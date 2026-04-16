package prebuilt

import (
	"math/big"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// TransferTypes is the canonical EIP-712 type definition for the Transfer message.
var TransferTypes = eip712.TypeDefinitions{
	"Transfer": {
		{Name: "from", Type: "address"},
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
	},
}

// TransferMessage is a strongly-typed Transfer message.
type TransferMessage struct {
	From  eip712.Address
	To    eip712.Address
	Value *big.Int
}

// ToMap converts TransferMessage to the map form required by HashTypedData.
func (t TransferMessage) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"from":  t.From,
		"to":    t.To,
		"value": t.Value,
	}
}
