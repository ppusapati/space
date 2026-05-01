//! Activity-Based Intelligence (GI-FR-010).
//!
//! Implements two foundational primitives:
//!
//! * [`PatternOfLife`] — accumulates a baseline distribution (mean and
//!   variance) over a stream of observations.
//! * [`anomaly_zscore`] — Z-score anomaly detector that flags
//!   observations whose deviation exceeds a configurable threshold.

#![cfg_attr(docsrs, feature(doc_cfg))]

use thiserror::Error;

/// Errors produced by `gi-abi`.
#[derive(Debug, Error, PartialEq)]
pub enum AbiError {
    /// Insufficient data to compute statistics.
    #[error("baseline has {0} samples; need ≥ 2")]
    InsufficientData(usize),
}

/// Pattern-of-life accumulator. Uses Welford's online algorithm for
/// numerically stable mean and variance updates.
#[derive(Debug, Clone, Default)]
pub struct PatternOfLife {
    /// Number of samples seen.
    pub n: u64,
    /// Running mean.
    pub mean: f64,
    /// Sum of squared deviations from the running mean (M₂ in Welford).
    pub m2: f64,
}

impl PatternOfLife {
    /// Construct an empty baseline.
    #[must_use]
    pub fn new() -> Self {
        Self::default()
    }

    /// Update with one new observation.
    pub fn push(&mut self, x: f64) {
        if !x.is_finite() {
            return;
        }
        self.n += 1;
        let delta = x - self.mean;
        self.mean += delta / (self.n as f64);
        let delta2 = x - self.mean;
        self.m2 += delta * delta2;
    }

    /// Sample variance.
    ///
    /// # Errors
    /// [`AbiError::InsufficientData`] if fewer than 2 samples have been
    /// observed.
    pub fn variance(&self) -> Result<f64, AbiError> {
        if self.n < 2 {
            return Err(AbiError::InsufficientData(self.n as usize));
        }
        Ok(self.m2 / ((self.n - 1) as f64))
    }

    /// Sample standard deviation.
    ///
    /// # Errors
    /// As [`PatternOfLife::variance`].
    pub fn std_dev(&self) -> Result<f64, AbiError> {
        self.variance().map(f64::sqrt)
    }
}

/// Z-score anomaly classification.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Anomaly {
    /// Within `±threshold · σ` of the mean.
    Normal,
    /// Beyond `+threshold · σ`.
    HighOutlier,
    /// Beyond `−threshold · σ`.
    LowOutlier,
}

/// Classify a value against a baseline.
///
/// # Errors
/// Propagates [`PatternOfLife::std_dev`] errors.
pub fn anomaly_zscore(
    baseline: &PatternOfLife,
    value: f64,
    threshold: f64,
) -> Result<(Anomaly, f64), AbiError> {
    let sigma = baseline.std_dev()?;
    if sigma == 0.0 {
        return Ok((Anomaly::Normal, 0.0));
    }
    let z = (value - baseline.mean) / sigma;
    let cls = if z > threshold {
        Anomaly::HighOutlier
    } else if z < -threshold {
        Anomaly::LowOutlier
    } else {
        Anomaly::Normal
    };
    Ok((cls, z))
}

#[cfg(test)]
mod tests {
    use approx::assert_abs_diff_eq;

    use super::*;

    #[test]
    fn welford_matches_textbook_mean_variance() {
        let mut pol = PatternOfLife::new();
        for v in &[1.0, 2.0, 3.0, 4.0, 5.0] {
            pol.push(*v);
        }
        assert_abs_diff_eq!(pol.mean, 3.0, epsilon = 1e-12);
        assert_abs_diff_eq!(pol.variance().unwrap(), 2.5, epsilon = 1e-12);
    }

    #[test]
    fn anomaly_classifies_three_categories() {
        let mut pol = PatternOfLife::new();
        for v in 0..100 {
            pol.push(f64::from(v));
        }
        let (cls, z) = anomaly_zscore(&pol, 50.0, 2.0).unwrap();
        assert_eq!(cls, Anomaly::Normal);
        let (cls, _z) = anomaly_zscore(&pol, 200.0, 2.0).unwrap();
        assert_eq!(cls, Anomaly::HighOutlier);
        let (cls, _z) = anomaly_zscore(&pol, -100.0, 2.0).unwrap();
        assert_eq!(cls, Anomaly::LowOutlier);
        let _ = z;
    }

    #[test]
    fn rejects_insufficient_data() {
        let pol = PatternOfLife::new();
        assert!(matches!(pol.variance(), Err(AbiError::InsufficientData(0))));
    }
}
