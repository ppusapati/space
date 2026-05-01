//! CRC-16-CCITT (polynomial `0x1021`, initial `0xFFFF`, no reflection).
//!
//! This is the variant used by CCSDS for TC and TM Frame Error Control,
//! and also by SDLC / X.25.

/// Compute the CRC-16-CCITT of the given byte slice.
#[must_use]
pub fn crc16_ccitt(data: &[u8]) -> u16 {
    let mut crc: u16 = 0xFFFF;
    for &b in data {
        crc ^= u16::from(b) << 8;
        for _ in 0..8 {
            if crc & 0x8000 != 0 {
                crc = (crc << 1) ^ 0x1021;
            } else {
                crc <<= 1;
            }
        }
    }
    crc
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Standard test vector: CRC-CCITT of "123456789" should be 0x29B1.
    #[test]
    fn known_test_vector() {
        let crc = crc16_ccitt(b"123456789");
        assert_eq!(crc, 0x29B1);
    }

    #[test]
    fn empty_input_returns_initial() {
        assert_eq!(crc16_ccitt(&[]), 0xFFFF);
    }
}
