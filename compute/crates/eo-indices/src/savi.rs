//! Soil-Adjusted Vegetation Index.
//!
//! `SAVI = (1 + L) · (NIR − Red) / (NIR + Red + L)` — Huete (1988).
//! `L` is the canopy background adjustment factor in `[0.0, 1.0]`. `L = 0.5`
//! is the canonical value for moderate vegetation.

use ndarray::{Array2, ArrayView2};

use crate::{INVALID_SENTINEL, IndicesError, clamp_unit, map_pair, validate_pair};

/// Inputs for [`compute_savi`].
#[derive(Debug, Clone, Copy)]
pub struct SaviInput<'a> {
    /// Red band reflectance.
    pub red: ArrayView2<'a, f32>,
    /// NIR band reflectance.
    pub nir: ArrayView2<'a, f32>,
    /// Soil-brightness correction factor `L` in `[0.0, 1.0]`.
    pub l: f32,
}

/// Output of [`compute_savi`].
#[derive(Debug, Clone)]
pub struct SaviResult {
    /// Per-pixel SAVI; non-finite pixels carry [`INVALID_SENTINEL`].
    pub savi: Array2<f32>,
}

/// Compute SAVI.
///
/// # Errors
/// Returns [`IndicesError::OutOfRange`] if `l` is not in `[0.0, 1.0]`,
/// [`IndicesError::ShapeMismatch`] if bands differ, or [`IndicesError::Empty`]
/// if either band has a zero dimension.
pub fn compute_savi(input: SaviInput<'_>) -> Result<SaviResult, IndicesError> {
    if !(input.l.is_finite() && (0.0..=1.0).contains(&input.l)) {
        return Err(IndicesError::OutOfRange {
            name: "l",
            value: f64::from(input.l),
            range: "[0.0, 1.0]",
        });
    }
    validate_pair(input.nir, input.red)?;
    let l = input.l;
    let savi = map_pair(input.nir, input.red, |nir, red| {
        if !(nir.is_finite() && red.is_finite()) {
            return INVALID_SENTINEL;
        }
        let denom = nir + red + l;
        if denom == 0.0 {
            return INVALID_SENTINEL;
        }
        clamp_unit((1.0 + l) * (nir - red) / denom)
    });
    Ok(SaviResult { savi })
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn savi_canonical_l_half() {
        // pixel: nir=0.4, red=0.1, l=0.5
        // (1+0.5)*(0.4-0.1)/(0.4+0.1+0.5) = 1.5*0.3/1.0 = 0.45
        let red = array![[0.1_f32]];
        let nir = array![[0.4_f32]];
        let r = compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: 0.5 }).unwrap();
        assert_abs_diff_eq!(r.savi[(0, 0)], 0.45, epsilon = 1e-6);
    }

    #[test]
    fn savi_l_zero_reduces_to_ndvi() {
        let red = array![[0.10_f32, 0.20], [0.50, 0.30]];
        let nir = array![[0.40_f32, 0.60], [0.50, 0.10]];
        let r = compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: 0.0 }).unwrap();
        let ndvi_00 = (0.40 - 0.10) / (0.40 + 0.10);
        assert_abs_diff_eq!(r.savi[(0, 0)], ndvi_00 as f32, epsilon = 1e-6);
    }

    #[test]
    fn savi_rejects_invalid_l() {
        let red = array![[0.0_f32]];
        let nir = array![[0.0_f32]];
        let err = compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: -0.1 }).unwrap_err();
        assert!(matches!(err, IndicesError::OutOfRange { name: "l", .. }));
    }
}
