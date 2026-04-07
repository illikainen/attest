package git

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/illikainen/attest/internal/bwrap"
)

func RevParse(dir string, ref string, short bool) (string, error) {
	cmd := []string{"rev-parse", "--verify"}
	if short {
		cmd = append(cmd, "--short")
	}
	cmd = append(cmd, ref)

	p, err := Git(dir, cmd...)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(p.Stdout), " \t\r\n"), nil
}

func IsGitHook() bool {
	return os.Getenv("GIT_EXEC_PATH") != ""
}

func Git(dir string, args ...string) (*bwrap.Process, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return bwrap.Bubblewrap(&bwrap.Options{
		Command: append([]string{"git", "-C", dir}, args...),
		RO: []string{
			filepath.Join(usr.HomeDir, ".gitconfig"),
			filepath.Join(usr.HomeDir, ".config", "git", "config"),
			"/etc/gitconfig",
		},
		RW: []string{dirAbs},
	})
}
