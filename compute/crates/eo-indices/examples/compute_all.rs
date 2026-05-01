//! Compute every supported index on a small synthetic 2-pixel array and print
//! the results. Run with `cargo run --example compute_all -p eo-indices`.

use eo_indices::{
    EviCoefficients, EviInput, NdviInput, NdwiInput, SaviInput, compute_evi, compute_ndvi,
    compute_ndwi, compute_savi,
};
use ndarray::array;

fn main() {
    let blue = array![[0.040_f32, 0.050]];
    let green = array![[0.080_f32, 0.090]];
    let red = array![[0.060_f32, 0.040]];
    let nir = array![[0.480_f32, 0.020]];

    let ndvi = compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).unwrap();
    let ndwi = compute_ndwi(NdwiInput { green: green.view(), nir: nir.view() }).unwrap();
    let savi = compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: 0.5 }).unwrap();
    let evi = compute_evi(EviInput {
        blue: blue.view(),
        red: red.view(),
        nir: nir.view(),
        coefficients: EviCoefficients::default(),
    })
    .unwrap();

    println!("NDVI: {:?}", ndvi.ndvi);
    println!("NDWI: {:?}", ndwi.ndwi);
    println!("SAVI: {:?}", savi.savi);
    println!("EVI:  {:?}", evi.evi);
}
