//! Image mosaicking with seamline generation and blending.
//!
//! Given a set of overlapping orthorectified tiles described by an
//! [`AffineGeo`] and a 2-D raster, this crate provides:
//!
//! * [`overlap`] — pairwise overlap computation in both tiles' pixel space.
//! * [`seamline`] — minimum-cost-path seamline via Dijkstra.
//! * [`blend`] — feather, average, max-value, and seamline-cut compositors.
//! * [`mosaic`] — driver that composes an arbitrary list of tiles onto a
//!   user-specified output grid.
//!
//! All algorithms operate in pixel space and assume tiles share the same
//! projection and pixel size. Reprojection is delegated to `eo-geometric`.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod blend;
pub mod mosaic;
pub mod overlap;
pub mod seamline;

pub use eo_geometric::ortho::AffineGeo;
use ndarray::Array2;
use thiserror::Error;

/// A single mosaic tile: a 2-D raster plus its geo-reference.
#[derive(Debug, Clone)]
pub struct Tile {
    /// Pixel grid (rows, cols) of `f32` values; `NaN` denotes nodata.
    pub raster: Array2<f32>,
    /// Affine georeference.
    pub geo: AffineGeo,
}

/// Errors produced by `eo-mosaic`.
#[derive(Debug, Error, PartialEq)]
pub enum MosaicError {
    /// Input tile list was empty.
    #[error("at least one tile is required")]
    NoTiles,
    /// Two tiles disagree on pixel size, so they cannot be combined without
    /// resampling.
    #[error("pixel-size mismatch between tiles: {a:?} vs {b:?}")]
    PixelSizeMismatch {
        /// First tile pixel size.
        a: (f64, f64),
        /// Second tile pixel size.
        b: (f64, f64),
    },
    /// Two tiles do not overlap.
    #[error("tiles do not overlap")]
    NoOverlap,
    /// Output raster shape was zero on at least one axis.
    #[error("output raster shape must be non-empty")]
    EmptyOutput,
}
