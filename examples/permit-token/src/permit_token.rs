extern crate alloc;

use alloc::string::{String, ToString};
use alloc::vec::Vec;

use casper_eip_712::prelude::*;
use casper_eip_712::verify::recover_eth_address;
use odra::casper_types::{account::AccountHash, bytesrepr::Bytes, U256};
use odra::prelude::*;
// Explicit import to resolve ambiguity: both casper_eip_712::prelude and odra::prelude export `Address`.
// The Odra Address (with Account/Contract variants) is the one used for Casper runtime addresses.
use odra::prelude::Address;
use odra_modules::cep18_token::Cep18;

pub const CHAIN_ID_CASPER_ASCII: u64 = 1_314_614_895;

#[derive(PartialEq, Eq, Debug)]
#[odra::odra_error]
pub enum Error {
    PermitExpired = 30_000,
    InvalidNonce = 30_001,
    InvalidSignature = 30_002,
    MalformedSignature = 30_003,
    InvalidOwnerEthAddress = 30_004,
    UnauthorizedMint = 30_005,
}

#[odra::module(errors = Error)]
pub struct PermitToken {
    token: SubModule<Cep18>,
    nonces: Mapping<Address, U256>,
    permit_allowances: Mapping<(Address, Address), U256>,
    domain_name: Var<String>,
    domain_version: Var<String>,
}

#[odra::module]
impl PermitToken {
    pub fn init(
        &mut self,
        name: String,
        symbol: String,
        decimals: u8,
        initial_supply: U256,
        domain_version: String,
    ) {
        self.token.init(symbol, name.clone(), decimals, initial_supply);
        self.domain_name.set(name);
        self.domain_version.set(domain_version);
    }

    delegate! {
        to self.token {
            fn name(&self) -> String;
            fn symbol(&self) -> String;
            fn decimals(&self) -> u8;
            fn total_supply(&self) -> U256;
            fn balance_of(&self, address: &Address) -> U256;
            fn approve(&mut self, spender: &Address, amount: &U256);
            fn decrease_allowance(&mut self, spender: &Address, decr_by: &U256);
            fn increase_allowance(&mut self, spender: &Address, inc_by: &U256);
            fn transfer(&mut self, recipient: &Address, amount: &U256);
        }
    }

    pub fn allowance(&self, owner: &Address, spender: &Address) -> U256 {
        let permit = self.permit_allowances.get(&(*owner, *spender)).unwrap_or_default();
        if permit > U256::zero() {
            // Permit-set allowances override CEP-18 approvals so off-chain signatures remain authoritative.
            permit
        } else {
            self.token.allowance(owner, spender)
        }
    }

    pub fn permit(
        &mut self,
        owner_eth_address: Bytes,
        spender: Address,
        value: U256,
        nonce: U256,
        deadline: u64,
        signature: Bytes,
        use_casper_domain: bool,
    ) {
        if deadline < self.current_block_time_u64() {
            self.env().revert(Error::PermitExpired);
        }
        if owner_eth_address.len() != 20 {
            self.env().revert(Error::InvalidOwnerEthAddress);
        }
        if signature.len() != 65 {
            self.env().revert(Error::MalformedSignature);
        }

        let owner_eth = slice_to_20(&owner_eth_address).unwrap_or_revert_with(self, Error::InvalidOwnerEthAddress);
        let owner = owner_proxy_address(owner_eth);
        let current_nonce = self.nonces.get(&owner).unwrap_or_default();
        if nonce != current_nonce {
            self.env().revert(Error::InvalidNonce);
        }

        let domain = self.build_domain(use_casper_domain);
        let permit = Permit {
            owner: owner_eth.into(),
            spender: address_to_eth_bytes(&spender).into(),
            value: u256_to_bytes32(value),
            nonce: u256_to_bytes32(nonce),
            deadline: u256_to_bytes32(U256::from(deadline)),
        };
        let digest = hash_typed_data(&domain, &permit);
        let sig_array: [u8; 65] = match signature.as_slice().try_into() {
            Ok(sig) => sig,
            Err(_) => self.env().revert(Error::MalformedSignature),
        };
        match recover_eth_address(digest, &sig_array) {
            Some(recovered) if recovered == owner_eth => {}
            _ => self.env().revert(Error::InvalidSignature),
        }

        self.nonces.set(&owner, current_nonce + U256::one());
        self.permit_allowances.set(&(owner, spender), value);
    }

