package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

import (
	"github.com/baris/ldap"
	"github.com/gorilla/sessions"
)

const JSONAPIError = "{\"error\":1}"
const JSONLoginError = "{\"login_error\":1}"

var BaseGitDir string
var GitwebServerName string

var port string
var runServer bool
var excludeRegexpString string
var excludeRegexp *regexp.Regexp
var repositories map[string]*GitRepo // Repo.Name:Repo
var repositoriesLock sync.Mutex
var sslCertFile string
var sslKeyFile string
var ldapURL string
var ldapDomain string
var cookieSecret string
var store *sessions.CookieStore

func addRepository(repo *GitRepo) {
	repositoriesLock.Lock()
	repositories[repo.Path] = repo
	repositoriesLock.Unlock()
}

func isExcluded(path string) bool {
	if excludeRegexpString != "" {
		return excludeRegexp.MatchString(path)
	}
	return false
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, "goit")
	if _, ok := session.Values["login"]; ok {
		return true
	}
	return false
}

func walk(path string, controlChannel chan bool) {
	walkerChannel := make(chan bool, 100)
	walkerCount := 0

	if isExcluded(path) {
		controlChannel <- false
		return
	}

	repo, ok := NewRepo(path)
	if ok {
		addRepository(repo)
		controlChannel <- true
		return
	}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		controlChannel <- false
		return
	}

	for _, info := range infos {
		if info.IsDir() {
			walkerCount += 1
			dirPath := filepath.Join(path, info.Name())
			go walk(dirPath, walkerChannel)
		}
	}

	// wait for walkers
	for i := 0; i < walkerCount; i++ {
		<-walkerChannel
	}

	controlChannel <- true
}

func findRepositories() {
	controlChannel := make(chan bool)
	go walk(BaseGitDir, controlChannel)
	<-controlChannel // wait for walkers
}

func sortedRepositories() GitRepos {
	pathList := make(GitRepos, len(repositories))
	i := 0
	for _, path := range repositories {
		pathList[i] = path
		i++
	}
	sort.Sort(pathList)
	return pathList
}

func handleRepository(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
		return
	}

	parts := strings.SplitN(r.URL.Path, "/", 3)
	repository := parts[2]
	http.Redirect(w, r, "/repository.html#"+repository, 302)
}

func handleAPIRepositories(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
		return
	}

	repositories = make(map[string]*GitRepo)
	findRepositories()
	pathList := sortedRepositories()
	stringPathList := []string{}
	for _, path := range pathList {
		stringPathList = append(stringPathList, path.Json())
	}
	fmt.Fprintf(w, "[\n")
	fmt.Fprintf(w, strings.Join(stringPathList, ",\n"))
	fmt.Fprintf(w, "\n]\n")
}

func handleAPIRepository(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
		return
	}

	repository := strings.SplitN(r.URL.Path, "/", 3)[2]
	path := filepath.Join(BaseGitDir, repository)
	if isExcluded(path) {
		fmt.Fprintf(w, JSONAPIError)
		return
	}
	repo, ok := NewRepo(path)
	if ok {
		fmt.Fprintf(w, repo.Json())
	} else {
		fmt.Fprintf(w, JSONAPIError)
	}
}

func handleAPIRepositoryTip(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
	}

	parts := strings.SplitN(r.URL.Path, "/", 4)
	branch := parts[2]
	repository := parts[3]

	if isExcluded(repository) {
		fmt.Fprintf(w, JSONAPIError)
		return
	}

	if branch != "master" {
		// TODO: support other branches.
		fmt.Fprintf(w, JSONAPIError)
		return
	}

	path := filepath.Join(BaseGitDir, repository)
	if repo, ok := NewRepo(path); ok {
		if tip, ok := repo.LastCommit(); ok {
			fmt.Fprintf(w, "["+repo.Json()+","+tip.Json()+"]")
			return
		}
	}
	fmt.Fprintf(w, JSONAPIError)
}

func handleAPIShow(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
		return
	}

	parts := strings.SplitN(r.URL.Path, "/", 5)
	branch := parts[2]
	sha := parts[3]
	repository := parts[4]

	if branch != "master" {
		// TODO: support other branches.
		fmt.Fprintf(w, JSONAPIError)
		return
	}

	path := filepath.Join(BaseGitDir, repository)
	if repo, ok := NewRepo(path); ok {
		if b, err := json.Marshal(repo.Show(sha)); err == nil {
			fmt.Fprintf(w, string(b))
			return
		}
	}
	fmt.Fprintf(w, JSONAPIError)
}

