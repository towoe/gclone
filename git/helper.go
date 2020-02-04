package git

import (
	"gopkg.in/src-d/go-git.v4"
)

func ExtractFetchRemotes(directory string) (map[string]string, error) {
	repo, err := git.PlainOpen(directory)
	if err != nil {
		return nil, err
	}

	r := make(map[string]string, 1)
	remotes, err := repo.Remotes()
	for _, v := range remotes {
		r[v.Config().Name] = v.Config().URLs[0]
	}
	return r, nil
}

type RepoStatus int

const (
	Undefined RepoStatus = iota + 1
	Clean
	Dirty
)

func (rp RepoStatus) String() string {
	if rp == Clean {
		return "Clean"
	} else if rp == Dirty {
		//return "Dirty\033[0m"
		return "Dirty"
	} else {
		return ""
	}
}

func (rp RepoStatus) Color() string {
	if rp == Clean {
		return "\033[32m"
	} else if rp == Dirty {
		return "\033[31m"
	} else {
		return ""
	}
}

func Status(directory string) (RepoStatus, error) {
	repo, err := git.PlainOpen(directory)
	if err != nil {
		return Undefined, err
	}

	work, err := repo.Worktree()
	if err != nil {
		return Undefined, err
	}

	status, err := work.Status()

	if len(status) == 0 {
		return Clean, nil
	} else {
		return Dirty, nil
	}
}