    pub fn transfer_from(&mut self, owner: &Address, recipient: &Address, amount: &U256) {
        let spender = self.env().caller();
        let permit_allowance = self.permit_allowances.get(&(*owner, spender)).unwrap_or_default();
        if permit_allowance >= *amount {
            self.permit_allowances.set(&(*owner, spender), permit_allowance - *amount);
            self.token.raw_transfer(owner, recipient, amount);
            return;
        }
        self.token.transfer_from(owner, recipient, amount);
    }

    pub fn nonces(&self, owner: &Address) -> U256 {
        self.nonces.get(owner).unwrap_or_default()
    }

    pub fn nonce_for_eth_address(&self, owner_eth_address: Bytes) -> U256 {
        let owner = owner_proxy_address(slice_to_20(&owner_eth_address).unwrap_or_revert_with(self, Error::InvalidOwnerEthAddress));
        self.nonces(&owner)
    }

    pub fn domain_separator_evm(&self) -> Vec<u8> {
        self.build_domain(false).separator_hash().to_vec()
    }

    pub fn domain_separator_casper(&self) -> Vec<u8> {
        self.build_domain(true).separator_hash().to_vec()
    }

    pub fn permit_owner_proxy(&self, owner_eth_address: Bytes) -> Address {
        owner_proxy_address(slice_to_20(&owner_eth_address).unwrap_or_revert_with(self, Error::InvalidOwnerEthAddress))
    }

    pub fn mint_to(&mut self, recipient: Address, amount: U256) {
        // DEMO ONLY: this helper stays intentionally unrestricted so the example and tests can mint
        // balances without adding a separate admin model. Do not copy this into production unchanged.
        self.token.raw_mint(&recipient, &amount);
    }
}

impl PermitToken {
    fn build_domain(&self, use_casper_domain: bool) -> DomainSeparator {
        let name = self.domain_name.get().unwrap_or_else(|| "PermitToken".to_string());
        let version = self.domain_version.get().unwrap_or_else(|| "1".to_string());
        if use_casper_domain {
            DomainBuilder::new()
                .name(&name)
                .version(&version)
                .custom_field("chain_name", DomainFieldValue::String("casper".into()))
                .custom_field(
                    "contract_package_hash",
                    DomainFieldValue::Bytes32(self.contract_package_hash_bytes()),
                )
                .build()
        } else {
            DomainBuilder::new()
                .name(&name)
                .version(&version)
                .chain_id(CHAIN_ID_CASPER_ASCII)
                .verifying_contract(self.contract_address_bytes())
                .build()
        }
    }

    fn contract_package_hash_bytes(&self) -> [u8; 32] {
        match self.env().self_address() {
            Address::Contract(hash) => hash.value(),
            Address::Account(hash) => hash.value(),
        }
    }

    fn contract_address_bytes(&self) -> [u8; 20] {
        let full = self.contract_package_hash_bytes();
        let mut addr = [0u8; 20];
        // EIP-712 expects an address-sized verifying contract, so the demo uses the leading 20 bytes.
        addr.copy_from_slice(&full[..20]);
        addr
    }

    fn current_block_time_u64(&self) -> u64 {
        self.env().get_block_time().into()
    }
}

fn owner_proxy_address(owner_eth: [u8; 20]) -> Address {
    let mut bytes = [0u8; 32];
    bytes[..20].copy_from_slice(&owner_eth);
    Address::Account(AccountHash::new(bytes))
}

pub fn address_to_eth_bytes(addr: &Address) -> [u8; 20] {
    let bytes = match addr {
        Address::Account(hash) => hash.value(),
        Address::Contract(hash) => hash.value(),
    };
    let mut result = [0u8; 20];
    result.copy_from_slice(&bytes[..20]);
    result
}

pub fn u256_to_bytes32(value: U256) -> [u8; 32] {
    let mut bytes = [0u8; 32];
    value.to_big_endian(&mut bytes);
    bytes
}

fn slice_to_20(bytes: &[u8]) -> Option<[u8; 20]> {
    if bytes.len() != 20 {
        return None;
    }
    let mut out = [0u8; 20];
    out.copy_from_slice(bytes);
    Some(out)
}
