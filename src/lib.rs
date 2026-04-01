//! # casper-eip-712
//!
//! EIP-712 style typed data hashing and domain separation for Casper Network.
//!
//! This crate provides a `no_std`-compatible toolkit for building EIP-712 domain
//! separators, hashing typed structs, and optionally recovering Ethereum-style
//! secp256k1 signers for hashes that can be produced by `ethers.js` or `viem`.
//!
//! ## Quick Start
//!
//! ```rust
//! use casper_eip_712::prelude::*;
//!
//! let domain = DomainBuilder::new()
//!     .name("MyToken")
//!     .version("1")
//!     .chain_id(1)
//!     .verifying_contract([0x11; 20])
//!     .build();
//!
//! let permit = Permit {
//!     owner: Address::Eth([0x22; 20]),
//!     spender: Address::Eth([0x33; 20]),
//!     value: [0u8; 32],
//!     nonce: [0u8; 32],
//!     deadline: [0u8; 32],
//! };
//!
//! let digest = hash_typed_data(&domain, &permit);
//! assert_eq!(digest.len(), 32);
//! ```

#![no_std]
extern crate alloc;

pub mod domain;
pub mod encoding;
pub mod hash;
pub mod keccak;
pub mod prelude;
pub mod traits;
pub mod type_hash;
pub mod types;

#[cfg(feature = "verify")]
pub mod verify;

pub use crate::domain::{DomainBuilder, DomainFieldValue, DomainSeparator};
pub use crate::encoding::*;
pub use crate::hash::hash_typed_data;
pub use crate::keccak::keccak256;
pub use crate::traits::Eip712Struct;
pub use crate::type_hash::compute_type_hash;
