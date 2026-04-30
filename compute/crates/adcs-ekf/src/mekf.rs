//! Multiplicative Extended Kalman Filter (MEKF) for spacecraft attitude.
//!
//! State vector: `(δθ, b_g) ∈ ℝ⁶` — three-axis attitude error angles plus
//! gyroscope bias. The full attitude is carried separately as a unit
//! quaternion that is updated multiplicatively after every measurement
//! reset, ensuring the quaternion always remains on the unit-sphere
//! manifold (Markley & Crassidis 2014, Eq. 6.65).
//!
//! Time propagation uses gyroscope rate measurements with bias correction:
//!
//! ```text
//! ω̂ = ω_meas − b_g
//! q̇ = ½ Ω(ω̂) q
//! ```
//!
//! The error-state propagation matrix (continuous-time):
//!
//! ```text
//! F = [ −[ω̂×]   −I_3  ]
//!     [   0_3     0_3  ]
//! ```
//!
//! discretised by zero-order hold over the gyro period.
//!
//! Vector-observation update: a unit reference vector `r̂_b` (e.g. sun
//! sensor or magnetometer) compared to the inertial reference `r̂_n`
//! rotated by the current attitude estimate. The measurement Jacobian for
//! the error state is `H = [[r̂_b ×]  0_3]`.

use nalgebra::{Matrix3, Matrix6, Quaternion, UnitQuaternion, Vector3};
use thiserror::Error;

/// Errors produced by [`MultiplicativeEkf`].
#[derive(Debug, Error, PartialEq)]
pub enum MekfError {
    /// Innovation covariance was singular.
    #[error("innovation covariance is singular")]
    SingularInnovation,
    /// Measurement vector was zero-length, so no direction information was
    /// available.
    #[error("zero-norm measurement or reference vector")]
    DegenerateMeasurement,
}

/// MEKF state.
#[derive(Debug, Clone)]
pub struct MekfState {
    /// Current best estimate of the body-to-inertial attitude quaternion.
    pub attitude: UnitQuaternion<f64>,
    /// Gyro bias estimate `b_g ∈ ℝ³` (rad/s).
    pub gyro_bias: Vector3<f64>,
    /// 6×6 error covariance `(δθ, b_g)`.
    pub covariance: Matrix6<f64>,
}

impl MekfState {
    /// Construct an initial MEKF state.
    #[must_use]
    pub fn new(attitude: UnitQuaternion<f64>, gyro_bias: Vector3<f64>, covariance: Matrix6<f64>) -> Self {
        Self { attitude, gyro_bias, covariance }
    }
}

/// Vector observation: a unit reference direction expressed in the inertial
/// frame and the corresponding measurement in the body frame.
#[derive(Debug, Clone, Copy)]
pub struct VectorObservation {
    /// Inertial-frame direction (unit vector).
    pub inertial: Vector3<f64>,
    /// Body-frame measurement (unit vector).
    pub body: Vector3<f64>,
    /// Per-axis measurement noise covariance (3×3).
    pub noise: Matrix3<f64>,
}

/// MEKF for ADCS.
#[derive(Debug, Clone)]
pub struct MultiplicativeEkf {
    /// Current state.
    pub state: MekfState,
    /// Gyro angular-random-walk noise covariance (rad/s)² per √Hz, applied
    /// to the upper 3×3 of the process-noise matrix.
    pub gyro_arw: f64,
    /// Gyro bias-random-walk noise (rad/s²)² per √Hz, applied to the lower
    /// 3×3 of the process-noise matrix.
    pub gyro_brw: f64,
}

impl MultiplicativeEkf {
    /// Construct a new MEKF.
    #[must_use]
    pub fn new(state: MekfState, gyro_arw: f64, gyro_brw: f64) -> Self {
        Self { state, gyro_arw, gyro_brw }
    }

