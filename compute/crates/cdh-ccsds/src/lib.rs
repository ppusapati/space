//! CCSDS protocol implementations for spacecraft Command & Data Handling.
//!
//! Implements:
//!
//! * [`spp`]  — Space Packet Protocol (CCSDS 133.0-B-2): primary header
//!   encoding/decoding plus full packet construction.
//! * [`tc_frame`]  — Telecommand Transfer Frame (CCSDS 232.0-B-3) primary
//!   header and frame construction with optional Frame Error Control.
//! * [`tm_frame`]  — Telemetry Transfer Frame (CCSDS 132.0-B-2) primary
//!   header and frame construction with optional Operational Control Field
//!   and Frame Error Control.
//! * [`crc`]  — CRC-CCITT (X.25 / 0x1021 polynomial) computation, used by
//!   both TC and TM as the Frame Error Control Field.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod crc;
pub mod spp;
pub mod tc_frame;
pub mod tm_frame;

use thiserror::Error;

/// Errors common to the CCSDS modules.
#[derive(Debug, Error, PartialEq)]
pub enum CcsdsError {
    /// Buffer was too short to contain the expected structure.
    #[error("buffer too short: needed {needed} bytes, got {got}")]
    Truncated {
        /// Required length.
        needed: usize,
        /// Actual length.
        got: usize,
    },
    /// Field value out of the spec-defined range.
    #[error("field `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: u32,
        /// Admissible range description.
        range: &'static str,
    },
    /// Computed CRC did not match the trailing FECF.
    #[error("FECF mismatch: computed {computed:#06x}, expected {expected:#06x}")]
    BadFecf {
        /// Computed CRC.
        computed: u16,
        /// Expected (received) CRC.
        expected: u16,
    },
}
