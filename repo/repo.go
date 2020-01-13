package repo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/towoe/gclone/git"
)

type remote struct {
	name string
	url  string
}

type repo struct {
	remotes   []remote
	Directory string `json:"directory"`
}

type directoryContent struct {
	valid   bool
	status  git.RepoStatus
	remotes []remote
	name    string
}

type Register struct {
	Repos map[string]directoryContent
}

var (
	_currentRegister     *Register
	_storageFileRegister string
)

var (
	ErrNotAGitDir = errors.New("No git structure found in directory")
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

// LoadRemotes load the remotes of a repository into the loaded data structure
func (r *Register) LoadRemotes() {
	for k := range r.Repos {
		d, err := getRemotes(k)
		if err == ErrNotAGitDir {
			log.Println("Error checking for remotes.", err, k)
		} else if err != nil {
			log.Fatalln(err)
		}
		r.Repos[k] = d
	}
}

// Store writes the current entries of r.Repo to the storage file
func (r *Register) Store() {
	newConfigService().Store(r, _storageFileRegister)
}

func getRemotes(gitDir string) (directoryContent, error) {

	remotes, err := git.ExtractFetchRemotes(gitDir)
	if err != nil {
		return directoryContent{valid: false}, ErrNotAGitDir
	}
	dirRemotes := make([]remote, 0)
	for name, url := range remotes {
		dirRemotes = append(dirRemotes, remote{name: name, url: url})
	}
	return directoryContent{valid: true, remotes: dirRemotes}, nil
}

// TODO error
func (r *Register) Add(gitDir string) error {
	gitDir, err := filepath.Abs(gitDir)
	if err != nil {
		return err
	}

	d, err := getRemotes(gitDir)
	if err != nil {
		return err
	}
	fmt.Println("Adding ", gitDir)
	r.Repos[gitDir] = d
	newConfigService().Store(r, _storageFileRegister)
	return nil
}

func (r *Register) remove(gitDir string) {
	delete(r.Repos, gitDir)
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
	r.removeInvalidEntries(DeleteAsk)
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

type DeleteMethod int

const (
	DeleteAll DeleteMethod = iota
	DeleteAsk
)

func (r *Register) removeInvalidEntries(m DeleteMethod) {
	for k, v := range r.Repos {
		if !v.valid {
			if m == DeleteAsk {
				fmt.Printf("Delete [%v] from the storage file [Yna] ", v.name)
				var ans string = "Y"
				fmt.Scanf("%s", &ans)
				if strings.HasPrefix(ans, "a") {
					m = DeleteAll
				} else if !(strings.HasPrefix(ans, "y") || strings.HasPrefix(ans, "Y")) {
					continue
				}
			}
			r.remove(k)
		}
	}
}

func (r *Register) listDirs() {

	sortedDirs := r.getSortedKeys()

	for _, dir := range sortedDirs {
		dirContent := r.Repos[dir]
		if dirContent.valid {
			fmt.Printf("%s: %s\tRemotes: ", substitueWithTilde(dir), dirContent.status)
			for k, remote := range dirContent.remotes {
				// TODO function for getting the status assembled
				// 	in a dynamic way
				stRemote := git.StatusRemote(dir, remote.name)
				fmt.Printf("%s: %s",
					remote.name, stRemote)
				if k < len(dirContent.remotes)-1 {
					fmt.Printf(", ")
				}
			}
			if len(dirContent.remotes) == 0 {
				fmt.Print("none set")
			}
			fmt.Println()
		}
	}
}

func (r *Register) listRemotes() {
	for url, dirs := range r.createRemoteList() {
		fmt.Println(url)
		for _, dir := range dirs {
			st, _ := git.Status(dir)
			fmt.Printf("%s:\t%s\n", st, dir)
		}
		fmt.Println()
	}
}

func (r *Register) setStatus() {
	res := make(chan directoryContent, len(r.Repos))
	for k, v := range r.Repos {
		go func(dir string, old directoryContent, status chan directoryContent) {
			st, err := git.Status(dir)
			old.name = dir
			if err == nil {
				old.status = st
				old.valid = true
			}
			status <- old
		}(k, v, res)
	}
	for range r.Repos {
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
		for _, remote := range repo.remotes {
			repoList[remote.url] = append(repoList[remote.url], d)
		}
	}
	return repoList
}
