package eip712

import (
	"fmt"
	"math/big"
	"strings"
)

// TypedField is a single named field in an EIP-712 type definition.
type TypedField struct {
	Name string
	Type string
}

// TypeDefinitions maps type names to their ordered field lists.
//
// Example:
//
//	TypeDefinitions{
//	    "Permit": {
//	        {Name: "owner",   Type: "address"},
//	        {Name: "spender", Type: "address"},
//	        {Name: "value",   Type: "uint256"},
//	    },
//	}
type TypeDefinitions map[string][]TypedField

// TypedDataOptions carries optional overrides for domain handling.
type TypedDataOptions struct {
	// DomainTypes, when non-nil, overrides auto-inference of domain fields.
	// Use CasperDomainTypes for Casper-native domains.
	DomainTypes []TypedField
}

// EIP712Domain holds domain separator fields.
// All fields are optional pointers; only non-nil fields are included in the
// domain type string. This mirrors the JS behavior of checking key presence.
type EIP712Domain struct {
	// Standard Ethereum fields
	Name              *string
	Version           *string
	ChainID           *big.Int
	VerifyingContract *Address
	Salt              *[32]byte

	// Casper-native fields
	ChainName           *string
	ContractPackageHash *[32]byte
}

// Address is a typed address value for EIP-712 encoding.
// It holds either a 20-byte Ethereum address or a 33-byte Casper address.
//
// Encoding:
//   - 20-byte (Eth):    left-padded with 12 zero bytes -> 32-byte slot
//   - 33-byte (Casper): keccak256(all 33 bytes)         -> 32-byte slot
type Address struct {
	raw []byte // invariant: len(raw) == 20 or 33
}

// NewEthAddress constructs an Address from a 20-byte Ethereum address.
func NewEthAddress(b [20]byte) Address {
	raw := make([]byte, 20)
	copy(raw, b[:])
	return Address{raw: raw}
}

// NewCasperAddress constructs an Address from a 33-byte Casper address
// (1-byte type prefix + 32-byte hash).
func NewCasperAddress(b [33]byte) Address {
	raw := make([]byte, 33)
	copy(raw, b[:])
	return Address{raw: raw}
}

// NewAddressFromHex parses a 0x-prefixed hex string into an Address.
// Accepts 20-byte (40 hex chars) and 33-byte (66 hex chars) values.
func NewAddressFromHex(s string) (Address, error) {
	b, err := FromHex(s)
	if err != nil {
		return Address{}, err
	}
	switch len(b) {
	case 20:
		var arr [20]byte
		copy(arr[:], b)
		return NewEthAddress(arr), nil
	case 33:
		var arr [33]byte
		copy(arr[:], b)
		return NewCasperAddress(arr), nil
	default:
		return Address{}, fmt.Errorf("address must be 20 or 33 bytes, got %d", len(b))
	}
}

// MustAddressFromHex is like NewAddressFromHex but panics on error.
// Intended for use in tests and package-level variables.
func MustAddressFromHex(s string) Address {
	a, err := NewAddressFromHex(s)
	if err != nil {
		panic(err)
	}
	return a
}

// Bytes returns the raw address bytes (20 or 33 bytes).
func (a Address) Bytes() []byte {
	out := make([]byte, len(a.raw))
	copy(out, a.raw)
	return out
}

// IsEth reports whether this is a 20-byte Ethereum address.
func (a Address) IsEth() bool { return len(a.raw) == 20 }

// IsCasper reports whether this is a 33-byte Casper address.
func (a Address) IsCasper() bool { return len(a.raw) == 33 }

// Hex returns the 0x-prefixed lowercase hex representation.
func (a Address) Hex() string { return "0x" + strings.ToLower(fmt.Sprintf("%x", a.raw)) }
