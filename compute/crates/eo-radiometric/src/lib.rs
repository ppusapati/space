//! Radiometric calibration for optical satellite imagery.
//!
//! Implements the standard transforms required to convert raw digital numbers
//! (DN) into physically meaningful quantities for Landsat-class and Sentinel-2
//! products:
//!
//! * **DN → spectral radiance**:       `L = M_L · DN + A_L`
//! * **DN → TOA reflectance** (linear): `ρ' = M_ρ · DN + A_ρ`
//! * **TOA reflectance — sun-angle corrected**: `ρ = ρ' / sin(θ_SE)`
//!   (equivalently `ρ = ρ' / cos(θ_SZ)`)
//! * **Radiance → TOA reflectance**:   `ρ = π · L · d² / (E_SUN · cos(θ_SZ))`
//! * **Radiance → brightness temperature** (Planck inversion):
//!   `T = K2 / ln(K1 / L + 1)`
//! * **Sentinel-2 L1C DN → TOA reflectance**: `ρ = DN / Q`
//! * **Dark Object Subtraction (BOA approximation)**:
//!   `ρ_BOA = max(ρ_TOA − ρ_dark, 0)`
//! * **Earth–Sun distance** at a given Julian Day from Spencer's series.
//!
//! References:
//!
//! * USGS Landsat 8-9 Collection 2 Level-1 Data Format Control Book.
//! * ESA Sentinel-2 MSI Level-1C Product Definition.
//! * Chander, G., Markham, B. & Helder, D. (2009) "Summary of current
//!   radiometric calibration coefficients..." *Remote Sensing of Environment*.
//! * Chavez, P. S. (1996) "Image-Based Atmospheric Corrections — Revisited
//!   and Improved" *Photogrammetric Engineering & Remote Sensing*.

#![cfg_attr(docsrs, feature(doc_cfg))]

use std::f64::consts::PI;

use ndarray::{Array2, ArrayView2};
use thiserror::Error;

/// Errors produced by `eo-radiometric`.
#[derive(Debug, Error, PartialEq)]
pub enum RadiometricError {
    /// A scalar parameter was outside its admissible range.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value as `f64`.
        value: f64,
        /// Human-readable description of the admissible range.
        range: &'static str,
    },
    /// An input array had a zero dimension.
    #[error("input array must be non-empty")]
    Empty,
}

/// Standard linear calibration: `y = gain · x + offset`.
///
/// Used for DN→radiance (`gain = M_L`, `offset = A_L`) **and** for the linear
/// part of DN→TOA reflectance (`gain = M_ρ`, `offset = A_ρ`).
#[derive(Debug, Clone, Copy)]
pub struct LinearCalibration {
    /// Multiplicative scaling factor.
    pub gain: f32,
    /// Additive scaling factor.
    pub offset: f32,
}

impl LinearCalibration {
    /// Apply the calibration to a single DN value.
    #[inline]
    #[must_use]
    pub fn apply(self, dn: f32) -> f32 {
        self.gain.mul_add(dn, self.offset)
    }
}

/// Apply a [`LinearCalibration`] to every pixel.
///
/// # Errors
/// Returns [`RadiometricError::Empty`] if `dn` has a zero dimension.
pub fn apply_linear(
    dn: ArrayView2<'_, f32>,
    cal: LinearCalibration,
) -> Result<Array2<f32>, RadiometricError> {
    let dim = dn.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(RadiometricError::Empty);
    }
    let mut out = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut out).and(dn).for_each(|o, &d| *o = cal.apply(d));
    Ok(out)
}

/// Sun-angle correction in-place: `ρ ← ρ / sin(θ_SE)` where `θ_SE` is the
/// solar elevation angle (radians).
///
/// # Errors
/// Returns [`RadiometricError::OutOfRange`] if `sun_elevation_rad` is not in
/// `(0, π/2]` — sin(0) = 0 produces a singularity.
pub fn correct_sun_elevation_in_place(
    arr: &mut Array2<f32>,
    sun_elevation_rad: f32,
) -> Result<(), RadiometricError> {
    let se = sun_elevation_rad;
    if !(se.is_finite() && (0.0..=core::f32::consts::FRAC_PI_2).contains(&se) && se > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "sun_elevation_rad",
            value: f64::from(se),
            range: "(0, π/2]",
        });
    }
    let denom = se.sin();
    arr.mapv_inplace(|v| v / denom);
    Ok(())
}

