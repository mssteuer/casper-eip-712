package eip712

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// ToHex encodes b as a 0x-prefixed lowercase hex string.
func ToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// FromHex decodes a hex string (with or without 0x prefix) into bytes.
// Returns an error on odd-length or invalid hex input.
func FromHex(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("hex string has odd length")
	}
	return hex.DecodeString(s)
}
