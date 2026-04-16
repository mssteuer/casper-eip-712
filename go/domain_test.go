package eip712_test

import (
	"math/big"
	"strings"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

func mustBytes32(hex string) [32]byte {
	b, err := eip712.FromHex(hex)
	if err != nil || len(b) != 32 {
		panic("invalid bytes32 hex: " + hex)
	}
	var out [32]byte
	copy(out[:], b)
	return out
}

func strPtr(s string) *string { return &s }
func bigPtr(i int64) *big.Int { return big.NewInt(i) }

func TestBuildDomainTypeStringCasper(t *testing.T) {
	domain := eip712.BuildDomain("MyToken", "1", "casper-test",
		mustBytes32("0x7777777777777777777777777777777777777777777777777777777777777777"))
	got := eip712.BuildDomainTypeString(domain,
		&eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes})
	want := "EIP712Domain(string name,string version,string chain_name,bytes32 contract_package_hash)"
	if got != want {
		t.Errorf("BuildDomainTypeString:\ngot  %q\nwant %q", got, want)
	}
}

func TestBuildDomainTypeStringStandardAutoInfer(t *testing.T) {
	chainID := big.NewInt(1)
	addr := eip712.MustAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	name := "MyToken"
	ver := "1"
	domain := eip712.EIP712Domain{
		Name:              &name,
		Version:           &ver,
		ChainID:           chainID,
		VerifyingContract: &addr,
	}
	got := eip712.BuildDomainTypeString(domain, nil)
	want := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	if got != want {
		t.Errorf("BuildDomainTypeString auto-infer:\ngot  %q\nwant %q", got, want)
	}
}

func TestHashDomainSeparatorPartialDomain(t *testing.T) {
	// name + version only
	name := "TestDomain"
	ver := "1"
	domain := eip712.EIP712Domain{Name: &name, Version: &ver}
	sep, err := eip712.HashDomainSeparator(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Must be a non-zero 32-byte hash
	var zero [32]byte
	if sep == zero {
		t.Error("HashDomainSeparator returned all-zero hash")
	}
}

func TestHashDomainSeparatorDeterministic(t *testing.T) {
	domain := eip712.BuildDomain("A", "1", "casper", mustBytes32("0x"+strings.Repeat("aa", 32)))
	opts := &eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes}
	a, err := eip712.HashDomainSeparator(domain, opts)
	if err != nil {
		t.Fatal(err)
	}
	b, err := eip712.HashDomainSeparator(domain, opts)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Error("HashDomainSeparator not deterministic")
	}
}
