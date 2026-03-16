use crate::keccak::keccak256;

/// Compute the EIP-712 type hash for a canonical type string.
pub fn compute_type_hash(type_string: &str) -> [u8; 32] {
    keccak256(type_string.as_bytes())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_permit_type_hash() {
        let type_string = "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)";
        let hash = compute_type_hash(type_string);
        let expected = crate::keccak::keccak256(type_string.as_bytes());
        assert_eq!(hash, expected);
    }

    #[test]
    fn test_eip712_domain_type_hash_full() {
        let type_string = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)";
        let hash = compute_type_hash(type_string);
        let expected = crate::keccak::keccak256(type_string.as_bytes());
        assert_eq!(hash, expected);
    }

    #[test]
    fn test_different_type_strings_different_hashes() {
        let h1 = compute_type_hash("Foo(uint256 a)");
        let h2 = compute_type_hash("Bar(uint256 a)");
        assert_ne!(h1, h2);
    }
}
