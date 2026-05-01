//! Sliding-Mode Controller.
//!
//! For a state error `e` and rate error `ė`, define the sliding surface
//! `s = ė + λ · e`. The continuous-time control law is
//!
//! ```text
//! u = u_eq − K · sat(s / Φ)
//! ```
//!
//! where `u_eq` is the equivalent control that would keep the system on
//! `s = 0`, `K` is a switching gain larger than the worst-case disturbance
//! upper bound, and `sat(·)` is a unit-saturation function with boundary
//! layer thickness `Φ` to soften the discontinuity (chattering reduction).

use crate::ControlError;

/// SMC tuning parameters.
#[derive(Debug, Clone, Copy)]
pub struct SmcGains {
    /// Sliding-surface slope `λ > 0`.
    pub lambda: f64,
    /// Switching gain `K > 0` — larger → more disturbance rejection,
    /// at the cost of more aggressive control.
    pub switching_gain: f64,
    /// Boundary-layer thickness `Φ > 0` to suppress chattering.
    pub boundary_layer: f64,
}

impl SmcGains {
    /// Validate the gain set.
    ///
    /// # Errors
    /// [`ControlError::OutOfRange`] if any gain is non-positive or non-finite.
    pub fn validate(&self) -> Result<(), ControlError> {
        for (name, v) in [
            ("lambda", self.lambda),
            ("switching_gain", self.switching_gain),
            ("boundary_layer", self.boundary_layer),
        ] {
            if !(v.is_finite() && v > 0.0) {
                return Err(ControlError::OutOfRange { name, value: v, range: "(0, +inf)" });
            }
        }
        Ok(())
    }
}

/// Compute the SMC control output.
///
/// `e` is the position error, `e_dot` the rate error, `u_eq` the equivalent
/// control. Returns the saturating SMC term.
///
/// # Errors
/// Propagates [`SmcGains::validate`] errors.
pub fn smc(e: f64, e_dot: f64, u_eq: f64, gains: SmcGains) -> Result<f64, ControlError> {
    gains.validate()?;
    let s = e_dot + gains.lambda * e;
    let switching = gains.switching_gain * sat(s / gains.boundary_layer);
    Ok(u_eq - switching)
}

/// Continuous saturation function: `sat(x) = clamp(x, -1, 1)`.
#[inline]
#[must_use]
pub fn sat(x: f64) -> f64 {
    x.clamp(-1.0, 1.0)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    #[test]
    fn smc_inside_boundary_layer_is_proportional() {
        // s/Φ = 0.5 → sat = 0.5 → u = u_eq − K·0.5.
        let g = SmcGains { lambda: 1.0, switching_gain: 2.0, boundary_layer: 1.0 };
        let u = smc(0.5, 0.0, 0.0, g).unwrap();
        assert_abs_diff_eq!(u, -1.0, epsilon = 1e-12);
    }

    #[test]
    fn smc_outside_boundary_layer_saturates() {
        let g = SmcGains { lambda: 1.0, switching_gain: 2.0, boundary_layer: 0.1 };
        let u = smc(10.0, 0.0, 0.0, g).unwrap();
        assert_abs_diff_eq!(u, -2.0, epsilon = 1e-12);
    }

    #[test]
    fn smc_drives_double_integrator_to_zero() {
        // y'' = u + d (d = sin disturbance amplitude 0.05).
        let g = SmcGains { lambda: 5.0, switching_gain: 1.0, boundary_layer: 0.05 };
        let mut y = 1.0_f64;
        let mut yd = 0.0_f64;
        let dt = 0.01;
        for k in 0..2000 {
            let d = 0.05 * (f64::from(k) * dt).sin();
            let u = smc(y, yd, 0.0, g).unwrap();
            yd += (u + d) * dt;
            y += yd * dt;
        }
        assert!(y.abs() < 0.05, "y not driven to zero: {y}");
    }

    #[test]
    fn rejects_zero_gain() {
        let g = SmcGains { lambda: 0.0, switching_gain: 1.0, boundary_layer: 0.1 };
        assert!(smc(0.0, 0.0, 0.0, g).is_err());
    }
}
