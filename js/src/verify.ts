import { secp256k1 } from "@noble/curves/secp256k1";
import { keccak256 } from "./keccak.js";
import { hashTypedData } from "./hash.js";
import { fromHex } from "./utils.js";
import type { EIP712Domain, TypeDefinitions, TypedDataOptions } from "./types.js";

function pubKeyToAddress(pubKeyUncompressed: Uint8Array): Uint8Array {
  const hash = keccak256(pubKeyUncompressed.slice(1));
  return hash.slice(12);
}

export function recoverAddress(digest: Uint8Array, signature: Uint8Array): Uint8Array {
  if (signature.length !== 65) throw new Error(`Signature must be 65 bytes, got ${signature.length}`);

  let v = signature[64];
  if (v >= 27) v -= 27;
  if (v > 1) throw new Error(`Invalid recovery id: ${signature[64]}`);

  const r = BigInt("0x" + Array.from(signature.slice(0, 32), (b) => b.toString(16).padStart(2, "0")).join(""));
  const s = BigInt("0x" + Array.from(signature.slice(32, 64), (b) => b.toString(16).padStart(2, "0")).join(""));

  const sig = new secp256k1.Signature(r, s);
  const pubKey = sig.addRecoveryBit(v).recoverPublicKey(digest).toRawBytes(false);
  return pubKeyToAddress(pubKey);
}

export function verifySignature(
  digest: Uint8Array,
  signature: Uint8Array,
  expectedAddress: string,
): boolean {
  try {
    const recovered = recoverAddress(digest, signature);
    const expected = fromHex(expectedAddress);
    if (recovered.length !== expected.length) return false;
    return recovered.every((b, i) => b === expected[i]);
  } catch {
    return false;
  }
}

export function recoverTypedDataSigner(
  domain: EIP712Domain,
  types: TypeDefinitions,
  primaryType: string,
  message: Record<string, unknown>,
  signature: Uint8Array,
  options?: TypedDataOptions,
): Uint8Array {
  const digest = hashTypedData(domain, types, primaryType, message, options);
  return recoverAddress(digest, signature);
}
