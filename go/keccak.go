package eip712

import "golang.org/x/crypto/sha3"

// Keccak256 computes the keccak256 hash of data.
// Uses the legacy Keccak-256 algorithm (NOT the NIST SHA3-256 standard),
// which is what Ethereum and EIP-712 specify.
func Keccak256(data []byte) [32]byte {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write(data)
	var out [32]byte
	h.Sum(out[:0])
	return out
}
