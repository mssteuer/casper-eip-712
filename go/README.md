# casper-eip712-go

Go package implementing EIP-712 typed data hashing for the Casper Network.

Mirrors the TypeScript and Rust implementations in this repository. Produces bit-for-bit identical digests, validated against shared cross-language test vectors.

## Install

```bash
go get github.com/casper-ecosystem/casper-eip-712/go
```

## Quick start — standard EVM domain

```go
import (
    "math/big"

    eip712 "github.com/casper-ecosystem/casper-eip-712/go"
    "github.com/casper-ecosystem/casper-eip-712/go/prebuilt"
)

chainID := big.NewInt(1)
contract := eip712.MustAddressFromHex("0x1111111111111111111111111111111111111111")
name, version := "My Token", "1"

domain := eip712.EIP712Domain{
    Name:              &name,
    Version:           &version,
    ChainID:           chainID,
    VerifyingContract: &contract,
}

msg := prebuilt.PermitMessage{
    Owner:    eip712.MustAddressFromHex("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
    Spender:  eip712.MustAddressFromHex("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
    Value:    big.NewInt(1000),
    Nonce:    big.NewInt(0),
    Deadline: big.NewInt(9999999999),
}

digest, err := eip712.HashTypedData(domain, prebuilt.PermitTypes, "Permit", msg.ToMap(), nil)
```

## Casper-native domains

For Casper-specific domains (`chain_name` + `contract_package_hash` instead of `chainId` + `verifyingContract`), use `BuildDomain` and pass `CasperDomainTypes` in the options:

```go
import (
    "encoding/hex"

    eip712 "github.com/casper-ecosystem/casper-eip-712/go"
    "github.com/casper-ecosystem/casper-eip-712/go/prebuilt"
)

raw, _ := hex.DecodeString("7777777777777777777777777777777777777777777777777777777777777777")
var contractPackageHash [32]byte
copy(contractPackageHash[:], raw)

domain := eip712.BuildDomain("CasperToken", "1", "casper-test", contractPackageHash)

opts := &eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes}

digest, err := eip712.HashTypedData(domain, prebuilt.PermitTypes, "Permit", msg.ToMap(), opts)
```

## Address types

The package distinguishes between 20-byte Ethereum addresses and 33-byte Casper addresses. Both are held in the same `Address` type and encoded differently into the 32-byte EIP-712 slot:

| Type | Encoding |
|------|----------|
| Eth (20 bytes) | left-padded with 12 zero bytes |
| Casper (33 bytes) | `keccak256(1-byte prefix + 32-byte hash)` |

```go
// Ethereum address
ethAddr := eip712.MustAddressFromHex("0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC")

// Casper AccountHash (prefix 0x00)
casperAddr, err := eip712.NewAddressFromHex("0x00" + strings.Repeat("ab", 32))

// From fixed-size arrays
var eth [20]byte
var casper [33]byte
a1 := eip712.NewEthAddress(eth)
a2 := eip712.NewCasperAddress(casper)
```

## EIP712Domain fields are optional pointers

Fields are `*string`, `*big.Int`, etc. rather than plain values. Only non-nil fields are included in the domain type string, mirroring how the JS implementation infers the schema from key presence:

```go
name, version := "MyToken", "1"
chainID := big.NewInt(1)

// Only name + version — no chainId, no contract
domain := eip712.EIP712Domain{
    Name:    &name,
    Version: &version,
}

// Standard EVM domain
domain := eip712.EIP712Domain{
    Name:              &name,
    Version:           &version,
    ChainID:           chainID,
    VerifyingContract: &addr,
}
```

## Low-level API

For callers that pre-compute the type hash and encode fields manually (useful for tight cross-language coordination or custom struct types):

```go
typeStr := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
typeHash := eip712.ComputeTypeHash(typeStr)

// Encode each field into 32-byte slots
encoded := make([]byte, 0, 32*5)
for _, slot := range []([32]byte){
    eip712.EncodeAddress(ownerAddr),
    eip712.EncodeAddress(spenderAddr),
    must(eip712.EncodeUint256(value)),
    must(eip712.EncodeUint256(nonce)),
    must(eip712.EncodeUint256(deadline)),
} {
    encoded = append(encoded, slot[:]...)
}

digest, err := eip712.HashTypedDataRaw(domain, typeHash, encoded, opts)
```

## Prebuilt message types

The `prebuilt` sub-package provides strongly-typed structs for the three common token operations. Each has a `ToMap()` method that produces the `map[string]interface{}` form expected by `HashTypedData`.

```go
import "github.com/casper-ecosystem/casper-eip-712/go/prebuilt"

// prebuilt.PermitTypes   — TypeDefinitions for Permit
// prebuilt.ApprovalTypes — TypeDefinitions for Approval
// prebuilt.TransferTypes — TypeDefinitions for Transfer

permit := prebuilt.PermitMessage{Owner: ..., Spender: ..., Value: ..., Nonce: ..., Deadline: ...}
approval := prebuilt.ApprovalMessage{Owner: ..., Spender: ..., Value: ...}
transfer := prebuilt.TransferMessage{From: ..., To: ..., Value: ...}
```

