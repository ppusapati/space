//! Battery state-of-charge and state-of-health estimators (SAT-FR-021).
//!
//! Two complementary techniques are exposed:
//!
//! * **Coulomb counting** — primary SOC estimator. Tracks accumulated
//!   ampere-hour throughput, with charging/discharging coulombic
//!   efficiency.
//! * **Open-Circuit-Voltage correction** — secondary SOC fix-up using a
//!   monotonically increasing OCV-vs-SOC curve. Applied when the
//!   battery is at rest (current near zero for a configurable settle
//!   time).
//!
//! State-of-health is estimated as the ratio of measured present
//! discharge capacity to nameplate capacity.

#![cfg_attr(docsrs, feature(doc_cfg))]

use thiserror::Error;

/// Errors common to the battery module.
#[derive(Debug, Error, PartialEq)]
pub enum BatteryError {
    /// A scalar parameter was out of range.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range description.
        range: &'static str,
    },
}

/// Open-Circuit-Voltage vs SOC curve.
///
/// Holds a monotonic look-up table of `(soc, voltage)` pairs with `soc`
/// strictly increasing and `voltage` non-decreasing.
#[derive(Debug, Clone)]
pub struct OcvSocCurve {
    /// Pairs `(soc ∈ [0, 1], voltage)`.
    points: Vec<(f64, f64)>,
}

impl OcvSocCurve {
    /// Construct from a sorted set of points.
    ///
    /// # Errors
    /// [`BatteryError::OutOfRange`] if fewer than two points or the
    /// points are not monotonic.
    pub fn new(points: Vec<(f64, f64)>) -> Result<Self, BatteryError> {
        if points.len() < 2 {
            return Err(BatteryError::OutOfRange {
                name: "points",
                value: points.len() as f64,
                range: "≥ 2",
            });
        }
        for i in 1..points.len() {
            if points[i - 1].0 >= points[i].0 {
                return Err(BatteryError::OutOfRange {
                    name: "soc",
                    value: points[i].0,
                    range: "strictly increasing",
                });
            }
            if points[i - 1].1 > points[i].1 {
                return Err(BatteryError::OutOfRange {
                    name: "voltage",
                    value: points[i].1,
                    range: "non-decreasing",
                });
            }
        }
        Ok(Self { points })
    }

    /// Linearly interpolate `voltage(soc)`.
    ///
    /// The constructor [`OcvSocCurve::new`] enforces `points.len() >= 2`,
    /// so the internal `first()` / `last()` accesses are infallible.
    #[must_use]
    pub fn voltage_at(&self, soc: f64) -> f64 {
        let s = soc.clamp(self.points.first().unwrap().0, self.points.last().unwrap().0);
        for w in self.points.windows(2) {
            let (s0, v0) = w[0];
            let (s1, v1) = w[1];
            if s >= s0 && s <= s1 {
                let t = (s - s0) / (s1 - s0);
                return v0 + t * (v1 - v0);
            }
        }
        self.points.last().unwrap().1
    }

    /// Inverse: linearly interpolate `soc(voltage)`. Saturates outside
    /// the table range.
    ///
    /// The constructor enforces `points.len() >= 2`, so the internal
    /// `first()` / `last()` accesses are infallible.
    #[must_use]
    pub fn soc_at(&self, voltage: f64) -> f64 {
        let v_first = self.points.first().unwrap().1;
        let v_last = self.points.last().unwrap().1;
        if voltage <= v_first {
            return self.points.first().unwrap().0;
        }
        if voltage >= v_last {
            return self.points.last().unwrap().0;
        }
        for w in self.points.windows(2) {
            let (s0, v0) = w[0];
            let (s1, v1) = w[1];
            if voltage >= v0 && voltage <= v1 {
                if (v1 - v0).abs() < 1e-12 {
                    return s0;
                }
                let t = (voltage - v0) / (v1 - v0);
                return s0 + t * (s1 - s0);
            }
        }
        self.points.last().unwrap().0
    }
}

