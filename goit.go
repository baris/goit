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

var port string
var runServer bool
var baseGitDir string
var gitwebServerName string
var repositories map[string] string // repositoryName:path

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

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func has(path, fileOrDir string) bool {
	return exists(filepath.Join(path, fileOrDir))
}

func isGitPath(path string) bool {
	return  has(path, "refs/heads") || has(path, ".git")
}

func walk(path string, controlChannel chan bool) {
	walkerChannel := make(chan bool, 100)
	walkerCount := 0

	path, _ = filepath.Abs(path)
	if isGitPath(path) {
		repositories[filepath.Base(path)] = path
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
	for i :=0; i < walkerCount; i++ {
		<-walkerChannel
	}

	controlChannel <- true
}

func init() {
	repositories = make(map[string] string)
	
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

func sortedRepositories() []string {
	pathList := make([]string, len(repositories))
	i := 0
	for _, path := range repositories {
		pathList[i] = path
		i++
	}
	sort.Strings(pathList)
	return pathList
}

func main() {
	curdir, _ := os.Getwd()
	if runServer {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			repositories = make(map[string] string)
			findRepositories()
			pathList := sortedRepositories()
			fmt.Fprintf(w, "<html><body>")
			for _, path := range pathList {
				fmt.Fprintf(w, "<a href='" + gitwebUrl(path) + "'>" + relativeGitPath(path) + "<a><br>")
			}
			fmt.Fprintf(w, "</body></html>")
		})
		http.ListenAndServe(":" + port, nil)
	} else {
		findRepositories()
		for _, path := range sortedRepositories() {
			fmt.Println(path)
		}
	}
	os.Chdir(curdir)
}
