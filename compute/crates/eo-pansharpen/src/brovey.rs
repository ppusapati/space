//! Brovey transform pan-sharpening.

use ndarray::{Array2, Array3, ArrayView2, ArrayView3};

use crate::PansharpenError;

/// Apply the Brovey transform.
///
/// `ms` has shape `(bands, rows, cols)`; `pan` has shape `(rows, cols)` at
/// the same spatial resolution as the multispectral bands.
///
/// `M'_i = M_i · Pan / I` where `I = (1/N) Σ M_i`. Pixels where `I == 0`
/// produce `0` (the band has no spectral content there).
///
/// # Errors
/// Returns [`PansharpenError`] when shapes mismatch or inputs are empty.
pub fn brovey(
    ms: ArrayView3<'_, f32>,
    pan: ArrayView2<'_, f32>,
) -> Result<Array3<f32>, PansharpenError> {
    let (b, r, c) = ms.dim();
    if b == 0 || r == 0 || c == 0 {
        return Err(PansharpenError::Empty);
    }
    if pan.dim() != (r, c) {
        return Err(PansharpenError::ShapeMismatch { expected: (r, c), got: pan.dim() });
    }
    let mut intensity = Array2::<f32>::zeros((r, c));
    for band in 0..b {
        for i in 0..r {
            for j in 0..c {
                intensity[(i, j)] += ms[(band, i, j)];
            }
        }
    }
    intensity.mapv_inplace(|v| v / b as f32);

    let mut out = Array3::<f32>::zeros((b, r, c));
    for band in 0..b {
        for i in 0..r {
            for j in 0..c {
                let denom = intensity[(i, j)];
                out[(band, i, j)] =
                    if denom == 0.0 { 0.0 } else { ms[(band, i, j)] * pan[(i, j)] / denom };
            }
        }
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::Array3;

    use super::*;

    #[test]
    fn brovey_pan_equals_intensity_is_identity() {
        let ms = Array3::from_shape_vec(
            (3, 2, 2),
            vec![0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2],
        )
        .unwrap();
        // Pan == intensity per pixel.
        let mut intensity = Array2::<f32>::zeros((2, 2));
        for b in 0..3 {
            for i in 0..2 {
                for j in 0..2 {
                    intensity[(i, j)] += ms[(b, i, j)];
                }
            }
        }
        intensity.mapv_inplace(|v| v / 3.0);
        let out = brovey(ms.view(), intensity.view()).unwrap();
        for b in 0..3 {
            for i in 0..2 {
                for j in 0..2 {
                    assert_abs_diff_eq!(out[(b, i, j)], ms[(b, i, j)], epsilon = 1e-6);
                }
            }
        }
    }

    #[test]
    fn brovey_zero_intensity_gives_zero() {
        let ms = Array3::<f32>::zeros((3, 2, 2));
        let pan = Array2::<f32>::from_elem((2, 2), 0.5);
        let out = brovey(ms.view(), pan.view()).unwrap();
        assert!(out.iter().all(|&v| v == 0.0));
    }

    #[test]
    fn brovey_shape_mismatch() {
        let ms = Array3::<f32>::zeros((3, 2, 2));
        let pan = Array2::<f32>::zeros((2, 3));
        assert!(brovey(ms.view(), pan.view()).is_err());
    }
}
