package cmd

import (
	"flag"
	"fmt"

	"github.com/illikainen/attest/internal/log"
	"github.com/illikainen/attest/internal/textutil"
	"github.com/illikainen/attest/pkg/attest"
)

func verify(name string, args []string, opts *Options) error {
	cmd := flag.NewFlagSet(name+" verify", flag.ExitOnError)
	file := cmd.String("m", "", "File to verify")
	sigFile := cmd.String("x", "", "Signature file")
	pubFile := cmd.String("p", "~/.attest/key.pub", "Public key file")
	verbosity := cmd.String("v", opts.Verbosity, "Verbosity")

	if err := cmd.Parse(args); err != nil {
		return err
	}
	opts.LogLevel.Set(log.ParseLevel(*verbosity))

	if *file == "" {
		return fmt.Errorf("the flag -m <file> is required")
	}

	if *sigFile == "" {
		*sigFile = *file + ".sig"
	}

	pubKey, err := attest.ReadPublicKey(*pubFile)
	if err != nil {
		return err
	}

	sig, err := attest.ReadSignature(*sigFile)
	if err != nil {
		return err
	}

	err = pubKey.VerifyFile(*file, sig)
	if err != nil {
		return err
	}

	textutil.Printf("Signature verified using %s\n", *pubFile)
	return nil
}
