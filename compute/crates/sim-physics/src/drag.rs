//! Atmospheric drag acceleration.
//!
//! Uses an exponential atmospheric density model (Vallado §8.6):
//!
//! ```text
//! ρ(h) = ρ₀ · exp(−(h − h₀) / H)
//! ```
//!
//! and the standard ballistic-drag formula:
//!
//! ```text
//! a_drag = −½ ρ |v_rel| (C_D · A / m) v_rel
//! ```
//!
//! For LEO, `v_rel` differs from the inertial velocity by the local
//! atmospheric co-rotation velocity `ω_E × r`. This module exposes both
//! the "no-rotation" version (`drag_inertial`) and the rotating-atmosphere
//! version (`drag_corotating`).

use nalgebra::Vector3;

use crate::R_EARTH;

/// Earth angular velocity (rad/s).
pub const OMEGA_EARTH: f64 = 7.292_115_146_7e-5;

/// Exponential-atmosphere density at altitude `altitude_m` (m above
/// mean Earth radius), using a single scale-height fit valid in LEO.
/// Reference value: `ρ₀ = 3.614e-13 kg/m³` at 400 km, `H = 50 km`.
#[must_use]
pub fn density_kg_per_m3(altitude_m: f64) -> f64 {
    let h_ref = 400_000.0;
    let rho_ref = 3.614e-13;
    let scale_height = 50_000.0;
    rho_ref * ((-(altitude_m - h_ref) / scale_height).exp())
}

/// Drag acceleration with the inertial velocity (no atmosphere co-rotation).
/// `cd_a_over_m` is the ballistic coefficient `C_D · A / m` (m²/kg).
#[must_use]
pub fn drag_inertial(r: Vector3<f64>, v: Vector3<f64>, cd_a_over_m: f64) -> Vector3<f64> {
    let altitude = r.norm() - R_EARTH;
    let rho = density_kg_per_m3(altitude);
    let v_mag = v.norm();
    if v_mag <= 0.0 {
        return Vector3::zeros();
    }
    -0.5 * rho * v_mag * cd_a_over_m * v
}

/// Drag acceleration with co-rotating atmosphere.
#[must_use]
pub fn drag_corotating(r: Vector3<f64>, v: Vector3<f64>, cd_a_over_m: f64) -> Vector3<f64> {
    let omega = Vector3::new(0.0, 0.0, OMEGA_EARTH);
    let v_atm = omega.cross(&r);
    let v_rel = v - v_atm;
    let altitude = r.norm() - R_EARTH;
    let rho = density_kg_per_m3(altitude);
    let v_rel_mag = v_rel.norm();
    if v_rel_mag <= 0.0 {
        return Vector3::zeros();
    }
    -0.5 * rho * v_rel_mag * cd_a_over_m * v_rel
}

#[cfg(test)]
mod tests {
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn density_decreases_with_altitude() {
        let rho_low = density_kg_per_m3(300_000.0);
        let rho_high = density_kg_per_m3(600_000.0);
        assert!(rho_low > rho_high);
    }

    #[test]
    fn drag_decelerates_motion() {
        let r = Vector3::new(R_EARTH + 400_000.0, 0.0, 0.0);
        let v = Vector3::new(0.0, 7_700.0, 0.0);
        let a = drag_inertial(r, v, 0.01);
        // Deceleration is opposite to velocity.
        assert!(a.dot(&v) < 0.0);
    }

    #[test]
    fn corotation_reduces_relative_velocity() {
        let r = Vector3::new(R_EARTH + 400_000.0, 0.0, 0.0);
        let v_inertial_only = Vector3::new(0.0, 7_700.0, 0.0);
        let a_inertial = drag_inertial(r, v_inertial_only, 0.01);
        let a_co = drag_corotating(r, v_inertial_only, 0.01);
        // Co-rotating drag should be smaller in magnitude (relative velocity
        // is smaller because the atmosphere moves with the spacecraft).
        assert!(a_co.norm() < a_inertial.norm());
    }
}
