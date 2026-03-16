use alloc::vec::Vec;

use crate::encoding::{encode_address, encode_uint256};
use crate::traits::Eip712Struct;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Approval {
    pub owner: [u8; 20],
    pub spender: [u8; 20],
    pub value: [u8; 32],
}

impl Eip712Struct for Approval {
    fn type_string() -> &'static str {
        "Approval(address owner,address spender,uint256 value)"
    }

    fn encode_data(&self) -> Vec<u8> {
        let mut data = Vec::with_capacity(96);
        data.extend_from_slice(&encode_address(self.owner));
        data.extend_from_slice(&encode_address(self.spender));
        data.extend_from_slice(&encode_uint256(self.value));
        data
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::keccak::keccak256;
    use crate::traits::Eip712Struct;

    #[test]
    fn test_approval_type_string() {
        assert_eq!(Approval::type_string(), "Approval(address owner,address spender,uint256 value)");
    }

    #[test]
    fn test_approval_type_hash_matches_string() {
        let expected = keccak256(Approval::type_string().as_bytes());
        assert_eq!(Approval::type_hash(), expected);
    }

    #[test]
    fn test_approval_encode_data_length() {
        let approval = Approval { owner: [0x11; 20], spender: [0x22; 20], value: [0; 32] };
        assert_eq!(approval.encode_data().len(), 96);
    }

    #[test]
    fn test_approval_hash_struct_deterministic() {
        let approval = Approval { owner: [0x11; 20], spender: [0x22; 20], value: [0; 32] };
        assert_eq!(approval.hash_struct(), approval.hash_struct());
    }

    #[test]
    fn test_approval_different_values_different_hash() {
        let a1 = Approval { owner: [0x11; 20], spender: [0x22; 20], value: [0; 32] };
        let a2 = Approval { owner: [0x11; 20], spender: [0x22; 20], value: [1; 32] };
        assert_ne!(a1.hash_struct(), a2.hash_struct());
    }
}
