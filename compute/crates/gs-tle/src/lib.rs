//! Ground-side TLE handling.
//!
//! Validates TLE checksums, parses the line into a structured form, and
//! provides convenience helpers to propagate via the bundled SGP4
//! propagator (re-exported from `orbit-prop`).

#![cfg_attr(docsrs, feature(doc_cfg))]

pub use orbit_prop::Sgp4Propagator;
pub use orbit_prop::sgp4_prop::StateTeme;

use nalgebra::{Matrix3, Vector3};
use thiserror::Error;

/// Errors produced by `gs-tle`.
#[derive(Debug, Error, PartialEq)]
pub enum TleError {
    /// Line had the wrong length.
    #[error("line {line} has {got} characters, expected 69")]
    WrongLength {
        /// Which line (1 or 2).
        line: u8,
        /// Actual length.
        got: usize,
    },
    /// Line did not start with the expected line number.
    #[error("line {line} starts with {got}, expected {expected}")]
    WrongLineNumber {
        /// Which line.
        line: u8,
        /// Got first character.
        got: char,
        /// Expected first character.
        expected: char,
    },
    /// Checksum mismatch.
    #[error("line {line} checksum mismatch: computed {computed}, expected {expected}")]
    BadChecksum {
        /// Which line.
        line: u8,
        /// Computed.
        computed: u8,
        /// From the line.
        expected: u8,
    },
}

/// Validate the modulo-10 checksum of a 69-character TLE line.
///
/// The TLE checksum sums every digit, with each `-` counting as 1, and
/// `+`/spaces/letters as 0; the last column holds the result modulo 10.
///
/// # Errors
/// [`TleError`] if length / line number / checksum is wrong.
pub fn validate_line(line: u8, s: &str) -> Result<(), TleError> {
    let bytes: Vec<u8> = s.bytes().collect();
    if bytes.len() != 69 {
        return Err(TleError::WrongLength { line, got: bytes.len() });
    }
    let expected_first = match line {
        1 => b'1',
        2 => b'2',
        _ => panic!("validate_line called with line != 1 or 2"),
    };
    if bytes[0] != expected_first {
        return Err(TleError::WrongLineNumber {
            line,
            got: bytes[0] as char,
            expected: expected_first as char,
        });
    }
    let mut sum = 0_u32;
    for &b in &bytes[..68] {
        if b.is_ascii_digit() {
            sum += u32::from(b - b'0');
        } else if b == b'-' {
            sum += 1;
        }
    }
    let computed = (sum % 10) as u8;
    let expected = bytes[68];
    if !expected.is_ascii_digit() {
        return Err(TleError::BadChecksum { line, computed, expected: 0 });
    }
    let expected_digit = expected - b'0';
    if computed != expected_digit {
        return Err(TleError::BadChecksum { line, computed, expected: expected_digit });
    }
    Ok(())
}

/// Validate a TLE pair (both lines must pass [`validate_line`]).
///
/// # Errors
/// First failing line propagates as [`TleError`].
pub fn validate_pair(line1: &str, line2: &str) -> Result<(), TleError> {
    validate_line(1, line1)?;
    validate_line(2, line2)?;
    Ok(())
}

/// Greenwich Mean Sidereal Time (GMST) at a given Julian Date (UT1) in
/// radians. The formula is taken from Vallado §3.5 (IAU 1982 model):
///
/// ```text
/// θ_GMST = 67_310.548_41 + (876_600·3600 + 8_640_184.812_866) · T + …
/// ```
///
/// Returns the angle modulo 2π.
#[must_use]
pub fn gmst_rad(jd_ut1: f64) -> f64 {
    let t_ut1 = (jd_ut1 - 2_451_545.0) / 36_525.0;
    let mut gmst_seconds = 67_310.548_41
        + (876_600.0 * 3600.0 + 8_640_184.812_866) * t_ut1
        + 0.093_104 * t_ut1 * t_ut1
        - 6.2e-6 * t_ut1 * t_ut1 * t_ut1;
    gmst_seconds = gmst_seconds.rem_euclid(86_400.0);
    let theta_rad = gmst_seconds * std::f64::consts::PI / 43_200.0;
    theta_rad.rem_euclid(2.0 * std::f64::consts::PI)
}

/// Rotate a TEME-frame position to ECEF (Earth-Centred-Earth-Fixed) by
/// applying the GMST rotation about Z. Polar motion is neglected (a
/// reasonable simplification at the metre level for ground-station
/// pointing in LEO).
#[must_use]
pub fn teme_to_ecef(r_teme_km: Vector3<f64>, jd_ut1: f64) -> Vector3<f64> {
    let theta = gmst_rad(jd_ut1);
    let c = theta.cos();
    let s = theta.sin();
    let r_z = Matrix3::new(c, s, 0.0, -s, c, 0.0, 0.0, 0.0, 1.0);
    r_z * r_teme_km
}

#[cfg(test)]
mod tests {
    use super::*;

    const VANGUARD_LINE1: &str =
        "1 00005U 58002B   00179.78495062  .00000023  00000-0  28098-4 0  4753";
    const VANGUARD_LINE2: &str =
        "2 00005  34.2682 348.7242 1859667 331.7664  19.3264 10.82419157413667";

    #[test]
    fn vanguard_checksums_validate() {
        validate_pair(VANGUARD_LINE1, VANGUARD_LINE2).unwrap();
    }

    #[test]
    fn corrupted_checksum_rejected() {
        let mut bad = VANGUARD_LINE1.to_string();
        // Flip the trailing checksum digit.
        bad.replace_range(68..69, "0");
        assert!(matches!(validate_line(1, &bad).unwrap_err(), TleError::BadChecksum { .. }));
    }

    #[test]
    fn wrong_length_rejected() {
        let short = "1 00005U 58002B   00179.78495062";
        assert!(matches!(validate_line(1, short).unwrap_err(), TleError::WrongLength { .. }));
    }

    #[test]
    fn gmst_at_j2000_epoch() {
        // GMST at J2000.0 (JD 2451545.0) ≈ 18h 41m 50.5s ≈ 4.894961 rad.
        let g = gmst_rad(2_451_545.0);
        // Tolerance: 1 milli-rad is plenty for an integer-noon epoch.
        assert!((g - 4.894_961).abs() < 1e-3, "GMST(J2000) = {g}");
    }

    #[test]
    fn teme_to_ecef_rotates_by_gmst() {
        let r = Vector3::new(7_000.0, 0.0, 0.0);
        let r_ecef = teme_to_ecef(r, 2_451_545.0);
        // Norm preserved.
        approx::assert_abs_diff_eq!(r_ecef.norm(), 7_000.0, epsilon = 1e-6);
    }
}
