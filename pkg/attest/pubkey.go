package attest

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/illikainen/attest/internal/bwrap"
	"github.com/illikainen/attest/internal/errutil"
	"github.com/illikainen/attest/internal/fsutil"
)

type PublicKey struct {
	SignatureAlgorithm [2]byte
	KeyID              [8]byte
	Key                [ed25519.PublicKeySize]byte
}

func ReadPublicKey(name string) (*PublicKey, error) {
	path, err := fsutil.ExpandPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, err
	}

	data, err = stripComment(data)
	if err != nil {
		return nil, err
	}

	return DecodePublicKey(data)
}

func DecodePublicKey(data []byte) (*PublicKey, error) {
	key := &PublicKey{}
	raw, err := decodeBase64(data, len(key.SignatureAlgorithm)+len(key.KeyID)+len(key.Key))
	if err != nil {
		return nil, err
	}

	copy(key.SignatureAlgorithm[:], raw[0:2])
	copy(key.KeyID[:], raw[2:10])
	copy(key.Key[:], raw[10:])

	if !bytes.Equal(key.SignatureAlgorithm[:], []byte{'E', 'd'}) {
		return nil, fmt.Errorf("invalid signature algorithm: %v", key.SignatureAlgorithm)
	}
	return key, nil
}

func (k *PublicKey) VerifyFile(name string, sig *Signature) error {
	data, err := os.ReadFile(name) // #nosec G304
	if err != nil {
		return err
	}
	return k.Verify(data, sig)
}

func (k *PublicKey) Verify(data []byte, sig *Signature) error {
	if !bytes.Equal(k.KeyID[:], sig.KeyID[:]) {
		return fmt.Errorf("invalid key id: %v != %v", k.KeyID, sig.KeyID)
	}

	key := ed25519.PublicKey(k.Key[:])
	if !ed25519.Verify(key, data, sig.Signature[:]) {
		return fmt.Errorf("invalid signature")
	}

	slog.Debug("signature verified with stdlib")
	return signify(data, sig, k)
}

func (k *PublicKey) Write(file string) error {
	path, err := fsutil.ExpandPath(file)
	if err != nil {
		return err
	}

	// Needed for compatibility with signify.
	comment := "untrusted comment: attest public key"
	return os.WriteFile(path, []byte(comment+"\n"+k.String()+"\n"), 0o600)
}

func (k *PublicKey) String() string {
	return encodeBase64(k.SignatureAlgorithm[:], k.KeyID[:], k.Key[:])
}

// If installed, signify is used for defense-in-depth to mitigate against
// implementation bugs in the message verification.
func signify(msg []byte, sig *Signature, pubKey *PublicKey) (err error) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer errutil.DeferRemove(tmp, &err)

	for _, path := range []string{"signify-openbsd", "signify"} {
		bin, err := exec.LookPath(path)
		if err == nil {
			pk := filepath.Join(tmp, "pubkey")
			if err := pubKey.Write(pk); err != nil {
				return err
			}

			s := filepath.Join(tmp, "sig")
			if err := sig.Write(s); err != nil {
				return err
			}

			m := filepath.Join(tmp, "msg")
			if err := os.WriteFile(m, msg, 0o600); err != nil {
				return err
			}

			p, err := bwrap.Bubblewrap(&bwrap.Options{
				Command: []string{bin, "-V", "-p", pk, "-x", s, "-m", m},
				RO:      []string{pk, s, m},
			})
			if err != nil {
				return err
			}

			slog.Debug("signify", "stdout", string(p.Stdout))
			break
		}
	}
	return nil
}
