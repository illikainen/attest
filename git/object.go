package git

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/illikainen/attest/crypto"
)

type Object struct {
	Mode         string `json:"mode"`
	SHA1         string `json:"sha1"`
	Path         string `json:"path"`
	Size         int    `json:"size"`
	Type         string `json:"type"`
	SubmoduleDir string `json:"submodule-dir"`
	SHA512       string `json:"sha512"`
}

func DecodeObject(dir, ref, mode, sha1, path, size, typ, submoduleDir string) (*Object, error) {
	if ref != "HEAD" && ref != "HEAD:"+submoduleDir {
		if _, err := ValidateSHA1(ref); err != nil {
			return nil, err
		}
	}

	if submoduleDir != "" {
		if _, err := ValidatePath(submoduleDir); err != nil {
			return nil, err
		}
	}

	if _, err := ValidateSHA1(sha1); err != nil {
		return nil, err
	}

	if _, err := ValidateObjectType(typ); err != nil {
		return nil, err
	}

	if _, err := ValidateObjectMode(mode); err != nil {
		return nil, err
	}

	if _, err := ValidatePath(path); err != nil {
		return nil, err
	}

	s := -1
	sha512 := ""
	if typ == "commit" {
		if mode != "160000" {
			return nil, fmt.Errorf("%s: %s: bad mode", path, mode)
		}
		if size != "-" {
			return nil, fmt.Errorf("%s: %s: bad size", path, size)
		}
	} else if typ == "blob" {
		var err error
		s, err = strconv.Atoi(size)
		if err != nil || s < 0 { // 0-size files can be committed
			return nil, fmt.Errorf("%s: %s: bad size: %w", path, size, err)
		}

		data, err := Git(filepath.Join(dir, submoduleDir), "cat-file", typ, sha1)
		if err != nil || len(data.Stdout) != s {
			return nil, fmt.Errorf("%s: cat-file: %w", path, err)
		}

		// sanity check that should probably be removed at some point
		showData, err := Git(filepath.Join(dir, submoduleDir), "show", ref+":"+path)
		if err != nil || !bytes.Equal(data.Stdout, showData.Stdout) {
			return nil, fmt.Errorf("%s: show: %w", path, err)
		}

		sha512, err = crypto.SHA512(data.Stdout)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	} else {
		return nil, fmt.Errorf("%s: %s: unsupported object type", path, typ)
	}

	return &Object{
		Mode:         mode,
		SHA1:         sha1,
		Type:         typ,
		Path:         path,
		Size:         s,
		SubmoduleDir: submoduleDir,
		SHA512:       sha512,
	}, nil
}
