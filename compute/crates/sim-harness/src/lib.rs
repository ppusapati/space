//! Simulation harness driving a spacecraft mission scenario (SAT-FR-031..032).
//!
//! Provides:
//!
//! * [`Scenario`] — fixed-step simulation loop holding the current
//!   `sim_physics::State`, an inertia tensor, a step size, and a final
//!   time. The loop calls a user-supplied controller to compute torque
//!   and acceleration commands, then advances via the
//!   `sim_physics::integrator::rk4_step` integrator.
//! * [`MonteCarlo`] — wraps a scenario factory closure and runs `n`
//!   independent replays with caller-controlled per-run perturbations,
//!   returning summary statistics on a user-supplied scalar metric.

#![cfg_attr(docsrs, feature(doc_cfg))]

use nalgebra::Vector3;
use sim_physics::{State, integrator::rk4_step, rigid_body::Inertia};

/// Output of a single scenario run.
#[derive(Debug, Clone)]
pub struct RunResult {
    /// Wall-clock simulated duration (s).
    pub duration_s: f64,
    /// Number of integration steps executed.
    pub steps: u64,
    /// Final state.
    pub final_state: State,
    /// Per-step recorded scalar metric (caller-supplied).
    pub metric: Vec<f64>,
}

/// Scenario configuration.
pub struct Scenario<A, T, M>
where
    A: Fn(&State, f64) -> Vector3<f64>,
    T: Fn(&State, f64) -> Vector3<f64>,
    M: Fn(&State, f64) -> f64,
{
    /// Initial state.
    pub initial_state: State,
    /// Spacecraft inertia.
    pub inertia: Inertia,
    /// Integration step size (s).
    pub dt: f64,
    /// Final simulated time (s).
    pub final_time: f64,
    /// Acceleration callback `(state, t) → a (m/s², inertial)`.
    pub acceleration: A,
    /// Torque callback `(state, t) → τ (N·m, body frame)`.
    pub torque: T,
    /// Metric callback recorded once per step.
    pub metric: M,
}

impl<A, T, M> Scenario<A, T, M>
where
    A: Fn(&State, f64) -> Vector3<f64>,
    T: Fn(&State, f64) -> Vector3<f64>,
    M: Fn(&State, f64) -> f64,
{
    /// Run the scenario and return the result.
    ///
    /// `acceleration(state, t)` and `torque(state, t)` are evaluated **once
    /// per outer integration step**; the four RK4 sub-steps reuse the same
    /// time-stamped callable. This is exact when the user-supplied
    /// callbacks depend only on `state`, and a good approximation when
    /// they depend slowly on `t` relative to `dt`.
    pub fn run(&self) -> RunResult {
        let mut state = self.initial_state;
        let mut t = 0.0_f64;
        let mut steps = 0_u64;
        let mut metric = Vec::with_capacity(((self.final_time / self.dt).abs() as usize).max(16));
        loop {
            metric.push((self.metric)(&state, t));
            if t >= self.final_time - 0.5 * self.dt.abs() {
                break;
            }
            let t_now = t;
            let accel_step = |s: &State| (self.acceleration)(s, t_now);
            let torque_step = |s: &State| (self.torque)(s, t_now);
            state = rk4_step(state, self.inertia, self.dt, &accel_step, &torque_step);
            t += self.dt;
            steps += 1;
        }
        RunResult { duration_s: t, steps, final_state: state, metric }
    }
}

/// Monte Carlo runner.
pub struct MonteCarlo<F>
where
    F: Fn(u64) -> RunResult,
{
    /// Number of runs.
    pub runs: u64,
    /// Factory that builds and runs one scenario for the given run index.
    /// Callers should perturb scenario parameters by deriving from the
    /// run index so the runs are reproducible.
    pub factory: F,
}

/// Aggregate statistics produced by [`MonteCarlo::execute`].
#[derive(Debug, Clone, Copy, Default)]
pub struct McStats {
    /// Number of runs executed.
    pub runs: u64,
    /// Mean of the per-run final metric.
    pub mean: f64,
    /// Sample standard deviation of the per-run final metric.
    pub std_dev: f64,
    /// Minimum.
    pub min: f64,
    /// Maximum.
    pub max: f64,
}

impl<F> MonteCarlo<F>
where
    F: Fn(u64) -> RunResult,
{
    /// Execute all runs serially and return aggregate statistics on the
    /// **final** value of each run's metric vector.
    #[must_use]
    pub fn execute(&self) -> McStats {
        if self.runs == 0 {
            return McStats::default();
        }
        let mut samples = Vec::with_capacity(self.runs as usize);
        for i in 0..self.runs {
            let r = (self.factory)(i);
            let val = r.metric.last().copied().unwrap_or(0.0);
            samples.push(val);
        }
        let n = samples.len() as f64;
        let mean = samples.iter().sum::<f64>() / n;
        let var = samples.iter().map(|v| (v - mean).powi(2)).sum::<f64>() / n;
        let min = samples.iter().copied().fold(f64::INFINITY, f64::min);
        let max = samples.iter().copied().fold(f64::NEG_INFINITY, f64::max);
        McStats { runs: self.runs, mean, std_dev: var.sqrt(), min, max }
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;
    use sim_physics::{R_EARTH, gravity};

    use super::*;

    #[test]
    fn scenario_propagates_circular_orbit() {
        let r0 = Vector3::new(R_EARTH + 500_000.0, 0.0, 0.0);
        let v_mag = (sim_physics::MU_EARTH / r0.norm()).sqrt();
        let v0 = Vector3::new(0.0, v_mag, 0.0);
        let inertia = Inertia { diag: Vector3::new(1.0, 1.0, 1.0) };
        let scenario = Scenario {
            initial_state: State {
                r: r0,
                v: v0,
                q: nalgebra::UnitQuaternion::identity(),
                omega: Vector3::zeros(),
            },
            inertia,
            dt: 10.0,
            final_time: 600.0,
            acceleration: |s: &State, _t: f64| gravity::total(s.r),
            torque: |_s: &State, _t: f64| Vector3::zeros(),
            metric: |s: &State, _t: f64| s.r.norm(),
        };
        let result = scenario.run();
        // After 600 s of two-body + J2, altitude should remain near initial.
        let expected = r0.norm();
        assert_abs_diff_eq!(result.final_state.r.norm(), expected, epsilon = 5_000.0);
    }

    #[test]
    fn monte_carlo_aggregates_statistics() {
        let mc = MonteCarlo {
            runs: 100,
            factory: |i: u64| RunResult {
                duration_s: 1.0,
                steps: 1,
                final_state: State::default(),
                metric: vec![(i as f64) * 0.01],
            },
        };
        let stats = mc.execute();
        assert_eq!(stats.runs, 100);
        // Mean of {0, 0.01, 0.02, ..., 0.99} = 0.495
        assert_abs_diff_eq!(stats.mean, 0.495, epsilon = 1e-9);
        assert_abs_diff_eq!(stats.min, 0.0, epsilon = 1e-9);
        assert_abs_diff_eq!(stats.max, 0.99, epsilon = 1e-9);
    }
}
