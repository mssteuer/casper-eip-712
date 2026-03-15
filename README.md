# casper-eip-712

EIP-712 style typed data hashing and domain separation for Casper Network.

## Status

🚧 Under active design — see [CCC task #324] for planning discussion.

## Motivation

Casper Network lacks a standardized approach to domain-separated, typed-data hashing.
Every project doing cross-chain bridges, multi-sig wallets, or gasless approvals
rolls their own signature scheme — leading to replay vulnerabilities and incompatibility.

This crate brings the battle-tested EIP-712 standard to Casper, enabling:

- **Cross-chain interop**: EVM wallets sign, Casper contracts verify
- **Replay protection**: Domain separators bind signatures to specific deployments
- **Developer ergonomics**: Type-safe API, pre-built helpers, comprehensive docs

## Components

| Package | Language | Status |
|---------|----------|--------|
| `casper-eip-712` | Rust | 🚧 Design |
| `casper-eip-712-js` | TypeScript | 🚧 Planned |
| Demo contract | Rust/Odra | 🚧 Planned |

## Born from a real audit

This crate was born from Halborn security findings (FIND-223, FIND-224) during
the Casper↔EVM Bridge audit. The bridge is its first consumer.

## License

Apache-2.0
