//! End-to-end calibration pipeline using realistic Landsat 8 OLI Band 4 (Red)
//! collection-2 coefficients.

use eo_radiometric::{
    LinearCalibration, apply_linear, correct_sun_elevation_in_place,
    radiance_to_brightness_temperature, radiance_to_toa_reflectance, sentinel2_dn_to_toa,
};
use ndarray::array;

#[test]
fn landsat8_band4_reflectance_pipeline() {
    // Landsat 8 OLI Band 4 (Red) representative Collection-2 coefficients:
    //   reflectance: M_ρ = 2.0e-5, A_ρ = -0.1
    //   sun elevation = 60°
    let dn = array![[10_000.0_f32, 30_000.0]];
    let cal = LinearCalibration { gain: 2.0e-5, offset: -0.1 };
    let mut rho = apply_linear(dn.view(), cal).unwrap();
    // ρ' values: 0.1 and 0.5
    correct_sun_elevation_in_place(&mut rho, 60.0_f32.to_radians()).unwrap();
    let s = 60.0_f32.to_radians().sin();
    assert!((rho[(0, 0)] - 0.1 / s).abs() < 1e-5);
    assert!((rho[(0, 1)] - 0.5 / s).abs() < 1e-5);
    // Reasonable physical bounds.
    for &v in &rho {
        assert!((0.0..=1.5).contains(&v), "reflectance out of bounds: {v}");
    }
}

#[test]
fn landsat5_thermal_pipeline() {
    // Radiance: L = 0.05518 * DN + 1.2378 (Landsat 5 TM B6 typical)
    // Brightness temperature with K1=607.76, K2=1260.56
    let dn = array![[150.0_f32, 200.0]];
    let cal = LinearCalibration { gain: 0.055_18, offset: 1.2378 };
    let radiance = apply_linear(dn.view(), cal).unwrap();
    let bt = radiance_to_brightness_temperature(radiance.view(), 607.76, 1260.56).unwrap();
    for &t in &bt {
        assert!((250.0..=350.0).contains(&t), "BT outside Earth temperature range: {t}");
    }
}

#[test]
fn sentinel2_l1c_to_toa_then_radiance_to_toa_consistent() {
    // S2 L1C TOA via DN/Q:
    let dn = array![[1234.0_f32, 9876.0]];
    let toa = sentinel2_dn_to_toa(dn.view(), 10_000.0).unwrap();
    assert!((toa[(0, 0)] - 0.1234).abs() < 1e-6);
    assert!((toa[(0, 1)] - 0.9876).abs() < 1e-6);
}

#[test]
fn radiance_to_toa_zero_input_yields_zero() {
    let r = ndarray::Array2::<f32>::zeros((4, 4));
    let toa = radiance_to_toa_reflectance(r.view(), 1969.0, 1.0, 0.5).unwrap();
    for &v in &toa {
        assert!(v.abs() < 1e-12, "expected zero, got {v}");
    }
}
