//! In-image resampling kernels.
//!
//! All kernels take a 2-D array and a real-valued `(line, sample)` location
//! and return the resampled value or `None` if the point is outside the
//! source array.

use ndarray::ArrayView2;

/// Resampling method.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Method {
    /// Nearest-neighbour.
    Nearest,
    /// Bilinear.
    Bilinear,
    /// Cubic-convolution (Keys, a = -0.5).
    Cubic,
}

/// Sample `src` at fractional `(line, sample)` using `method`.
#[must_use]
pub fn sample(src: ArrayView2<'_, f32>, line: f64, sample: f64, method: Method) -> Option<f32> {
    let (rows, cols) = src.dim();
    if rows == 0 || cols == 0 {
        return None;
    }
    if !(line.is_finite() && sample.is_finite()) {
        return None;
    }
    if line < 0.0 || sample < 0.0 || line > rows as f64 - 1.0 || sample > cols as f64 - 1.0 {
        return None;
    }
    Some(match method {
        Method::Nearest => nearest(src, line, sample),
        Method::Bilinear => bilinear(src, line, sample),
        Method::Cubic => cubic(src, line, sample),
    })
}

fn nearest(src: ArrayView2<'_, f32>, line: f64, sample: f64) -> f32 {
    let r = line.round() as usize;
    let c = sample.round() as usize;
    src[(r.min(src.dim().0 - 1), c.min(src.dim().1 - 1))]
}

fn bilinear(src: ArrayView2<'_, f32>, line: f64, sample: f64) -> f32 {
    let (rows, cols) = src.dim();
    let r0 = line.floor() as usize;
    let c0 = sample.floor() as usize;
    let r1 = (r0 + 1).min(rows - 1);
    let c1 = (c0 + 1).min(cols - 1);
    let dr = (line - r0 as f64) as f32;
    let dc = (sample - c0 as f64) as f32;
    let v00 = src[(r0, c0)];
    let v01 = src[(r0, c1)];
    let v10 = src[(r1, c0)];
    let v11 = src[(r1, c1)];
    let v0 = v00 * (1.0 - dc) + v01 * dc;
    let v1 = v10 * (1.0 - dc) + v11 * dc;
    v0 * (1.0 - dr) + v1 * dr
}

#[inline]
fn cubic_kernel(t: f32) -> f32 {
    let a: f32 = -0.5;
    let t = t.abs();
    if t < 1.0 {
        ((a + 2.0) * t - (a + 3.0)) * t * t + 1.0
    } else if t < 2.0 {
        ((a * t - 5.0 * a) * t + 8.0 * a) * t - 4.0 * a
    } else {
        0.0
    }
}

fn cubic(src: ArrayView2<'_, f32>, line: f64, sample: f64) -> f32 {
    let (rows, cols) = src.dim();
    let r = line.floor() as i64;
    let c = sample.floor() as i64;
    let dr = (line - r as f64) as f32;
    let dc = (sample - c as f64) as f32;
    let mut acc = 0.0_f32;
    for di in -1..=2_i64 {
        let ri = (r + di).clamp(0, rows as i64 - 1) as usize;
        let wr = cubic_kernel(di as f32 - dr);
        for dj in -1..=2_i64 {
            let cj = (c + dj).clamp(0, cols as i64 - 1) as usize;
            let wc = cubic_kernel(dj as f32 - dc);
            acc += wr * wc * src[(ri, cj)];
        }
    }
    acc
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn nearest_picks_closest() {
        let a = array![[1.0_f32, 2.0], [3.0, 4.0]];
        assert_abs_diff_eq!(sample(a.view(), 0.49, 0.49, Method::Nearest).unwrap(), 1.0);
        assert_abs_diff_eq!(sample(a.view(), 0.51, 0.51, Method::Nearest).unwrap(), 4.0);
    }

    #[test]
    fn bilinear_centre() {
        let a = array![[0.0_f32, 1.0], [2.0, 3.0]];
        assert_abs_diff_eq!(sample(a.view(), 0.5, 0.5, Method::Bilinear).unwrap(), 1.5);
    }

    #[test]
    fn cubic_reproduces_constant() {
        let a = ndarray::Array2::<f32>::from_elem((6, 6), 7.0);
        assert_abs_diff_eq!(
            sample(a.view(), 2.3, 3.7, Method::Cubic).unwrap(),
            7.0,
            epsilon = 1e-5
        );
    }

    #[test]
    fn cubic_reproduces_linear_ramp() {
        let a = ndarray::Array2::from_shape_fn((6, 6), |(i, j)| (i + j) as f32);
        // Cubic-convolution is exact for linear functions.
        assert_abs_diff_eq!(sample(a.view(), 2.3, 3.7, Method::Cubic).unwrap(), 6.0, epsilon = 1e-4);
    }

    #[test]
    fn out_of_bounds_returns_none() {
        let a = array![[1.0_f32, 2.0], [3.0, 4.0]];
        assert!(sample(a.view(), -0.1, 0.5, Method::Bilinear).is_none());
        assert!(sample(a.view(), 1.5, 0.5, Method::Bilinear).is_none());
    }
}
