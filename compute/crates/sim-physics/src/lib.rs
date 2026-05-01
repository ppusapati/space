//! 6-DOF spacecraft simulation primitives (SAT-FR-030).
//!
//! Provides:
//!
//! * [`State`] — full 13-element 6-DOF state: position (m), velocity
//!   (m/s), attitude quaternion (body→inertial), and body angular
//!   velocity (rad/s).
//! * [`gravity`] — two-body and J2 zonal-harmonic Earth gravity.
//! * [`drag`] — exponential atmospheric density model + ballistic drag.
//! * [`srp`] — cannonball solar-radiation-pressure acceleration.
//! * [`rigid_body`] — Euler's rotational equations for a body with a
//!   diagonal inertia tensor.
//! * [`integrator`] — RK4 step that advances the full 6-DOF state given
//!   user-supplied force / torque models.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod drag;
pub mod gravity;
pub mod integrator;
pub mod rigid_body;
pub mod srp;

use nalgebra::{Quaternion, UnitQuaternion, Vector3};
use thiserror::Error;

/// Errors common to the simulation modules.
#[derive(Debug, Error, PartialEq)]
pub enum SimError {
    /// Out-of-range scalar parameter.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range.
        range: &'static str,
    },
}

/// Earth gravitational parameter (m³/s²).
pub const MU_EARTH: f64 = 3.986_004_418e14;
/// Earth equatorial radius (m).
pub const R_EARTH: f64 = 6_378_137.0;
/// Second zonal harmonic of Earth's potential.
pub const J2_EARTH: f64 = 1.082_626_68e-3;
/// Solar radiation flux at 1 AU (W/m²).
pub const SOLAR_FLUX_1AU: f64 = 1_361.0;
/// Speed of light (m/s).
pub const C_LIGHT: f64 = 299_792_458.0;
/// 1 AU in metres.
pub const AU_M: f64 = 1.495_978_707e11;

/// 6-DOF state vector.
#[derive(Debug, Clone, Copy)]
pub struct State {
    /// Position in inertial frame (m).
    pub r: Vector3<f64>,
    /// Velocity in inertial frame (m/s).
    pub v: Vector3<f64>,
    /// Attitude quaternion (body → inertial), unit-norm.
    pub q: UnitQuaternion<f64>,
    /// Body-frame angular velocity (rad/s).
    pub omega: Vector3<f64>,
}

impl Default for State {
    fn default() -> Self {
        Self {
            r: Vector3::zeros(),
            v: Vector3::zeros(),
            q: UnitQuaternion::identity(),
            omega: Vector3::zeros(),
        }
    }
}

impl State {
    /// Construct a state with the given position, velocity, attitude
    /// (axis-angle in radians), and angular velocity.
    #[must_use]
    pub fn new(r: Vector3<f64>, v: Vector3<f64>, q: UnitQuaternion<f64>, omega: Vector3<f64>) -> Self {
        Self { r, v, q, omega }
    }

    /// Rotate a body-frame vector to the inertial frame.
    #[must_use]
    pub fn body_to_inertial(&self, vec_body: Vector3<f64>) -> Vector3<f64> {
        self.q * vec_body
    }

    /// Rotate an inertial-frame vector to the body frame.
    #[must_use]
    pub fn inertial_to_body(&self, vec_inertial: Vector3<f64>) -> Vector3<f64> {
        self.q.inverse() * vec_inertial
    }
}

/// Quaternion derivative `q̇ = ½ q ⊗ ω` for body-frame `ω`.
#[must_use]
pub fn quaternion_derivative(q: UnitQuaternion<f64>, omega: Vector3<f64>) -> Quaternion<f64> {
    let omega_q = Quaternion::new(0.0, omega.x, omega.y, omega.z);
    (q.into_inner() * omega_q) * 0.5
}
