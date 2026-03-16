import { describe, it, expect } from "vitest";
import { hashDomainSeparator, buildDomainTypeString } from "../src/domain.js";
import { toHex } from "../src/utils.js";

describe("domain", () => {
  it("standard EVM domain produces correct type hash", () => {
    const domain = {
      name: "Test",
      version: "1",
      chainId: 1,
      verifyingContract: "0x1111111111111111111111111111111111111111",
    };
    const typeString = buildDomainTypeString(domain);
    expect(typeString).toBe(
      "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
    );
  });

  it("partial domain omits missing fields", () => {
    const domain = { name: "Test", version: "1" };
    const typeString = buildDomainTypeString(domain);
    expect(typeString).toBe("EIP712Domain(string name,string version)");
  });

  it("custom Casper domain with explicit types", () => {
    const domain = { name: "Test", chain_name: "casper-test", contract_package_hash: "0x" + "77".repeat(32) };
    const domainTypes = [
      { name: "name", type: "string" },
      { name: "chain_name", type: "string" },
      { name: "contract_package_hash", type: "bytes32" },
    ];
    const typeString = buildDomainTypeString(domain, domainTypes);
    expect(typeString).toBe(
      "EIP712Domain(string name,string chain_name,bytes32 contract_package_hash)"
    );
  });

  it("matches vectors.json permit_basic domain separator", () => {
    const domain = {
      name: "MyToken",
      version: "1",
      chainId: 1,
      verifyingContract: "0x1111111111111111111111111111111111111111",
    };
    const hash = hashDomainSeparator(domain);
    expect(toHex(hash)).toBe(
      "0x98a000c59add584118697238deb65a508461c73c65bc49956fb97bfc568c30c8"
    );
  });

  it("matches vectors.json casper_native_domain_permit domain separator", () => {
    const domain = {
      name: "CasperToken",
      version: "1",
      chain_name: "casper-test",
      contract_package_hash: "0x7777777777777777777777777777777777777777777777777777777777777777",
    };
    const domainTypes = [
      { name: "name", type: "string" },
      { name: "version", type: "string" },
      { name: "chain_name", type: "string" },
      { name: "contract_package_hash", type: "bytes32" },
    ];
    const hash = hashDomainSeparator(domain, domainTypes);
    expect(toHex(hash)).toBe(
      "0xb277bd5cc55f7517eaa22010c78cd3e74387ded279c8e3f8f29cc5ea31f2a402"
    );
  });

  it("deterministic — same inputs produce same hash", () => {
    const domain = { name: "A", version: "1" };
    const h1 = hashDomainSeparator(domain);
    const h2 = hashDomainSeparator(domain);
    expect(toHex(h1)).toBe(toHex(h2));
  });

  it("different domains produce different hashes", () => {
    const d1 = hashDomainSeparator({ name: "A", version: "1" });
    const d2 = hashDomainSeparator({ name: "B", version: "1" });
    expect(toHex(d1)).not.toBe(toHex(d2));
  });
});
