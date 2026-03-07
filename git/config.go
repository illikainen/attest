package git

import (
	"strings"

	"github.com/illikainen/attest/log"
)

func IsDisabled(dir string) bool {
	disabled, err := ConfigGet(dir, "attest", "disabled", "false")
	if err != nil {
		return false
	}

	if disabled == "true" {
		log.Warnf("signing and verification is disabled")
		return true
	}
	return false
}

func ConfigGet(dir string, group string, key string, fallback string) (string, error) {
	p, err := Git(dir, "config", "get", "--default", fallback, "--", group+"."+key)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(p.Stdout), " \t\r\n"), nil
}
