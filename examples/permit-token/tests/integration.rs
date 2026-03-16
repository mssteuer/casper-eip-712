use casper_eip_712::prelude::*;
use k256::ecdsa::{signature::hazmat::PrehashSigner, SigningKey};
use odra::casper_types::{bytesrepr::Bytes, U256};
use odra::host::{Deployer, HostEnv};
use odra::prelude::{Address, Addressable};
use permit_token::permit_token::{address_to_eth_bytes, u256_to_bytes32, PermitToken, PermitTokenHostRef, PermitTokenInitArgs, CHAIN_ID_CASPER_ASCII};

const TOKEN_NAME: &str = "PermitToken";
const TOKEN_SYMBOL: &str = "PTK";
const TOKEN_DECIMALS: u8 = 18;
const INITIAL_SUPPLY: u64 = 1_000_000;
const DOMAIN_VERSION: &str = "1";

fn setup() -> (HostEnv, PermitTokenHostRef) {
    let env = odra_test::env();
    let token = PermitToken::deploy(
        &env,
        PermitTokenInitArgs {
            name: TOKEN_NAME.into(),
            symbol: TOKEN_SYMBOL.into(),
            decimals: TOKEN_DECIMALS,
            initial_supply: INITIAL_SUPPLY.into(),
            domain_version: DOMAIN_VERSION.into(),
        },
    );
    (env, token)
}

fn test_keypair(seed: u8) -> (SigningKey, [u8; 20]) {
    let key = SigningKey::from_bytes((&[seed; 32]).into()).unwrap();
    let verifying = key.verifying_key();
    let encoded = verifying.to_encoded_point(false);
    let bytes = encoded.as_bytes();
    let hash = casper_eip_712::keccak::keccak256(&bytes[1..]);
    let mut addr = [0u8; 20];
    addr.copy_from_slice(&hash[12..32]);
    (key, addr)
}

fn sign_digest(key: &SigningKey, digest: [u8; 32]) -> Vec<u8> {
    let (sig, recid) = key.sign_prehash(&digest).unwrap();
    let mut sig_bytes = vec![0u8; 65];
    sig_bytes[..64].copy_from_slice(&sig.to_bytes());
    sig_bytes[64] = recid.to_byte();
    sig_bytes
}

fn contract_bytes(addr: Address) -> [u8; 32] {
    match addr {
        Address::Contract(hash) => hash.value(),
        Address::Account(hash) => hash.value(),
    }
}

fn evm_domain(token: &PermitTokenHostRef) -> DomainSeparator {
    let contract = contract_bytes(token.address());
    let mut verifying_contract = [0u8; 20];
    verifying_contract.copy_from_slice(&contract[..20]);
    DomainBuilder::new()
        .name(TOKEN_NAME)
        .version(DOMAIN_VERSION)
        .chain_id(CHAIN_ID_CASPER_ASCII)
        .verifying_contract(verifying_contract)
        .build()
}

fn casper_domain(token: &PermitTokenHostRef) -> DomainSeparator {
    DomainBuilder::new()
        .name(TOKEN_NAME)
        .version(DOMAIN_VERSION)
        .custom_field("chain_name", DomainFieldValue::String("casper".into()))
        .custom_field("contract_package_hash", DomainFieldValue::Bytes32(contract_bytes(token.address())))
        .build()
}

fn build_permit(owner: [u8; 20], spender: &Address, value: U256, nonce: U256, deadline: u64) -> Permit {
    Permit {
        owner,
        spender: address_to_eth_bytes(spender),
        value: u256_to_bytes32(value),
        nonce: u256_to_bytes32(nonce),
        deadline: u256_to_bytes32(U256::from(deadline)),
    }
}

#[test]
fn test_permit_sets_allowance() {
    let (env, mut token) = setup();
    let (key, owner_eth) = test_keypair(0x11);
    let spender = env.get_account(1);
    let owner_proxy = token.permit_owner_proxy(Bytes::from(owner_eth.to_vec()));
    token.mint_to(owner_proxy, U256::from(5_000u64));

    let value = U256::from(1_000u64);
    let nonce = U256::zero();
    let deadline = u64::MAX;
    let signature = sign_digest(&key, hash_typed_data(&evm_domain(&token), &build_permit(owner_eth, &spender, value, nonce, deadline)));

    token.permit(Bytes::from(owner_eth.to_vec()), spender, value, nonce, deadline, Bytes::from(signature), false);
    assert_eq!(token.allowance(&owner_proxy, &spender), value);
    assert_eq!(token.nonces(&owner_proxy), U256::one());
}

