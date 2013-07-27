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

var repositories map[string]*GitRepo // Repo.Name:Repo
var repositoriesLock sync.Mutex

var cookieStore *sessions.CookieStore
var excludeRegexp *regexp.Regexp
var configFile string
var config Config

type Config struct {
	Git_dir string
	Server bool
	Port string
	Exclude string
	SslCertFile string
	SslKeyFile string
	Ldap_url string
	Ldap_domain string
	Cookie_secret string
	Gitwebserver_name string
}

func addRepository(repo *GitRepo) {
	repositoriesLock.Lock()
	repositories[repo.Path] = repo
	repositoriesLock.Unlock()
}

func isExcluded(path string) bool {
	if config.Exclude != "" {
		return excludeRegexp.MatchString(path)
	}
	return false
}

func enableAuthentication() bool {
	return config.Ldap_url != "" && config.Ldap_domain != ""
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	if !enableAuthentication() {
		return true
	}

	session, _ := cookieStore.Get(r, "goit")
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
	go walk(config.Git_dir, controlChannel)
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
	path := filepath.Join(config.Git_dir, repository)
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

	path := filepath.Join(config.Git_dir, repository)
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

	path := filepath.Join(config.Git_dir, repository)
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

	path := filepath.Join(config.Git_dir, repository)
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

	path := filepath.Join(config.Git_dir, repository)
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

	if l, err := ldap.Dial("tcp", config.Ldap_url); err == nil {
		defer l.Close()
		if err := l.Bind(username+"@"+config.Ldap_domain, password); err == nil {
			session, _ := cookieStore.Get(r, "goit")
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
	flag.StringVar(&configFile, "config", "config.json", "JSON formatted configuration file")
	flag.Parse()

	conf_contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(conf_contents, &config)

	if re, err := regexp.Compile(config.Exclude); err != nil {
		fmt.Println("Regexp error")
		os.Exit(1)
	} else {
		excludeRegexp = re
	}

	if config.Server {
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
		http.HandleFunc("/r/", handleRepository)
		http.HandleFunc("/repositories/", handleAPIRepositories)
		http.HandleFunc("/repository/", handleAPIRepository)
		http.HandleFunc("/commits/", handleAPICommits)
		http.HandleFunc("/heads/", handleAPIHeads)
		http.HandleFunc("/show/", handleAPIShow)
		http.HandleFunc("/tip/", handleAPIRepositoryTip)
		if enableAuthentication() {
			http.HandleFunc("/login/", handleLDAPLogin)
			cookieStore = sessions.NewCookieStore([]byte(config.Cookie_secret))
		}

		if config.SslCertFile != "" && config.SslKeyFile != "" {
			http.ListenAndServeTLS(":"+config.Port, config.SslCertFile, config.SslKeyFile, nil)
		} else {
			http.ListenAndServe(":"+config.Port, nil)
		}
	} else {
		printRepositories()
	}
}
