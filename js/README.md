# @casper-ecosystem/casper-eip-712

TypeScript companion package for the Rust `casper-eip-712` crate.

It provides EIP-712 typed data hashing plus ECDSA signature recovery for Casper-oriented flows, including support for Casper-specific domain fields when you supply explicit domain types.

## Install

```bash
npm install @casper-ecosystem/casper-eip-712
```

## Quick start

```ts
import {
  hashTypedData,
  verifySignature,
  PermitTypes,
  type PermitMessage,
} from "@casper-ecosystem/casper-eip-712";

const domain = {
  name: "My Token",
  version: "1",
  chainId: 1,
  verifyingContract: "0x1111111111111111111111111111111111111111",
};

const message: PermitMessage = {
  owner: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  spender: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  value: "0x01",
  nonce: "0x00",
  deadline: "0xffff",
};

const digest = hashTypedData(domain, PermitTypes, "Permit", message);

const signer = verifySignature(
  domain,
  PermitTypes,
  "Permit",
  message,
  {
    r: "0x...",
    s: "0x...",
    v: 27,
  },
);
```

## Casper-specific domains

If your domain uses Casper-specific fields such as `chain_name` or `contract_package_hash`, pass explicit `domainTypes` so the package knows the intended domain schema.

```ts
import { hashTypedData } from "@casper-ecosystem/casper-eip-712";

const domain = {
  name: "CSPR",
  version: "1",
  chain_name: "casper-test",
  contract_package_hash:
    "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
};

const digest = hashTypedData(domain, types, "Permit", message, {
  domainTypes: [
    { name: "name", type: "string" },
    { name: "version", type: "string" },
    { name: "chain_name", type: "string" },
    { name: "contract_package_hash", type: "bytes32" },
  ],
});
```

## API surface

- `hashDomainSeparator(domain, domainTypes?)`
- `hashStruct(primaryType, types, message)`
- `hashTypedData(domain, types, primaryType, message, options?)`
- `recoverAddress(digest, signature)`
- `recoverTypedDataSigner(domain, types, primaryType, message, signature, options?)`
- `verifySignature(domain, types, primaryType, message, signature, options?)`
- Encoding helpers: `encodeAddress`, `encodeUint256`, `encodeUint64`, `encodeBytes32`, `encodeBytes`, `encodeString`, `encodeBool`, `encodeField`
- Prebuilt message types: `PermitTypes`, `ApprovalTypes`, `TransferTypes`

## Development

```bash
npm test
npm run build
```

Cross-language parity is covered by test vectors in `../tests/vectors.json`.
