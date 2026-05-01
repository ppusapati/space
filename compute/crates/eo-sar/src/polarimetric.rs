//! Polarimetric SAR decomposition.
//!
//! Operates on full-pol scattering vectors per pixel using the Pauli basis
//! `k = (1/√2) [S_HH + S_VV, S_HH − S_VV, 2 S_HV]ᵀ`.
//!
//! * [`pauli_decomposition`] — returns the three Pauli intensity images
//!   `|k_1|², |k_2|², |k_3|²` (often shown in red / green / blue).
//! * [`cloude_pottier`] — eigenvalue decomposition of the 3×3 coherency
//!   matrix `[T] = E[k · kᴴ]` averaged over a window. Returns Entropy `H`,
//!   Anisotropy `A`, and mean alpha angle `ᾱ` per pixel.

use nalgebra::Matrix3;
use ndarray::{Array2, ArrayView2};
use num_complex::Complex32;

use crate::SarError;

/// Outputs of [`pauli_decomposition`].
#[derive(Debug, Clone)]
pub struct Pauli {
    /// `|S_HH + S_VV|² / 2` (red / single-bounce).
    pub red: Array2<f32>,
    /// `|S_HH − S_VV|² / 2` (green / double-bounce).
    pub green: Array2<f32>,
    /// `|2 S_HV|² / 2 = 2|S_HV|²` (blue / volume).
    pub blue: Array2<f32>,
}

/// Compute the Pauli decomposition.
///
/// # Errors
/// [`SarError::ShapeMismatch`] if any band shape disagrees;
/// [`SarError::Empty`] for empty input.
pub fn pauli_decomposition(
    s_hh: ArrayView2<'_, Complex32>,
    s_hv: ArrayView2<'_, Complex32>,
    s_vv: ArrayView2<'_, Complex32>,
) -> Result<Pauli, SarError> {
    let dim = s_hh.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(SarError::Empty);
    }
    if s_hv.dim() != dim {
        return Err(SarError::ShapeMismatch { expected: dim, got: s_hv.dim() });
    }
    if s_vv.dim() != dim {
        return Err(SarError::ShapeMismatch { expected: dim, got: s_vv.dim() });
    }
    let mut red = Array2::<f32>::zeros(dim);
    let mut green = Array2::<f32>::zeros(dim);
    let mut blue = Array2::<f32>::zeros(dim);
    ndarray::Zip::from(&mut red)
        .and(&mut green)
        .and(&mut blue)
        .and(s_hh)
        .and(s_hv)
        .and(s_vv)
        .for_each(|r, g, b, &hh, &hv, &vv| {
            *r = (hh + vv).norm_sqr() * 0.5;
            *g = (hh - vv).norm_sqr() * 0.5;
            *b = 2.0 * hv.norm_sqr();
        });
    Ok(Pauli { red, green, blue })
}

/// Outputs of [`cloude_pottier`].
#[derive(Debug, Clone)]
pub struct ClouePottier {
    /// Entropy `H ∈ [0, 1]`.
    pub entropy: Array2<f32>,
    /// Anisotropy `A ∈ [0, 1]`.
    pub anisotropy: Array2<f32>,
    /// Mean alpha angle `ᾱ ∈ [0, π/2]` (radians).
    pub alpha: Array2<f32>,
}

