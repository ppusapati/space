//! Magnetorquer (magnetic rod / coil) model.
//!
//! Each torquer commands a body-frame dipole moment `m ∈ ℝ³` (A·m²)
//! within per-axis limits. The induced body torque in a magnetic field
//! `B` (T) is
//!
//! ```text
//! τ = m × B
//! ```
//!
//! Note the resulting torque is always perpendicular to `B`, so a single
//! magnetorquer cannot produce torque along the local field direction.

use nalgebra::Vector3;

use crate::ActuatorError;

/// A 3-axis magnetorquer assembly.
#[derive(Debug, Clone, Copy)]
pub struct Magnetorquer {
    /// Per-axis maximum dipole moment (A·m²). Each component is the
    /// absolute upper bound on the corresponding body-frame dipole axis.
    pub max_dipole: Vector3<f64>,
}

impl Magnetorquer {
    /// Construct and validate a magnetorquer.
    ///
    /// # Errors
    /// [`ActuatorError::OutOfRange`] if any component is non-positive.
    pub fn new(max_dipole: Vector3<f64>) -> Result<Self, ActuatorError> {
        for (i, &v) in max_dipole.iter().enumerate() {
            if !(v.is_finite() && v > 0.0) {
                let names = ["max_dipole.x", "max_dipole.y", "max_dipole.z"];
                return Err(ActuatorError::OutOfRange {
                    name: names[i],
                    value: v,
                    range: "(0, +inf)",
                });
            }
        }
        Ok(Self { max_dipole })
    }

    /// Clamp a commanded dipole to the per-axis limits and compute the
    /// torque produced in field `b_field` (T).
    #[must_use]
    pub fn torque(&self, dipole_command: Vector3<f64>, b_field: Vector3<f64>) -> Vector3<f64> {
        let m = Vector3::new(
            dipole_command.x.clamp(-self.max_dipole.x, self.max_dipole.x),
            dipole_command.y.clamp(-self.max_dipole.y, self.max_dipole.y),
            dipole_command.z.clamp(-self.max_dipole.z, self.max_dipole.z),
        );
        m.cross(&b_field)
    }

    /// Best-effort dipole that yields a body torque closest to `desired`
    /// in the perpendicular subspace of `b_field`. Implemented by solving
    /// `m × B = τ` in the least-squares sense:
    /// `m = (B × τ) / |B|²` (the component of m parallel to B is chosen
    /// to zero because it produces no torque).
    #[must_use]
    pub fn dipole_for_torque(desired: Vector3<f64>, b_field: Vector3<f64>) -> Vector3<f64> {
        let b_norm_sq = b_field.norm_squared();
        if b_norm_sq <= f64::EPSILON {
            return Vector3::zeros();
        }
        b_field.cross(&desired) / b_norm_sq
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn torque_perpendicular_to_field() {
        let m = Magnetorquer::new(Vector3::new(1.0, 1.0, 1.0)).unwrap();
        let dipole = Vector3::new(0.5, 0.0, 0.0);
        let b = Vector3::new(0.0, 1e-5, 0.0);
        let tau = m.torque(dipole, b);
        // x × y = z, so τ should point in +z with magnitude 0.5 * 1e-5 = 5e-6
        assert_abs_diff_eq!(tau.x, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(tau.y, 0.0, epsilon = 1e-12);
        assert_abs_diff_eq!(tau.z, 5e-6, epsilon = 1e-15);
    }

    #[test]
    fn dipole_for_torque_recovers_perpendicular_demand() {
        let b = Vector3::new(0.0, 1e-5, 0.0);
        let desired = Vector3::new(0.0, 0.0, 5e-6); // perpendicular to b
        let m = Magnetorquer::dipole_for_torque(desired, b);
        // τ_actual = m × b should equal desired.
        let tau = m.cross(&b);
        for i in 0..3 {
            assert_abs_diff_eq!(tau[i], desired[i], epsilon = 1e-12);
        }
    }

    #[test]
    fn rejects_non_positive_max_dipole() {
        let err = Magnetorquer::new(Vector3::new(1.0, -1.0, 1.0)).unwrap_err();
        assert!(matches!(err, ActuatorError::OutOfRange { name: "max_dipole.y", .. }));
    }

    #[test]
    fn dipole_for_torque_with_zero_field_returns_zero() {
        let m = Magnetorquer::dipole_for_torque(Vector3::new(0.0, 0.0, 1e-6), Vector3::zeros());
        assert!(m.iter().all(|v| *v == 0.0));
    }
}
