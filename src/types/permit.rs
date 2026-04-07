use alloc::vec::Vec;

use crate::encoding::{encode_address, encode_uint256, Address};
use crate::traits::Eip712Struct;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Permit {
    pub owner: Address,
    pub spender: Address,
    pub value: [u8; 32],
    pub nonce: [u8; 32],
    pub deadline: [u8; 32],
}

impl Eip712Struct for Permit {
    fn type_string() -> &'static str {
        "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
    }

    fn encode_data(&self) -> Vec<u8> {
        let mut data = Vec::with_capacity(160);
        data.extend_from_slice(&encode_address(self.owner));
        data.extend_from_slice(&encode_address(self.spender));
        data.extend_from_slice(&encode_uint256(self.value));
        data.extend_from_slice(&encode_uint256(self.nonce));
        data.extend_from_slice(&encode_uint256(self.deadline));
        data
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::keccak::keccak256;
    use crate::traits::Eip712Struct;

    #[test]
    fn test_permit_type_string() {
        assert_eq!(
            Permit::type_string(),
            "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
        );
    }

    #[test]
    fn test_permit_type_hash_matches_string() {
        let expected = keccak256(Permit::type_string().as_bytes());
        assert_eq!(Permit::type_hash(), expected);
    }

    #[test]
    fn test_permit_encode_data_length() {
        let permit = Permit {
            owner: Address::Eth([0x11; 20]),
            spender: Address::Eth([0x22; 20]),
            value: [0; 32],
            nonce: [0; 32],
            deadline: [0; 32],
        };
        assert_eq!(permit.encode_data().len(), 160);
    }

    #[test]
    fn test_permit_hash_struct_deterministic() {
        let permit = Permit {
            owner: Address::Eth([0x11; 20]),
            spender: Address::Eth([0x22; 20]),
            value: [0; 32],
            nonce: [0; 32],
            deadline: [0; 32],
        };
        assert_eq!(permit.hash_struct(), permit.hash_struct());
    }

    #[test]
    fn test_permit_different_values_different_hash() {
        let p1 = Permit { owner: Address::Eth([0x11; 20]), spender: Address::Eth([0x22; 20]), value: [0; 32], nonce: [0; 32], deadline: [0; 32] };
        let p2 = Permit { owner: Address::Eth([0x11; 20]), spender: Address::Eth([0x22; 20]), value: [1; 32], nonce: [0; 32], deadline: [0; 32] };
        assert_ne!(p1.hash_struct(), p2.hash_struct());
    }

    #[test]
    fn test_permit_casper_owner_changes_hash() {
        let eth_permit = Permit {
            owner: Address::Eth([0x11; 20]),
            spender: Address::Eth([0x22; 20]),
            value: [0; 32],
            nonce: [0; 32],
            deadline: [0; 32],
        };
        let mut casper_raw = [0x11u8; 33];
        casper_raw[0] = 0x00;
        let casper_permit = Permit {
            owner: Address::Casper(casper_raw),
            spender: Address::Eth([0x22; 20]),
            value: [0; 32],
            nonce: [0; 32],
            deadline: [0; 32],
        };
        assert_ne!(eth_permit.hash_struct(), casper_permit.hash_struct());
    }
}
