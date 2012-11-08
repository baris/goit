package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type GitRepoType int8

const (
	Unknown GitRepoType = iota
	NonBare
	Bare
)

type GitCommitInfo struct {
	SHA     string
	Author  string
	Date    string
	Subject string
}

type GitCommitInfos []*GitCommitInfo

type GitRepo struct {
	Name         string
	Path         string
	RelativePath string
	Type         GitRepoType
}

type GitRepos []*GitRepo

func (i *GitCommitInfo) String() string {
	return i.SHA + ", " + i.Author + ", " + i.Date + ", " + i.Subject
}

func (i *GitCommitInfo) Json() string {
	if b, err := json.Marshal(i); err == nil {
		return string(b)
	}
	return "{status=\"error\"}"
}

// sort.Interface for array type
func (r GitRepos) Less(i, j int) bool { return r[i].RelativePath < r[j].RelativePath }
func (r GitRepos) Len() int           { return len(r) }
func (r GitRepos) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func (r *GitRepo) String() string {
	return r.Name
}

func (r *GitRepo) Json() string {
	if b, err := json.Marshal(r); err == nil {
		return string(b)
	}
	return JSONAPIError
}

func (r *GitRepo) GitwebUrl() string {
	return "https://" + GitwebServerName + "?p=" + r.RelativePath
}

func (r *GitRepo) RepoTip() (info *GitCommitInfo, ok bool) {
	gitDir := r.Path
	if r.Type == NonBare {
		gitDir += "/.git"
	}
	out, err := exec.Command("git", "--git-dir", gitDir, "log", "-1", "--format=%h#%ae#%ar#%s").Output()
	if err != nil {
		return nil, false
	}
	info = new(GitCommitInfo)
	line := strings.Trim(string(out), " \n")
	parts := strings.Split(line, "#")
	info.SHA = parts[0]
	info.Author = parts[1]
	info.Date = parts[2]
	info.Subject = parts[3]
	return info, true
}

func (r *GitRepo) LastCommitsN(n int) (infos GitCommitInfos, ok bool) {
	gitDir := r.Path
	if r.Type == NonBare {
		gitDir += "/.git"
	}
	out, err := exec.Command("git", "--git-dir", gitDir, "log", "-"+strconv.Itoa(n), "--format=%h#%ae#%ar#%s").Output()
	if err != nil {
		return nil, false
	}
	lines := strings.Split(strings.Trim(string(out), " \n"), "\n")
	infos = make(GitCommitInfos, len(lines))
	for index, line := range lines {
		info := new(GitCommitInfo)
		parts := strings.Split(strings.Trim(line, " \n"), "#")
		info.SHA = parts[0]
		info.Author = parts[1]
		info.Date = parts[2]
		info.Subject = parts[3]
		infos[index] = info
	}
	return infos, true
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

	relPath := path[len(BaseGitDir):]
	if relPath[0] == '/' {
		relPath = relPath[1:]
	}
	repo.RelativePath = relPath

	return repo, true
}
