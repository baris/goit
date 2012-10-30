package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

import "git"

var port string
var runServer bool
var baseGitDir string
var gitwebServerName string
var repositories map[string] *git.GitRepo // Repo.Name:Repo


func relativeGitPath(path string) string {
	gitPath := path[len(baseGitDir):]
	if gitPath[0] == '/' {
		gitPath = gitPath[1:]
	}
	return gitPath
}

func gitwebUrl(path string) string {
	return "https://" + gitwebServerName + "?p=" + relativeGitPath(path)
}

func walk(path string, controlChannel chan bool) {
	walkerChannel := make(chan bool, 100)
	walkerCount := 0

	repo, ok := GetRepo(path)
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

func init() {
	repositories = make(map[string] *git.GitRepo)

	flag.StringVar(&gitwebServerName, "gitwebServer", "localhost", "Gitweb server's hostname")
	flag.StringVar(&baseGitDir, "baseGitDir", "/git", "Base Git directory on server")
	flag.BoolVar(&runServer, "runServer", false, "Run web server or just print repositories")
	flag.StringVar(&port, "port", "8080", "Port to listen from")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
}

func findRepositories() {
	controlChannel := make(chan bool)

	for _, dir := range flag.Args() {
		go walk(dir, controlChannel)
	}

	// wait walkers
	for _ = range flag.Args() {
		<-controlChannel
	}
}

func sortedRepositories() git.GitRepos {
	pathList := make(git.GitRepos, len(repositories))
	i := 0
	for _, path := range repositories {
		pathList[i] = path
		i++
	}
	sort.Sort(pathList)
	return pathList
}

func main() {
	curdir, _ := os.Getwd()
	if runServer {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			repositories = make(map[string] *git.GitRepo)
			findRepositories()
			pathList := sortedRepositories()
			fmt.Fprintf(w, "<html><body>")
			for _, repo := range pathList {
				fmt.Fprintf(w, "<a href='"+gitwebUrl(repo.Path)+"'>"+relativeGitPath(repo.Path)+"<a><br>")
			}
			fmt.Fprintf(w, "</body></html>")
		})
		http.ListenAndServe(":"+port, nil)
	} else {
		findRepositories()
		for _, repo := range sortedRepositories() {
			fmt.Println(repo)
		}
	}
	os.Chdir(curdir)
}
