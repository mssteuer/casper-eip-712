# casper-eip-712

`casper-eip-712` is a `no_std`-compatible Rust crate for EIP-712 typed data hashing and domain separation on Casper. It provides reusable encoding helpers, domain construction, typed-struct hashing, and an optional `verify` feature for Ethereum-style secp256k1 signer recovery.

## Features

- `no_std` + `alloc`
- EIP-712-compatible type hashing and `\x19\x01` typed data digest construction
- Flexible `DomainBuilder` with standard EVM fields and custom Casper-native fields
- Prebuilt `Permit`, `Approval`, and `Transfer` structs
- Optional `verify` feature for signer recovery and verification

## Quick start

```rust
use casper_eip_712::prelude::*;

let domain = DomainBuilder::new()
    .name("MyToken")
    .version("1")
    .chain_id(1)
    .verifying_contract([0x11; 20])
    .build();

let permit = Permit {
    owner: [0x22; 20],
    spender: [0x33; 20],
    value: [0u8; 32],
    nonce: [0u8; 32],
    deadline: [0u8; 32],
};

let digest = hash_typed_data(&domain, &permit);
assert_eq!(digest.len(), 32);
```

## Custom struct example

```rust
use alloc::vec::Vec;
use casper_eip_712::prelude::*;

struct Attestation {
    subject: [u8; 20],
    claim_hash: [u8; 32],
}

impl Eip712Struct for Attestation {
    fn type_string() -> &'static str {
        "Attestation(address subject,bytes32 claim_hash)"
    }

    fn type_hash() -> [u8; 32] {
        keccak256(Self::type_string().as_bytes())
    }

    fn encode_data(&self) -> Vec<u8> {
        let mut out = Vec::with_capacity(64);
        out.extend_from_slice(&encode_address(self.subject));
        out.extend_from_slice(&encode_bytes32(self.claim_hash));
        out
    }
}
```

## Casper-native domain example

```rust
use casper_eip_712::prelude::*;

let domain = DomainBuilder::new()
    .name("Bridge")
    .version("1")
    .custom_field("chain_name", DomainFieldValue::String("casper-test".into()))
    .custom_field("contract_package_hash", DomainFieldValue::Bytes32([0x99; 32]))
    .build();
```

## Feature flags

- default: minimal hashing/encoding support
- `verify`: enables secp256k1 signer recovery via `k256`

## no_std

The crate is designed for `#![no_std]` environments with `alloc` available, making it suitable for WASM-based contract targets.

## Spec

- EIP-712: <https://eips.ethereum.org/EIPS/eip-712>
