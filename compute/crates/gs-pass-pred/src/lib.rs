//! Pass prediction for an SGP4-defined satellite over a ground station.
//!
//! For each pass we report:
//!
//! * **AOS** (Acquisition of Signal) — first time the elevation crosses
//!   above the configured horizon mask.
//! * **TCA** (Time of Closest Approach) — instant of maximum elevation.
//! * **LOS** (Loss of Signal) — last time the elevation drops below the
//!   horizon mask.
//!
//! The ground station is an `(lat, lon, height_m)` geodetic location.
//! Elevation / azimuth are computed from the TEME → ECEF satellite
//! position and a topocentric (East-North-Up) transform.

#![cfg_attr(docsrs, feature(doc_cfg))]

use gs_tle::{Sgp4Propagator, teme_to_ecef};
use nalgebra::{Matrix3, Vector3};

/// Geodetic ground-station location.
#[derive(Debug, Clone, Copy)]
pub struct GroundStation {
    /// Latitude (rad, geodetic).
    pub lat_rad: f64,
    /// Longitude (rad, east-positive).
    pub lon_rad: f64,
    /// Height above the WGS-84 ellipsoid (m).
    pub height_m: f64,
    /// Horizon-mask elevation (rad). Passes are reported only when the
    /// satellite climbs above this.
    pub min_elevation_rad: f64,
}

/// One pass.
#[derive(Debug, Clone, Copy)]
pub struct Pass {
    /// Acquisition-of-signal Julian Date (UT1).
    pub aos_jd: f64,
    /// Time of closest approach (Julian Date, UT1).
    pub tca_jd: f64,
    /// Loss-of-signal Julian Date (UT1).
    pub los_jd: f64,
    /// Maximum elevation reached during the pass (rad).
    pub max_elevation_rad: f64,
}

/// Convert geodetic `(lat, lon, h)` to ECEF using WGS-84.
#[must_use]
pub fn geodetic_to_ecef(lat_rad: f64, lon_rad: f64, height_m: f64) -> Vector3<f64> {
    const A: f64 = 6_378_137.0;
    const E2: f64 = 6.694_379_990_141_316_e-3;
    let n = A / (1.0 - E2 * lat_rad.sin().powi(2)).sqrt();
    let x = (n + height_m) * lat_rad.cos() * lon_rad.cos();
    let y = (n + height_m) * lat_rad.cos() * lon_rad.sin();
    let z = (n * (1.0 - E2) + height_m) * lat_rad.sin();
    Vector3::new(x, y, z)
}

/// Build the rotation matrix from ECEF to the topocentric (East-North-Up)
/// frame at `(lat, lon)`.
#[must_use]
pub fn ecef_to_enu_rotation(lat_rad: f64, lon_rad: f64) -> Matrix3<f64> {
    let sl = lat_rad.sin();
    let cl = lat_rad.cos();
    let so = lon_rad.sin();
    let co = lon_rad.cos();
    Matrix3::new(
        -so, co, 0.0, // east
        -sl * co, -sl * so, cl, // north
        cl * co, cl * so, sl, // up
    )
}

/// Compute (azimuth, elevation, range) of the satellite from the ground
/// station, with `r_sat_km_teme` from the SGP4 propagator and `jd_ut1`
/// the corresponding Julian Date.
#[must_use]
pub fn look_angles(
    station: GroundStation,
    r_sat_km_teme: Vector3<f64>,
    jd_ut1: f64,
) -> (f64, f64, f64) {
    let r_sat_m_ecef = teme_to_ecef(r_sat_km_teme, jd_ut1) * 1_000.0;
    let r_station_ecef = geodetic_to_ecef(station.lat_rad, station.lon_rad, station.height_m);
    let r_topo_ecef = r_sat_m_ecef - r_station_ecef;
    let r_enu = ecef_to_enu_rotation(station.lat_rad, station.lon_rad) * r_topo_ecef;
    let range = r_enu.norm();
    let elevation = (r_enu.z / range).asin();
    let azimuth = r_enu.x.atan2(r_enu.y).rem_euclid(2.0 * std::f64::consts::PI);
    (azimuth, elevation, range)
}

