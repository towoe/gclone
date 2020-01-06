package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetXDGDataFolderHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	os.Setenv("HOME", home)

	gcloneDir := filepath.Join(home, ".local/share/gclone")

	p := getXDGDataFoler()
	if p != gcloneDir {
		t.Errorf("getXDGDataFoler() == %q, want %q", p, gcloneDir)
	}
}

func TestGetXDGDataFolderXDG(t *testing.T) {
	home, _ := os.UserHomeDir()
	os.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local/share"))

	gcloneDir := filepath.Join(home, ".local/share/gclone")

	p := getXDGDataFoler()
	if p != gcloneDir {
		t.Errorf("getXDGDataFoler() == %q, want %q", p, gcloneDir)
	}
}

func TestGetStorageFile(t *testing.T) {

	homeDir, _ := os.UserHomeDir()
	workingDir, _ := os.Getwd()
	os.Setenv("HOME", homeDir)

	cases := []struct {
		in, want string
	}{
		{"", filepath.Join(homeDir, ".local/share/gclone/register.json")},
		{"repo.json", filepath.Join(workingDir, "repo.json")},
	}
	for _, c := range cases {
		got := getStorageFile(c.in)
		if got != c.want {
			t.Errorf("getStorageFile(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
