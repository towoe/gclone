package main

import (
	"flag"
	"os"
	"regexp"

	"github.com/towoe/gclone/repo"
)

// Usage
// gclone https://github.com/user/repo [folder]
// 	- clone to [folder] or ./repo
// gclone path/to/folder
// 	- print status for folder OR
// 	- add folder to index
// gclone
// 	- print index

// Options
// indexFile

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

	registry := repo.CurrentRegister(*indexFile)

	if *fetch {
		registry.Fetch(*fetchOptions)
	}

	if flag.NArg() > 0 {
		urlExp, _ := regexp.Compile("^(git|http).*")
		// Decide if argument is address for cloning or a local path to add
		if urlExp.MatchString(flag.Arg(0)) {
			dest, _ := os.Getwd()
			if flag.NArg() > 1 {
				// Use second argument as clone destination
				dest = flag.Arg(1)
			}
			registry.Clone(flag.Arg(0), dest)
		} else {
			registry.Add(flag.Arg(0))
		}
	} else {
		sort := repo.Directory
		if *listOrder == "remote" {
			sort = repo.Remote
		}
		// Print list of entries
		registry.List(sort)
	}
}
