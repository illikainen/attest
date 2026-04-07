package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"log/slog"

	"github.com/illikainen/attest/internal/git"
	"github.com/illikainen/attest/internal/log"
	"github.com/illikainen/attest/internal/textutil"
)

func verifyTree(name string, args []string, opts *Options) error {
	cmd := flag.NewFlagSet(name+" verify-tree [<ref>]", flag.ExitOnError)
	repo := cmd.String("C", ".", "Git repository")
	verbosity := cmd.String("v", opts.Verbosity, "Verbosity")

	if err := cmd.Parse(args); err != nil {
		return err
	}
	opts.LogLevel.Set(log.ParseLevel(*verbosity))

	refs := cmd.Args()
	if len(refs) == 0 {
		refs = []string{"HEAD"}
	}

	for _, ref := range refs {
		tree, uid, err := git.VerifyTree(*repo, ref)
		if err != nil {
			return err
		}

		var prettyTree bytes.Buffer
		if err := json.Indent(&prettyTree, tree, "", "  "); err != nil {
			return err
		}
		slog.Debug("tree", "json", prettyTree.String())

		abbrev, err := git.RevParse(*repo, ref, true)
		if err != nil {
			return err
		}
		textutil.Printf("Signature for %s (%s) verified as coming from %s\n", ref, abbrev, uid)
	}
	return nil
}
