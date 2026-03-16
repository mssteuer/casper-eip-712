use casper_eip_712::prelude::*;
use hex::FromHex;
use serde::Deserialize;
use serde_json::Value;

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct VectorsFile {
    vectors: Vec<TestVector>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct TestVector {
    name: String,
    primary_type: String,
    domain: Value,
    message: Value,
    domain_separator: String,
    struct_hash: String,
    digest: String,
}

fn parse_hex_array<const N: usize>(value: &str) -> [u8; N] {
    let trimmed = value.strip_prefix("0x").unwrap_or(value);
    let bytes = <Vec<u8>>::from_hex(trimmed).expect("valid hex input");
    let array: [u8; N] = bytes.try_into().expect("hex input has expected byte length");
    array
}

fn get_str<'a>(value: &'a Value, field: &str) -> &'a str {
    value.get(field)
        .and_then(Value::as_str)
        .unwrap_or_else(|| panic!("missing string field {field}"))
}

fn build_domain(domain: &Value) -> DomainSeparator {
    let mut builder = DomainBuilder::new();

    if let Some(name) = domain.get("name").and_then(Value::as_str) {
        builder = builder.name(name);
    }
    if let Some(version) = domain.get("version").and_then(Value::as_str) {
        builder = builder.version(version);
    }
    if let Some(chain_id) = domain.get("chainId").and_then(Value::as_u64) {
        builder = builder.chain_id(chain_id);
    }
    if let Some(verifying_contract) = domain.get("verifyingContract").and_then(Value::as_str) {
        builder = builder.verifying_contract(parse_hex_array::<20>(verifying_contract));
    }
    if let Some(salt) = domain.get("salt").and_then(Value::as_str) {
        builder = builder.salt(parse_hex_array::<32>(salt));
    }
    // This test helper only wires the custom domain fields covered by the current
    // committed vectors. If future vectors add new custom domain fields, extend this
    // mapping so they are not silently ignored by the builder.
    if let Some(chain_name) = domain.get("chain_name").and_then(Value::as_str) {
        builder = builder.custom_field("chain_name", DomainFieldValue::String(chain_name.into()));
    }
    if let Some(contract_package_hash) = domain.get("contract_package_hash").and_then(Value::as_str) {
        builder = builder.custom_field(
            "contract_package_hash",
            DomainFieldValue::Bytes32(parse_hex_array::<32>(contract_package_hash)),
        );
    }

    builder.build()
}

#[test]
fn vectors_match_ethers_reference_values() {
    let vectors: VectorsFile = serde_json::from_str(include_str!("vectors.json")).expect("vectors.json parses");
    assert!(vectors.vectors.len() >= 6, "expected expanded cross-language coverage");

    for vector in vectors.vectors {
        let domain = build_domain(&vector.domain);

        let actual_struct_hash = match vector.primary_type.as_str() {
            "Permit" => {
                let permit = Permit {
                    owner: parse_hex_array::<20>(get_str(&vector.message, "owner")),
                    spender: parse_hex_array::<20>(get_str(&vector.message, "spender")),
                    value: parse_hex_array::<32>(get_str(&vector.message, "value")),
                    nonce: parse_hex_array::<32>(get_str(&vector.message, "nonce")),
                    deadline: parse_hex_array::<32>(get_str(&vector.message, "deadline")),
                };

                assert_eq!(
                    hash_typed_data(&domain, &permit),
                    parse_hex_array::<32>(&vector.digest),
                    "typed data digest mismatch for {}",
                    vector.name
                );

                permit.hash_struct()
            }
            "Approval" => {
                let approval = Approval {
                    owner: parse_hex_array::<20>(get_str(&vector.message, "owner")),
                    spender: parse_hex_array::<20>(get_str(&vector.message, "spender")),
                    value: parse_hex_array::<32>(get_str(&vector.message, "value")),
                };

                assert_eq!(
                    hash_typed_data(&domain, &approval),
                    parse_hex_array::<32>(&vector.digest),
                    "typed data digest mismatch for {}",
                    vector.name
                );

                approval.hash_struct()
            }
            "Transfer" => {
                let transfer = Transfer {
                    from: parse_hex_array::<20>(get_str(&vector.message, "from")),
                    to: parse_hex_array::<20>(get_str(&vector.message, "to")),
                    value: parse_hex_array::<32>(get_str(&vector.message, "value")),
                };

                assert_eq!(
                    hash_typed_data(&domain, &transfer),
                    parse_hex_array::<32>(&vector.digest),
                    "typed data digest mismatch for {}",
                    vector.name
                );

                transfer.hash_struct()
            }
            other => panic!("unsupported primary type {other} in {}", vector.name),
        };

        assert_eq!(
            domain.separator_hash(),
            parse_hex_array::<32>(&vector.domain_separator),
            "domain separator mismatch for {}",
            vector.name
        );
        assert_eq!(
            actual_struct_hash,
            parse_hex_array::<32>(&vector.struct_hash),
            "struct hash mismatch for {}",
            vector.name
        );
    }
}

#[test]
fn casper_native_domain_type_string_matches_builder_order() {
    let vectors: VectorsFile = serde_json::from_str(include_str!("vectors.json")).expect("vectors.json parses");
    let vector = vectors
        .vectors
        .into_iter()
        .find(|entry| entry.name == "casper_native_domain_permit")
        .expect("casper-native vector exists");

    let domain = build_domain(&vector.domain);
    assert_eq!(
        domain.type_string(),
        "EIP712Domain(string name,string version,string chain_name,bytes32 contract_package_hash)"
    );
}

#[test]
fn partial_domain_vector_omits_chain_fields() {
    let vectors: VectorsFile = serde_json::from_str(include_str!("vectors.json")).expect("vectors.json parses");
    let vector = vectors
        .vectors
        .into_iter()
        .find(|entry| entry.name == "permit_partial_domain")
        .expect("partial domain vector exists");

    let domain = build_domain(&vector.domain);
    assert_eq!(domain.type_string(), "EIP712Domain(string name,string version)");
}
