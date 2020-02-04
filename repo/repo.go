package repo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/towoe/gclone/git"
)

// remote is the data structure for a remote entry of a repository.
// A repository can have multiple remotes.
type remote struct {
	name   string
	url    string
	status git.RepoRemoteDiff
}

// directoryContent contains all the information for a single repository
type directoryContent struct {
	valid   bool
	status  git.RepoStatus
	remotes []remote
	dirName string
}

// Register stores a map of all registered repositories
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
// (directory) and adds them to the Register
func (r *Register) LoadRemotes() {
	for directory := range r.Repos {
		content, err := getRemotes(directory)
		if err == ErrNotAGitDir {
			log.Println("Error checking for remotes.",
				err, directory)
		} else if err != nil {
			log.Fatalln(err)
		}
		r.Repos[directory] = content
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
			fmt.Printf("\033[36mFetching: %v\033[m\n", k)
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
func (r *Register) RemoveInvalidEntries(m DeleteMethod) bool {
	var removed bool
	for k, v := range r.Repos {
		if !v.valid {
			if m == DeleteAsk {
				fmt.Printf("Delete [%v] from the storage file [Y/n/a/q] ", v.dirName)
				var ans string = "Y"
				fmt.Scanf("%s", &ans)
				if strings.HasPrefix(ans, "q") {
					break
				} else if strings.HasPrefix(ans, "a") {
					m = DeleteAll
				} else if !(strings.HasPrefix(ans, "y") || strings.HasPrefix(ans, "Y")) {
					continue
				}
			}
			r.remove(k)
			removed = true
		}
	}
	return removed
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
	bar := pb.StartNew(len(r.Repos))
	bar.SetTemplate(`Collecting statuses: {{counters . }}`)
	for range r.Repos {
		v1, ok := <-res
		bar.Increment()
		if !ok {
			break
		}
		r.Repos[v1.dirName] = v1
	}
	bar.Finish()
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

// List prints all the entries from the storage file
func (r *Register) List() {
	for _, dir := range sortedKeysContent(&r.Repos) {
		fmt.Println(substituteWithTilde(dir))
	}
}

// Status gathers the state for each entry and prints the status in a list
// sorted by the specified argument
func (r *Register) Status(key string, sorted string, reverse bool) {
	r.setStatus()
	r.updateRemotestatus()
	var l []statusLine
	if key == "remote" {
		l = r.listByRemotes()
	} else {
		l = r.listByDirs()
	}
	// Sorting by key always happens, so if the status is sorted
	// the keys are sorted in the status group as well
	var kk sort.Interface = byKey(l)
	if reverse {
		kk = sort.Reverse(byKey(l))
	}
	sort.Stable(kk)
	if sorted == "status" {
		var ls sort.Interface = byStatus(l)
		if reverse {
			ls = sort.Reverse(byStatus(l))
		}
		sort.Stable(ls)
	}
	removeDuplicateKey(l)
	alignWithSpace(l)
	printLines(l)
}

// statusLine stores the status for each entry.
// The purpose for this is to have the ability to change the text
// which is printed as an output. An example for this would be to
// insert spaces in order to align the second entries.
type statusLine struct {
	key         string
	status      string
	statusColor string
	info        string
}

type byKey []statusLine

func (k byKey) Len() int {
	return len(k)
}
func (k byKey) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}
func (k byKey) Less(i, j int) bool {
	return strings.Compare(k[i].key, k[j].key) == -1
}

type byStatus []statusLine

func (s byStatus) Len() int {
	return len(s)
}
func (s byStatus) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byStatus) Less(i, j int) bool {
	return strings.Compare(s[i].status, s[j].status) == -1
}

func printLines(s []statusLine) {
	for _, v := range s {
		fmt.Printf("%v  %v%v\033[0m  %v\n", v.key, v.statusColor, v.status, v.info)
	}
}

// removeDuplicateKey from a list sorted by key
func removeDuplicateKey(s []statusLine) {
	for i, v := range s {
		// Check all previous keys
		for j := i; j > 0; j-- {
			if s[j-1].key == "" {
				// Go back further for comparison
				continue
			}
			if strings.Compare(s[i].key, s[j-1].key) == 0 {
				// Current key matches one from a previous entry
				v.key = ""
				s[i] = v
			}
		}
	}
}

func alignWithSpace(s []statusLine) {
	var maxLenK, maxLenS int
	for i := range s {
		if entryLen := len(s[i].status); entryLen > maxLenS {
			maxLenS = entryLen
		}
		if entryLen := len(s[i].key); entryLen > maxLenK {
			maxLenK = entryLen
		}
	}
	for i := range s {
		s[i].key += strings.Repeat(" ", maxLenK-len(s[i].key))
		s[i].status += strings.Repeat(" ", maxLenS-len(s[i].status))
	}
}

func (r *Register) listByDirs() []statusLine {
	sList := make([]statusLine, 0, 10)
	for k, dirContent := range r.Repos {
		sEntry := statusLine{}
		if dirContent.valid {
			sEntry.key = fmt.Sprintf("%s:", substituteWithTilde(k))
			sEntry.status = dirContent.status.String()
			sEntry.statusColor = dirContent.status.Color()
			for k, remote := range dirContent.remotes {
				rem := fmt.Sprintf("%s: %s",
					remote.name, remote.status)
				if k < len(dirContent.remotes)-1 {
					rem += fmt.Sprintf(", ")
				}
				sEntry.info += rem
			}
			if len(dirContent.remotes) == 0 {
				sEntry.info = fmt.Sprint("none set")
			}
			sList = append(sList, sEntry)
		}
	}
	return sList
}

func (r *Register) listByRemotes() []statusLine {
	sList := make([]statusLine, 0)
	for directory, dirContent := range r.Repos {
		for _, remote := range dirContent.remotes {
			sEntry := statusLine{}
			sEntry.key = remote.url
			sEntry.status = dirContent.status.String()
			sEntry.statusColor = dirContent.status.Color()
			sEntry.info = substituteWithTilde(directory)
			sList = append(sList, sEntry)
		}
	}
	return sList
}
