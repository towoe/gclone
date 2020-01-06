package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/towoe/gclone/git"
)

type remote struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type repo struct {
	Remotes   []remote `json:"remotes"`
	Directory string   `json:"directory"`
}

type directoryContent struct {
	valid   bool
	status  git.RepoStatus
	Remotes []remote `json:"remotes"`
	name    string
}

type Register struct {
	Repos map[string]directoryContent
}

var (
	_currentRegister     *Register
	_storageFileRegister string
)

func getStorageFile(storageFile string) string {
	var storage string
	var err error
	storage, err = filepath.Abs(storageFile)
	if storageFile == "" || err != nil {
		storageFolder := getXDGDataFoler()
		storage = filepath.Join(storageFolder, "register.json")
	}
	return storage
}

func getXDGDataFoler() string {
	var xdgData string
	if xdgData = os.Getenv("XDG_DATA_HOME"); xdgData == "" {
		xdgData = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(xdgData, "gclone")
}

func CurrentRegister(storageFile string) *Register {
	_storageFileRegister = getStorageFile(storageFile)
	os.MkdirAll(filepath.Dir(_storageFileRegister), 0755)

	_currentRegister = &Register{}
	_currentRegister.Repos = make(map[string]directoryContent)
	newConfigService().Load(_currentRegister, _storageFileRegister)
	return _currentRegister
}

// TODO error
func (r *Register) Add(gitDir string) error {
	gitDir, err := filepath.Abs(gitDir)
	if err != nil {
		return err
	}

	remotes, err := git.ExtractFetchRemotes(gitDir)
	if err != nil {
		return err
	}
	dirRemotes := make([]remote, 0)
	for name, url := range remotes {
		dirRemotes = append(dirRemotes, remote{Name: name, URL: url})
	}
	fmt.Println("Adding ", gitDir)
	r.Repos[gitDir] = directoryContent{
		valid:   true,
		Remotes: dirRemotes,
	}
	newConfigService().Store(r, _storageFileRegister)
	return nil
}

func (r *Register) remove(rep repo) {
}

func (r *Register) Fetch(options string) error {
	for k := range r.Repos {
		func(k string, options string) {
			fmt.Println("Fetching: ", k)
			c := git.NewFetch(k, options)
			c.Run()
		}(k, options)
	}
	return nil
}

func (r *Register) Clone(url string, destDir string) error {
	c := git.NewClone(url, destDir)
	err := c.Run()
	if err != nil {
		return err
	}
	if destDir == "" {
		// get the last path element from the URL
		urlWithoutSlash := strings.TrimRight(url, "/")
		urlWithoutGit := strings.TrimSuffix(urlWithoutSlash, ".git")
		k := strings.Split(urlWithoutGit, "/")
		destDir = k[len(k)-1]
	}
	err = r.Add(destDir)
	if err != nil {
		return err
	}
	return nil
}

type ListSort int

const (
	Directory ListSort = iota
	Remote
)

func (r *Register) List(s ListSort) {
	r.setStatus()
	if s == Directory {
		r.listDirs()
	} else if s == Remote {
		r.listRemotes()
	}
}

func (r *Register) getSortedKeys() []string {
	// Sort the output based on the directory path
	var dirs []string
	for d := range r.Repos {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}

func (r *Register) listDirs() {

	sortedDirs := r.getSortedKeys()

	for _, dir := range sortedDirs {
		dirContent := r.Repos[dir]
		if dirContent.valid {
			fmt.Printf("%s: %s\tRemotes: ", dir, dirContent.status)
			for k, remote := range dirContent.Remotes {
				// TODO function for getting the status assembled
				// 	in a dynamic way
				stRemote := git.StatusRemote(dir, remote.Name)
				fmt.Printf("%s: %s",
					remote.Name, stRemote)
				if k < len(dirContent.Remotes)-1 {
					fmt.Printf(", ")
				}
			}
			if len(dirContent.Remotes) == 0 {
				fmt.Print("none set")
			}
		} else {
			// TODO: collect all entries and handle after the loop
			fmt.Println("Could not validate: ", dir)
		}
		fmt.Println()
	}
	// TODO: add option to remove invalid entriies
}

func (r *Register) listRemotes() {
	for URL, dirs := range r.createRemoteList() {
		fmt.Println(URL)
		for _, dir := range dirs {
			st, _ := git.Status(dir)
			fmt.Printf("%s:\t%s\n", st, dir)
		}
		fmt.Println()
	}
}

func (r *Register) setStatus() {
	var wg sync.WaitGroup
	res := make(chan directoryContent, len(r.Repos))
	for k, v := range r.Repos {
		wg.Add(1)
		go func(wg *sync.WaitGroup, dir string, old directoryContent, status chan directoryContent) {
			st, err := git.Status(dir)
			old.name = dir
			if err == nil {
				old.status = st
				old.valid = true
			}
			wg.Done()
			status <- old
		}(&wg, k, v, res)
	}
	wg.Wait()
	close(res)
	for {
		v1, ok := <-res
		if !ok {
			break
		}
		r.Repos[v1.name] = v1
	}
}

func (r *Register) createRemoteList() map[string][]string {
	// TODO: handle repos without remotes
	repoList := make(map[string][]string)
	for d, repo := range r.Repos {
		for _, remote := range repo.Remotes {
			repoList[remote.URL] = append(repoList[remote.URL], d)
		}
	}
	return repoList
}
