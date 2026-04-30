//! Bit and frame synchronization (GS-FR-030).
//!
//! * [`BitSync`] — early-late gate detector that estimates the optimal
//!   sample point within a symbol given a stream of NRZ-L baseband
//!   samples and a known bit period.
//! * [`FrameSync`] — sliding-window correlator that finds the CCSDS
//!   Attached Sync Marker (`0x1ACFFC1D`, big-endian) in a hard-decision
//!   bit stream. Tolerates `max_errors` bit slips for noisy links.

#![cfg_attr(docsrs, feature(doc_cfg))]

/// CCSDS Attached Sync Marker (CCSDS 132.0-B-2).
pub const ASM: u32 = 0x1ACF_FC1D;

/// Early-late gate bit synchronizer.
///
/// Maintains a phase estimate `phase ∈ [0, samples_per_bit)`. For each
/// `samples_per_bit` chunk of input samples, the gate compares the
/// energy in the early half (samples before `phase`) and late half
/// (samples after `phase`) and adjusts `phase` to reduce the asymmetry.
#[derive(Debug, Clone)]
pub struct BitSync {
    /// Samples per bit (≥ 2).
    pub samples_per_bit: usize,
    /// Current phase estimate in samples.
    pub phase: f64,
    /// Loop gain in `[0, 1]`.
    pub loop_gain: f64,
}

impl BitSync {
    /// Construct a bit synchronizer.
    #[must_use]
    pub fn new(samples_per_bit: usize, loop_gain: f64) -> Self {
        Self { samples_per_bit, phase: 0.0, loop_gain }
    }

    /// Process `samples_per_bit` baseband samples and return the
    /// hard-decision bit (`0` or `1`) plus the updated phase estimate.
    /// Samples are assumed real-valued NRZ-L (positive = `1`, negative = `0`).
    #[must_use]
    pub fn step(&mut self, samples: &[f64]) -> u8 {
        debug_assert_eq!(samples.len(), self.samples_per_bit);
        let n = self.samples_per_bit as f64;
        // Energy in early/late halves around the current phase.
        let mid = self.phase + n * 0.5;
        let mut early = 0.0_f64;
        let mut late = 0.0_f64;
        for (i, &s) in samples.iter().enumerate() {
            let pos = i as f64;
            if pos < mid {
                early += s * s;
            } else {
                late += s * s;
            }
        }
        let err = late - early;
        // Update phase via PI loop (only the integrator term here).
        self.phase = (self.phase + self.loop_gain * err.signum() * 0.1).rem_euclid(n);
        // Hard decision at the centre of the symbol relative to phase.
        let center = (self.phase + n * 0.5) as usize % self.samples_per_bit;
        u8::from(samples[center] >= 0.0)
    }
}

/// Sliding-window correlator that scans a hard-decision bit stream for
/// the CCSDS attached sync marker.
#[derive(Debug, Clone)]
pub struct FrameSync {
    /// Marker pattern (typically [`ASM`]).
    pub marker: u32,
    /// Maximum allowed Hamming-distance bit errors.
    pub max_errors: u32,
}

impl FrameSync {
    /// Construct with the canonical CCSDS ASM (0x1ACFFC1D).
    #[must_use]
    pub fn ccsds(max_errors: u32) -> Self {
        Self { marker: ASM, max_errors }
    }

    /// Locate the marker in `bits` (each element is `0` or `1`). Returns
    /// the bit-index of the first sample of the matched marker, or
    /// `None` if the marker is not found within `max_errors` of any
    /// 32-bit window.
    #[must_use]
    pub fn locate(&self, bits: &[u8]) -> Option<usize> {
        if bits.len() < 32 {
            return None;
        }
        let mut window: u32 = 0;
        for &b in bits.iter().take(32) {
            window = (window << 1) | u32::from(b & 1);
        }
        if (window ^ self.marker).count_ones() <= self.max_errors {
            return Some(0);
        }
        for i in 32..bits.len() {
            window = (window << 1) | u32::from(bits[i] & 1);
            if (window ^ self.marker).count_ones() <= self.max_errors {
                return Some(i + 1 - 32);
            }
        }
        None
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn bit_sync_recovers_clean_pulse() {
        let mut bs = BitSync::new(8, 0.1);
        // Symbol = +1 NRZ-L for 8 samples → bit = 1.
        let samples = vec![1.0_f64; 8];
        let bit = bs.step(&samples);
        assert_eq!(bit, 1);
        // Symbol = -1 → bit = 0.
        let samples = vec![-1.0_f64; 8];
        let bit = bs.step(&samples);
        assert_eq!(bit, 0);
    }

    #[test]
    fn frame_sync_finds_clean_marker() {
        let mut bits = Vec::new();
        // 16 zero bits followed by the ASM bit-by-bit.
        for _ in 0..16 {
            bits.push(0);
        }
        for i in (0..32).rev() {
            bits.push(((ASM >> i) & 1) as u8);
        }
        let fs = FrameSync::ccsds(0);
        assert_eq!(fs.locate(&bits), Some(16));
    }

    #[test]
    fn frame_sync_tolerates_few_errors() {
        let mut bits = Vec::new();
        for _ in 0..10 {
            bits.push(0);
        }
        // Inject 2 bit errors in the marker.
        let mut asm_with_errors = ASM;
        asm_with_errors ^= 0b101;
        for i in (0..32).rev() {
            bits.push(((asm_with_errors >> i) & 1) as u8);
        }
        let fs = FrameSync::ccsds(3);
        assert_eq!(fs.locate(&bits), Some(10));
    }

    #[test]
    fn frame_sync_returns_none_when_marker_absent() {
        let bits = vec![0_u8; 64];
        let fs = FrameSync::ccsds(0);
        assert_eq!(fs.locate(&bits), None);
    }
}
