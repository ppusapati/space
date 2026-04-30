//! SGP4 / SDP4 propagation via the upstream `sgp4` crate.

use nalgebra::Vector3;
use thiserror::Error;

/// Errors produced by [`Sgp4Propagator`].
#[derive(Debug, Error)]
pub enum Sgp4Error {
    /// TLE could not be parsed (bad checksum, malformed columns, etc.).
    #[error("invalid TLE: {0}")]
    InvalidTle(String),
    /// Constants could not be initialised from the TLE.
    #[error("SGP4 constants initialisation failed: {0}")]
    ConstantsInit(String),
    /// Propagation produced a numerical error (decayed orbit, unstable
    /// integration).
    #[error("SGP4 propagation failed: {0}")]
    PropagationFailed(String),
}

/// Position / velocity expressed in the TEME frame.
#[derive(Debug, Clone, Copy)]
pub struct StateTeme {
    /// Position (km).
    pub position_km: Vector3<f64>,
    /// Velocity (km/s).
    pub velocity_km_s: Vector3<f64>,
}

/// SGP4 / SDP4 propagator built from a NORAD two-line element set.
#[derive(Debug, Clone)]
pub struct Sgp4Propagator {
    elements: sgp4::Elements,
    constants: sgp4::Constants,
}

impl Sgp4Propagator {
    /// Construct from raw TLE lines.
    ///
    /// # Errors
    /// [`Sgp4Error::InvalidTle`] if either line fails to parse, or
    /// [`Sgp4Error::ConstantsInit`] if SGP4 cannot derive its derived
    /// quantities (e.g., satellites with degenerate orbits).
    pub fn from_tle(line1: &str, line2: &str) -> Result<Self, Sgp4Error> {
        let elements = sgp4::Elements::from_tle(None, line1.as_bytes(), line2.as_bytes())
            .map_err(|e| Sgp4Error::InvalidTle(e.to_string()))?;
        let constants = sgp4::Constants::from_elements(&elements)
            .map_err(|e| Sgp4Error::ConstantsInit(e.to_string()))?;
        Ok(Self { elements, constants })
    }

    /// Propagate to `minutes_since_epoch` minutes after the TLE epoch.
    ///
    /// # Errors
    /// [`Sgp4Error::PropagationFailed`] for SGP4 numerical errors.
    pub fn propagate(&self, minutes_since_epoch: f64) -> Result<StateTeme, Sgp4Error> {
        let p = self
            .constants
            .propagate(sgp4::MinutesSinceEpoch(minutes_since_epoch))
            .map_err(|e| Sgp4Error::PropagationFailed(format!("{e:?}")))?;
        Ok(StateTeme {
            position_km: Vector3::new(p.position[0], p.position[1], p.position[2]),
            velocity_km_s: Vector3::new(p.velocity[0], p.velocity[1], p.velocity[2]),
        })
    }

    /// Access the parsed elements (epoch, mean motion, eccentricity, etc.).
    #[must_use]
    pub fn elements(&self) -> &sgp4::Elements {
        &self.elements
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    /// Vanguard-1 (NORAD 5) — the canonical Spacetrack Report #3
    /// verification case (Hoots, Schumacher & Glover, 2004).
    const VANGUARD_LINE1: &str =
        "1 00005U 58002B   00179.78495062  .00000023  00000-0  28098-4 0  4753";
    const VANGUARD_LINE2: &str =
        "2 00005  34.2682 348.7242 1859667 331.7664  19.3264 10.82419157413667";

    #[test]
    fn vanguard_at_epoch_matches_reference() {
        let p = Sgp4Propagator::from_tle(VANGUARD_LINE1, VANGUARD_LINE2).unwrap();
        let s = p.propagate(0.0).unwrap();
        // Reference values from the upstream sgp4 crate's verification run
        // for this canonical test case.
        assert_abs_diff_eq!(s.position_km.x, 7_022.466_472, epsilon = 1e-3);
        assert_abs_diff_eq!(s.position_km.y, -1_400.066_561, epsilon = 1e-3);
        assert_abs_diff_eq!(s.position_km.z, 0.051_065_5, epsilon = 1e-3);
    }

    #[test]
    fn vanguard_propagates_forward() {
        let p = Sgp4Propagator::from_tle(VANGUARD_LINE1, VANGUARD_LINE2).unwrap();
        let s0 = p.propagate(0.0).unwrap();
        let s1 = p.propagate(60.0).unwrap();
        // After 1 hour the position must change.
        let dr = (s1.position_km - s0.position_km).norm();
        assert!(dr > 100.0, "position barely changed in 60 min: {dr} km");
    }

    #[test]
    fn invalid_tle_rejected() {
        let err = Sgp4Propagator::from_tle("garbage", "garbage").unwrap_err();
        assert!(matches!(err, Sgp4Error::InvalidTle(_)));
    }
}
