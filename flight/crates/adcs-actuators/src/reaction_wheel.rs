//! Reaction-wheel array model and torque allocation.
//!
//! Each wheel has a unit mounting axis `aᵢ` in the body frame, a maximum
//! commanded torque `τ_max`, a maximum stored angular momentum `h_max`,
//! and a viscous-friction coefficient `c_f`. A 3-axis body torque demand
//! `τ_b ∈ ℝ³` is allocated across `N ≥ 3` wheels by the Moore–Penrose
//! pseudo-inverse:
//!
//! ```text
//! τ_w = A⁺ τ_b      with A = [a_1  a_2  ⋯  a_N]   (3 × N)
//! ```
//!
//! After allocation each wheel command is clamped to its torque limit and
//! the wheel state (angular momentum) is integrated:
//!
//! ```text
//! h_i_{k+1} = clamp(h_i_k + (τ_i − c_f · ω_i) · dt, ±h_max)
//! ```
//!
//! where `ω_i = h_i / J_i` is the wheel rate (`J_i` is the polar inertia).

use nalgebra::{DMatrix, DVector, Vector3};

use crate::ActuatorError;

/// Single reaction wheel.
#[derive(Debug, Clone, Copy)]
pub struct ReactionWheel {
    /// Mounting axis (unit vector in body frame).
    pub axis: Vector3<f64>,
    /// Maximum commanded torque (N·m).
    pub max_torque: f64,
    /// Maximum stored angular momentum (N·m·s).
    pub max_momentum: f64,
    /// Polar moment of inertia of the wheel (kg·m²).
    pub inertia: f64,
    /// Viscous friction coefficient (N·m·s/rad).
    pub viscous_friction: f64,
    /// Current stored angular momentum (N·m·s).
    pub momentum: f64,
}

/// An array of reaction wheels.
#[derive(Debug, Clone)]
pub struct ReactionWheelArray {
    /// Wheels in fixed order.
    pub wheels: Vec<ReactionWheel>,
    /// Cached pseudo-inverse of the 3×N axis matrix.
    pseudo_inv: DMatrix<f64>,
}

impl ReactionWheelArray {
    /// Build the array. Validates the configuration and computes `A⁺`.
    ///
    /// # Errors
    /// [`ActuatorError::OutOfRange`] if any wheel parameter is non-positive
    /// or any axis is not unit-norm; [`ActuatorError::Allocation`] if the
    /// 3×N axis matrix is rank-deficient.
    pub fn new(wheels: Vec<ReactionWheel>) -> Result<Self, ActuatorError> {
        if wheels.len() < 3 {
            return Err(ActuatorError::Allocation("at least 3 wheels are required"));
        }
        let mut a = DMatrix::<f64>::zeros(3, wheels.len());
        for (i, w) in wheels.iter().enumerate() {
            for (name, v) in [
                ("max_torque", w.max_torque),
                ("max_momentum", w.max_momentum),
                ("inertia", w.inertia),
            ] {
                if !(v.is_finite() && v > 0.0) {
                    return Err(ActuatorError::OutOfRange {
                        name,
                        value: v,
                        range: "(0, +inf)",
                    });
                }
            }
            let n = w.axis.norm();
            if !(n.is_finite() && (n - 1.0).abs() < 1e-6) {
                return Err(ActuatorError::OutOfRange {
                    name: "axis",
                    value: n,
                    range: "unit vector",
                });
            }
            a[(0, i)] = w.axis.x;
            a[(1, i)] = w.axis.y;
            a[(2, i)] = w.axis.z;
        }
        // Pseudo-inverse via SVD.
        let svd = a.clone().svd(true, true);
        let pseudo = svd.pseudo_inverse(1e-12).map_err(|_| {
            ActuatorError::Allocation("axis matrix is rank-deficient")
        })?;
        Ok(Self { wheels, pseudo_inv: pseudo })
    }

    /// Allocate a *demanded body torque* to per-wheel torque commands.
    ///
    /// Convention: when a wheel applies internal torque `τ_w` to its
    /// rotor, the spacecraft body experiences the reaction `−τ_w` along
    /// the wheel axis. To produce a body torque `τ_b`, the wheels must
    /// supply `τ_w = −A⁺ · τ_b`. Each wheel command is then clamped to
    /// `±max_torque`.
    #[must_use]
    pub fn allocate(&self, body_torque_demand: Vector3<f64>) -> Vec<f64> {
        let tau_vec = DVector::<f64>::from_vec(vec![
            body_torque_demand.x,
            body_torque_demand.y,
            body_torque_demand.z,
        ]);
        let cmd = -&self.pseudo_inv * tau_vec;
        cmd.iter()
            .zip(self.wheels.iter())
            .map(|(t, w)| t.clamp(-w.max_torque, w.max_torque))
            .collect()
    }