    /// Time-propagate the MEKF using a gyroscope rate measurement `ω_meas`
    /// (rad/s) over period `dt` (s). Implements the canonical first-order
    /// discrete propagation.
    pub fn propagate(&mut self, omega_meas: Vector3<f64>, dt: f64) {
        let omega_hat = omega_meas - self.state.gyro_bias;
        // Quaternion propagation: q_{k+1} = q_k ⊗ exp(½ ω̂ dt)
        let half_angle = 0.5 * dt * omega_hat.norm();
        let dq = if half_angle > 0.0 {
            let axis = omega_hat.normalize();
            let s = half_angle.sin();
            UnitQuaternion::from_quaternion(Quaternion::new(
                half_angle.cos(),
                axis.x * s,
                axis.y * s,
                axis.z * s,
            ))
        } else {
            UnitQuaternion::identity()
        };
        self.state.attitude *= dq;

        // Error-state covariance propagation (zero-order hold).
        let mut f = Matrix6::<f64>::identity();
        let omega_skew = skew(omega_hat);
        // Top-left: I − [ω̂ ×] dt
        let top_left = Matrix3::<f64>::identity() - omega_skew * dt;
        let top_right = -Matrix3::<f64>::identity() * dt;
        f.fixed_view_mut::<3, 3>(0, 0).copy_from(&top_left);
        f.fixed_view_mut::<3, 3>(0, 3).copy_from(&top_right);
        // Bottom-left: 0; bottom-right: I (already from identity).
        let mut q = Matrix6::<f64>::zeros();
        q.fixed_view_mut::<3, 3>(0, 0).copy_from(&(Matrix3::<f64>::identity() * (self.gyro_arw * dt)));
        q.fixed_view_mut::<3, 3>(3, 3).copy_from(&(Matrix3::<f64>::identity() * (self.gyro_brw * dt)));
        self.state.covariance = f * self.state.covariance * f.transpose() + q;
        symmetrise(&mut self.state.covariance);
    }

    /// Update step using a single vector observation.
    ///
    /// # Errors
    /// [`MekfError::SingularInnovation`] if the 3×3 innovation covariance
    /// cannot be inverted; [`MekfError::DegenerateMeasurement`] if either
    /// vector has zero norm.
    pub fn update(&mut self, obs: &VectorObservation) -> Result<(), MekfError> {
        if obs.body.norm() == 0.0 || obs.inertial.norm() == 0.0 {
            return Err(MekfError::DegenerateMeasurement);
        }
        let r_b = obs.body.normalize();
        let r_n = obs.inertial.normalize();

        // Predicted measurement: rotate r_n into body frame using current attitude.
        let q = self.state.attitude;
        let predicted_b = q.inverse() * r_n;

        // Innovation y = r_b − predicted.
        let y = r_b - predicted_b;

        // Measurement Jacobian H = [[r_b ×]  0]
        let mut h = nalgebra::Matrix3x6::<f64>::zeros();
        h.fixed_view_mut::<3, 3>(0, 0).copy_from(&skew(r_b));

        let p = &self.state.covariance;
        let s = h * p * h.transpose() + obs.noise;
        let s_inv = s.try_inverse().ok_or(MekfError::SingularInnovation)?;
        let k = p * h.transpose() * s_inv;
        let dx = k * y;

        // Apply δθ via multiplicative update; bias additively.
        let dtheta = Vector3::new(dx[0], dx[1], dx[2]);
        let dbias = Vector3::new(dx[3], dx[4], dx[5]);
        let dq = small_angle_quaternion(dtheta);
        self.state.attitude *= dq;
        self.state.gyro_bias += dbias;

        // Joseph form covariance update.
        let identity = Matrix6::<f64>::identity();
        let i_kh = identity - k * h;
        self.state.covariance = i_kh * p * i_kh.transpose() + k * obs.noise * k.transpose();
        symmetrise(&mut self.state.covariance);
        Ok(())
    }
}

#[inline]
fn skew(v: Vector3<f64>) -> Matrix3<f64> {
    Matrix3::new(0.0, -v.z, v.y, v.z, 0.0, -v.x, -v.y, v.x, 0.0)
}

