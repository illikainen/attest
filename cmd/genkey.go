package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/illikainen/attest/internal/log"
	"github.com/illikainen/attest/internal/textutil"
	"github.com/illikainen/attest/pkg/attest"
)

func genkey(name string, args []string, opts *Options) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	cmd := flag.NewFlagSet(name+" genkey", flag.ExitOnError)
	pubFile := cmd.String("p", filepath.Join(usr.HomeDir, ".attest", "key.pub"), "Public key file")
	privFile := cmd.String("s", filepath.Join(usr.HomeDir, ".attest", "key.sec"), "Private key file")
	verbosity := cmd.String("v", opts.Verbosity, "Verbosity")

	if err := cmd.Parse(args); err != nil {
		return err
	}
	opts.LogLevel.Set(log.ParseLevel(*verbosity))

	if _, err := os.Stat(*pubFile); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", *pubFile)
	}

	if _, err := os.Stat(*privFile); err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s already exists", *privFile)
	}

	pubKey, privKey, err := attest.GenerateKey()
	if err != nil {
		return err
	}

	if err := pubKey.Write(*pubFile); err != nil {
		return err
	}

	if err := privKey.Write(*privFile); err != nil {
		return err
	}

	textutil.Printf("Public key saved as %s\n", *pubFile)
	textutil.Printf("Private key saved as %s\n", *privFile)
	return nil
}
