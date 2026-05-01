//! Telecommand Transfer Frame (CCSDS 232.0-B-3).
//!
//! Primary Header (5 octets, big-endian):
//!
//! ```text
//! ┌────┬──────┬──────┬─────┬───────┬─────────────────────┬───────────────┬───────────┐
//! │ TF │ Bypa │ Cont │ Spr │  SCID │      Virtual         │ Frame Length  │  Frame Seq │
//! │ Ver│ Flag │ Flag │ res │       │      Channel         │   (10 bits)   │   Number   │
//! │ 2  │  1   │  1   │  2  │  10   │       6              │      10       │     8      │   bits
//! └────┴──────┴──────┴─────┴───────┴─────────────────────┴───────────────┴───────────┘
//! ```
//!
//! Frame structure:
//!
//! ```text
//! [primary header (5)][data field (N)][optional FECF (2)]
//! ```
//!
//! `frame_length` is the length of the entire frame (including primary
//! header and FECF if present) MINUS one, expressed as a 10-bit field.

use crate::{CcsdsError, crc::crc16_ccitt};

/// Telecommand frame primary header.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct TcPrimaryHeader {
    /// Transfer Frame Version Number (2 bits, always 0).
    pub version: u8,
    /// Bypass Flag (1 bit).
    pub bypass: bool,
    /// Control Command Flag (1 bit).
    pub control_command: bool,
    /// Spacecraft Identifier (10 bits).
    pub scid: u16,
    /// Virtual Channel Identifier (6 bits).
    pub vcid: u8,
    /// Frame Length minus 1 (10 bits) — total frame length in octets, − 1.
    pub frame_length_minus_one: u16,
    /// Frame Sequence Number (8 bits).
    pub frame_seq: u8,
}

impl TcPrimaryHeader {
    /// Encoded primary header length (5 octets).
    pub const SIZE: usize = 5;

    /// Encode to 5 big-endian bytes.
    ///
    /// # Errors
    /// [`CcsdsError::OutOfRange`] if any field exceeds its bit-width.
    pub fn encode(&self) -> Result<[u8; 5], CcsdsError> {
        if self.version > 0x03 {
            return Err(CcsdsError::OutOfRange {
                name: "tc.version",
                value: u32::from(self.version),
                range: "[0, 3]",
            });
        }
        if self.scid > 0x03FF {
            return Err(CcsdsError::OutOfRange {
                name: "tc.scid",
                value: u32::from(self.scid),
                range: "[0, 0x3FF]",
            });
        }
        if self.vcid > 0x3F {
            return Err(CcsdsError::OutOfRange {
                name: "tc.vcid",
                value: u32::from(self.vcid),
                range: "[0, 0x3F]",
            });
        }
        if self.frame_length_minus_one > 0x03FF {
            return Err(CcsdsError::OutOfRange {
                name: "tc.frame_length_minus_one",
                value: u32::from(self.frame_length_minus_one),
                range: "[0, 0x3FF]",
            });
        }
        let bypass_bit = u16::from(self.bypass);
        let control_bit = u16::from(self.control_command);
        let word0: u16 = (u16::from(self.version) << 14)
            | (bypass_bit << 13)
            | (control_bit << 12)
            | self.scid;
        let word1: u16 = (u16::from(self.vcid) << 10) | self.frame_length_minus_one;
        Ok([
            (word0 >> 8) as u8,
            (word0 & 0xFF) as u8,
            (word1 >> 8) as u8,
            (word1 & 0xFF) as u8,
            self.frame_seq,
        ])
    }

    /// Transfer Frame Version Number mandated by CCSDS 232.0-B-3
    /// (always `0` for current TC frames). Mission policy rejects
    /// anything else at the decode boundary via [`Self::decode_strict`].
    pub const EXPECTED_VERSION: u8 = 0;

    /// Decode from a slice (must be ≥ 5 bytes). Permissive: accepts any
    /// version bits and surfaces them to the caller. Use
    /// [`Self::decode_strict`] when you want the spec-mandated version
    /// policy enforced at the protocol boundary.
    ///
    /// # Errors
    /// [`CcsdsError::Truncated`] if the slice is too short.
    pub fn decode(buf: &[u8]) -> Result<Self, CcsdsError> {
        if buf.len() < Self::SIZE {
            return Err(CcsdsError::Truncated { needed: Self::SIZE, got: buf.len() });
        }
        let word0 = (u16::from(buf[0]) << 8) | u16::from(buf[1]);
        let word1 = (u16::from(buf[2]) << 8) | u16::from(buf[3]);
        Ok(Self {
            version: ((word0 >> 14) & 0x03) as u8,
            bypass: (word0 >> 13) & 0x01 != 0,
            control_command: (word0 >> 12) & 0x01 != 0,
            scid: word0 & 0x03FF,
            vcid: ((word1 >> 10) & 0x3F) as u8,
            frame_length_minus_one: word1 & 0x03FF,
            frame_seq: buf[4],
        })
    }

