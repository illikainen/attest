package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illikainen/attest/git"
	"github.com/illikainen/attest/log"
	"github.com/illikainen/attest/utils"
)

func signTree(args []string) error {
	cmd := flag.NewFlagSet("sign-tree [ref]", flag.ExitOnError)
	repo := cmd.String("C", ".", "Git repository")
	treeFile := cmd.String("m", "", "Signed tree file")
	sigFile := cmd.String("x", "", "Signature file")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	var ref string
	refs := cmd.Args()

	if len(refs) > 1 {
		return fmt.Errorf("only one ref can be specified")
	} else if len(refs) == 1 {
		ref = refs[0]
	} else {
		ref = "HEAD"
	}

	tree, sig, err := git.SignTree(*repo, ref)
	if err != nil {
		return err
	}

	if *treeFile == "" {
		*treeFile = strings.ReplaceAll(ref, string(filepath.Separator), "_") + ".json"
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

	abbrev, err := git.RevParse(*repo, ref, true)
	if err != nil {
		return err
	}
	utils.Printf("Tree for %s saved as %s\n", abbrev, *treeFile)
	utils.Printf("Signature saved as %s\n", *sigFile)
	return nil
}
