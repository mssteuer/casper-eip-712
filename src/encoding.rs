use crate::keccak::keccak256;

/// Encode an Ethereum address (20 bytes) as a 32-byte left-padded value.
pub fn encode_address(addr: [u8; 20]) -> [u8; 32] {
    let mut encoded = [0u8; 32];
    encoded[12..32].copy_from_slice(&addr);
    encoded
}

/// Encode a uint256 value (already 32 bytes big-endian). Identity function.
pub fn encode_uint256(value: [u8; 32]) -> [u8; 32] {
    value
}

/// Encode a uint64 value as 32-byte left-padded big-endian.
pub fn encode_uint64(value: u64) -> [u8; 32] {
    let mut encoded = [0u8; 32];
    encoded[24..32].copy_from_slice(&value.to_be_bytes());
    encoded
}

/// Encode a bytes32 value. Identity function.
pub fn encode_bytes32(value: [u8; 32]) -> [u8; 32] {
    value
}

/// Encode a dynamic string per EIP-712: keccak256(value).
pub fn encode_string(value: &str) -> [u8; 32] {
    keccak256(value.as_bytes())
}

/// Encode dynamic bytes per EIP-712: keccak256(value).
pub fn encode_bytes(value: &[u8]) -> [u8; 32] {
    keccak256(value)
}

/// Encode a boolean as a 32-byte left-padded value (0 or 1).
pub fn encode_bool(value: bool) -> [u8; 32] {
    let mut encoded = [0u8; 32];
    encoded[31] = value as u8;
    encoded
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encode_address_left_pads() {
        let addr = [0x11u8; 20];
        let encoded = encode_address(addr);
        assert_eq!(&encoded[0..12], &[0u8; 12]);
        assert_eq!(&encoded[12..32], &addr);
    }

    #[test]
    fn test_encode_uint256_identity() {
        let value = [0xffu8; 32];
        assert_eq!(encode_uint256(value), value);
    }

    #[test]
    fn test_encode_uint64_big_endian_padded() {
        let encoded = encode_uint64(1u64);
        assert_eq!(&encoded[0..24], &[0u8; 24]);
        assert_eq!(&encoded[24..32], &1u64.to_be_bytes());
    }

    #[test]
    fn test_encode_uint64_max() {
        let encoded = encode_uint64(u64::MAX);
        assert_eq!(&encoded[0..24], &[0u8; 24]);
        assert_eq!(&encoded[24..32], &u64::MAX.to_be_bytes());
    }

    #[test]
    fn test_encode_bytes32_identity() {
        let value = [0xab; 32];
        assert_eq!(encode_bytes32(value), value);
    }

    #[test]
    fn test_encode_string_hashes() {
        use crate::keccak::keccak256;
        let encoded = encode_string("hello");
        assert_eq!(encoded, keccak256(b"hello"));
    }

    #[test]
    fn test_encode_bytes_hashes() {
        use crate::keccak::keccak256;
        let data = b"some bytes";
        let encoded = encode_bytes(data);
        assert_eq!(encoded, keccak256(data));
    }

    #[test]
    fn test_encode_bool_true() {
        let encoded = encode_bool(true);
        assert_eq!(encoded[31], 1);
        assert_eq!(&encoded[0..31], &[0u8; 31]);
    }

    #[test]
    fn test_encode_bool_false() {
        let encoded = encode_bool(false);
        assert_eq!(encoded, [0u8; 32]);
    }
}