/// Predict passes by sampling the propagator at `step_seconds`-second
/// intervals over `[start_jd, start_jd + duration_days]`. Each pass is
/// refined with bisection to better than 1 second on AOS / LOS and TCA.
#[must_use]
pub fn predict_passes(
    propagator: &Sgp4Propagator,
    epoch_jd: f64,
    station: GroundStation,
    start_jd: f64,
    duration_days: f64,
    step_seconds: f64,
) -> Vec<Pass> {
    let mut passes = Vec::new();
    let mut t = start_jd;
    let end = start_jd + duration_days;
    let dt_days = step_seconds / 86_400.0;
    let mut prev_above = false;
    let mut entry_jd = 0.0_f64;
    let mut max_elev = f64::NEG_INFINITY;
    let mut max_jd = 0.0_f64;
    while t <= end + 1e-12 {
        let elev = sample_elevation(propagator, epoch_jd, station, t);
        let above = elev >= station.min_elevation_rad;
        if above && !prev_above {
            entry_jd = t;
            max_elev = elev;
            max_jd = t;
        } else if above && elev > max_elev {
            max_elev = elev;
            max_jd = t;
        } else if !above && prev_above {
            // Pass just ended.
            let aos = refine_crossing(
                propagator,
                epoch_jd,
                station,
                entry_jd - dt_days,
                entry_jd,
                false,
            );
            let los = refine_crossing(
                propagator,
                epoch_jd,
                station,
                t - dt_days,
                t,
                true,
            );
            let tca = refine_extremum(propagator, epoch_jd, station, max_jd - dt_days, max_jd + dt_days);
            passes.push(Pass {
                aos_jd: aos,
                tca_jd: tca,
                los_jd: los,
                max_elevation_rad: sample_elevation(propagator, epoch_jd, station, tca),
            });
            max_elev = f64::NEG_INFINITY;
        }
        prev_above = above;
        t += dt_days;
    }
    passes
}

fn sample_elevation(
    propagator: &Sgp4Propagator,
    epoch_jd: f64,
    station: GroundStation,
    jd: f64,
) -> f64 {
    let minutes = (jd - epoch_jd) * 1_440.0;
    match propagator.propagate(minutes) {
        Ok(s) => look_angles(station, s.position_km, jd).1,
        Err(_) => f64::NEG_INFINITY,
    }
}

fn refine_crossing(
    propagator: &Sgp4Propagator,
    epoch_jd: f64,
    station: GroundStation,
    mut t_below: f64,
    mut t_above: f64,
    above_then_below: bool,
) -> f64 {
    if above_then_below {
        std::mem::swap(&mut t_below, &mut t_above);
    }
    for _ in 0..40 {
        let mid = 0.5 * (t_below + t_above);
        let e = sample_elevation(propagator, epoch_jd, station, mid);
        if e >= station.min_elevation_rad {
            t_above = mid;
        } else {
            t_below = mid;
        }
        if (t_above - t_below).abs() * 86_400.0 < 0.5 {
            break;
        }
    }
    0.5 * (t_below + t_above)
}

