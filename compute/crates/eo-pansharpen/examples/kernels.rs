//! Demonstrate every pan-sharpening kernel on a synthetic 3-band MS cube
//! plus a synthetic Pan band.

use eo_pansharpen::{GsWeights, brovey, gram_schmidt, ihs, pca};
use ndarray::{Array2, Array3};

fn main() {
    let ms = Array3::from_shape_fn((3, 4, 4), |(b, i, j)| {
        0.1 + (b as f32) * 0.05 + (i as f32) * 0.02 + (j as f32) * 0.01
    });
    let pan = Array2::from_shape_fn((4, 4), |(i, j)| 0.3 + (i as f32) * 0.025 + (j as f32) * 0.012);

    let b = brovey(ms.view(), pan.view()).unwrap();
    let h = ihs(ms.view(), pan.view()).unwrap();
    let p = pca(ms.view(), pan.view()).unwrap();
    let g = gram_schmidt(ms.view(), pan.view(), &GsWeights::equal(3)).unwrap();

    println!("Brovey output band 0:\n{:?}", b.index_axis(ndarray::Axis(0), 0));
    println!("IHS    output band 0:\n{:?}", h.index_axis(ndarray::Axis(0), 0));
    println!("PCA    output band 0:\n{:?}", p.index_axis(ndarray::Axis(0), 0));
    println!("GS     output band 0:\n{:?}", g.index_axis(ndarray::Axis(0), 0));
}
