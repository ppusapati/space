//! Unscented Kalman Filter for nonlinear state estimation.
//!
//! Implements the symmetric scaled-sigma-point UKF (Julier & Uhlmann
//! 2004; Wan & van der Merwe 2000) with parameters `(α, β, κ)`. For an
//! `n`-dimensional state, `2n + 1` sigma points are generated using the
//! Cholesky factor of the scaled covariance, then propagated through the
//! user-supplied nonlinear motion / measurement functions. Mean and
//! covariance are recovered from weighted sums.
//!
//! Algorithm:
//!
//! ```text
//! λ      = α² (n + κ) − n
//! W₀ᵐ    = λ / (n + λ)
//! W₀ᶜ    = λ / (n + λ) + (1 − α² + β)
//! Wᵢᵐ_ᶜ = 1 / (2(n + λ))                 for i = 1 … 2n
//! χ₀     = x̂
//! χᵢ     = x̂ + (√((n + λ) P))ᵢ            for i = 1 … n
//! χᵢ₊ₙ   = x̂ − (√((n + λ) P))ᵢ            for i = 1 … n
//! ```
//!
//! Predict:
//!
//! ```text
//! Yᵢ   = f(χᵢ)
//! x̂⁻   = Σ Wᵢᵐ Yᵢ
//! P⁻   = Σ Wᵢᶜ (Yᵢ − x̂⁻)(Yᵢ − x̂⁻)ᵀ + Q
//! ```
//!
//! Update:
//!
//! ```text
//! Zᵢ   = h(Yᵢ)
//! ẑ    = Σ Wᵢᵐ Zᵢ
//! Pzz  = Σ Wᵢᶜ (Zᵢ − ẑ)(Zᵢ − ẑ)ᵀ + R
//! Pxz  = Σ Wᵢᶜ (Yᵢ − x̂⁻)(Zᵢ − ẑ)ᵀ
//! K    = Pxz · Pzz⁻¹
//! x̂⁺   = x̂⁻ + K (z − ẑ)
//! P⁺   = P⁻ − K Pzz Kᵀ
//! ```

#![cfg_attr(docsrs, feature(doc_cfg))]

use nalgebra::{DMatrix, DVector};
use thiserror::Error;

/// Errors produced by [`Ukf`].
#[derive(Debug, Error, PartialEq)]
pub enum UkfError {
    /// State / covariance / model dimensions disagree.
    #[error("dimension mismatch: {0}")]
    DimensionMismatch(&'static str),
    /// Cholesky decomposition of the scaled covariance failed (matrix is
    /// not positive-definite).
    #[error("Cholesky decomposition failed: covariance not positive-definite")]
    NotPositiveDefinite,
    /// Innovation covariance was numerically singular.
    #[error("innovation covariance is singular")]
    SingularInnovation,
}

/// Process / motion model used by [`Ukf::predict`].
pub trait MotionFn: Fn(&DVector<f64>, f64) -> DVector<f64> {}
impl<F> MotionFn for F where F: Fn(&DVector<f64>, f64) -> DVector<f64> {}

/// Measurement function used by [`Ukf::update`].
pub trait MeasurementFn: Fn(&DVector<f64>) -> DVector<f64> {}
impl<F> MeasurementFn for F where F: Fn(&DVector<f64>) -> DVector<f64> {}

/// UKF tuning parameters.
#[derive(Debug, Clone, Copy)]
pub struct UkfParams {
    /// Sigma-point spread `α ∈ (0, 1]`. Smaller values keep sigma points
    /// closer to the mean.
    pub alpha: f64,
    /// Distribution prior knowledge `β`. `2.0` is optimal for Gaussian.
    pub beta: f64,
    /// Secondary scaling `κ`. `3 − n` is a common choice but `0` is safer
    /// when `n > 3`.
    pub kappa: f64,
}

impl Default for UkfParams {
    fn default() -> Self {
        Self { alpha: 1e-3, beta: 2.0, kappa: 0.0 }
    }
}

/// Discrete-time UKF.
#[derive(Debug, Clone)]
pub struct Ukf {
    /// State estimate.
    pub state: DVector<f64>,
    /// State error covariance.
    pub covariance: DMatrix<f64>,
    /// Process noise covariance.
    pub process_noise: DMatrix<f64>,
    /// Tuning parameters.
    pub params: UkfParams,
}

impl Ukf {
    /// Construct a UKF.
    ///
    /// # Errors
    /// [`UkfError::DimensionMismatch`] if shapes disagree.
    pub fn new(
        state: DVector<f64>,
        covariance: DMatrix<f64>,
        process_noise: DMatrix<f64>,
        params: UkfParams,
    ) -> Result<Self, UkfError> {
        let n = state.len();
        if covariance.shape() != (n, n) {
            return Err(UkfError::DimensionMismatch("covariance shape != (n, n)"));
        }
        if process_noise.shape() != (n, n) {
            return Err(UkfError::DimensionMismatch("process_noise shape != (n, n)"));
        }
        Ok(Self { state, covariance, process_noise, params })
    }

