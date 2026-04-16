package prebuilt

import (
	"math/big"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// ApprovalTypes is the canonical EIP-712 type definition for the Approval message.
var ApprovalTypes = eip712.TypeDefinitions{
	"Approval": {
		{Name: "owner", Type: "address"},
		{Name: "spender", Type: "address"},
		{Name: "value", Type: "uint256"},
	},
}

// ApprovalMessage is a strongly-typed Approval message.
type ApprovalMessage struct {
	Owner   eip712.Address
	Spender eip712.Address
	Value   *big.Int
}

// ToMap converts ApprovalMessage to the map form required by HashTypedData.
func (a ApprovalMessage) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"owner":   a.Owner,
		"spender": a.Spender,
		"value":   a.Value,
	}
}
