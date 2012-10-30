package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepoType int8

const (
	Unknown GitRepoType = iota
	NonBare
	Bare
)

type GitCommitInfo struct {
	SHA string
	Author string
	Date string
}

type GitRepo struct {
	Name string
	Path string
	Type GitRepoType
}

type GitRepos []*GitRepo

// sort.Interface for array type
func (r GitRepos) Less(i, j int) bool { return r[i].Name < r[j].Name }
func (r GitRepos) Len() int { return len(r) }
func (r GitRepos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }


func (r *GitRepo) String() string {
	return r.Name
}


func (r *GitRepo) LatestCommit() (info *GitCommitInfo, ok bool) {
	out, err := exec.Command("git", "log", "HEAD^..HEAD", "--format='%h#%ae#%ar'").Output()
	if err != nil {
		return nil, false
	}
	info = new(GitCommitInfo)
	parts := strings.Split(string(out), "#")
	info.SHA = parts[0]
	info.Author = parts[1]
	info.Date = parts[2]
	return info, true
}


func GitPathType(path string) (repoType GitRepoType, ok bool) {
	ok = false
	repoType = Unknown
	if has(path, "refs/heads") && has(path, "objects") {
		ok = true
		repoType = Bare
	} else if has(path, ".git") {
		ok = true
		repoType = NonBare
	}
	return
}


func NewRepo(path string) (repo *GitRepo, ok bool) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, false
	}

	repoType, ok := GitPathType(path)
	if ok == false {
		return nil, false
	}

	repo = new(GitRepo)
	repo.Type = repoType
	repo.Path = path
	repo.Name = filepath.Base(path)
	return repo, true
}
