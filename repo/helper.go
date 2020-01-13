package repo

import (
	"os/user"
	"strings"
)

func substituteWithTilde(dir string) string {
	u, _ := user.Current()
	if strings.HasPrefix(dir, u.HomeDir) {
		return strings.Replace(dir, u.HomeDir, "~", 1)
	}
	return dir
}
