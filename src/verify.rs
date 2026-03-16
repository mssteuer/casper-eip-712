use k256::ecdsa::{RecoveryId, Signature, VerifyingKey};

use crate::domain::DomainSeparator;
use crate::hash::hash_typed_data;
use crate::keccak::keccak256;
use crate::traits::Eip712Struct;

fn pubkey_to_eth_address(verifying_key: &VerifyingKey) -> [u8; 20] {
    let encoded = verifying_key.to_encoded_point(false);
    let bytes = encoded.as_bytes();
    let hash = keccak256(&bytes[1..]);
    let mut address = [0u8; 20];
    address.copy_from_slice(&hash[12..32]);
    address
}

pub fn recover_eth_address(message_hash: [u8; 32], signature: &[u8; 65]) -> Option<[u8; 20]> {
    let sig = Signature::try_from(&signature[..64]).ok()?;
    let mut v = signature[64];
    if v >= 27 {
        v = v.checked_sub(27)?;
    }
    if v > 1 {
        return None;
    }
    let recid = RecoveryId::try_from(v).ok()?;
    let verifying_key = VerifyingKey::recover_from_prehash(&message_hash, &sig, recid).ok()?;
    Some(pubkey_to_eth_address(&verifying_key))
}

pub fn verify_signature(message_hash: [u8; 32], signature: &[u8; 65], expected_signer: [u8; 20]) -> bool {
    recover_eth_address(message_hash, signature) == Some(expected_signer)
}

pub fn recover_typed_data_signer(
    domain: &DomainSeparator,
    message: &impl Eip712Struct,
    signature: &[u8; 65],
) -> Option<[u8; 20]> {
    let digest = hash_typed_data(domain, message);
    recover_eth_address(digest, signature)
}

#[cfg(test)]
mod tests {
    use super::*;
    use k256::ecdsa::{signature::hazmat::PrehashSigner, SigningKey};

    #[test]
    fn test_recover_eth_address_known_key_roundtrip() {
        let signing_key = SigningKey::from_bytes((&[0x11u8; 32]).into()).unwrap();
        let verifying_key = signing_key.verifying_key();
        let expected = pubkey_to_eth_address(verifying_key);
        let message_hash = [0x42u8; 32];
        let (sig, recid) = signing_key.sign_prehash(&message_hash).unwrap();
        let mut sig_bytes = [0u8; 65];
        sig_bytes[..64].copy_from_slice(&sig.to_bytes());
        sig_bytes[64] = recid.to_byte();
        let recovered = recover_eth_address(message_hash, &sig_bytes).unwrap();
        assert_eq!(recovered, expected);
    }

    #[test]
    fn test_verify_signature_correct_signer() {
        let signing_key = SigningKey::from_bytes((&[0x11u8; 32]).into()).unwrap();
        let expected = pubkey_to_eth_address(signing_key.verifying_key());
        let message_hash = [0x24u8; 32];
        let (sig, recid) = signing_key.sign_prehash(&message_hash).unwrap();
        let mut sig_bytes = [0u8; 65];
        sig_bytes[..64].copy_from_slice(&sig.to_bytes());
        sig_bytes[64] = recid.to_byte();
        assert!(verify_signature(message_hash, &sig_bytes, expected));
    }

    #[test]
    fn test_verify_signature_wrong_signer() {
        let signing_key = SigningKey::from_bytes((&[0x11u8; 32]).into()).unwrap();
        let message_hash = [0x24u8; 32];
        let (sig, recid) = signing_key.sign_prehash(&message_hash).unwrap();
        let mut sig_bytes = [0u8; 65];
        sig_bytes[..64].copy_from_slice(&sig.to_bytes());
        sig_bytes[64] = recid.to_byte();
        assert!(!verify_signature(message_hash, &sig_bytes, [0u8; 20]));
    }

    #[test]
    fn test_recover_invalid_v_value() {
        let hash = [0u8; 32];
        let mut sig = [0u8; 65];
        sig[64] = 5;
        assert!(recover_eth_address(hash, &sig).is_none());
    }
}
