package attest

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func GenerateKey() (*PublicKey, *PrivateKey, error) {
	if strings.ToLower(runtime.GOOS) == "linux" {
		avail, err := os.ReadFile("/proc/sys/kernel/random/entropy_avail")
		if err != nil {
			return nil, nil, err
		}

		n, err := strconv.Atoi(strings.Trim(string(avail), " \n"))
		if err != nil {
			return nil, nil, err
		}

		if n < 256 {
			return nil, nil, fmt.Errorf("insufficient entropy (%d)", n)
		}
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	if len(pub) != ed25519.PublicKeySize || len(priv) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("invalid key lengths: pub=%d, priv=%d", len(pub), len(priv))
	}

	id, err := SHA512(pub)
	if err != nil {
		return nil, nil, err
	}

	pubKey := &PublicKey{}
	copy(pubKey.SignatureAlgorithm[:], []byte{'E', 'd'})
	copy(pubKey.KeyID[:], id[:8])
	copy(pubKey.Key[:], pub)

	privKey := &PrivateKey{}
	copy(privKey.SignatureAlgorithm[:], []byte{'E', 'd'})
	copy(privKey.KeyID[:], id[:8])
	copy(privKey.Key[:], priv)

	return pubKey, privKey, nil
}
