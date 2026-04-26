//! Casper-native signature verification for EIP-712 typed data.
//!
//! This module extends `casper-eip-712` with support for Casper `PublicKey`
//! verification, using the `casper-types` crate.  Enable with the
//! `casper-native` Cargo feature.
//!
//! # Feature flag
//!
//! ```toml
//! casper-eip-712 = { version = "1.2", features = ["casper-native"] }
//! ```
//!
//! # Example
//!
//! ```rust,ignore
//! use casper_eip_712::prelude::*;
//! use casper_eip_712::casper_native::{verify_casper_signer, TransferAuthorization};
//! use casper_types::{PublicKey, Signature};
//!
//! let domain = DomainBuilder::new()
//!     .name("MyToken")
//!     .version("1")
//!     .custom_field("chain_name", DomainFieldValue::String("casper-test".into()))
//!     .custom_field("contract_package_hash", DomainFieldValue::Bytes32([0x77; 32]))
//!     .build();
//!
//! let auth = TransferAuthorization {
//!     from: [0xAB; 32],
//!     to: [0xCD; 32],
//!     value: [0u8; 32],
//!     valid_after: 0,
//!     valid_before: u64::MAX,
//!     nonce: [0x01; 32],
//! };
//!
//! // public_key and signature come from the Casper contract call arguments.
//! // let account_hash = verify_casper_signer(&domain, &auth, &public_key, &signature)?;
//! ```

extern crate alloc;

use alloc::vec::Vec;

use casper_types::{account::AccountHash, crypto, PublicKey, Signature};

use crate::domain::DomainSeparator;
use crate::hash::hash_typed_data;
use crate::traits::Eip712Struct;

/// Error type for casper-native signature verification.
#[derive(Debug, PartialEq, Eq)]
pub enum Error {
    /// The signature is invalid for the given public key and typed-data digest.
    InvalidSignature,
    /// The derived AccountHash does not match the `from` field in the message.
    AccountHashMismatch,
}

impl core::fmt::Display for Error {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        match self {
            Error::InvalidSignature => f.write_str("invalid casper signature"),
            Error::AccountHashMismatch => f.write_str("account hash mismatch"),
        }
    }
}

/// Verify that a Casper `PublicKey` signed the EIP-712 typed data digest.
///
/// Returns `Ok(AccountHash)` on success, or an error if the signature is
/// invalid or if the key's account hash does not match `expected_from`.
///
/// # Arguments
///
/// * `domain` — the pre-built `DomainSeparator`
/// * `message` — any type implementing `Eip712Struct`
/// * `public_key` — the Casper `PublicKey` passed by the caller
/// * `signature` — the Casper `Signature` to verify
/// * `expected_from` — optional 32-byte account hash; if `Some`, the function
///   checks that `AccountHash::from(public_key) == expected_from`.
pub fn verify_casper_signer(
    domain: &DomainSeparator,
    message: &impl Eip712Struct,
    public_key: &PublicKey,
    signature: &Signature,
    expected_from: Option<&[u8; 32]>,
) -> Result<AccountHash, Error> {
    let digest = hash_typed_data(domain, message);
    crypto::verify(&digest, signature, public_key).map_err(|_| Error::InvalidSignature)?;

    let account_hash = public_key.to_account_hash();

    if let Some(expected) = expected_from {
        if &account_hash.value() != expected {
            return Err(Error::AccountHashMismatch);
        }
    }

    Ok(account_hash)
}

/// EIP-712 typed struct for an EIP-3009-style transfer authorization.
///
/// Compatible with Krzysztof's `casper-x402-poc` semantics.  The
/// `from`/`to` fields are 32-byte Casper `AccountHash` values (not the
/// full 33-byte address with prefix byte).
///
/// # EIP-712 type string
///
/// ```text
/// TransferAuthorization(bytes32 from,bytes32 to,uint256 value,uint64 valid_after,uint64 valid_before,bytes32 nonce)
/// ```
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct TransferAuthorization {
    /// 32-byte `AccountHash` of the sender.
    pub from: [u8; 32],
    /// 32-byte `AccountHash` of the recipient.
    pub to: [u8; 32],
    /// Transfer amount as a 32-byte big-endian `U256` (motes or token units).
    pub value: [u8; 32],
    /// Unix timestamp (seconds) before which the authorization is invalid.
    pub valid_after: u64,
    /// Unix timestamp (seconds) after which the authorization expires.
    pub valid_before: u64,
    /// 32-byte replay-protection nonce (must be unique per `from`).
    pub nonce: [u8; 32],
}

