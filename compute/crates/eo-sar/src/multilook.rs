//! Multi-looking: block averaging of intensity to reduce speckle and
//! produce ground-resolution imagery.

use ndarray::{Array2, ArrayView2};
use num_complex::Complex32;

use crate::SarError;

/// Convert complex SLC pixels to intensity (`|z|²`).
#[must_use]
pub fn intensity_from_slc(slc: ArrayView2<'_, Complex32>) -> Array2<f32> {
    let mut out = Array2::<f32>::zeros(slc.dim());
    ndarray::Zip::from(&mut out).and(slc).for_each(|o, z| *o = z.norm_sqr());
    out
}

/// Block-average `intensity` by `looks_az` rows and `looks_rg` columns. The
/// output shape is `(rows / looks_az, cols / looks_rg)`. Trailing rows /
/// columns that don't fit a full block are dropped.
///
/// # Errors
/// Returns [`SarError::OutOfRange`] if either look factor is zero, or
/// [`SarError::Empty`] if the input is empty.
pub fn multilook_intensity(
    intensity: ArrayView2<'_, f32>,
    looks_az: usize,
    looks_rg: usize,
) -> Result<Array2<f32>, SarError> {
    if looks_az == 0 {
        return Err(SarError::OutOfRange {
            name: "looks_az",
            value: 0.0,
            range: "[1, +inf)",
        });
    }
    if looks_rg == 0 {
        return Err(SarError::OutOfRange {
            name: "looks_rg",
            value: 0.0,
            range: "[1, +inf)",
        });
    }
    let (rows, cols) = intensity.dim();
    if rows == 0 || cols == 0 {
        return Err(SarError::Empty);
    }
    let out_rows = rows / looks_az;
    let out_cols = cols / looks_rg;
    if out_rows == 0 || out_cols == 0 {
        return Err(SarError::OutOfRange {
            name: "looks_*",
            value: f64::from(looks_az.max(looks_rg) as u32),
            range: "<= input dim",
        });
    }
    let denom = (looks_az * looks_rg) as f32;
    let mut out = Array2::<f32>::zeros((out_rows, out_cols));
    for ir in 0..out_rows {
        for ic in 0..out_cols {
            let mut sum = 0.0_f32;
            for dr in 0..looks_az {
                for dc in 0..looks_rg {
                    sum += intensity[(ir * looks_az + dr, ic * looks_rg + dc)];
                }
            }
            out[(ir, ic)] = sum / denom;
        }
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;
    use num_complex::Complex32;

    use super::*;

    #[test]
    fn intensity_from_slc_correct() {
        let slc = array![[Complex32::new(3.0, 4.0)]]; // |z|^2 = 25
        let i = intensity_from_slc(slc.view());
        assert_abs_diff_eq!(i[(0, 0)], 25.0, epsilon = 1e-6);
    }

    #[test]
    fn multilook_2x2_averages() {
        let a = array![[1.0_f32, 2.0, 3.0, 4.0], [5.0, 6.0, 7.0, 8.0]];
        let m = multilook_intensity(a.view(), 2, 2).unwrap();
        // Block 1: (1+2+5+6)/4 = 3.5; block 2: (3+4+7+8)/4 = 5.5
        assert_eq!(m.dim(), (1, 2));
        assert_abs_diff_eq!(m[(0, 0)], 3.5, epsilon = 1e-6);
        assert_abs_diff_eq!(m[(0, 1)], 5.5, epsilon = 1e-6);
    }

    #[test]
    fn multilook_zero_rejected() {
        let a = array![[1.0_f32]];
        assert!(multilook_intensity(a.view(), 0, 1).is_err());
        assert!(multilook_intensity(a.view(), 1, 0).is_err());
    }
}