fn refine_extremum(
    propagator: &Sgp4Propagator,
    epoch_jd: f64,
    station: GroundStation,
    a: f64,
    b: f64,
) -> f64 {
    // Golden-section search for the maximum.
    let phi = 0.618_033_988_749_894_9_f64;
    let mut x1 = b - phi * (b - a);
    let mut x2 = a + phi * (b - a);
    let mut a = a;
    let mut b = b;
    let mut f1 = sample_elevation(propagator, epoch_jd, station, x1);
    let mut f2 = sample_elevation(propagator, epoch_jd, station, x2);
    for _ in 0..40 {
        if f1 < f2 {
            a = x1;
            x1 = x2;
            f1 = f2;
            x2 = a + phi * (b - a);
            f2 = sample_elevation(propagator, epoch_jd, station, x2);
        } else {
            b = x2;
            x2 = x1;
            f2 = f1;
            x1 = b - phi * (b - a);
            f1 = sample_elevation(propagator, epoch_jd, station, x1);
        }
        if (b - a) * 86_400.0 < 0.1 {
            break;
        }
    }
    0.5 * (a + b)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    #[test]
    fn enu_rotation_orthonormal() {
        let r = ecef_to_enu_rotation(0.5, 1.0);
        let rt = r.transpose();
        let id = r * rt;
        for i in 0..3 {
            for j in 0..3 {
                let expect = if i == j { 1.0 } else { 0.0 };
                assert_abs_diff_eq!(id[(i, j)], expect, epsilon = 1e-12);
            }
        }
    }

    #[test]
    fn geodetic_to_ecef_at_north_pole_yields_z_only() {
        let r = geodetic_to_ecef(std::f64::consts::FRAC_PI_2, 0.0, 0.0);
        assert_abs_diff_eq!(r.x, 0.0, epsilon = 1e-6);
        assert_abs_diff_eq!(r.y, 0.0, epsilon = 1e-6);
        // Polar radius ≈ 6356752 m.
        assert!(r.z > 6_350_000.0 && r.z < 6_360_000.0);
    }

    #[test]
    fn elevation_overhead_is_pi_over_two() {
        // Place the satellite directly above the station at 1000 km
        // altitude. Check that elevation ≈ 90°.
        let lat = 0.0;
        let lon = 0.0;
        let height = 0.0;
        let station_ecef = geodetic_to_ecef(lat, lon, height);
        let sat_ecef = station_ecef.normalize() * (station_ecef.norm() + 1_000_000.0);
        // Pretend ECEF == TEME (we apply identity by passing jd that gives gmst = 0).
        // gmst(2451545.0) ≈ 4.89 rad, so we approximate by directly providing
        // a jd that yields gmst ≈ 0 — easier just to bypass with a manual
        // teme_to_ecef inverse via an identity rotation.
        // Construct a "TEME" position equal to ECEF (i.e., use jd where
        // gmst = 0). Instead, directly compute angles using r_topo_ecef.
        let r_topo_ecef = sat_ecef - station_ecef;
        let r_enu = ecef_to_enu_rotation(lat, lon) * r_topo_ecef;
        let range = r_enu.norm();
        let elevation = (r_enu.z / range).asin();
        assert_abs_diff_eq!(elevation, std::f64::consts::FRAC_PI_2, epsilon = 1e-6);
    }

    /// End-to-end pass prediction over a Vanguard-1 TLE for a single
    /// 24-hour window. Vanguard-1 has a ~133-minute orbit, so a station
    /// at the equator with a wide horizon mask should see at least one
    /// pass per day. Each detected pass must have AOS < TCA < LOS.
    #[test]
    fn predict_passes_finds_at_least_one_pass_per_day() {
        const VANGUARD_LINE1: &str =
            "1 00005U 58002B   00179.78495062  .00000023  00000-0  28098-4 0  4753";
        const VANGUARD_LINE2: &str =
            "2 00005  34.2682 348.7242 1859667 331.7664  19.3264 10.82419157413667";
        let propagator = Sgp4Propagator::from_tle(VANGUARD_LINE1, VANGUARD_LINE2).unwrap();
        // Vanguard-1 TLE epoch field "00179.78495062" → year 2000,
        // day-of-year 179.78495062. JD(2000-01-01 00:00 UTC) = 2_451_544.5.
        let epoch_jd = 2_451_544.5 + (179.784_950_62 - 1.0);
        // Equatorial station with a permissive 5° horizon mask.
        let station = GroundStation {
            lat_rad: 0.0,
            lon_rad: 0.0,
            height_m: 0.0,
            min_elevation_rad: 5.0_f64.to_radians(),
        };
        let passes = predict_passes(
            &propagator,
            epoch_jd,
            station,
            epoch_jd,
            1.0,    // 24 hours
            30.0,   // 30-second coarse step
        );
        assert!(!passes.is_empty(), "no passes found in 24h");
        for p in &passes {
            assert!(p.aos_jd < p.tca_jd, "AOS not before TCA: {} ≥ {}", p.aos_jd, p.tca_jd);
            assert!(p.tca_jd < p.los_jd, "TCA not before LOS: {} ≥ {}", p.tca_jd, p.los_jd);
            assert!(
                p.max_elevation_rad >= station.min_elevation_rad,
                "TCA elevation below mask"
            );
        }
    }
}
