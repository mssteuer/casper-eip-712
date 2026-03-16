//! Convenience re-exports for common usage.
//!
//! ```rust
//! use casper_eip_712::prelude::*;
//! ```

pub use crate::domain::{DomainBuilder, DomainFieldValue, DomainSeparator};
pub use crate::encoding::*;
pub use crate::hash::hash_typed_data;
pub use crate::keccak::keccak256;
pub use crate::traits::Eip712Struct;
pub use crate::types::{Approval, Permit, Transfer};

#[cfg(feature = "verify")]
pub use crate::verify::{recover_eth_address, recover_typed_data_signer, verify_signature};
