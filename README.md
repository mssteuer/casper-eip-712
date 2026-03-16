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

## Getting Started

### Prerequisites

- Rust (nightly recommended — see `rust-toolchain.toml`)
- Node.js 18+ (for the TypeScript package and demo)
- npm

### Build the Rust crate

```bash
cargo build
cargo test
```

### Build the TypeScript companion

```bash
cd js
npm install
npm run build
npm test
```

### Run the demo

The demo showcases a CEP-18 token with gasless permit/approve using EIP-712 signatures. It runs locally without a Casper node.

```bash
# 1. Build the TypeScript package first (the demo depends on it)
cd js && npm install && npm run build && cd ..

# 2. Install and run the demo
cd examples/permit-token/demo
npm install
npx tsx demo.ts
```

### Run the Rust integration tests (Odra)

```bash
cd examples/permit-token
cargo odra test
```

> **Note:** `casper-client` is pinned to 5.0.0 in the lock file. Version 5.0.1 introduced
> breaking API changes that are incompatible with Odra 2.5.0. If you regenerate the lock file,
> run `cargo update casper-client --precise 5.0.0` to re-pin.

## Repository Structure

```
src/                    — Core Rust crate (no_std, EIP-712 encoding + hashing)
js/                     — TypeScript companion package (@casper-ecosystem/casper-eip-712)
  src/                  — TypeScript source
  dist/                 — Built output (run npm run build)
examples/
  permit-token/         — Demo contract: CEP-18 with permit/approve pattern
    src/                — Odra smart contract (Rust)
    tests/              — Rust integration tests
    demo/               — Standalone TypeScript demo
scripts/                — Test vector generation
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

## Demo: Permit Token

See [`examples/permit-token/`](./examples/permit-token/) for a complete working
example — a CEP-18 token with gasless permit/approve using EIP-712 signatures,
supporting both EVM-compatible and Casper-native domain separators. Includes a
standalone TypeScript demo you can run without a Casper node.

## Feature flags

- default: minimal hashing/encoding support
- `verify`: enables secp256k1 signer recovery via `k256`

## no_std

The crate is designed for `#![no_std]` environments with `alloc` available, making it suitable for WASM-based contract targets.

## Development notes

- Generate test vectors with `npm --prefix scripts run generate` (the generator is TypeScript and is executed with `tsx`).
- `Cargo.lock` is intentionally not committed because this repository ships a library crate. Regenerate a fresh lockfile locally when auditing dependency resolution.

## Spec

- EIP-712: <https://eips.ethereum.org/EIPS/eip-712>
