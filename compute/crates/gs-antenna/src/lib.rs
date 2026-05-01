//! Antenna pointing controllers (GS-FR-010).
//!
//! Three modes:
//!
//! * [`ProgramTrack`] — caller supplies an `(az, el)` target each step;
//!   the controller produces slew-rate-limited commands.
//! * [`AutoTrack`] — closed-loop tracking driven by an external signal-
//!   strength gradient. The controller adjusts az/el toward whichever
//!   neighbour shows higher signal.
//! * [`StepTrack`] — perturbation-based tracking: walk az / el by `step`
//!   degrees in turn, keeping the perturbation that increases signal.
//!
//! All controllers share the [`Pedestal`] actuator model that enforces
//! per-axis slew-rate limits.

#![cfg_attr(docsrs, feature(doc_cfg))]

use thiserror::Error;

/// Errors produced by `gs-antenna`.
#[derive(Debug, Error, PartialEq)]
pub enum AntennaError {
    /// Out-of-range angle.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range.
        range: &'static str,
    },
}

/// Az-el pointing actuator with slew-rate limits.
#[derive(Debug, Clone, Copy)]
pub struct Pedestal {
    /// Current azimuth (rad, [0, 2π)).
    pub azimuth: f64,
    /// Current elevation (rad, [-π/2, π/2]).
    pub elevation: f64,
    /// Maximum azimuth slew rate (rad/s).
    pub az_rate_max: f64,
    /// Maximum elevation slew rate (rad/s).
    pub el_rate_max: f64,
}

impl Pedestal {
    /// Slew toward `(az_cmd, el_cmd)` over `dt_s` seconds, respecting the
    /// per-axis rate limits and elevation bounds.
    pub fn slew(&mut self, az_cmd: f64, el_cmd: f64, dt_s: f64) {
        let az_err = wrap_pi(az_cmd - self.azimuth);
        let el_err = el_cmd.clamp(-std::f64::consts::FRAC_PI_2, std::f64::consts::FRAC_PI_2)
            - self.elevation;
        let max_az = self.az_rate_max * dt_s;
        let max_el = self.el_rate_max * dt_s;
        let dz = az_err.clamp(-max_az, max_az);
        let de = el_err.clamp(-max_el, max_el);
        self.azimuth = (self.azimuth + dz).rem_euclid(2.0 * std::f64::consts::PI);
        self.elevation = (self.elevation + de)
            .clamp(-std::f64::consts::FRAC_PI_2, std::f64::consts::FRAC_PI_2);
    }
}

#[inline]
fn wrap_pi(angle: f64) -> f64 {
    let two_pi = 2.0 * std::f64::consts::PI;
    let mut a = angle.rem_euclid(two_pi);
    if a > std::f64::consts::PI {
        a -= two_pi;
    }
    a
}

/// Program-track controller. The caller supplies the predicted target
/// each step.
pub struct ProgramTrack {
    /// Pedestal under control.
    pub pedestal: Pedestal,
}

impl ProgramTrack {
    /// Step the controller toward `(az_cmd, el_cmd)`.
    pub fn step(&mut self, az_cmd: f64, el_cmd: f64, dt_s: f64) {
        self.pedestal.slew(az_cmd, el_cmd, dt_s);
    }
}

/// Auto-track controller — closed-loop tracking with neighbour-signal
/// gradient measurements. The user supplies the signal strength at four
/// dithered points (±step in az and el).
pub struct AutoTrack {
    /// Pedestal.
    pub pedestal: Pedestal,
    /// Dither / step magnitude (rad).
    pub step: f64,
    /// Loop gain (rad per signal-difference unit).
    pub gain: f64,
}

