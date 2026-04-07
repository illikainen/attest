package git

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/illikainen/attest/internal/log"
	"github.com/illikainen/attest/internal/textutil"
	"github.com/illikainen/attest/pkg/attest"
)

func Run() error {
	hooks := map[string]func(string) error{
		"prepare-commit-msg":    prepareCommitMsg,
		"reference-transaction": referenceTransaction,
		"post-checkout":         postCheckout,
	}
	hook := os.Getenv("GIT_HOOK")

	dir := os.Getenv("GIT_PREFIX")
	if dir == "" {
		dir = "."
	}

	if IsDisabled(dir) {
		return nil
	}

	verbosity, err := ConfigGet(dir, "attest", "verbosity", "warn")
	if err != nil {
		return err
	}

	handler := log.NewSanitizedHandler(os.Stderr, &log.HandlerOptions{
		Name:  "attest",
		Level: log.ParseLevel(verbosity),
	})
	logger := slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("hook", hook),
	}))
	slog.SetDefault(logger)

	if fn := hooks[hook]; fn != nil {
		return fn(dir)
	}
	return nil
}

func prepareCommitMsg(dir string) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("unexpected number of arguments")
	}

	source := ""
	if len(os.Args) == 3 {
		source = os.Args[2]
	}

	if source != "" && source != "message" {
		return fmt.Errorf("invalid source: %s", source)
	}

	ref, err := WriteTree(dir)
	if err != nil {
		return err
	}

	// `ref` would be misleading because that's not the to-be-committed sha1
	slog.Info("signing index")

	tree, sig, err := SignTree(dir, ref)
	if err != nil {
		return err
	}

	var prettyTree bytes.Buffer
	if err := json.Indent(&prettyTree, tree, "", "  "); err != nil {
		return err
	}

	msg, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}

	slog.Debug(
		"commit",
		"tree", prettyTree.String(),
		"ref", ref,
		"sig", sig,
		"msg", strings.Split(string(msg), "\n")[0],
		"source", source,
	)
	newMsg, err := replaceTreeSignature(sig, string(msg))
	if err != nil {
		return err
	}

	return os.WriteFile(os.Args[1], []byte(newMsg), 0o600)
}

func replaceTreeSignature(sig *attest.Signature, msg string) (string, error) {
	rx, err := regexp.Compile(`^[a-zA-Z0-9+/=]{100}$`)
	if err != nil {
		return "", err
	}

	var lines []string
	var generatedLines []string
	generated := false
	replaced := false

	for _, line := range strings.Split(msg, "\n") {
		if strings.HasPrefix(line, "#") {
			generated = true
		}

		if !generated {
			elts := strings.Split(line, " ")
			if len(elts) == 2 && strings.HasPrefix(elts[0], signatureMarker) {
				if rx.MatchString(elts[1]) {
					line = signatureMarker + " " + sig.String()
					replaced = true
				}
			}
			lines = append(lines, line)
		} else {
			generatedLines = append(generatedLines, line)
		}
	}

	if !replaced {
		lines = append(lines, signatureMarker+" "+sig.String())
	}

	return strings.Join(append(lines, generatedLines...), "\n"), nil
}

func referenceTransaction(dir string) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("unexpected number of arguments")
	}

	state := os.Args[1]
	if state != "prepared" {
		return nil
	}

	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		elts := strings.Split(scan.Text(), " ")
		newRef := elts[1]
		refName := elts[2]

		if slices.Contains([]string{"HEAD", "ORIG_HEAD", "AUTO_MERGE", "CHERRY_PICK_HEAD"}, refName) {
			continue
		}

		if newRef == strings.Repeat("0", 40) {
			continue
		}

		if strings.Contains(newRef, ":") {
			newRef = strings.SplitN(newRef, ":", 2)[1]
		}

		_, uid, err := VerifyTree(dir, newRef)
		if err != nil {
			return err
		}

		newAbbrev, err := RevParse(dir, newRef, true)
		if err != nil {
			return err
		}

		textutil.Printf("\033[32mattest\033[0m: %s (%s): signed by %s (reference-transaction)\n",
			refName, newAbbrev, uid)
	}

	return scan.Err()
}

func postCheckout(dir string) error {
	if len(os.Args) != 4 {
		return fmt.Errorf("unexpected number of arguments")
	}

	newRef := os.Args[2]
	_, uid, err := VerifyTree(dir, newRef)
	if err != nil {
		return err
	}

	newAbbrev, err := RevParse(dir, newRef, true)
	if err != nil {
		return err
	}

	textutil.Printf("\033[32mattest\033[0m: %s: signed by %s (post-checkout)\n", newAbbrev, uid)
	return nil
}
