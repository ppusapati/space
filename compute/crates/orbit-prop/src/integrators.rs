//! Numerical integrators for ODE propagation.
//!
//! * [`rk4`] — fixed-step classical Runge-Kutta of order 4.
//! * [`rk45_dopri`] — embedded Dormand-Prince RK4(5) with PI step-size
//!   control. Suitable for orbit propagation under perturbations such as
//!   J2/J3, atmospheric drag, solar radiation pressure, and third-body
//!   gravity.
//!
//! The state is represented as a [`nalgebra::DVector<f64>`] so that the
//! same routines can integrate any system size — for orbital mechanics,
//! the canonical 6-dimensional state vector is `[r_x, r_y, r_z, v_x, v_y,
//! v_z]`.

use nalgebra::DVector;

/// One classical RK4 step.
///
/// `f` is `dy/dt = f(t, y)`. Returns `y_{k+1} = y_k + Δy`.
#[must_use]
pub fn rk4<F>(f: &F, t: f64, y: &DVector<f64>, h: f64) -> DVector<f64>
where
    F: Fn(f64, &DVector<f64>) -> DVector<f64>,
{
    let k1 = f(t, y);
    let k2 = f(t + 0.5 * h, &(y + &k1 * (0.5 * h)));
    let k3 = f(t + 0.5 * h, &(y + &k2 * (0.5 * h)));
    let k4 = f(t + h, &(y + &k3 * h));
    y + (k1 + 2.0 * k2 + 2.0 * k3 + k4) * (h / 6.0)
}

/// One adaptive RK4(5) step using the Dormand-Prince tableau. Returns
/// the new state, the *suggested* next step, and the estimated error.
/// If the local error exceeds `tol`, the step is rejected by the caller.
///
/// The integration loop is performed by [`rk45_dopri`].
fn dopri_step<F>(
    f: &F,
    t: f64,
    y: &DVector<f64>,
    h: f64,
) -> (DVector<f64>, DVector<f64>, f64)
where
    F: Fn(f64, &DVector<f64>) -> DVector<f64>,
{
    // Dormand-Prince coefficients (Hairer, Norsett & Wanner, 1993).
    const C2: f64 = 1.0 / 5.0;
    const C3: f64 = 3.0 / 10.0;
    const C4: f64 = 4.0 / 5.0;
    const C5: f64 = 8.0 / 9.0;

    const A21: f64 = 1.0 / 5.0;
    const A31: f64 = 3.0 / 40.0;
    const A32: f64 = 9.0 / 40.0;
    const A41: f64 = 44.0 / 45.0;
    const A42: f64 = -56.0 / 15.0;
    const A43: f64 = 32.0 / 9.0;
    const A51: f64 = 19_372.0 / 6_561.0;
    const A52: f64 = -25_360.0 / 2_187.0;
    const A53: f64 = 64_448.0 / 6_561.0;
    const A54: f64 = -212.0 / 729.0;
    const A61: f64 = 9_017.0 / 3_168.0;
    const A62: f64 = -355.0 / 33.0;
    const A63: f64 = 46_732.0 / 5_247.0;
    const A64: f64 = 49.0 / 176.0;
    const A65: f64 = -5_103.0 / 18_656.0;
    const A71: f64 = 35.0 / 384.0;
    const A73: f64 = 500.0 / 1_113.0;
    const A74: f64 = 125.0 / 192.0;
    const A75: f64 = -2_187.0 / 6_784.0;
    const A76: f64 = 11.0 / 84.0;

    // 5th-order coefficients (b).
    const B1: f64 = 35.0 / 384.0;
    const B3: f64 = 500.0 / 1_113.0;
    const B4: f64 = 125.0 / 192.0;
    const B5: f64 = -2_187.0 / 6_784.0;
    const B6: f64 = 11.0 / 84.0;
    // 4th-order coefficients (b̂).
    const E1: f64 = 71.0 / 57_600.0;
    const E3: f64 = -71.0 / 16_695.0;
    const E4: f64 = 71.0 / 1_920.0;
    const E5: f64 = -17_253.0 / 339_200.0;
    const E6: f64 = 22.0 / 525.0;
    const E7: f64 = -1.0 / 40.0;

    let k1 = f(t, y);
    let k2 = f(t + C2 * h, &(y + &k1 * (A21 * h)));
    let k3 = f(t + C3 * h, &(y + &k1 * (A31 * h) + &k2 * (A32 * h)));
    let k4 = f(t + C4 * h, &(y + &k1 * (A41 * h) + &k2 * (A42 * h) + &k3 * (A43 * h)));
    let k5 = f(
        t + C5 * h,
        &(y + &k1 * (A51 * h) + &k2 * (A52 * h) + &k3 * (A53 * h) + &k4 * (A54 * h)),
    );
    let k6 = f(
        t + h,
        &(y
            + &k1 * (A61 * h)
            + &k2 * (A62 * h)
            + &k3 * (A63 * h)
            + &k4 * (A64 * h)
            + &k5 * (A65 * h)),
    );
    let y5 = y
        + &k1 * (A71 * h)
        + &k3 * (A73 * h)
        + &k4 * (A74 * h)
        + &k5 * (A75 * h)
        + &k6 * (A76 * h);
    let k7 = f(t + h, &y5);

    let y_high = y + (&k1 * B1 + &k3 * B3 + &k4 * B4 + &k5 * B5 + &k6 * B6) * h;
    let err_vec =
        (&k1 * E1 + &k3 * E3 + &k4 * E4 + &k5 * E5 + &k6 * E6 + &k7 * E7) * h;
    let err_norm = err_vec.norm();
    (y_high, err_vec, err_norm)
}

