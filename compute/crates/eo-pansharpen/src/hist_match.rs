//! Linear histogram matching: rescale `source` so its mean and standard
//! deviation match those of `reference`. Used to align the panchromatic band
//! to the synthetic intensity / first principal component before substitution.

use ndarray::ArrayView2;

use crate::PansharpenError;

/// Returns a new array containing `source` rescaled to the mean and standard
/// deviation of `reference`:
///
/// `out = (source − μ_s) · (σ_r / σ_s) + μ_r`
///
/// If `source` has zero variance the function returns a constant array equal
/// to `μ_r`.
///
/// # Errors
/// [`PansharpenError::ShapeMismatch`] if shapes differ; [`PansharpenError::Empty`]
/// if either input is empty.
pub fn histogram_match(
    source: ArrayView2<'_, f32>,
    reference: ArrayView2<'_, f32>,
) -> Result<ndarray::Array2<f32>, PansharpenError> {
    let dim = source.dim();
    if dim.0 == 0 || dim.1 == 0 {
        return Err(PansharpenError::Empty);
    }
    if dim != reference.dim() {
        return Err(PansharpenError::ShapeMismatch { expected: dim, got: reference.dim() });
    }
    let (mu_s, sigma_s) = mean_std(source);
    let (mu_r, sigma_r) = mean_std(reference);
    let mut out = ndarray::Array2::<f32>::zeros(dim);
    if sigma_s == 0.0 {
        out.fill(mu_r);
        return Ok(out);
    }
    let scale = sigma_r / sigma_s;
    ndarray::Zip::from(&mut out).and(source).for_each(|o, &v| {
        *o = (v - mu_s) * scale + mu_r;
    });
    Ok(out)
}

/// Compute the mean and (uncorrected) standard deviation of an array.
pub(crate) fn mean_std(a: ArrayView2<'_, f32>) -> (f32, f32) {
    let n = (a.dim().0 * a.dim().1) as f32;
    if n == 0.0 {
        return (0.0, 0.0);
    }
    let mean = a.iter().copied().sum::<f32>() / n;
    let var = a.iter().copied().map(|v| (v - mean) * (v - mean)).sum::<f32>() / n;
    (mean, var.sqrt())
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;
    use ndarray::array;

    use super::*;

    #[test]
    fn matches_first_two_moments() {
        let src = array![[0.0_f32, 1.0], [2.0, 3.0]]; // mean=1.5, std=sqrt(1.25)
        let r = array![[10.0_f32, 20.0], [30.0, 40.0]]; // mean=25, std=sqrt(125)
        let out = histogram_match(src.view(), r.view()).unwrap();
        let (mu_o, sd_o) = mean_std(out.view());
        let (mu_r, sd_r) = mean_std(r.view());
        assert_abs_diff_eq!(mu_o, mu_r, epsilon = 1e-4);
        assert_abs_diff_eq!(sd_o, sd_r, epsilon = 1e-4);
    }

    #[test]
    fn zero_variance_source() {
        let src = ndarray::Array2::<f32>::from_elem((2, 2), 5.0);
        let r = array![[0.0_f32, 1.0], [2.0, 3.0]];
        let out = histogram_match(src.view(), r.view()).unwrap();
        let (mu_o, sd_o) = mean_std(out.view());
        let (mu_r, _) = mean_std(r.view());
        assert_abs_diff_eq!(mu_o, mu_r, epsilon = 1e-6);
        assert_abs_diff_eq!(sd_o, 0.0, epsilon = 1e-6);
    }
}
