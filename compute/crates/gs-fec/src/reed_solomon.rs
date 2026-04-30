//! Reed-Solomon `(255, 223)` over `GF(2⁸)` via the upstream `reed-solomon`
//! crate. This is the canonical CCSDS RS code with `t = 16` correctable
//! errors.

use thiserror::Error;

/// Number of message bytes.
pub const K: usize = 223;
/// Number of parity bytes.
pub const PARITY: usize = 32;
/// Maximum correctable errors.
pub const T: usize = 16;
/// Total codeword length.
pub const N: usize = 255;

/// Errors produced by the RS decoder.
#[derive(Debug, Error, PartialEq)]
pub enum RsError {
    /// More errors than `t`; uncorrectable.
    #[error("too many errors to correct")]
    Uncorrectable,
}

/// Encode a message of exactly `K` bytes into an `N`-byte codeword.
#[must_use]
pub fn encode(message: &[u8; K]) -> [u8; N] {
    let encoder = reed_solomon::Encoder::new(PARITY);
    let buf = encoder.encode(message);
    let mut out = [0_u8; N];
    out.copy_from_slice(&buf[..N]);
    out
}

/// Decode a possibly-corrupted codeword in place. Returns the number
/// of byte errors corrected.
///
/// # Errors
/// [`RsError::Uncorrectable`] if more than `T` errors are present.
pub fn decode(codeword: &mut [u8; N]) -> Result<u32, RsError> {
    let decoder = reed_solomon::Decoder::new(PARITY);
    match decoder.correct(codeword.as_mut_slice(), None) {
        Ok(buffer) => {
            let mut errors = 0_u32;
            for (i, &b) in buffer.data().iter().take(K).enumerate() {
                if b != codeword[i] {
                    errors += 1;
                }
            }
            // Rebuild the canonical codeword from the corrected message.
            let mut msg = [0_u8; K];
            msg.copy_from_slice(&buffer.data()[..K]);
            *codeword = encode(&msg);
            Ok(errors)
        }
        Err(_) => Err(RsError::Uncorrectable),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn random_message(seed: u64) -> [u8; K] {
        let mut m = [0_u8; K];
        let mut s = seed;
        for slot in m.iter_mut() {
            s = s.wrapping_mul(6_364_136_223_846_793_005).wrapping_add(1_442_695_040_888_963_407);
            *slot = (s >> 32) as u8;
        }
        m
    }

    #[test]
    fn round_trip_no_errors() {
        let m = random_message(1);
        let mut c = encode(&m);
        let n_corrected = decode(&mut c).unwrap();
        assert_eq!(n_corrected, 0);
        assert_eq!(&c[..K], &m[..]);
    }

    #[test]
    fn corrects_up_to_t_errors() {
        let m = random_message(2);
        let mut c = encode(&m);
        for i in 0..T {
            c[i * 5] ^= 0xA5;
        }
        let _n = decode(&mut c).unwrap();
        assert_eq!(&c[..K], &m[..]);
    }

    #[test]
    fn detects_too_many_errors() {
        let m = random_message(3);
        let mut c = encode(&m);
        for i in 0..(T + 1) {
            c[i * 5] ^= 0xA5;
        }
        assert!(matches!(decode(&mut c), Err(RsError::Uncorrectable)));
    }
}