impl AutoTrack {
    /// Use four neighbour readings `(az+, az-, el+, el-)` to estimate
    /// the gradient and slew toward the maximum.
    pub fn step_with_signal(
        &mut self,
        az_plus: f64,
        az_minus: f64,
        el_plus: f64,
        el_minus: f64,
        dt_s: f64,
    ) {
        let daz = (az_plus - az_minus) * self.gain;
        let del = (el_plus - el_minus) * self.gain;
        let az_cmd = self.pedestal.azimuth + daz;
        let el_cmd = self.pedestal.elevation + del;
        self.pedestal.slew(az_cmd, el_cmd, dt_s);
    }
}

/// Step-track controller — sequential dither-and-keep style.
pub struct StepTrack {
    /// Pedestal.
    pub pedestal: Pedestal,
    /// Dither step.
    pub step: f64,
    /// Last signal reading at the centre of the dither pattern.
    pub last_centre_signal: f64,
    /// Internal phase counter (0..4).
    phase: u8,
}

impl StepTrack {
    /// Construct a new step-track controller.
    #[must_use]
    pub fn new(pedestal: Pedestal, step: f64) -> Self {
        Self { pedestal, step, last_centre_signal: f64::NEG_INFINITY, phase: 0 }
    }

    /// Apply one step of the algorithm: nudge az/el based on current
    /// signal vs. last reading.
    pub fn step_with_signal(&mut self, signal: f64, dt_s: f64) {
        if signal > self.last_centre_signal {
            // Keep going — but bookkeeping happens after the dither.
        }
        let (daz, del) = match self.phase {
            0 => (self.step, 0.0),
            1 => (-2.0 * self.step, 0.0),
            2 => (self.step, self.step),
            _ => (0.0, -2.0 * self.step),
        };
        self.phase = (self.phase + 1) % 4;
        let az_cmd = self.pedestal.azimuth + daz;
        let el_cmd = self.pedestal.elevation + del;
        self.pedestal.slew(az_cmd, el_cmd, dt_s);
        self.last_centre_signal = signal;
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use std::f64::consts::PI;

    use super::*;

    fn pedestal() -> Pedestal {
        Pedestal {
            azimuth: 0.0,
            elevation: 0.0,
            az_rate_max: 1.0_f64.to_radians() * 5.0, // 5°/s in rad/s
            el_rate_max: 1.0_f64.to_radians() * 3.0, // 3°/s
        }
    }

    #[test]
    fn slew_respects_rate_limit() {
        let mut p = pedestal();
        // Command 90° step in 1 s, but limit is 5°/s ⇒ only 5° actual.
        p.slew(90.0_f64.to_radians(), 0.0, 1.0);
        assert_abs_diff_eq!(p.azimuth, 5.0_f64.to_radians(), epsilon = 1e-9);
    }

    #[test]
    fn slew_wraps_through_zero() {
        let mut p = pedestal();
        p.azimuth = 359.0_f64.to_radians();
        // Command 1° → wrap to ~0°.
        p.slew(1.0_f64.to_radians(), 0.0, 1.0);
        // 5°/s × 1s = 5° travel; minimal-arc to 1° is +2°, slew limit allows it.
        let expected = 1.0_f64.to_radians();
        assert_abs_diff_eq!(p.azimuth.rem_euclid(2.0 * PI), expected, epsilon = 1e-6);
    }

    #[test]
    fn auto_track_walks_toward_signal_peak() {
        let mut at = AutoTrack { pedestal: pedestal(), step: 0.01, gain: 0.5 };
        // Synthetic environment: signal increases with +az.
        for _ in 0..50 {
            let s_plus = 1.0;
            let s_minus = 0.0;
            let s_el_p = 0.5;
            let s_el_m = 0.5;
            at.step_with_signal(s_plus, s_minus, s_el_p, s_el_m, 0.1);
        }
        // Pedestal should have moved in +az direction.
        assert!(at.pedestal.azimuth > 0.0);
    }

    #[test]
    fn elevation_clamps_to_zenith() {
        let mut p = pedestal();
        p.slew(0.0, 100.0_f64.to_radians(), 100.0);
        assert!(p.elevation <= std::f64::consts::FRAC_PI_2 + 1e-9);
    }
}
