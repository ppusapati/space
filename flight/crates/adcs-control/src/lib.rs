//! Spacecraft attitude / rate controllers.
//!
//! Implements four controllers required by SAT-FR-003:
//!
//! * [`pid`] — Proportional-Integral-Derivative with optional integrator
//!   anti-windup and output saturation.
//! * [`lqr`] — Linear Quadratic Regulator via iterative discrete algebraic
//!   Riccati equation (Hewer 1971).
//! * [`smc`] — Sliding Mode Controller with continuous saturation function
//!   to mitigate chattering.
//! * [`mpc`] — Unconstrained finite-horizon Model Predictive Control with
//!   quadratic cost, solved analytically per step.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod lqr;
pub mod mpc;
pub mod pid;
pub mod smc;

use thiserror::Error;

/// Errors common to all controllers in this crate.
#[derive(Debug, Error, PartialEq)]
pub enum ControlError {
    /// Dimension mismatch in matrix arguments.
    #[error("dimension mismatch: {0}")]
    DimensionMismatch(&'static str),
    /// Iterative algorithm did not converge.
    #[error("iteration did not converge after {iters} steps (residual {residual:.3e})")]
    DidNotConverge {
        /// Iterations performed.
        iters: u32,
        /// Final residual.
        residual: f64,
    },
    /// A scalar parameter is out of range.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range.
        range: &'static str,
    },
}
