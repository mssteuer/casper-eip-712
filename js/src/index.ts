export { hashTypedData, hashTypedDataRaw, hashStruct } from "./hash.js";
export { buildDomain, hashDomainSeparator, buildDomainTypeString, CASPER_DOMAIN_TYPES } from "./domain.js";
export { computeTypeHash } from "./type-hash.js";

export {
  encodeAddress,
  encodeUint256,
  encodeUint64,
  encodeString,
  encodeBytes32,
  encodeBytes,
  encodeBool,
  encodeField,
} from "./encoding.js";

export { recoverAddress, verifySignature, recoverTypedDataSigner } from "./verify.js";
export { keccak256 } from "./keccak.js";
export { toHex, fromHex } from "./utils.js";

export {
  PermitTypes,
  ApprovalTypes,
  TransferTypes,
  type PermitMessage,
  type ApprovalMessage,
  type TransferMessage,
  TransferAuthorizationTypes,
  BatchTransferAuthorizationTypes,
  type TransferAuthorizationMessage,
  type BatchEntryMessage,
  type BatchTransferAuthorizationMessage,
} from "./prebuilt/index.js";

export type { EIP712Domain, CasperDomain, TypedField, TypeDefinitions, TypedDataOptions } from "./types.js";