/// Convert spectral radiance (`W·m⁻²·sr⁻¹·µm⁻¹`) to TOA reflectance:
///
/// `ρ = π · L · d² / (E_SUN · cos(θ_SZ))`.
///
/// # Errors
/// Returns [`RadiometricError`] if `esun_w_per_m2_um <= 0`,
/// `earth_sun_distance_au <= 0`, or `sun_zenith_rad` is not in `[0, π/2)`.
pub fn radiance_to_toa_reflectance(
    radiance: ArrayView2<'_, f32>,
    esun_w_per_m2_um: f32,
    earth_sun_distance_au: f32,
    sun_zenith_rad: f32,
) -> Result<Array2<f32>, RadiometricError> {
    if !(esun_w_per_m2_um.is_finite() && esun_w_per_m2_um > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "esun_w_per_m2_um",
            value: f64::from(esun_w_per_m2_um),
            range: "(0, +inf)",
        });
    }
    if !(earth_sun_distance_au.is_finite() && earth_sun_distance_au > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "earth_sun_distance_au",
            value: f64::from(earth_sun_distance_au),
            range: "(0, +inf)",
        });
    }
    let sz = sun_zenith_rad;
    if !(sz.is_finite() && (0.0..core::f32::consts::FRAC_PI_2).contains(&sz)) {
        return Err(RadiometricError::OutOfRange {
            name: "sun_zenith_rad",
            value: f64::from(sz),
            range: "[0, π/2)",
        });
    }
    let dim = radiance.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(RadiometricError::Empty);
    }
    let factor = (PI * f64::from(earth_sun_distance_au).powi(2))
        / (f64::from(esun_w_per_m2_um) * f64::from(sz.cos()));
    let factor_f32 = factor as f32;
    let mut out = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut out).and(radiance).for_each(|o, &l| *o = l * factor_f32);
    Ok(out)
}

/// Convert spectral radiance to brightness temperature (Kelvin) via the Planck
/// inversion `T = K2 / ln(K1 / L + 1)`.
///
/// # Errors
/// Returns [`RadiometricError::OutOfRange`] if `k1` or `k2` is non-positive.
pub fn radiance_to_brightness_temperature(
    radiance: ArrayView2<'_, f32>,
    k1: f32,
    k2: f32,
) -> Result<Array2<f32>, RadiometricError> {
    if !(k1.is_finite() && k1 > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "k1",
            value: f64::from(k1),
            range: "(0, +inf)",
        });
    }
    if !(k2.is_finite() && k2 > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "k2",
            value: f64::from(k2),
            range: "(0, +inf)",
        });
    }
    let dim = radiance.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(RadiometricError::Empty);
    }
    let mut out = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut out).and(radiance).for_each(|o, &l| {
        if l <= 0.0 || !l.is_finite() {
            *o = f32::NAN;
        } else {
            let ln_arg = k1 / l + 1.0;
            *o = k2 / ln_arg.ln();
        }
    });
    Ok(out)
}

/// Convert Sentinel-2 L1C digital numbers to TOA reflectance:
/// `ρ = DN / quantification_value`. The default `quantification_value` for
/// Sentinel-2 L1C/L2A is `10000`.
///
/// # Errors
/// Returns [`RadiometricError::OutOfRange`] if `quantification_value <= 0`.
pub fn sentinel2_dn_to_toa(
    dn: ArrayView2<'_, f32>,
    quantification_value: f32,
) -> Result<Array2<f32>, RadiometricError> {
    if !(quantification_value.is_finite() && quantification_value > 0.0) {
        return Err(RadiometricError::OutOfRange {
            name: "quantification_value",
            value: f64::from(quantification_value),
            range: "(0, +inf)",
        });
    }
    let dim = dn.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(RadiometricError::Empty);
    }
    let inv = 1.0_f32 / quantification_value;
    let mut out = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut out).and(dn).for_each(|o, &d| *o = d * inv);
    Ok(out)
}

/// Apply a Dark Object Subtraction to obtain a first-order BOA reflectance:
/// `ρ_BOA = max(ρ_TOA − ρ_dark, 0)`.
///
/// `dark_value` is normally taken from the histogram minimum of a dark target
/// (water, cloud shadow). Negative or non-finite results are clamped to zero.
///
/// # Errors
/// Returns [`RadiometricError`] if `dark_value` is non-finite or `toa` empty.
pub fn dark_object_subtraction(
    toa: ArrayView2<'_, f32>,
    dark_value: f32,
) -> Result<Array2<f32>, RadiometricError> {
    if !dark_value.is_finite() {
        return Err(RadiometricError::OutOfRange {
            name: "dark_value",
            value: f64::from(dark_value),
            range: "finite",
        });
    }
    let dim = toa.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(RadiometricError::Empty);
    }
    let mut out = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut out).and(toa).for_each(|o, &v| {
        let r = v - dark_value;
        *o = if r.is_finite() && r > 0.0 { r } else if r.is_finite() { 0.0 } else { f32::NAN };
    });
    Ok(out)
}

