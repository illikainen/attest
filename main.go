package main

import (
	"os"

	"github.com/illikainen/attest/cmd"
	"github.com/illikainen/attest/git"
	"github.com/illikainen/attest/log"
)

func main() {
	if git.IsGitHook() {
		if err := git.Run(); err != nil {
			log.Errorf("%s", err)
			os.Exit(1)
		}
	} else {
		if err := cmd.Run(); err != nil {
			log.Errorf("%s", err)
			os.Exit(1)
		}
	}
}
