package main

import (
	"log/slog"
	"os"

	"github.com/illikainen/attest/cmd"
	"github.com/illikainen/attest/internal/git"
)

func main() {
	if git.IsGitHook() {
		if err := git.Run(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	} else {
		if err := cmd.Run(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
}
