//! Discrete-time Linear Quadratic Regulator.
//!
//! For the LTI plant `x_{k+1} = A x_k + B u_k` and quadratic cost
//!
//! ```text
//! J = Σ_{k=0}^{∞} xₖᵀ Q xₖ + uₖᵀ R uₖ
//! ```
//!
//! the optimal feedback law is `u_k = −K · x_k` where
//!
//! ```text
//! K = (R + Bᵀ P B)⁻¹ Bᵀ P A
//! ```
//!
//! and `P` is the unique positive-definite solution of the Discrete
//! Algebraic Riccati Equation (DARE):
//!
//! ```text
//! P = Aᵀ P A − Aᵀ P B (R + Bᵀ P B)⁻¹ Bᵀ P A + Q
//! ```
//!
//! This module solves the DARE iteratively (Hewer 1971): start with
//! `P₀ = Q` and iterate the equation until `‖Pₙ₊₁ − Pₙ‖_F < ε`.

use nalgebra::DMatrix;

use crate::ControlError;

/// Solve the DARE and return `(P, K)`.
///
/// # Errors
/// [`ControlError::DimensionMismatch`] if dimensions disagree;
/// [`ControlError::DidNotConverge`] if the iteration fails to reduce the
/// residual below `tol` within `max_iters`.
pub fn solve_dare(
    a: &DMatrix<f64>,
    b: &DMatrix<f64>,
    q: &DMatrix<f64>,
    r: &DMatrix<f64>,
    tol: f64,
    max_iters: u32,
) -> Result<(DMatrix<f64>, DMatrix<f64>), ControlError> {
    let n = a.nrows();
    let m = b.ncols();
    if a.shape() != (n, n) {
        return Err(ControlError::DimensionMismatch("A must be n×n"));
    }
    if b.shape() != (n, m) {
        return Err(ControlError::DimensionMismatch("B must be n×m"));
    }
    if q.shape() != (n, n) {
        return Err(ControlError::DimensionMismatch("Q must be n×n"));
    }
    if r.shape() != (m, m) {
        return Err(ControlError::DimensionMismatch("R must be m×m"));
    }
    let mut p = q.clone();
    let mut last_p = p.clone();
    for it in 0..max_iters {
        let bt_p = b.transpose() * &p;
        let s = r + &bt_p * b;
        let s_inv = s
            .try_inverse()
            .ok_or(ControlError::DimensionMismatch("R + BᵀPB is singular"))?;
        let at_p = a.transpose() * &p;
        let at_p_a = &at_p * a;
        let at_p_b = &at_p * b;
        let next = &at_p_a - &at_p_b * s_inv * &bt_p * a + q;
        let residual = (&next - &p).norm();
        last_p = p;
        p = next;
        if residual < tol {
            let bt_p = b.transpose() * &p;
            let s = r + &bt_p * b;
            let s_inv = s
                .try_inverse()
                .ok_or(ControlError::DimensionMismatch("R + BᵀPB is singular"))?;
            let k = s_inv * bt_p * a;
            return Ok((p, k));
        }
        let _ = it;
    }
    Err(ControlError::DidNotConverge {
        iters: max_iters,
        residual: (&p - &last_p).norm(),
    })
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::{Complex, DMatrix};

    use super::*;

    #[test]
    fn lqr_double_integrator() {
        // Discrete double integrator at dt = 0.1 s.
        let dt = 0.1;
        let a = DMatrix::<f64>::from_row_slice(2, 2, &[1.0, dt, 0.0, 1.0]);
        let b = DMatrix::<f64>::from_row_slice(2, 1, &[0.5 * dt * dt, dt]);
        let q = DMatrix::<f64>::identity(2, 2);
        let r = DMatrix::<f64>::from_element(1, 1, 0.1);
        let (_p, k) = solve_dare(&a, &b, &q, &r, 1e-10, 10_000).unwrap();
        // Closed-loop must be stable: spectral radius of (A - B K) < 1.
        let cl = &a - &b * &k;
        let eig = cl.complex_eigenvalues();
        let max_abs = eig.iter().map(|z: &Complex<f64>| z.norm()).fold(0.0_f64, f64::max);
        assert!(max_abs < 1.0, "closed-loop spectral radius = {max_abs}");
    }

    #[test]
    fn rejects_dimension_mismatch() {
        let a = DMatrix::<f64>::identity(2, 2);
        let b = DMatrix::<f64>::zeros(3, 1);
        let q = DMatrix::<f64>::identity(2, 2);
        let r = DMatrix::<f64>::from_element(1, 1, 0.1);
        assert!(solve_dare(&a, &b, &q, &r, 1e-9, 10).is_err());
    }

    #[test]
    fn lqr_drives_state_to_zero() {
        let dt = 0.1;
        let a = DMatrix::<f64>::from_row_slice(2, 2, &[1.0, dt, 0.0, 1.0]);
        let b = DMatrix::<f64>::from_row_slice(2, 1, &[0.5 * dt * dt, dt]);
        let q = DMatrix::<f64>::identity(2, 2);
        let r = DMatrix::<f64>::from_element(1, 1, 0.05);
        let (_, k) = solve_dare(&a, &b, &q, &r, 1e-10, 10_000).unwrap();
        let mut x = nalgebra::DVector::from_vec(vec![1.0, 0.0]);
        for _ in 0..200 {
            let u = -&k * &x;
            x = &a * &x + &b * &u;
        }
        assert_abs_diff_eq!(x[0], 0.0, epsilon = 1e-2);
        assert_abs_diff_eq!(x[1], 0.0, epsilon = 1e-2);
    }
}
