# Permit Token Demo

A minimal CEP-18 token example that demonstrates a gasless `permit()` / approve flow on Casper using `casper-eip-712`.

It is intentionally small and educational rather than production-ready. Think of it as a reference implementation you can steal shamelessly from — the legal kind of stealing.

## What this is

This example wraps Odra's `Cep18` module and adds a `permit()` entrypoint that:

1. accepts an off-chain EIP-712 signature,
2. reconstructs the typed-data digest on-chain,
3. recovers the secp256k1 signer,
4. checks nonce + deadline,
5. installs an allowance that can later be consumed by `transfer_from()`.

The design is inspired by ERC-2612 / OpenZeppelin `ERC20Permit`, but adapted for Casper and demonstrated with both:

- **EVM-compatible domain separation** — `name`, `version`, `chainId`, `verifyingContract`
- **Casper-native domain separation** — `name`, `version`, `chain_name`, `contract_package_hash`

## File layout

```text
examples/permit-token/
├── Cargo.toml
├── Odra.toml
├── build.rs
├── rust-toolchain.toml
├── src/
│   ├── lib.rs
│   ├── permit_token.rs
│   └── bin/
│       ├── build_contract.rs
│       └── build_schema.rs
├── tests/
│   └── integration.rs
└── demo/
    ├── package.json
    ├── tsconfig.json
    └── demo.ts
```

## How the flow works

```text
owner wallet (ethers.js)
  -> signs Permit(owner, spender, value, nonce, deadline)
  -> relayer / backend submits PermitToken::permit(...)
  -> contract rebuilds EIP-712 domain + message digest
  -> contract recovers signer with casper-eip-712 verify feature
  -> contract checks deadline + nonce
  -> contract stores allowance
  -> spender calls transfer_from()
```

The owner never has to submit the transaction themselves.

## Quick start

### Rust contract + tests

```bash
cd examples/permit-token
cargo test
```

### TypeScript signing demo

```bash
cd examples/permit-token/demo
npm install
npx tsx demo.ts
```

## Contract walkthrough

### Domain construction

The contract supports two domain modes:

```rust
fn build_domain(&self, use_casper_domain: bool) -> DomainSeparator {
    let name = self.domain_name.get().unwrap_or_else(|| "PermitToken".to_string());
    let version = self.domain_version.get().unwrap_or_else(|| "1".to_string());
    if use_casper_domain {
        DomainBuilder::new()
            .name(&name)
            .version(&version)
            .custom_field("chain_name", DomainFieldValue::String("casper".into()))
            .custom_field(
                "contract_package_hash",
                DomainFieldValue::Bytes32(self.contract_package_hash_bytes()),
            )
            .build()
    } else {
        DomainBuilder::new()
            .name(&name)
            .version(&version)
            .chain_id(1_314_614_895)
            .verifying_contract(self.contract_address_bytes())
            .build()
    }
}
```

### Permit verification

At a high level, `permit()` does this:

```rust
pub fn permit(
    &mut self,
    owner_eth_address: Bytes,
    spender: Address,
    value: U256,
    nonce: U256,
    deadline: u64,
    signature: Bytes,
    use_casper_domain: bool,
) {
    // 1. deadline check
    // 2. owner address / signature length validation
    // 3. nonce check
    // 4. build domain
    // 5. hash typed Permit struct
    // 6. recover signer
    // 7. increment nonce
    // 8. store permit allowance
}
```

The `Permit` struct hashed on-chain matches the one signed off-chain with ethers.

### Allowance model used in the demo

Odra's `Cep18::approve()` is caller-based, which is perfect for normal approvals but not for gasless permits. For the demo, the contract keeps a separate:

```rust
permit_allowances: Mapping<(Address, Address), U256>
```

Then `transfer_from()` checks that mapping first and falls back to the native CEP-18 allowance path if no permit allowance is available.

That makes the demo simple and easy to reason about without modifying the Odra module itself.

## Tests included

The Rust integration suite covers:

- happy path permit
- wrong signer
- expired deadline
- replayed nonce
- `permit()` -> `transfer_from()` flow
- both EVM-compatible and Casper-native domain modes

Run them with:

```bash
cd examples/permit-token
cargo test
```

## TypeScript demo

The standalone TS demo shows the same flow from the client side:

- generate an `ethers` wallet,
- build permit data,
- sign with `wallet.signTypedData(...)`,
- verify locally with `@casper-ecosystem/casper-eip-712`,
- display both EVM and Casper-native domain separator digests.

This is useful as:

- a walkthrough for integrators,
- a smoke test for the JS package,
- a starting point for wallet / frontend integration.

## Using this as a template

If you want to adapt this example for your own project:

1. rename the token and domain fields,
2. replace the `Permit` struct with your own typed message if needed,
3. decide whether you want EVM-compatible, Casper-native, or both domain modes,
4. keep the nonce and deadline protections,
5. decide where the resulting authorization is stored and consumed,
6. add production-grade error handling and audits.

## Security notes

This demo includes the important basics:

- **nonce replay protection**
- **deadline expiry**
- **domain separation**
- **signature / signer validation**

But it is still a demo.

Before using this pattern in production, you should:

- replace demo-only helpers like unrestricted `mint_to()` with explicit access control,
- audit the allowance storage design,
- review how owner identity maps from Ethereum-style addresses to Casper-side authority,
- confirm your domain schema is stable and documented,
- test with real deployment packaging and client code,
- perform a proper security review.

## Related references

- ERC-2612 / ERC20 Permit
- OpenZeppelin `ERC20Permit`
- Uniswap Permit2
- Rust crate: [`casper-eip-712`](../..)
- TypeScript package: [`js/`](../../js)
