//! Space Packet Protocol (CCSDS 133.0-B-2).
//!
//! Primary Header (6 octets, big-endian):
//!
//! ```text
//! ┌───┬───┬───┬─────────────┬────────┬────────┬───────────────┐
//! │ V │ T │ S │   APID      │  SQF   │  SQC   │  Packet Len   │
//! │ 3 │ 1 │ 1 │     11      │   2    │   14   │      16       │   bits
//! └───┴───┴───┴─────────────┴────────┴────────┴───────────────┘
//! ```
//!
//! * **V** — Packet Version Number (always `0` for SPP).
//! * **T** — Packet Type (`0` = telemetry, `1` = telecommand).
//! * **S** — Secondary Header Flag (`1` = present).
//! * **APID** — Application Process Identifier (11 bits, `0x000–0x7FF`).
//! * **SQF** — Sequence Flags (00 = continuation, 01 = first, 10 = last,
//!   11 = unsegmented).
//! * **SQC** — Packet Sequence Count (or Packet Name) (14 bits).
//! * **Packet Length** — Number of octets in the data field MINUS one.

use crate::CcsdsError;

/// Packet type discriminator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PacketType {
    /// Telemetry (downlink).
    Telemetry,
    /// Telecommand (uplink).
    Telecommand,
}

/// Sequence Flag values per CCSDS 133.0-B-2.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SequenceFlag {
    /// Continuation segment.
    Continuation = 0b00,
    /// First segment of a multi-packet sequence.
    First = 0b01,
    /// Last segment.
    Last = 0b10,
    /// Unsegmented (the packet is a complete unit).
    Unsegmented = 0b11,
}

/// SPP primary header.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct PrimaryHeader {
    /// Version (always 0 for SPP).
    pub version: u8,
    /// Packet type.
    pub packet_type: PacketType,
    /// Secondary header flag.
    pub secondary_header: bool,
    /// Application Process Identifier (11 bits).
    pub apid: u16,
    /// Sequence flag.
    pub sequence_flag: SequenceFlag,
    /// Packet sequence count (14 bits).
    pub sequence_count: u16,
    /// Packet data length minus 1 (16 bits).
    pub data_length_minus_one: u16,
}

impl PrimaryHeader {
    /// Encoded header length (always 6 octets).
    pub const SIZE: usize = 6;

    /// Encode to 6 big-endian bytes.
    ///
    /// # Errors
    /// [`CcsdsError::OutOfRange`] if APID > 0x7FF or sequence_count > 0x3FFF.
    pub fn encode(&self) -> Result<[u8; 6], CcsdsError> {
        if self.apid > 0x07FF {
            return Err(CcsdsError::OutOfRange {
                name: "apid",
                value: u32::from(self.apid),
                range: "[0, 0x7FF]",
            });
        }
        if self.sequence_count > 0x3FFF {
            return Err(CcsdsError::OutOfRange {
                name: "sequence_count",
                value: u32::from(self.sequence_count),
                range: "[0, 0x3FFF]",
            });
        }
        if self.version > 0x07 {
            return Err(CcsdsError::OutOfRange {
                name: "version",
                value: u32::from(self.version),
                range: "[0, 7]",
            });
        }
        let type_bit = match self.packet_type {
            PacketType::Telemetry => 0_u16,
            PacketType::Telecommand => 1_u16,
        };
        let sec_hdr_bit = u16::from(self.secondary_header);
        let word0: u16 = (u16::from(self.version) << 13)
            | (type_bit << 12)
            | (sec_hdr_bit << 11)
            | self.apid;
        let seq_flag = self.sequence_flag as u16;
        let word1: u16 = (seq_flag << 14) | self.sequence_count;
        let word2: u16 = self.data_length_minus_one;
        Ok([
            (word0 >> 8) as u8,
            (word0 & 0xFF) as u8,
            (word1 >> 8) as u8,
            (word1 & 0xFF) as u8,
            (word2 >> 8) as u8,
            (word2 & 0xFF) as u8,
        ])
    }

    /// Decode from a slice (must be ≥ 6 bytes).
    ///
    /// # Errors
    /// [`CcsdsError::Truncated`] if the slice is too short.
    pub fn decode(buf: &[u8]) -> Result<Self, CcsdsError> {
        if buf.len() < Self::SIZE {
            return Err(CcsdsError::Truncated { needed: Self::SIZE, got: buf.len() });
        }
        let word0 = (u16::from(buf[0]) << 8) | u16::from(buf[1]);
        let word1 = (u16::from(buf[2]) << 8) | u16::from(buf[3]);
        let word2 = (u16::from(buf[4]) << 8) | u16::from(buf[5]);
        let version = ((word0 >> 13) & 0x07) as u8;
        let type_bit = (word0 >> 12) & 0x01;
        let sec_hdr = (word0 >> 11) & 0x01 != 0;
        let apid = word0 & 0x07FF;
        let seq_flag_bits = ((word1 >> 14) & 0x03) as u8;
        let seq_flag = match seq_flag_bits {
            0b00 => SequenceFlag::Continuation,
            0b01 => SequenceFlag::First,
            0b10 => SequenceFlag::Last,
            _ => SequenceFlag::Unsegmented,
        };
        let seq_count = word1 & 0x3FFF;
        let pkt_type = if type_bit == 0 { PacketType::Telemetry } else { PacketType::Telecommand };
        Ok(Self {
            version,
            packet_type: pkt_type,
            secondary_header: sec_hdr,
            apid,
            sequence_flag: seq_flag,
            sequence_count: seq_count,
            data_length_minus_one: word2,
        })
    }

