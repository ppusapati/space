//! Hamming(7,4) and extended Hamming(8,4) SEC-DED.
//!
//! Bit layout for Hamming(7,4) (1-indexed, MSB first):
//! ```text
//! position : 1   2   3   4   5   6   7
//! content  : p1  p2  d1  p3  d2  d3  d4
//! ```
//!
//! Parity equations:
//! ```text
//! p1 = d1 ⊕ d2 ⊕ d4
//! p2 = d1 ⊕ d3 ⊕ d4
//! p3 = d2 ⊕ d3 ⊕ d4
//! ```
//!
//! Extended Hamming(8,4) prepends an overall-parity bit `p0` (XOR of
//! all seven Hamming bits) at position 0 to detect double-bit errors.

use crate::DecodeOutcome;

/// Encode a 4-bit nibble (`data ∈ [0, 15]`) into a Hamming(7,4) codeword.
/// Returns the lower 7 bits packed into a `u8`.
#[must_use]
pub fn encode_7_4(data: u8) -> u8 {
    let d1 = data & 1;
    let d2 = (data >> 1) & 1;
    let d3 = (data >> 2) & 1;
    let d4 = (data >> 3) & 1;
    let p1 = d1 ^ d2 ^ d4;
    let p2 = d1 ^ d3 ^ d4;
    let p3 = d2 ^ d3 ^ d4;
    // Bits 1..7 are p1, p2, d1, p3, d2, d3, d4. We pack as:
    // bit0 = p1, bit1 = p2, bit2 = d1, bit3 = p3, bit4 = d2, bit5 = d3, bit6 = d4.
    p1 | (p2 << 1) | (d1 << 2) | (p3 << 3) | (d2 << 4) | (d3 << 5) | (d4 << 6)
}

/// Decode a Hamming(7,4) codeword. Returns `(data, outcome)` — `data` is
/// the recovered nibble (after correction) and `outcome` indicates whether
/// any bit was flipped.
#[must_use]
pub fn decode_7_4(codeword: u8) -> (u8, DecodeOutcome) {
    let p1 = codeword & 1;
    let p2 = (codeword >> 1) & 1;
    let d1 = (codeword >> 2) & 1;
    let p3 = (codeword >> 3) & 1;
    let d2 = (codeword >> 4) & 1;
    let d3 = (codeword >> 5) & 1;
    let d4 = (codeword >> 6) & 1;
    let s1 = p1 ^ d1 ^ d2 ^ d4;
    let s2 = p2 ^ d1 ^ d3 ^ d4;
    let s3 = p3 ^ d2 ^ d3 ^ d4;
    let syndrome = s1 | (s2 << 1) | (s3 << 2);
    let mut cw = codeword;
    let mut outcome = DecodeOutcome::NoError;
    if syndrome != 0 {
        // Map syndrome (1..7) to a 0-indexed bit position.
        let bit_position = match syndrome {
            1 => 0_u8, // p1
            2 => 1,    // p2
            3 => 2,    // d1
            4 => 3,    // p3
            5 => 4,    // d2
            6 => 5,    // d3
            7 => 6,    // d4
            _ => unreachable!(),
        };
        cw ^= 1 << bit_position;
        outcome = DecodeOutcome::Corrected { bit_index: u32::from(bit_position) };
    }
    let d1c = (cw >> 2) & 1;
    let d2c = (cw >> 4) & 1;
    let d3c = (cw >> 5) & 1;
    let d4c = (cw >> 6) & 1;
    let data = d1c | (d2c << 1) | (d3c << 2) | (d4c << 3);
    (data, outcome)
}

/// Encode using extended Hamming(8, 4): prepends an overall-parity bit
/// (XOR of all 7 inner bits) at position 7 of the returned byte.
#[must_use]
pub fn encode_8_4(data: u8) -> u8 {
    let cw = encode_7_4(data);
    let p0 = parity_u8(cw);
    cw | (p0 << 7)
}

/// Decode extended Hamming(8, 4). Single-bit errors are corrected.
/// Double-bit errors are detected as `Uncorrectable`.
#[must_use]
pub fn decode_8_4(codeword: u8) -> (u8, DecodeOutcome) {
    let inner = codeword & 0x7F;
    let received_overall = (codeword >> 7) & 1;
    let computed_overall = parity_u8(inner);
    let overall_match = received_overall == computed_overall;

    let (data, inner_outcome) = decode_7_4(inner);
    match inner_outcome {
        DecodeOutcome::NoError => {
            if overall_match {
                (data, DecodeOutcome::NoError)
            } else {
                // Single-bit error in the overall parity bit → still decode OK.
                (data, DecodeOutcome::Corrected { bit_index: 7 })
            }
        }
        DecodeOutcome::Corrected { bit_index } => {
            if overall_match {
                // Hamming syndrome flagged a single-bit error but the
                // overall parity is now consistent → uncorrectable double-bit error.
                (data, DecodeOutcome::Uncorrectable)
            } else {
                // Hamming corrected a single-bit error and the overall parity
                // mismatches → genuine SEC.
                (data, DecodeOutcome::Corrected { bit_index })
            }
        }
        DecodeOutcome::Uncorrectable => (data, DecodeOutcome::Uncorrectable),
    }
}

#[inline]
fn parity_u8(x: u8) -> u8 {
    let mut v = x;
    v ^= v >> 4;
    v ^= v >> 2;
    v ^= v >> 1;
    v & 1
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn round_trip_7_4_no_error() {
        for data in 0u8..16 {
            let cw = encode_7_4(data);
            let (decoded, outcome) = decode_7_4(cw);
            assert_eq!(decoded, data);
            assert_eq!(outcome, DecodeOutcome::NoError);
        }
    }

    #[test]
    fn corrects_single_bit_error_7_4() {
        for data in 0u8..16 {
            for bit in 0u8..7 {
                let cw = encode_7_4(data);
                let corrupted = cw ^ (1 << bit);
                let (decoded, outcome) = decode_7_4(corrupted);
                assert_eq!(decoded, data, "data={data} bit={bit}");
                assert_eq!(outcome, DecodeOutcome::Corrected { bit_index: u32::from(bit) });
            }
        }
    }

    #[test]
    fn round_trip_8_4_no_error() {
        for data in 0u8..16 {
            let cw = encode_8_4(data);
            let (decoded, outcome) = decode_8_4(cw);
            assert_eq!(decoded, data);
            assert_eq!(outcome, DecodeOutcome::NoError);
        }
    }

    #[test]
    fn corrects_single_bit_error_8_4() {
        for data in 0u8..16 {
            for bit in 0u8..8 {
                let cw = encode_8_4(data);
                let corrupted = cw ^ (1 << bit);
                let (decoded, outcome) = decode_8_4(corrupted);
                assert_eq!(decoded, data, "data={data} bit={bit}");
                assert!(matches!(outcome, DecodeOutcome::Corrected { .. }), "bit={bit}");
            }
        }
    }

    #[test]
    fn detects_double_bit_error_8_4() {
        for data in 0u8..16 {
            for b1 in 0u8..7 {
                for b2 in (b1 + 1)..8 {
                    let cw = encode_8_4(data);
                    let corrupted = cw ^ (1 << b1) ^ (1 << b2);
                    let (_, outcome) = decode_8_4(corrupted);
                    assert_eq!(
                        outcome,
                        DecodeOutcome::Uncorrectable,
                        "data={data} b1={b1} b2={b2}"
                    );
                }
            }
        }
    }
}
