//! Normalized Difference Water Index.
//!
//! `NDWI = (Green − NIR) / (Green + NIR)` — McFeeters (1996). High values
//! indicate open water; vegetation produces negative values.

use ndarray::{Array2, ArrayView2};

use crate::{IndicesError, map_pair, normalized_difference, validate_pair};

/// Inputs for [`compute_ndwi`].
#[derive(Debug, Clone, Copy)]
pub struct NdwiInput<'a> {
    /// Green band reflectance.
    pub green: ArrayView2<'a, f32>,
    /// NIR band reflectance.
    pub nir: ArrayView2<'a, f32>,
}

/// Output of [`compute_ndwi`].
#[derive(Debug, Clone)]
pub struct NdwiResult {
    /// Per-pixel NDWI; non-finite pixels carry [`crate::INVALID_SENTINEL`].
    pub ndwi: Array2<f32>,
}

/// Compute NDWI for the given Green / NIR bands.
///
/// # Errors
/// Returns [`IndicesError::ShapeMismatch`] if `green` and `nir` differ in
/// shape, or [`IndicesError::Empty`] if either band has a zero dimension.
pub fn compute_ndwi(input: NdwiInput<'_>) -> Result<NdwiResult, IndicesError> {
    validate_pair(input.green, input.nir)?;
    let ndwi = map_pair(input.green, input.nir, normalized_difference);
    Ok(NdwiResult { ndwi })
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn ndwi_water_signature() {
        // Water: high green, low NIR -> positive NDWI.
        let green = array![[0.10_f32, 0.20]];
        let nir = array![[0.02_f32, 0.40]];
        let r = compute_ndwi(NdwiInput { green: green.view(), nir: nir.view() }).unwrap();
        assert_abs_diff_eq!(r.ndwi[(0, 0)], (0.10 - 0.02) / (0.10 + 0.02), epsilon = 1e-6);
        // Vegetation: green < nir -> negative NDWI.
        assert!(r.ndwi[(0, 1)] < 0.0);
    }
}
