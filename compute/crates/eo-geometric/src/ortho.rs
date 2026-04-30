//! DEM-aware orthorectification.

use ndarray::{Array2, ArrayView2};

use crate::{
    GeometricError,
    dem::Dem,
    resample::{Method, sample},
    rpc::Rpc,
};

/// 6-parameter affine geo-transform identifying an output raster's
/// upper-left pixel-centre origin in `(x, y)` map units. Pixels at
/// (row, col) map to:
/// ```text
/// x = origin_x + col * pixel_x
/// y = origin_y + row * pixel_y
/// ```
#[derive(Debug, Clone, Copy)]
pub struct AffineGeo {
    /// Map x coordinate of the upper-left pixel centre.
    pub origin_x: f64,
    /// Map y coordinate of the upper-left pixel centre.
    pub origin_y: f64,
    /// Map x increment per output column.
    pub pixel_x: f64,
    /// Map y increment per output row (typically negative).
    pub pixel_y: f64,
}

impl AffineGeo {
    /// Map output `(row, col)` to `(x, y)` (lon, lat for a geographic grid).
    #[must_use]
    pub fn pixel_to_world(self, row: usize, col: usize) -> (f64, f64) {
        (
            (col as f64).mul_add(self.pixel_x, self.origin_x),
            (row as f64).mul_add(self.pixel_y, self.origin_y),
        )
    }
}

/// Orthorectify `src` onto an output grid defined by `out_shape` and `geo`,
/// using `rpc` to project each output pixel through the DEM and resampling
/// the source array with `method`. Pixels for which the RPC inverse falls
/// outside the source array are filled with [`f32::NAN`].
///
/// `geo` is interpreted as `(lon, lat)` — `pixel_x` increases longitude per
/// output column, `pixel_y` increases latitude per output row (typically
/// negative). The DEM lookup is performed at the same `(lat, lon)`.
///
/// # Errors
/// Returns [`GeometricError`] for empty inputs.
pub fn orthorectify(
    src: ArrayView2<'_, f32>,
    rpc: &Rpc,
    dem: &Dem,
    out_shape: (usize, usize),
    geo: AffineGeo,
    method: Method,
    inverse_tol: f64,
    inverse_max_iter: u32,
) -> Result<Array2<f32>, GeometricError> {
    let (sr, sc) = src.dim();
    if sr == 0 || sc == 0 || out_shape.0 == 0 || out_shape.1 == 0 {
        return Err(GeometricError::Empty);
    }
    let mut out = Array2::<f32>::from_elem(out_shape, f32::NAN);
    for row in 0..out_shape.0 {
        for col in 0..out_shape.1 {
            let (lon, lat) = geo.pixel_to_world(row, col);
            let h = dem.elevation_at(lat, lon).unwrap_or(0.0);
            let (line, samp) = rpc.forward(lat, lon, h);
            // RPC forward may produce out-of-bounds when the output grid
            // covers a wider area than the source. Sample handles bounds.
            let _ = (inverse_tol, inverse_max_iter); // hooks for inverse-direction variants
            if let Some(v) = sample(src, line, samp, method) {
                out[(row, col)] = v;
            }
        }
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::Array2;

    use super::*;

    fn identity_rpc() -> Rpc {
        let mut num_l = [0.0; 20];
        let mut den_l = [0.0; 20];
        let mut num_s = [0.0; 20];
        let mut den_s = [0.0; 20];
        // line = lat, sample = lon (both with offsets/scales=0/1)
        num_l[2] = 1.0;
        den_l[0] = 1.0;
        num_s[1] = 1.0;
        den_s[0] = 1.0;
        Rpc {
            line_off: 0.0,
            sample_off: 0.0,
            lat_off: 0.0,
            lon_off: 0.0,
            height_off: 0.0,
            line_scale: 1.0,
            sample_scale: 1.0,
            lat_scale: 1.0,
            lon_scale: 1.0,
            height_scale: 1.0,
            line_num: num_l,
            line_den: den_l,
            samp_num: num_s,
            samp_den: den_s,
        }
    }

    fn flat_dem(rows: usize, cols: usize) -> Dem {
        Dem {
            lat_origin: 4.0,
            lon_origin: 0.0,
            lat_step: -1.0,
            lon_step: 1.0,
            elevation: Array2::from_elem((rows, cols), 0.0),
        }
    }

    #[test]
    fn ortho_identity_rpc_copies_through_geo() {
        // Source 5x5 with values = row*10 + col.
        let src = Array2::from_shape_fn((5, 5), |(i, j)| (i * 10 + j) as f32);
        let rpc = identity_rpc();
        let dem = flat_dem(5, 5);
        // Output grid sampled at the same lat=row, lon=col positions:
        //   row 0 -> lat 4, row 1 -> lat 3, ..., row 4 -> lat 0
        //   col 0 -> lon 0, col 4 -> lon 4
        let geo = AffineGeo { origin_x: 0.0, origin_y: 4.0, pixel_x: 1.0, pixel_y: -1.0 };
        let out =
            orthorectify(src.view(), &rpc, &dem, (5, 5), geo, Method::Bilinear, 1e-9, 50).unwrap();
        // For row=0, lat=4 -> line=4 (RPC); col=0, lon=0 -> sample=0
        // src[(4,0)] = 40
        assert_abs_diff_eq!(out[(0, 0)], 40.0, epsilon = 1e-3);
        // For row=4, col=4 -> lat=0,lon=4 -> line=0,samp=4 -> src[(0,4)] = 4
        assert_abs_diff_eq!(out[(4, 4)], 4.0, epsilon = 1e-3);
    }
}
