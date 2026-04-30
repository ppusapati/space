//! Normalized Difference Vegetation Index.
//!
//! `NDVI = (NIR − Red) / (NIR + Red)` — Rouse et al. (1974).

use ndarray::{Array2, ArrayView2};

use crate::{IndicesError, map_pair, normalized_difference, validate_pair};

/// Inputs for [`compute_ndvi`].
#[derive(Debug, Clone, Copy)]
pub struct NdviInput<'a> {
    /// Red band reflectance, dimensionless `[0, 1]`.
    pub red: ArrayView2<'a, f32>,
    /// Near-Infra-Red band reflectance, dimensionless `[0, 1]`.
    pub nir: ArrayView2<'a, f32>,
}

/// Output of [`compute_ndvi`].
#[derive(Debug, Clone)]
pub struct NdviResult {
    /// `[-1.0, 1.0]` per-pixel NDVI; non-finite pixels carry [`crate::INVALID_SENTINEL`].
    pub ndvi: Array2<f32>,
}

/// Compute NDVI for the given Red / NIR bands.
///
/// # Errors
/// Returns [`IndicesError::ShapeMismatch`] if `red` and `nir` differ in shape,
/// or [`IndicesError::Empty`] if either band has a zero dimension.
pub fn compute_ndvi(input: NdviInput<'_>) -> Result<NdviResult, IndicesError> {
    validate_pair(input.nir, input.red)?;
    let ndvi = map_pair(input.nir, input.red, normalized_difference);
    Ok(NdviResult { ndvi })
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn ndvi_known_values() {
        let red = array![[0.10_f32, 0.20], [0.50, 0.00]];
        let nir = array![[0.40_f32, 0.60], [0.50, 0.00]];
        let r = compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).unwrap();
        // (0.4-0.1)/(0.4+0.1)=0.6
        assert_abs_diff_eq!(r.ndvi[(0, 0)], 0.6, epsilon = 1e-6);
        // (0.6-0.2)/(0.8)=0.5
        assert_abs_diff_eq!(r.ndvi[(0, 1)], 0.5, epsilon = 1e-6);
        // (0.5-0.5)/(1.0)=0.0
        assert_abs_diff_eq!(r.ndvi[(1, 0)], 0.0, epsilon = 1e-6);
        // 0/0 -> sentinel
        assert!(r.ndvi[(1, 1)].is_nan());
    }

    #[test]
    fn ndvi_shape_mismatch() {
        let red = Array2::<f32>::zeros((4, 4));
        let nir = Array2::<f32>::zeros((4, 5));
        let err = compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).unwrap_err();
        // `validate_pair` is called with (nir, red), so `expected` reflects nir.
        assert_eq!(err, IndicesError::ShapeMismatch { expected: (4, 5), got: (4, 4) });
    }

    #[test]
    fn ndvi_empty() {
        let red = Array2::<f32>::zeros((0, 0));
        let nir = Array2::<f32>::zeros((0, 0));
        assert_eq!(
            compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).unwrap_err(),
            IndicesError::Empty
        );
    }
}