/// SOC / SOH estimator.
#[derive(Debug, Clone)]
pub struct BatteryEstimator {
    /// Nameplate capacity (A·h).
    pub nameplate_ah: f64,
    /// Present (degraded) capacity (A·h).
    pub present_ah: f64,
    /// Coulombic efficiency on charge (η_c ∈ (0, 1]).
    pub charge_efficiency: f64,
    /// Coulombic efficiency on discharge (η_d ∈ (0, 1]).
    pub discharge_efficiency: f64,
    /// Current SOC estimate ∈ [0, 1].
    pub soc: f64,
    /// OCV-SOC curve.
    pub curve: OcvSocCurve,
    /// Time the battery has been "at rest" — current magnitude below
    /// `rest_current_a` — accumulated since the last non-rest sample.
    pub rest_time: f64,
    /// Threshold below which the battery is treated as resting (A).
    pub rest_current_threshold: f64,
    /// Required rest time before applying an OCV correction (s).
    pub rest_settle_time: f64,
    /// Total cumulative discharged capacity (A·h) — used by SOH.
    pub cumulative_discharge_ah: f64,
}

impl BatteryEstimator {
    /// Construct from initial conditions.
    ///
    /// # Errors
    /// [`BatteryError::OutOfRange`] for invalid parameters.
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        nameplate_ah: f64,
        present_ah: f64,
        charge_efficiency: f64,
        discharge_efficiency: f64,
        initial_soc: f64,
        curve: OcvSocCurve,
        rest_current_threshold: f64,
        rest_settle_time: f64,
    ) -> Result<Self, BatteryError> {
        for (name, v) in [
            ("nameplate_ah", nameplate_ah),
            ("present_ah", present_ah),
            ("charge_efficiency", charge_efficiency),
            ("discharge_efficiency", discharge_efficiency),
            ("rest_current_threshold", rest_current_threshold),
            ("rest_settle_time", rest_settle_time),
        ] {
            if !(v.is_finite() && v >= 0.0) {
                return Err(BatteryError::OutOfRange { name, value: v, range: "[0, +inf)" });
            }
        }
        if !(charge_efficiency > 0.0 && charge_efficiency <= 1.0) {
            return Err(BatteryError::OutOfRange {
                name: "charge_efficiency",
                value: charge_efficiency,
                range: "(0, 1]",
            });
        }
        if !(discharge_efficiency > 0.0 && discharge_efficiency <= 1.0) {
            return Err(BatteryError::OutOfRange {
                name: "discharge_efficiency",
                value: discharge_efficiency,
                range: "(0, 1]",
            });
        }
        if !(0.0..=1.0).contains(&initial_soc) {
            return Err(BatteryError::OutOfRange {
                name: "initial_soc",
                value: initial_soc,
                range: "[0, 1]",
            });
        }
        Ok(Self {
            nameplate_ah,
            present_ah,
            charge_efficiency,
            discharge_efficiency,
            soc: initial_soc,
            curve,
            rest_time: 0.0,
            rest_current_threshold,
            rest_settle_time,
            cumulative_discharge_ah: 0.0,
        })
    }

    /// State-of-health: present_ah / nameplate_ah, clamped to `[0, 1]`.
    #[must_use]
    pub fn soh(&self) -> f64 {
        if self.nameplate_ah <= 0.0 {
            0.0
        } else {
            (self.present_ah / self.nameplate_ah).clamp(0.0, 1.0)
        }
    }

    /// Update SOC using a measured `current_a` (positive = discharge,
    /// negative = charge), terminal `voltage_v`, and time step `dt_s`.
    ///
    /// Coulomb counting drives SOC continuously; if the battery has been
    /// resting for `rest_settle_time` seconds the SOC is reset to the
    /// OCV curve's reading.
    pub fn step(&mut self, current_a: f64, voltage_v: f64, dt_s: f64) {
        if !(dt_s.is_finite() && dt_s > 0.0) {
            return;
        }
        let cap_seconds = self.present_ah * 3600.0;
        let charge_delta = if current_a >= 0.0 {
            // Discharging: η_d < 1 ⇒ each A·s drawn drops SOC by more than
            // 1/cap_seconds (internal losses).
            -current_a * dt_s / (cap_seconds * self.discharge_efficiency)
        } else {
            // Charging: η_c < 1 ⇒ each A·s pushed in stores less than
            // 1/cap_seconds.
            -current_a * dt_s * self.charge_efficiency / cap_seconds
        };
        self.soc = (self.soc + charge_delta).clamp(0.0, 1.0);
        if current_a > 0.0 {
            self.cumulative_discharge_ah += current_a * dt_s / 3600.0;
        }
        // OCV correction when at rest.
        if current_a.abs() < self.rest_current_threshold {
            self.rest_time += dt_s;
            if self.rest_time >= self.rest_settle_time {
                self.soc = self.curve.soc_at(voltage_v);
            }
        } else {
            self.rest_time = 0.0;
        }
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    fn nominal_curve() -> OcvSocCurve {
        OcvSocCurve::new(vec![
            (0.0, 3.00),
            (0.2, 3.30),
            (0.4, 3.55),
            (0.6, 3.80),
            (0.8, 3.95),
            (1.0, 4.20),
        ])
        .unwrap()
    }

    fn nominal_battery() -> BatteryEstimator {
        BatteryEstimator::new(
            10.0, // nameplate Ah
            10.0, // present Ah
            0.98, // η_c
            0.97, // η_d
            0.5,  // initial SOC
            nominal_curve(),
            0.05,  // rest current threshold (A)
            300.0, // rest settle time (s)
        )
        .unwrap()
    }

    #[test]
    fn coulomb_counting_discharge() {
        let mut b = nominal_battery();
        // Discharge at 1 A for 1800 s = 0.5 Ah → ΔSOC ≈ −0.5/(10·η_d) ≈ −0.0515.
        for _ in 0..1800 {
            b.step(1.0, 3.7, 1.0);
        }
        let expected = 0.5 - 0.5 / (10.0 * 0.97);
        assert_abs_diff_eq!(b.soc, expected, epsilon = 1e-3);
    }

    /// Regression: charge-path SOC growth must use the **charge** efficiency
    /// (η_c = 0.98), distinct from the discharge path's η_d = 0.97. A
    /// previous formula bug applied the wrong factor to discharge; this
    /// test pins both paths so a future copy-paste cannot reintroduce it.
    #[test]
    fn coulomb_counting_charge_uses_charge_efficiency() {
        let mut b = nominal_battery();
        b.soc = 0.5;
        // Charge at 1 A for 1800 s = 0.5 Ah; ΔSOC = +0.5·η_c / 10 = 0.049.
        for _ in 0..1800 {
            b.step(-1.0, 3.7, 1.0);
        }
        let expected = 0.5 + 0.5 * 0.98 / 10.0;
        assert_abs_diff_eq!(b.soc, expected, epsilon = 1e-3);
    }

    /// Regression: a full charge-then-discharge cycle should leave the
    /// battery at a slightly *lower* SOC than where it started, by
    /// exactly the round-trip-efficiency loss `(1 − η_c · η_d)`.
    #[test]
    fn round_trip_efficiency_loss() {
        let mut b = nominal_battery();
        b.soc = 0.5;
        // Charge 1 A for 360 s = 0.1 Ah, then discharge 1 A for 360 s.
        for _ in 0..360 {
            b.step(-1.0, 3.7, 1.0);
        }
        for _ in 0..360 {
            b.step(1.0, 3.7, 1.0);
        }
        // Net delta = (0.1 · η_c) − (0.1 / η_d) per nameplate Ah.
        let expected =
            0.5 + (0.1 * 0.98) / 10.0 - (0.1 / 0.97) / 10.0;
        assert_abs_diff_eq!(b.soc, expected, epsilon = 1e-4);
        assert!(b.soc < 0.5, "round-trip should lose energy");
    }

    #[test]
    fn ocv_correction_after_rest() {
        let mut b = nominal_battery();
        b.soc = 0.1; // misaligned estimate
        // Rest for 600 s at 0 A and the true OCV (≈ 3.55 V at SOC 0.4).
        for _ in 0..600 {
            b.step(0.0, 3.55, 1.0);
        }
        assert_abs_diff_eq!(b.soc, 0.4, epsilon = 1e-3);
    }

    #[test]
    fn soh_reflects_present_capacity_loss() {
        let mut b = nominal_battery();
        b.present_ah = 8.0;
        assert_abs_diff_eq!(b.soh(), 0.8, epsilon = 1e-12);
    }

    #[test]
    fn rejects_invalid_curve() {
        let curve = OcvSocCurve::new(vec![(0.5, 3.7), (0.5, 3.8)]);
        assert!(curve.is_err());
    }
}
