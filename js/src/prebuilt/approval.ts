import type { TypeDefinitions } from "../types.js";

export interface ApprovalMessage extends Record<string, unknown> {
  owner: string;
  spender: string;
  value: string | bigint;
}

export const ApprovalTypes: TypeDefinitions = {
  Approval: [
    { name: "owner", type: "address" },
    { name: "spender", type: "address" },
    { name: "value", type: "uint256" },
  ],
};
