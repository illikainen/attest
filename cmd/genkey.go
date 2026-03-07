package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/illikainen/attest/crypto"
	"github.com/illikainen/attest/log"
)

func genkey(args []string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	cmd := flag.NewFlagSet("genkey", flag.ExitOnError)
	pubFile := cmd.String("p", filepath.Join(usr.HomeDir, ".attest", "key.pub"), "Public key file")
	privFile := cmd.String("s", filepath.Join(usr.HomeDir, ".attest", "key.sec"), "Private key file")
	cmd.Func("v", "Verbosity", log.SetVerbosity)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if _, err := os.Stat(*pubFile); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", *pubFile)
	}

	if _, err := os.Stat(*privFile); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", *privFile)
	}

	pubKey, privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	if err := pubKey.Write(*pubFile); err != nil {
		return err
	}

	if err := privKey.Write(*privFile); err != nil {
		return err
	}

	printf("Public key saved as %s\n", *pubFile)
	printf("Private key saved as %s\n", *privFile)
	return nil
}
