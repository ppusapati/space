//! Export geospatial features (GI-FR-024).
//!
//! Two formats are implemented:
//!
//! * [`to_geojson`] — RFC 7946 GeoJSON via the upstream `geojson` crate.
//! * [`to_kml`] — minimal OGC KML 2.2 emitter producing a single
//!   `Document` with one `Folder` per feature collection. KML is XML so
//!   the output is built by direct string assembly to avoid a heavy
//!   dependency for what is essentially a serialiser.
//!
//! Shapefile (which uses three binary files plus optional `.dbf`) and
//! GeoPackage (a SQLite database) are intentionally out of scope and
//! require dedicated crates (`shapefile`, `gpkg`, `rusqlite`).

#![cfg_attr(docsrs, feature(doc_cfg))]

use thiserror::Error;

/// Errors produced by `gi-export`.
#[derive(Debug, Error)]
pub enum ExportError {
    /// GeoJSON serialisation error.
    #[error("GeoJSON encoding failed: {0}")]
    GeoJson(String),
}

/// One geometry to export. Coordinates are `(lon, lat)` per GeoJSON / KML
/// convention.
#[derive(Debug, Clone)]
pub enum Geometry {
    /// `(lon, lat)`.
    Point(f64, f64),
    /// Sequence of `(lon, lat)` vertices.
    Linestring(Vec<(f64, f64)>),
    /// Outer ring of a polygon (no holes).
    Polygon(Vec<(f64, f64)>),
}

/// One feature: geometry + properties (string key/value pairs).
#[derive(Debug, Clone)]
pub struct Feature {
    /// Geometry.
    pub geometry: Geometry,
    /// Properties.
    pub properties: Vec<(String, String)>,
}

/// Render a `FeatureCollection` to a GeoJSON string.
///
/// # Errors
/// [`ExportError::GeoJson`] for serialisation failures (effectively
/// out-of-memory).
pub fn to_geojson(features: &[Feature]) -> Result<String, ExportError> {
    let collection = geojson::FeatureCollection {
        bbox: None,
        features: features.iter().map(geojson_feature).collect(),
        foreign_members: None,
    };
    serde_json::to_string(&collection).map_err(|e| ExportError::GeoJson(e.to_string()))
}

fn geojson_feature(f: &Feature) -> geojson::Feature {
    let geom = match &f.geometry {
        Geometry::Point(lon, lat) => {
            geojson::Geometry::new(geojson::Value::Point(vec![*lon, *lat]))
        }
        Geometry::Linestring(pts) => geojson::Geometry::new(geojson::Value::LineString(
            pts.iter().map(|&(x, y)| vec![x, y]).collect(),
        )),
        Geometry::Polygon(ring) => geojson::Geometry::new(geojson::Value::Polygon(vec![ring
            .iter()
            .map(|&(x, y)| vec![x, y])
            .collect()])),
    };
    let mut properties = serde_json::Map::new();
    for (k, v) in &f.properties {
        properties.insert(k.clone(), serde_json::Value::String(v.clone()));
    }
    geojson::Feature {
        bbox: None,
        geometry: Some(geom),
        id: None,
        properties: Some(properties),
        foreign_members: None,
    }
}

/// Render a feature collection to OGC KML 2.2.
#[must_use]
pub fn to_kml(name: &str, features: &[Feature]) -> String {
    let mut out = String::new();
    out.push_str(r#"<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2"><Document>"#);
    out.push_str("<name>");
    push_xml_escaped(&mut out, name);
    out.push_str("</name>");
    for (i, f) in features.iter().enumerate() {
        out.push_str("<Placemark>");
        out.push_str(&format!("<name>feature-{i}</name>"));
        for (k, v) in &f.properties {
            out.push_str("<ExtendedData><Data name=\"");
            push_xml_escaped(&mut out, k);
            out.push_str("\"><value>");
            push_xml_escaped(&mut out, v);
            out.push_str("</value></Data></ExtendedData>");
        }
        match &f.geometry {
            Geometry::Point(lon, lat) => {
                out.push_str(&format!(
                    "<Point><coordinates>{lon},{lat},0</coordinates></Point>"
                ));
            }
            Geometry::Linestring(pts) => {
                out.push_str("<LineString><coordinates>");
                for (x, y) in pts {
                    out.push_str(&format!("{x},{y},0 "));
                }
                out.push_str("</coordinates></LineString>");
            }
            Geometry::Polygon(ring) => {
                out.push_str(
                    "<Polygon><outerBoundaryIs><LinearRing><coordinates>",
                );
                for (x, y) in ring {
                    out.push_str(&format!("{x},{y},0 "));
                }
                out.push_str("</coordinates></LinearRing></outerBoundaryIs></Polygon>");
            }
        }
        out.push_str("</Placemark>");
    }
    out.push_str("</Document></kml>");
    out
}

fn push_xml_escaped(out: &mut String, s: &str) {
    for ch in s.chars() {
        match ch {
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '&' => out.push_str("&amp;"),
            '"' => out.push_str("&quot;"),
            '\'' => out.push_str("&apos;"),
            other => out.push(other),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn geojson_round_trip_point() {
        let features = vec![Feature {
            geometry: Geometry::Point(10.0, 20.0),
            properties: vec![("name".into(), "Origin".into())],
        }];
        let json = to_geojson(&features).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed["type"], "FeatureCollection");
        assert_eq!(parsed["features"][0]["geometry"]["type"], "Point");
        assert_eq!(parsed["features"][0]["geometry"]["coordinates"][0], 10.0);
    }

    #[test]
    fn kml_emits_xml_envelope() {
        let features = vec![Feature {
            geometry: Geometry::Point(10.0, 20.0),
            properties: vec![("name".into(), "<Origin>".into())],
        }];
        let kml = to_kml("Test", &features);
        assert!(kml.starts_with("<?xml"));
        // Properties value `<Origin>` must be XML-escaped.
        assert!(kml.contains("&lt;Origin&gt;"));
        assert!(kml.contains("<Point><coordinates>10,20,0</coordinates></Point>"));
    }
}
