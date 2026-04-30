//! Geometric correction primitives.
//!
//! Modules:
//!
//! * [`rpc`] — Rational Polynomial Coefficient (RPC) forward (ground → image)
//!   and inverse (image → ground) projection per the OGC RPC sensor model.
//! * [`dem`] — Digital Elevation Model wrapper providing geographic-grid lookup
//!   with bilinear interpolation.
//! * [`resample`] — Nearest-neighbour, bilinear, and cubic-convolution
//!   resampling kernels for in-image sampling.
//! * [`ortho`] — DEM-aware orthorectification using either an RPC sensor model
//!   or a 6-parameter affine geo-transform.
//!
//! All public functions return `Result<…, GeometricError>`; no panics on
//! validated input.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod dem;
pub mod ortho;
pub mod resample;
pub mod rpc;

use thiserror::Error;

/// Errors produced by `eo-geometric`.
#[derive(Debug, Error, PartialEq)]
pub enum GeometricError {
    /// Out-of-range scalar parameter.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Parameter name.
        name: &'static str,
        /// Offending value as `f64`.
        value: f64,
        /// Human-readable description of the admissible range.
        range: &'static str,
    },
    /// Empty input array.
    #[error("input array must be non-empty")]
    Empty,
    /// Inverse RPC iteration failed to converge.
    #[error("inverse RPC failed to converge after {iters} iterations (residual {residual:.3e})")]
    DidNotConverge {
        /// Iterations attempted.
        iters: u32,
        /// Final residual (image-space pixel distance).
        residual: f64,
    },
    /// Coordinates outside the input array bounds.
    #[error("coordinate ({x}, {y}) is outside array bounds ({rows}, {cols})")]
    OutOfBounds {
        /// X coordinate.
        x: f64,
        /// Y coordinate.
        y: f64,
        /// Array rows.
        rows: usize,
        /// Array columns.
        cols: usize,
    },
}
