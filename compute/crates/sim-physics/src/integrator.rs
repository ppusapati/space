//! 6-DOF integrator that advances a [`State`] by `dt` using RK4.
//!
//! The integrator combines:
//!
//! * Translational dynamics: `r̈ = a_total(r, v)` from gravity + drag + SRP
//!   + user-supplied additional accelerations.
//! * Rotational dynamics: `ω̇` from Euler's equation; `q̇ = ½ q ⊗ ω`.

use nalgebra::{Quaternion, UnitQuaternion, Vector3};

use crate::{State, quaternion_derivative, rigid_body::Inertia, rigid_body::euler_rate};

/// One full RK4 step.
///
/// `accel_fn` returns the inertial-frame acceleration on the spacecraft
/// (m/s²) given the current `State`. `torque_fn` returns the body-frame
/// torque (N·m) given the current `State`.
#[must_use]
pub fn rk4_step<A, T>(
    state: State,
    inertia: Inertia,
    dt: f64,
    accel_fn: &A,
    torque_fn: &T,
) -> State
where
    A: Fn(&State) -> Vector3<f64>,
    T: Fn(&State) -> Vector3<f64>,
{
    let f = |s: &State| Derivative {
        dr: s.v,
        dv: accel_fn(s),
        dq: quaternion_derivative(s.q, s.omega),
        domega: euler_rate(s.omega, torque_fn(s), inertia),
    };
    let k1 = f(&state);
    let s2 = advance(&state, &k1, 0.5 * dt);
    let k2 = f(&s2);
    let s3 = advance(&state, &k2, 0.5 * dt);
    let k3 = f(&s3);
    let s4 = advance(&state, &k3, dt);
    let k4 = f(&s4);

    let dr = (k1.dr + k2.dr * 2.0 + k3.dr * 2.0 + k4.dr) * (dt / 6.0);
    let dv = (k1.dv + k2.dv * 2.0 + k3.dv * 2.0 + k4.dv) * (dt / 6.0);
    let dq = (k1.dq + k2.dq * 2.0 + k3.dq * 2.0 + k4.dq) * (dt / 6.0);
    let domega = (k1.domega + k2.domega * 2.0 + k3.domega * 2.0 + k4.domega) * (dt / 6.0);

    let new_q_inner = state.q.into_inner() + dq;
    let new_q = UnitQuaternion::from_quaternion(new_q_inner);
    State {
        r: state.r + dr,
        v: state.v + dv,
        q: new_q,
        omega: state.omega + domega,
    }
}

struct Derivative {
    dr: Vector3<f64>,
    dv: Vector3<f64>,
    dq: Quaternion<f64>,
    domega: Vector3<f64>,
}

fn advance(s: &State, k: &Derivative, h: f64) -> State {
    let q_inner = s.q.into_inner() + k.dq * h;
    State {
        r: s.r + k.dr * h,
        v: s.v + k.dv * h,
        q: UnitQuaternion::from_quaternion(q_inner),
        omega: s.omega + k.domega * h,
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::{UnitQuaternion, Vector3};

    use super::*;
    use crate::{State, gravity};

    #[test]
    fn circular_orbit_closes() {
        // Two-body circular orbit at 7000 km radius — should return to
        // start after one period.
        let r = 7_000_000.0;
        let v = (crate::MU_EARTH / r).sqrt();
        let mut state = State {
            r: Vector3::new(r, 0.0, 0.0),
            v: Vector3::new(0.0, v, 0.0),
            q: UnitQuaternion::identity(),
            omega: Vector3::zeros(),
        };
        let inertia = crate::rigid_body::Inertia { diag: Vector3::new(1.0, 1.0, 1.0) };
        let accel = |s: &State| gravity::two_body(s.r);
        let torque = |_s: &State| Vector3::zeros();
        let period = 2.0 * std::f64::consts::PI * (r.powi(3) / crate::MU_EARTH).sqrt();
        let dt = 30.0;
        let n = (period / dt).round() as i32;
        for _ in 0..n {
            state = rk4_step(state, inertia, dt, &accel, &torque);
        }
        // RK4 with a 30-s step over one ~5800-s orbit: orbit-closure
        // accuracy is on the order of a few km. We assert the radius is
        // preserved (energy conservation is the relevant invariant) and
        // the angular position has wrapped close to the start.
        let radius_after = state.r.norm();
        assert_abs_diff_eq!(radius_after, r, epsilon = 5_000.0);
        assert_abs_diff_eq!(state.r.x, r, epsilon = 50_000.0);
    }

    #[test]
    fn pure_torque_about_z_increases_omega_z() {
        let mut state = State::default();
        let inertia = crate::rigid_body::Inertia { diag: Vector3::new(1.0, 1.0, 1.0) };
        let accel = |_s: &State| Vector3::zeros();
        let torque = |_s: &State| Vector3::new(0.0, 0.0, 0.01);
        for _ in 0..10 {
            state = rk4_step(state, inertia, 1.0, &accel, &torque);
        }
        assert!(state.omega.z > 0.0);
    }
}
