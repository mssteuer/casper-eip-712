//go:build secp256k1

package eip712_test

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// TestRecoverAddressAcceptsLegacyV exercises the normalization of the recovery
// id byte. Rust (src/verify.rs) and JS (js/src/verify.ts) both accept v in
// {0, 1, 27, 28}; Go must do the same to stay wire-compatible with signatures
// produced by Ethereum wallets (MetaMask, ethers.js, hardware wallets), which
// commonly emit v=27 or v=28.
func TestRecoverAddressAcceptsLegacyV(t *testing.T) {
	// Deterministic test private key — NOT for production use.
	// Derived from Ethereum test account #0.
	privKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		t.Fatalf("decode priv key: %v", err)
	}
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)

	var digest [32]byte
	for i := range digest {
		digest[i] = byte(i)
	}

	// SignCompact returns <1-byte recovery code (27 or 28 + compression bit)><r><s>.
	// We request an uncompressed recovery code so the byte is exactly 27 or 28.
	compact := ecdsa.SignCompact(privKey, digest[:], false)
	if len(compact) != 65 {
		t.Fatalf("SignCompact returned %d bytes, want 65", len(compact))
	}
	recoveryCode := compact[0]
	if recoveryCode != 27 && recoveryCode != 28 {
		t.Fatalf("unexpected recovery code %d", recoveryCode)
	}
	rawV := recoveryCode - 27

	// Ethereum-style sig layout: r (32) || s (32) || v (1).
	var sigBase [65]byte
	copy(sigBase[:32], compact[1:33])
	copy(sigBase[32:64], compact[33:65])

	// Derive the expected address directly from the public key so the test
	// doesn't depend on RecoverAddress being correct.
	pubUncompressed := privKey.PubKey().SerializeUncompressed()
	addrHash := eip712.Keccak256(pubUncompressed[1:])
	var expected [20]byte
	copy(expected[:], addrHash[12:])

	cases := []struct {
		name string
		v    byte
	}{
		{"raw v=0 or v=1", rawV},
		{"legacy v=27 or v=28", rawV + 27},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sig := sigBase
			sig[64] = tc.v
			got, err := eip712.RecoverAddress(digest, sig)
			if err != nil {
				t.Fatalf("RecoverAddress(v=%d): %v", tc.v, err)
			}
			if got != expected {
				t.Errorf("RecoverAddress(v=%d) = %x, want %x", tc.v, got, expected)
			}
		})
	}
}

// TestRecoverAddressRejectsInvalidV confirms the normalization only accepts the
// four canonical values {0, 1, 27, 28}; anything else (e.g. 2, 26, 29, 35+ from
// EIP-155 transaction-level encoding) must be rejected at the EIP-712 layer.
func TestRecoverAddressRejectsInvalidV(t *testing.T) {
	var digest [32]byte
	var sig [65]byte
	// r and s need to be non-zero to reach the v check, but their value is
	// irrelevant — we expect to fail before RecoverCompact is called.
	sig[0] = 1
	sig[32] = 1

	for _, v := range []byte{2, 3, 26, 29, 30, 35, 36, 255} {
		sig[64] = v
		if _, err := eip712.RecoverAddress(digest, sig); err == nil {
			t.Errorf("RecoverAddress accepted invalid v=%d", v)
		}
	}
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