    fn weights(&self) -> (f64, Vec<f64>, Vec<f64>) {
        let n = self.state.len() as f64;
        let lambda = self.params.alpha * self.params.alpha * (n + self.params.kappa) - n;
        let denom = n + lambda;
        let mut wm = vec![0.0_f64; 2 * self.state.len() + 1];
        let mut wc = vec![0.0_f64; 2 * self.state.len() + 1];
        wm[0] = lambda / denom;
        wc[0] = lambda / denom + (1.0 - self.params.alpha * self.params.alpha + self.params.beta);
        for i in 1..wm.len() {
            wm[i] = 0.5 / denom;
            wc[i] = 0.5 / denom;
        }
        (lambda, wm, wc)
    }

    fn sigma_points(&self) -> Result<Vec<DVector<f64>>, UkfError> {
        let n = self.state.len();
        let n_f = n as f64;
        let lambda = self.params.alpha * self.params.alpha * (n_f + self.params.kappa) - n_f;
        let scaled = (n_f + lambda) * &self.covariance;
        let chol = scaled.cholesky().ok_or(UkfError::NotPositiveDefinite)?;
        let l = chol.l();
        let mut points = Vec::with_capacity(2 * n + 1);
        points.push(self.state.clone());
        for i in 0..n {
            let col = l.column(i).into_owned();
            points.push(&self.state + &col);
        }
        for i in 0..n {
            let col = l.column(i).into_owned();
            points.push(&self.state - &col);
        }
        Ok(points)
    }

    /// Predict step.
    ///
    /// # Errors
    /// [`UkfError`] for shape disagreement or non-PD covariance.
    pub fn predict<F: MotionFn>(&mut self, f: F, dt: f64) -> Result<(), UkfError> {
        let sigmas = self.sigma_points()?;
        let propagated: Vec<_> = sigmas.iter().map(|x| f(x, dt)).collect();
        let n = propagated[0].len();
        let (_, wm, wc) = self.weights();
        // Mean.
        let mut mean = DVector::<f64>::zeros(n);
        for (i, y) in propagated.iter().enumerate() {
            mean += y * wm[i];
        }
        // Covariance.
        let mut cov = DMatrix::<f64>::zeros(n, n);
        for (i, y) in propagated.iter().enumerate() {
            let diff = y - &mean;
            cov += &diff * diff.transpose() * wc[i];
        }
        cov += &self.process_noise;
        symmetrise(&mut cov);
        // Resize state slot to match output.
        if self.state.len() != n {
            self.state = DVector::<f64>::zeros(n);
            self.covariance = DMatrix::<f64>::zeros(n, n);
        }
        self.state = mean;
        self.covariance = cov;
        Ok(())
    }

