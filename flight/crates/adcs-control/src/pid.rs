//! PID controller.
//!
//! Discrete form:
//!
//! ```text
//! e_k     = setpoint − measurement
//! P_k     = Kp · e_k
//! I_k     = I_{k-1} + Ki · e_k · dt              (subject to anti-windup)
//! D_k     = Kd · (e_k − e_{k-1}) / dt            (filtered if Tf > 0)
//! u_raw   = P_k + I_k + D_k
//! u_k     = clamp(u_raw, u_min, u_max)
//! ```
//!
//! Anti-windup uses the back-calculation method: when `u_raw` is clamped,
//! the integrator is rolled back so that `u_raw` would have ended up at the
//! saturation bound.

use crate::ControlError;

/// PID gains.
#[derive(Debug, Clone, Copy)]
pub struct PidGains {
    /// Proportional gain.
    pub kp: f64,
    /// Integral gain.
    pub ki: f64,
    /// Derivative gain.
    pub kd: f64,
    /// Derivative low-pass filter time-constant (seconds). Zero disables filtering.
    pub tf: f64,
}

/// Output saturation limits for [`Pid`].
#[derive(Debug, Clone, Copy)]
pub struct PidLimits {
    /// Lower bound on the control output.
    pub min: f64,
    /// Upper bound on the control output.
    pub max: f64,
}

impl Default for PidLimits {
    fn default() -> Self {
        Self { min: f64::NEG_INFINITY, max: f64::INFINITY }
    }
}

/// PID controller state.
#[derive(Debug, Clone)]
pub struct Pid {
    /// Gain set.
    pub gains: PidGains,
    /// Saturation limits.
    pub limits: PidLimits,
    /// Integral accumulator.
    pub integrator: f64,
    /// Filtered derivative state.
    pub derivative_state: f64,
    /// Last error (for unfiltered derivative).
    pub last_error: f64,
}

impl Pid {
    /// Construct a new PID with zero integrator/derivative state.
    ///
    /// # Errors
    /// [`ControlError::OutOfRange`] if `tf < 0` or `min > max`.
    pub fn new(gains: PidGains, limits: PidLimits) -> Result<Self, ControlError> {
        if gains.tf < 0.0 || !gains.tf.is_finite() {
            return Err(ControlError::OutOfRange {
                name: "tf",
                value: gains.tf,
                range: "[0, +inf)",
            });
        }
        if limits.min > limits.max {
            return Err(ControlError::OutOfRange {
                name: "limits.min/max",
                value: limits.min,
                range: "min <= max",
            });
        }
        Ok(Self { gains, limits, integrator: 0.0, derivative_state: 0.0, last_error: 0.0 })
    }

    /// Reset all internal state to zero.
    pub fn reset(&mut self) {
        self.integrator = 0.0;
        self.derivative_state = 0.0;
        self.last_error = 0.0;
    }

    /// Compute the next control output.
    ///
    /// `dt` must be positive.
    ///
    /// # Errors
    /// [`ControlError::OutOfRange`] if `dt <= 0`.
    pub fn step(&mut self, setpoint: f64, measurement: f64, dt: f64) -> Result<f64, ControlError> {
        if !(dt > 0.0 && dt.is_finite()) {
            return Err(ControlError::OutOfRange {
                name: "dt",
                value: dt,
                range: "(0, +inf)",
            });
        }
        let error = setpoint - measurement;
        let p = self.gains.kp * error;
        // Tentative integrator update.
        let mut i_new = self.integrator + self.gains.ki * error * dt;
        // Derivative term (with optional first-order low-pass filter).
        let raw_deriv = (error - self.last_error) / dt;
        let d = if self.gains.tf > 0.0 {
            let alpha = dt / (self.gains.tf + dt);
            self.derivative_state += alpha * (raw_deriv - self.derivative_state);
            self.gains.kd * self.derivative_state
        } else {
            self.gains.kd * raw_deriv
        };
        let u_raw = p + i_new + d;
        let u_sat = u_raw.clamp(self.limits.min, self.limits.max);
        // Back-calculation anti-windup. Equality is exact because `clamp`
        // returns the input value byte-for-byte when it is inside the range.
        #[allow(clippy::float_cmp)]
        let saturated = u_raw != u_sat;
        if saturated && self.gains.ki != 0.0 {
            i_new -= u_raw - u_sat;
        }
        self.integrator = i_new;
        self.last_error = error;
        Ok(u_sat)
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    #[test]
    fn proportional_only_steady_offset() {
        let mut pid = Pid::new(
            PidGains { kp: 2.0, ki: 0.0, kd: 0.0, tf: 0.0 },
            PidLimits::default(),
        )
        .unwrap();
        let u = pid.step(10.0, 7.0, 0.1).unwrap();
        // P = 2 * 3 = 6
        assert_abs_diff_eq!(u, 6.0, epsilon = 1e-12);
    }

    #[test]
    fn integrator_drives_steady_state_to_setpoint() {
        // First-order plant: y_{k+1} = a·y_k + b·u_k with a = 0.9, b = 0.1.
        // PI control (Kp=1, Ki=1) gives an asymptotically stable closed-loop
        // that drives the steady-state error to zero.
        let mut pid = Pid::new(
            PidGains { kp: 1.0, ki: 1.0, kd: 0.0, tf: 0.0 },
            PidLimits::default(),
        )
        .unwrap();
        let mut y = 0.0_f64;
        let setpoint = 5.0;
        for _ in 0..2000 {
            let u = pid.step(setpoint, y, 0.1).unwrap();
            y = 0.9 * y + 0.1 * u;
        }
        assert_abs_diff_eq!(y, setpoint, epsilon = 0.05);
    }

    #[test]
    fn anti_windup_keeps_integrator_finite() {
        let mut pid = Pid::new(
            PidGains { kp: 0.0, ki: 100.0, kd: 0.0, tf: 0.0 },
            PidLimits { min: -1.0, max: 1.0 },
        )
        .unwrap();
        for _ in 0..1000 {
            let _ = pid.step(10.0, 0.0, 0.01).unwrap();
        }
        // Without anti-windup the integrator would explode to ~1e6.
        assert!(pid.integrator.abs() < 5.0, "integrator should not run away: {}", pid.integrator);
    }

    #[test]
    fn saturation_respected() {
        let mut pid = Pid::new(
            PidGains { kp: 1000.0, ki: 0.0, kd: 0.0, tf: 0.0 },
            PidLimits { min: -1.0, max: 1.0 },
        )
        .unwrap();
        let u = pid.step(10.0, 0.0, 0.01).unwrap();
        assert_abs_diff_eq!(u, 1.0, epsilon = 1e-12);
    }

    #[test]
    fn rejects_negative_tf() {
        let err = Pid::new(
            PidGains { kp: 1.0, ki: 0.0, kd: 0.0, tf: -1.0 },
            PidLimits::default(),
        )
        .unwrap_err();
        assert!(matches!(err, ControlError::OutOfRange { name: "tf", .. }));
    }
}
