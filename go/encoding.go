package eip712

import (
	"fmt"
	"math/big"
	"strings"
)

// EncodeAddress encodes an Address as a 32-byte EIP-712 slot.
//   - Eth (20-byte): left-padded with 12 zero bytes.
//   - Casper (33-byte): keccak256(all 33 bytes).
func EncodeAddress(addr Address) [32]byte {
	switch len(addr.raw) {
	case 20:
		var out [32]byte
		copy(out[12:], addr.raw)
		return out
	case 33:
		return Keccak256(addr.raw)
	default:
		panic(fmt.Sprintf("eip712: invalid address length %d", len(addr.raw)))
	}
}

// EncodeUint256 encodes a non-negative integer as a 32-byte big-endian slot.
// Returns an error if value is nil, negative, or exceeds 32 bytes.
func EncodeUint256(value *big.Int) ([32]byte, error) {
	if value == nil {
		return [32]byte{}, fmt.Errorf("eip712: uint256 value must not be nil")
	}
	if value.Sign() < 0 {
		return [32]byte{}, fmt.Errorf("eip712: uint256 value must be non-negative")
	}
	b := value.Bytes()
	if len(b) > 32 {
		return [32]byte{}, fmt.Errorf("eip712: uint256 value exceeds 32 bytes (%d bytes)", len(b))
	}
	var out [32]byte
	copy(out[32-len(b):], b)
	return out, nil
}

// EncodeUint256FromBytes encodes a raw 32-byte big-endian value as-is.
// Infallible; useful when the caller already has a pre-encoded slot.
func EncodeUint256FromBytes(b [32]byte) [32]byte {
	return b
}

// EncodeUint64 encodes a uint64 as a 32-byte big-endian slot.
func EncodeUint64(value uint64) [32]byte {
	return encodeUint64Internal(value)
}

func encodeUint64Internal(value uint64) [32]byte {
	var out [32]byte
	out[24] = byte(value >> 56)
	out[25] = byte(value >> 48)
	out[26] = byte(value >> 40)
	out[27] = byte(value >> 32)
	out[28] = byte(value >> 24)
	out[29] = byte(value >> 16)
	out[30] = byte(value >> 8)
	out[31] = byte(value)
	return out
}

// EncodeString encodes a UTF-8 string as keccak256(bytes).
func EncodeString(value string) [32]byte {
	return Keccak256([]byte(value))
}

// EncodeBytes encodes a byte slice as keccak256(bytes).
func EncodeBytes(value []byte) [32]byte {
	return Keccak256(value)
}

// EncodeBytes32 returns the 32-byte value as a slot (identity function).
func EncodeBytes32(value [32]byte) [32]byte {
	return value
}

// EncodeBool encodes a boolean as a 32-byte slot (0 = false, 1 = true).
func EncodeBool(value bool) [32]byte {
	var out [32]byte
	if value {
		out[31] = 1
	}
	return out
}

// EncodeField dispatches encoding based on the EIP-712 type name.
//
// Supported types: address, uint256, uint64, uint8-uint248 (all via uint256),
// int* (via uint256), bytes32, string, bytes, bool, and any nested struct type
// registered in types.
//
// value may be: Address, *big.Int, uint64, string (hex or decimal for integers),
// []byte, bool, [32]byte, or map[string]interface{} for nested structs.
func EncodeField(typ string, value interface{}, types TypeDefinitions) ([32]byte, error) {
	switch typ {
	case "address":
		return encodeFieldAddress(value)

	case "uint256", "uint128", "uint64", "uint32", "uint16", "uint8",
		"int256", "int128", "int64", "int32", "int16", "int8":
		return encodeFieldInt(value)

	case "bytes32":
		return encodeFieldBytes32(value)

	case "string":
		s, ok := value.(string)
		if !ok {
			return [32]byte{}, fmt.Errorf("eip712: field type string requires string value, got %T", value)
		}
		return EncodeString(s), nil

	case "bytes":
		return encodeFieldBytes(value)

	case "bool":
		b, ok := value.(bool)
		if !ok {
			return [32]byte{}, fmt.Errorf("eip712: field type bool requires bool value, got %T", value)
		}
		return EncodeBool(b), nil

	default:
		if strings.HasPrefix(typ, "uint") || strings.HasPrefix(typ, "int") {
			return encodeFieldInt(value)
		}
		if _, ok := types[typ]; ok {
			msg, ok := value.(map[string]interface{})
			if !ok {
				return [32]byte{}, fmt.Errorf("eip712: nested struct %q requires map[string]interface{}, got %T", typ, value)
			}
			return HashStruct(typ, types, msg)
		}
		return [32]byte{}, fmt.Errorf("eip712: unsupported EIP-712 type %q", typ)
	}
}

func encodeFieldAddress(value interface{}) ([32]byte, error) {
	switch v := value.(type) {
	case Address:
		return EncodeAddress(v), nil
	case string:
		addr, err := NewAddressFromHex(v)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: invalid address hex: %w", err)
		}
		return EncodeAddress(addr), nil
	default:
		return [32]byte{}, fmt.Errorf("eip712: field type address requires Address or string, got %T", value)
	}
}

func encodeFieldInt(value interface{}) ([32]byte, error) {
	switch v := value.(type) {
	case *big.Int:
		return EncodeUint256(v)
	case uint64:
		return EncodeUint256(new(big.Int).SetUint64(v))
	case int64:
		return EncodeUint256(new(big.Int).SetInt64(v))
	case int:
		return EncodeUint256(big.NewInt(int64(v)))
	case float64:
		return EncodeUint256(new(big.Int).SetInt64(int64(v)))
	case string:
		i := new(big.Int)
		s := strings.TrimPrefix(v, "0x")
		s = strings.TrimPrefix(s, "0X")
		base := 10
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			base = 16
		}
		if _, ok := i.SetString(s, base); !ok {
			return [32]byte{}, fmt.Errorf("eip712: invalid integer value %q", v)
		}
		return EncodeUint256(i)
	case [32]byte:
		return v, nil
	default:
		return [32]byte{}, fmt.Errorf("eip712: unsupported integer value type %T", value)
	}
}

func encodeFieldBytes32(value interface{}) ([32]byte, error) {
	switch v := value.(type) {
	case [32]byte:
		return EncodeBytes32(v), nil
	case string:
		b, err := FromHex(v)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: invalid bytes32 hex: %w", err)
		}
		if len(b) != 32 {
			return [32]byte{}, fmt.Errorf("eip712: bytes32 must be 32 bytes, got %d", len(b))
		}
		var arr [32]byte
		copy(arr[:], b)
		return EncodeBytes32(arr), nil
	default:
		return [32]byte{}, fmt.Errorf("eip712: field type bytes32 requires [32]byte or string, got %T", value)
	}
}

func encodeFieldBytes(value interface{}) ([32]byte, error) {
	switch v := value.(type) {
	case []byte:
		return EncodeBytes(v), nil
	case string:
		b, err := FromHex(v)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: invalid bytes hex: %w", err)
		}
		return EncodeBytes(b), nil
	default:
		return [32]byte{}, fmt.Errorf("eip712: field type bytes requires []byte or string, got %T", value)
	}
}