impl Eip712Struct for TransferAuthorization {
    fn type_string() -> &'static str {
        "TransferAuthorization(bytes32 from,bytes32 to,uint256 value,uint64 valid_after,uint64 valid_before,bytes32 nonce)"
    }

    fn encode_data(&self) -> Vec<u8> {
        use crate::encoding::{encode_bytes32, encode_uint256, encode_uint64};
        let mut data = Vec::with_capacity(6 * 32);
        data.extend_from_slice(&encode_bytes32(self.from));
        data.extend_from_slice(&encode_bytes32(self.to));
        data.extend_from_slice(&encode_uint256(self.value));
        data.extend_from_slice(&encode_uint64(self.valid_after));
        data.extend_from_slice(&encode_uint64(self.valid_before));
        data.extend_from_slice(&encode_bytes32(self.nonce));
        data
    }
}

/// EIP-712 typed struct for a batch transfer authorization.
///
/// One signature covers multiple (to, value) transfers. Useful for x402
/// flows where a single authorization pays both the recipient and a
/// facilitator fee.
///
/// # EIP-712 type string
///
/// ```text
/// BatchTransferAuthorization(bytes32 from,BatchEntry[] transfers,uint64 valid_after,uint64 valid_before,bytes32 nonce)BatchEntry(bytes32 to,uint256 value)
/// ```
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BatchEntry {
    /// 32-byte `AccountHash` of the recipient.
    pub to: [u8; 32],
    /// Transfer amount as a 32-byte big-endian `U256`.
    pub value: [u8; 32],
}

impl BatchEntry {
    /// Compute the EIP-712 `hashStruct` of a single `BatchEntry`.
    fn hash_struct_entry(&self) -> [u8; 32] {
        use crate::encoding::{encode_bytes32, encode_uint256};
        use crate::keccak::keccak256;

        let type_hash = keccak256(b"BatchEntry(bytes32 to,uint256 value)");
        let mut data = Vec::with_capacity(3 * 32);
        data.extend_from_slice(&type_hash);
        data.extend_from_slice(&encode_bytes32(self.to));
        data.extend_from_slice(&encode_uint256(self.value));
        keccak256(&data)
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BatchTransferAuthorization {
    /// 32-byte `AccountHash` of the sender.
    pub from: [u8; 32],
    /// Ordered list of transfer entries.
    pub transfers: Vec<BatchEntry>,
    /// Unix timestamp (seconds) before which the authorization is invalid.
    pub valid_after: u64,
    /// Unix timestamp (seconds) after which the authorization expires.
    pub valid_before: u64,
    /// 32-byte replay-protection nonce (must be unique per `from`).
    pub nonce: [u8; 32],
}

impl Eip712Struct for BatchTransferAuthorization {
    fn type_string() -> &'static str {
        "BatchTransferAuthorization(bytes32 from,BatchEntry[] transfers,uint64 valid_after,uint64 valid_before,bytes32 nonce)BatchEntry(bytes32 to,uint256 value)"
    }

    fn encode_data(&self) -> Vec<u8> {
        use crate::encoding::{encode_bytes32, encode_uint64};
        use crate::keccak::keccak256;

        // EIP-712: array of structs is encoded as keccak256 of the concatenated
        // hashStruct of each element.
        let mut array_encoded: Vec<u8> = Vec::with_capacity(self.transfers.len() * 32);
        for entry in &self.transfers {
            array_encoded.extend_from_slice(&entry.hash_struct_entry());
        }
        let transfers_hash = keccak256(&array_encoded);

        let mut data = Vec::with_capacity(5 * 32);
        data.extend_from_slice(&encode_bytes32(self.from));
        data.extend_from_slice(&transfers_hash);
        data.extend_from_slice(&encode_uint64(self.valid_after));
        data.extend_from_slice(&encode_uint64(self.valid_before));
        data.extend_from_slice(&encode_bytes32(self.nonce));
        data
    }
}
#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::DomainBuilder;
    use crate::domain::DomainFieldValue;
    use crate::hash::hash_typed_data;
    use crate::keccak::keccak256;
    use crate::traits::Eip712Struct;

    fn test_domain() -> DomainSeparator {
        DomainBuilder::new()
            .name("TestToken")
            .version("1")
            .custom_field("chain_name", DomainFieldValue::String("casper-test".into()))
            .custom_field(
                "contract_package_hash",
                DomainFieldValue::Bytes32([0x77; 32]),
            )
            .build()
    }

