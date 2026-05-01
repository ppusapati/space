//! Error Detection And Correction codes for radiation-hardened memory.
//!
//! Implements:
//!
//! * [`hamming`] — Hamming(7, 4) and extended Hamming(8, 4) SEC-DED
//!   codes (Single-Error-Correcting, Double-Error-Detecting).
//! * [`secded64`] — extended Hamming(72, 64) SEC-DED suitable for
//!   protecting 64-bit memory words, the standard ECC-DRAM scheme.
//! * [`scrub`] — memory scrubber: walks a buffer of 64-bit + 8-parity
//!   words, corrects every detected single-bit error, counts
//!   uncorrectable double-bit errors, and rewrites every word.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod hamming;
pub mod scrub;
pub mod secded64;

use thiserror::Error;

/// Errors produced by `cdh-edac`.
#[derive(Debug, Error, PartialEq)]
pub enum EdacError {
    /// Decoder detected an uncorrectable double-bit error.
    #[error("uncorrectable double-bit error detected")]
    Uncorrectable,
    /// Input field out of range for the chosen code.
    #[error("input out of range: {0}")]
    OutOfRange(&'static str),
}

/// Outcome of a SEC-DED decode operation.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DecodeOutcome {
    /// No error detected; data is the decoded payload.
    NoError,
    /// Single-bit error corrected at the indicated bit index.
    Corrected {
        /// Bit position (0-based) that was flipped before correction.
        bit_index: u32,
    },
    /// Double-bit error detected; data is unreliable.
    Uncorrectable,
}
