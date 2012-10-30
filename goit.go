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

var BaseGitDir string
var GitwebServerName string

var port string
var runServer bool
var repositories map[string] *GitRepo // Repo.Name:Repo


func walk(path string, controlChannel chan bool) {
	walkerChannel := make(chan bool, 100)
	walkerCount := 0

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

	for _, dir := range flag.Args() {
		go walk(dir, controlChannel)
	}

	// wait walkers
	for _ = range flag.Args() {
		<-controlChannel
	}
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
	fmt.Fprintf(w, "<html><body><table>")
	for _, repo := range pathList {
		fmt.Fprintf(w, "<tr><td><a href='"+repo.GitwebUrl()+"'>"+repo.RelativePath()+"<a></td>")
		info := repo.LatestCommit
		if info != nil {
			fmt.Fprintf(w, "<td>"+info.SHA+"<td>"+info.Author+"<td>"+info.Date+"</td></tr>")
		} else {
			fmt.Fprintf(w, "<td span=3></td></tr>")
		}
	}
	fmt.Fprintf(w, "</table></body></html>")
}


func printRepositories() {
	repositories = make(map[string] *GitRepo)
	findRepositories()
	for _, repo := range sortedRepositories() {
		info := repo.LatestCommit
		if info != nil {
			fmt.Println(repo.RelativePath() + " " + info.String())
		} else {
			fmt.Println(repo.RelativePath())
		}
	}
}


func main() {
	flag.StringVar(&GitwebServerName, "gitwebServer", "localhost", "Gitweb server's hostname")
	flag.StringVar(&BaseGitDir, "baseGitDir", "/git", "Base Git directory on server")
	flag.BoolVar(&runServer, "runServer", false, "Run web server or just print repositories")
	flag.StringVar(&port, "port", "8080", "Port to listen from")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	curdir, _ := os.Getwd()
	if runServer {
		http.HandleFunc("/", handleRoot)
		http.ListenAndServe(":"+port, nil)
	} else {
		printRepositories()
	}
	os.Chdir(curdir)
}