/// Earth–Sun distance in astronomical units at a given Julian day-of-year
/// using Spencer's truncated Fourier series (Spencer, 1971), accurate to
/// roughly 1·10⁻⁴ AU:
///
/// `1/d² = 1.000110 + 0.034221·cos Γ + 0.001280·sin Γ
///       + 0.000719·cos 2Γ + 0.000077·sin 2Γ`
///
/// where `Γ = 2π·(N−1)/365`.
///
/// # Errors
/// Returns [`RadiometricError::OutOfRange`] if `day_of_year` is not in `1..=366`.
pub fn earth_sun_distance_au(day_of_year: u32) -> Result<f32, RadiometricError> {
    if !(1..=366).contains(&day_of_year) {
        return Err(RadiometricError::OutOfRange {
            name: "day_of_year",
            value: f64::from(day_of_year),
            range: "[1, 366]",
        });
    }
    let n = f64::from(day_of_year);
    let gamma = 2.0 * PI * (n - 1.0) / 365.0;
    let inv_d2 = 1.000_110
        + 0.034_221 * gamma.cos()
        + 0.001_280 * gamma.sin()
        + 0.000_719 * (2.0 * gamma).cos()
        + 0.000_077 * (2.0 * gamma).sin();
    Ok((1.0 / inv_d2.sqrt()) as f32)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn linear_calibration_applies() {
        let dn = array![[100.0_f32, 200.0]];
        let cal = LinearCalibration { gain: 0.01, offset: -1.0 };
        let r = apply_linear(dn.view(), cal).unwrap();
        assert_abs_diff_eq!(r[(0, 0)], 0.0, epsilon = 1e-6);
        assert_abs_diff_eq!(r[(0, 1)], 1.0, epsilon = 1e-6);
    }

    #[test]
    fn sun_elevation_correction() {
        let mut a = array![[0.5_f32]];
        // sin(30°) = 0.5 -> result 1.0
        correct_sun_elevation_in_place(&mut a, 30.0_f32.to_radians()).unwrap();
        assert_abs_diff_eq!(a[(0, 0)], 1.0, epsilon = 1e-5);
    }

    #[test]
    fn sun_elevation_zero_rejected() {
        let mut a = array![[0.5_f32]];
        assert!(correct_sun_elevation_in_place(&mut a, 0.0).is_err());
    }

    #[test]
    fn radiance_to_toa_reflectance_known_value() {
        // L = 100, ESUN = 1969 (Landsat 5 TM B3), d=1.0 AU, sun zenith = 45°
        // ρ = π·L·d² / (ESUN·cos(45°))
        let l = array![[100.0_f32]];
        let r = radiance_to_toa_reflectance(l.view(), 1969.0, 1.0, 45.0_f32.to_radians()).unwrap();
        let expected = (PI * 100.0 / (1969.0 * 45.0_f64.to_radians().cos())) as f32;
        assert_abs_diff_eq!(r[(0, 0)], expected, epsilon = 1e-5);
    }

    #[test]
    fn brightness_temperature_landsat_b6() {
        // Landsat 5 TM B6 calibration: K1 = 607.76, K2 = 1260.56
        // L = 10 W/m^2/sr/um -> T = 1260.56 / ln(607.76/10 + 1)
        let l = array![[10.0_f32]];
        let bt = radiance_to_brightness_temperature(l.view(), 607.76, 1260.56).unwrap();
        let expected = 1260.56_f32 / (607.76_f32 / 10.0 + 1.0).ln();
        assert_abs_diff_eq!(bt[(0, 0)], expected, epsilon = 1e-3);
    }

    #[test]
    fn brightness_temperature_zero_radiance_is_nan() {
        let l = array![[0.0_f32]];
        let bt = radiance_to_brightness_temperature(l.view(), 600.0, 1200.0).unwrap();
        assert!(bt[(0, 0)].is_nan());
    }

    #[test]
    fn sentinel2_dn() {
        let dn = array![[10000.0_f32, 5000.0]];
        let r = sentinel2_dn_to_toa(dn.view(), 10000.0).unwrap();
        assert_abs_diff_eq!(r[(0, 0)], 1.0, epsilon = 1e-6);
        assert_abs_diff_eq!(r[(0, 1)], 0.5, epsilon = 1e-6);
    }

    #[test]
    fn dos_clamps_negative() {
        let toa = array![[0.20_f32, 0.05]];
        let r = dark_object_subtraction(toa.view(), 0.10).unwrap();
        assert_abs_diff_eq!(r[(0, 0)], 0.10, epsilon = 1e-6);
        assert_abs_diff_eq!(r[(0, 1)], 0.0, epsilon = 1e-6);
    }

    #[test]
    fn earth_sun_distance_perihelion_aphelion() {
        // Perihelion (~Jan 4) and aphelion (~Jul 4) bracket [0.983, 1.017].
        let d_peri = earth_sun_distance_au(4).unwrap();
        let d_aphe = earth_sun_distance_au(185).unwrap();
        assert!(d_peri < 1.0);
        assert!(d_aphe > 1.0);
        assert!((d_peri - 0.983).abs() < 0.01);
        assert!((d_aphe - 1.017).abs() < 0.01);
    }

    #[test]
    fn earth_sun_distance_invalid_day() {
        assert!(earth_sun_distance_au(0).is_err());
        assert!(earth_sun_distance_au(367).is_err());
    }
}
