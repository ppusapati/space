//! Spacecraft actuator models and torque allocation.
//!
//! Implements SAT-FR-004:
//!
//! * [`reaction_wheel`] — RW model with maximum torque / momentum / friction
//!   and a pseudo-inverse torque-allocation routine for an arbitrary array of
//!   wheel mounting axes.
//! * [`magnetorquer`] — magnetic dipole moment → body torque via `τ = m × B`,
//!   with per-axis maximum dipole limits.
//! * [`thruster`] — on/off thruster model with thrust, specific impulse, and
//!   moment-arm geometry.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod magnetorquer;
pub mod reaction_wheel;
pub mod thruster;

use thiserror::Error;

/// Errors common to the actuator models.
#[derive(Debug, Error, PartialEq)]
pub enum ActuatorError {
    /// A configuration parameter is out of range or non-finite.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range.
        range: &'static str,
    },
    /// Wheel allocation matrix is singular / underdetermined.
    #[error("wheel allocation: {0}")]
    Allocation(&'static str),
}
