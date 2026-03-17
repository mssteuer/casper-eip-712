# Image Specifications — EIP-712 Blog Post

## Image 1 — Architecture Diagram

**Title:** "EIP-712 Gasless Signing Flow"

**Layout:** Left-to-right horizontal flow, 3 stages.

**Elements:**
1. **Left — User (wallet icon)**
   - Label: "User"
   - Sub-label: "Signs typed data off-chain (zero cost)"
   - Below: A stylized EIP-712 typed data card showing field names:
     ```
     Permit
     ├── owner: 0x22...
     ├── spender: 0x33...
     ├── value: 1000
     ├── nonce: 0
     └── deadline: 4294967295
     ```

2. **Center — Arrow + Relayer**
   - Arrow from User to Contract
   - Label on arrow: "Relayer submits to Casper contract"
   - Small note: "Relayer pays deploy cost"

3. **Right — Casper Contract (hexagonal or shield icon)**
   - Label: "Smart Contract"
   - Sub-label: "Verifies signature → Recovers signer → Executes action"
   - Small checkmark icon

**Style:** Clean, minimal. Casper brand colors (Casper red #FF473E + dark navy). White background. Thin connecting arrows. Modern sans-serif font.

**Dimensions:** 1200×600px (Medium hero image ratio)

---

## Image 2 — Domain Separator Comparison

**Title:** "Same Standard, Extended for Casper"

**Layout:** Two code-card columns side by side with a shared header.

**Shared Header:** "DomainBuilder supports both"

**Left Column — "EVM Domain"**
```
EIP712Domain {
  name: "CasperBridge"
  version: "1"
  chainId: 11155111
  verifyingContract: 0xAb5801a7...
}
```
Small Ethereum logo watermark

**Right Column — "Casper-Native Domain"**
```
EIP712Domain {
  name: "Bridge"
  version: "1"
  chain_name: "casper-test"
  contract_package_hash: 0x99ab...
}
```
Small Casper logo watermark

**Connecting element:** A subtle "=" sign or bridge icon between the columns, indicating equivalence in security properties.

**Style:** Code-card style with dark backgrounds (#1a1a2e or similar) and syntax-highlighted text (field names in one color, values in another). Casper brand accent color for the Casper column. Slight rounded corners on the cards.

**Dimensions:** 1200×500px
