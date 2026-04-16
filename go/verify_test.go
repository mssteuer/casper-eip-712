//go:build secp256k1

package eip712_test

import (
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// TestRecoverAddressRoundTrip signs a digest with a known private key and
// verifies RecoverAddress returns the correct address.
// This test uses a well-known test private key (NOT for production use).
func TestRecoverAddressRoundTrip(t *testing.T) {
	// Known test private key (Ethereum test account #0)
	// Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
	// Private key (hex): ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
	//
	// This test confirms the round-trip works; it does NOT test specific
	// digest/sig values (those are covered by the cross-language vector test).
	t.Skip("Implement after secp256k1 signing helper is available in test utils")
}

func TestVerifySignatureReturnsFalseOnMismatch(t *testing.T) {
	var digest [32]byte
	var sig [65]byte
	var wrongAddr [20]byte
	for i := range wrongAddr {
		wrongAddr[i] = 0xff
	}
	// A random sig won't match any real address - but RecoverAddress may error
	// on invalid sig bytes. VerifySignature must not panic and must return
	// either (false, nil) or (false, err).
	ok, _ := eip712.VerifySignature(digest, sig, wrongAddr)
	if ok {
		t.Error("VerifySignature returned true for mismatched address")
	}
}
