//! Unconstrained finite-horizon Model Predictive Control.
//!
//! For LTI dynamics `x_{k+1} = A x_k + B u_k`, horizon `N`, and quadratic
//! cost
//!
//! ```text
//! J = Σ_{k=0}^{N-1} xₖᵀ Q xₖ + uₖᵀ R uₖ + x_Nᵀ Q_f x_N
//! ```
//!
//! the optimal feedback law is the time-varying state-feedback obtained
//! from a backward Riccati recursion:
//!
//! ```text
//! P_N    = Q_f
//! P_k    = Q + Aᵀ P_{k+1} A − Aᵀ P_{k+1} B (R + Bᵀ P_{k+1} B)⁻¹ Bᵀ P_{k+1} A
//! K_k    = (R + Bᵀ P_{k+1} B)⁻¹ Bᵀ P_{k+1} A
//! u_k    = −K_k x_k
//! ```
//!
//! `Mpc::compute_gains` runs the Riccati recursion once (typically called
//! when the model changes); `Mpc::action` evaluates the time-varying
//! gain at the current step.
//!
//! Constraints (input/state bounds) require a QP and are out of scope for
//! this lightweight implementation. For constrained MPC use a dedicated
//! solver; the `gain` schedule produced here is a useful warm-start.

use nalgebra::{DMatrix, DVector};

use crate::ControlError;

/// Per-step state-feedback gain schedule produced by [`compute_gains`].
#[derive(Debug, Clone)]
pub struct GainSchedule {
    /// `K_0, K_1, …, K_{N-1}` — `K_k` is `m × n`.
    pub gains: Vec<DMatrix<f64>>,
}

/// Run the backward Riccati recursion for the given model + horizon.
///
/// # Errors
/// [`ControlError::DimensionMismatch`] for shape disagreement.
pub fn compute_gains(
    a: &DMatrix<f64>,
    b: &DMatrix<f64>,
    q: &DMatrix<f64>,
    r: &DMatrix<f64>,
    qf: &DMatrix<f64>,
    horizon: usize,
) -> Result<GainSchedule, ControlError> {
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
    if qf.shape() != (n, n) {
        return Err(ControlError::DimensionMismatch("Qf must be n×n"));
    }
    if horizon == 0 {
        return Err(ControlError::DimensionMismatch("horizon must be positive"));
    }
    let mut p = qf.clone();
    let mut gains = Vec::with_capacity(horizon);
    for _ in 0..horizon {
        let bt_p = b.transpose() * &p;
        let s = r + &bt_p * b;
        let s_inv = s
            .try_inverse()
            .ok_or(ControlError::DimensionMismatch("R + BᵀPB singular"))?;
        let bt_p_a = &bt_p * a;
        let k = &s_inv * &bt_p_a;
        let at_p = a.transpose() * &p;
        let at_p_a = &at_p * a;
        let at_p_b = &at_p * b;
        let next_p = q + &at_p_a - &at_p_b * s_inv * &bt_p_a;
        gains.push(k);
        p = next_p;
    }
    gains.reverse();
    Ok(GainSchedule { gains })
}

/// Evaluate the time-varying feedback law at step `k` (0-indexed from the
/// current time).
///
/// # Errors
/// [`ControlError::OutOfRange`] if `k >= horizon`.
pub fn action(
    schedule: &GainSchedule,
    state: &DVector<f64>,
    k: usize,
) -> Result<DVector<f64>, ControlError> {
    if k >= schedule.gains.len() {
        return Err(ControlError::OutOfRange {
            name: "k",
            value: k as f64,
            range: "[0, horizon)",
        });
    }
    Ok(-&schedule.gains[k] * state)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::{DMatrix, DVector};

    use super::*;

    #[test]
    fn mpc_double_integrator_drives_state_to_zero() {
        let dt = 0.1;
        let a = DMatrix::<f64>::from_row_slice(2, 2, &[1.0, dt, 0.0, 1.0]);
        let b = DMatrix::<f64>::from_row_slice(2, 1, &[0.5 * dt * dt, dt]);
        let q = DMatrix::<f64>::identity(2, 2);
        let r = DMatrix::<f64>::from_element(1, 1, 0.1);
        let qf = DMatrix::<f64>::identity(2, 2) * 10.0;
        let horizon = 50;
        let sched = compute_gains(&a, &b, &q, &r, &qf, horizon).unwrap();

        let mut x = DVector::from_vec(vec![1.0, 0.0]);
        for k in 0..horizon {
            let u = action(&sched, &x, k).unwrap();
            x = &a * &x + &b * &u;
        }
        assert_abs_diff_eq!(x[0], 0.0, epsilon = 1e-2);
        assert_abs_diff_eq!(x[1], 0.0, epsilon = 1e-2);
    }

    #[test]
    fn rejects_zero_horizon() {
        let a = DMatrix::<f64>::identity(2, 2);
        let b = DMatrix::<f64>::zeros(2, 1);
        let q = DMatrix::<f64>::identity(2, 2);
        let r = DMatrix::<f64>::from_element(1, 1, 0.1);
        let qf = DMatrix::<f64>::identity(2, 2);
        assert!(compute_gains(&a, &b, &q, &r, &qf, 0).is_err());
    }
}
