# Why We Brought Ethereum's Most Important Signing Standard to Casper

*Subtitle: How a security audit led us to build EIP-712 infrastructure for the entire Casper ecosystem*

*Tags: Blockchain, Casper Network, EIP-712, Web3, Smart Contracts, Rust, TypeScript, Security*

Every blockchain that wants to play in the big leagues (cross-chain bridges, gasless transactions, AI agents moving money on your behalf) needs EIP-712. It's the standard that makes typed, domain-separated signatures possible. Ethereum figured this out years ago. Uniswap uses it. Aave uses it. OpenSea uses it. It's the plumbing behind every "approve without paying gas" interaction you've ever done on an EVM chain.

Casper has it now.

This isn't something we planned on a roadmap six months ago. It started the way the best infrastructure usually starts: with a security auditor telling us we had a gap. During the Halborn security audit of [CSPRbridge.com](https://csprbridge.com), our cross-chain bridge connecting Casper to EVM networks, the auditors flagged that our attestation verification was using ad-hoc signature encoding. Custom `encodePacked`-style hashing, hand-rolled for each message type. It worked, but it was brittle, non-standard, and the kind of thing that keeps security engineers up at night.

We could have patched the finding and moved on. That's what most teams do: fix the specific issue, close the ticket, ship the update. Instead, we asked ourselves: "If this is a gap in the bridge, isn't it a gap in the entire ecosystem?" The answer was yes. So we built infrastructure.

## Why Roll-Your-Own Signatures Fail

Here's a scenario that plays out more often than anyone in crypto likes to admit. A team builds a bridge, or a cross-chain messaging protocol, or a multi-chain dApp. They need signatures: attestations that something happened on Chain A, verified on Chain B. So they roll their own encoding. They pick a hash function, concatenate some fields, maybe add a prefix. Ship it.

Then someone replays a signature from testnet on mainnet. Or submits an attestation meant for the ERC-20 locker contract against the NFT bridge. Or finds that the "domain" is just the contract name hard-coded as a string, with nothing tying the signature to a specific deployment.

This is what domain separation solves, and it's what the ad-hoc approach almost always gets wrong. When every project invents its own encoding, you get inconsistency, incompatibility between systems, and attack surfaces that are invisible until someone exploits them.

Security auditors don't tell you what's nice to have. They tell you what will get you exploited.

That Halborn audit was the catalyst. The bridge's attestation verification was using raw `keccak256` with `encodePacked`-style concatenation: four different hash functions, each assembling bytes manually for different message types. Functional? Yes. Standardized? No. Upgradeable without breaking every deployed relayer? Absolutely not.

We decided that if we were going to fix the bridge, we'd fix it for the whole ecosystem.

## What EIP-712 Is and Why Ethereum Got It Right

If you've been building on Ethereum, you've already used EIP-712, even if you didn't know it by name. Every time you've signed a "permit" in MetaMask and seen a nicely formatted breakdown of what you're approving (the token, the spender, the amount), that's EIP-712 at work.

Before EIP-712, signing on Ethereum meant signing opaque byte blobs. Your wallet would show you a hex string and essentially ask: "Do you trust this?" The answer should always have been "no," but people clicked "sign" anyway. EIP-712 changed this by introducing typed structured data hashing:

**Human-readable in wallets.** Instead of a hex blob, users see the actual fields they're signing: token name, amount, recipient, deadline. The wallet can display this because the data has a schema.

**Domain-separated.** Every signature is bound to a specific domain: a contract address, a chain ID, a protocol version. A signature for Uniswap on Ethereum mainnet cannot be replayed against a copycat contract on Arbitrum. The domain is baked into the hash.

**Machine-verifiable.** Smart contracts can reconstruct the exact hash from the typed data, recover the signer, and verify the authorization on-chain. No ambiguity, no parsing, no custom deserialization logic.

This combination is why EIP-712 became the backbone of gasless approvals (ERC-2612), meta-transactions, and most of modern DeFi's authorization infrastructure. Uniswap's Permit2 uses it to let users approve token spending with a signature instead of a transaction. Aave uses it for credit delegation. OpenSea used it for gasless NFT listings. The pattern is always the same: user signs structured data off-chain, contract verifies on-chain, and the domain separator ensures that signature can't be replayed anywhere it wasn't intended.