    /// Decode and enforce CCSDS 232.0-B-3 §4.1.2.2: the Transfer Frame
    /// Version Number must be `0`. Use this at the ingest boundary
    /// (uplink path) to reject unsupported frame versions before they
    /// propagate further into command processing.
    ///
    /// # Errors
    /// [`CcsdsError::Truncated`] if the slice is too short;
    /// [`CcsdsError::UnsupportedVersion`] if the version field is not
    /// [`Self::EXPECTED_VERSION`].
    pub fn decode_strict(buf: &[u8]) -> Result<Self, CcsdsError> {
        let header = Self::decode(buf)?;
        if header.version != Self::EXPECTED_VERSION {
            return Err(CcsdsError::UnsupportedVersion {
                field: "tc.version",
                got: header.version,
                expected: Self::EXPECTED_VERSION,
            });
        }
        Ok(header)
    }
}

/// Build a complete TC frame: primary header + data field + optional FECF.
/// `with_fecf = true` appends a CRC-16-CCITT computed over the entire
/// frame.
///
/// # Errors
/// [`CcsdsError::OutOfRange`] if header fields are invalid.
pub fn build_tc_frame(
    mut header: TcPrimaryHeader,
    data: &[u8],
    with_fecf: bool,
) -> Result<Vec<u8>, CcsdsError> {
    let total_len = TcPrimaryHeader::SIZE + data.len() + if with_fecf { 2 } else { 0 };
    if total_len == 0 || total_len > 1024 {
        return Err(CcsdsError::OutOfRange {
            name: "tc.total_length",
            value: total_len as u32,
            range: "[1, 1024]",
        });
    }
    header.frame_length_minus_one = (total_len - 1) as u16;
    let hdr = header.encode()?;
    let mut out = Vec::with_capacity(total_len);
    out.extend_from_slice(&hdr);
    out.extend_from_slice(data);
    if with_fecf {
        let crc = crc16_ccitt(&out);
        out.push((crc >> 8) as u8);
        out.push((crc & 0xFF) as u8);
    }
    Ok(out)
}

/// Validate a TC frame's FECF.
///
/// # Errors
/// [`CcsdsError::Truncated`] if frame too short;
/// [`CcsdsError::BadFecf`] on CRC mismatch.
pub fn validate_fecf(frame: &[u8]) -> Result<(), CcsdsError> {
    if frame.len() < 7 {
        return Err(CcsdsError::Truncated { needed: 7, got: frame.len() });
    }
    let computed = crc16_ccitt(&frame[..frame.len() - 2]);
    let received = (u16::from(frame[frame.len() - 2]) << 8) | u16::from(frame[frame.len() - 1]);
    if computed != received {
        return Err(CcsdsError::BadFecf { computed, expected: received });
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn header_round_trip() {
        let h = TcPrimaryHeader {
            version: 0,
            bypass: true,
            control_command: false,
            scid: 0x12,
            vcid: 0x05,
            frame_length_minus_one: 99,
            frame_seq: 0xA5,
        };
        let bytes = h.encode().unwrap();
        assert_eq!(TcPrimaryHeader::decode(&bytes).unwrap(), h);
    }

    #[test]
    fn build_and_validate_fecf() {
        let h = TcPrimaryHeader {
            version: 0,
            bypass: false,
            control_command: false,
            scid: 0x12,
            vcid: 0,
            frame_length_minus_one: 0,
            frame_seq: 0,
        };
        let frame = build_tc_frame(h, &[0x10, 0x20, 0x30, 0x40], true).unwrap();
        assert_eq!(frame.len(), 5 + 4 + 2);
        validate_fecf(&frame).unwrap();
    }

    #[test]
    fn decode_strict_accepts_spec_version() {
        let h = TcPrimaryHeader {
            version: 0,
            bypass: false,
            control_command: true,
            scid: 0x42,
            vcid: 0x07,
            frame_length_minus_one: 32,
            frame_seq: 9,
        };
        let bytes = h.encode().unwrap();
        let decoded = TcPrimaryHeader::decode_strict(&bytes).unwrap();
        assert_eq!(decoded, h);
    }

    #[test]
    fn decode_strict_rejects_nonzero_version() {
        // Hand-craft a header byte sequence with version = 1 in the top
        // two bits. Permissive decode must succeed; strict decode must
        // reject with UnsupportedVersion.
        let mut bytes = TcPrimaryHeader {
            version: 0,
            bypass: false,
            control_command: false,
            scid: 0,
            vcid: 0,
            frame_length_minus_one: 0,
            frame_seq: 0,
        }
        .encode()
        .unwrap();
        bytes[0] |= 0b0100_0000;
        let lax = TcPrimaryHeader::decode(&bytes).unwrap();
        assert_eq!(lax.version, 1);
        let err = TcPrimaryHeader::decode_strict(&bytes).unwrap_err();
        assert!(matches!(
            err,
            CcsdsError::UnsupportedVersion { field: "tc.version", got: 1, expected: 0 }
        ));
    }

    #[test]
    fn fecf_corruption_rejected() {
        let h = TcPrimaryHeader {
            version: 0,
            bypass: false,
            control_command: false,
            scid: 0x12,
            vcid: 0,
            frame_length_minus_one: 0,
            frame_seq: 0,
        };
        let mut frame = build_tc_frame(h, &[0; 8], true).unwrap();
        frame[3] ^= 0x01;
        assert!(matches!(validate_fecf(&frame).unwrap_err(), CcsdsError::BadFecf { .. }));
    }
}
