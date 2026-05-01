//! Generic linearised Extended Kalman Filter.
//!
//! The classic discrete-time EKF executes two steps per cycle:
//!
//! **Predict** (given the current state estimate `x` and covariance `P`,
//! a [`MotionModel`] supplying the deterministic propagation `f(x)` and its
//! Jacobian `F = ∂f/∂x`, and the process-noise covariance `Q`):
//!
//! ```text
//! x⁻ = f(x)
//! P⁻ = F · P · Fᵀ + Q
//! ```
//!
//! **Update** (given a measurement `z`, a [`MeasurementModel`] supplying
//! `h(x)` and its Jacobian `H = ∂h/∂x`, and the measurement-noise
//! covariance `R`):
//!
//! ```text
//! y = z − h(x⁻)
//! S = H · P⁻ · Hᵀ + R
//! K = P⁻ · Hᵀ · S⁻¹
//! x⁺ = x⁻ + K · y
//! P⁺ = (I − K · H) · P⁻ · (I − K · H)ᵀ + K · R · Kᵀ
//! ```
//!
//! The Joseph form of the covariance update is used because it is
//! numerically stable even when the Kalman gain is non-optimal.

use nalgebra::{DMatrix, DVector};
use thiserror::Error;

/// Errors produced by [`Ekf`].
#[derive(Debug, Error, PartialEq)]
pub enum EkfError {
    /// State / covariance / model dimensions disagree.
    #[error("dimension mismatch: {0}")]
    DimensionMismatch(&'static str),
    /// Innovation covariance was numerically singular.
    #[error("innovation covariance is singular")]
    SingularInnovation,
}

/// Process / motion model used by [`Ekf::predict`].
pub trait MotionModel {
    /// Deterministic state propagation: `x_{k+1|k} = f(x_k, dt)`.
    fn propagate(&self, state: &DVector<f64>, dt: f64) -> DVector<f64>;
    /// Jacobian `F = ∂f/∂x` evaluated at `state`.
    fn jacobian(&self, state: &DVector<f64>, dt: f64) -> DMatrix<f64>;
}

/// Observation / measurement model used by [`Ekf::update`].
pub trait MeasurementModel {
    /// Predicted measurement: `z = h(x)`.
    fn predict(&self, state: &DVector<f64>) -> DVector<f64>;
    /// Jacobian `H = ∂h/∂x` evaluated at `state`.
    fn jacobian(&self, state: &DVector<f64>) -> DMatrix<f64>;
}

/// Discrete-time EKF with dynamically sized state.
#[derive(Debug, Clone)]
pub struct Ekf {
    /// State estimate `x ∈ ℝⁿ`.
    pub state: DVector<f64>,
    /// State error covariance `P ∈ ℝⁿˣⁿ`, symmetric positive-definite.
    pub covariance: DMatrix<f64>,
    /// Process noise covariance `Q ∈ ℝⁿˣⁿ` used by [`Ekf::predict`].
    pub process_noise: DMatrix<f64>,
}

impl Ekf {
    /// Construct an EKF from initial state, covariance, and process noise.
    /// Returns [`EkfError::DimensionMismatch`] if dimensions disagree.
    ///
    /// # Errors
    /// See above.
    pub fn new(
        state: DVector<f64>,
        covariance: DMatrix<f64>,
        process_noise: DMatrix<f64>,
    ) -> Result<Self, EkfError> {
        let n = state.len();
        if covariance.shape() != (n, n) {
            return Err(EkfError::DimensionMismatch("covariance shape != (n, n)"));
        }
        if process_noise.shape() != (n, n) {
            return Err(EkfError::DimensionMismatch("process_noise shape != (n, n)"));
        }
        Ok(Self { state, covariance, process_noise })
    }

    /// Predict step.
    ///
    /// # Errors
    /// Returns [`EkfError::DimensionMismatch`] if the model returns
    /// inconsistently sized vectors / matrices.
    pub fn predict<M: MotionModel + ?Sized>(
        &mut self,
        model: &M,
        dt: f64,
    ) -> Result<(), EkfError> {
        let n = self.state.len();
        let f_jac = model.jacobian(&self.state, dt);
        if f_jac.shape() != (n, n) {
            return Err(EkfError::DimensionMismatch("motion jacobian shape"));
        }
        let propagated = model.propagate(&self.state, dt);
        if propagated.len() != n {
            return Err(EkfError::DimensionMismatch("motion output length"));
        }
        self.state = propagated;
        self.covariance = &f_jac * &self.covariance * f_jac.transpose() + &self.process_noise;
        symmetrise(&mut self.covariance);
        Ok(())
    }