/// Cloude-Pottier H/A/Alpha decomposition over a square boxcar averaging
/// window (`window`, odd ≥ 3).
///
/// # Errors
/// [`SarError`] for shape, emptiness, or window violations.
pub fn cloude_pottier(
    s_hh: ArrayView2<'_, Complex32>,
    s_hv: ArrayView2<'_, Complex32>,
    s_vv: ArrayView2<'_, Complex32>,
    window: usize,
) -> Result<ClouePottier, SarError> {
    if window < 3 || window % 2 == 0 {
        return Err(SarError::OutOfRange {
            name: "window",
            value: window as f64,
            range: "odd >= 3",
        });
    }
    let dim = s_hh.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(SarError::Empty);
    }
    if s_hv.dim() != dim {
        return Err(SarError::ShapeMismatch { expected: dim, got: s_hv.dim() });
    }
    if s_vv.dim() != dim {
        return Err(SarError::ShapeMismatch { expected: dim, got: s_vv.dim() });
    }
    let half = window / 2;
    let (rows, cols) = dim;
    let mut entropy = Array2::<f32>::zeros(dim);
    let mut anisotropy = Array2::<f32>::zeros(dim);
    let mut alpha = Array2::<f32>::zeros(dim);
    let sqrt2 = 2.0_f64.sqrt();
    for r in 0..rows {
        for c in 0..cols {
            let r0 = r.saturating_sub(half);
            let r1 = (r + half).min(rows - 1);
            let c0 = c.saturating_sub(half);
            let c1 = (c + half).min(cols - 1);
            let mut t = Matrix3::<num_complex::Complex<f64>>::zeros();
            let mut n = 0u32;
            for rr in r0..=r1 {
                for cc in c0..=c1 {
                    let hh = c32_to_c64(s_hh[(rr, cc)]);
                    let hv = c32_to_c64(s_hv[(rr, cc)]);
                    let vv = c32_to_c64(s_vv[(rr, cc)]);
                    let k1 = (hh + vv) / sqrt2;
                    let k2 = (hh - vv) / sqrt2;
                    let k3 = num_complex::Complex::<f64>::new(2.0, 0.0) * hv / sqrt2;
                    // T = k k^H (3x3 Hermitian).
                    let k = nalgebra::Vector3::new(k1, k2, k3);
                    let outer = k * k.adjoint();
                    t += outer;
                    n += 1;
                }
            }
            if n == 0 {
                continue;
            }
            let inv_n = 1.0 / f64::from(n);
            t *= num_complex::Complex::<f64>::new(inv_n, 0.0);
            // Hermitian eigendecomposition: build a 3x3 Hermitian matrix.
            let (eigvals, eigvecs) = hermitian_eig3(&t);
            let total: f64 = eigvals.iter().sum();
            if total <= 0.0 {
                continue;
            }
            let mut p = [eigvals[0] / total, eigvals[1] / total, eigvals[2] / total];
            for v in &mut p {
                *v = v.max(0.0);
            }
            let h: f64 = p
                .iter()
                .filter(|&&pv| pv > 0.0)
                .map(|&pv| -pv * pv.log(3.0))
                .sum();
            // Anisotropy A = (λ2 - λ3) / (λ2 + λ3) using sorted eigvals descending.
            let mut sorted = eigvals;
            sorted.sort_by(|a, b| b.partial_cmp(a).unwrap_or(std::cmp::Ordering::Equal));
            let a_val = if sorted[1] + sorted[2] > 0.0 {
                (sorted[1] - sorted[2]) / (sorted[1] + sorted[2])
            } else {
                0.0
            };
            // Mean alpha: ᾱ = Σ p_i α_i, α_i = acos(|u_i[0]|).
            let mean_alpha: f64 = (0..3)
                .map(|i| p[i] * (eigvecs[i].0.norm()).acos())
                .sum();
            entropy[(r, c)] = h.clamp(0.0, 1.0) as f32;
            anisotropy[(r, c)] = a_val.clamp(0.0, 1.0) as f32;
            alpha[(r, c)] = mean_alpha as f32;
        }
    }
    Ok(ClouePottier { entropy, anisotropy, alpha })
}

#[inline]
fn c32_to_c64(z: Complex32) -> num_complex::Complex<f64> {
    num_complex::Complex::<f64>::new(f64::from(z.re), f64::from(z.im))
}

