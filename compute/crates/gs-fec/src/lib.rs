//! Forward Error Correction codecs (GS-FR-030).
//!
//! Implements the two FEC schemes most commonly used on satellite
//! downlinks:
//!
//! * [`reed_solomon`] — systematic Reed-Solomon over `GF(2⁸)` with the
//!   CCSDS-standard primitive polynomial `x⁸ + x⁷ + x² + x + 1`. Encoder
//!   and decoder for the canonical `(255, 223)` code (`t = 16`),
//!   capable of correcting up to 16 byte errors per codeword.
//! * [`convolutional`] — rate-1/2 constraint-length-7 convolutional
//!   encoder using the standard NASA polynomials `(g₀ = 0o171, g₁ = 0o133)`
//!   and a Viterbi decoder over the resulting trellis.
//!
//! LDPC and Turbo decoders, also referenced in EO-FR-015 / GS-FR-030,
//! are large self-contained modules and are intentionally not part of
//! this crate; consumers needing them should depend on dedicated
//! crates (e.g. `ldpc-toolbox`).

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod convolutional;
pub mod reed_solomon;
