extern crate alloc;

use alloc::vec::Vec;

use crate::keccak::keccak256;

/// Trait for types that can be hashed per EIP-712.
pub trait Eip712Struct {
    /// The canonical EIP-712 type string.
    fn type_string() -> &'static str;

    /// keccak256 of the canonical type string.
    fn type_hash() -> [u8; 32];

    /// ABI-encode the struct fields (without the type hash prefix).
    fn encode_data(&self) -> Vec<u8>;

    /// Compute hashStruct = keccak256(typeHash ‖ encodeData).
    fn hash_struct(&self) -> [u8; 32] {
        let encoded = self.encode_data();
        let mut data = Vec::with_capacity(32 + encoded.len());
        data.extend_from_slice(&Self::type_hash());
        data.extend_from_slice(&encoded);
        keccak256(&data)
    }
}
