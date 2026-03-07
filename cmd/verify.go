package cmd

import (
	"flag"
	"fmt"

	"github.com/illikainen/attest/crypto"
	"github.com/illikainen/attest/log"
)

func verify(args []string) error {
	cmd := flag.NewFlagSet("verify", flag.ExitOnError)
	file := cmd.String("m", "", "File to verify")
	sigFile := cmd.String("x", "", "Signature file")
	pubFile := cmd.String("p", "~/.attest/key.pub", "Public key file")
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

	pubKey, err := crypto.ReadPublicKey(*pubFile)
	if err != nil {
		return err
	}

	sig, err := crypto.ReadSignature(*sigFile)
	if err != nil {
		return err
	}

	err = pubKey.VerifyFile(*file, sig)
	if err != nil {
		return err
	}

	printf("Signature verified using %s\n", *pubFile)
	return nil
}
