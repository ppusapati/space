//! Adaptive speckle filters.
//!
//! * **Lee filter** (Lee, 1980): linear MMSE estimator using local mean and
//!   variance, modulated by a noise variance estimate `σ²_v`.
//! * **Frost filter** (Frost et al., 1982): exponentially weighted local
//!   mean with damping factor `K` proportional to the local coefficient of
//!   variation.

use ndarray::{Array2, ArrayView2};

use crate::SarError;

/// Apply the Lee filter with a square `window` (odd ≥ 3) and a noise
/// variance `noise_var` (in the same units as the squared pixel value).
///
/// Output:
/// `R = mean + W · (I − mean)` with `W = max(0, σ_x²) / σ_x²`,
/// `σ_x² = (σ² − noise_var · mean²) / (1 + noise_var)`.
///
/// # Errors
/// [`SarError::OutOfRange`] for invalid `window` or negative `noise_var`;
/// [`SarError::Empty`] for empty input.
pub fn lee(
    intensity: ArrayView2<'_, f32>,
    window: usize,
    noise_var: f32,
) -> Result<Array2<f32>, SarError> {
    validate_window(window)?;
    if !(noise_var.is_finite() && noise_var >= 0.0) {
        return Err(SarError::OutOfRange {
            name: "noise_var",
            value: f64::from(noise_var),
            range: "[0, +inf)",
        });
    }
    let (rows, cols) = intensity.dim();
    if rows == 0 || cols == 0 {
        return Err(SarError::Empty);
    }
    let half = window / 2;
    let mut out = Array2::<f32>::zeros((rows, cols));
    for r in 0..rows {
        for c in 0..cols {
            let (mean, var) = local_mean_var(intensity, r, c, half);
            let mean_sq = mean * mean;
            let sigma_x2 = (var - noise_var * mean_sq).max(0.0) / (1.0 + noise_var);
            let w = if var > 0.0 { sigma_x2 / var } else { 0.0 };
            out[(r, c)] = w.mul_add(intensity[(r, c)] - mean, mean);
        }
    }
    Ok(out)
}

/// Apply the Frost filter with a square `window` and damping factor `k`.
///
/// `out(p) = Σ I(q)·m(q) / Σ m(q)`,
/// where `m(q) = exp(−k · CV² · |p − q|)` and `CV` is the local
/// coefficient of variation `σ / μ`.
///
/// # Errors
/// [`SarError::OutOfRange`] for invalid `window` or non-finite `k`.
pub fn frost(
    intensity: ArrayView2<'_, f32>,
    window: usize,
    k: f32,
) -> Result<Array2<f32>, SarError> {
    validate_window(window)?;
    if !k.is_finite() {
        return Err(SarError::OutOfRange {
            name: "k",
            value: f64::from(k),
            range: "finite",
        });
    }
    let (rows, cols) = intensity.dim();
    if rows == 0 || cols == 0 {
        return Err(SarError::Empty);
    }
    let half = window / 2;
    let mut out = Array2::<f32>::zeros((rows, cols));
    for r in 0..rows {
        for c in 0..cols {
            let (mean, var) = local_mean_var(intensity, r, c, half);
            let cv = if mean > 0.0 { var.sqrt() / mean } else { 0.0 };
            let damping = k * cv * cv;
            let r0 = r.saturating_sub(half);
            let r1 = (r + half).min(rows - 1);
            let c0 = c.saturating_sub(half);
            let c1 = (c + half).min(cols - 1);
            let mut num = 0.0_f32;
            let mut den = 0.0_f32;
            for rr in r0..=r1 {
                for cc in c0..=c1 {
                    let dr = (rr as f32 - r as f32).abs();
                    let dc = (cc as f32 - c as f32).abs();
                    let dist = (dr * dr + dc * dc).sqrt();
                    let w = (-damping * dist).exp();
                    num += w * intensity[(rr, cc)];
                    den += w;
                }
            }
            out[(r, c)] = if den > 0.0 { num / den } else { intensity[(r, c)] };
        }
    }
    Ok(out)
}

fn validate_window(window: usize) -> Result<(), SarError> {
    if window < 3 || window % 2 == 0 {
        return Err(SarError::OutOfRange {
            name: "window",
            value: window as f64,
            range: "odd >= 3",
        });
    }
    Ok(())
}

fn local_mean_var(intensity: ArrayView2<'_, f32>, r: usize, c: usize, half: usize) -> (f32, f32) {
    let (rows, cols) = intensity.dim();
    let r0 = r.saturating_sub(half);
    let r1 = (r + half).min(rows - 1);
    let c0 = c.saturating_sub(half);
    let c1 = (c + half).min(cols - 1);
    let n = ((r1 - r0 + 1) * (c1 - c0 + 1)) as f32;
    let mut s = 0.0_f32;
    let mut s2 = 0.0_f32;
    for rr in r0..=r1 {
        for cc in c0..=c1 {
            let v = intensity[(rr, cc)];
            s += v;
            s2 += v * v;
        }
    }
    let mean = s / n;
    let var = (s2 / n) - mean * mean;
    (mean, var.max(0.0))
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::Array2;

    use super::*;

    #[test]
    fn lee_constant_input_returns_input() {
        let i = Array2::<f32>::from_elem((10, 10), 4.0);
        let out = lee(i.view(), 5, 0.5).unwrap();
        for v in &out {
            assert_abs_diff_eq!(*v, 4.0, epsilon = 1e-5);
        }
    }

    #[test]
    fn lee_high_variance_passes_pixel_through() {
        // With noise_var=0 the filter weight w = 1, so the output equals input.
        let i = Array2::<f32>::from_shape_fn((6, 6), |(r, c)| (r as f32) + 5.0 * (c as f32));
        let out = lee(i.view(), 3, 0.0).unwrap();
        for r in 0..6 {
            for c in 0..6 {
                assert_abs_diff_eq!(out[(r, c)], i[(r, c)], epsilon = 1e-4);
            }
        }
    }

    #[test]
    fn frost_constant_input_returns_input() {
        let i = Array2::<f32>::from_elem((10, 10), 7.0);
        let out = frost(i.view(), 5, 1.0).unwrap();
        for v in &out {
            assert_abs_diff_eq!(*v, 7.0, epsilon = 1e-5);
        }
    }

    #[test]
    fn lee_invalid_window() {
        let i = Array2::<f32>::zeros((4, 4));
        assert!(lee(i.view(), 4, 0.5).is_err());
        assert!(lee(i.view(), 1, 0.5).is_err());
    }
}
