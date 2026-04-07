package attest

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"os"
)

type Signature struct {
	SignatureAlgorithm [2]byte
	KeyID              [8]byte
	Signature          [ed25519.SignatureSize]byte
}

func ReadSignature(name string) (*Signature, error) {
	data, err := os.ReadFile(name) // #nosec G304
	if err != nil {
		return nil, err
	}

	data, err = stripComment(data)
	if err != nil {
		return nil, err
	}

	return DecodeSignature(data)
}

func DecodeSignature(data []byte) (*Signature, error) {
	s := &Signature{}
	raw, err := decodeBase64(data, len(s.SignatureAlgorithm)+len(s.KeyID)+len(s.Signature))
	if err != nil {
		return nil, err
	}

	copy(s.SignatureAlgorithm[:], raw[0:2])
	copy(s.KeyID[:], raw[2:10])
	copy(s.Signature[:], raw[10:])

	if !bytes.Equal(s.SignatureAlgorithm[:], []byte{'E', 'd'}) {
		return nil, fmt.Errorf("invalid signature algorithm: %v", s.SignatureAlgorithm)
	}

	return s, nil
}

func (s *Signature) Write(name string) error {
	// Needed for compatibility with signify.
	comment := "untrusted comment: verify with attest or signify"
	return os.WriteFile(name, []byte(comment+"\n"+s.String()+"\n"), 0o600)
}

func (s *Signature) String() string {
	return encodeBase64(s.SignatureAlgorithm[:], s.KeyID[:], s.Signature[:])
}
