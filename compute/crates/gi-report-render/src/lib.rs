//! Intelligence-report PDF rendering (GI-FR-022).
//!
//! Builds a structured one-page report (A4) with:
//! * a title bar
//! * a subtitle / classification banner
//! * an arbitrary number of section headers
//! * key / value rows under each section
//!
//! The implementation wraps `printpdf` and uses the built-in Helvetica
//! font so reports render without external assets.

#![cfg_attr(docsrs, feature(doc_cfg))]

use printpdf::{BuiltinFont, Mm, PdfDocument};
use thiserror::Error;

/// Errors produced by `gi-report-render`.
#[derive(Debug, Error)]
pub enum ReportError {
    /// Underlying `printpdf` error.
    #[error("PDF generation failed: {0}")]
    Pdf(String),
}

/// One section of the report.
#[derive(Debug, Clone)]
pub struct Section {
    /// Heading (rendered in bold).
    pub heading: String,
    /// Key/value rows.
    pub rows: Vec<(String, String)>,
}

/// Render a report to a PDF byte buffer.
///
/// # Errors
/// [`ReportError::Pdf`] if `printpdf` fails to serialise the document.
pub fn render_pdf(
    title: &str,
    classification: &str,
    sections: &[Section],
) -> Result<Vec<u8>, ReportError> {
    let (doc, page, layer) = PdfDocument::new(title, Mm(210.0), Mm(297.0), "main");
    let regular = doc
        .add_builtin_font(BuiltinFont::Helvetica)
        .map_err(|e| ReportError::Pdf(format!("{e:?}")))?;
    let bold = doc
        .add_builtin_font(BuiltinFont::HelveticaBold)
        .map_err(|e| ReportError::Pdf(format!("{e:?}")))?;
    let l = doc.get_page(page).get_layer(layer);
    let mut y: f32 = 280.0;
    l.use_text(title, 18.0, Mm(20.0), Mm(y), &bold);
    y -= 8.0;
    l.use_text(classification, 10.0, Mm(20.0), Mm(y), &regular);
    y -= 12.0;

    for s in sections {
        l.use_text(&s.heading, 14.0, Mm(20.0), Mm(y), &bold);
        y -= 7.0;
        for (k, v) in &s.rows {
            l.use_text(k, 11.0, Mm(22.0), Mm(y), &bold);
            l.use_text(v, 11.0, Mm(60.0), Mm(y), &regular);
            y -= 5.5;
        }
        y -= 4.0;
    }

    doc.save_to_bytes().map_err(|e| ReportError::Pdf(format!("{e:?}")))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn render_emits_pdf_header() {
        let sections = vec![
            Section {
                heading: "Mission".into(),
                rows: vec![
                    ("Target".into(), "Port of XYZ".into()),
                    ("Start".into(), "2026-04-30T00:00Z".into()),
                ],
            },
            Section {
                heading: "Findings".into(),
                rows: vec![("Vessels".into(), "12 detected".into())],
            },
        ];
        let bytes = render_pdf("Activity Report", "INTERNAL", &sections).unwrap();
        // First 4 bytes of every PDF are "%PDF".
        assert_eq!(&bytes[..4], b"%PDF");
    }
}
