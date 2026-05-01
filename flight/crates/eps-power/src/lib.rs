//! Electrical Power Subsystem primitives (SAT-FR-020/022/024).
//!
//! * [`budget`] — power-budget bookkeeping with per-load priority and
//!   load-shedding allocation when total demand exceeds available power.
//! * [`mppt`] — Perturb-and-Observe Maximum-Power-Point-Tracker for solar
//!   array boost converters.
//! * [`eclipse`] — analytical conical-shadow eclipse predictor for a
//!   spacecraft state vector and Sun direction.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod budget;
pub mod eclipse;
pub mod mppt;

use thiserror::Error;

/// Errors common to the EPS modules.
#[derive(Debug, Error, PartialEq)]
pub enum EpsError {
    /// Out-of-range scalar parameter.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range description.
        range: &'static str,
    },
}
