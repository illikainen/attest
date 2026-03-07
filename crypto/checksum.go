package crypto

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func SHA512(data []byte) (string, error) {
	size := len(data)
	cksum := sha512.New()

	if n, err := cksum.Write(data); err != nil || n != size {
		return "", fmt.Errorf("invalid sha512 write: %w (%d/%d)", err, n, size)
	}

	sum := cksum.Sum(nil)
	if len(sum) != 64 {
		return "", fmt.Errorf("invalid sha512 length")
	}

	return hex.EncodeToString(sum), nil
}
