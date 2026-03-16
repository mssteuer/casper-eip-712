import { describe, it, expect } from "vitest";
import { recoverAddress, verifySignature, recoverTypedDataSigner } from "../src/verify.js";
import { hashTypedData } from "../src/hash.js";
import { toHex } from "../src/utils.js";
import { secp256k1 } from "@noble/curves/secp256k1";
import { keccak256 } from "../src/keccak.js";

function privateKeyToAddress(privKey: Uint8Array): Uint8Array {
  const pubKey = secp256k1.getPublicKey(privKey, false);
  const hash = keccak256(pubKey.slice(1));
  return hash.slice(12);
}

function signDigest(privKey: Uint8Array, digest: Uint8Array): Uint8Array {
  const sig = secp256k1.sign(digest, privKey);
  const r = sig.r;
  const s = sig.s;
  const signature = new Uint8Array(65);
  const rBytes = hexToBytes(r.toString(16).padStart(64, "0"));
  const sBytes = hexToBytes(s.toString(16).padStart(64, "0"));
  signature.set(rBytes, 0);
  signature.set(sBytes, 32);
  signature[64] = sig.recovery;
  return signature;
}

function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
  }
  return bytes;
}

describe("verify", () => {
  const privKey = new Uint8Array(32).fill(0x11);
  const address = privateKeyToAddress(privKey);

  it("recovers address from signed digest", () => {
    const digest = new Uint8Array(32).fill(0x42);
    const signature = signDigest(privKey, digest);
    const recovered = recoverAddress(digest, signature);
    expect(toHex(recovered)).toBe(toHex(address));
  });

  it("verifySignature returns true for correct signer", () => {
    const digest = new Uint8Array(32).fill(0x24);
    const signature = signDigest(privKey, digest);
    expect(verifySignature(digest, signature, toHex(address))).toBe(true);
  });

  it("verifySignature returns false for wrong signer", () => {
    const digest = new Uint8Array(32).fill(0x24);
    const signature = signDigest(privKey, digest);
    expect(verifySignature(digest, signature, "0x" + "00".repeat(20))).toBe(false);
  });

  it("recoverTypedDataSigner roundtrip", () => {
    const domain = { name: "Test", version: "1" };
    const types = { Simple: [{ name: "value", type: "uint256" }] };
    const message = { value: "0x" + "00".repeat(31) + "01" };
    const digest = hashTypedData(domain, types, "Simple", message);
    const signature = signDigest(privKey, digest);
    const recovered = recoverTypedDataSigner(domain, types, "Simple", message, signature);
    expect(toHex(recovered)).toBe(toHex(address));
  });
});
