//! Solar Radiation Pressure (cannonball model).
//!
//! ```text
//! a_SRP = −P_SRP · C_R · (A / m) · ŝ
//! ```
//!
//! where `P_SRP = Φ / c` is the solar pressure at 1 AU, scaled by
//! `(AU / |r_sun|)²`, `C_R` is the reflectivity coefficient (1 for fully
//! absorbing, 2 for perfectly reflecting), `A/m` is the area-to-mass
//! ratio, and `ŝ` is the unit vector from the spacecraft toward the Sun.

use nalgebra::Vector3;

use crate::{AU_M, C_LIGHT, SOLAR_FLUX_1AU};

/// SRP acceleration on a spacecraft at `r_sat` (m) given the Sun's
/// inertial position `r_sun` (m), reflectivity coefficient `c_r`, and
/// area-to-mass ratio `a_over_m` (m²/kg).
///
/// In eclipse (`in_shadow == true`) the SRP is zero.
#[must_use]
pub fn srp_acceleration(
    r_sat: Vector3<f64>,
    r_sun: Vector3<f64>,
    c_r: f64,
    a_over_m: f64,
    in_shadow: bool,
) -> Vector3<f64> {
    if in_shadow {
        return Vector3::zeros();
    }
    let r_sun_sat = r_sun - r_sat;
    let dist = r_sun_sat.norm();
    if dist <= 0.0 {
        return Vector3::zeros();
    }
    let s_hat = r_sun_sat / dist;
    let pressure_at_satellite = (SOLAR_FLUX_1AU / C_LIGHT) * (AU_M / dist).powi(2);
    -pressure_at_satellite * c_r * a_over_m * s_hat
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn shadow_returns_zero() {
        let a = srp_acceleration(
            Vector3::new(7e6, 0.0, 0.0),
            Vector3::new(AU_M, 0.0, 0.0),
            1.5,
            0.01,
            true,
        );
        assert!(a.iter().all(|v| *v == 0.0));
    }

    #[test]
    fn unit_au_acceleration_pushes_away_from_sun() {
        let a = srp_acceleration(
            Vector3::zeros(),
            Vector3::new(AU_M, 0.0, 0.0),
            1.0,
            1.0,
            false,
        );
        // -P·ŝ → x component negative (Sun is in +x).
        let expected_mag = SOLAR_FLUX_1AU / C_LIGHT;
        assert_abs_diff_eq!(a.x, -expected_mag, epsilon = 1e-15);
    }

    #[test]
    fn pressure_inverse_square_with_distance() {
        let close = srp_acceleration(
            Vector3::zeros(),
            Vector3::new(0.5 * AU_M, 0.0, 0.0),
            1.0,
            1.0,
            false,
        );
        let far = srp_acceleration(
            Vector3::zeros(),
            Vector3::new(2.0 * AU_M, 0.0, 0.0),
            1.0,
            1.0,
            false,
        );
        // Magnitude should scale by 1/d²: (0.5 AU)/(2 AU) = 1/4 → 16× ratio.
        let ratio = close.norm() / far.norm();
        assert_abs_diff_eq!(ratio, 16.0, epsilon = 1e-9);
    }
}
