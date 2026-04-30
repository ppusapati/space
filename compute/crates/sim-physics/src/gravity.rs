//! Earth gravity acceleration.
//!
//! Two-body central gravity:
//!
//! ```text
//! a_TB = −μ r / |r|³
//! ```
//!
//! Plus the J2 zonal-harmonic perturbation (Vallado §8.4):
//!
//! ```text
//! a_J2 = −(3 μ J₂ R_E²) / (2 |r|⁵) · [
//!         (1 − 5 z² / |r|²) [x, y, 0]ᵀ + (3 − 5 z² / |r|²) [0, 0, z]ᵀ
//!     ] / |r|
//! ```

use nalgebra::Vector3;

use crate::{J2_EARTH, MU_EARTH, R_EARTH};

/// Two-body acceleration at position `r` (m) in an Earth-centred inertial frame.
#[must_use]
pub fn two_body(r: Vector3<f64>) -> Vector3<f64> {
    let r_mag = r.norm();
    if r_mag <= 0.0 {
        return Vector3::zeros();
    }
    -MU_EARTH * r / r_mag.powi(3)
}

/// J2 zonal-harmonic perturbation at position `r` (m).
#[must_use]
pub fn j2(r: Vector3<f64>) -> Vector3<f64> {
    let r_mag = r.norm();
    if r_mag <= 0.0 {
        return Vector3::zeros();
    }
    let factor = -1.5 * J2_EARTH * MU_EARTH * R_EARTH.powi(2) / r_mag.powi(5);
    let z2_over_r2 = r.z.powi(2) / r_mag.powi(2);
    Vector3::new(
        factor * r.x * (1.0 - 5.0 * z2_over_r2),
        factor * r.y * (1.0 - 5.0 * z2_over_r2),
        factor * r.z * (3.0 - 5.0 * z2_over_r2),
    )
}

/// Total Earth gravitational acceleration (two-body + J2).
#[must_use]
pub fn total(r: Vector3<f64>) -> Vector3<f64> {
    two_body(r) + j2(r)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn two_body_radial() {
        // At 7000 km altitude (radius), gravity ≈ −μ/r² in radial direction.
        let r = Vector3::new(7_000_000.0, 0.0, 0.0);
        let a = two_body(r);
        let expected = -crate::MU_EARTH / 7_000_000.0_f64.powi(2);
        assert_abs_diff_eq!(a.x, expected, epsilon = 1e-3);
        assert_abs_diff_eq!(a.y, 0.0, epsilon = 1e-12);
    }

    #[test]
    fn j2_pole_perturbation_points_outward() {
        // At the north pole (z = r) the J2 perturbation acceleration
        // (relative to the spherical-Earth two-body gravity) points
        // *outward* in +z, reflecting the equatorial bulge: there is less
        // mass beneath the spacecraft at the pole, so the J2 correction
        // adds a positive radial acceleration.
        let r = Vector3::new(0.0, 0.0, 7_000_000.0);
        let a = j2(r);
        assert!(a.z > 0.0, "J2 perturbation at pole should be +z, got {}", a.z);
        assert_abs_diff_eq!(a.x, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(a.y, 0.0, epsilon = 1e-12);
    }

    #[test]
    fn j2_equator_perturbation_points_inward() {
        // At the equator, J2 augments inward gravity along the radial
        // axis (the bulge mass is right under the spacecraft).
        let r = Vector3::new(7_000_000.0, 0.0, 0.0);
        let a = j2(r);
        assert!(a.x < 0.0, "J2 perturbation at equator should be -x, got {}", a.x);
    }

    #[test]
    fn total_includes_both_components() {
        let r = Vector3::new(7_000_000.0, 0.0, 1_000_000.0);
        let a_total = total(r);
        let a_tb = two_body(r);
        let a_j2 = j2(r);
        for i in 0..3 {
            assert_abs_diff_eq!(a_total[i], a_tb[i] + a_j2[i], epsilon = 1e-12);
        }
    }
}
