//! Extended Hamming(72, 64) SEC-DED.
//!
//! 64 data bits protected by 7 inner Hamming parity bits and 1 overall
//! parity bit, giving an industry-standard ECC-DRAM scheme:
//!
//! * Single-bit errors are corrected.
//! * Double-bit errors are detected.
//!
//! Codeword bit positions (1-indexed). Parity positions are powers of two:
//! `1, 2, 4, 8, 16, 32, 64`. The overall parity sits at position 72. The
//! remaining 64 positions hold data bits.

use crate::DecodeOutcome;

const PARITY_POSITIONS: [u32; 7] = [1, 2, 4, 8, 16, 32, 64];
const OVERALL_POSITION: u32 = 72;

/// 0-indexed positions of the 64 data bits within a 72-bit codeword.
const DATA_POSITIONS: [u32; 64] = build_data_positions();

const fn build_data_positions() -> [u32; 64] {
    let mut out = [0_u32; 64];
    let mut idx = 0_usize;
    let mut pos = 1_u32;
    while pos < OVERALL_POSITION {
        // Skip parity positions (1, 2, 4, 8, 16, 32, 64).
        let is_pow2 = pos.is_power_of_two();
        if !is_pow2 {
            out[idx] = pos;
            idx += 1;
        }
        pos += 1;
    }
    out
}

#[inline]
fn get_bit(cw: u128, position_one_indexed: u32) -> u8 {
    ((cw >> (position_one_indexed - 1)) & 1) as u8
}

#[inline]
fn set_bit(cw: &mut u128, position_one_indexed: u32, value: u8) {
    let mask = 1_u128 << (position_one_indexed - 1);
    if value & 1 == 1 {
        *cw |= mask;
    } else {
        *cw &= !mask;
    }
}

#[inline]
fn flip_bit(cw: &mut u128, position_one_indexed: u32) {
    *cw ^= 1_u128 << (position_one_indexed - 1);
}

fn compute_inner_parity(cw: u128, parity_index: u32) -> u8 {
    // Parity p_i covers all positions in 1..=71 where bit `parity_index` of
    // the position is set, except the parity position itself.
    let bit = parity_index;
    let parity_pos = 1_u32 << bit;
    let mut acc = 0_u8;
    let mut pos = 1_u32;
    while pos < OVERALL_POSITION {
        if pos != parity_pos && (pos >> bit) & 1 == 1 {
            acc ^= get_bit(cw, pos);
        }
        pos += 1;
    }
    acc
}

fn compute_syndrome(cw: u128) -> u8 {
    let mut s = 0_u8;
    for i in 0..7 {
        let computed = compute_inner_parity(cw, i);
        let received = get_bit(cw, PARITY_POSITIONS[i as usize]);
        if computed ^ received != 0 {
            s |= 1 << i;
        }
    }
    s
}

fn compute_overall_parity(cw: u128) -> u8 {
    let mut acc = 0_u8;
    let mut pos = 1_u32;
    while pos <= OVERALL_POSITION {
        acc ^= get_bit(cw, pos);
        pos += 1;
    }
    acc
}

/// Encode a 64-bit data word into a 72-bit codeword represented as a
/// `(u64, u8)` pair: the data is unchanged, and the eight parity bits
/// are packed into a single `u8` in the order
/// `[p1, p2, p4, p8, p16, p32, p64, overall]`.
#[must_use]
pub fn encode(data: u64) -> (u64, u8) {
    let mut cw: u128 = 0;
    // Place data bits at the 64 non-parity positions.
    for (i, &pos) in DATA_POSITIONS.iter().enumerate() {
        let bit = ((data >> i) & 1) as u8;
        set_bit(&mut cw, pos, bit);
    }
    // Compute inner parities.
    for i in 0..7_u32 {
        let p = compute_inner_parity(cw, i);
        set_bit(&mut cw, PARITY_POSITIONS[i as usize], p);
    }
    // Overall parity over positions 1..=71.
    let mut acc = 0_u8;
    for pos in 1..OVERALL_POSITION {
        acc ^= get_bit(cw, pos);
    }
    set_bit(&mut cw, OVERALL_POSITION, acc);

    let parity_byte = (get_bit(cw, 1))
        | (get_bit(cw, 2) << 1)
        | (get_bit(cw, 4) << 2)
        | (get_bit(cw, 8) << 3)
        | (get_bit(cw, 16) << 4)
        | (get_bit(cw, 32) << 5)
        | (get_bit(cw, 64) << 6)
        | (get_bit(cw, OVERALL_POSITION) << 7);
    (data, parity_byte)
}