func handleAPIHeads(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
		return
	}

	parts := strings.SplitN(r.URL.Path, "/", 3)
	repository := parts[2]

	path := filepath.Join(BaseGitDir, repository)
	if repo, ok := NewRepo(path); ok {
		if b, err := json.Marshal(repo.Heads()); err == nil {
			fmt.Fprintf(w, "["+repo.Json()+",")
			fmt.Fprintf(w, string(b))
			fmt.Fprintf(w, "]")
			return
		}
	}
	fmt.Fprintf(w, JSONAPIError)
}

func handleAPICommits(w http.ResponseWriter, r *http.Request) {
	if !isLoggedIn(w, r) {
		fmt.Fprintf(w, JSONLoginError)
	}

	parts := strings.SplitN(r.URL.Path, "/", 5)
	branch := parts[2]
	numCommits, err := strconv.Atoi(parts[3])
	if err != nil {
		fmt.Fprintf(w, JSONAPIError)
		return
	}
	repository := parts[4]

	if isExcluded(repository) {
		fmt.Fprintf(w, JSONAPIError)
		return
	}

	if branch != "master" {
		// TODO: support other branches.
		fmt.Fprintf(w, JSONAPIError)
		return
	}

	path := filepath.Join(BaseGitDir, repository)
	if repo, ok := NewRepo(path); ok {
		infos, ok := repo.LastCommitsN(numCommits)
		if ok != true {
			fmt.Fprintf(w, JSONAPIError)
			return
		}

		infoStrings := []string{}
		for _, info := range infos {
			infoStrings = append(infoStrings, info.Json())
		}
		fmt.Fprintf(w, "["+repo.Json()+",[\n")
		fmt.Fprintf(w, strings.Join(infoStrings, ",\n"))
		fmt.Fprintf(w, "]]")
	}
}

func handleLDAPLogin(w http.ResponseWriter, r *http.Request) {
	username, password := "", ""
	if err := r.ParseForm(); err == nil {
		username = r.PostFormValue("username")
		password = r.PostFormValue("password")
	}

	if l, err := ldap.Dial("tcp", ldapURL); err == nil {
		defer l.Close()
		if err := l.Bind(username+"@"+ldapDomain, password); err == nil {
			session, _ := store.Get(r, "goit")
			session.Values["login"] = username
			session.Save(r, w)
			http.Redirect(w, r, "/index.html", 302)
			return
		}
	}
	http.Redirect(w, r, "/login.html", 302)
}

func printRepositories() {
	repositories = make(map[string]*GitRepo)
	findRepositories()
	for _, repo := range sortedRepositories() {
		fmt.Println(repo.Json())
	}
}

func main() {
	flag.StringVar(&GitwebServerName, "gitwebServer", "localhost", "Gitweb server's hostname")
	flag.BoolVar(&runServer, "runServer", false, "Run web server or just print repositories")
	flag.StringVar(&port, "port", "8080", "Port to listen from")
	flag.StringVar(&excludeRegexpString, "excludeRegexp", "", "Exlude paths from being listed")
	flag.StringVar(&sslCertFile, "certFile", "", "SSL Certificate path")
	flag.StringVar(&sslKeyFile, "keyFile", "", "SSL Key path")
	flag.StringVar(&ldapURL, "ldapUrl", "", "LDAP URL (i.e. ldap.example.com:321)")
	flag.StringVar(&ldapDomain, "ldapDomain", "", "LDAP domain (i.e. example.com)")
	flag.StringVar(&cookieSecret, "cookieSecret", "", "Secret key to authenticate sessions")
	flag.Parse()

	if re, err := regexp.Compile(excludeRegexpString); err != nil {
		println("Regexp error")
		os.Exit(1)
	} else {
		excludeRegexp = re
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	if path, err := filepath.Abs(flag.Arg(0)); err == nil {
		BaseGitDir = path
	} else {
		BaseGitDir = flag.Arg(0)
	}

	if runServer {
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
		http.HandleFunc("/r/", handleRepository)
		http.HandleFunc("/repositories/", handleAPIRepositories)
		http.HandleFunc("/repository/", handleAPIRepository)
		http.HandleFunc("/commits/", handleAPICommits)
		http.HandleFunc("/heads/", handleAPIHeads)
		http.HandleFunc("/show/", handleAPIShow)
		http.HandleFunc("/tip/", handleAPIRepositoryTip)
		http.HandleFunc("/login/", handleLDAPLogin)

		store = sessions.NewCookieStore([]byte(cookieSecret))

		if sslCertFile != "" && sslKeyFile != "" {
			http.ListenAndServeTLS(":"+port, sslCertFile, sslKeyFile, nil)
		} else {
			http.ListenAndServe(":"+port, nil)
		}
	} else {
		printRepositories()
	}
}
