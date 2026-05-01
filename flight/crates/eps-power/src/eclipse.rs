//! Conical-shadow eclipse predictor.
//!
//! Given a spacecraft inertial position `r_sc` (m) and the Sun direction
//! unit vector `s_hat` (pointing from the central body to the Sun), this
//! module determines whether the spacecraft is in umbra (full shadow),
//! penumbra (partial shadow), or sunlit. The classical conical-shadow
//! geometry of Vallado (2013) §5.3 is used.
//!
//! For Earth shadow with a typical satellite altitude, the umbra cone
//! converges to a single point at ~1.4 million km from Earth, so within
//! GEO altitudes the umbra can be treated as a cylinder to good
//! accuracy. We provide the cylindrical-shadow approximation for the
//! common LEO/MEO case as well.

use nalgebra::Vector3;

/// Eclipse state.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Eclipse {
    /// Direct sunlight.
    Sunlit,
    /// Partial shadow (penumbra).
    Penumbra,
    /// Full shadow (umbra).
    Umbra,
}

/// Determine eclipse state using the **cylindrical-shadow** approximation,
/// which is exact when the central body's umbra is treated as an infinite
/// cylinder of radius `body_radius_m` aligned along `−s_hat`.
///
/// This is the simplest and most common LEO eclipse predictor.
#[must_use]
pub fn cylindrical_shadow(
    r_sc_m: Vector3<f64>,
    s_hat: Vector3<f64>,
    body_radius_m: f64,
) -> Eclipse {
    let s = s_hat.normalize();
    let r_dot_s = r_sc_m.dot(&s);
    if r_dot_s >= 0.0 {
        return Eclipse::Sunlit;
    }
    let perpendicular = r_sc_m - r_dot_s * s;
    if perpendicular.norm() <= body_radius_m {
        Eclipse::Umbra
    } else {
        Eclipse::Sunlit
    }
}

/// Determine eclipse state using the full **conical-shadow** geometry
/// (Vallado §5.3). `body_radius_m` is the central body radius and
/// `sun_radius_m` is the Sun's radius. `sun_distance_m` is the distance
/// from the central body to the Sun (used to derive the umbra/penumbra
/// half-angles).
#[must_use]
pub fn conical_shadow(
    r_sc_m: Vector3<f64>,
    s_hat: Vector3<f64>,
    body_radius_m: f64,
    sun_radius_m: f64,
    sun_distance_m: f64,
) -> Eclipse {
    let s = s_hat.normalize();
    let r_dot_s = r_sc_m.dot(&s);
    if r_dot_s >= 0.0 {
        return Eclipse::Sunlit;
    }
    let depth = -r_dot_s; // distance behind the body along the anti-Sun direction
    let perpendicular = (r_sc_m - r_dot_s * s).norm();
    // Half-angles at the body (from cone geometry).
    let alpha_umb = ((sun_radius_m - body_radius_m) / sun_distance_m).asin();
    let alpha_pen = ((sun_radius_m + body_radius_m) / sun_distance_m).asin();
    // Cone radii at the spacecraft's depth.
    let umbra_radius = body_radius_m - depth * alpha_umb.tan();
    let penumbra_radius = body_radius_m + depth * alpha_pen.tan();
    if umbra_radius > 0.0 && perpendicular <= umbra_radius {
        Eclipse::Umbra
    } else if perpendicular <= penumbra_radius {
        Eclipse::Penumbra
    } else {
        Eclipse::Sunlit
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Earth radius in metres for the test cases.
    const R_EARTH: f64 = 6_378_137.0;
    /// Sun radius in metres.
    const R_SUN: f64 = 696_000_000.0;
    /// AU in metres.
    const D_SUN: f64 = 1.495_978_707e11;

    #[test]
    fn cylindrical_sunlit_when_r_aligned_with_sun() {
        let s = Vector3::new(1.0, 0.0, 0.0);
        let r = Vector3::new(7e6, 0.0, 0.0);
        assert_eq!(cylindrical_shadow(r, s, R_EARTH), Eclipse::Sunlit);
    }

    #[test]
    fn cylindrical_umbra_directly_behind_earth() {
        // Spacecraft at (-7e6, 0, 0) with Sun in +x → r·s < 0, perpendicular = 0.
        let s = Vector3::new(1.0, 0.0, 0.0);
        let r = Vector3::new(-7e6, 0.0, 0.0);
        assert_eq!(cylindrical_shadow(r, s, R_EARTH), Eclipse::Umbra);
    }

    #[test]
    fn cylindrical_sunlit_when_perpendicular_distance_exceeds_radius() {
        let s = Vector3::new(1.0, 0.0, 0.0);
        let r = Vector3::new(-7e6, 8e6, 0.0);
        assert_eq!(cylindrical_shadow(r, s, R_EARTH), Eclipse::Sunlit);
    }

    #[test]
    fn conical_classifies_three_regions() {
        let s = Vector3::new(1.0, 0.0, 0.0);
        // Deep behind Earth, on-axis: umbra.
        let r_umb = Vector3::new(-7e6, 0.0, 0.0);
        assert_eq!(conical_shadow(r_umb, s, R_EARTH, R_SUN, D_SUN), Eclipse::Umbra);
        // Sun-facing side: sunlit.
        let r_sun = Vector3::new(7e6, 0.0, 0.0);
        assert_eq!(conical_shadow(r_sun, s, R_EARTH, R_SUN, D_SUN), Eclipse::Sunlit);
        // Just outside the umbra cone but inside penumbra → penumbra.
        let r_pen = Vector3::new(-7e6, R_EARTH * 1.005, 0.0);
        assert_eq!(conical_shadow(r_pen, s, R_EARTH, R_SUN, D_SUN), Eclipse::Penumbra);
    }
}