    /// Step the wheel array forward by `dt`. The applied torque on the
    /// spacecraft is the negative of the wheel torque in the body frame:
    /// `τ_body = −Σ_i a_i · τ_i_clamped`.
    ///
    /// The wheel momentum integrator clamps to `±max_momentum`; whatever is
    /// not absorbed (because the wheel saturated) is discarded.
    pub fn step(&mut self, body_torque_command: Vector3<f64>, dt: f64) -> Vector3<f64> {
        let cmds = self.allocate(body_torque_command);
        let mut applied = Vector3::<f64>::zeros();
        for (w, &cmd) in self.wheels.iter_mut().zip(cmds.iter()) {
            let omega = w.momentum / w.inertia;
            let net = cmd - w.viscous_friction * omega;
            let new_h = (w.momentum + net * dt).clamp(-w.max_momentum, w.max_momentum);
            // Effective torque exerted on the spacecraft is the change in
            // wheel momentum / dt with sign reversed.
            let effective = (new_h - w.momentum) / dt;
            applied -= effective * w.axis;
            w.momentum = new_h;
        }
        applied
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    fn standard_3() -> ReactionWheelArray {
        // Three orthogonal wheels along body axes.
        let w = |axis: Vector3<f64>| ReactionWheel {
            axis,
            max_torque: 0.05,
            max_momentum: 0.1,
            inertia: 1e-4,
            viscous_friction: 0.0,
            momentum: 0.0,
        };
        ReactionWheelArray::new(vec![
            w(Vector3::x()),
            w(Vector3::y()),
            w(Vector3::z()),
        ])
        .unwrap()
    }

    #[test]
    fn pseudo_inverse_orthogonal_basis_inverts_with_sign_flip() {
        // Demanded body torque +0.01 about x ⇒ wheel command must be -0.01
        // (Newton's 3rd law: body sees −τ_w).
        let mut arr = standard_3();
        let cmds = arr.allocate(Vector3::new(0.01, -0.02, 0.03));
        assert_abs_diff_eq!(cmds[0], -0.01, epsilon = 1e-9);
        assert_abs_diff_eq!(cmds[1], 0.02, epsilon = 1e-9);
        assert_abs_diff_eq!(cmds[2], -0.03, epsilon = 1e-9);
        let _ = arr.step(Vector3::new(0.01, 0.0, 0.0), 0.01);
    }

    #[test]
    fn step_returns_negative_torque_on_spacecraft() {
        let mut arr = standard_3();
        // Demand 0.01 N·m about x.
        let applied = arr.step(Vector3::new(0.01, 0.0, 0.0), 0.1);
        // Newton's 3rd law: spacecraft sees −τ_w, but the array reports the
        // *applied* torque sign convention as positive when the wheel spins
        // up to produce +x body torque (ie. effective is negative; applied = -effective).
        assert_abs_diff_eq!(applied.x, 0.01, epsilon = 1e-12);
    }

    #[test]
    fn momentum_saturates() {
        let mut arr = standard_3();
        // Drive x wheel to saturation. Demand +0.05 N·m on body ⇒ wheel
        // commanded −0.05 N·m, so its momentum runs to −max_momentum.
        for _ in 0..1000 {
            let _ = arr.step(Vector3::new(0.05, 0.0, 0.0), 0.1);
        }
        assert_abs_diff_eq!(arr.wheels[0].momentum, -0.1, epsilon = 1e-9);
    }

    /// Regression: closed-loop check that demanded body torque equals
    /// the body torque actually applied while the wheels are below
    /// saturation. A previous bug had the sign wrong on allocation,
    /// which this guards against.
    #[test]
    fn applied_body_torque_matches_demand_unsaturated() {
        let mut arr = standard_3();
        let demand = Vector3::new(0.01, -0.005, 0.02);
        let applied = arr.step(demand, 0.01);
        for i in 0..3 {
            assert_abs_diff_eq!(applied[i], demand[i], epsilon = 1e-9);
        }
    }

    #[test]
    fn rejects_non_unit_axis() {
        let w = ReactionWheel {
            axis: Vector3::new(2.0, 0.0, 0.0),
            max_torque: 0.01,
            max_momentum: 0.1,
            inertia: 1e-4,
            viscous_friction: 0.0,
            momentum: 0.0,
        };
        let err = ReactionWheelArray::new(vec![w; 3]).unwrap_err();
        assert!(matches!(err, ActuatorError::OutOfRange { name: "axis", .. }));
    }

    #[test]
    fn rejects_too_few_wheels() {
        let w = ReactionWheel {
            axis: Vector3::x(),
            max_torque: 0.01,
            max_momentum: 0.1,
            inertia: 1e-4,
            viscous_friction: 0.0,
            momentum: 0.0,
        };
        let err = ReactionWheelArray::new(vec![w; 2]).unwrap_err();
        assert!(matches!(err, ActuatorError::Allocation(_)));
    }
}
