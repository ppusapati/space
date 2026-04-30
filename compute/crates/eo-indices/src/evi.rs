//! Enhanced Vegetation Index.
//!
//! `EVI = G · (NIR − Red) / (NIR + C1·Red − C2·Blue + L)` — Huete et al. (2002).
//! Default coefficients match the MODIS / Sentinel-2 EVI formulation:
//! `G = 2.5`, `C1 = 6.0`, `C2 = 7.5`, `L = 1.0`.

use ndarray::{Array2, ArrayView2};

use crate::{INVALID_SENTINEL, IndicesError, map_triple, validate_triple};

/// Coefficients for the EVI formulation.
#[derive(Debug, Clone, Copy)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct EviCoefficients {
    /// Gain factor `G`. Default `2.5`.
    pub gain: f32,
    /// Aerosol-resistance coefficient `C1`. Default `6.0`.
    pub c1: f32,
    /// Aerosol-resistance coefficient `C2`. Default `7.5`.
    pub c2: f32,
    /// Canopy-background adjustment `L`. Default `1.0`.
    pub l: f32,
    /// Whether to clamp the result to `[-1.0, 1.0]`. Default `true`.
    pub clamp: bool,
}

impl Default for EviCoefficients {
    fn default() -> Self {
        Self { gain: 2.5, c1: 6.0, c2: 7.5, l: 1.0, clamp: true }
    }
}

impl EviCoefficients {
    fn validate(&self) -> Result<(), IndicesError> {
        if !(self.gain.is_finite() && self.gain > 0.0) {
            return Err(IndicesError::OutOfRange {
                name: "gain",
                value: f64::from(self.gain),
                range: "(0, +inf)",
            });
        }
        if !(self.c1.is_finite() && self.c1 >= 0.0) {
            return Err(IndicesError::OutOfRange {
                name: "c1",
                value: f64::from(self.c1),
                range: "[0, +inf)",
            });
        }
        if !(self.c2.is_finite() && self.c2 >= 0.0) {
            return Err(IndicesError::OutOfRange {
                name: "c2",
                value: f64::from(self.c2),
                range: "[0, +inf)",
            });
        }
        if !(self.l.is_finite() && self.l >= 0.0) {
            return Err(IndicesError::OutOfRange {
                name: "l",
                value: f64::from(self.l),
                range: "[0, +inf)",
            });
        }
        Ok(())
    }
}

/// Inputs for [`compute_evi`].
#[derive(Debug, Clone, Copy)]
pub struct EviInput<'a> {
    /// Blue band reflectance.
    pub blue: ArrayView2<'a, f32>,
    /// Red band reflectance.
    pub red: ArrayView2<'a, f32>,
    /// NIR band reflectance.
    pub nir: ArrayView2<'a, f32>,
    /// Coefficients (use [`EviCoefficients::default`] for the standard set).
    pub coefficients: EviCoefficients,
}

/// Output of [`compute_evi`].
#[derive(Debug, Clone)]
pub struct EviResult {
    /// Per-pixel EVI; non-finite pixels carry [`INVALID_SENTINEL`].
    pub evi: Array2<f32>,
}

/// Compute EVI for the given Blue / Red / NIR bands.
///
/// # Errors
/// Returns [`IndicesError`] for shape, emptiness, or coefficient violations.
pub fn compute_evi(input: EviInput<'_>) -> Result<EviResult, IndicesError> {
    input.coefficients.validate()?;
    validate_triple(input.nir, input.red, input.blue)?;
    let EviCoefficients { gain, c1, c2, l, clamp } = input.coefficients;
    let evi = map_triple(input.nir, input.red, input.blue, |nir, red, blue| {
        if !(nir.is_finite() && red.is_finite() && blue.is_finite()) {
            return INVALID_SENTINEL;
        }
        let denom = nir + c1 * red - c2 * blue + l;
        if denom == 0.0 {
            return INVALID_SENTINEL;
        }
        let v = gain * (nir - red) / denom;
        if !v.is_finite() {
            return INVALID_SENTINEL;
        }
        if clamp { v.clamp(-1.0, 1.0) } else { v }
    });
    Ok(EviResult { evi })
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn evi_default_coefficients_known_value() {
        // Pixel: blue=0.05, red=0.10, nir=0.40
        // denom = 0.40 + 6*0.10 - 7.5*0.05 + 1.0 = 0.40 + 0.60 - 0.375 + 1.0 = 1.625
        // num   = 2.5 * (0.40 - 0.10) = 0.75
        // evi   = 0.75 / 1.625 = 0.461538...
        let blue = array![[0.05_f32]];
        let red = array![[0.10_f32]];
        let nir = array![[0.40_f32]];
        let r = compute_evi(EviInput {
            blue: blue.view(),
            red: red.view(),
            nir: nir.view(),
            coefficients: EviCoefficients::default(),
        })
        .unwrap();
        assert_abs_diff_eq!(r.evi[(0, 0)], 0.75 / 1.625, epsilon = 1e-6);
    }

    #[test]
    fn evi_zero_denominator_emits_sentinel() {
        // Choose values that make the denominator zero.
        // l = 0, blue = 1.0, red = 0.0, nir = 0.0 -> denom = 0 + 0 - 7.5*1.0 + 0 = -7.5
        // To get exactly zero: nir=0, red=0, blue=0, l=0 -> denom=0
        let blue = array![[0.0_f32]];
        let red = array![[0.0_f32]];
        let nir = array![[0.0_f32]];
        let r = compute_evi(EviInput {
            blue: blue.view(),
            red: red.view(),
            nir: nir.view(),
            coefficients: EviCoefficients { l: 0.0, ..EviCoefficients::default() },
        })
        .unwrap();
        assert!(r.evi[(0, 0)].is_nan());
    }

    #[test]
    fn evi_invalid_coefficients() {
        let blue = array![[0.0_f32]];
        let red = array![[0.0_f32]];
        let nir = array![[0.0_f32]];
        let err = compute_evi(EviInput {
            blue: blue.view(),
            red: red.view(),
            nir: nir.view(),
            coefficients: EviCoefficients { gain: -1.0, ..EviCoefficients::default() },
        })
        .unwrap_err();
        match err {
            IndicesError::OutOfRange { name, .. } => assert_eq!(name, "gain"),
            other => panic!("unexpected error: {other:?}"),
        }
    }
}
