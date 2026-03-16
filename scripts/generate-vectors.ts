import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { TypedDataEncoder, concat, getBytes, keccak256, toUtf8Bytes, zeroPadValue } from "ethers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.join(__dirname, "..");

type Field = { name: string; type: string };

type Vector = {
  name: string;
  primaryType: string;
  domain: Record<string, unknown>;
  message: Record<string, unknown>;
  domainSeparator: string;
  structHash: string;
  digest: string;
};

function typeString(primaryType: string, fields: Field[]): string {
  return `${primaryType}(${fields.map((field) => `${field.type} ${field.name}`).join(",")})`;
}

function encodeFieldValue(fieldType: string, value: unknown): string {
  switch (fieldType) {
    case "string":
      return keccak256(toUtf8Bytes(String(value)));
    case "bytes":
      return keccak256(getBytes(String(value)));
    case "address":
      return zeroPadValue(String(value), 32);
    case "bool":
      return zeroPadValue((value ? "0x01" : "0x00"), 32);
    case "uint256":
    case "bytes32":
      return zeroPadValue(String(value), 32);
    default:
      throw new Error(`unsupported field type ${fieldType}`);
  }
}

function manualStructHash(primaryType: string, fields: Field[], value: Record<string, unknown>): string {
  const encodedFields = [keccak256(toUtf8Bytes(typeString(primaryType, fields)))];
  for (const field of fields) {
    encodedFields.push(encodeFieldValue(field.type, value[field.name]));
  }
  return keccak256(concat(encodedFields));
}

function manualTypedDataDigest(domainSeparator: string, structHash: string): string {
  return keccak256(concat(["0x1901", domainSeparator, structHash]));
}

function standardVector(
  name: string,
  primaryType: string,
  domain: Record<string, unknown>,
  types: Record<string, Field[]>,
  message: Record<string, unknown>,
): Vector {
  return {
    name,
    primaryType,
    domain,
    message,
    domainSeparator: TypedDataEncoder.hashDomain(domain),
    structHash: TypedDataEncoder.from(types).hash(message),
    digest: TypedDataEncoder.hash(domain, types, message),
  };
}

function customDomainVector(
  name: string,
  primaryType: string,
  domain: Record<string, unknown>,
  domainFields: Field[],
  messageFields: Field[],
  message: Record<string, unknown>,
): Vector {
  const domainSeparator = manualStructHash("EIP712Domain", domainFields, domain);
  const structHash = manualStructHash(primaryType, messageFields, message);

  return {
    name,
    primaryType,
    domain,
    message,
    domainSeparator,
    structHash,
    digest: manualTypedDataDigest(domainSeparator, structHash),
  };
}

const permitFields: Field[] = [
  { name: "owner", type: "address" },
  { name: "spender", type: "address" },
  { name: "value", type: "uint256" },
  { name: "nonce", type: "uint256" },
  { name: "deadline", type: "uint256" },
];

const approvalFields: Field[] = [
  { name: "owner", type: "address" },
  { name: "spender", type: "address" },
  { name: "value", type: "uint256" },
];

const transferFields: Field[] = [
  { name: "from", type: "address" },
  { name: "to", type: "address" },
  { name: "value", type: "uint256" },
];

const vectors: Vector[] = [
  standardVector(
    "permit_basic",
    "Permit",
    {
      name: "MyToken",
      version: "1",
      chainId: 1,
      verifyingContract: "0x1111111111111111111111111111111111111111",
    },
    { Permit: permitFields },
    {
      owner: "0x2222222222222222222222222222222222222222",
      spender: "0x3333333333333333333333333333333333333333",
      value: "0x4444444444444444444444444444444444444444444444444444444444444444",
      nonce: "0x5555555555555555555555555555555555555555555555555555555555555555",
      deadline: "0x6666666666666666666666666666666666666666666666666666666666666666",
    },
  ),
  standardVector(
    "approval_basic",
    "Approval",
    {
      name: "MyToken",
      version: "1",
      chainId: 1,
      verifyingContract: "0x1111111111111111111111111111111111111111",
    },
    { Approval: approvalFields },
    {
      owner: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      spender: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      value: "0x0000000000000000000000000000000000000000000000000000000000001234",
    },
  ),
  standardVector(
    "transfer_basic",
    "Transfer",
    {
      name: "MyToken",
      version: "1",
      chainId: 1,
      verifyingContract: "0x1111111111111111111111111111111111111111",
    },
    { Transfer: transferFields },
    {
      from: "0x1234567890abcdef1234567890abcdef12345678",
      to: "0x876543210fedcba9876543210fedcba987654321",
      value: "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
    },
  ),
  standardVector(
    "permit_partial_domain",
    "Permit",
    {
      name: "MyToken",
      version: "1",
    },
    { Permit: permitFields },
    {
      owner: "0x9999999999999999999999999999999999999999",
      spender: "0x8888888888888888888888888888888888888888",
      value: "0x0000000000000000000000000000000000000000000000000000000000000000",
      nonce: "0x0000000000000000000000000000000000000000000000000000000000000001",
      deadline: "0x0000000000000000000000000000000000000000000000000000000000000002",
    },
  ),
  customDomainVector(
    "casper_native_domain_permit",
    "Permit",
    {
      name: "CasperToken",
      version: "1",
      chain_name: "casper-test",
      contract_package_hash:
        "0x7777777777777777777777777777777777777777777777777777777777777777",
    },
    [
      { name: "name", type: "string" },
      { name: "version", type: "string" },
      { name: "chain_name", type: "string" },
      { name: "contract_package_hash", type: "bytes32" },
    ],
    permitFields,
    {
      owner: "0x0101010101010101010101010101010101010101",
      spender: "0x0202020202020202020202020202020202020202",
      value: "0x000000000000000000000000000000000000000000000000000000000000002a",
      nonce: "0x0000000000000000000000000000000000000000000000000000000000000007",
      deadline: "0x00000000000000000000000000000000000000000000000000000000000003e8",
    },
  ),
  standardVector(
    "approval_zero_value_edge",
    "Approval",
    {
      name: "EdgeToken",
      version: "9",
      chainId: 42,
      verifyingContract: "0x0000000000000000000000000000000000000001",
    },
    { Approval: approvalFields },
    {
      owner: "0x0000000000000000000000000000000000000000",
      spender: "0xffffffffffffffffffffffffffffffffffffffff",
      value: "0x0000000000000000000000000000000000000000000000000000000000000000",
    },
  ),
];

const payload = {
  generatedBy: "ethers.TypedDataEncoder.hash",
  generatedAt: new Date().toISOString(),
  vectors,
};

const outputPath = path.join(repoRoot, "tests", "vectors.json");
mkdirSync(path.dirname(outputPath), { recursive: true });
writeFileSync(outputPath, `${JSON.stringify(payload, null, 2)}\n`);
console.log(`wrote ${outputPath}`);
