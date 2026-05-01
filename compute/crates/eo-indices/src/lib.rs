//! Vegetation and water spectral indices.
//!
//! Implements:
//!
//! - **NDVI** — Normalized Difference Vegetation Index
//!   `(NIR − Red) / (NIR + Red)`  *(Rouse et al., 1974)*
//! - **EVI**  — Enhanced Vegetation Index
//!   `G · (NIR − Red) / (NIR + C1·Red − C2·Blue + L)`  *(Huete et al., 2002)*
//! - **SAVI** — Soil-Adjusted Vegetation Index
//!   `(1 + L) · (NIR − Red) / (NIR + Red + L)`  *(Huete, 1988)*
//! - **NDWI** — Normalized Difference Water Index
//!   `(Green − NIR) / (Green + NIR)`  *(McFeeters, 1996)*
//!
//! All operators take dense 2-D reflectance arrays with identical shape and
//! return a real-valued image. Per-pixel results are clamped to the canonical
//! range `[-1.0, 1.0]` for NDVI / NDWI / SAVI and `[-1.0, 1.0]` (after optional
//! clamping) for EVI. A configurable sentinel value is emitted for invalid
//! pixels (zero denominator or non-finite input).

#![cfg_attr(docsrs, feature(doc_cfg))]

use ndarray::{Array2, ArrayView2, Zip};
use thiserror::Error;

mod evi;
mod ndvi;
mod ndwi;
mod savi;

pub use evi::{EviCoefficients, EviInput, EviResult, compute_evi};
pub use ndvi::{NdviInput, NdviResult, compute_ndvi};
pub use ndwi::{NdwiInput, NdwiResult, compute_ndwi};
pub use savi::{SaviInput, SaviResult, compute_savi};

/// Errors produced by the indices crate.
#[derive(Debug, Error, PartialEq)]
pub enum IndicesError {
    /// Input arrays do not share the same `(rows, cols)` shape.
    #[error("input band shape mismatch: expected {expected:?}, got {got:?}")]
    ShapeMismatch {
        /// Shape of the first/reference band.
        expected: (usize, usize),
        /// Shape of the offending band.
        got: (usize, usize),
    },
    /// One or more input arrays are empty.
    #[error("input bands must be non-empty")]
    Empty,
    /// A scalar parameter was outside its admissible range.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value as an `f64`.
        value: f64,
        /// Human-readable description of the admissible range.
        range: &'static str,
    },
}

/// Sentinel emitted at pixels for which the index is not defined (e.g.
/// zero denominator, NaN reflectance).
pub const INVALID_SENTINEL: f32 = f32::NAN;

#[inline]
fn validate_pair(a: ArrayView2<'_, f32>, b: ArrayView2<'_, f32>) -> Result<(), IndicesError> {
    let dim = a.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(IndicesError::Empty);
    }
    if dim != b.dim() {
        return Err(IndicesError::ShapeMismatch { expected: dim, got: b.dim() });
    }
    Ok(())
}

#[inline]
fn validate_triple(
    a: ArrayView2<'_, f32>,
    b: ArrayView2<'_, f32>,
    c: ArrayView2<'_, f32>,
) -> Result<(), IndicesError> {
    let dim = a.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(IndicesError::Empty);
    }
    if dim != b.dim() {
        return Err(IndicesError::ShapeMismatch { expected: dim, got: b.dim() });
    }
    if dim != c.dim() {
        return Err(IndicesError::ShapeMismatch { expected: dim, got: c.dim() });
    }
    Ok(())
}

/// Clamp a finite value to `[-1.0, 1.0]`, returning [`INVALID_SENTINEL`] for
/// non-finite inputs.
#[inline]
#[must_use]
pub fn clamp_unit(v: f32) -> f32 {
    if v.is_finite() { v.clamp(-1.0, 1.0) } else { INVALID_SENTINEL }
}

/// Apply a binary kernel pixel-wise into a freshly allocated output array.
fn map_pair<F>(a: ArrayView2<'_, f32>, b: ArrayView2<'_, f32>, kernel: F) -> Array2<f32>
where
    F: Fn(f32, f32) -> f32 + Sync + Send,
{
    let mut out = Array2::<f32>::from_elem(a.dim(), 0.0);
    Zip::from(&mut out).and(a).and(b).for_each(|o, &x, &y| *o = kernel(x, y));
    out
}

/// Apply a ternary kernel pixel-wise into a freshly allocated output array.
fn map_triple<F>(
    a: ArrayView2<'_, f32>,
    b: ArrayView2<'_, f32>,
    c: ArrayView2<'_, f32>,
    kernel: F,
) -> Array2<f32>
where
    F: Fn(f32, f32, f32) -> f32 + Sync + Send,
{
    let mut out = Array2::<f32>::from_elem(a.dim(), 0.0);
    Zip::from(&mut out).and(a).and(b).and(c).for_each(|o, &x, &y, &z| *o = kernel(x, y, z));
    out
}

/// Compute the normalized difference of two bands `(a − b) / (a + b)` with
/// the [`INVALID_SENTINEL`] policy applied to the zero-denominator and
/// non-finite cases. The result is clamped to `[-1.0, 1.0]`.
#[must_use]
pub fn normalized_difference(a: f32, b: f32) -> f32 {
    if !(a.is_finite() && b.is_finite()) {
        return INVALID_SENTINEL;
    }
    let num = a - b;
    let den = a + b;
    if den == 0.0 { INVALID_SENTINEL } else { clamp_unit(num / den) }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn normalized_difference_basic() {
        assert!((normalized_difference(0.5, 0.1) - (0.4 / 0.6)).abs() < 1e-6);
        assert!(normalized_difference(0.0, 0.0).is_nan());
        assert!(normalized_difference(f32::NAN, 0.1).is_nan());
        // (0.5 - 0.0)/(0.5 + 0.0) = 1.0 (upper bound).
        assert!((normalized_difference(0.5, 0.0) - 1.0).abs() < 1e-6);
        // (0.0 - 0.5)/(0.0 + 0.5) = -1.0 (lower bound).
        assert!((normalized_difference(0.0, 0.5) + 1.0).abs() < 1e-6);
        // Sum-to-zero with non-zero terms -> sentinel.
        assert!(normalized_difference(1.0, -1.0).is_nan());
    }

    #[test]
    fn clamp_unit_rules() {
        use approx::assert_abs_diff_eq;
        assert_abs_diff_eq!(clamp_unit(2.0), 1.0, epsilon = 0.0);
        assert_abs_diff_eq!(clamp_unit(-2.0), -1.0, epsilon = 0.0);
        assert!(clamp_unit(f32::NAN).is_nan());
        assert_abs_diff_eq!(clamp_unit(0.5), 0.5, epsilon = 0.0);
    }
}
