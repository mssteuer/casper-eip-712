import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it, expect } from "vitest";
import { hashDomainSeparator, hashStruct, hashTypedData, toHex } from "../src/index.js";
import type { TypedField, TypeDefinitions } from "../src/types.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const vectorsPath = path.join(__dirname, "..", "..", "tests", "vectors.json");
const vectorsFile = JSON.parse(readFileSync(vectorsPath, "utf-8"));

interface Vector {
  name: string;
  primaryType: string;
  domain: Record<string, unknown>;
  message: Record<string, unknown>;
  domainSeparator: string;
  structHash: string;
  digest: string;
}

const typeDefinitions: Record<string, TypeDefinitions> = {
  Permit: {
    Permit: [
      { name: "owner", type: "address" },
      { name: "spender", type: "address" },
      { name: "value", type: "uint256" },
      { name: "nonce", type: "uint256" },
      { name: "deadline", type: "uint256" },
    ],
  },
  Approval: {
    Approval: [
      { name: "owner", type: "address" },
      { name: "spender", type: "address" },
      { name: "value", type: "uint256" },
    ],
  },
  Transfer: {
    Transfer: [
      { name: "from", type: "address" },
      { name: "to", type: "address" },
      { name: "value", type: "uint256" },
    ],
  },
};

function getDomainTypes(vector: Vector): TypedField[] | undefined {
  if (vector.domain.chain_name !== undefined || vector.domain.contract_package_hash !== undefined) {
    const fields: TypedField[] = [];
    if (vector.domain.name !== undefined) fields.push({ name: "name", type: "string" });
    if (vector.domain.version !== undefined) fields.push({ name: "version", type: "string" });
    if (vector.domain.chain_name !== undefined) fields.push({ name: "chain_name", type: "string" });
    if (vector.domain.contract_package_hash !== undefined) fields.push({ name: "contract_package_hash", type: "bytes32" });
    return fields;
  }
  return undefined;
}

describe("cross-language vectors", () => {
  const vectors: Vector[] = vectorsFile.vectors;

  it(`loaded ${vectors.length} vectors (expect >= 6)`, () => {
    expect(vectors.length).toBeGreaterThanOrEqual(6);
  });

  for (const vector of vectors) {
    describe(vector.name, () => {
      const types = typeDefinitions[vector.primaryType];
      if (!types) {
        throw new Error(`Missing type definitions for primaryType \"${vector.primaryType}\" in cross-language test vectors`);
      }
      const domainTypes = getDomainTypes(vector);

      it("domain separator matches", () => {
        const hash = hashDomainSeparator(vector.domain, domainTypes);
        expect(toHex(hash)).toBe(vector.domainSeparator);
      });

      it("struct hash matches", () => {
        const hash = hashStruct(vector.primaryType, types, vector.message);
        expect(toHex(hash)).toBe(vector.structHash);
      });

      it("digest matches", () => {
        const digest = hashTypedData(
          vector.domain,
          types,
          vector.primaryType,
          vector.message,
          domainTypes ? { domainTypes } : undefined,
        );
        expect(toHex(digest)).toBe(vector.digest);
      });
    });
  }
});