It's not just a convenience. It's a security primitive. And it's the kind of standard that separates blockchains that are ready for real-world financial applications from those that aren't.

If you're building on Casper, you're about to start using it too.

[IMAGE 1: Architecture diagram — EIP-712 signing flow: User signs typed data off-chain (zero cost) → Relayer submits to Casper contract → Contract verifies signature, recovers signer, executes action]

## How We Brought EIP-712 to Casper

[`casper-eip-712`](https://github.com/casper-ecosystem/casper-eip-712) is a `no_std`-compatible Rust crate that brings the full EIP-712 toolkit to Casper. It gives you domain construction, typed struct hashing, encoding helpers, and optional secp256k1 signer recovery. Everything you need to implement the same signature patterns that power the EVM ecosystem.

### Building a Domain

The `DomainBuilder` supports both standard EVM fields and Casper-native extensions:

```rust
use casper_eip_712::prelude::*;

// EVM-compatible domain, interoperable with ethers.js and MetaMask
let domain = DomainBuilder::new()
    .name("MyToken")
    .version("1")
    .chain_id(1)
    .verifying_contract([0x11; 20])
    .build();

// Casper-native domain: uses chain_name and contract_package_hash
let domain = DomainBuilder::new()
    .name("Bridge")
    .version("1")
    .custom_field("chain_name", DomainFieldValue::String("casper-test".into()))
    .custom_field("contract_package_hash", DomainFieldValue::Bytes32([0x99; 32]))
    .build();
```

[IMAGE 2: Domain separator comparison — Side-by-side: EVM domain (name, version, chainId, verifyingContract) vs Casper-native domain (name, version, chain_name, contract_package_hash)]

That dual-domain support is intentional. If you're building a bridge that verifies Ethereum-origin signatures, use the EVM-compatible domain. If you're building a Casper-native dApp and want domain fields that make sense for Casper (like `chain_name` and `contract_package_hash` instead of integer chain IDs and 20-byte addresses), use the custom fields. Same hashing algorithm, same security properties, native semantics for each chain.

### Defining Custom Typed Structs

Any struct can be EIP-712 hashable by implementing the `Eip712Struct` trait:

```rust
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

Then hashing is one line:

```rust
let digest = hash_typed_data(&domain, &attestation);
```

The crate ships with prebuilt `Permit`, `Approval`, and `Transfer` structs for common patterns, so you don't have to implement these yourself for standard use cases.

### Design Decisions

A few things worth noting:

- **`no_std` + `alloc`**: The crate is designed for WASM-based contract targets. It works inside Casper smart contracts without pulling in `std`.
- **Optional `verify` feature**: Enable `verify` in your `Cargo.toml` to get secp256k1 signer recovery via `k256`. Leave it off if you only need hashing.
- **Framework-agnostic**: We built `casper-eip-712` as a standalone Rust crate, not as a module for any specific framework. This was intentional: any Casper contract can use it, regardless of tooling. But since most contracts on Casper today are developed using [Odra](https://odra.dev), the rapid-development smart contracting framework, we made sure it integrates seamlessly. Add it to your `Cargo.toml`, import the prelude, and you're signing.

## Real-World Usage: The Bridge

Remember that Halborn audit? Here's the payoff.

The [CSPRbridge.com](https://csprbridge.com) cross-chain bridge connects Casper to EVM networks. Its core security mechanism is multi-relayer attestation: when tokens move between chains, a threshold of relayers must sign an attestation confirming the transaction. The Casper-side contract then verifies these signatures before minting or releasing tokens.

Before `casper-eip-712`, this verification used hand-rolled encoding: four separate hash functions in `crypto.rs`, each manually assembling bytes with `keccak256` and `encodePacked`-style concatenation. Each one was correct for its specific message type, but each was also bespoke. No shared schema. No domain binding per deployment. No way for a wallet or monitoring tool to introspect what was being signed.

The audit identified the gap: signatures weren't bound to a specific contract deployment. An attestation for one bridge instance could theoretically be replayed against another. It also flagged the lack of a standardized encoding scheme, which made the system harder to audit, harder to extend, and harder to reason about.

We didn't just patch the finding. We built infrastructure so that every project on Casper gets this right by default. The bridge is migrating its attestation verification to `casper-eip-712`, binding every signature to a specific chain, contract, and protocol version. What used to be four custom hash functions becomes one standard pattern. And critically, any new message type the bridge needs to support (new token standards, new cross-chain operations) follows the same trait implementation. Add your struct, implement `Eip712Struct`, and the hashing, encoding, and verification are handled.

This is what I mean when I say we built infrastructure, not a patch. The next team that builds a bridge, an oracle, or any system that needs cross-chain attestations on Casper doesn't have to solve this problem again.

## The TypeScript Companion

Smart contracts verify signatures. But someone has to create them first, and that someone is usually a frontend developer writing TypeScript.

[`@casper-ecosystem/casper-eip-712`](https://www.npmjs.com/package/@casper-ecosystem/casper-eip-712) is the TypeScript mirror of the Rust crate. It provides the same domain construction, type hashing, and encoding functions, plus signer recovery for client-side verification:

```typescript
import {
  hashTypedData,
  recoverTypedDataSigner,
  PermitTypes,
  type EIP712Domain,
  type PermitMessage,
} from "@casper-ecosystem/casper-eip-712";

const domain: EIP712Domain = {
  name: "PermitToken",
  version: "1",
  chainId: 1314614895,
  verifyingContract: contractAddress,
};

const permit: PermitMessage = {
  owner: wallet.address,
  spender: spenderAddress,
  value: "0x" + (1000n).toString(16).padStart(64, "0"),
  nonce: "0x" + "0".padStart(64, "0"),
  deadline: "0x" + (0xffffffffn).toString(16).padStart(64, "0"),
};

// Sign with ethers.js, standard Ethereum tooling
const signature = await wallet.signTypedData(domain, PermitTypes, permitPayload);

// Verify locally before submitting to chain
const digest = hashTypedData(domain, PermitTypes, "Permit", permit);
const recovered = recoverTypedDataSigner(domain, PermitTypes, "Permit", permit, sigBytes);
```

Frontend devs don't need to touch Rust. The TypeScript package produces the exact same hashes and digests as the Rust crate, verified by shared test vectors generated from a single source of truth. Sign in the browser, verify on Casper.

## Demo: The Permit/Approve Pattern in Action

Talk is cheap. Let me show you.

The [`examples/permit-token`](https://github.com/casper-ecosystem/casper-eip-712/tree/master/examples/permit-token) directory contains a complete working CEP-18 token with gasless permit/approve, the same pattern that powers Uniswap's token approvals, now running on Casper.

Here's what happens:

1. **A user signs a Permit off-chain** using standard Ethereum tooling (ethers.js `signTypedData`). Zero cost. No transaction.
2. **A relayer submits the signature** to the Casper contract, paying the deploy cost on the user's behalf.
3. **The contract verifies the EIP-712 signature**, recovers the signer's Ethereum address, checks the nonce and deadline, and writes the token allowance. All without the token owner ever submitting a transaction.

The demo is a CEP-18 token built with [Odra](https://odra.dev). The EIP-712 permit pattern drops in as a regular crate dependency. No special adapters, no glue code. Odra handles the smart contract scaffolding; `casper-eip-712` handles the cryptographic verification.

Run the Rust tests:
```bash
cd examples/permit-token
cargo odra test
```

Run the TypeScript demo:
```bash
cd js && npm install && npm run build && cd ..
cd examples/permit-token/demo
npm install
npx tsx demo.ts
```

You can run this demo without a Casper node. Right now. Go.

The TypeScript demo generates a keypair, signs a permit using ethers.js, verifies it client-side with our TypeScript package, and shows you exactly what the on-chain contract call would look like: the 65-byte signature, the reconstructed digest, and the recovered signer address. The Rust integration tests do the full round-trip inside an Odra test environment: deploy the contract, sign a permit, submit it, verify the allowance was set, and confirm that invalid signatures, expired deadlines, and replayed nonces are rejected.

## What This Unlocks for Casper

I didn't write a Rust crate because I was bored on a Saturday night. (Well, partially. But mostly for what it enables.)

### Gasless Transactions and Relayer Services

The number one UX barrier in crypto is "buy gas first." A new user wants to use your dApp. They don't own CSPR. Under the current model, they bounce. With EIP-712 permits and a relayer service, that user signs an authorization off-chain (free) and the relayer submits it on their behalf, paying the deploy cost. The user interacts with Casper without ever acquiring CSPR upfront. This is how you onboard the next million users. By not asking them to figure out gas fees before they can do anything useful.

### Agentic Commerce via x402

The [x402 protocol](https://www.x402.org/) is enabling AI agents to pay for API access and services using HTTP 402 payment flows with stablecoins. Think about that for a second. Autonomous agents, making purchasing decisions, signing payment authorizations. EIP-712 is the natural signing layer for this: structured, domain-separated, human-auditable authorizations that smart contracts can verify. As x402 comes to Casper, `casper-eip-712` provides the cryptographic foundation for agents to sign verifiable, scoped authorizations. Not opaque byte blobs. Typed, structured data with clear semantics that a compliance system, a monitoring tool, or a human can read and understand.

### Cross-Chain Interoperability

In a multi-chain world, domain separation isn't a nice-to-have. It's a security requirement. A signature for your mainnet bridge deployment must not be valid against your testnet deployment. An attestation for the ERC-20 locker must not work against the NFT bridge. `casper-eip-712`'s `DomainBuilder` supports hybrid environments natively: EVM-standard fields for Ethereum-origin signatures, Casper-native fields for Casper-native protocols, and the flexibility to mix custom fields for whatever your architecture needs.

### Batch Authorization

A relayer collects signed permits from multiple users and submits them in a single deploy. Users get instant, gasless interactions. The relayer amortizes deploy costs. Smart contracts verify each signature independently. This pattern is already standard on Ethereum (Uniswap's Permit2, for example). Now it works on Casper.

### Non-Crypto-Native Onboarding

Let's be blunt: the current onboarding flow for most blockchain applications is terrible. I [wrote about this extensively](/csprsuite/enabling-web3-adoption-with-cspr-click-8c551875985a): the "connect your wallet" gauntlet, the gas token requirement, the 24-word seed phrase anxiety. EIP-712 is a key piece of the puzzle for fixing this. Combined with a relayer architecture, you can build Casper dApps where the user never sees a wallet popup, never buys gas, and never knows they're interacting with a blockchain. They sign a message, a message they can actually read and understand, and the relayer handles the rest. That's the user experience that gets us from 420 million crypto users to a billion.

### The Big Picture

This isn't just a library. It's a building block that positions Casper to compete for the next wave of adoption, the wave where users don't know they're using a blockchain, and AI agents transact on their behalf. Gasless UX, agentic commerce, cross-chain interop, batch transactions: these are the patterns that the most successful chains will support natively. `casper-eip-712` gives Casper that foundation.

When I look at where crypto is headed (account abstraction, intent-based architectures, agent-driven commerce), every one of those paths runs through structured, domain-separated signing. We're not building for where blockchain is today. We're building for where it's going.

## Build With Us

The crate is open source, published, and ready to use:

- **GitHub:** [casper-ecosystem/casper-eip-712](https://github.com/casper-ecosystem/casper-eip-712)
- **Rust Crate:** [crates.io/crates/casper-eip-712](https://crates.io/crates/casper-eip-712)
- **NPM Package:** [@casper-ecosystem/casper-eip-712](https://www.npmjs.com/package/@casper-ecosystem/casper-eip-712)
- **Rust Docs:** [docs.rs/casper-eip-712](https://docs.rs/casper-eip-712/1.0.0/casper_eip_712/)
- **Working Demo:** [examples/permit-token](https://github.com/casper-ecosystem/casper-eip-712/tree/master/examples/permit-token)

If you're building on Casper, use it. If you're building on Ethereum and wondering what else is out there, try it. The signatures are the same.
