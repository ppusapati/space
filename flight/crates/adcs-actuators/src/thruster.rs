//! Cold-gas / monopropellant thruster model.
//!
//! Each thruster has a body-frame thrust vector direction `d̂` (unit), a
//! mounting position `r` (body coordinates), a peak thrust `F`, a
//! specific impulse `Isp`, and a minimum-impulse-bit duration `t_min`.
//! When fired for a duration `t`, the thruster produces:
//!
//! * Force on the spacecraft: `F · d̂`
//! * Torque on the spacecraft about the body origin: `r × (F · d̂)`
//! * Propellant mass consumed: `F · t / (Isp · g₀)`

use nalgebra::Vector3;

use crate::ActuatorError;

/// Standard gravity (m/s²) used to convert Isp (seconds) to exit velocity.
pub const G0: f64 = 9.806_65;

/// Single thruster.
#[derive(Debug, Clone, Copy)]
pub struct Thruster {
    /// Direction of thrust expressed in the body frame (unit vector).
    pub direction: Vector3<f64>,
    /// Mounting position in the body frame relative to the centre of mass (m).
    pub position: Vector3<f64>,
    /// Peak thrust (N).
    pub thrust: f64,
    /// Specific impulse (s).
    pub isp: f64,
    /// Minimum impulse bit duration (s) — pulses shorter than this are
    /// rejected.
    pub min_pulse: f64,
}

impl Thruster {
    /// Validate the thruster.
    ///
    /// # Errors
    /// [`ActuatorError::OutOfRange`] for invalid parameter values.
    pub fn validate(&self) -> Result<(), ActuatorError> {
        for (name, v) in [
            ("thrust", self.thrust),
            ("isp", self.isp),
            ("min_pulse", self.min_pulse),
        ] {
            if !(v.is_finite() && v > 0.0) {
                return Err(ActuatorError::OutOfRange { name, value: v, range: "(0, +inf)" });
            }
        }
        let n = self.direction.norm();
        if (n - 1.0).abs() >= 1e-6 {
            return Err(ActuatorError::OutOfRange {
                name: "direction",
                value: n,
                range: "unit vector",
            });
        }
        Ok(())
    }

    /// Body-frame force produced by firing this thruster.
    #[must_use]
    pub fn force(&self) -> Vector3<f64> {
        self.thrust * self.direction
    }

    /// Body-frame torque produced about the centre of mass.
    #[must_use]
    pub fn torque(&self) -> Vector3<f64> {
        self.position.cross(&self.force())
    }

    /// Propellant mass consumed for a `duration` (s) firing.
    /// Returns 0 if `duration` is below the minimum impulse bit.
    #[must_use]
    pub fn propellant(&self, duration: f64) -> f64 {
        if duration <= 0.0 || duration < self.min_pulse {
            return 0.0;
        }
        self.thrust * duration / (self.isp * G0)
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    fn nominal() -> Thruster {
        Thruster {
            direction: Vector3::new(0.0, 0.0, 1.0),
            position: Vector3::new(0.5, 0.0, 0.0),
            thrust: 1.0,
            isp: 220.0,
            min_pulse: 0.01,
        }
    }

    #[test]
    fn force_and_torque_basic() {
        let t = nominal();
        t.validate().unwrap();
        let f = t.force();
        assert_abs_diff_eq!(f.z, 1.0, epsilon = 1e-12);
        let tau = t.torque();
        // r × F = (0.5, 0, 0) × (0, 0, 1) = (0, -0.5, 0)
        assert_abs_diff_eq!(tau.x, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(tau.y, -0.5, epsilon = 1e-12);
        assert_abs_diff_eq!(tau.z, 0.0, epsilon = 1e-12);
    }

    #[test]
    fn propellant_rocket_equation() {
        let t = nominal();
        // 1 second @ 1 N, Isp 220 s -> mass = 1*1/(220*9.80665) ≈ 4.6377e-4 kg
        let m = t.propellant(1.0);
        assert_abs_diff_eq!(m, 1.0 / (220.0 * G0), epsilon = 1e-9);
    }

    #[test]
    fn min_pulse_returns_zero_propellant() {
        let t = nominal();
        assert_abs_diff_eq!(t.propellant(0.005), 0.0, epsilon = 1e-12);
    }

    #[test]
    fn invalid_direction_rejected() {
        let mut t = nominal();
        t.direction = Vector3::new(1.0, 1.0, 0.0); // not unit
        assert!(t.validate().is_err());
    }
}
