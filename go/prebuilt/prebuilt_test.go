package prebuilt_test

import (
	"math/big"
	"strings"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
	"github.com/casper-ecosystem/casper-eip-712/go/prebuilt"
)

func TestPermitTypesTypeString(t *testing.T) {
	ts, err := eip712.BuildCanonicalTypeString("Permit", prebuilt.PermitTypes)
	if err != nil {
		t.Fatal(err)
	}
	want := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
	if ts != want {
		t.Errorf("Permit type string:\ngot  %q\nwant %q", ts, want)
	}
}

func TestPermitMessageToMap(t *testing.T) {
	msg := prebuilt.PermitMessage{
		Owner:    eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20)),
		Spender:  eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20)),
		Value:    big.NewInt(1000),
		Nonce:    big.NewInt(0),
		Deadline: big.NewInt(9999999),
	}
	m := msg.ToMap()
	if _, ok := m["owner"]; !ok {
		t.Error("ToMap missing 'owner'")
	}
	if _, ok := m["deadline"]; !ok {
		t.Error("ToMap missing 'deadline'")
	}
}

func TestApprovalTypesTypeString(t *testing.T) {
	ts, err := eip712.BuildCanonicalTypeString("Approval", prebuilt.ApprovalTypes)
	if err != nil {
		t.Fatal(err)
	}
	want := "Approval(address owner,address spender,uint256 value)"
	if ts != want {
		t.Errorf("Approval type string:\ngot  %q\nwant %q", ts, want)
	}
}

// TestPrebuiltMessagesErrorOnNilNumericField confirms that an un-set numeric
// field (left as a nil *big.Int) causes HashStruct to fail rather than hash as
// a zero slot — matching Rust/JS behavior for missing required fields.
func TestPrebuiltMessagesErrorOnNilNumericField(t *testing.T) {
	owner := eip712.MustAddressFromHex("0x" + strings.Repeat("11", 20))
	spender := eip712.MustAddressFromHex("0x" + strings.Repeat("22", 20))

	cases := []struct {
		name        string
		primaryType string
		types       eip712.TypeDefinitions
		msg         map[string]interface{}
	}{
		{
			name:        "Permit with nil Value",
			primaryType: "Permit",
			types:       prebuilt.PermitTypes,
			msg: prebuilt.PermitMessage{
				Owner: owner, Spender: spender,
				Value: nil, Nonce: big.NewInt(0), Deadline: big.NewInt(1),
			}.ToMap(),
		},
		{
			name:        "Permit with nil Nonce",
			primaryType: "Permit",
			types:       prebuilt.PermitTypes,
			msg: prebuilt.PermitMessage{
				Owner: owner, Spender: spender,
				Value: big.NewInt(1), Nonce: nil, Deadline: big.NewInt(1),
			}.ToMap(),
		},
		{
			name:        "Approval with nil Value",
			primaryType: "Approval",
			types:       prebuilt.ApprovalTypes,
			msg: prebuilt.ApprovalMessage{
				Owner: owner, Spender: spender, Value: nil,
			}.ToMap(),
		},
		{
			name:        "Transfer with nil Value",
			primaryType: "Transfer",
			types:       prebuilt.TransferTypes,
			msg: prebuilt.TransferMessage{
				From: owner, To: spender, Value: nil,
			}.ToMap(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := eip712.HashStruct(tc.primaryType, tc.types, tc.msg); err == nil {
				t.Errorf("%s: HashStruct expected error, got nil", tc.name)
			}
		})
	}
}

func TestTransferTypesTypeString(t *testing.T) {
	ts, err := eip712.BuildCanonicalTypeString("Transfer", prebuilt.TransferTypes)
	if err != nil {
		t.Fatal(err)
	}
	want := "Transfer(address from,address to,uint256 value)"
	if ts != want {
		t.Errorf("Transfer type string:\ngot  %q\nwant %q", ts, want)
	}
}
