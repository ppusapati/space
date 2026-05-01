//! Euler's rotational equations for a rigid body.
//!
//! For a body with diagonal principal-axis inertia tensor
//! `J = diag(J₁, J₂, J₃)` and applied torque `τ` (body-frame),
//! the angular acceleration is:
//!
//! ```text
//! ω̇ = J⁻¹ (τ − ω × Jω)
//! ```

use nalgebra::Vector3;

use crate::SimError;

/// Diagonal principal-axis inertia tensor.
#[derive(Debug, Clone, Copy)]
pub struct Inertia {
    /// Principal moments of inertia `[J₁, J₂, J₃]` (kg·m²).
    pub diag: Vector3<f64>,
}

impl Inertia {
    /// Validate the inertia.
    ///
    /// # Errors
    /// [`SimError::OutOfRange`] if any component is non-positive.
    pub fn validate(&self) -> Result<(), SimError> {
        for (i, &j) in self.diag.iter().enumerate() {
            if !(j.is_finite() && j > 0.0) {
                let names = ["diag.x", "diag.y", "diag.z"];
                return Err(SimError::OutOfRange {
                    name: names[i],
                    value: j,
                    range: "(0, +inf)",
                });
            }
        }
        Ok(())
    }
}

/// Euler equation: `ω̇ = J⁻¹ (τ − ω × Jω)` for diagonal inertia.
#[must_use]
pub fn euler_rate(omega: Vector3<f64>, torque: Vector3<f64>, j: Inertia) -> Vector3<f64> {
    let j_omega = Vector3::new(
        j.diag.x * omega.x,
        j.diag.y * omega.y,
        j.diag.z * omega.z,
    );
    let cross = omega.cross(&j_omega);
    let net = torque - cross;
    Vector3::new(net.x / j.diag.x, net.y / j.diag.y, net.z / j.diag.z)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn pure_torque_produces_proportional_rate() {
        let j = Inertia { diag: Vector3::new(1.0, 1.0, 1.0) };
        let omega = Vector3::zeros();
        let torque = Vector3::new(0.1, 0.0, 0.0);
        let dot = euler_rate(omega, torque, j);
        assert_abs_diff_eq!(dot.x, 0.1, epsilon = 1e-12);
    }

    #[test]
    fn gyroscopic_coupling() {
        // Spinning about x with non-equal inertia → gyroscopic torque
        // produces non-zero ω̇ on other axes when ω has multiple components.
        let j = Inertia { diag: Vector3::new(1.0, 2.0, 3.0) };
        let omega = Vector3::new(1.0, 0.5, 0.0);
        let torque = Vector3::zeros();
        let dot = euler_rate(omega, torque, j);
        assert!(dot.norm() > 0.0);
    }
}
