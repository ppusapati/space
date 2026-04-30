//! Doppler shift prediction for satellite RF links.
//!
//! Given the relative velocity `v_rel` between the spacecraft and the
//! ground station along the line-of-sight, the (non-relativistic)
//! Doppler-shifted observed frequency is
//!
//! ```text
//! f_obs = f_emit · (1 − v_rad / c)
//! ```
//!
//! where `v_rad` is the range-rate (positive moving away). Equivalently
//! the shift is `Δf = −f_emit · v_rad / c`. For uplink the ground
//! transmitter must compensate by *adding* `+f_emit · v_rad / c` so that
//! the spacecraft sees the nominal frequency. For downlink the ground
//! receiver tunes to `f_emit + Δf`.

#![cfg_attr(docsrs, feature(doc_cfg))]

use nalgebra::Vector3;

/// Speed of light (m/s).
pub const C_LIGHT: f64 = 299_792_458.0;

/// Range-rate (closing speed: positive means range increasing).
#[must_use]
pub fn range_rate(
    sat_position_m: Vector3<f64>,
    sat_velocity_m_s: Vector3<f64>,
    station_position_m: Vector3<f64>,
    station_velocity_m_s: Vector3<f64>,
) -> f64 {
    let r = sat_position_m - station_position_m;
    let v = sat_velocity_m_s - station_velocity_m_s;
    let r_mag = r.norm();
    if r_mag <= 0.0 {
        return 0.0;
    }
    r.dot(&v) / r_mag
}

/// Observed downlink frequency at the ground given the spacecraft's
/// transmit frequency and the radial velocity.
#[must_use]
pub fn doppler_shifted_downlink(emit_hz: f64, range_rate_m_s: f64) -> f64 {
    emit_hz * (1.0 - range_rate_m_s / C_LIGHT)
}

/// Compensated uplink frequency the ground transmitter should emit so
/// that the spacecraft receiver sees `nominal_rx_hz`.
#[must_use]
pub fn uplink_compensated(nominal_rx_hz: f64, range_rate_m_s: f64) -> f64 {
    // Spacecraft sees f_received = f_tx * (1 − v / c). To deliver
    // nominal_rx_hz we set f_tx = nominal_rx_hz / (1 − v / c).
    nominal_rx_hz / (1.0 - range_rate_m_s / C_LIGHT)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use nalgebra::Vector3;

    use super::*;

    #[test]
    fn range_rate_zero_when_stationary() {
        let r = range_rate(
            Vector3::new(7e6, 0.0, 0.0),
            Vector3::zeros(),
            Vector3::zeros(),
            Vector3::zeros(),
        );
        assert_abs_diff_eq!(r, 0.0, epsilon = 1e-12);
    }

    #[test]
    fn range_rate_positive_when_receding() {
        let r = range_rate(
            Vector3::new(7e6, 0.0, 0.0),
            Vector3::new(100.0, 0.0, 0.0),
            Vector3::zeros(),
            Vector3::zeros(),
        );
        assert!(r > 0.0);
    }

    #[test]
    fn downlink_red_shift_when_receding() {
        let f0 = 437.5e6;
        let f = doppler_shifted_downlink(f0, 7_500.0);
        // Receding → observed lower.
        assert!(f < f0);
    }

    #[test]
    fn uplink_compensation_round_trips() {
        let f_target = 145.0e6;
        let v_rad = 6_500.0;
        let f_tx = uplink_compensated(f_target, v_rad);
        // The spacecraft sees f_tx * (1 − v / c) ≈ f_target.
        let f_seen = f_tx * (1.0 - v_rad / C_LIGHT);
        assert_abs_diff_eq!(f_seen, f_target, epsilon = 1e-6);
    }
}
