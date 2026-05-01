//! Pairwise tile overlap computation.

use crate::{AffineGeo, MosaicError, Tile};

/// Overlap rectangle in both tiles' pixel coordinates and in world units.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Overlap {
    /// Overlap rows×cols.
    pub shape: (usize, usize),
    /// Top-left pixel of the overlap inside tile A `(row, col)`.
    pub a_origin: (usize, usize),
    /// Top-left pixel of the overlap inside tile B `(row, col)`.
    pub b_origin: (usize, usize),
    /// World-space upper-left corner `(x, y)`.
    pub world_min: (f64, f64),
    /// World-space lower-right corner `(x, y)`.
    pub world_max: (f64, f64),
}

/// World-space bounding box of a tile.
#[must_use]
pub fn tile_bbox(t: &Tile) -> ((f64, f64), (f64, f64)) {
    let (rows, cols) = t.raster.dim();
    let x0 = t.geo.origin_x;
    let y0 = t.geo.origin_y;
    let x1 = t.geo.origin_x + (cols as f64) * t.geo.pixel_x;
    let y1 = t.geo.origin_y + (rows as f64) * t.geo.pixel_y;
    let xmin = x0.min(x1);
    let xmax = x0.max(x1);
    let ymin = y0.min(y1);
    let ymax = y0.max(y1);
    ((xmin, ymin), (xmax, ymax))
}

/// Compute the pairwise overlap of two tiles in pixel coordinates.
///
/// # Errors
/// Returns [`MosaicError::PixelSizeMismatch`] if pixel sizes disagree (within
/// 1·10⁻⁹) or [`MosaicError::NoOverlap`] if the world bounding boxes are
/// disjoint.
pub fn pairwise_overlap(a: &Tile, b: &Tile) -> Result<Overlap, MosaicError> {
    let dx_eq = (a.geo.pixel_x - b.geo.pixel_x).abs() < 1e-9;
    let dy_eq = (a.geo.pixel_y - b.geo.pixel_y).abs() < 1e-9;
    if !(dx_eq && dy_eq) {
        return Err(MosaicError::PixelSizeMismatch {
            a: (a.geo.pixel_x, a.geo.pixel_y),
            b: (b.geo.pixel_x, b.geo.pixel_y),
        });
    }
    let ((axmin, aymin), (axmax, aymax)) = tile_bbox(a);
    let ((bxmin, bymin), (bxmax, bymax)) = tile_bbox(b);
    let xmin = axmin.max(bxmin);
    let ymin = aymin.max(bymin);
    let xmax = axmax.min(bxmax);
    let ymax = aymax.min(bymax);
    if xmax <= xmin || ymax <= ymin {
        return Err(MosaicError::NoOverlap);
    }

    // Convert world bbox back to pixel offsets in each tile, accounting for
    // the sign of pixel_y.
    let (a_orig, b_orig, shape) = pixel_origin_and_shape(a, b, xmin, ymin, xmax, ymax);
    Ok(Overlap {
        shape,
        a_origin: a_orig,
        b_origin: b_orig,
        world_min: (xmin, ymin),
        world_max: (xmax, ymax),
    })
}

fn pixel_origin_and_shape(
    a: &Tile,
    b: &Tile,
    xmin: f64,
    ymin: f64,
    xmax: f64,
    ymax: f64,
) -> ((usize, usize), (usize, usize), (usize, usize)) {
    fn world_to_pixel(geo: AffineGeo, x: f64, y: f64) -> (f64, f64) {
        let col = (x - geo.origin_x) / geo.pixel_x;
        let row = (y - geo.origin_y) / geo.pixel_y;
        (row, col)
    }
    // Use the upper-left in *world* coordinates: that's (xmin, ymax) since
    // y typically decreases with row index when pixel_y < 0.
    let (ar0, ac0) = world_to_pixel(a.geo, xmin, ymax);
    let (br0, bc0) = world_to_pixel(b.geo, xmin, ymax);
    let cols = ((xmax - xmin) / a.geo.pixel_x.abs()).round() as usize;
    let rows = ((ymax - ymin) / a.geo.pixel_y.abs()).round() as usize;
    (
        (ar0.round().max(0.0) as usize, ac0.round().max(0.0) as usize),
        (br0.round().max(0.0) as usize, bc0.round().max(0.0) as usize),
        (rows, cols),
    )
}

#[cfg(test)]
mod tests {
    use ndarray::Array2;

    use super::*;

    fn tile(rows: usize, cols: usize, ox: f64, oy: f64) -> Tile {
        Tile {
            raster: Array2::<f32>::zeros((rows, cols)),
            geo: AffineGeo { origin_x: ox, origin_y: oy, pixel_x: 1.0, pixel_y: -1.0 },
        }
    }

    #[test]
    fn axis_aligned_overlap() {
        let a = tile(10, 10, 0.0, 10.0); // covers x in [0,10], y in [0,10]
        let b = tile(10, 10, 5.0, 12.0); // covers x in [5,15], y in [2,12]
        let o = pairwise_overlap(&a, &b).unwrap();
        assert_eq!(o.shape, (8, 5));
    }

    #[test]
    fn no_overlap() {
        let a = tile(5, 5, 0.0, 5.0);
        let b = tile(5, 5, 100.0, 100.0);
        assert!(matches!(pairwise_overlap(&a, &b).unwrap_err(), MosaicError::NoOverlap));
    }

    #[test]
    fn mismatched_pixel_size() {
        let a = tile(5, 5, 0.0, 5.0);
        let mut b = tile(5, 5, 1.0, 5.0);
        b.geo.pixel_x = 2.0;
        assert!(matches!(
            pairwise_overlap(&a, &b).unwrap_err(),
            MosaicError::PixelSizeMismatch { .. }
        ));
    }
}
