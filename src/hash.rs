use crate::domain::DomainSeparator;
use crate::keccak::keccak256;
use crate::traits::Eip712Struct;

/// Compute the final EIP-712 typed data digest.
pub fn hash_typed_data(domain: &DomainSeparator, message: &impl Eip712Struct) -> [u8; 32] {
    let mut data = [0u8; 66];
    data[0] = 0x19;
    data[1] = 0x01;
    data[2..34].copy_from_slice(&domain.separator_hash());
    data[34..66].copy_from_slice(&message.hash_struct());
    keccak256(&data)
}

#[cfg(test)]
mod tests {
    extern crate alloc;

    use alloc::vec::Vec;

    use super::*;
    use crate::domain::DomainBuilder;
    use crate::keccak::keccak256;
    use crate::traits::Eip712Struct;

    struct SimpleStruct {
        value: [u8; 32],
    }

    impl Eip712Struct for SimpleStruct {
        fn type_string() -> &'static str {
            "SimpleStruct(uint256 value)"
        }

        fn type_hash() -> [u8; 32] {
            keccak256(Self::type_string().as_bytes())
        }

        fn encode_data(&self) -> Vec<u8> {
            self.value.to_vec()
        }
    }

    #[test]
    fn test_hash_typed_data_format() {
        let domain = DomainBuilder::new().name("Test").version("1").build();
        let msg = SimpleStruct { value: [0x01; 32] };
        let hash = hash_typed_data(&domain, &msg);

        let mut expected_input = Vec::new();
        expected_input.push(0x19);
        expected_input.push(0x01);
        expected_input.extend_from_slice(&domain.separator_hash());
        expected_input.extend_from_slice(&msg.hash_struct());
        let expected = keccak256(&expected_input);

        assert_eq!(hash, expected);
    }
}