#[test]
fn test_permit_wrong_signer_reverts() {
    let (env, mut token) = setup();
    let (wrong_key, _) = test_keypair(0x22);
    let (_, claimed_owner) = test_keypair(0x11);
    let spender = env.get_account(1);
    let value = U256::from(100u64);
    let nonce = U256::zero();
    let deadline = u64::MAX;
    let signature = sign_digest(&wrong_key, hash_typed_data(&evm_domain(&token), &build_permit(claimed_owner, &spender, value, nonce, deadline)));

    assert!(token.try_permit(Bytes::from(claimed_owner.to_vec()), spender, value, nonce, deadline, Bytes::from(signature), false).is_err());
}

#[test]
fn test_permit_expired_deadline_reverts() {
    let (env, mut token) = setup();
    let (key, owner_eth) = test_keypair(0x33);
    let spender = env.get_account(1);
    let value = U256::from(100u64);
    let nonce = U256::zero();
    env.advance_block_time(1);
    // The block time is now 1, so a deadline of 0 is already expired.
    let deadline = 0u64;
    let signature = sign_digest(&key, hash_typed_data(&evm_domain(&token), &build_permit(owner_eth, &spender, value, nonce, deadline)));

    assert!(token.try_permit(Bytes::from(owner_eth.to_vec()), spender, value, nonce, deadline, Bytes::from(signature), false).is_err());
}

#[test]
fn test_permit_replayed_nonce_reverts() {
    let (env, mut token) = setup();
    let (key, owner_eth) = test_keypair(0x44);
    let spender = env.get_account(1);
    let value = U256::from(100u64);
    let nonce = U256::zero();
    let deadline = u64::MAX;
    let signature = sign_digest(&key, hash_typed_data(&evm_domain(&token), &build_permit(owner_eth, &spender, value, nonce, deadline)));

    token.permit(Bytes::from(owner_eth.to_vec()), spender, value, nonce, deadline, Bytes::from(signature.clone()), false);
    assert!(token.try_permit(Bytes::from(owner_eth.to_vec()), spender, value, nonce, deadline, Bytes::from(signature), false).is_err());
}

#[test]
fn test_permit_then_transfer_from() {
    let (env, mut token) = setup();
    let (key, owner_eth) = test_keypair(0x55);
    let spender = env.get_account(1);
    let recipient = env.get_account(2);
    let owner_proxy = token.permit_owner_proxy(Bytes::from(owner_eth.to_vec()));
    token.mint_to(owner_proxy, U256::from(2_000u64));

    let value = U256::from(750u64);
    let nonce = U256::zero();
    let deadline = u64::MAX;
    let signature = sign_digest(&key, hash_typed_data(&evm_domain(&token), &build_permit(owner_eth, &spender, value, nonce, deadline)));
    token.permit(Bytes::from(owner_eth.to_vec()), spender, value, nonce, deadline, Bytes::from(signature), false);

    env.set_caller(spender);
    token.transfer_from(&owner_proxy, &recipient, &U256::from(300u64));

    assert_eq!(token.balance_of(&recipient), U256::from(300u64));
    assert_eq!(token.balance_of(&owner_proxy), U256::from(1_700u64));
    assert_eq!(token.allowance(&owner_proxy, &spender), U256::from(450u64));
}

#[test]
fn test_evm_and_casper_domain_both_work() {
    let (env, mut token) = setup();
    let (key, owner_eth) = test_keypair(0x66);
    let spender = env.get_account(1);
    let owner_proxy = token.permit_owner_proxy(Bytes::from(owner_eth.to_vec()));
    token.mint_to(owner_proxy, U256::from(1_000u64));

    let deadline = u64::MAX;
    let sig1 = sign_digest(&key, hash_typed_data(&evm_domain(&token), &build_permit(owner_eth, &spender, U256::from(100u64), U256::zero(), deadline)));
    token.permit(Bytes::from(owner_eth.to_vec()), spender, U256::from(100u64), U256::zero(), deadline, Bytes::from(sig1), false);

    let sig2 = sign_digest(&key, hash_typed_data(&casper_domain(&token), &build_permit(owner_eth, &spender, U256::from(200u64), U256::one(), deadline)));
    token.permit(Bytes::from(owner_eth.to_vec()), spender, U256::from(200u64), U256::one(), deadline, Bytes::from(sig2), true);

    assert_eq!(token.allowance(&owner_proxy, &spender), U256::from(200u64));
    assert_eq!(token.nonces(&owner_proxy), U256::from(2u8));
}
