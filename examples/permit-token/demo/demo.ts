import { Wallet } from "ethers";
import {
  CASPER_DOMAIN_TYPES,
  PermitTypes,
  hashDomainSeparator,
  hashTypedData,
  recoverTypedDataSigner,
  toHex,
  type EIP712Domain,
  type PermitMessage,
  type TypedField,
} from "@casper-ecosystem/casper-eip-712";

console.log("═══════════════════════════════════════════════════════");
console.log(" EIP-712 Permit Demo — casper-eip-712");
console.log("═══════════════════════════════════════════════════════\n");

const wallet = Wallet.createRandom();
console.log("Step 1: Generated keypair");
console.log(` Address: ${wallet.address}`);
console.log(` Public key: ${wallet.publicKey.slice(0, 42)}...\n`);

const TOKEN_NAME = "PermitToken";
const DOMAIN_VERSION = "1";
const CONTRACT_HASH = `0x${"ab".repeat(32)}`;
const CONTRACT_ADDR = `0x${"ab".repeat(20)}`;
const spender = `0x${"cc".repeat(20)}`;

const evmDomain: EIP712Domain = {
  name: TOKEN_NAME,
  version: DOMAIN_VERSION,
  chainId: 1314614895,
  verifyingContract: CONTRACT_ADDR,
};

const casperDomain = {
  name: TOKEN_NAME,
  version: DOMAIN_VERSION,
  chain_name: "casper",
  contract_package_hash: CONTRACT_HASH,
};

const casperDomainTypes: TypedField[] = CASPER_DOMAIN_TYPES;

console.log("Step 2: Built EIP-712 domains");
console.log(` EVM domain: name=\"${TOKEN_NAME}\", chainId=1314614895`);
console.log(` Casper domain: name=\"${TOKEN_NAME}\", chain_name=\"casper\"`);
console.log(` EVM separator: ${toHex(hashDomainSeparator(evmDomain))}`);
console.log(` Casper separator: ${toHex(hashDomainSeparator(casperDomain, casperDomainTypes))}\n`);

const permitMessage: PermitMessage = {
  owner: wallet.address,
  spender,
  value: `0x${(1000n).toString(16).padStart(64, "0")}`,
  nonce: `0x${(0n).toString(16).padStart(64, "0")}`,
  deadline: `0x${(0xffffffffn).toString(16).padStart(64, "0")}`,
};

console.log("Step 3: Constructed Permit message");
console.log(` Owner: ${permitMessage.owner}`);
console.log(` Spender: ${permitMessage.spender}`);
console.log(" Value: 1000");
console.log(" Nonce: 0");
console.log(" Deadline: 4294967295\n");

const ethersPayload = {
  owner: wallet.address,
  spender,
  value: 1000n,
  nonce: 0n,
  deadline: 0xffffffffn,
};
const signatureHex = await wallet.signTypedData(evmDomain, PermitTypes, ethersPayload);

console.log("Step 4: Signed with ethers.js signTypedData()");
console.log(` Signature: ${signatureHex.slice(0, 42)}...`);
console.log(` Length: ${(signatureHex.length - 2) / 2} bytes\n`);

const sigBytes = new Uint8Array(signatureHex.slice(2).match(/.{2}/g)!.map((b) => Number.parseInt(b, 16)));
const digest = hashTypedData(evmDomain, PermitTypes, "Permit", permitMessage);
const recoveredBytes = recoverTypedDataSigner(evmDomain, PermitTypes, "Permit", permitMessage, sigBytes);
const recoveredAddress = toHex(recoveredBytes);
const matches = recoveredAddress.toLowerCase() === wallet.address.toLowerCase();

console.log("Step 5: Verified locally with @casper-ecosystem/casper-eip-712");
console.log(` Digest: ${toHex(digest)}`);
console.log(` Recovered: ${recoveredAddress}`);
console.log(` Matches: ${matches ? "✅ YES" : "❌ NO"}\n`);

const casperDigest = hashTypedData(casperDomain, PermitTypes, "Permit", permitMessage, {
  domainTypes: casperDomainTypes,
});
console.log("Step 5b: Casper-native domain variant");
console.log(` Digest: ${toHex(casperDigest)}`);
console.log(" Same message schema, different domain binding.\n");

console.log("Step 6: On-chain contract call shape");
console.log(" Contract: PermitToken::permit()");
console.log(" Arguments:");
console.log(` owner_eth_address: ${wallet.address}`);
console.log(" spender: <casper-account-hash>");
console.log(" value: 1000");
console.log(" nonce: 0");
console.log(" deadline: 4294967295");
console.log(` signature: ${signatureHex.slice(0, 20)}...`);
console.log(" use_casper_domain: false\n");

console.log(" The same 65-byte signature produced by ethers.js is submitted");
console.log(" to the Casper contract. The Rust contract reconstructs the");
console.log(" EIP-712 digest, recovers the signer, checks the nonce and");
console.log(" deadline, then writes the allowance without the owner paying gas.\n");

console.log("═══════════════════════════════════════════════════════");
console.log(" ✅ Demo complete — sign with Ethereum tooling,");
console.log(" verify on Casper.");
console.log("═══════════════════════════════════════════════════════");
