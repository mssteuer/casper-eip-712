import { describe, it, expect } from "vitest";
import {
  encodeAddress,
  encodeUint256,
  encodeUint64,
  encodeString,
  encodeBytes32,
  encodeBytes,
  encodeBool,
} from "../src/encoding.js";
import { toHex, fromHex } from "../src/utils.js";
import { keccak256 } from "../src/keccak.js";

describe("encoding", () => {
  describe("encodeAddress", () => {
    it("left-pads 20-byte address to 32 bytes", () => {
      const encoded = encodeAddress("0x1111111111111111111111111111111111111111");
      expect(encoded.length).toBe(32);
      expect(toHex(encoded.slice(0, 12))).toBe("0x000000000000000000000000");
      expect(toHex(encoded.slice(12))).toBe("0x1111111111111111111111111111111111111111");
    });

    it("handles zero address", () => {
      const encoded = encodeAddress("0x0000000000000000000000000000000000000000");
      expect(encoded).toEqual(new Uint8Array(32));
    });

    it("encodes 33-byte Casper account hash as keccak256", () => {
      // AccountHash: 0x00 prefix + 32 bytes of 0x11
      const hex = "0x00" + "11".repeat(32);
      const encoded = encodeAddress(hex);
      expect(encoded.length).toBe(32);
      const raw = fromHex(hex);
      expect(toHex(encoded)).toBe(toHex(keccak256(raw)));
    });

    it("encodes 33-byte Casper package hash as keccak256", () => {
      // PackageHash: 0x01 prefix + 32 bytes of 0x11
      const hex = "0x01" + "11".repeat(32);
      const encoded = encodeAddress(hex);
      expect(encoded.length).toBe(32);
      const raw = fromHex(hex);
      expect(toHex(encoded)).toBe(toHex(keccak256(raw)));
    });

    it("account hash and package hash with same payload produce different slots", () => {
      const accountHash = "0x00" + "42".repeat(32);
      const packageHash = "0x01" + "42".repeat(32);
      expect(toHex(encodeAddress(accountHash))).not.toBe(toHex(encodeAddress(packageHash)));
    });

    it("rejects addresses that are neither 20 nor 33 bytes", () => {
      expect(() => encodeAddress("0x" + "aa".repeat(21))).toThrow("Address must be 20 or 33 bytes");
      expect(() => encodeAddress("0x" + "aa".repeat(32))).toThrow("Address must be 20 or 33 bytes");
    });
  });

  describe("encodeUint256", () => {
    it("encodes bigint as 32-byte big-endian", () => {
      const encoded = encodeUint256(1n);
      expect(encoded.length).toBe(32);
      expect(encoded[31]).toBe(1);
      expect(encoded[0]).toBe(0);
    });

    it("encodes hex string", () => {
      const encoded = encodeUint256("0x0000000000000000000000000000000000000000000000000000000000001234");
      expect(encoded.length).toBe(32);
      expect(encoded[30]).toBe(0x12);
      expect(encoded[31]).toBe(0x34);
    });

    it("encodes max uint256", () => {
      const max = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff";
      const encoded = encodeUint256(max);
      expect(encoded.every((b) => b === 0xff)).toBe(true);
    });
  });

  describe("encodeString", () => {
    it("returns keccak256 of UTF-8 bytes", () => {
      const encoded = encodeString("hello");
      const expected = keccak256(new TextEncoder().encode("hello"));
      expect(toHex(encoded)).toBe(toHex(expected));
    });
  });

  describe("encodeBytes32", () => {
    it("passes through 32 bytes", () => {
      const hex = "0x" + "ab".repeat(32);
      const encoded = encodeBytes32(hex);
      expect(encoded.length).toBe(32);
      expect(encoded.every((b) => b === 0xab)).toBe(true);
    });
  });

  describe("encodeBytes", () => {
    it("returns keccak256 of raw bytes", () => {
      const data = new Uint8Array([1, 2, 3]);
      const encoded = encodeBytes(data);
      expect(toHex(encoded)).toBe(toHex(keccak256(data)));
    });
  });

  describe("encodeBool", () => {
    it("encodes true as 1 in last byte", () => {
      const encoded = encodeBool(true);
      expect(encoded[31]).toBe(1);
      expect(encoded.slice(0, 31).every((b) => b === 0)).toBe(true);
    });

    it("encodes false as all zeros", () => {
      const encoded = encodeBool(false);
      expect(encoded.every((b) => b === 0)).toBe(true);
    });
  });
});
