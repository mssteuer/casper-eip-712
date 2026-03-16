use alloc::vec::Vec;

use crate::encoding::{encode_address, encode_uint256};
use crate::traits::Eip712Struct;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Transfer {
    pub from: [u8; 20],
    pub to: [u8; 20],
    pub value: [u8; 32],
}

impl Eip712Struct for Transfer {
    fn type_string() -> &'static str {
        "Transfer(address from,address to,uint256 value)"
    }

    fn encode_data(&self) -> Vec<u8> {
        let mut data = Vec::with_capacity(96);
        data.extend_from_slice(&encode_address(self.from));
        data.extend_from_slice(&encode_address(self.to));
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
    fn test_transfer_type_string() {
        assert_eq!(Transfer::type_string(), "Transfer(address from,address to,uint256 value)");
    }

    #[test]
    fn test_transfer_type_hash_matches_string() {
        let expected = keccak256(Transfer::type_string().as_bytes());
        assert_eq!(Transfer::type_hash(), expected);
    }

    #[test]
    fn test_transfer_encode_data_length() {
        let transfer = Transfer { from: [0x11; 20], to: [0x22; 20], value: [0; 32] };
        assert_eq!(transfer.encode_data().len(), 96);
    }

    #[test]
    fn test_transfer_hash_struct_deterministic() {
        let transfer = Transfer { from: [0x11; 20], to: [0x22; 20], value: [0; 32] };
        assert_eq!(transfer.hash_struct(), transfer.hash_struct());
    }

    #[test]
    fn test_transfer_different_values_different_hash() {
        let t1 = Transfer { from: [0x11; 20], to: [0x22; 20], value: [0; 32] };
        let t2 = Transfer { from: [0x11; 20], to: [0x22; 20], value: [1; 32] };
        assert_ne!(t1.hash_struct(), t2.hash_struct());
    }
}
