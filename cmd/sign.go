package cmd

import (
	"flag"
	"fmt"

	"github.com/illikainen/attest/internal/log"
	"github.com/illikainen/attest/internal/textutil"
	"github.com/illikainen/attest/pkg/attest"
)

func sign(name string, args []string, opts *Options) error {
	cmd := flag.NewFlagSet(name+" sign", flag.ExitOnError)
	file := cmd.String("m", "", "File to sign")
	sigFile := cmd.String("x", "", "Signature file")
	privFile := cmd.String("s", "~/.attest/key.sec", "Private key file")
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

	privKey, err := attest.ReadPrivateKey(*privFile)
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

	textutil.Printf("Signature for %s saved as %s\n", *file, *sigFile)
	return nil
}
