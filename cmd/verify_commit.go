package cmd

import (
	"bytes"
	"encoding/json"
	"flag"

	"github.com/illikainen/attest/git"
	"github.com/illikainen/attest/log"
)

func verifyCommit(args []string) error {
	cmd := flag.NewFlagSet("sign", flag.ExitOnError)
	repo := cmd.String("C", ".", "Git repository")
	ref := cmd.String("r", "HEAD", "Ref to verify")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	tree, uid, err := git.VerifyTree(*repo, *ref)
	if err != nil {
		return err
	}

	var prettyTree bytes.Buffer
	if err := json.Indent(&prettyTree, tree, "", "  "); err != nil {
		return err
	}

	log.Debugf("tree:\n%s", prettyTree.String())
	printf("Signature for %s verified as coming from %s\n", *ref, uid)
	return nil
}