    /// Update step using measurement `z` with covariance `r`.
    ///
    /// # Errors
    /// [`EkfError::DimensionMismatch`] for shape disagreement;
    /// [`EkfError::SingularInnovation`] if `S = H P Hᵀ + R` cannot be
    /// inverted.
    pub fn update<M: MeasurementModel + ?Sized>(
        &mut self,
        model: &M,
        z: &DVector<f64>,
        r: &DMatrix<f64>,
    ) -> Result<(), EkfError> {
        let m = z.len();
        if r.shape() != (m, m) {
            return Err(EkfError::DimensionMismatch("R shape != (m, m)"));
        }
        let h = model.jacobian(&self.state);
        if h.shape() != (m, self.state.len()) {
            return Err(EkfError::DimensionMismatch("H shape != (m, n)"));
        }
        let predicted = model.predict(&self.state);
        if predicted.len() != m {
            return Err(EkfError::DimensionMismatch("h(x) length"));
        }
        let y = z - &predicted;
        let s = &h * &self.covariance * h.transpose() + r;
        let s_inv = s.try_inverse().ok_or(EkfError::SingularInnovation)?;
        let k = &self.covariance * h.transpose() * &s_inv;
        self.state = &self.state + &k * &y;
        // Joseph form covariance update.
        let identity = DMatrix::<f64>::identity(self.state.len(), self.state.len());
        let i_kh = &identity - &k * &h;
        self.covariance = &i_kh * &self.covariance * i_kh.transpose() + &k * r * k.transpose();
        symmetrise(&mut self.covariance);
        Ok(())
    }
}

fn symmetrise(p: &mut DMatrix<f64>) {
    let pt = p.transpose();
    *p += &pt;
    p.scale_mut(0.5);
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::{DMatrix, DVector};

    use super::*;

    /// Constant-velocity 1-D model: state = (position, velocity).
    /// f(x) = [pos + v·dt; v]    F = [[1, dt],[0, 1]]
    struct CvModel;

    impl MotionModel for CvModel {
        fn propagate(&self, state: &DVector<f64>, dt: f64) -> DVector<f64> {
            DVector::from_vec(vec![state[0] + state[1] * dt, state[1]])
        }
        fn jacobian(&self, _state: &DVector<f64>, dt: f64) -> DMatrix<f64> {
            DMatrix::<f64>::from_row_slice(2, 2, &[1.0, dt, 0.0, 1.0])
        }
    }

    /// Position-only observation: z = pos.
    struct PosModel;

    impl MeasurementModel for PosModel {
        fn predict(&self, state: &DVector<f64>) -> DVector<f64> {
            DVector::from_vec(vec![state[0]])
        }
        fn jacobian(&self, _state: &DVector<f64>) -> DMatrix<f64> {
            DMatrix::<f64>::from_row_slice(1, 2, &[1.0, 0.0])
        }
    }

    #[test]
    fn predict_then_update_drives_state_to_measurement() {
        let mut ekf = Ekf::new(
            DVector::from_vec(vec![0.0, 0.0]),
            DMatrix::<f64>::identity(2, 2) * 10.0,
            DMatrix::<f64>::identity(2, 2) * 0.01,
        )
        .unwrap();
        // Simulate ground-truth: position grows by 1 m per step, velocity = 1 m/s.
        for k in 1..=20 {
            ekf.predict(&CvModel, 1.0).unwrap();
            let z = DVector::from_vec(vec![f64::from(k)]);
            ekf.update(&PosModel, &z, &DMatrix::<f64>::from_element(1, 1, 0.04)).unwrap();
        }
        // After 20 perfect measurements the position should track tightly.
        assert_abs_diff_eq!(ekf.state[0], 20.0, epsilon = 0.5);
        assert_abs_diff_eq!(ekf.state[1], 1.0, epsilon = 0.5);
    }

    #[test]
    fn covariance_remains_symmetric() {
        let mut ekf = Ekf::new(
            DVector::from_vec(vec![0.0, 0.0]),
            DMatrix::<f64>::identity(2, 2),
            DMatrix::<f64>::identity(2, 2) * 0.05,
        )
        .unwrap();
        for _ in 0..50 {
            ekf.predict(&CvModel, 0.1).unwrap();
            ekf.update(
                &PosModel,
                &DVector::from_vec(vec![0.0]),
                &DMatrix::<f64>::from_element(1, 1, 0.1),
            )
            .unwrap();
        }
        let p = &ekf.covariance;
        for i in 0..p.nrows() {
            for j in 0..p.ncols() {
                assert_abs_diff_eq!(p[(i, j)], p[(j, i)], epsilon = 1e-9);
            }
        }
    }

    #[test]
    fn rejects_dimension_mismatch() {
        let s = DVector::from_vec(vec![0.0, 0.0]);
        let p = DMatrix::<f64>::identity(3, 3);
        let q = DMatrix::<f64>::identity(2, 2);
        let err = Ekf::new(s, p, q).unwrap_err();
        assert!(matches!(err, EkfError::DimensionMismatch(_)));
    }
}
