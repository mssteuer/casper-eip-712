# casper-eip-712

`casper-eip-712` is a multi-language EIP-712 toolkit for Casper. This repository contains the `no_std`-compatible Rust core crate, plus companion TypeScript and Go packages that application developers can use to generate and verify EIP-712 typed data messages.

## Features

- `no_std` + `alloc`
- EIP-712-compatible type hashing and `\x19\x01` typed data digest construction
- Flexible `DomainBuilder` with standard EVM fields and custom Casper-native fields
- Prebuilt `Permit`, `Approval`, and `Transfer` structs
- Optional `verify` feature for signer recovery and verification
- Companion TypeScript and Go packages for application integration
- Shared cross-language vectors to keep Rust, TypeScript, and Go outputs in sync

## Packages in this repository

- **Rust (`src/`)**: core `casper-eip-712` crate for hashing, domain separation, encoding helpers, and optional secp256k1 recovery.
- **TypeScript (`js/`)**: companion package for dApps/services to generate and verify EIP-712 typed data messages.
- **Go (`go/`)**: companion package for backend/services to hash typed data and verify EIP-712 signatures.

## Choose your package

- Rust crate docs: [`README.md`](./README.md)
- TypeScript package docs: [`js/README.md`](./js/README.md)
- Go package docs: [`go/README.md`](./go/README.md)

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
go/                     — Go companion package (typed data hashing + verification helpers)
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
    subject: Address,
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

## Future Use Cases

### Gasless Transactions and Relayer Services

EIP-712 signatures enable a "gasless" user experience where token owners authorize actions by signing a typed message off-chain — without submitting a transaction or paying gas. A relayer or counterparty submits the signed authorization on-chain on their behalf.

This pattern, widely adopted on Ethereum as ERC-2612 (used by Uniswap, Aave, OpenSea, and others), becomes possible on Casper through this crate:

1. A user signs an EIP-712 `Permit` message with their private key (off-chain, zero cost)
2. A relayer or dApp backend submits the permit to the Casper contract, paying the deploy cost
3. The contract verifies the signature, recovers the signer, and executes the authorized action

This unlocks several powerful patterns:
- **User onboarding without CSPR** — new users can interact with dApps before acquiring native tokens for gas
- **Meta-transactions** — dApp operators subsidize gas costs to reduce friction
- **Batch authorization** — a relayer collects multiple signed permits and submits them in a single deploy

### Agentic Transactions via x402

The [x402 protocol](https://www.x402.org/) enables AI agents to pay for API access and services using HTTP 402 payment flows with stablecoins. EIP-712 typed data signing is the natural authorization layer for agent-initiated transactions on Casper:

- An AI agent signs a structured, domain-separated payment authorization
- A facilitator or smart contract verifies the signature and executes the payment
- The typed data schema makes the authorization human-readable and auditable — critical when autonomous agents are moving value

As x402 expands to Casper, `casper-eip-712` provides the cryptographic foundation for agents to sign verifiable, scoped authorizations that smart contracts can trust.

### Cross-Chain Bridge and Multi-Chain Domain Separation

In multi-chain environments — bridges, cross-chain messaging protocols, and multi-deployment systems — domain separation is critical for security. Without it, a signature intended for one deployment or chain can be replayed against another, potentially draining funds or corrupting state.

EIP-712 domain separators solve this by binding every signature to a specific:
- **Contract identity** (address or package hash)
- **Chain** (chain ID or chain name)
- **Version** (protocol version for upgrade safety)

This crate's `DomainBuilder` supports both standard EVM fields (`chainId`, `verifyingContract`) and Casper-native fields (`chain_name`, `contract_package_hash`), making it suitable for hybrid environments where Casper contracts verify attestations originating from EVM chains — or where multiple Casper deployments (testnet, mainnet, staging) must be cryptographically isolated from each other.

## Feature flags

- default: minimal hashing/encoding support
- `verify`: enables secp256k1 signer recovery via `k256`

## no_std

The crate is designed for `#![no_std]` environments with `alloc` available, making it suitable for WASM-based contract targets.

## Development notes

- Generate test vectors with `npm --prefix scripts run generate` (the generator is TypeScript and is executed with `tsx`).
- `Cargo.lock` is committed to pin `casper-client` to 5.0.0 (see note above about Odra compatibility). If you regenerate it, re-pin with `cargo update casper-client --precise 5.0.0`.

## Spec

- EIP-712: <https://eips.ethereum.org/EIPS/eip-712>
