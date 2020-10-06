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