fn small_angle_quaternion(dtheta: Vector3<f64>) -> UnitQuaternion<f64> {
    // q ≈ (1, 0.5 δθ); renormalise.
    let q = Quaternion::new(1.0, 0.5 * dtheta.x, 0.5 * dtheta.y, 0.5 * dtheta.z);
    UnitQuaternion::from_quaternion(q)
}

fn symmetrise(p: &mut Matrix6<f64>) {
    let pt = p.transpose();
    *p += pt;
    *p *= 0.5;
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::{UnitQuaternion, Vector3};

    use super::*;

    #[test]
    fn propagate_zero_rate_keeps_attitude() {
        let init = MekfState::new(
            UnitQuaternion::identity(),
            Vector3::zeros(),
            Matrix6::identity() * 1e-3,
        );
        let mut mekf = MultiplicativeEkf::new(init, 1e-6, 1e-9);
        mekf.propagate(Vector3::zeros(), 0.1);
        let q = mekf.state.attitude;
        assert_abs_diff_eq!(q.w, 1.0, epsilon = 1e-12);
        assert_abs_diff_eq!(q.i, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(q.j, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(q.k, 0.0, epsilon = 1e-12);
    }

    #[test]
    fn propagate_x_rate_rotates_about_x() {
        let init = MekfState::new(
            UnitQuaternion::identity(),
            Vector3::zeros(),
            Matrix6::identity() * 1e-3,
        );
        let mut mekf = MultiplicativeEkf::new(init, 1e-6, 1e-9);
        // 1 rad/s about x for 1 s -> rotation of 1 rad around x.
        mekf.propagate(Vector3::new(1.0, 0.0, 0.0), 1.0);
        let aa = mekf.state.attitude.scaled_axis();
        assert_abs_diff_eq!(aa.x, 1.0, epsilon = 1e-9);
        assert_abs_diff_eq!(aa.y, 0.0, epsilon = 1e-9);
        assert_abs_diff_eq!(aa.z, 0.0, epsilon = 1e-9);
    }

    #[test]
    fn vector_update_corrects_misaligned_attitude() {
        // Truth: identity. Initial estimate: 5° off about z.
        let theta0 = 5.0_f64.to_radians();
        let init_q = UnitQuaternion::from_axis_angle(&Vector3::z_axis(), theta0);
        let init = MekfState::new(init_q, Vector3::zeros(), Matrix6::identity() * 0.01);
        let mut mekf = MultiplicativeEkf::new(init, 1e-6, 1e-9);

        // Reference vector: x_inertial; body measurement: x_body (truth = identity).
        for _ in 0..30 {
            let obs = VectorObservation {
                inertial: Vector3::new(1.0, 0.0, 0.0),
                body: Vector3::new(1.0, 0.0, 0.0),
                noise: Matrix3::identity() * 1e-4,
            };
            mekf.update(&obs).unwrap();
            let obs2 = VectorObservation {
                inertial: Vector3::new(0.0, 1.0, 0.0),
                body: Vector3::new(0.0, 1.0, 0.0),
                noise: Matrix3::identity() * 1e-4,
            };
            mekf.update(&obs2).unwrap();
        }
        let final_angle = mekf.state.attitude.angle();
        assert!(final_angle < 0.5_f64.to_radians(), "angle = {final_angle:.6} rad");
    }

    #[test]
    fn rejects_zero_norm_measurement() {
        let init = MekfState::new(UnitQuaternion::identity(), Vector3::zeros(), Matrix6::identity());
        let mut mekf = MultiplicativeEkf::new(init, 1e-6, 1e-9);
        let obs = VectorObservation {
            inertial: Vector3::zeros(),
            body: Vector3::new(1.0, 0.0, 0.0),
            noise: Matrix3::identity(),
        };
        assert_eq!(mekf.update(&obs).unwrap_err(), MekfError::DegenerateMeasurement);
    }
}
