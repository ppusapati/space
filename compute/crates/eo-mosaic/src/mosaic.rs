//! Multi-tile mosaic driver.

use ndarray::{Array2, ArrayView2};

use crate::{
    AffineGeo, MosaicError, Tile, blend,
    overlap::{Overlap, pairwise_overlap},
    seamline::{Direction, min_cost_path},
};

/// Compositing rule used by [`mosaic`].
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Compositor {
    /// Average overlapping pixels.
    Average,
    /// Take the maximum value at each overlapping pixel.
    Maximum,
    /// Build a minimum-cost seamline through the overlap and feather across it.
    Seamline {
        /// Half-width (pixels) of the feather ramp.
        feather: usize,
    },
}

/// Mosaic a list of tiles onto an output grid defined by `out_geo` with shape
/// `out_shape`. Tiles outside the output extent contribute nothing. Pixels
/// not covered by any tile carry `NaN`.
///
/// All tiles must share the same `pixel_x` and `pixel_y` as `out_geo`.
///
/// # Errors
/// [`MosaicError::NoTiles`] for an empty list,
/// [`MosaicError::PixelSizeMismatch`] for inconsistent pixel sizes,
/// [`MosaicError::EmptyOutput`] for a zero-shape output.
pub fn mosaic(
    tiles: &[Tile],
    out_geo: AffineGeo,
    out_shape: (usize, usize),
    compositor: Compositor,
) -> Result<Array2<f32>, MosaicError> {
    if tiles.is_empty() {
        return Err(MosaicError::NoTiles);
    }
    if out_shape.0 == 0 || out_shape.1 == 0 {
        return Err(MosaicError::EmptyOutput);
    }
    for t in tiles {
        if (t.geo.pixel_x - out_geo.pixel_x).abs() > 1e-9
            || (t.geo.pixel_y - out_geo.pixel_y).abs() > 1e-9
        {
            return Err(MosaicError::PixelSizeMismatch {
                a: (t.geo.pixel_x, t.geo.pixel_y),
                b: (out_geo.pixel_x, out_geo.pixel_y),
            });
        }
    }
    // Output canvas filled with NaN.
    let mut canvas = Array2::<f32>::from_elem(out_shape, f32::NAN);

    // Place each tile in the output canvas; combine with the existing canvas
    // pixel value via the compositor when both are valid.
    for t in tiles {
        place_tile(&mut canvas, t, out_geo, compositor);
    }
    Ok(canvas)
}

fn place_tile(canvas: &mut Array2<f32>, tile: &Tile, geo: AffineGeo, compositor: Compositor) {
    let (rows, cols) = canvas.dim();
    let (trows, tcols) = tile.raster.dim();
    // Top-left of the tile in canvas pixel coordinates.
    let row0 = ((tile.geo.origin_y - geo.origin_y) / geo.pixel_y).round() as i64;
    let col0 = ((tile.geo.origin_x - geo.origin_x) / geo.pixel_x).round() as i64;
    let row_start = row0.max(0) as usize;
    let col_start = col0.max(0) as usize;
    let row_end = ((row0 + trows as i64).min(rows as i64)).max(0) as usize;
    let col_end = ((col0 + tcols as i64).min(cols as i64)).max(0) as usize;
    if row_start >= row_end || col_start >= col_end {
        return;
    }
    // For Seamline compositor, we need the existing canvas region to
    // compute a cost map; build views.
    let canvas_view = canvas
        .slice(ndarray::s![row_start..row_end, col_start..col_end])
        .to_owned();
    let tile_offset_row = (row_start as i64 - row0) as usize;
    let tile_offset_col = (col_start as i64 - col0) as usize;
    let tile_view = tile
        .raster
        .slice(ndarray::s![
            tile_offset_row..tile_offset_row + (row_end - row_start),
            tile_offset_col..tile_offset_col + (col_end - col_start)
        ])
        .to_owned();

    let blended = match compositor {
        Compositor::Average => blend::average(canvas_view.view(), tile_view.view()),
        Compositor::Maximum => blend::maximum(canvas_view.view(), tile_view.view()),
        Compositor::Seamline { feather } => {
            seamline_blend(canvas_view.view(), tile_view.view(), feather)
        }
    };
    canvas
        .slice_mut(ndarray::s![row_start..row_end, col_start..col_end])
        .assign(&blended);
}

