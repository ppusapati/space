//! Blending compositors.
//!
//! Each compositor takes views over two registered rasters of identical
//! shape and writes a single output. Pixels carrying NaN in one input adopt
//! the other input's value; pixels NaN in both yield NaN.

use ndarray::{Array2, ArrayView2};

/// Average compositor: `out = (a + b) / 2` with NaN-aware fallback.
#[must_use]
pub fn average(a: ArrayView2<'_, f32>, b: ArrayView2<'_, f32>) -> Array2<f32> {
    let mut out = Array2::<f32>::from_elem(a.dim(), f32::NAN);
    ndarray::Zip::from(&mut out).and(a).and(b).for_each(|o, &x, &y| {
        *o = match (x.is_finite(), y.is_finite()) {
            (true, true) => 0.5 * (x + y),
            (true, false) => x,
            (false, true) => y,
            (false, false) => f32::NAN,
        };
    });
    out
}

/// Maximum-value compositor.
#[must_use]
pub fn maximum(a: ArrayView2<'_, f32>, b: ArrayView2<'_, f32>) -> Array2<f32> {
    let mut out = Array2::<f32>::from_elem(a.dim(), f32::NAN);
    ndarray::Zip::from(&mut out).and(a).and(b).for_each(|o, &x, &y| {
        *o = match (x.is_finite(), y.is_finite()) {
            (true, true) => x.max(y),
            (true, false) => x,
            (false, true) => y,
            (false, false) => f32::NAN,
        };
    });
    out
}

/// Feather blend across a vertical seamline `seamline_col[row]`. Pixels to
/// the left of the seam favour `a`, pixels to the right favour `b`, with a
/// linear ramp of half-width `feather` over the transition. NaNs follow the
/// rules of [`average`].
///
/// `seamline_col.len()` must equal the number of rows.
#[must_use]
pub fn feather_vertical(
    a: ArrayView2<'_, f32>,
    b: ArrayView2<'_, f32>,
    seamline_col: &[usize],
    feather: usize,
) -> Array2<f32> {
    let (rows, cols) = a.dim();
    debug_assert_eq!(seamline_col.len(), rows);
    let mut out = Array2::<f32>::from_elem(a.dim(), f32::NAN);
    let f = feather.max(1) as f32;
    for r in 0..rows {
        let s = seamline_col[r] as i64;
        for c in 0..cols {
            let dist = c as i64 - s;
            let w = if dist < -(feather as i64) {
                1.0_f32 // fully a
            } else if dist > feather as i64 {
                0.0_f32 // fully b
            } else {
                0.5 - 0.5 * (dist as f32 / f)
            };
            let av = a[(r, c)];
            let bv = b[(r, c)];
            out[(r, c)] = match (av.is_finite(), bv.is_finite()) {
                (true, true) => w * av + (1.0 - w) * bv,
                (true, false) => av,
                (false, true) => bv,
                (false, false) => f32::NAN,
            };
        }
    }
    out
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn average_handles_nan() {
        let a = array![[1.0_f32, f32::NAN], [3.0, 4.0]];
        let b = array![[2.0_f32, 8.0], [f32::NAN, 4.0]];
        let out = average(a.view(), b.view());
        assert_abs_diff_eq!(out[(0, 0)], 1.5);
        assert_abs_diff_eq!(out[(0, 1)], 8.0);
        assert_abs_diff_eq!(out[(1, 0)], 3.0);
        assert_abs_diff_eq!(out[(1, 1)], 4.0);
    }

    #[test]
    fn maximum_handles_nan() {
        let a = array![[1.0_f32, f32::NAN]];
        let b = array![[2.0_f32, 8.0]];
        let out = maximum(a.view(), b.view());
        assert_abs_diff_eq!(out[(0, 0)], 2.0);
        assert_abs_diff_eq!(out[(0, 1)], 8.0);
    }

    #[test]
    fn feather_left_right_extremes() {
        let a = ndarray::Array2::<f32>::from_elem((1, 11), 0.0);
        let b = ndarray::Array2::<f32>::from_elem((1, 11), 1.0);
        let out = feather_vertical(a.view(), b.view(), &[5], 2);
        // Far left -> a (0.0), far right -> b (1.0)
        assert_abs_diff_eq!(out[(0, 0)], 0.0);
        assert_abs_diff_eq!(out[(0, 10)], 1.0);
        // At seam: average
        assert_abs_diff_eq!(out[(0, 5)], 0.5);
    }
}
