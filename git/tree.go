package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/illikainen/attest/crypto"
	"github.com/illikainen/attest/log"
	"github.com/illikainen/attest/utils"
)

type Tree struct {
	Type    string    `json:"type"`
	Repo    string    `json:"repo"`
	Objects []*Object `json:"objects"`
}

func (t *Tree) Len() int {
	return len(t.Objects)
}

func (t *Tree) Encode() ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	return data, nil
}

func ListTree(dir string, ref string) (*Tree, error) {
	realRef, err := RevParse(dir, ref, false)
	if err != nil {
		return nil, err
	}

	objs, err := listTree(dir, realRef)
	if err != nil {
		return nil, err
	}

	subObjs, err := listSubmoduleTrees(dir, "HEAD")
	if err != nil {
		return nil, err
	}

	objs = append(objs, subObjs...)
	sort.Slice(objs, func(i int, j int) bool {
		if objs[i].SubmoduleDir == "" && objs[j].SubmoduleDir == "" {
			return objs[i].Path < objs[j].Path
		}

		if objs[i].SubmoduleDir == "" {
			return true
		}

		if objs[j].SubmoduleDir == "" {
			return false
		}

		if objs[i].SubmoduleDir != objs[j].SubmoduleDir {
			return objs[i].SubmoduleDir < objs[j].SubmoduleDir
		}
		return objs[i].Path < objs[j].Path
	})

	repo, err := repoName(dir)
	if err != nil {
		return nil, err
	}

	return &Tree{
		Type:    "git",
		Repo:    repo,
		Objects: objs,
	}, nil
}

func listTree(dir string, ref string) ([]*Object, error) {
	// %x00 isn't used because ls-tree is executed with `-z`
	format := "%(objectmode)%x01%(objectname)%x01%(path)%x01%(objectsize)%x01%(objecttype)"
	p, err := Git(dir, "ls-tree", "-z", "-r", "--format", format, "--", ref)
	if err != nil {
		return nil, err
	}

	var objs []*Object
	for _, line := range strings.Split(string(p.Stdout), "\x00") {
		if line == "" {
			continue
		}

		elts := strings.Split(line, "\x01")
		if len(elts) != 5 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		obj, err := DecodeObject(dir, ref, elts[0], elts[1], elts[2], elts[3], elts[4], "")
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}

	return objs, nil
}

func listSubmoduleTrees(dir string, ref string) ([]*Object, error) {
	// %x00 isn't used because ls-tree is executed with `-z`
	format := "%(objectmode)%x01%(objectname)%x01%(path)%x01%(objectsize)%x01%(objecttype)"
	p, err := Git(dir, "submodule", "foreach", "--recursive",
		"git ls-tree -z -r --format '"+format+"' "+ref)
	if err != nil {
		return nil, err
	}

	if len(p.Stdout) == 0 {
		return nil, nil
	}

	rx, err := regexp.Compile(`^Entering '([^']+)'$`)
	if err != nil {
		return nil, err
	}

	var submoduleDir string
	var objs []*Object
	for _, elt := range strings.Split(string(p.Stdout), "\x00") {
		// Split on \n because the initial "Entering '<submodule>'"
		// from `git submodule foreach` and the first object line from
		// `git ls-tree -z` are separated by \n.
		for _, line := range strings.Split(elt, "\n") {
			m := rx.FindStringSubmatch(line)
			if m != nil {
				submoduleDir = m[1]
			} else {
				if submoduleDir == "" {
					return nil, fmt.Errorf("not in a submodule")
				}

				if line != "" {
					elts := strings.Split(line, "\x01")
					if len(elts) != 5 {
						return nil, fmt.Errorf("invalid line: %s", line)
					}

					obj, err := DecodeObject(dir, ref, elts[0], elts[1],
						elts[2], elts[3], elts[4], submoduleDir)
					if err != nil {
						return nil, err
					}
					objs = append(objs, obj)
				}
			}
		}
	}

	return objs, nil
}

func repoName(dir string) (string, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	return filepath.Base(filepath.Clean(path)), nil
}

