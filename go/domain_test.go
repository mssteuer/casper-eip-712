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

// TestHashDomainSeparatorExplicitTypeMismatchErrors confirms that declared
// tf.Type and the domain field's actual value type must agree. Previously the
// explicit-DomainTypes path ignored tf.Type entirely and dispatched encoding
// from the name, so {Name: "chainId", Type: "bytes32"} would still encode the
// ChainID *big.Int as uint256 — silently producing an internally inconsistent
// digest (type hash says bytes32, encoded bytes are uint256).
//
// Rust cannot express this hazard (DomainFieldValue enum ties type and value
// together); JS catches it in encodeField. Go now dispatches encoding from
// tf.Type, so a mismatch errors at the encoder.
func TestHashDomainSeparatorExplicitTypeMismatchErrors(t *testing.T) {
	chainID := big.NewInt(1)
	addr := eip712.MustAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")
	name := "T"
	ver := "1"
	domain := eip712.EIP712Domain{
		Name:              &name,
		Version:           &ver,
		ChainID:           chainID,
		VerifyingContract: &addr,
	}

	cases := []struct {
		name  string
		types []eip712.TypedField
	}{
		{
			name: "chainId declared as bytes32 but ChainID is *big.Int",
			types: []eip712.TypedField{
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "bytes32"},
			},
		},
		{
			name: "name declared as uint256 but Name is string",
			types: []eip712.TypedField{
				{Name: "name", Type: "uint256"},
			},
		},
		{
			name: "verifyingContract declared as bytes32 but value is Address",
			types: []eip712.TypedField{
				{Name: "verifyingContract", Type: "bytes32"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &eip712.TypedDataOptions{DomainTypes: tc.types}
			if _, err := eip712.HashDomainSeparator(domain, opts); err == nil {
				t.Errorf("%s: HashDomainSeparator expected error, got nil", tc.name)
			}
		})
	}
}

// TestHashDomainSeparatorExplicitUint64ChainIdAccepted confirms that the new
// type-dispatch path still allows legitimate alternative encodings the name-
// dispatch path couldn't easily support — e.g. some verifiers declare chainId
// as uint64 rather than uint256. Both encode as left-padded 32-byte big-endian
// for the same numeric value, so hashing must succeed.
func TestHashDomainSeparatorExplicitUint64ChainIdAccepted(t *testing.T) {
	chainID := big.NewInt(1)
	name := "T"
	ver := "1"
	domain := eip712.EIP712Domain{Name: &name, Version: &ver, ChainID: chainID}
	opts := &eip712.TypedDataOptions{DomainTypes: []eip712.TypedField{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint64"},
	}}
	if _, err := eip712.HashDomainSeparator(domain, opts); err != nil {
		t.Errorf("HashDomainSeparator with chainId as uint64 should succeed: %v", err)
	}
}

func TestHashDomainSeparatorExplicitChainIdNilErrors(t *testing.T) {
	name := "T"
	ver := "1"
	domain := eip712.EIP712Domain{Name: &name, Version: &ver} // ChainID intentionally nil
	opts := &eip712.TypedDataOptions{DomainTypes: []eip712.TypedField{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
	}}
	if _, err := eip712.HashDomainSeparator(domain, opts); err == nil {
		t.Error("HashDomainSeparator with explicit chainId but nil ChainID expected error, got nil")
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