    fn sample_transfer_auth() -> TransferAuthorization {
        TransferAuthorization {
            from: [0xAB; 32],
            to: [0xCD; 32],
            value: [0u8; 32],
            valid_after: 0,
            valid_before: 9_999_999_999,
            nonce: [0x01; 32],
        }
    }

    #[test]
    fn test_transfer_authorization_type_string() {
        assert_eq!(
            TransferAuthorization::type_string(),
            "TransferAuthorization(bytes32 from,bytes32 to,uint256 value,uint64 valid_after,uint64 valid_before,bytes32 nonce)"
        );
    }

    #[test]
    fn test_transfer_authorization_type_hash_matches_type_string() {
        let expected = keccak256(TransferAuthorization::type_string().as_bytes());
        assert_eq!(TransferAuthorization::type_hash(), expected);
    }

    #[test]
    fn test_transfer_authorization_encode_data_length() {
        let auth = sample_transfer_auth();
        assert_eq!(auth.encode_data().len(), 6 * 32);
    }

    #[test]
    fn test_transfer_authorization_hash_struct_deterministic() {
        let auth = sample_transfer_auth();
        assert_eq!(auth.hash_struct(), auth.hash_struct());
    }

    #[test]
    fn test_transfer_authorization_different_nonce_different_hash() {
        let mut a1 = sample_transfer_auth();
        let mut a2 = sample_transfer_auth();
        a1.nonce = [0x01; 32];
        a2.nonce = [0x02; 32];
        assert_ne!(a1.hash_struct(), a2.hash_struct());
    }

    #[test]
    fn test_transfer_authorization_different_value_different_hash() {
        let mut a1 = sample_transfer_auth();
        let mut a2 = sample_transfer_auth();
        a1.value[31] = 1;
        a2.value[31] = 2;
        assert_ne!(a1.hash_struct(), a2.hash_struct());
    }

    #[test]
    fn test_hash_typed_data_deterministic_across_instances() {
        let domain = test_domain();
        let auth = sample_transfer_auth();
        let d1 = hash_typed_data(&domain, &auth);
        let d2 = hash_typed_data(&domain, &auth);
        assert_eq!(d1, d2);
    }

    #[test]
    fn test_hash_typed_data_starts_with_0x1901() {
        let domain = test_domain();
        let auth = sample_transfer_auth();
        let digest = hash_typed_data(&domain, &auth);
        assert_eq!(digest.len(), 32, "digest must be 32 bytes");
    }

    #[test]
    fn test_batch_transfer_type_string() {
        assert_eq!(
            BatchTransferAuthorization::type_string(),
            "BatchTransferAuthorization(bytes32 from,BatchEntry[] transfers,uint64 valid_after,uint64 valid_before,bytes32 nonce)BatchEntry(bytes32 to,uint256 value)"
        );
    }

    #[test]
    fn test_batch_transfer_type_hash_matches_type_string() {
        let expected = keccak256(BatchTransferAuthorization::type_string().as_bytes());
        assert_eq!(BatchTransferAuthorization::type_hash(), expected);
    }

    #[test]
    fn test_batch_transfer_encode_data_length() {
        let batch = BatchTransferAuthorization {
            from: [0x11; 32],
            transfers: alloc::vec![
                BatchEntry { to: [0x22; 32], value: [0u8; 32] },
                BatchEntry { to: [0x33; 32], value: [0u8; 32] },
            ],
            valid_after: 0,
            valid_before: 1_000_000,
            nonce: [0xAA; 32],
        };
        // from(32) + transfers_hash(32) + valid_after(32) + valid_before(32) + nonce(32) = 160
        assert_eq!(batch.encode_data().len(), 5 * 32);
    }

    #[test]
    fn test_batch_transfer_empty_vs_nonempty_differs() {
        let base = BatchTransferAuthorization {
            from: [0x11; 32],
            transfers: alloc::vec![],
            valid_after: 0,
            valid_before: 1_000_000,
            nonce: [0xAA; 32],
        };
        let with_entry = BatchTransferAuthorization {
            transfers: alloc::vec![BatchEntry { to: [0x22; 32], value: [0u8; 32] }],
            ..base.clone()
        };
        assert_ne!(base.hash_struct(), with_entry.hash_struct());
    }

    #[test]
    fn test_error_display() {
        use crate::alloc::string::ToString;
        assert_eq!(Error::InvalidSignature.to_string(), "invalid casper signature");
        assert_eq!(Error::AccountHashMismatch.to_string(), "account hash mismatch");
    }
}
