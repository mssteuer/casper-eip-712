//go:build secp256k1

package eip712

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// RecoverAddress recovers the 20-byte Ethereum address from digest and a
// 65-byte secp256k1 signature (r || s || v, where v is 0 or 1).
func RecoverAddress(digest [32]byte, signature [65]byte) ([20]byte, error) {
	// Ethereum signature format: r (32) || s (32) || v (1)
	// v must be 0 or 1 (not 27/28)
	v := signature[64]
	if v != 0 && v != 1 {
		return [20]byte{}, fmt.Errorf("eip712: invalid recovery id %d (must be 0 or 1)", v)
	}

	// secp256k1 library uses [v, r, s] order internally for recovery
	// Build compact sig: [v+27, r..., s...] - some libs expect 27/28
	// decred/secp256k1 RecoverCompact expects [v, r, s] with v as 0/1
	// Build the 65-byte compact sig in [v || r || s] format expected by RecoverCompact
	var compactSig [65]byte
	compactSig[0] = v + 27                    // RecoverCompact expects 27 or 28
	copy(compactSig[1:33], signature[:32])    // r
	copy(compactSig[33:65], signature[32:64]) // s

	pubKey, _, err := ecdsa.RecoverCompact(compactSig[:], digest[:])
	if err != nil {
		return [20]byte{}, fmt.Errorf("eip712: signature recovery failed: %w", err)
	}

	return pubKeyToAddress(pubKey), nil
}

// VerifySignature verifies that signature over digest matches expectedAddress.
// Returns false on address mismatch; returns error only on malformed input.
func VerifySignature(digest [32]byte, signature [65]byte, expectedAddress [20]byte) (bool, error) {
	recovered, err := RecoverAddress(digest, signature)
	if err != nil {
		return false, err
	}
	return recovered == expectedAddress, nil
}

// RecoverTypedDataSigner hashes typed data and recovers the signer's address.
func RecoverTypedDataSigner(
	domain EIP712Domain,
	types TypeDefinitions,
	primaryType string,
	message map[string]interface{},
	signature [65]byte,
	opts *TypedDataOptions,
) ([20]byte, error) {
	digest, err := HashTypedData(domain, types, primaryType, message, opts)
	if err != nil {
		return [20]byte{}, err
	}
	return RecoverAddress(digest, signature)
}

// pubKeyToAddress converts a secp256k1 public key to an Ethereum address.
// Ethereum address = last 20 bytes of keccak256(uncompressed_pub_key[1:])
func pubKeyToAddress(pub *secp256k1.PublicKey) [20]byte {
	// Uncompressed public key: 65 bytes, first byte is 0x04
	uncompressed := pub.SerializeUncompressed()
	// Hash bytes [1:] (skip the 0x04 prefix)
	hash := Keccak256(uncompressed[1:])
	var addr [20]byte
	copy(addr[:], hash[12:]) // last 20 bytes of the 32-byte hash
	return addr
}
