//! Synthetic Aperture Radar (SAR) processing.
//!
//! Implements the algorithms required by EO-FR-015:
//!
//! * [`multilook`] — block averaging in range and azimuth to convert
//!   complex single-look (SLC) imagery into multi-look intensity.
//! * [`speckle`] — Lee and Frost adaptive speckle filters.
//! * [`polarimetric`] — Pauli decomposition and Cloude-Pottier
//!   H / A / Alpha eigenvalue decomposition for full-pol coherency.
//!
//! All public functions return `Result<…, SarError>` and never panic on
//! validated input. Complex pixels use [`num_complex::Complex32`].

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod multilook;
pub mod polarimetric;
pub mod speckle;

use thiserror::Error;

/// Errors produced by `eo-sar`.
#[derive(Debug, Error, PartialEq)]
pub enum SarError {
    /// Input array had a zero dimension.
    #[error("input array must be non-empty")]
    Empty,
    /// Two arrays disagreed in shape.
    #[error("input shape mismatch: expected {expected:?}, got {got:?}")]
    ShapeMismatch {
        /// Reference shape.
        expected: (usize, usize),
        /// Offending shape.
        got: (usize, usize),
    },
    /// A scalar parameter is outside its admissible range.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range description.
        range: &'static str,
    },
}
