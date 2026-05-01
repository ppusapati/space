//! Multi-layer raster fusion (GI-FR-001/003).
//!
//! Provides three NaN-aware fusion primitives operating on a list of
//! co-registered `f32` rasters of identical shape:
//!
//! * [`weighted_sum`] — `Σ w_i · L_i` with the weights normalised over
//!   per-pixel finite contributions only.
//! * [`max_value`] — pixel-wise maximum across all layers.
//! * [`coverage`] — count of finite contributions per pixel (useful for
//!   masking and quality reporting).
//!
//! Designed for the multi-source data-fusion workflow described in
//! GI-FR-001 (optical / SAR / thermal / hyperspectral).

#![cfg_attr(docsrs, feature(doc_cfg))]

use ndarray::{Array2, ArrayView2};
use thiserror::Error;

/// Errors produced by `gi-fusion-eng`.
#[derive(Debug, Error, PartialEq)]
pub enum FusionError {
    /// Empty input list.
    #[error("at least one layer is required")]
    NoLayers,
    /// Layer shapes disagree.
    #[error("layer shape mismatch: expected {expected:?}, got {got:?}")]
    ShapeMismatch {
        /// Reference shape.
        expected: (usize, usize),
        /// Offending shape.
        got: (usize, usize),
    },
    /// Weight count does not match layer count.
    #[error("weight count {weights} does not match layer count {layers}")]
    WeightCount {
        /// Number of weights.
        weights: usize,
        /// Number of layers.
        layers: usize,
    },
}

/// Compute a per-pixel weighted sum, normalising the weights to sum to
/// one across the *finite* layer values at each pixel.
///
/// # Errors
/// [`FusionError`] for empty / shape mismatched / weight count issues.
pub fn weighted_sum(
    layers: &[ArrayView2<'_, f32>],
    weights: &[f32],
) -> Result<Array2<f32>, FusionError> {
    if layers.is_empty() {
        return Err(FusionError::NoLayers);
    }
    if weights.len() != layers.len() {
        return Err(FusionError::WeightCount {
            weights: weights.len(),
            layers: layers.len(),
        });
    }
    let dim = layers[0].dim();
    for l in layers {
        if l.dim() != dim {
            return Err(FusionError::ShapeMismatch { expected: dim, got: l.dim() });
        }
    }
    let mut out = Array2::<f32>::from_elem(dim, f32::NAN);
    for (r, c) in (0..dim.0).flat_map(|r| (0..dim.1).map(move |c| (r, c))) {
        let mut wsum = 0.0_f32;
        let mut acc = 0.0_f32;
        for (l, &w) in layers.iter().zip(weights) {
            let v = l[(r, c)];
            if v.is_finite() && w.is_finite() && w >= 0.0 {
                acc += w * v;
                wsum += w;
            }
        }
        if wsum > 0.0 {
            out[(r, c)] = acc / wsum;
        }
    }
    Ok(out)
}

/// Per-pixel maximum across layers (NaN-aware).
///
/// # Errors
/// [`FusionError`] for empty / shape mismatched layers.
pub fn max_value(layers: &[ArrayView2<'_, f32>]) -> Result<Array2<f32>, FusionError> {
    if layers.is_empty() {
        return Err(FusionError::NoLayers);
    }
    let dim = layers[0].dim();
    for l in layers {
        if l.dim() != dim {
            return Err(FusionError::ShapeMismatch { expected: dim, got: l.dim() });
        }
    }
    let mut out = Array2::<f32>::from_elem(dim, f32::NAN);
    for (r, c) in (0..dim.0).flat_map(|r| (0..dim.1).map(move |c| (r, c))) {
        let mut best = f32::NEG_INFINITY;
        for l in layers {
            let v = l[(r, c)];
            if v.is_finite() && v > best {
                best = v;
            }
        }
        if best.is_finite() {
            out[(r, c)] = best;
        }
    }
    Ok(out)
}

/// Per-pixel count of finite contributions across layers.
///
/// # Errors
/// [`FusionError`] for empty / shape mismatched layers.
pub fn coverage(layers: &[ArrayView2<'_, f32>]) -> Result<Array2<u16>, FusionError> {
    if layers.is_empty() {
        return Err(FusionError::NoLayers);
    }
    let dim = layers[0].dim();
    for l in layers {
        if l.dim() != dim {
            return Err(FusionError::ShapeMismatch { expected: dim, got: l.dim() });
        }
    }
    let mut out = Array2::<u16>::zeros(dim);
    for (r, c) in (0..dim.0).flat_map(|r| (0..dim.1).map(move |c| (r, c))) {
        let mut n = 0_u16;
        for l in layers {
            if l[(r, c)].is_finite() {
                n += 1;
            }
        }
        out[(r, c)] = n;
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn weighted_sum_handles_nan() {
        let a = array![[1.0_f32, f32::NAN]];
        let b = array![[3.0_f32, 5.0]];
        let r = weighted_sum(&[a.view(), b.view()], &[1.0, 1.0]).unwrap();
        assert_abs_diff_eq!(r[(0, 0)], 2.0, epsilon = 1e-6);
        assert_abs_diff_eq!(r[(0, 1)], 5.0, epsilon = 1e-6);
    }

    #[test]
    fn max_value_handles_nan() {
        let a = array![[1.0_f32, f32::NAN]];
        let b = array![[3.0_f32, 4.0]];
        let r = max_value(&[a.view(), b.view()]).unwrap();
        assert_abs_diff_eq!(r[(0, 0)], 3.0, epsilon = 1e-6);
        assert_abs_diff_eq!(r[(0, 1)], 4.0, epsilon = 1e-6);
    }

    #[test]
    fn coverage_counts_finite() {
        let a = array![[1.0_f32, f32::NAN]];
        let b = array![[3.0_f32, 4.0]];
        let r = coverage(&[a.view(), b.view()]).unwrap();
        assert_eq!(r[(0, 0)], 2);
        assert_eq!(r[(0, 1)], 1);
    }
}
