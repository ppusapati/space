//! IHS (Intensity-Hue-Saturation) pan-sharpening for 3-band RGB imagery.
//!
//! The intensity transform is the row-sum of the standard sphere model:
//! `I = (R + G + B) / 3`. The kernel:
//!
//! 1. Computes `I` from the input RGB.
//! 2. Histogram-matches `Pan` to `I` to remove gain bias.
//! 3. Replaces `I` with the matched Pan and reconstructs RGB by adding the
//!    per-band difference `(Pan' − I)` back to each band.
//!
//! Step 3 is mathematically equivalent to the IHS round-trip via the standard
//! linear transform but avoids unnecessary 3-channel matrix algebra.

use ndarray::{Array3, ArrayView2, ArrayView3};

use crate::{PansharpenError, hist_match::histogram_match};

/// Apply IHS pan-sharpening to a 3-band RGB cube.
///
/// `ms` shape is `(3, rows, cols)`; `pan` shape is `(rows, cols)`.
///
/// # Errors
/// [`PansharpenError::BandCount`] if `ms` does not have exactly 3 bands;
/// [`PansharpenError::ShapeMismatch`] if `pan` shape disagrees;
/// [`PansharpenError::Empty`] for empty inputs.
pub fn ihs(
    ms: ArrayView3<'_, f32>,
    pan: ArrayView2<'_, f32>,
) -> Result<Array3<f32>, PansharpenError> {
    let (b, r, c) = ms.dim();
    if b == 0 || r == 0 || c == 0 {
        return Err(PansharpenError::Empty);
    }
    if b != 3 {
        return Err(PansharpenError::BandCount { kernel: "ihs", expected: 3, got: b });
    }
    if pan.dim() != (r, c) {
        return Err(PansharpenError::ShapeMismatch { expected: (r, c), got: pan.dim() });
    }

    // Intensity = (R + G + B) / 3.
    let mut intensity = ndarray::Array2::<f32>::zeros((r, c));
    for band in 0..3 {
        for i in 0..r {
            for j in 0..c {
                intensity[(i, j)] += ms[(band, i, j)];
            }
        }
    }
    intensity.mapv_inplace(|v| v / 3.0);

    let pan_matched = histogram_match(pan, intensity.view())?;

    let mut out = Array3::<f32>::zeros((3, r, c));
    for band in 0..3 {
        for i in 0..r {
            for j in 0..c {
                let delta = pan_matched[(i, j)] - intensity[(i, j)];
                out[(band, i, j)] = ms[(band, i, j)] + delta;
            }
        }
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::{Array2, Array3};

    use super::*;

    #[test]
    fn ihs_pan_equals_intensity_is_identity() {
        let ms = Array3::from_shape_vec(
            (3, 2, 2),
            vec![0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2],
        )
        .unwrap();
        let mut intensity = Array2::<f32>::zeros((2, 2));
        for b in 0..3 {
            for i in 0..2 {
                for j in 0..2 {
                    intensity[(i, j)] += ms[(b, i, j)];
                }
            }
        }
        intensity.mapv_inplace(|v| v / 3.0);
        let out = ihs(ms.view(), intensity.view()).unwrap();
        for b in 0..3 {
            for i in 0..2 {
                for j in 0..2 {
                    assert_abs_diff_eq!(out[(b, i, j)], ms[(b, i, j)], epsilon = 1e-5);
                }
            }
        }
    }

    #[test]
    fn ihs_rejects_non_three_bands() {
        let ms = Array3::<f32>::zeros((4, 2, 2));
        let pan = Array2::<f32>::zeros((2, 2));
        assert!(matches!(
            ihs(ms.view(), pan.view()).unwrap_err(),
            PansharpenError::BandCount { expected: 3, .. }
        ));
    }
}
