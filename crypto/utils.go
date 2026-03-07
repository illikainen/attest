package crypto

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"
)

func decodeBase64(data []byte, size int) ([]byte, error) {
	data = bytes.Trim(data, " \t\r\n")

	expectedLen := base64.StdEncoding.EncodedLen(size)
	if len(data) != expectedLen {
		return nil, fmt.Errorf("invalid length: %d != %d", len(data), expectedLen)
	}

	result, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func encodeBase64(data ...[]byte) string {
	b := append([]byte{}, data[0]...)
	for i := 1; i < len(data); i++ {
		b = append(b, data[i]...)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// The untrusted comment line is needed for compatibility with signify.
func stripComment(data []byte) ([]byte, error) {
	data = bytes.Trim(data, " \t\r\n")
	lines := bytes.SplitN(data, []byte("\n"), 2)

	if len(lines) == 0 {
		return nil, fmt.Errorf("malformed signature file (%d bytes)", len(data))
	} else if len(lines) == 1 {
		return data, nil
	}

	if ok, err := regexp.Match(`^untrusted comment: [a-zA-Z0-9,.@()/ -]+$`, lines[0]); err != nil || !ok {
		return nil, fmt.Errorf("malformed signature comment: %w", err)
	}
	return lines[1], nil
}
