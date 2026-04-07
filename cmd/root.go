package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/illikainen/attest/internal/log"
)

type Options struct {
	Verbosity string
	LogLevel  *slog.LevelVar
}

func Run() error {
	opts := &Options{
		LogLevel: setupLog(),
	}
	commands := map[string]func(string, []string, *Options) error{
		"genkey":      genkey,
		"sign":        sign,
		"verify":      verify,
		"verify-tree": verifyTree,
	}

	bin := filepath.Base(os.Args[0])
	subcommands := strings.Join(keys(commands), "|")
	cmd := flag.NewFlagSet(fmt.Sprintf("%s [flags] <%s>", bin, subcommands), flag.ExitOnError)
	cmd.StringVar(&opts.Verbosity, "v", "warn", "Verbosity")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		return err
	}
	opts.LogLevel.Set(log.ParseLevel(opts.Verbosity))

	args := cmd.Args()
	if len(args) < 1 {
		cmd.Usage()
		return fmt.Errorf("missing command")
	}

	fn := commands[args[0]]
	if fn == nil {
		cmd.Usage()
		return fmt.Errorf("invalid command")
	}
	return fn(bin, args[1:], opts)
}

func setupLog() *slog.LevelVar {
	level := new(slog.LevelVar)
	slog.SetDefault(slog.New(log.NewSanitizedHandler(os.Stderr, &log.HandlerOptions{
		Name:  "attest",
		Level: level,
	})))
	return level
}

func keys[K comparable, V any](m map[K]V) []K {
	rv := []K{}
	for k := range m {
		rv = append(rv, k)
	}
	return rv
}
