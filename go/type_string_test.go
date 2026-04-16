package eip712_test

import (
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

func TestBuildTypeStringSingle(t *testing.T) {
	fields := []eip712.TypedField{
		{Name: "owner", Type: "address"},
		{Name: "spender", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "deadline", Type: "uint256"},
	}
	got := eip712.BuildTypeString("Permit", fields)
	want := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
	if got != want {
		t.Errorf("BuildTypeString:\ngot  %q\nwant %q", got, want)
	}
}

func TestBuildTypeStringEmpty(t *testing.T) {
	got := eip712.BuildTypeString("Empty", nil)
	if got != "Empty()" {
		t.Errorf("BuildTypeString(empty fields) = %q, want \"Empty()\"", got)
	}
}

func TestBuildCanonicalTypeStringNoDeps(t *testing.T) {
	types := eip712.TypeDefinitions{
		"Permit": {
			{Name: "owner", Type: "address"},
			{Name: "value", Type: "uint256"},
		},
	}
	got, err := eip712.BuildCanonicalTypeString("Permit", types)
	if err != nil {
		t.Fatal(err)
	}
	want := "Permit(address owner,uint256 value)"
	if got != want {
		t.Errorf("BuildCanonicalTypeString:\ngot  %q\nwant %q", got, want)
	}
}

func TestBuildCanonicalTypeStringWithDeps(t *testing.T) {
	// EIP-712 spec example: Mail references Person
	types := eip712.TypeDefinitions{
		"Mail": {
			{Name: "from", Type: "Person"},
			{Name: "to", Type: "Person"},
			{Name: "contents", Type: "string"},
		},
		"Person": {
			{Name: "name", Type: "string"},
			{Name: "wallet", Type: "address"},
		},
	}
	got, err := eip712.BuildCanonicalTypeString("Mail", types)
	if err != nil {
		t.Fatal(err)
	}
	// Per EIP-712: primary type first, then deps sorted alphabetically
	want := "Mail(Person from,Person to,string contents)Person(string name,address wallet)"
	if got != want {
		t.Errorf("BuildCanonicalTypeString:\ngot  %q\nwant %q", got, want)
	}
}

func TestComputeTypeHash(t *testing.T) {
	// keccak256 of the Permit type string (known reference value)
	typeStr := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
	hash := eip712.ComputeTypeHash(typeStr)
	// This matches the value produced by the JS/Rust implementations
	want := "0x6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c9"
	got := eip712.ToHex(hash[:])
	if got != want {
		t.Errorf("ComputeTypeHash(Permit):\ngot  %s\nwant %s", got, want)
	}
}
