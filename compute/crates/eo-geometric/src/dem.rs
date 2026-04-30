//! Digital Elevation Model lookup with bilinear interpolation.

use ndarray::Array2;

use crate::GeometricError;

/// A regular geographic-grid DEM. The grid origin is at the upper-left pixel
/// centre `(lat_origin, lon_origin)`; latitudes decrease with row, longitudes
/// increase with column.
#[derive(Debug, Clone)]
pub struct Dem {
    /// Latitude at row 0 (degrees).
    pub lat_origin: f64,
    /// Longitude at column 0 (degrees).
    pub lon_origin: f64,
    /// Latitude step per row (degrees, must be < 0 for typical DEM).
    pub lat_step: f64,
    /// Longitude step per column (degrees, must be > 0).
    pub lon_step: f64,
    /// Elevation in metres, shape `(rows, cols)`.
    pub elevation: Array2<f32>,
}

impl Dem {
    /// Bilinear elevation lookup at `(lat, lon)`. Returns the value as `f64`.
    ///
    /// # Errors
    /// Returns [`GeometricError::OutOfBounds`] if the point falls outside
    /// the DEM extent.
    pub fn elevation_at(&self, lat: f64, lon: f64) -> Result<f64, GeometricError> {
        if !self.lat_step.is_finite() || self.lat_step == 0.0 {
            return Err(GeometricError::OutOfRange {
                name: "lat_step",
                value: self.lat_step,
                range: "non-zero",
            });
        }
        if !self.lon_step.is_finite() || self.lon_step == 0.0 {
            return Err(GeometricError::OutOfRange {
                name: "lon_step",
                value: self.lon_step,
                range: "non-zero",
            });
        }
        let (rows, cols) = self.elevation.dim();
        if rows < 2 || cols < 2 {
            return Err(GeometricError::Empty);
        }
        let row = (lat - self.lat_origin) / self.lat_step;
        let col = (lon - self.lon_origin) / self.lon_step;
        if !(0.0..=(rows as f64 - 1.0)).contains(&row)
            || !(0.0..=(cols as f64 - 1.0)).contains(&col)
        {
            return Err(GeometricError::OutOfBounds { x: col, y: row, rows, cols });
        }
        let r0 = row.floor();
        let c0 = col.floor();
        let dr = row - r0;
        let dc = col - c0;
        let r0i = r0 as usize;
        let c0i = c0 as usize;
        let r1i = (r0i + 1).min(rows - 1);
        let c1i = (c0i + 1).min(cols - 1);
        let v00 = f64::from(self.elevation[(r0i, c0i)]);
        let v10 = f64::from(self.elevation[(r1i, c0i)]);
        let v01 = f64::from(self.elevation[(r0i, c1i)]);
        let v11 = f64::from(self.elevation[(r1i, c1i)]);
        let v0 = v00 * (1.0 - dc) + v01 * dc;
        let v1 = v10 * (1.0 - dc) + v11 * dc;
        Ok(v0 * (1.0 - dr) + v1 * dr)
    }
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    fn ramp_dem() -> Dem {
        // 3×3 DEM with elevations equal to row+col.
        let elevation =
            Array2::from_shape_fn((3, 3), |(i, j)| ((i as i32) + (j as i32)) as f32 * 100.0);
        Dem {
            lat_origin: 10.0,
            lon_origin: 0.0,
            lat_step: -1.0, // row 0 = lat 10, row 1 = lat 9, row 2 = lat 8
            lon_step: 1.0,
            elevation,
        }
    }

    #[test]
    fn elevation_at_pixel_centres() {
        let d = ramp_dem();
        assert_abs_diff_eq!(d.elevation_at(10.0, 0.0).unwrap(), 0.0, epsilon = 1e-9);
        assert_abs_diff_eq!(d.elevation_at(9.0, 1.0).unwrap(), 200.0, epsilon = 1e-9);
        assert_abs_diff_eq!(d.elevation_at(8.0, 2.0).unwrap(), 400.0, epsilon = 1e-9);
    }

    #[test]
    fn elevation_bilinear_midpoint() {
        let d = ramp_dem();
        // Midway between (10,0)=0 and (9,1)=200 should be 100.
        assert_abs_diff_eq!(d.elevation_at(9.5, 0.5).unwrap(), 100.0, epsilon = 1e-9);
    }

    #[test]
    fn out_of_bounds() {
        let d = ramp_dem();
        assert!(d.elevation_at(11.0, 0.0).is_err());
        assert!(d.elevation_at(8.0, -1.0).is_err());
    }
}