    /// Total packet length on the wire = 6 (primary header) + (data_length_minus_one + 1).
    #[must_use]
    pub fn total_length(&self) -> usize {
        Self::SIZE + usize::from(self.data_length_minus_one) + 1
    }
}

/// A complete Space Packet (header + data field).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SpacePacket {
    /// Primary header.
    pub header: PrimaryHeader,
    /// Packet data field (secondary header + user data, if present).
    pub data: Vec<u8>,
}

impl SpacePacket {
    /// Build a packet from header info and a data payload. The `data_length_minus_one`
    /// field is set automatically.
    ///
    /// # Errors
    /// [`CcsdsError::OutOfRange`] if `data` is empty or longer than 65 536 bytes.
    pub fn build(
        version: u8,
        packet_type: PacketType,
        secondary_header: bool,
        apid: u16,
        sequence_flag: SequenceFlag,
        sequence_count: u16,
        data: Vec<u8>,
    ) -> Result<Self, CcsdsError> {
        if data.is_empty() || data.len() > 65_536 {
            return Err(CcsdsError::OutOfRange {
                name: "data.len",
                value: data.len() as u32,
                range: "[1, 65536]",
            });
        }
        let header = PrimaryHeader {
            version,
            packet_type,
            secondary_header,
            apid,
            sequence_flag,
            sequence_count,
            data_length_minus_one: (data.len() - 1) as u16,
        };
        Ok(Self { header, data })
    }

    /// Serialise the packet to bytes.
    ///
    /// # Errors
    /// Propagates [`PrimaryHeader::encode`] errors.
    pub fn encode(&self) -> Result<Vec<u8>, CcsdsError> {
        let hdr = self.header.encode()?;
        let mut out = Vec::with_capacity(self.data.len() + 6);
        out.extend_from_slice(&hdr);
        out.extend_from_slice(&self.data);
        Ok(out)
    }

    /// Decode a packet from a byte slice. The decoder uses the
    /// `data_length_minus_one` field to determine how many bytes belong to
    /// the data field; trailing bytes are ignored.
    ///
    /// # Errors
    /// [`CcsdsError::Truncated`] if the slice is shorter than the
    /// declared total length.
    pub fn decode(buf: &[u8]) -> Result<Self, CcsdsError> {
        let header = PrimaryHeader::decode(buf)?;
        let total = header.total_length();
        if buf.len() < total {
            return Err(CcsdsError::Truncated { needed: total, got: buf.len() });
        }
        let data = buf[PrimaryHeader::SIZE..total].to_vec();
        Ok(Self { header, data })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn header_round_trip() {
        let h = PrimaryHeader {
            version: 0,
            packet_type: PacketType::Telecommand,
            secondary_header: true,
            apid: 0x123,
            sequence_flag: SequenceFlag::Unsegmented,
            sequence_count: 0x2345,
            data_length_minus_one: 9,
        };
        let bytes = h.encode().unwrap();
        let h2 = PrimaryHeader::decode(&bytes).unwrap();
        assert_eq!(h, h2);
    }

    #[test]
    fn apid_out_of_range_rejected() {
        let h = PrimaryHeader {
            version: 0,
            packet_type: PacketType::Telemetry,
            secondary_header: false,
            apid: 0x1000,
            sequence_flag: SequenceFlag::Unsegmented,
            sequence_count: 0,
            data_length_minus_one: 0,
        };
        assert!(matches!(h.encode().unwrap_err(), CcsdsError::OutOfRange { name: "apid", .. }));
    }

    #[test]
    fn space_packet_round_trip() {
        let pkt = SpacePacket::build(
            0,
            PacketType::Telemetry,
            false,
            0x100,
            SequenceFlag::Unsegmented,
            42,
            vec![1, 2, 3, 4, 5],
        )
        .unwrap();
        let bytes = pkt.encode().unwrap();
        let pkt2 = SpacePacket::decode(&bytes).unwrap();
        assert_eq!(pkt, pkt2);
        assert_eq!(pkt.header.total_length(), bytes.len());
    }

    #[test]
    fn truncated_packet_rejected() {
        let pkt = SpacePacket::build(
            0,
            PacketType::Telemetry,
            false,
            0x100,
            SequenceFlag::Unsegmented,
            0,
            vec![0; 100],
        )
        .unwrap();
        let bytes = pkt.encode().unwrap();
        let err = SpacePacket::decode(&bytes[..50]).unwrap_err();
        assert!(matches!(err, CcsdsError::Truncated { .. }));
    }
}
