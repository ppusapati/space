//! Perturb-and-Observe Maximum Power Point Tracker.
//!
//! Maintains a duty-cycle command for a boost converter that tracks the
//! solar array's maximum power point. The classic P&O algorithm:
//!
//! ```text
//! ΔP = P_now − P_prev
//! ΔV = V_now − V_prev
//! sign = (ΔP > 0) XOR (ΔV < 0)   ; True ⇒ increase duty
//! D ← clamp(D + sign × step, D_min, D_max)
//! ```
//!
//! When the array reaches the MPP the duty oscillates by ±1 step around
//! the optimum; the ripple amplitude is `step`, and the average tracks
//! the true MPP.

use crate::EpsError;

/// MPPT controller state.
#[derive(Debug, Clone)]
pub struct PerturbAndObserve {
    /// Current duty-cycle command in `[0, 1]`.
    pub duty: f64,
    /// Step magnitude per cycle.
    pub step: f64,
    /// Lower bound on duty.
    pub duty_min: f64,
    /// Upper bound on duty.
    pub duty_max: f64,
    /// Previous voltage (V).
    pub last_voltage: f64,
    /// Previous power (W).
    pub last_power: f64,
    /// Whether the controller has been initialised.
    initialised: bool,
}

impl PerturbAndObserve {
    /// Construct from initial duty and step parameters.
    ///
    /// # Errors
    /// [`EpsError::OutOfRange`] for invalid duty / step / limits.
    pub fn new(initial_duty: f64, step: f64, duty_min: f64, duty_max: f64) -> Result<Self, EpsError> {
        for (name, v) in [
            ("initial_duty", initial_duty),
            ("step", step),
            ("duty_min", duty_min),
            ("duty_max", duty_max),
        ] {
            if !v.is_finite() {
                return Err(EpsError::OutOfRange { name, value: v, range: "finite" });
            }
        }
        if duty_min >= duty_max {
            return Err(EpsError::OutOfRange {
                name: "duty_min/max",
                value: duty_min,
                range: "duty_min < duty_max",
            });
        }
        if !(duty_min >= 0.0 && duty_max <= 1.0) {
            return Err(EpsError::OutOfRange {
                name: "duty_min/max",
                value: duty_min,
                range: "[0, 1]",
            });
        }
        if !(initial_duty >= duty_min && initial_duty <= duty_max) {
            return Err(EpsError::OutOfRange {
                name: "initial_duty",
                value: initial_duty,
                range: "[duty_min, duty_max]",
            });
        }
        if step <= 0.0 {
            return Err(EpsError::OutOfRange {
                name: "step",
                value: step,
                range: "(0, +inf)",
            });
        }
        Ok(Self {
            duty: initial_duty,
            step,
            duty_min,
            duty_max,
            last_voltage: 0.0,
            last_power: 0.0,
            initialised: false,
        })
    }

    /// Run one P&O cycle given the current measured voltage (V) and
    /// current (A). Returns the updated duty command.
    pub fn step(&mut self, voltage: f64, current: f64) -> f64 {
        let power = voltage * current;
        if !self.initialised {
            self.last_voltage = voltage;
            self.last_power = power;
            self.initialised = true;
            return self.duty;
        }
        let dp = power - self.last_power;
        let dv = voltage - self.last_voltage;
        // Decide whether to increase or decrease duty.
        let increase = (dp > 0.0) ^ (dv < 0.0);
        let delta = if increase { self.step } else { -self.step };
        self.duty = (self.duty + delta).clamp(self.duty_min, self.duty_max);
        self.last_voltage = voltage;
        self.last_power = power;
        self.duty
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Linear IV curve: I = (1 − V) for V ∈ [0, 1] yields a parabolic
    /// power curve P(V) = V·(1 − V) with a unique MPP at V = 0.5.
    fn iv(v: f64) -> f64 {
        (1.0 - v).max(0.0)
    }

    #[test]
    fn p_and_o_climbs_to_mpp() {
        // Map duty → array voltage as V = duty (boost-converter convention
        // where higher duty operates the array at a higher voltage). MPP
        // therefore lies at duty = 0.5 for a P(V) = V·(1 − V) curve.
        // Start away from boundary so the algorithm sees a gradient.
        let mut mppt = PerturbAndObserve::new(0.1, 0.005, 0.0, 1.0).unwrap();
        for _ in 0..2000 {
            let v = mppt.duty.clamp(0.0, 1.0);
            let i = iv(v);
            mppt.step(v, i);
        }
        assert!(
            (mppt.duty - 0.5).abs() < 0.02,
            "duty did not converge near MPP (0.5): {}",
            mppt.duty
        );
    }

    #[test]
    fn rejects_invalid_step() {
        assert!(PerturbAndObserve::new(0.5, 0.0, 0.0, 1.0).is_err());
    }
}
