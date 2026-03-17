# X Announcement Drafts — EIP-712 Blog Post

## @Casper_Network Thread (3–4 tweets)

### Tweet 1 (Hook)
We just open-sourced EIP-712 for Casper Network.

Typed, domain-separated signatures — the same standard behind Uniswap permits, Aave credit delegation, and gasless approvals on Ethereum.

Here's why this matters for Casper 🧵

### Tweet 2 (Problem)
Every cross-chain bridge and multi-chain dApp needs signatures that can't be replayed.

Without a standard, teams roll their own encoding. That leads to security gaps — the exact kind our Halborn security audit flagged in @CSPRbridge.

So we didn't just patch the bridge. We built infrastructure for the ecosystem.

### Tweet 3 (What it enables)
What casper-eip-712 unlocks:

🔹 Gasless transactions — users sign off-chain, relayers submit on-chain
🔹 Agentic commerce via x402 — AI agents signing structured authorizations
🔹 Cross-chain interop — hybrid domain separators for EVM ↔ Casper
🔹 CEP-18 permit/approve — the pattern that powers DeFi

### Tweet 4 (CTA)
Rust crate (no_std, WASM-ready):
https://crates.io/crates/casper-eip-712

TypeScript companion:
https://www.npmjs.com/package/@casper-ecosystem/casper-eip-712

Working demo you can run right now:
https://github.com/casper-ecosystem/casper-eip-712

Full blog post: [LINK]

---

## @mssteuer Signal Boost

Wrote about why we brought EIP-712 to Casper.

It started with a security audit. Ended with open-source infrastructure that unlocks gasless UX, agentic transactions, and real cross-chain interop.

Rust crate + TypeScript companion + working demo. All open source.

[LINK]
