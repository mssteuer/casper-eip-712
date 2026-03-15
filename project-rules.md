# casper-eip-712 — Project Rules

## Tech Stack
- **Rust crate**: `no_std` compatible, sha3 for keccak256
- **TypeScript package**: ESM, zero runtime deps, vitest for tests
- **Demo contract**: Odra framework (Casper smart contracts)

## Build & Verification

### Rust crate
```bash
cargo test
cargo test --no-default-features  # verify no_std builds
```

### TypeScript package
```bash
cd js && npm install && npm test
```

## Key Design Principles
- API must mirror EIP-712 semantics exactly
- Cross-language test vectors: Rust and TypeScript MUST produce identical hashes for same inputs
- No external dependencies beyond sha3/keccak256
- Must work in WASM (Casper contract target)

## Quality
- Every public function has doc comments with examples
- Integration tests with known EIP-712 test vectors from Ethereum
- README with quick-start guide
