//! Integration tests for `eo-indices` exercising every public index against
//! synthetic-but-realistic surface reflectance arrays.

use eo_indices::{
    EviCoefficients, EviInput, NdviInput, NdwiInput, SaviInput, compute_evi, compute_ndvi,
    compute_ndwi, compute_savi,
};
use ndarray::array;

#[test]
fn computes_all_indices_for_sentinel2_like_pixels() {
    // Sentinel-2 surface-reflectance approximations for three pixels:
    //   row 0: dense vegetation
    //   row 1: open water
    //   row 2: bare soil
    let blue = array![[0.040_f32, 0.050, 0.140]];
    let green = array![[0.080_f32, 0.090, 0.160]];
    let red = array![[0.060_f32, 0.040, 0.220]];
    let nir = array![[0.480_f32, 0.020, 0.290]];

    let ndvi = compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).unwrap().ndvi;
    assert!(ndvi[(0, 0)] > 0.7, "vegetation NDVI should be high, got {}", ndvi[(0, 0)]);
    assert!(ndvi[(0, 1)] < 0.0, "water NDVI should be negative, got {}", ndvi[(0, 1)]);
    assert!(ndvi[(0, 2)].abs() < 0.3, "bare soil NDVI should be near zero, got {}", ndvi[(0, 2)]);

    let ndwi = compute_ndwi(NdwiInput { green: green.view(), nir: nir.view() }).unwrap().ndwi;
    assert!(ndwi[(0, 1)] > 0.5, "water NDWI should be high, got {}", ndwi[(0, 1)]);
    assert!(ndwi[(0, 0)] < 0.0, "vegetation NDWI should be negative, got {}", ndwi[(0, 0)]);

    let savi = compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: 0.5 }).unwrap().savi;
    assert!(savi[(0, 0)] > 0.5, "vegetation SAVI should be high, got {}", savi[(0, 0)]);

    let evi = compute_evi(EviInput {
        blue: blue.view(),
        red: red.view(),
        nir: nir.view(),
        coefficients: EviCoefficients::default(),
    })
    .unwrap()
    .evi;
    assert!(evi[(0, 0)] > 0.4, "vegetation EVI should be high, got {}", evi[(0, 0)]);
    assert!(evi[(0, 1)] < 0.1, "water EVI should be low, got {}", evi[(0, 1)]);
}

#[test]
fn rejects_shape_mismatch_consistently() {
    let red = ndarray::Array2::<f32>::zeros((10, 10));
    let nir = ndarray::Array2::<f32>::zeros((10, 9));
    assert!(compute_ndvi(NdviInput { red: red.view(), nir: nir.view() }).is_err());
    assert!(compute_savi(SaviInput { red: red.view(), nir: nir.view(), l: 0.5 }).is_err());
    let blue = ndarray::Array2::<f32>::zeros((10, 10));
    assert!(
        compute_evi(EviInput {
            blue: blue.view(),
            red: red.view(),
            nir: nir.view(),
            coefficients: EviCoefficients::default(),
        })
        .is_err()
    );
    let green = ndarray::Array2::<f32>::zeros((10, 9));
    assert!(compute_ndwi(NdwiInput { green: green.view(), nir: nir.view() }).is_ok());
    assert!(compute_ndwi(NdwiInput { green: red.view(), nir: nir.view() }).is_err());
}
