//! Pan-sharpening kernels for multispectral imagery fusion with a
//! panchromatic band of higher spatial resolution.
//!
//! The pan band must be supplied at the **same spatial dimensions** as the
//! multispectral bands; resampling (e.g. bicubic upsampling of the MS bands)
//! is the caller's responsibility and is delegated to a separate
//! resampling stage.
//!
//! Kernels implemented:
//!
//! * **Brovey** — `M'_i = M_i · Pan / I` with `I = mean(M_1..M_n)`.
//!   Simple, hue-preserving, suitable for visual products.
//! * **IHS**    — RGB → IHS, replace I with histogram-matched Pan, IHS → RGB.
//!   Strictly 3-band (Red/Green/Blue).
//! * **PCA**    — PCA decomposition of MS bands, replace PC1 with
//!   histogram-matched Pan, inverse PCA. Works for any band count ≥ 2.
//! * **Gram-Schmidt** — Laben-Brower (1998) classical Gram-Schmidt fusion.
//!
//! All kernels operate on `f32` reflectance arrays and return new owned
//! arrays. Inputs are validated for shape and band-count consistency.

#![cfg_attr(docsrs, feature(doc_cfg))]

mod brovey;
mod gs;
mod hist_match;
mod ihs;
mod pca;

pub use brovey::brovey;
pub use gs::{GsWeights, gram_schmidt};
pub use hist_match::histogram_match;
pub use ihs::ihs;
pub use pca::pca;

use thiserror::Error;

/// Errors produced by `eo-pansharpen`.
#[derive(Debug, Error, PartialEq)]
pub enum PansharpenError {
    /// Input arrays do not share `(rows, cols)`.
    #[error("input shape mismatch: expected {expected:?}, got {got:?}")]
    ShapeMismatch {
        /// Reference shape.
        expected: (usize, usize),
        /// Offending shape.
        got: (usize, usize),
    },
    /// An input array had a zero dimension or no bands.
    #[error("input must be non-empty")]
    Empty,
    /// Wrong number of input bands for a kernel that requires a fixed count.
    #[error("kernel `{kernel}` requires {expected} bands, got {got}")]
    BandCount {
        /// Kernel name.
        kernel: &'static str,
        /// Expected band count.
        expected: usize,
        /// Actual band count.
        got: usize,
    },
    /// Algorithm failure (e.g., singular covariance during PCA).
    #[error("kernel `{kernel}` failed: {reason}")]
    Algorithm {
        /// Kernel name.
        kernel: &'static str,
        /// Human-readable reason.
        reason: &'static str,
    },
}