    /// Update step.
    ///
    /// # Errors
    /// [`UkfError`] for shape disagreement, non-PD covariance, or singular
    /// innovation covariance.
    pub fn update<H: MeasurementFn>(
        &mut self,
        h: H,
        z: &DVector<f64>,
        r: &DMatrix<f64>,
    ) -> Result<(), UkfError> {
        let m = z.len();
        if r.shape() != (m, m) {
            return Err(UkfError::DimensionMismatch("R shape != (m, m)"));
        }
        let sigmas = self.sigma_points()?;
        let predicted: Vec<_> = sigmas.iter().map(&h).collect();
        if predicted[0].len() != m {
            return Err(UkfError::DimensionMismatch("h(x) length != z length"));
        }
        let (_, wm, wc) = self.weights();
        // Predicted measurement mean.
        let mut z_hat = DVector::<f64>::zeros(m);
        for (i, zi) in predicted.iter().enumerate() {
            z_hat += zi * wm[i];
        }
        // Innovation covariance Pzz and cross-covariance Pxz.
        let mut pzz = DMatrix::<f64>::zeros(m, m);
        let mut pxz = DMatrix::<f64>::zeros(self.state.len(), m);
        for (i, zi) in predicted.iter().enumerate() {
            let dz = zi - &z_hat;
            let dx = &sigmas[i] - &self.state;
            pzz += &dz * dz.transpose() * wc[i];
            pxz += &dx * dz.transpose() * wc[i];
        }
        pzz += r;
        let pzz_inv = pzz.clone().try_inverse().ok_or(UkfError::SingularInnovation)?;
        let k = &pxz * &pzz_inv;
        let y = z - &z_hat;
        self.state = &self.state + &k * &y;
        self.covariance = &self.covariance - &k * &pzz * k.transpose();
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

    #[test]
    fn ukf_linear_constant_velocity_matches_truth() {
        // 1-D constant-velocity model: x = (pos, vel); same as the EKF test.
        let mut ukf = Ukf::new(
            DVector::from_vec(vec![0.0, 0.0]),
            DMatrix::<f64>::identity(2, 2) * 10.0,
            DMatrix::<f64>::identity(2, 2) * 0.01,
            UkfParams::default(),
        )
        .unwrap();
        let f = |x: &DVector<f64>, dt: f64| DVector::from_vec(vec![x[0] + x[1] * dt, x[1]]);
        let h = |x: &DVector<f64>| DVector::from_vec(vec![x[0]]);
        for k in 1..=20 {
            ukf.predict(f, 1.0).unwrap();
            ukf.update(
                h,
                &DVector::from_vec(vec![f64::from(k)]),
                &DMatrix::<f64>::from_element(1, 1, 0.04),
            )
            .unwrap();
        }
        assert_abs_diff_eq!(ukf.state[0], 20.0, epsilon = 0.5);
        assert_abs_diff_eq!(ukf.state[1], 1.0, epsilon = 0.5);
    }

    #[test]
    fn ukf_recovers_from_strongly_nonlinear_observation() {
        // Truth: pos = 5, vel = 0. Measurement = pos² (heavily nonlinear).
        let mut ukf = Ukf::new(
            DVector::from_vec(vec![3.0, 0.0]),
            DMatrix::<f64>::identity(2, 2) * 4.0,
            DMatrix::<f64>::identity(2, 2) * 0.01,
            UkfParams::default(),
        )
        .unwrap();
        let f = |x: &DVector<f64>, dt: f64| DVector::from_vec(vec![x[0] + x[1] * dt, x[1]]);
        let h = |x: &DVector<f64>| DVector::from_vec(vec![x[0] * x[0]]);
        for _ in 0..40 {
            ukf.predict(f, 0.1).unwrap();
            ukf.update(h, &DVector::from_vec(vec![25.0]), &DMatrix::<f64>::from_element(1, 1, 0.04))
                .unwrap();
        }
        assert_abs_diff_eq!(ukf.state[0], 5.0, epsilon = 0.5);
    }

    #[test]
    fn rejects_dimension_mismatch() {
        let s = DVector::from_vec(vec![0.0, 0.0]);
        let p = DMatrix::<f64>::identity(3, 3);
        let q = DMatrix::<f64>::identity(2, 2);
        assert!(matches!(
            Ukf::new(s, p, q, UkfParams::default()).unwrap_err(),
            UkfError::DimensionMismatch(_)
        ));
    }

    #[test]
    fn rejects_non_pd_covariance() {
        let mut ukf = Ukf::new(
            DVector::from_vec(vec![0.0, 0.0]),
            DMatrix::<f64>::zeros(2, 2),
            DMatrix::<f64>::identity(2, 2),
            UkfParams::default(),
        )
        .unwrap();
        let f = |x: &DVector<f64>, _: f64| x.clone();
        assert_eq!(ukf.predict(f, 1.0).unwrap_err(), UkfError::NotPositiveDefinite);
    }
}
