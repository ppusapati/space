//! Demonstrate the standard Landsat 8 OLI Band 4 calibration chain on a tiny
//! synthetic DN tile.

use eo_radiometric::{
    LinearCalibration, apply_linear, correct_sun_elevation_in_place, dark_object_subtraction,
    earth_sun_distance_au,
};
use ndarray::array;

fn main() {
    let dn = array![[10_000.0_f32, 20_000.0, 30_000.0]];
    let cal = LinearCalibration { gain: 2.0e-5, offset: -0.1 };
    let mut toa = apply_linear(dn.view(), cal).expect("linear calibration");
    correct_sun_elevation_in_place(&mut toa, 60.0_f32.to_radians()).expect("sun-angle correction");
    let boa = dark_object_subtraction(toa.view(), 0.05).expect("dos");
    let d = earth_sun_distance_au(80).expect("earth-sun distance");
    println!("Earth-Sun distance @ DOY 80 = {d:.6} AU");
    println!("TOA reflectance = {toa:?}");
    println!("BOA (DOS)      = {boa:?}");
}