/// Hermitian eigen-decomposition of a 3×3 matrix.
///
/// Returns `(eigvals, [u1, u2, u3])` where `u_i` is the eigenvector
/// corresponding to eigenvalue `eigvals[i]`. Uses nalgebra's Schur on the
/// real 6×6 embedding `[[A_re, -A_im],[A_im, A_re]]` because nalgebra has no
/// complex Hermitian eigensolver in stable releases.
///
/// We instead use a direct Jacobi rotation method on a real 6×6 symmetric
/// embedding, then collapse pairs of eigenvalues.
fn hermitian_eig3(
    t: &Matrix3<num_complex::Complex<f64>>,
) -> (
    [f64; 3],
    [(num_complex::Complex<f64>, num_complex::Complex<f64>, num_complex::Complex<f64>); 3],
) {
    use nalgebra::Matrix6;

    let mut a = Matrix6::<f64>::zeros();
    for i in 0..3 {
        for j in 0..3 {
            let z = t[(i, j)];
            a[(i, j)] = z.re;
            a[(i + 3, j + 3)] = z.re;
            a[(i, j + 3)] = -z.im;
            a[(i + 3, j)] = z.im;
        }
    }
    let eigen = a.symmetric_eigen();
    // Eigen pairs are in arbitrary order; eigenvalues come in duplicate
    // pairs because of the embedding.
    let mut pairs: Vec<(f64, [f64; 6])> = (0..6)
        .map(|i| {
            let v = eigen.eigenvectors.column(i);
            let mut arr = [0.0_f64; 6];
            for k in 0..6 {
                arr[k] = v[k];
            }
            (eigen.eigenvalues[i], arr)
        })
        .collect();
    pairs.sort_by(|x, y| y.0.partial_cmp(&x.0).unwrap_or(std::cmp::Ordering::Equal));
    // Take three (the three largest distinct eigenvalues, accounting for
    // pair degeneracy by skipping every other one).
    let chosen = [&pairs[0], &pairs[2], &pairs[4]];
    let mut vals = [0.0_f64; 3];
    let zero = num_complex::Complex::<f64>::new(0.0, 0.0);
    let mut vecs: [(num_complex::Complex<f64>, num_complex::Complex<f64>, num_complex::Complex<f64>); 3] =
        [(zero, zero, zero); 3];
    for (k, p) in chosen.iter().enumerate() {
        vals[k] = p.0;
        let v = p.1;
        // Recover complex eigenvector u = (v[0..3]) + i (v[3..6])
        let u = (
            num_complex::Complex::<f64>::new(v[0], v[3]),
            num_complex::Complex::<f64>::new(v[1], v[4]),
            num_complex::Complex::<f64>::new(v[2], v[5]),
        );
        // Normalise.
        let norm = (u.0.norm_sqr() + u.1.norm_sqr() + u.2.norm_sqr()).sqrt();
        if norm > 0.0 {
            vecs[k] = (
                u.0 / num_complex::Complex::<f64>::new(norm, 0.0),
                u.1 / num_complex::Complex::<f64>::new(norm, 0.0),
                u.2 / num_complex::Complex::<f64>::new(norm, 0.0),
            );
        } else {
            vecs[k] = u;
        }
    }
    (vals, vecs)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn pauli_zeros_for_zero_input() {
        let s = array![[Complex32::new(0.0, 0.0)]];
        let p = pauli_decomposition(s.view(), s.view(), s.view()).unwrap();
        assert_abs_diff_eq!(p.red[(0, 0)], 0.0);
        assert_abs_diff_eq!(p.green[(0, 0)], 0.0);
        assert_abs_diff_eq!(p.blue[(0, 0)], 0.0);
    }

    #[test]
    fn pauli_pure_double_bounce() {
        // S_HH = -S_VV, S_HV = 0  → red = 0, green = 2|S_HH|², blue = 0.
        let hh = array![[Complex32::new(1.0, 0.0)]];
        let hv = array![[Complex32::new(0.0, 0.0)]];
        let vv = array![[Complex32::new(-1.0, 0.0)]];
        let p = pauli_decomposition(hh.view(), hv.view(), vv.view()).unwrap();
        assert_abs_diff_eq!(p.red[(0, 0)], 0.0, epsilon = 1e-6);
        assert_abs_diff_eq!(p.green[(0, 0)], 2.0, epsilon = 1e-6);
        assert_abs_diff_eq!(p.blue[(0, 0)], 0.0, epsilon = 1e-6);
    }

    #[test]
    fn cloude_pottier_pure_surface_low_entropy() {
        // Pure single-bounce: S_HH = S_VV, S_HV = 0 → only k1 non-zero,
        // T = diag(2, 0, 0) → λ = (2, 0, 0) → H = 0, A = 0, α = 0.
        let n = 3;
        let hh = ndarray::Array2::<Complex32>::from_elem((n, n), Complex32::new(1.0, 0.0));
        let hv = ndarray::Array2::<Complex32>::from_elem((n, n), Complex32::new(0.0, 0.0));
        let vv = ndarray::Array2::<Complex32>::from_elem((n, n), Complex32::new(1.0, 0.0));
        let cp = cloude_pottier(hh.view(), hv.view(), vv.view(), 3).unwrap();
        assert!(cp.entropy[(1, 1)] < 1e-3);
        assert!(cp.alpha[(1, 1)] < 1e-3);
    }
}
