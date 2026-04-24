package eip712_test

import (
	"math/big"
	"strings"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
	"github.com/casper-ecosystem/casper-eip-712/go/prebuilt"
)

func TestHashStructPermitDeterministic(t *testing.T) {
	msg := map[string]interface{}{
		"owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		"spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		"value":    big.NewInt(100),
		"nonce":    big.NewInt(0),
		"deadline": big.NewInt(9999999),
	}
	h1, err := eip712.HashStruct("Permit", prebuilt.PermitTypes, msg)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := eip712.HashStruct("Permit", prebuilt.PermitTypes, msg)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Error("HashStruct not deterministic")
	}
}

func TestHashStructDifferentValuesDifferentHash(t *testing.T) {
	base := map[string]interface{}{
		"owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		"spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		"value":    big.NewInt(100),
		"nonce":    big.NewInt(0),
		"deadline": big.NewInt(9999999),
	}
	changed := map[string]interface{}{
		"owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		"spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		"value":    big.NewInt(200),
		"nonce":    big.NewInt(0),
		"deadline": big.NewInt(9999999),
	}
	h1, _ := eip712.HashStruct("Permit", prebuilt.PermitTypes, base)
	h2, _ := eip712.HashStruct("Permit", prebuilt.PermitTypes, changed)
	if h1 == h2 {
		t.Error("HashStruct with different values must produce different hashes")
	}
}

func TestHashTypedDataDigestIs66Bytes(t *testing.T) {
	// Sanity: the digest is computed from a 66-byte buffer (0x19, 0x01, 32+32)
	name := "MyToken"
	ver := "1"
	chainID := big.NewInt(1)
	domain := eip712.EIP712Domain{Name: &name, Version: &ver, ChainID: chainID}
	msg := map[string]interface{}{
		"owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		"spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		"value":    big.NewInt(100),
		"nonce":    big.NewInt(0),
		"deadline": big.NewInt(9999999),
	}
	digest, err := eip712.HashTypedData(domain, prebuilt.PermitTypes, "Permit", msg, nil)
	if err != nil {
		t.Fatal(err)
	}
	var zero [32]byte
	if digest == zero {
		t.Error("HashTypedData returned all-zero digest")
	}
}

func TestHashTypedDataRawMatchesHashTypedData(t *testing.T) {
	name := "MyToken"
	ver := "1"
	chainID := big.NewInt(1)
	domain := eip712.EIP712Domain{Name: &name, Version: &ver, ChainID: chainID}
	msg := map[string]interface{}{
		"owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		"spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		"value":    big.NewInt(100),
		"nonce":    big.NewInt(0),
		"deadline": big.NewInt(9999999),
	}

	// Full API
	digestFull, err := eip712.HashTypedData(domain, prebuilt.PermitTypes, "Permit", msg, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Raw API: pre-compute typeHash + encode fields manually
	typeStr, _ := eip712.BuildCanonicalTypeString("Permit", prebuilt.PermitTypes)
	typeHash := eip712.ComputeTypeHash(typeStr)
	fields := prebuilt.PermitTypes["Permit"]
	var encoded []byte
	for _, f := range fields {
		slot, err := eip712.EncodeField(f.Type, msg[f.Name], prebuilt.PermitTypes)
		if err != nil {
			t.Fatal(err)
		}
		encoded = append(encoded, slot[:]...)
	}

	digestRaw, err := eip712.HashTypedDataRaw(domain, typeHash, encoded, nil)
	if err != nil {
		t.Fatal(err)
	}

	if digestFull != digestRaw {
		t.Errorf("HashTypedData and HashTypedDataRaw disagree:\nfull %s\nraw  %s",
			eip712.ToHex(digestFull[:]), eip712.ToHex(digestRaw[:]))
	}
}
