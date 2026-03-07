package cmd

import (
	"flag"
	"os"

	"github.com/illikainen/attest/git"
	"github.com/illikainen/attest/log"
)

func signTree(args []string) error {
	cmd := flag.NewFlagSet("sign-tree", flag.ExitOnError)
	repo := cmd.String("C", ".", "Git repository")
	ref := cmd.String("ref", "HEAD", "Ref to sign")
	treeFile := cmd.String("m", "", "Signed tree file")
	sigFile := cmd.String("x", "", "Signature file")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	tree, sig, err := git.SignTree(*repo, *ref)
	if err != nil {
		return err
	}

	if *treeFile == "" {
		*treeFile = *ref + ".json"
	}

	if *sigFile == "" {
		*sigFile = *treeFile + ".sig"
	}

	if err := os.WriteFile(*treeFile, tree, 0o600); err != nil {
		return err
	}

	if err := sig.Write(*sigFile); err != nil {
		return err
	}

	printf("Tree content saved as %s\n", *treeFile)
	printf("Signature saved as %s\n", *sigFile)
	return nil
}
