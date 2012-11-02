package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const JSONAPIError = "{status=\"error\"}"
var BaseGitDir string
var GitwebServerName string

var port string
var runServer bool
var excludeRegexpString string
var excludeRegexp *regexp.Regexp
var repositories map[string] *GitRepo // Repo.Name:Repo


func isExcluded(path string) bool {
	if excludeRegexpString != "" {
		return excludeRegexp.MatchString(path)
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
		repositories[repo.Name] = repo
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


func handleRoot(w http.ResponseWriter, r *http.Request) {
	repositories = make(map[string] *GitRepo)
	findRepositories()
	pathList := sortedRepositories()
	fmt.Fprintf(w, `<html><head>
<title>Goit: Go Git web interface</title>
<link rel="stylesheet" type="text/css" href="/files/bootstrap.min.css">
<link rel="stylesheet" type="text/css" href="/files/base.css">
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.7.1/jquery.min.js"></script>
<script type="text/javascript" src="/files/goit.js"></script>
</head>
<body>
<div class="page-header">
  <h1> [[ Goit ]] <small>for %s</small></h1>
</div>
<input id=search type="text" placeholder="Search…">
<table>
`, GitwebServerName)
	for _, repo := range pathList {
		fmt.Fprintf(w,
			"<tr id=" + toCSSName(repo.RelativePath) + ">" +
			"<td><a href='" + repo.GitwebUrl() + "'>" + repo.RelativePath + "<a></td>" +
			"<td id=" + toCSSName(repo.RelativePath) + "-sha></td>" +
			"<td id=" + toCSSName(repo.RelativePath) + "-author></td>" +
			"<td id=" + toCSSName(repo.RelativePath) + "-date></td>" +
			"</tr>")
	}
	fmt.Fprintf(w, "</table></body></html>")
}


func handleAPIRepositories(w http.ResponseWriter, r *http.Request) {
	repositories = make(map[string] *GitRepo)
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
		if tip, ok := repo.GetRepoTip(); ok {
			fmt.Fprintf(w, tip.Json())
			return
		}
	}
	fmt.Fprintf(w, JSONAPIError)
}


func printRepositories() {
	repositories = make(map[string] *GitRepo)
	findRepositories()
	for _, repo := range sortedRepositories() {
		println(repo.Json())
	}
}


func main() {
	flag.StringVar(&GitwebServerName, "gitwebServer", "localhost", "Gitweb server's hostname")
	flag.BoolVar(&runServer, "runServer", false, "Run web server or just print repositories")
	flag.StringVar(&port, "port", "8080", "Port to listen from")
	flag.StringVar(&excludeRegexpString, "excludeRegexp", "", "Exlude paths from being listed")
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
	if path, err := filepath.Abs(flag.Arg(0)); err != nil {
		BaseGitDir = path
	} else {
		BaseGitDir = flag.Arg(0)
	}

	curdir, _ := os.Getwd()
	if runServer {
		http.HandleFunc("/", handleRoot)
		http.HandleFunc("/repositories/", handleAPIRepositories)
		http.HandleFunc("/repository/", handleAPIRepository)
		http.HandleFunc("/tip/", handleAPIRepositoryTip)

		http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./static"))))
		http.ListenAndServe(":"+port, nil)
	} else {
		printRepositories()
	}
	os.Chdir(curdir)
}
