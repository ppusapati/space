//! Classical Gram-Schmidt pan-sharpening (Laben & Brower, 1998).
//!
//! 1. Build a synthetic low-resolution Pan from a weighted sum of the MS bands.
//! 2. Treat that synthetic Pan as the first GS component (`GS_1`).
//! 3. Orthogonalise each MS band against the prior GS components.
//! 4. Histogram-match the high-resolution Pan to `GS_1`.
//! 5. Replace `GS_1` with the matched Pan.
//! 6. Invert the GS transform: add the projection of every higher-order GS
//!    component back through the same coefficients.

use ndarray::{Array2, Array3, ArrayView2, ArrayView3};

use crate::{PansharpenError, hist_match::histogram_match};

/// Per-band weights used to synthesise the low-resolution Pan in step 1.
#[derive(Debug, Clone)]
pub struct GsWeights {
    /// Length must equal the number of MS bands; entries must sum to a finite
    /// non-zero value. Weights are normalised internally.
    pub weights: Vec<f32>,
}

impl GsWeights {
    /// Equal weights `[1/B, …, 1/B]` for `b_count` bands.
    #[must_use]
    pub fn equal(b_count: usize) -> Self {
        Self { weights: vec![1.0 / (b_count as f32); b_count] }
    }
}

/// Apply Gram-Schmidt pan-sharpening.
///
/// # Errors
/// Returns [`PansharpenError`] for shape, band-count, or weight issues.
pub fn gram_schmidt(
    ms: ArrayView3<'_, f32>,
    pan: ArrayView2<'_, f32>,
    weights: &GsWeights,
) -> Result<Array3<f32>, PansharpenError> {
    let (b, r, c) = ms.dim();
    if b == 0 || r == 0 || c == 0 {
        return Err(PansharpenError::Empty);
    }
    if pan.dim() != (r, c) {
        return Err(PansharpenError::ShapeMismatch { expected: (r, c), got: pan.dim() });
    }
    if weights.weights.len() != b {
        return Err(PansharpenError::BandCount {
            kernel: "gram_schmidt",
            expected: b,
            got: weights.weights.len(),
        });
    }
    let wsum: f32 = weights.weights.iter().copied().sum();
    if !wsum.is_finite() || wsum == 0.0 {
        return Err(PansharpenError::Algorithm {
            kernel: "gram_schmidt",
            reason: "weights sum to zero",
        });
    }
    let w: Vec<f32> = weights.weights.iter().map(|x| x / wsum).collect();

    // Step 1: synthetic low-res Pan = Σ w_i · M_i.
    let mut gs_components: Vec<Array2<f32>> = Vec::with_capacity(b + 1);
    let mut synthetic = Array2::<f32>::zeros((r, c));
    for (band, &wi) in w.iter().enumerate() {
        for i in 0..r {
            for j in 0..c {
                synthetic[(i, j)] += wi * ms[(band, i, j)];
            }
        }
    }
    gs_components.push(synthetic);

    // Step 2-3: GS orthogonalise each MS band against prior components.
    // gs_b = M_b - Σ_{k<b} ((M_b · gs_k) / (gs_k · gs_k)) · gs_k
    // We store coefficients to reuse on the inverse step.
    let mut coefficients: Vec<Vec<f32>> = vec![Vec::new(); b];
    for band in 0..b {
        let mb_view = ms.index_axis(ndarray::Axis(0), band);
        let mut gs_b = mb_view.to_owned();
        let mut row = Vec::with_capacity(band + 1);
        for k in 0..=band {
            let gsk = &gs_components[k];
            let dot_mk: f32 = ndarray::Zip::from(&gs_b).and(gsk).fold(0.0_f32, |s, &a, &kv| s + a * kv);
            let dot_kk: f32 = gsk.iter().map(|v| v * v).sum();
            let coeff = if dot_kk == 0.0 { 0.0 } else { dot_mk / dot_kk };
            row.push(coeff);
            // gs_b -= coeff * gsk
            ndarray::Zip::from(&mut gs_b).and(gsk).for_each(|x, &kv| *x -= coeff * kv);
        }
        coefficients[band] = row;
        gs_components.push(gs_b);
    }

    // Step 4: histogram-match Pan to GS_1.
    let pan_matched = histogram_match(pan, gs_components[0].view())?;

    // Step 5: keep the *original* GS_1 mean/std for sanity but substitute.
    // Replace gs_components[0] with pan_matched.
    gs_components[0] = pan_matched;

    // Step 6: inverse Gram-Schmidt — reconstruct each MS band:
    // M_b' = gs_{b+1} + Σ_{k<=b} coefficients[b][k] · gs_components[k]
    let mut out = Array3::<f32>::zeros((b, r, c));
    for band in 0..b {
        let mut accum = gs_components[band + 1].clone();
        for (k, &coeff) in coefficients[band].iter().enumerate() {
            ndarray::Zip::from(&mut accum)
                .and(&gs_components[k])
                .for_each(|x, &kv| *x += coeff * kv);
        }
        for i in 0..r {
            for j in 0..c {
                out[(band, i, j)] = accum[(i, j)];
            }
        }
    }

    debug_assert!(out.iter().all(|v| v.is_finite()));
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::{Array2, Array3};

    use super::*;

    #[test]
    fn gs_pan_equals_synthetic_is_identity() {
        let ms = Array3::from_shape_vec(
            (3, 4, 4),
            (0..48).map(|x| 0.1 + (x as f32) * 0.01).collect(),
        )
        .unwrap();
        let weights = GsWeights::equal(3);
        // Synthetic Pan = Σ w_i M_i with equal weights.
        let mut synthetic = Array2::<f32>::zeros((4, 4));
        for b in 0..3 {
            for i in 0..4 {
                for j in 0..4 {
                    synthetic[(i, j)] += ms[(b, i, j)] / 3.0;
                }
            }
        }
        let out = gram_schmidt(ms.view(), synthetic.view(), &weights).unwrap();
        for b in 0..3 {
            for i in 0..4 {
                for j in 0..4 {
                    assert_abs_diff_eq!(out[(b, i, j)], ms[(b, i, j)], epsilon = 1e-4);
                }
            }
        }
    }

    #[test]
    fn gs_rejects_zero_weights() {
        let ms = Array3::<f32>::zeros((3, 2, 2));
        let pan = Array2::<f32>::zeros((2, 2));
        let w = GsWeights { weights: vec![0.0, 0.0, 0.0] };
        assert!(gram_schmidt(ms.view(), pan.view(), &w).is_err());
    }
}
