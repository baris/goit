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
	GitwebUrl    string
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

//func (r *GitRepo) GitwebUrl() string {
//	return "https://" + GitwebServerName + "?p=" + r.RelativePath
//}

func (r *GitRepo) GitDir() string {
	if r.Type == Bare {
		return r.Path
	}
	return r.Path + "/.git"
}

func (r *GitRepo) RunCommand(args ...string) (out string, ok bool) {
	cmd := exec.Command("git")
	cmd.Args = append([]string{"git", "--git-dir", r.GitDir()}, args...)
	cmd_out, err := cmd.Output()
	if err != nil {
		return string(cmd_out), false
	}
	return string(cmd_out), true
}

func (r *GitRepo) LastCommit() (info *GitCommitInfo, ok bool) {
	infos, ok := r.LastCommitsN(1)
	if ok {
		return infos[0], true
	}
	return nil, false
}

func (r *GitRepo) LastCommitsN(n int) (infos GitCommitInfos, ok bool) {
	out, ok := r.RunCommand("log", "-"+strconv.Itoa(n), "--format=%H#%ae#%ar#%s")
	if ok != true {
		return nil, false
	}
	lines := strings.Split(strings.Trim(out, " \n"), "\n")
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

func (r *GitRepo) Show(sha string) string {
	out, ok := r.RunCommand("show", sha)
	if ok != true {
		return ""
	}
	return out
}

func (r *GitRepo) Heads() (heads []string) {
	out, ok := r.RunCommand("for-each-ref", "--count", "10")
	if ok != true {
		return
	}
	for _, line := range strings.Split(strings.Trim(out, " \n"), "\n") {
		head := strings.Split(strings.Split(line, " ")[1], "\t")[1]
		if strings.HasPrefix(head, "refs/heads/") {
			heads = append(heads, head[len("refs/heads/"):])
		}
	}
	return
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
	relPath = strings.TrimLeft(relPath, "/")
	if len(relPath) == 0 {
		relPath = "."
	}
	repo.RelativePath = relPath

	repo.GitwebUrl = "https://" + GitwebServerName + "?p=" + repo.RelativePath

	return repo, true
}
