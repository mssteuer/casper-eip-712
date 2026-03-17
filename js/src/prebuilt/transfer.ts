import type { TypeDefinitions } from "../types.js";

export interface TransferMessage extends Record<string, unknown> {
  from: string;
  to: string;
  value: string | bigint;
}

export const TransferTypes: TypeDefinitions = {
  Transfer: [
    { name: "from", type: "address" },
    { name: "to", type: "address" },
    { name: "value", type: "uint256" },
  ],
};