/// Decode a 72-bit codeword.
#[must_use]
pub fn decode(data: u64, parity_byte: u8) -> (u64, DecodeOutcome) {
    // Reconstruct the 128-bit codeword from data + parity.
    let mut cw: u128 = 0;
    for (i, &pos) in DATA_POSITIONS.iter().enumerate() {
        let bit = ((data >> i) & 1) as u8;
        set_bit(&mut cw, pos, bit);
    }
    set_bit(&mut cw, 1, parity_byte & 1);
    set_bit(&mut cw, 2, (parity_byte >> 1) & 1);
    set_bit(&mut cw, 4, (parity_byte >> 2) & 1);
    set_bit(&mut cw, 8, (parity_byte >> 3) & 1);
    set_bit(&mut cw, 16, (parity_byte >> 4) & 1);
    set_bit(&mut cw, 32, (parity_byte >> 5) & 1);
    set_bit(&mut cw, 64, (parity_byte >> 6) & 1);
    set_bit(&mut cw, OVERALL_POSITION, (parity_byte >> 7) & 1);

    let syndrome = compute_syndrome(cw);
    let overall = compute_overall_parity(cw);

    let outcome = match (syndrome, overall) {
        (0, 0) => DecodeOutcome::NoError,
        (0, 1) => DecodeOutcome::Corrected { bit_index: OVERALL_POSITION - 1 },
        (s, 0) if s != 0 => DecodeOutcome::Uncorrectable,
        (s, _) => {
            // Single-bit error at codeword position `s`.
            flip_bit(&mut cw, s as u32);
            DecodeOutcome::Corrected { bit_index: s as u32 - 1 }
        }
    };

    // Recover data bits from (possibly corrected) codeword.
    let mut recovered: u64 = 0;
    for (i, &pos) in DATA_POSITIONS.iter().enumerate() {
        recovered |= u64::from(get_bit(cw, pos)) << i;
    }
    (recovered, outcome)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn fixtures() -> [u64; 5] {
        [
            0x0000_0000_0000_0000,
            0xFFFF_FFFF_FFFF_FFFF,
            0xDEAD_BEEF_CAFE_BABE,
            0x0123_4567_89AB_CDEF,
            0xA5A5_A5A5_5A5A_5A5A,
        ]
    }

    #[test]
    fn round_trip_no_error() {
        for &d in &fixtures() {
            let (data, parity) = encode(d);
            let (decoded, outcome) = decode(data, parity);
            assert_eq!(decoded, d, "data {d:#x}");
            assert_eq!(outcome, DecodeOutcome::NoError);
        }
    }

    #[test]
    fn corrects_single_bit_in_data() {
        let d = 0xDEAD_BEEF_CAFE_BABE;
        let (data, parity) = encode(d);
        for bit in 0u32..64 {
            let corrupted = data ^ (1u64 << bit);
            let (decoded, outcome) = decode(corrupted, parity);
            assert_eq!(decoded, d, "bit={bit}");
            assert!(matches!(outcome, DecodeOutcome::Corrected { .. }), "bit={bit}");
        }
    }

    #[test]
    fn corrects_single_bit_in_parity() {
        let d = 0xDEAD_BEEF_CAFE_BABE;
        let (data, parity) = encode(d);
        for bit in 0u32..8 {
            let corrupted = parity ^ (1u8 << bit);
            let (decoded, outcome) = decode(data, corrupted);
            assert_eq!(decoded, d);
            assert!(matches!(outcome, DecodeOutcome::Corrected { .. }), "bit={bit}");
        }
    }

    #[test]
    fn detects_double_bit_in_data() {
        let d = 0xDEAD_BEEF_CAFE_BABE;
        let (data, parity) = encode(d);
        for b1 in 0u32..63 {
            for b2 in (b1 + 1)..64 {
                let corrupted = data ^ (1u64 << b1) ^ (1u64 << b2);
                let (_, outcome) = decode(corrupted, parity);
                assert_eq!(outcome, DecodeOutcome::Uncorrectable, "b1={b1} b2={b2}");
            }
        }
    }
}
