// Package eip712 implements EIP-712 typed data hashing for the Casper Network.
//
// It produces bit-for-bit identical digests to the companion TypeScript and Rust
// implementations, validated against shared cross-language test vectors.
//
// # Quick start
//
//	domain := eip712.BuildDomain("MyToken", "1", "casper-test",
//	    [32]byte{0x77, ...})
//
//	digest, err := eip712.HashTypedData(domain, prebuilt.PermitTypes, "Permit",
//	    map[string]interface{}{
//	        "owner":    eip712.MustAddressFromHex("0x" + strings.Repeat("ab", 20)),
//	        "spender":  eip712.MustAddressFromHex("0x" + strings.Repeat("cd", 20)),
//	        "value":    new(big.Int).SetInt64(100),
//	        "nonce":    new(big.Int).SetInt64(0),
//	        "deadline": new(big.Int).SetInt64(9999999999),
//	    },
//	    &eip712.TypedDataOptions{DomainTypes: eip712.CasperDomainTypes},
//	)
//
// # Casper-native domains
//
// Use [BuildDomain] and pass [CasperDomainTypes] in [TypedDataOptions] to select
// the Casper-native domain schema (chain_name + contract_package_hash) instead of
// the standard Ethereum schema (chainId + verifyingContract).
//
// # Signature verification
//
// Build with the `secp256k1` build tag to enable [RecoverAddress],
// [VerifySignature], and [RecoverTypedDataSigner].
package eip712