fn seamline_blend(
    a: ArrayView2<'_, f32>,
    b: ArrayView2<'_, f32>,
    feather: usize,
) -> Array2<f32> {
    let (rows, _cols) = a.dim();
    // Cost = absolute difference; NaN cells get a high cost so the seam
    // routes around them.
    let cost = ndarray::Zip::from(a).and(b).map_collect(|&x, &y| {
        if x.is_finite() && y.is_finite() { (x - y).abs() } else { f32::MAX / 4.0 }
    });
    let path = min_cost_path(cost.view(), Direction::TopToBottom);
    if path.is_empty() {
        return blend::average(a, b);
    }
    let mut col_per_row = vec![0_usize; rows];
    for (r, c) in path {
        col_per_row[r] = c;
    }
    blend::feather_vertical(a, b, &col_per_row, feather)
}

/// Compute the pairwise overlap and (optionally) blend two registered tiles
/// onto a third tile that exactly covers their union.
///
/// # Errors
/// Bubbles up [`pairwise_overlap`] errors.
pub fn blend_pair(
    a: &Tile,
    b: &Tile,
    compositor: Compositor,
) -> Result<(Tile, Overlap), MosaicError> {
    let ov = pairwise_overlap(a, b)?;
    let geo = AffineGeo {
        origin_x: ov.world_min.0,
        origin_y: ov.world_max.1,
        pixel_x: a.geo.pixel_x,
        pixel_y: a.geo.pixel_y,
    };
    let tiles = [a.clone(), b.clone()];
    let out = mosaic(&tiles, geo, ov.shape, compositor)?;
    Ok((Tile { raster: out, geo }, ov))
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::Array2;

    use super::*;

    fn tile(rows: usize, cols: usize, ox: f64, oy: f64, fill: f32) -> Tile {
        Tile {
            raster: Array2::<f32>::from_elem((rows, cols), fill),
            geo: AffineGeo { origin_x: ox, origin_y: oy, pixel_x: 1.0, pixel_y: -1.0 },
        }
    }

    #[test]
    fn mosaic_placed_tile_appears() {
        let a = tile(4, 4, 0.0, 4.0, 7.0);
        let geo = AffineGeo { origin_x: 0.0, origin_y: 4.0, pixel_x: 1.0, pixel_y: -1.0 };
        let m = mosaic(&[a], geo, (4, 4), Compositor::Average).unwrap();
        for v in &m {
            assert_abs_diff_eq!(*v, 7.0, epsilon = 1e-6);
        }
    }

    #[test]
    fn mosaic_two_tiles_average_in_overlap() {
        let a = tile(4, 4, 0.0, 4.0, 1.0);
        let b = tile(4, 4, 2.0, 4.0, 3.0);
        // Output covers x in [0,6], y in [0,4]: 4 rows × 6 cols.
        let geo = AffineGeo { origin_x: 0.0, origin_y: 4.0, pixel_x: 1.0, pixel_y: -1.0 };
        let m = mosaic(&[a, b], geo, (4, 6), Compositor::Average).unwrap();
        // Cols 0..2 only A -> 1; cols 4..6 only B -> 3; cols 2..4 average -> 2
        for r in 0..4 {
            for c in 0..2 {
                assert_abs_diff_eq!(m[(r, c)], 1.0, epsilon = 1e-6);
            }
            for c in 2..4 {
                assert_abs_diff_eq!(m[(r, c)], 2.0, epsilon = 1e-6);
            }
            for c in 4..6 {
                assert_abs_diff_eq!(m[(r, c)], 3.0, epsilon = 1e-6);
            }
        }
    }

    #[test]
    fn mosaic_max_compositor() {
        let a = tile(2, 2, 0.0, 2.0, 1.0);
        let b = tile(2, 2, 0.0, 2.0, 3.0);
        let geo = AffineGeo { origin_x: 0.0, origin_y: 2.0, pixel_x: 1.0, pixel_y: -1.0 };
        let m = mosaic(&[a, b], geo, (2, 2), Compositor::Maximum).unwrap();
        for v in &m {
            assert_abs_diff_eq!(*v, 3.0, epsilon = 1e-6);
        }
    }

    #[test]
    fn mosaic_uncovered_pixels_are_nan() {
        // Tile occupies the upper-left 2×2 of the 4×4 canvas. With output
        // origin_y = 4 and pixel_y = -1, row 0 corresponds to y = 4, so the
        // tile must also have origin_y = 4 to align at row 0.
        let a = tile(2, 2, 0.0, 4.0, 1.0);
        let geo = AffineGeo { origin_x: 0.0, origin_y: 4.0, pixel_x: 1.0, pixel_y: -1.0 };
        let m = mosaic(&[a], geo, (4, 4), Compositor::Average).unwrap();
        for r in 0..4 {
            for c in 0..4 {
                let v = m[(r, c)];
                if r < 2 && c < 2 {
                    assert_abs_diff_eq!(v, 1.0, epsilon = 1e-6);
                } else {
                    assert!(v.is_nan(), "expected NaN at ({r},{c}), got {v}");
                }
            }
        }
    }
}
