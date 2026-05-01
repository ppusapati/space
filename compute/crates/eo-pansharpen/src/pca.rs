//! PCA-based pan-sharpening.
//!
//! 1. Centre each band by its mean.
//! 2. Compute the symmetric `B×B` covariance matrix.
//! 3. Eigendecompose; sort eigenvectors by descending eigenvalue.
//! 4. Project the centred MS into PC space.
//! 5. Histogram-match Pan to PC1.
//! 6. Replace PC1 with matched Pan, inverse-rotate, add band means.

use nalgebra::DMatrix;
use ndarray::{Array3, ArrayView2, ArrayView3};

use crate::{PansharpenError, hist_match::histogram_match};

/// Apply PCA pan-sharpening for any band count `≥ 2`.
///
/// # Errors
/// Returns [`PansharpenError::BandCount`] if `ms` has < 2 bands,
/// [`PansharpenError::ShapeMismatch`] if `pan` disagrees, or
/// [`PansharpenError::Algorithm`] if eigendecomposition fails.
pub fn pca(
    ms: ArrayView3<'_, f32>,
    pan: ArrayView2<'_, f32>,
) -> Result<Array3<f32>, PansharpenError> {
    let (b, r, c) = ms.dim();
    if b == 0 || r == 0 || c == 0 {
        return Err(PansharpenError::Empty);
    }
    if b < 2 {
        return Err(PansharpenError::BandCount { kernel: "pca", expected: 2, got: b });
    }
    if pan.dim() != (r, c) {
        return Err(PansharpenError::ShapeMismatch { expected: (r, c), got: pan.dim() });
    }

    let n = r * c;
    let n_f64 = n as f64;

    // Means per band and centred (column-major) data matrix N×B (in f64 for stability).
    let mut means = vec![0.0_f64; b];
    let mut centred = DMatrix::<f64>::zeros(n, b);
    for band in 0..b {
        let s: f64 = (0..r).flat_map(|i| (0..c).map(move |j| (i, j)))
            .map(|(i, j)| f64::from(ms[(band, i, j)]))
            .sum();
        let mu = s / n_f64;
        means[band] = mu;
        let mut idx = 0;
        for i in 0..r {
            for j in 0..c {
                centred[(idx, band)] = f64::from(ms[(band, i, j)]) - mu;
                idx += 1;
            }
        }
    }

    // Covariance B×B = (1/N) · Cᵀ · C
    let cov = centred.transpose() * &centred / n_f64;
    let eigen = cov.symmetric_eigen();
    // Sort eigenvalues descending, permute eigenvectors.
    let mut idx: Vec<usize> = (0..b).collect();
    idx.sort_by(|&a, &z| {
        eigen.eigenvalues[z].partial_cmp(&eigen.eigenvalues[a]).unwrap_or(std::cmp::Ordering::Equal)
    });
    let mut eigvecs = DMatrix::<f64>::zeros(b, b);
    for (k, &col) in idx.iter().enumerate() {
        eigvecs.set_column(k, &eigen.eigenvectors.column(col));
    }

    // Project: PC = centred · eigvecs. Shape N×B.
    let pcs = &centred * &eigvecs;

    // Build PC1 array as 2-D for histogram matching.
    let mut pc1 = ndarray::Array2::<f32>::zeros((r, c));
    let mut idx2 = 0;
    for i in 0..r {
        for j in 0..c {
            pc1[(i, j)] = pcs[(idx2, 0)] as f32;
            idx2 += 1;
        }
    }
    let pan_matched = histogram_match(pan, pc1.view())?;

    // Substitute PC1 with matched Pan (kept in f32 → cast to f64 for back-projection).
    let mut pcs_replaced = pcs;
    let mut idx3 = 0;
    for i in 0..r {
        for j in 0..c {
            pcs_replaced[(idx3, 0)] = f64::from(pan_matched[(i, j)]);
            idx3 += 1;
        }
    }

    // Inverse rotation: out_centred = pcs_replaced · eigvecsᵀ
    let recon = pcs_replaced * eigvecs.transpose();

    // Add per-band means and reshape back to (B, R, C).
    let mut out = Array3::<f32>::zeros((b, r, c));
    let mut row = 0;
    for i in 0..r {
        for j in 0..c {
            for band in 0..b {
                out[(band, i, j)] = (recon[(row, band)] + means[band]) as f32;
            }
            row += 1;
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
    fn pca_reconstructs_when_pan_equals_pc1() {
        // A simple cube whose first PC carries most variance.
        let ms = Array3::from_shape_fn((4, 8, 8), |(b, i, j)| {
            (b as f32) * 0.05 + (i as f32) * 0.01 + (j as f32) * 0.001
        });
        // Compute PC1 by running pca with a placeholder Pan, then re-run with that PC1.
        let pan_placeholder = Array2::<f32>::zeros((8, 8));
        let _ = pca(ms.view(), pan_placeholder.view()).unwrap();
        // Round-trip: when we feed back the *first* principal component of the input
        // as the "high-resolution" Pan, the output must be identical to the input.
        // Compute PC1 ourselves:
        let n = 64;
        let mut means = vec![0.0_f64; 4];
        let mut centred = nalgebra::DMatrix::<f64>::zeros(n, 4);
        for b in 0..4 {
            let mu: f64 = (0..8)
                .flat_map(|i| (0..8).map(move |j| (i, j)))
                .map(|(i, j)| f64::from(ms[(b, i, j)]))
                .sum::<f64>()
                / 64.0;
            means[b] = mu;
            let mut idx = 0;
            for i in 0..8 {
                for j in 0..8 {
                    centred[(idx, b)] = f64::from(ms[(b, i, j)]) - mu;
                    idx += 1;
                }
            }
        }
        let cov = centred.transpose() * &centred / 64.0;
        let eig = cov.symmetric_eigen();
        let mut order: Vec<usize> = (0..4).collect();
        order.sort_by(|&a, &z| {
            eig.eigenvalues[z]
                .partial_cmp(&eig.eigenvalues[a])
                .unwrap_or(std::cmp::Ordering::Equal)
        });
        let pc_col = &eig.eigenvectors.column(order[0]);
        let pc1_vec = &centred * pc_col;
        let mut pc1_img = Array2::<f32>::zeros((8, 8));
        for i in 0..8 {
            for j in 0..8 {
                pc1_img[(i, j)] = pc1_vec[(i * 8 + j, 0)] as f32;
            }
        }
        let out = pca(ms.view(), pc1_img.view()).unwrap();
        for b in 0..4 {
            for i in 0..8 {
                for j in 0..8 {
                    assert_abs_diff_eq!(out[(b, i, j)], ms[(b, i, j)], epsilon = 1e-3);
                }
            }
        }
    }

    #[test]
    fn pca_rejects_single_band() {
        let ms = Array3::<f32>::zeros((1, 2, 2));
        let pan = Array2::<f32>::zeros((2, 2));
        assert!(matches!(
            pca(ms.view(), pan.view()).unwrap_err(),
            PansharpenError::BandCount { .. }
        ));
    }
}
