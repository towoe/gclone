package main

import (
	"flag"
	"regexp"

	"github.com/towoe/gclone/repo"
)

var (
	indexFile = flag.String("index", "", "Path to file containg the index")
	listOrder = flag.String("list", "dir", "Sort the list by: \"dir\", "+
		"\"remote\"")
	fetch        = flag.Bool("fetch", false, "Fetch remotes")
	fetchOptions = flag.String("fetch-options", "", "Additional arguments "+
		"for fetch")
)

func main() {
	flag.Parse()

	r := repo.CurrentRegister(*indexFile)
	r.LoadRemotes()

	if *fetch {
		r.Fetch(*fetchOptions)
	}

	if flag.NArg() > 0 {
		urlExp, _ := regexp.Compile("^(git|http).*")
		// Decide if argument is address for cloning or a local path to add
		if urlExp.MatchString(flag.Arg(0)) {
			// Use second argument as clone destination. Will be an
			// empty string if no second argument is provided
			r.Clone(flag.Arg(0), flag.Arg(1))
		} else {
			r.Add(flag.Arg(0))
		}
	} else {
		sort := repo.Directory
		if *listOrder == "remote" {
			sort = repo.Remote
		}
		// Print list of entries
		r.List(sort)
	}
	r.Store()
}
