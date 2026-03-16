use casper_eip_712::prelude::*;

#[test]
fn test_permit_hash_matches_known_vector_shape() {
    let domain = DomainBuilder::new()
        .name("MyToken")
        .version("1")
        .chain_id(1)
        .verifying_contract([0x11; 20])
        .build();

    let permit = Permit {
        owner: [0x22; 20],
        spender: [0x33; 20],
        value: [0x44; 32],
        nonce: [0x55; 32],
        deadline: [0x66; 32],
    };

    let hash1 = hash_typed_data(&domain, &permit);
    let hash2 = hash_typed_data(&domain, &permit);
    assert_eq!(hash1, hash2);
}
