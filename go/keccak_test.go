package eip712_test

import (
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

func TestKeccak256EmptyInput(t *testing.T) {
	// keccak256("") = c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
	got := eip712.Keccak256([]byte{})
	want := "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	if eip712.ToHex(got[:]) != want {
		t.Errorf("Keccak256(\"\") = %s, want %s", eip712.ToHex(got[:]), want)
	}
}

func TestKeccak256KnownValue(t *testing.T) {
	// keccak256("hello") = 1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8
	got := eip712.Keccak256([]byte("hello"))
	want := "0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
	if eip712.ToHex(got[:]) != want {
		t.Errorf("Keccak256(\"hello\") = %s, want %s", eip712.ToHex(got[:]), want)
	}
}

func TestKeccak256NotSHA3(t *testing.T) {
	// Confirm it is legacy Keccak, NOT NIST SHA3-256.
	// NIST SHA3-256("") = a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a
	h := eip712.Keccak256([]byte{})
	got := eip712.ToHex(h[:])
	sha3Standard := "0xa7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	if got == sha3Standard {
		t.Error("Keccak256 returned NIST SHA3-256 output; must use legacy Keccak-256")
	}
}
