//! Mapbox Vector Tile (MVT) renderer (GI-FR-020).
//!
//! Wraps the upstream `mvt` crate to produce a serialised tile from a
//! list of features. Each feature is a `(GeomType, [(x, y); …])` pair
//! in tile-local coordinates `[0, extent]`, the standard for slippy-map
//! WMS/WMTS/MVT endpoints.

#![cfg_attr(docsrs, feature(doc_cfg))]

use mvt::{GeomEncoder, GeomType, Tile};
use thiserror::Error;

/// Errors produced by `gi-tile-render`.
#[derive(Debug, Error)]
pub enum TileError {
    /// Underlying MVT serialisation error.
    #[error("MVT encoding failed: {0}")]
    Encoding(String),
}

/// One geometry feature destined for an MVT layer.
#[derive(Debug, Clone)]
pub struct Feature {
    /// Feature identifier.
    pub id: u64,
    /// Geometry type.
    pub kind: GeomType,
    /// Geometry vertex coordinates in tile-local units `[0, extent]`.
    pub vertices: Vec<(f64, f64)>,
}

/// Build a vector tile from a single layer's features.
///
/// `extent` is the tile resolution (typically 4096 for slippy maps).
///
/// # Errors
/// [`TileError::Encoding`] if the upstream encoder rejects the geometry
/// (e.g., empty vertex list).
pub fn build_tile(layer_name: &str, features: &[Feature], extent: u32) -> Result<Vec<u8>, TileError> {
    let mut tile = Tile::new(extent);
    let mut layer = tile.create_layer(layer_name);
    for f in features {
        let mut enc = GeomEncoder::new(f.kind);
        for &(x, y) in &f.vertices {
            enc = enc.point(x, y).map_err(|e| TileError::Encoding(format!("{e:?}")))?;
        }
        let geom = enc.encode().map_err(|e| TileError::Encoding(format!("{e:?}")))?;
        let mut feature = layer.into_feature(geom);
        feature.set_id(f.id);
        layer = feature.into_layer();
    }
    tile.add_layer(layer).map_err(|e| TileError::Encoding(format!("{e:?}")))?;
    tile.to_bytes().map_err(|e| TileError::Encoding(format!("{e:?}")))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn build_tile_with_one_point() {
        let features = vec![Feature {
            id: 1,
            kind: GeomType::Point,
            vertices: vec![(2048.0, 2048.0)],
        }];
        let bytes = build_tile("test", &features, 4096).unwrap();
        assert!(!bytes.is_empty());
        // The first byte of an MVT is the protobuf field tag; the upstream
        // `mvt` crate emits a non-empty buffer for a single-feature layer.
        assert!(bytes.len() > 5);
    }

    #[test]
    fn build_tile_with_multiple_features() {
        let features = vec![
            Feature { id: 1, kind: GeomType::Point, vertices: vec![(100.0, 200.0)] },
            Feature { id: 2, kind: GeomType::Linestring, vertices: vec![(0.0, 0.0), (50.0, 50.0)] },
        ];
        let bytes = build_tile("multi", &features, 4096).unwrap();
        assert!(!bytes.is_empty());
    }
}
