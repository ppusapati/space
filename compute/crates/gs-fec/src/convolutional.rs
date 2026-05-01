//! Rate-1/2 K=7 convolutional encoder + Viterbi decoder.
//!
//! NASA standard polynomials: `g₀ = 0o171 = 0b1111001`,
//! `g₁ = 0o133 = 0b1011011`. These are the same polynomials used by
//! Voyager, ISEE, and many CCSDS missions.

const G0: u8 = 0b111_1001; // 0o171
const G1: u8 = 0b101_1011; // 0o133
const STATES: usize = 64;
const K_LEN: usize = 6; // memory bits

/// Encode a bit stream. The encoder is initialised in the all-zero
/// state; trailing zero bits should be appended by the caller to
/// terminate the trellis.
#[must_use]
pub fn encode(bits: &[u8]) -> Vec<u8> {
    let mut state: u8 = 0;
    let mut out = Vec::with_capacity(bits.len() * 2);
    for &b in bits {
        let bit = b & 1;
        let reg = (u16::from(state) << 1) | u16::from(bit);
        let v0 = parity_u16(reg & u16::from(G0));
        let v1 = parity_u16(reg & u16::from(G1));
        out.push(v0);
        out.push(v1);
        state = (reg & 0x3F) as u8;
    }
    out
}

#[inline]
fn parity_u16(mut x: u16) -> u8 {
    x ^= x >> 8;
    x ^= x >> 4;
    x ^= x >> 2;
    x ^= x >> 1;
    (x & 1) as u8
}

/// Hard-decision Viterbi decoder. Returns the decoded bit stream
/// (excluding the K-1 termination bits).
#[must_use]
pub fn viterbi_decode(received: &[u8]) -> Vec<u8> {
    let n_bits = received.len() / 2;
    if n_bits == 0 {
        return Vec::new();
    }
    let mut path_metric = [u32::MAX; STATES];
    path_metric[0] = 0;
    let mut history: Vec<[u8; STATES]> = Vec::with_capacity(n_bits);

    for step in 0..n_bits {
        let r0 = received[step * 2];
        let r1 = received[step * 2 + 1];
        let mut next_metric = [u32::MAX; STATES];
        let mut chosen_prev = [0_u8; STATES];
        for prev in 0..STATES {
            if path_metric[prev] == u32::MAX {
                continue;
            }
            for bit in 0..=1_u8 {
                let reg = (prev << 1 | usize::from(bit)) as u16;
                let v0 = parity_u16(reg & u16::from(G0));
                let v1 = parity_u16(reg & u16::from(G1));
                let branch = u32::from(v0 ^ r0) + u32::from(v1 ^ r1);
                let new_state = ((prev as u16) << 1 & 0x3F) | u16::from(bit);
                let new_state_us = new_state as usize;
                let cand = path_metric[prev].saturating_add(branch);
                if cand < next_metric[new_state_us] {
                    next_metric[new_state_us] = cand;
                    chosen_prev[new_state_us] = prev as u8;
                }
            }
        }
        path_metric = next_metric;
        history.push(chosen_prev);
    }

    // Find the best terminal state (the encoder is typically terminated
    // back to state 0, but we accept the lowest-metric state in general).
    let mut best_state = 0_usize;
    let mut best_metric = u32::MAX;
    for (i, &m) in path_metric.iter().enumerate() {
        if m < best_metric {
            best_metric = m;
            best_state = i;
        }
    }
    let mut out = Vec::with_capacity(n_bits);
    let mut state = best_state;
    for step in (0..n_bits).rev() {
        out.push((state & 1) as u8);
        state = history[step][state] as usize;
    }
    out.reverse();
    // Strip trailing K_LEN termination bits if the caller appended them.
    if out.len() > K_LEN { out[..out.len() - K_LEN].to_vec() } else { out }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn round_trip_clean_channel() {
        let info: Vec<u8> = b"\x01\x02\x03\x04\x05\x06\x07\x08"
            .iter()
            .flat_map(|byte| (0..8).rev().map(move |i| (byte >> i) & 1))
            .collect();
        // Append K-1 = 6 zero bits to terminate the trellis.
        let mut padded = info.clone();
        for _ in 0..K_LEN {
            padded.push(0);
        }
        let coded = encode(&padded);
        let decoded = viterbi_decode(&coded);
        assert_eq!(decoded, info);
    }

    #[test]
    fn corrects_isolated_bit_errors() {
        let info: Vec<u8> = (0..32).map(|i| (i & 1) as u8).collect();
        let mut padded = info.clone();
        for _ in 0..K_LEN {
            padded.push(0);
        }
        let mut coded = encode(&padded);
        // Flip one symbol in the middle.
        coded[20] ^= 1;
        let decoded = viterbi_decode(&coded);
        assert_eq!(decoded, info);
    }
}