func VerifyTree(dir string, ref string) ([]byte, string, error) {
	allowedSigners, err := readAllowedSigners(dir)
	if err != nil {
		return nil, "", err
	}

	sig, err := getTreeSignature(dir, ref)
	if err != nil {
		return nil, "", err
	}

	tree, err := ListTree(dir, ref)
	if err != nil {
		return nil, "", err
	}

	if tree.Len() <= 0 {
		return nil, "", fmt.Errorf("%s: no objects to sign", ref)
	}

	msg, err := tree.Encode()
	if err != nil {
		return nil, "", err
	}

	for _, signer := range allowedSigners {
		log.Debugf("%s: trying %s", ref, signer.UID)
		if bytes.Equal(signer.PublicKey.KeyID[:], sig.KeyID[:]) {
			if signer.PublicKey.Verify(msg, sig) == nil {
				log.Debugf("%s: signed by %s", ref, signer.UID)
				return msg, signer.UID, nil
			}
		}
	}

	return nil, "", fmt.Errorf("%s: not signed by a trusted key", ref)
}

const signatureMarker = "Tree-Signature:"

func getTreeSignature(dir string, ref string) (*crypto.Signature, error) {
	p, err := Git(dir, "log", "--no-show-signature", "-1", "--pretty=%B", ref)
	if err != nil {
		return nil, err
	}
	msg := strings.Trim(string(p.Stdout), " \t\r\n")

	if _, err := ValidatePrintable(msg, false); err != nil {
		return nil, err
	}

	for _, line := range strings.Split(msg, "\n") {
		if strings.HasPrefix(line, "#") {
			break
		}

		elts := strings.Split(line, " ")
		if len(elts) == 2 && strings.HasPrefix(elts[0], signatureMarker) {
			sig, err := crypto.DecodeSignature([]byte(elts[1]))
			if err != nil {
				return nil, err
			}
			return sig, nil
		}
	}

	return nil, fmt.Errorf("%s is not signed by a trusted key", ref)
}

func SignTree(dir string, ref string) ([]byte, *crypto.Signature, error) {
	submodules, err := SubmoduleStatus(dir)
	if err != nil {
		return nil, nil, err
	}

	for _, submodule := range submodules {
		if !submodule.Initialized {
			return nil, nil, fmt.Errorf("%s: recursively initialize submodules before signing", dir)
		}
	}

	privKey, err := readPrivateKey(dir)
	if err != nil {
		return nil, nil, err
	}

	tree, err := ListTree(dir, ref)
	if err != nil {
		return nil, nil, err
	}

	if tree.Len() <= 0 {
		return nil, nil, fmt.Errorf("%s: no objects to verify", ref)
	}

	msg, err := tree.Encode()
	if err != nil {
		return nil, nil, err
	}

	sig, err := privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return msg, sig, err
}

type Signer struct {
	UID       string
	PublicKey *crypto.PublicKey
}

func readAllowedSigners(dir string) ([]*Signer, error) {
	allowedSigners, err := ConfigGet(dir, "attest", "allowedSigners", "")
	if err != nil {
		return nil, err
	}

	if allowedSigners == "" {
		return nil, fmt.Errorf("attest.allowedSigners must be configured in .gitconfig")
	}

	path, err := utils.ExpandPath(allowedSigners)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, err
	}

	var signers []*Signer
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.Trim(line, " \t")
		if line != "" && !strings.HasPrefix(line, "#") {
			elts := strings.Split(line, " ")
			if len(elts) != 2 {
				return nil, fmt.Errorf("%s: bad line: %s", allowedSigners, line)
			}

			pubKey, err := crypto.DecodePublicKey([]byte(elts[1]))
			if err != nil {
				return nil, err
			}

			log.Debugf("trusting %s (%s)", elts[0], pubKey)
			signers = append(signers, &Signer{UID: elts[0], PublicKey: pubKey})
		}
	}

	return signers, nil
}

func readPrivateKey(dir string) (*crypto.PrivateKey, error) {
	privFile, err := ConfigGet(dir, "attest", "privkey", "")
	if err != nil {
		return nil, err
	}

	if privFile == "" {
		return nil, fmt.Errorf("attest.privkey must be configured in .gitconfig")
	}

	path, err := utils.ExpandPath(privFile)
	if err != nil {
		return nil, err
	}

	privKey, err := crypto.ReadPrivateKey(path)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func WriteTree(dir string) (string, error) {
	p, err := Git(dir, "write-tree")
	if err != nil {
		return "", err
	}

	return strings.Trim(string(p.Stdout), " \t\r\n"), nil
}
