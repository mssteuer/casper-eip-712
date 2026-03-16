use alloc::string::{String, ToString};
use alloc::vec::Vec;

use crate::encoding::{encode_address, encode_bool, encode_bytes, encode_bytes32, encode_string, encode_uint64, encode_uint256};
use crate::keccak::keccak256;

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum DomainFieldValue {
    String(String),
    Uint256([u8; 32]),
    Uint64(u64),
    Address([u8; 20]),
    Bytes32([u8; 32]),
    Bytes(Vec<u8>),
    Bool(bool),
}

impl DomainFieldValue {
    fn type_name(&self) -> &'static str {
        match self {
            Self::String(_) => "string",
            Self::Uint256(_) | Self::Uint64(_) => "uint256",
            Self::Address(_) => "address",
            Self::Bytes32(_) => "bytes32",
            Self::Bytes(_) => "bytes",
            Self::Bool(_) => "bool",
        }
    }

    fn encode(&self) -> [u8; 32] {
        match self {
            Self::String(value) => encode_string(value),
            Self::Uint256(value) => encode_uint256(*value),
            Self::Uint64(value) => encode_uint64(*value),
            Self::Address(value) => encode_address(*value),
            Self::Bytes32(value) => encode_bytes32(*value),
            Self::Bytes(value) => encode_bytes(value),
            Self::Bool(value) => encode_bool(*value),
        }
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct DomainSeparator {
    type_hash: [u8; 32],
    separator_hash: [u8; 32],
    type_string: String,
}

impl DomainSeparator {
    pub fn type_hash(&self) -> [u8; 32] {
        self.type_hash
    }

    pub fn separator_hash(&self) -> [u8; 32] {
        self.separator_hash
    }

    pub fn type_string(&self) -> &str {
        &self.type_string
    }
}

#[derive(Default, Clone, Debug)]
pub struct DomainBuilder {
    name: Option<String>,
    version: Option<String>,
    chain_id: Option<[u8; 32]>,
    verifying_contract: Option<[u8; 20]>,
    salt: Option<[u8; 32]>,
    custom_fields: Vec<(String, DomainFieldValue)>,
}

impl DomainBuilder {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn name(mut self, value: &str) -> Self {
        self.name = Some(value.to_string());
        self
    }

    pub fn version(mut self, value: &str) -> Self {
        self.version = Some(value.to_string());
        self
    }

    pub fn chain_id(mut self, value: u64) -> Self {
        self.chain_id = Some(encode_uint64(value));
        self
    }

    pub fn chain_id_bytes(mut self, value: [u8; 32]) -> Self {
        self.chain_id = Some(value);
        self
    }

    pub fn verifying_contract(mut self, value: [u8; 20]) -> Self {
        self.verifying_contract = Some(value);
        self
    }

    pub fn salt(mut self, value: [u8; 32]) -> Self {
        self.salt = Some(value);
        self
    }

    pub fn custom_field(mut self, name: &str, value: DomainFieldValue) -> Self {
        self.custom_fields.push((name.to_string(), value));
        self
    }

    pub fn build(self) -> DomainSeparator {
        let mut fields: Vec<(&'static str, String, DomainFieldValue)> = Vec::new();

        if let Some(value) = self.name {
            fields.push(("string", "name".to_string(), DomainFieldValue::String(value)));
        }
        if let Some(value) = self.version {
            fields.push(("string", "version".to_string(), DomainFieldValue::String(value)));
        }
        if let Some(value) = self.chain_id {
            fields.push(("uint256", "chainId".to_string(), DomainFieldValue::Uint256(value)));
        }
        if let Some(value) = self.verifying_contract {
            fields.push(("address", "verifyingContract".to_string(), DomainFieldValue::Address(value)));
        }
        if let Some(value) = self.salt {
            fields.push(("bytes32", "salt".to_string(), DomainFieldValue::Bytes32(value)));
        }
        for (name, value) in self.custom_fields {
            fields.push((value.type_name(), name, value));
        }

        let mut type_string = String::from("EIP712Domain(");
        for (index, (field_type, field_name, _)) in fields.iter().enumerate() {
            if index > 0 {
                type_string.push(',');
            }
            type_string.push_str(field_type);
            type_string.push(' ');
            type_string.push_str(field_name);
        }
        type_string.push(')');

        let type_hash = keccak256(type_string.as_bytes());
        let mut encoded = Vec::with_capacity(32 + fields.len() * 32);
        encoded.extend_from_slice(&type_hash);
        for (_, _, value) in &fields {
            encoded.extend_from_slice(&value.encode());
        }
        let separator_hash = keccak256(&encoded);

        DomainSeparator {
            type_hash,
            separator_hash,
            type_string,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::keccak::keccak256;

    #[test]
    fn test_standard_evm_domain_type_hash() {
        let domain = DomainBuilder::new()
            .name("Test")
            .version("1")
            .chain_id(1)
            .verifying_contract([0x11; 20])
            .build();
        let expected_type_hash = keccak256(
            b"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
        );
        assert_eq!(domain.type_hash(), expected_type_hash);
    }

    #[test]
    fn test_partial_domain_type_string() {
        let domain = DomainBuilder::new().name("Test").version("1").build();
        assert_eq!(domain.type_string(), "EIP712Domain(string name,string version)");
    }

    #[test]
    fn test_custom_fields_are_appended() {
        let domain = DomainBuilder::new()
            .name("Test")
            .custom_field("chain_name", DomainFieldValue::String("casper-test".into()))
            .custom_field("contract_package_hash", DomainFieldValue::Bytes32([0x77; 32]))
            .build();
        assert_eq!(
            domain.type_string(),
            "EIP712Domain(string name,string chain_name,bytes32 contract_package_hash)"
        );
    }

    #[test]
    fn test_minimal_domain() {
        let domain = DomainBuilder::new().name("OnlyName").build();
        assert_eq!(domain.type_string(), "EIP712Domain(string name)");
    }

    #[test]
    fn test_different_domains_different_hashes() {
        let d1 = DomainBuilder::new().name("A").version("1").build();
        let d2 = DomainBuilder::new().name("B").version("1").build();
        assert_ne!(d1.separator_hash(), d2.separator_hash());
    }

    #[test]
    fn test_deterministic_separator_hash() {
        let d1 = DomainBuilder::new().name("A").version("1").build();
        let d2 = DomainBuilder::new().name("A").version("1").build();
        assert_eq!(d1.separator_hash(), d2.separator_hash());
    }
}
