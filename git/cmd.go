package git

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type Cmd struct {
	Name   string
	Args   []string
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
	Dir    string
}

func NewGit() *Cmd {
	return &Cmd{
		Name:   "git",
		Args:   []string{},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    "",
	}
}

func (cmd *Cmd) Run() error {
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdin = cmd.Stdin
	// TODO: Collect output to buffer and return it
	c.Stdout = cmd.Stdout
	c.Stderr = cmd.Stderr
	c.Dir = cmd.Dir
	err := c.Run()
	if err != nil {
		log.Println("Error running", c)
		return err
	}
	return nil
}

func (cmd *Cmd) Output() ([]byte, error) {
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stderr = cmd.Stderr
	c.Dir = cmd.Dir
	out, err := c.Output()
	return out, err
}

func (cmd *Cmd) SetArg(arg string) *Cmd {
	cmd.Args = append(cmd.Args, arg)
	return cmd
}

func (cmd *Cmd) SetArgs(args ...string) *Cmd {
	for _, a := range args {
		if a != "" {
			cmd.SetArg(a)
		}
	}
	return cmd
}

func NewClone(url string, directory string) *Cmd {
	return NewGit().SetArgs("clone", url, directory)
}

func NewFetch(directory string, options string) *Cmd {
	f := []string{"fetch"}
	f = append(f, strings.Split(options, " ")...)
	g := NewGit().SetArgs(f...)
	g.Dir = directory
	return g
}

func NewCurrentBranch(directory string) *Cmd {
	g := NewGit().SetArgs("branch", "--show-current")
	g.Dir = directory
	return g
}

func NewRevListCount(directory string, remote string, branch string) *Cmd {
	g := NewGit().SetArgs("rev-list", "--count", remote+"/"+branch+"...")
	g.Dir = directory
	return g
}

type RepoRemoteDiff int

const (
	Ahead RepoRemoteDiff = iota + 1
	Behind
	UpToDate
	Changed
)

func (rr RepoRemoteDiff) String() string {
	if rr == Ahead {
		return "\033[31mAhead\033[0m"
	} else if rr == Behind {
		return "\033[31mBehind\033[0m"
	} else if rr == UpToDate {
		return "\033[32mUp to date\033[0m"
	} else if rr == Changed {
		return "\033[31mChanged\033[0m"
	} else {
		return ""
	}
}

func StatusRemote(directory string, remote string) RepoRemoteDiff {
	branch, _ := NewCurrentBranch(directory).Output()
	branch_name := strings.Trim(string(branch), "\n")
	rev, _ := NewRevListCount(directory, remote, string(branch_name)).Output()
	if strings.HasPrefix(string(rev), "0") {
		return UpToDate
	} else {
		return Changed
	}
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
	g := NewGit().SetArgs("--no-optional-locks", "status", "--porcelain=v1")
	g.Dir = directory
	s, _ := g.Output()
	if len(s) == 0 {
		return Clean, nil
	} else {
		return Dirty, nil
	}
}
