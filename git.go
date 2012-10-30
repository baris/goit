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
	LatestCommit *GitCommitInfo
}

type GitRepos []*GitRepo


func (i *GitCommitInfo) String() string {
	return i.SHA + ", " + i.Author + ", " + i.Date
}


// sort.Interface for array type
func (r GitRepos) Less(i, j int) bool { return r[i].RelativePath() < r[j].RelativePath() }
func (r GitRepos) Len() int { return len(r) }
func (r GitRepos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }


func (r *GitRepo) String() string {
	return r.Name
}


func (r *GitRepo) RelativePath() string {
	gitPath := r.Path[len(BaseGitDir):]
	if gitPath[0] == '/' {
		gitPath = gitPath[1:]
	}
	return gitPath
}


func (r *GitRepo) GitwebUrl() string {
	return "https://" + GitwebServerName + "?p=" + r.RelativePath()
}


func (r *GitRepo) GetLatestCommit() (info *GitCommitInfo, ok bool) {
	gitDir := r.Path
	if r.Type == NonBare {
		gitDir += "/.git"
	}
	out, err := exec.Command("git", "--git-dir", gitDir, "log", "-1", "--format=%h#%ae#%ar").Output()
	if err != nil {
		return nil, false
	}
	info = new(GitCommitInfo)
	line := strings.Trim(string(out), " \n")
	parts := strings.Split(line, "#")
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
	repo.LatestCommit = nil
	info, ok := repo.GetLatestCommit()
	if ok {
		repo.LatestCommit = info
	}
	return repo, true
}
