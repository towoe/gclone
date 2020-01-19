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
	name   string
	url    string
	status git.RepoRemoteDiff
}

type directoryContent struct {
	valid   bool
	status  git.RepoStatus
	remotes []remote
	dirName string
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

// LoadRemotes looks up the currently available remotes of a repository
// (direcotry) and adds them to the Register
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

// use an argument with a pointer and create the slice for different maps
func sortedKeysContent(m *map[string]directoryContent) []string {
	var l []string
	for k := range *m {
		l = append(l, k)
	}
	sort.Strings(l)
	return l
}

// TODO How to use the same signature as in sortedKeysContent
func sortedKeysString(m *map[string][]directoryContent) []string {
	var l []string
	for k := range *m {
		l = append(l, k)
	}
	sort.Strings(l)
	return l
}

type DeleteMethod int

const (
	DeleteAll DeleteMethod = iota
	DeleteAsk
)

// RemoveInvalidEntries remove entries which are not marked valid from the
// storage
func (r *Register) RemoveInvalidEntries(m DeleteMethod) {
	for k, v := range r.Repos {
		if !v.valid {
			if m == DeleteAsk {
				fmt.Printf("Delete [%v] from the storage file [Yna] ", v.dirName)
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

func (r *Register) updateRemotestatus() {
	for k, v := range r.Repos {
		if v.valid {
			for n, remote := range v.remotes {
				r.Repos[k].remotes[n].status =
					git.StatusRemote(v.dirName, remote.name)
			}
		}
	}
}

type ListSort int

const (
	Directory ListSort = iota
	Remote
)

func (r *Register) List(s string) {
	r.setStatus()
	r.updateRemotestatus()
	if s == "remote" {
		r.listByRemotes()
	} else {
		r.listByDirs()
	}
}

func (r *Register) listByDirs() {
	for _, dir := range sortedKeysContent(&r.Repos) {
		dirContent := r.Repos[dir]
		if dirContent.valid {
			fmt.Printf("%s: %s\tRemotes: ", substituteWithTilde(dir), dirContent.status)
			for k, remote := range dirContent.remotes {
				fmt.Printf("%s: %s",
					remote.name, remote.status)
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

func (r *Register) listByRemotes() {
	m := r.mapRemotes()
	for _, url := range sortedKeysString(&m) {
		dirCont := m[url]
		fmt.Printf("%s:\t", url)
		for _, d := range dirCont {
			fmt.Printf("%s: %s\n", substituteWithTilde(d.dirName), d.status)
		}
	}
}

func (r *Register) setStatus() {
	res := make(chan directoryContent, len(r.Repos))
	for k, v := range r.Repos {
		go func(dir string, old directoryContent, status chan directoryContent) {
			st, err := git.Status(dir)
			old.dirName = dir
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
		r.Repos[v1.dirName] = v1
	}
}

// iterate over all entries and just create a dumb slice
func (r *Register) mapRemotes() map[string][]directoryContent {
	rem := make(map[string][]directoryContent)
	for _, dirCont := range r.Repos {
		for _, remote := range dirCont.remotes {
			val, ok := rem[remote.url]
			if !ok {
				val = make([]directoryContent, 0)
			}
			rem[remote.url] = append(val, dirCont)
		}
	}
	return rem
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
