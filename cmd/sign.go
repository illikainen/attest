package cmd

import (
	"flag"
	"fmt"

	"github.com/illikainen/attest/crypto"
	"github.com/illikainen/attest/log"
	"github.com/illikainen/attest/utils"
)

func sign(args []string) error {
	cmd := flag.NewFlagSet("sign", flag.ExitOnError)
	file := cmd.String("m", "", "File to sign")
	sigFile := cmd.String("x", "", "Signature file")
	privFile := cmd.String("s", "~/.attest/key.sec", "Private key file")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("the flag -m <file> is required")
	}

	if *sigFile == "" {
		*sigFile = *file + ".sig"
	}

	privKey, err := crypto.ReadPrivateKey(*privFile)
	if err != nil {
		return err
	}

	sig, err := privKey.SignFile(*file)
	if err != nil {
		return err
	}

	if err := sig.Write(*sigFile); err != nil {
		return err
	}

	utils.Printf("Signature for %s saved as %s\n", *file, *sigFile)
	return nil
}
