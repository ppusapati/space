//! Telemetry Transfer Frame (CCSDS 132.0-B-2).
//!
//! Primary Header (6 octets, big-endian):
//!
//! ```text
//! ┌────┬─────┬───────┬────┬──────┬──────────┬─────────────┬────┬─────┬────┬─────┐
//! │ TF │ SCI │ Virt. │ OC │ Master Frame  │  Virt.     │ Sec. Hdr. │ Sync │ Pkt │ Seg │
//! │ Ver│  D  │ Chan  │ F  │ Counter (8b)  │ Counter(8b)│  Flag (1) │  (1) │  (1)│ (2) │
//! │ 2  │ 10  │  3    │ 1  │       8        │     8      │     1     │   1  │  1  │  2  │   bits
//! └────┴─────┴───────┴────┴──────────────┴─────────────┴───────────┴──────┴─────┴─────┘
//! ```
//!
//! Frame structure:
//!
//! ```text
//! [primary header (6)][secondary hdr (opt)][data field][OCF (opt 4)][FECF (opt 2)]
//! ```

use crate::{CcsdsError, crc::crc16_ccitt};

/// TM frame primary header.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct TmPrimaryHeader {
    /// Transfer Frame Version Number (2 bits).
    pub version: u8,
    /// Spacecraft Identifier (10 bits).
    pub scid: u16,
    /// Virtual Channel ID (3 bits).
    pub vcid: u8,
    /// Operational Control Field flag (1 bit).
    pub ocf_flag: bool,
    /// Master Channel Frame Count (8 bits).
    pub master_frame_count: u8,
    /// Virtual Channel Frame Count (8 bits).
    pub vc_frame_count: u8,
    /// Secondary header flag (1 bit).
    pub sec_header_flag: bool,
    /// Synchronization flag (1 bit).
    pub sync_flag: bool,
    /// Packet order flag (1 bit).
    pub packet_order: bool,
    /// Segment length identifier (2 bits).
    pub segment_length_id: u8,
    /// First Header Pointer (11 bits).
    pub first_header_pointer: u16,
}

impl TmPrimaryHeader {
    /// Encoded primary header length.
    pub const SIZE: usize = 6;

    /// Encode to 6 big-endian bytes.
    ///
    /// # Errors
    /// [`CcsdsError::OutOfRange`] if any field exceeds its bit-width.
    pub fn encode(&self) -> Result<[u8; 6], CcsdsError> {
        if self.version > 0x03 {
            return Err(CcsdsError::OutOfRange {
                name: "tm.version",
                value: u32::from(self.version),
                range: "[0, 3]",
            });
        }
        if self.scid > 0x03FF {
            return Err(CcsdsError::OutOfRange {
                name: "tm.scid",
                value: u32::from(self.scid),
                range: "[0, 0x3FF]",
            });
        }
        if self.vcid > 0x07 {
            return Err(CcsdsError::OutOfRange {
                name: "tm.vcid",
                value: u32::from(self.vcid),
                range: "[0, 7]",
            });
        }
        if self.segment_length_id > 0x03 {
            return Err(CcsdsError::OutOfRange {
                name: "tm.segment_length_id",
                value: u32::from(self.segment_length_id),
                range: "[0, 3]",
            });
        }
        if self.first_header_pointer > 0x07FF {
            return Err(CcsdsError::OutOfRange {
                name: "tm.first_header_pointer",
                value: u32::from(self.first_header_pointer),
                range: "[0, 0x7FF]",
            });
        }
        let ocf = u16::from(self.ocf_flag);
        let word0 = (u16::from(self.version) << 14)
            | (self.scid << 4)
            | (u16::from(self.vcid) << 1)
            | ocf;
        let word2_lower: u16 = (u16::from(self.sec_header_flag) << 15)
            | (u16::from(self.sync_flag) << 14)
            | (u16::from(self.packet_order) << 13)
            | (u16::from(self.segment_length_id) << 11)
            | self.first_header_pointer;
        Ok([
            (word0 >> 8) as u8,
            (word0 & 0xFF) as u8,
            self.master_frame_count,
            self.vc_frame_count,
            (word2_lower >> 8) as u8,
            (word2_lower & 0xFF) as u8,
        ])
    }

    /// Decode from a slice.
    ///
    /// # Errors
    /// [`CcsdsError::Truncated`] if too short.
    pub fn decode(buf: &[u8]) -> Result<Self, CcsdsError> {
        if buf.len() < Self::SIZE {
            return Err(CcsdsError::Truncated { needed: Self::SIZE, got: buf.len() });
        }
        let word0 = (u16::from(buf[0]) << 8) | u16::from(buf[1]);
        let word2 = (u16::from(buf[4]) << 8) | u16::from(buf[5]);
        Ok(Self {
            version: ((word0 >> 14) & 0x03) as u8,
            scid: (word0 >> 4) & 0x03FF,
            vcid: ((word0 >> 1) & 0x07) as u8,
            ocf_flag: word0 & 0x01 != 0,
            master_frame_count: buf[2],
            vc_frame_count: buf[3],
            sec_header_flag: (word2 >> 15) & 0x01 != 0,
            sync_flag: (word2 >> 14) & 0x01 != 0,
            packet_order: (word2 >> 13) & 0x01 != 0,
            segment_length_id: ((word2 >> 11) & 0x03) as u8,
            first_header_pointer: word2 & 0x07FF,
        })
    }
}

/// Build a complete TM frame.
///
/// # Errors
/// Header encode errors.
pub fn build_tm_frame(
    header: TmPrimaryHeader,
    data: &[u8],
    ocf: Option<[u8; 4]>,
    with_fecf: bool,
) -> Result<Vec<u8>, CcsdsError> {
    let mut out = Vec::with_capacity(TmPrimaryHeader::SIZE + data.len() + 6);
    out.extend_from_slice(&header.encode()?);
    out.extend_from_slice(data);
    if let Some(o) = ocf {
        out.extend_from_slice(&o);
    }
    if with_fecf {
        let crc = crc16_ccitt(&out);
        out.push((crc >> 8) as u8);
        out.push((crc & 0xFF) as u8);
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn header_round_trip() {
        let h = TmPrimaryHeader {
            version: 0,
            scid: 0x12,
            vcid: 5,
            ocf_flag: true,
            master_frame_count: 0xAB,
            vc_frame_count: 0xCD,
            sec_header_flag: true,
            sync_flag: false,
            packet_order: true,
            segment_length_id: 3,
            first_header_pointer: 0x123,
        };
        let bytes = h.encode().unwrap();
        assert_eq!(TmPrimaryHeader::decode(&bytes).unwrap(), h);
    }

    #[test]
    fn build_with_ocf_and_fecf() {
        let h = TmPrimaryHeader {
            version: 0,
            scid: 0x12,
            vcid: 0,
            ocf_flag: true,
            master_frame_count: 0,
            vc_frame_count: 0,
            sec_header_flag: false,
            sync_flag: false,
            packet_order: false,
            segment_length_id: 3,
            first_header_pointer: 0,
        };
        let frame = build_tm_frame(h, &[0u8; 100], Some([0xDE, 0xAD, 0xBE, 0xEF]), true).unwrap();
        assert_eq!(frame.len(), 6 + 100 + 4 + 2);
        // Validate FECF.
        let crc = crc16_ccitt(&frame[..frame.len() - 2]);
        let received =
            (u16::from(frame[frame.len() - 2]) << 8) | u16::from(frame[frame.len() - 1]);
        assert_eq!(crc, received);
    }
}