/// Adaptive Dormand-Prince RK4(5) integrator with PI step-size control.
/// Integrates `dy/dt = f(t, y)` from `(t0, y0)` to `tf`. Returns the
/// final state.
///
/// `tol_abs` is the absolute error tolerance; the step size is shrunk
/// when the local truncation error exceeds it.
///
/// # Errors
/// Returns the final state on success. Step size is bounded below by
/// `h_min`; if the requested step falls below `h_min` the integrator
/// returns the best-effort state at the current `t` and `Err`.
#[allow(clippy::missing_errors_doc)]
pub fn rk45_dopri<F>(
    f: F,
    t0: f64,
    y0: DVector<f64>,
    tf: f64,
    tol_abs: f64,
    h_init: f64,
    h_min: f64,
) -> Result<DVector<f64>, &'static str>
where
    F: Fn(f64, &DVector<f64>) -> DVector<f64>,
{
    let mut t = t0;
    let mut y = y0;
    let mut h = h_init.copysign(tf - t0);
    let max_iter: u32 = 1_000_000;
    let mut iter: u32 = 0;
    let safety = 0.9_f64;
    while (tf - t).abs() > 1e-15 {
        if iter >= max_iter {
            return Err("RK45 step iteration cap exceeded");
        }
        iter += 1;
        // Clamp h so we don't step past tf in the requested direction.
        if (tf - t).abs() < h.abs() {
            h = tf - t;
        }
        let (y_new, _err_vec, err) = dopri_step(&f, t, &y, h);
        if err <= tol_abs || h.abs() <= h_min {
            // Accept.
            t += h;
            y = y_new;
            // PI step-size scale-up.
            let factor = if err > 0.0 {
                safety * (tol_abs / err).powf(1.0 / 5.0)
            } else {
                5.0
            };
            h *= factor.clamp(0.2, 5.0);
        } else {
            // Reject — shrink step.
            let factor = safety * (tol_abs / err).powf(1.0 / 5.0);
            h *= factor.clamp(0.1, 1.0);
            if h.abs() < h_min {
                return Err("step size shrank below h_min");
            }
        }
    }
    Ok(y)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::DVector;

    use super::*;

    #[test]
    fn rk4_solves_linear_ode_exact() {
        // y' = -y, y(0) = 1; analytical: y = exp(-t).
        let f = |_t: f64, y: &DVector<f64>| -y.clone();
        let mut y = DVector::from_vec(vec![1.0]);
        let dt = 0.01;
        let mut t = 0.0;
        for _ in 0..1000 {
            y = rk4(&f, t, &y, dt);
            t += dt;
        }
        // exp(-10) ≈ 4.54e-5
        assert_abs_diff_eq!(y[0], (-10.0_f64).exp(), epsilon = 1e-7);
    }

    #[test]
    fn rk45_simple_harmonic_oscillator() {
        // y'' = -y → state = (y, y'); analytical y(t) = cos(t).
        let f = |_t: f64, x: &DVector<f64>| DVector::from_vec(vec![x[1], -x[0]]);
        let y0 = DVector::from_vec(vec![1.0, 0.0]);
        let yf = rk45_dopri(f, 0.0, y0, 10.0, 1e-6, 0.1, 1e-9).unwrap();
        assert_abs_diff_eq!(yf[0], (10.0_f64).cos(), epsilon = 1e-4);
        assert_abs_diff_eq!(yf[1], -(10.0_f64).sin(), epsilon = 1e-4);
    }

    #[test]
    fn rk45_two_body_circular_orbit() {
        // Two-body in km / km/s with μ = 398_600.4418.
        const MU: f64 = 398_600.441_8;
        let f = |_t: f64, s: &DVector<f64>| {
            let r = nalgebra::Vector3::new(s[0], s[1], s[2]);
            let r_mag = r.norm();
            let acc = -MU * r / r_mag.powi(3);
            DVector::from_vec(vec![s[3], s[4], s[5], acc.x, acc.y, acc.z])
        };
        // Circular orbit at 7000 km radius: v_circ = sqrt(MU / r).
        let r = 7000.0;
        let v = (MU / r).sqrt();
        let y0 = DVector::from_vec(vec![r, 0.0, 0.0, 0.0, v, 0.0]);
        let period = 2.0 * std::f64::consts::PI * (r.powi(3) / MU).sqrt();
        let yf = rk45_dopri(f, 0.0, y0, period, 1e-6, 60.0, 1e-3).unwrap();
        // After one period the orbit should close to within tol.
        assert_abs_diff_eq!(yf[0], r, epsilon = 1.0);
        assert_abs_diff_eq!(yf[1], 0.0, epsilon = 1.0);
        assert_abs_diff_eq!(yf[2], 0.0, epsilon = 1e-6);
    }
}
