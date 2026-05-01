//! Memory scrubber.
//!
//! A scrubber walks an array of `(data, parity)` 72-bit words, decodes
//! each, and rewrites it. Single-bit errors are corrected in place;
//! double-bit errors are counted and reported but not rewritten so that
//! upstream consumers can take corrective action.

use crate::{DecodeOutcome, secded64};

/// Statistics produced by [`scrub`].
#[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
pub struct ScrubStats {
    /// Words scanned.
    pub words: u64,
    /// Words found clean.
    pub clean: u64,
    /// Words with a single-bit error that was corrected.
    pub corrected: u64,
    /// Words with an uncorrectable double-bit error.
    pub uncorrectable: u64,
}

/// Walk a slice of `(data, parity)` words, correcting single-bit errors
/// in place. Returns aggregate statistics.
#[must_use]
pub fn scrub(words: &mut [(u64, u8)]) -> ScrubStats {
    let mut stats = ScrubStats::default();
    for w in words.iter_mut() {
        stats.words += 1;
        let (decoded, outcome) = secded64::decode(w.0, w.1);
        match outcome {
            DecodeOutcome::NoError => stats.clean += 1,
            DecodeOutcome::Corrected { .. } => {
                let (data, parity) = secded64::encode(decoded);
                w.0 = data;
                w.1 = parity;
                stats.corrected += 1;
            }
            DecodeOutcome::Uncorrectable => stats.uncorrectable += 1,
        }
    }
    stats
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::secded64;

    #[test]
    fn scrub_repairs_random_single_bit_flips() {
        let originals = [
            0xDEAD_BEEF_CAFE_BABE_u64,
            0x0123_4567_89AB_CDEF_u64,
            0xFFFF_FFFF_FFFF_FFFF_u64,
        ];
        let mut words: Vec<(u64, u8)> = originals.iter().map(|&d| secded64::encode(d)).collect();
        // Flip bit 5 in word 0 and bit 60 in word 2.
        words[0].0 ^= 1u64 << 5;
        words[2].0 ^= 1u64 << 60;
        let stats = scrub(&mut words);
        assert_eq!(stats.words, 3);
        assert_eq!(stats.corrected, 2);
        assert_eq!(stats.clean, 1);
        assert_eq!(stats.uncorrectable, 0);
        // Decoded values must match originals.
        for (i, (data, parity)) in words.iter().enumerate() {
            let (decoded, outcome) = secded64::decode(*data, *parity);
            assert_eq!(decoded, originals[i]);
            assert_eq!(outcome, DecodeOutcome::NoError);
        }
    }

    #[test]
    fn scrub_counts_uncorrectable() {
        let mut words: Vec<(u64, u8)> = vec![secded64::encode(0xAA)];
        words[0].0 ^= 0b11; // double-bit
        let stats = scrub(&mut words);
        assert_eq!(stats.uncorrectable, 1);
        assert_eq!(stats.corrected, 0);
    }
}