## Signature verification

### EVM signer verification (`secp256k1` build tag)

Signature recovery is an optional feature gated behind the `secp256k1` build tag (analogous to `--features verify` in the Rust crate). It is not compiled by default, keeping the base module free of the `secp256k1` dependency.

```bash
# Build without verify (default)
go build ./...

# Build with signature verification enabled
go build -tags secp256k1 ./...
```

```go
//go:build secp256k1

// Recover the signer's 20-byte address from a 65-byte signature (r || s || v).
addr, err := eip712.RecoverAddress(digest, signature)

// Verify the signature matches an expected address.
ok, err := eip712.VerifySignature(digest, signature, expectedAddr)

// End-to-end: hash typed data and recover signer in one call.
signer, err := eip712.RecoverTypedDataSigner(domain, types, "Permit", msg, signature, opts)
```

Signature format: 65 bytes as `r (32) || s (32) || v (1)` where `v` is `0` or `1` (not `27`/`28`).

### Casper public key verification

If you need to verify an EIP-712 digest/signature with a Casper public key, use `casper-go-sdk`:

```go
import (
    "github.com/make-software/casper-go-sdk/v2/types/keypair"
)

func VerifyEIP712SignatureWithCasperKey(digest [32]byte, sig [65]byte, publicKey string) (bool, error) {
    pk, err := keypair.NewPublicKey(publicKey)
    if err != nil {
        return false, err
    }
    err = pk.VerifySignature(digest[:], sig[:])
    if err != nil {
        return false, err
    }
    return true, nil
}
```

`publicKey` must be a Casper-formatted public key string (for example `01...` for Ed25519 or `02...` for secp256k1).

## API reference

### Hashing

| Function | Description |
|----------|-------------|
| `HashTypedData(domain, types, primaryType, message, opts)` | Full EIP-712 digest |
| `HashTypedDataRaw(domain, typeHash, encodedStruct, opts)` | Low-level variant with pre-computed inputs |
| `HashStruct(primaryType, types, message)` | Struct hash only |
| `HashDomainSeparator(domain, opts)` | Domain separator hash |

### Encoding

| Function | Description |
|----------|-------------|
| `EncodeAddress(addr)` | 32-byte slot from `Address` |
| `EncodeUint256(value *big.Int)` | 32-byte big-endian slot |
| `EncodeUint256FromBytes(b [32]byte)` | Raw slot passthrough |
| `EncodeUint64(value uint64)` | 32-byte big-endian slot |
| `EncodeString(s)` | `keccak256(UTF-8 bytes)` |
| `EncodeBytes(b)` | `keccak256(b)` |
| `EncodeBytes32(b [32]byte)` | Identity (32-byte passthrough) |
| `EncodeBool(b)` | 0 or 1 in a 32-byte slot |
| `EncodeField(typ, value, types)` | Dispatcher for any EIP-712 type |

### Domain

| Symbol | Description |
|--------|-------------|
| `BuildDomain(name, version, chainName, contractPackageHash)` | Constructs a Casper-native domain |
| `BuildDomainTypeString(domain, opts)` | Canonical domain type string |
| `CasperDomainTypes` | `[]TypedField` constant for Casper-native domains |

### Type strings

| Function | Description |
|----------|-------------|
| `BuildTypeString(typeName, fields)` | Simple type string |
| `BuildCanonicalTypeString(primaryType, types)` | With sorted transitive dependencies |
| `ComputeTypeHash(typeString)` | `keccak256(typeString)` |

### Address constructors

| Function | Description |
|----------|-------------|
| `NewEthAddress(b [20]byte)` | From a 20-byte array |
| `NewCasperAddress(b [33]byte)` | From a 33-byte array |
| `NewAddressFromHex(s string)` | From a 0x-prefixed hex string (20 or 33 bytes) |
| `MustAddressFromHex(s string)` | Like above, panics on error |

### Utilities

| Function | Description |
|----------|-------------|
| `Keccak256(data []byte)` | `[32]byte` legacy Keccak-256 hash |
| `ToHex(b []byte)` | `0x`-prefixed lowercase hex string |
| `FromHex(s string)` | Hex string → `[]byte` |

### Verification (build tag `secp256k1`)

| Function | Description |
|----------|-------------|
| `RecoverAddress(digest, signature)` | Recover 20-byte signer address |
| `VerifySignature(digest, signature, expected)` | Verify signature against address |
| `RecoverTypedDataSigner(domain, types, primaryType, message, signature, opts)` | End-to-end recovery |

## Development

```bash
# Run all tests
go test ./...

# Run with signature verification
go test -tags secp256k1 ./...

# Run cross-language vector test only
go test -run TestCrossLanguageVectors ./...

# Static analysis
go vet ./...
```

Cross-language correctness is validated by `TestCrossLanguageVectors`, which reads the shared `../tests/vectors.json` and asserts identical `domainSeparator`, `structHash`, and `digest` values against the JS and Rust reference implementations.
