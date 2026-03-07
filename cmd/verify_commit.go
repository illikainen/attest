package cmd

import (
	"bytes"
	"encoding/json"
	"flag"

	"github.com/illikainen/attest/git"
	"github.com/illikainen/attest/log"
	"github.com/illikainen/attest/utils"
)

func verifyCommit(args []string) error {
	cmd := flag.NewFlagSet("sign", flag.ExitOnError)
	repo := cmd.String("C", ".", "Git repository")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

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

		log.Debugf("tree:\n%s", prettyTree.String())

		abbrev, err := git.RevParse(*repo, ref, true)
		if err != nil {
			return err
		}
		utils.Printf("Signature for %s (%s) verified as coming from %s\n", ref, abbrev, uid)
	}
	return nil
}
