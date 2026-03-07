package crypto

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/illikainen/attest/utils"
)

type PrivateKey struct {
	SignatureAlgorithm [2]byte
	KeyID              [8]byte
	Key                [ed25519.PrivateKeySize]byte
}

func ReadPrivateKey(name string) (*PrivateKey, error) {
	path, err := utils.ExpandPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, err
	}

	key, err := DecodePrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("%s: can't read private key: %w", name, err)
	}
	return key, nil
}

func DecodePrivateKey(data []byte) (*PrivateKey, error) {
	key := &PrivateKey{}
	raw, err := decodeBase64(data, len(key.SignatureAlgorithm)+len(key.KeyID)+len(key.Key))
	if err != nil {
		return nil, err
	}

	copy(key.SignatureAlgorithm[:], raw[0:2])
	copy(key.KeyID[:], raw[2:10])
	copy(key.Key[:], raw[10:])

	if !bytes.Equal(key.SignatureAlgorithm[:], []byte{'E', 'd'}) {
		return nil, fmt.Errorf("invalid signature algorithm: %s", key.SignatureAlgorithm)
	}

	return key, nil
}

func (k *PrivateKey) Sign(msg []byte) (*Signature, error) {
	key := ed25519.PrivateKey(k.Key[:])

	sig, err := key.Sign(rand.Reader, msg, crypto.Hash(0))
	if err != nil {
		return nil, err
	}

	s := &Signature{
		SignatureAlgorithm: k.SignatureAlgorithm,
		KeyID:              k.KeyID,
	}
	copy(s.Signature[:], sig)
	return s, nil
}

func (k *PrivateKey) SignFile(name string) (*Signature, error) {
	data, err := os.ReadFile(name) // #nosec G304
	if err != nil {
		return nil, err
	}

	return k.Sign(data)
}

func (k *PrivateKey) Write(file string) error {
	path, err := utils.ExpandPath(file)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(k.String()+"\n"), 0o600)
}

func (k *PrivateKey) String() string {
	return encodeBase64(k.SignatureAlgorithm[:], k.KeyID[:], k.Key[:])
}
