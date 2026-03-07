package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illikainen/attest/log"
)

func Run() error {
	log.Setup(&log.Options{
		ShowPrefix: true,
	})

	cmd := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}

	commands := map[string]func([]string) error{
		"genkey":        genkey,
		"sign":          sign,
		"sign-tree":     signTree,
		"verify":        verify,
		"verify-commit": verifyCommit,
	}

	var names []string
	for name := range commands {
		names = append(names, name)
	}

	args := cmd.Args()
	if len(args) < 1 {
		return fmt.Errorf("missing action: %s", strings.Join(names, ", "))
	}

	fn := commands[args[0]]
	if fn == nil {
		return fmt.Errorf("invalid action: %s\nactions: %s", args[0], strings.Join(names, ", "))
	}
	return fn(args[1:])
}
