package eip712_test

import (
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
	"github.com/casper-ecosystem/casper-eip-712/go/prebuilt"
)

// vectorsFile is the shared cross-language test vectors.
// Path is relative to this file (go/ directory), one level up is repo root.
const vectorsFile = "../tests/vectors.json"

type vectorsJSON struct {
	Vectors []vector `json:"vectors"`
}

type vector struct {
	Name            string                 `json:"name"`
	PrimaryType     string                 `json:"primaryType"`
	Domain          map[string]interface{} `json:"domain"`
	Message         map[string]interface{} `json:"message"`
	DomainSeparator string                 `json:"domainSeparator"`
	StructHash      string                 `json:"structHash"`
	Digest          string                 `json:"digest"`
}

func TestCrossLanguageVectors(t *testing.T) {
	data, err := os.ReadFile(vectorsFile)
	if err != nil {
		t.Fatalf("cannot read %s: %v", vectorsFile, err)
	}

	var vf vectorsJSON
	if err := json.Unmarshal(data, &vf); err != nil {
		t.Fatalf("cannot parse vectors.json: %v", err)
	}

	if len(vf.Vectors) < 6 {
		t.Fatalf("expected at least 6 vectors, got %d", len(vf.Vectors))
	}

	for _, v := range vf.Vectors {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			domain := parseDomain(t, v.Domain)
			opts := resolveOpts(v.Domain)
			types, msg := resolveTypesAndMessage(t, v.PrimaryType, v.Message)

			// Assert domain separator
			gotSep, err := eip712.HashDomainSeparator(domain, opts)
			if err != nil {
				t.Fatalf("HashDomainSeparator: %v", err)
			}
			assertHex(t, "domainSeparator", eip712.ToHex(gotSep[:]), v.DomainSeparator)

			// Assert struct hash
			gotStruct, err := eip712.HashStruct(v.PrimaryType, types, msg)
			if err != nil {
				t.Fatalf("HashStruct: %v", err)
			}
			assertHex(t, "structHash", eip712.ToHex(gotStruct[:]), v.StructHash)

			// Assert final digest
			gotDigest, err := eip712.HashTypedData(domain, types, v.PrimaryType, msg, opts)
			if err != nil {
				t.Fatalf("HashTypedData: %v", err)
			}
			assertHex(t, "digest", eip712.ToHex(gotDigest[:]), v.Digest)
		})
	}
}

func assertHex(t *testing.T, label, got, want string) {
	t.Helper()
	got = strings.ToLower(got)
	want = strings.ToLower(want)
	if got != want {
		t.Errorf("%s mismatch:\ngot  %s\nwant %s", label, got, want)
	}
}

// parseDomain converts a raw JSON domain map into an EIP712Domain struct.
func parseDomain(t *testing.T, raw map[string]interface{}) eip712.EIP712Domain {
	t.Helper()
	var d eip712.EIP712Domain

	if v, ok := raw["name"].(string); ok {
		d.Name = &v
	}
	if v, ok := raw["version"].(string); ok {
		d.Version = &v
	}
	if v, ok := raw["chainId"]; ok {
		chainID := parseJSONNumber(t, v)
		d.ChainID = chainID
	}
	if v, ok := raw["verifyingContract"].(string); ok {
		addr, err := eip712.NewAddressFromHex(v)
		if err != nil {
			t.Fatalf("verifyingContract: %v", err)
		}
		d.VerifyingContract = &addr
	}
	if v, ok := raw["salt"].(string); ok {
		b, err := eip712.FromHex(v)
		if err != nil || len(b) != 32 {
			t.Fatalf("salt: invalid bytes32 %q", v)
		}
		var arr [32]byte
		copy(arr[:], b)
		d.Salt = &arr
	}
	if v, ok := raw["chain_name"].(string); ok {
		d.ChainName = &v
	}
	if v, ok := raw["contract_package_hash"].(string); ok {
		b, err := eip712.FromHex(v)
		if err != nil || len(b) != 32 {
			t.Fatalf("contract_package_hash: invalid bytes32 %q", v)
		}
		var arr [32]byte
		copy(arr[:], b)
		d.ContractPackageHash = &arr
	}
	return d
}

// resolveOpts returns TypedDataOptions with CasperDomainTypes if the domain
// uses Casper-native fields; nil otherwise (standard auto-inference).
func resolveOpts(raw map[string]interface{}) *eip712.TypedDataOptions {
	_, hasChainName := raw["chain_name"]
	_, hasCPH := raw["contract_package_hash"]
	if hasChainName || hasCPH {
		return &eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes}
	}
	return nil
}

// resolveTypesAndMessage converts the JSON message into the correct Go types
// based on primaryType, and returns the TypeDefinitions for that type.
func resolveTypesAndMessage(t *testing.T, primaryType string, raw map[string]interface{}) (eip712.TypeDefinitions, map[string]interface{}) {
	t.Helper()
	switch primaryType {
	case "Permit":
		msg := map[string]interface{}{
			"owner":    parseAddress(t, raw["owner"]),
			"spender":  parseAddress(t, raw["spender"]),
			"value":    parseUint256(t, raw["value"]),
			"nonce":    parseUint256(t, raw["nonce"]),
			"deadline": parseUint256(t, raw["deadline"]),
		}
		return prebuilt.PermitTypes, msg
	case "Approval":
		msg := map[string]interface{}{
			"owner":   parseAddress(t, raw["owner"]),
			"spender": parseAddress(t, raw["spender"]),
			"value":   parseUint256(t, raw["value"]),
		}
		return prebuilt.ApprovalTypes, msg
	case "Transfer":
		msg := map[string]interface{}{
			"from":  parseAddress(t, raw["from"]),
			"to":    parseAddress(t, raw["to"]),
			"value": parseUint256(t, raw["value"]),
		}
		return prebuilt.TransferTypes, msg
	default:
		t.Fatalf("unknown primaryType %q", primaryType)
		return nil, nil
	}
}

func parseAddress(t *testing.T, v interface{}) eip712.Address {
	t.Helper()
	s, ok := v.(string)
	if !ok {
		t.Fatalf("address value is not a string: %T", v)
	}
	addr, err := eip712.NewAddressFromHex(s)
	if err != nil {
		t.Fatalf("parseAddress(%q): %v", s, err)
	}
	return addr
}

func parseUint256(t *testing.T, v interface{}) *big.Int {
	t.Helper()
	switch val := v.(type) {
	case string:
		i := new(big.Int)
		s := strings.TrimPrefix(val, "0x")
		if _, ok := i.SetString(s, 16); !ok {
			t.Fatalf("parseUint256(%q): invalid hex", val)
		}
		return i
	case float64:
		return new(big.Int).SetInt64(int64(val))
	default:
		t.Fatalf("parseUint256: unexpected type %T", v)
		return nil
	}
}

func parseJSONNumber(t *testing.T, v interface{}) *big.Int {
	t.Helper()
	switch val := v.(type) {
	case float64:
		return new(big.Int).SetInt64(int64(val))
	case string:
		i := new(big.Int)
		s := strings.TrimPrefix(val, "0x")
		if _, ok := i.SetString(s, 16); ok {
			return i
		}
		if _, ok := i.SetString(val, 10); ok {
			return i
		}
		t.Fatalf("parseJSONNumber(%q): not a valid number", val)
		return nil
	default:
		t.Fatalf("parseJSONNumber: unexpected type %T", v)
		return nil
	}
}
