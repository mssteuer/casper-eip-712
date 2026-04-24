// Package prebuilt provides strongly-typed EIP-712 message structs and their
// corresponding TypeDefinitions for common Casper token operations.
package prebuilt

import (
	"math/big"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// PermitTypes is the canonical EIP-712 type definition for the Permit message.
var PermitTypes = eip712.TypeDefinitions{
	"Permit": {
		{Name: "owner", Type: "address"},
		{Name: "spender", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "deadline", Type: "uint256"},
	},
}

// PermitMessage is a strongly-typed Permit message.
type PermitMessage struct {
	Owner    eip712.Address
	Spender  eip712.Address
	Value    *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

// ToMap converts PermitMessage to the map form required by HashTypedData.
func (p PermitMessage) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"owner":    p.Owner,
		"spender":  p.Spender,
		"value":    p.Value,
		"nonce":    p.Nonce,
		"deadline": p.Deadline,
	}
}
