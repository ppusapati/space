//! On-board log-structured file system.
//!
//! Implements SAT-FR-011: an append-only journal of records protected by
//! per-record CRC-CCITT and indexed by 32-bit file ID. Records are tagged
//! `Write` (insert/replace) or `Delete`, allowing users to overwrite or
//! remove a file by appending a new record. A garbage-collection pass
//! ([`Filesystem::compact`]) rewrites the journal keeping only the
//! latest live record per file.
//!
//! Recovery from media corruption: on [`Filesystem::open`] every record
//! is CRC-validated; corrupted records are skipped and counted, allowing
//! upstream code to surface the event via housekeeping telemetry while
//! preserving the rest of the journal.
//!
//! Storage abstraction is a `Vec<u8>`-style buffer; on real flight
//! hardware this would back to NAND / NOR through a HAL layer.

#![cfg_attr(docsrs, feature(doc_cfg))]

use cdh_ccsds::crc::crc16_ccitt;
use std::collections::BTreeMap;
use thiserror::Error;

/// Errors produced by `cdh-fs`.
#[derive(Debug, Error)]
pub enum FsError {
    /// Backing buffer was truncated or malformed.
    #[error("invalid record at offset {offset}: {reason}")]
    InvalidRecord {
        /// Byte offset within the journal.
        offset: usize,
        /// Diagnostic.
        reason: &'static str,
    },
    /// Requested file ID does not exist (after applying tombstones).
    #[error("file id {0} not found")]
    NotFound(u32),
    /// Backing buffer is full.
    #[error("journal full: capacity {capacity}, requested {requested}")]
    OutOfSpace {
        /// Buffer capacity.
        capacity: usize,
        /// Bytes that would have been written.
        requested: usize,
    },
    /// Compaction could not fit the live set into the buffer; old state
    /// is preserved.
    #[error(
        "compaction cannot fit live set: required {required} bytes, capacity {capacity} bytes"
    )]
    CompactionWouldOverflow {
        /// Bytes required for the live set after compaction.
        required: usize,
        /// Buffer capacity.
        capacity: usize,
    },
}

/// Record tag.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
enum Tag {
    Write = 0x01,
    Delete = 0x02,
}

impl Tag {
    fn from_byte(b: u8) -> Option<Self> {
        match b {
            0x01 => Some(Tag::Write),
            0x02 => Some(Tag::Delete),
            _ => None,
        }
    }
}

/// Journal record header (12 bytes):
/// `[magic_ver(1)][tag(1)][file_id(4)][len(4)][crc(2)]`
///
/// `magic_ver` packs a 4-bit magic nibble (`0xA`) and a 4-bit format
/// version (currently `0x1`), giving a stable on-disk sentinel of
/// `0xA1`. This lets the recovery walker distinguish between:
/// * a valid header (magic nibble `0xA`),
/// * blank/erased flash (`0x00` for NOR-zeroed, `0xFF` for NAND-erased), and
/// * structural corruption (any other byte at the magic position).
const HEADER_LEN: usize = 12;

/// Magic + version byte stored at offset 0 of every record header.
const MAGIC_VERSION: u8 = 0xA1;
const MAGIC_NIBBLE: u8 = 0xA0;
const MAGIC_MASK: u8 = 0xF0;
const VERSION_MASK: u8 = 0x0F;
const SUPPORTED_VERSION: u8 = 0x01;

/// Recovery statistics produced by [`Filesystem::open`].
#[derive(Debug, Default, Clone, Copy)]
pub struct RecoveryStats {
    /// Records loaded successfully.
    pub valid_records: u32,
    /// Records skipped due to CRC failure (header was structurally
    /// valid; payload integrity check failed).
    pub corrupted_records: u32,
    /// Resyncs triggered by an unrecognised header byte. Distinct from
    /// `corrupted_records` because it indicates the record-boundary
    /// stream itself was lost (not just a single payload).
    pub desync_events: u32,
    /// Bytes skipped during resync. Useful for sizing housekeeping
    /// alerts.
    pub bytes_skipped: u32,
}

/// Statistics returned by a successful [`Filesystem::compact`].
#[derive(Debug, Default, Clone, Copy)]
pub struct CompactionStats {
    /// Live records written into the new journal.
    pub records_kept: u32,
    /// Bytes used by the journal after compaction (header + payload).
    pub bytes_used: usize,
    /// Bytes reclaimed (`old_tail - new_tail`).
    pub bytes_reclaimed: usize,
}

/// Read a big-endian `u32` from `buf[offset..offset+4]`. Panics only if
/// the caller violates the slice precondition; callers must ensure
/// `offset + 4 <= buf.len()`.
#[inline]
fn read_u32_be(buf: &[u8], offset: usize) -> u32 {
    u32::from_be_bytes([
        buf[offset],
        buf[offset + 1],
        buf[offset + 2],
        buf[offset + 3],
    ])
}

/// Read a big-endian `u16` from `buf[offset..offset+2]`.
#[inline]
fn read_u16_be(buf: &[u8], offset: usize) -> u16 {
    u16::from_be_bytes([buf[offset], buf[offset + 1]])
}

/// Returns `true` if every byte of `region` is the same erase-state
/// pattern (`0x00` or `0xFF`). Used to disambiguate "end of journal"
/// from "structural corruption".
fn is_blank(region: &[u8]) -> bool {
    if region.is_empty() {
        return true;
    }
    let first = region[0];
    if first != 0x00 && first != 0xFF {
        return false;
    }
    region.iter().all(|&b| b == first)
}

/// Log-structured filesystem.
pub struct Filesystem {
    buffer: Vec<u8>,
    tail: usize,
    /// `file_id -> (offset_of_payload, length)` for the *latest* live write.
    index: BTreeMap<u32, (usize, usize)>,
    /// Recovery statistics from the last `open`.
    pub recovery: RecoveryStats,
}

impl Filesystem {
    /// Construct an empty filesystem with the given byte capacity.
    #[must_use]
    pub fn new(capacity: usize) -> Self {
        Self {
            buffer: vec![0; capacity],
            tail: 0,
            index: BTreeMap::new(),
            recovery: RecoveryStats::default(),
        }
    }

    /// Open an existing journal in `buffer`. Walks the journal record by
    /// record, rebuilds the index, and counts corrupt records.
    ///
    /// Boundary recovery model:
    /// * A header is recognised by its magic nibble (`0xA`) in the high
    ///   four bits of byte 0. The low four bits encode the format
    ///   version; only [`SUPPORTED_VERSION`] is accepted.
    /// * If the magic byte is missing and the remainder of the journal
    ///   is uniformly `0x00` or `0xFF`, the walker treats it as blank
    ///   media (clean end-of-log).
    /// * Otherwise the walker enters resync: it scans forward
    ///   byte-by-byte for the next plausible header whose CRC also
    ///   verifies, counting `desync_events` and `bytes_skipped`.
    ///
    /// # Errors
    /// Currently never returns an error — even malformed input is
    /// surfaced via [`RecoveryStats`]. The signature is preserved for
    /// future hard-fail policies.
    pub fn open(buffer: Vec<u8>) -> Result<Self, FsError> {
        let mut fs = Self {
            tail: 0,
            buffer,
            index: BTreeMap::new(),
            recovery: RecoveryStats::default(),
        };
        let mut offset = 0;
        while offset + HEADER_LEN <= fs.buffer.len() {
            match fs.try_parse_record(offset) {
                ParseOutcome::Valid { tag, file_id, len } => {
                    match tag {
                        Tag::Write => {
                            fs.index.insert(file_id, (offset + HEADER_LEN, len));
                        }
                        Tag::Delete => {
                            fs.index.remove(&file_id);
                        }
                    }
                    fs.recovery.valid_records += 1;
                    offset += HEADER_LEN + len;
                }
                ParseOutcome::EndOfLog => break,
                ParseOutcome::CorruptPayload { len } => {
                    // Header looked structurally valid (magic + tag +
                    // bounded length) but the CRC failed. Skip exactly
                    // one record's worth and resume — record boundaries
                    // are still trustworthy.
                    fs.recovery.corrupted_records += 1;
                    offset += HEADER_LEN + len;
                }
                ParseOutcome::Desync => {
                    // Unknown / out-of-policy header. Scan forward
                    // until we find another plausible header or run
                    // out of buffer.
                    let resync_start = offset;
                    let new_offset = fs.resync_from(offset + 1);
                    fs.recovery.desync_events += 1;
                    fs.recovery.bytes_skipped =
                        fs.recovery.bytes_skipped.saturating_add(
                            (new_offset - resync_start) as u32,
                        );
                    offset = new_offset;
                }
            }
        }
        fs.tail = offset;
        Ok(fs)
    }

    /// Attempt to parse the record beginning at `offset`. Does not
    /// mutate state; only inspects bytes.
    fn try_parse_record(&self, offset: usize) -> ParseOutcome {
        let magic_ver = self.buffer[offset];
        // End-of-log: a non-magic byte where the rest of the buffer is
        // uniform erase-state.
        if magic_ver & MAGIC_MASK != MAGIC_NIBBLE {
            if is_blank(&self.buffer[offset..]) {
                return ParseOutcome::EndOfLog;
            }
            return ParseOutcome::Desync;
        }
        if magic_ver & VERSION_MASK != SUPPORTED_VERSION {
            // Recognised our magic but a future version we cannot
            // parse — treat as a desync rather than guessing.
            return ParseOutcome::Desync;
        }
        let tag_byte = self.buffer[offset + 1];
        let Some(tag) = Tag::from_byte(tag_byte) else {
            return ParseOutcome::Desync;
        };
        let file_id = read_u32_be(&self.buffer, offset + 2);
        let len = read_u32_be(&self.buffer, offset + 6) as usize;
        let crc_received = read_u16_be(&self.buffer, offset + 10);
        if offset + HEADER_LEN + len > self.buffer.len() {
            // Length spans past the buffer — almost certainly a
            // corrupted length field. Treat as desync so we can resync
            // from a later byte rather than blindly advance.
            return ParseOutcome::Desync;
        }
        let mut crc_input = Vec::with_capacity(HEADER_LEN - 2 + len);
        crc_input.extend_from_slice(&self.buffer[offset..offset + HEADER_LEN - 2]);
        crc_input.extend_from_slice(&self.buffer[offset + HEADER_LEN..offset + HEADER_LEN + len]);
        if crc16_ccitt(&crc_input) != crc_received {
            return ParseOutcome::CorruptPayload { len };
        }
        ParseOutcome::Valid { tag, file_id, len }
    }

    /// Scan forward starting at `from` for the next byte that parses
    /// successfully as a record (passing magic, version, tag, length,
    /// and CRC checks). Returns the next offset to resume from — either
    /// the next valid record start or `buffer.len()` if no candidate
    /// remains.
    fn resync_from(&self, from: usize) -> usize {
        let mut probe = from;
        while probe + HEADER_LEN <= self.buffer.len() {
            match self.try_parse_record(probe) {
                ParseOutcome::Valid { .. } | ParseOutcome::EndOfLog => return probe,
                _ => probe += 1,
            }
        }
        self.buffer.len()
    }

    /// Write (or overwrite) a file.
    ///
    /// # Errors
    /// [`FsError::OutOfSpace`] if the journal is full.
    pub fn write(&mut self, file_id: u32, data: &[u8]) -> Result<(), FsError> {
        self.append(Tag::Write, file_id, data)
    }

    /// Delete a file (writes a tombstone record).
    ///
    /// # Errors
    /// [`FsError::OutOfSpace`].
    pub fn delete(&mut self, file_id: u32) -> Result<(), FsError> {
        self.append(Tag::Delete, file_id, &[])
    }

    /// Read the latest contents of a file.
    ///
    /// # Errors
    /// [`FsError::NotFound`] if the file has been deleted or never written.
    pub fn read(&self, file_id: u32) -> Result<&[u8], FsError> {
        let &(offset, len) = self.index.get(&file_id).ok_or(FsError::NotFound(file_id))?;
        Ok(&self.buffer[offset..offset + len])
    }

    /// Number of live files.
    #[must_use]
    pub fn len(&self) -> usize {
        self.index.len()
    }

    /// True if no live files.
    #[must_use]
    pub fn is_empty(&self) -> bool {
        self.index.is_empty()
    }

    /// Bytes used by the journal so far (live + tombstoned + holes).
    #[must_use]
    pub fn used_bytes(&self) -> usize {
        self.tail
    }

    /// Backing buffer total capacity.
    #[must_use]
    pub fn capacity(&self) -> usize {
        self.buffer.len()
    }

    /// Compact the journal: rewrite only the latest live record per file
    /// in a fresh buffer, dropping tombstones and old versions.
    ///
    /// # Errors
    /// [`FsError::CompactionWouldOverflow`] if the live set cannot fit
    /// into the existing capacity. The old journal state is preserved
    /// untouched in that case.
    pub fn compact(&mut self) -> Result<CompactionStats, FsError> {
        let snapshot: Vec<(u32, Vec<u8>)> = self
            .index
            .iter()
            .map(|(&id, &(off, len))| (id, self.buffer[off..off + len].to_vec()))
            .collect();
        let required: usize = snapshot.iter().map(|(_, d)| HEADER_LEN + d.len()).sum();
        if required > self.buffer.len() {
            return Err(FsError::CompactionWouldOverflow {
                required,
                capacity: self.buffer.len(),
            });
        }
        let old_tail = self.tail;
        let mut new_buffer = vec![0_u8; self.buffer.len()];
        let mut new_index = BTreeMap::new();
        let mut new_tail = 0_usize;
        let mut records_kept = 0_u32;
        for (file_id, data) in snapshot {
            let total = HEADER_LEN + data.len();
            let mut hdr = [0_u8; HEADER_LEN];
            hdr[0] = MAGIC_VERSION;
            hdr[1] = Tag::Write as u8;
            hdr[2..6].copy_from_slice(&file_id.to_be_bytes());
            hdr[6..10].copy_from_slice(&(data.len() as u32).to_be_bytes());
            let mut crc_input = Vec::with_capacity(HEADER_LEN - 2 + data.len());
            crc_input.extend_from_slice(&hdr[..HEADER_LEN - 2]);
            crc_input.extend_from_slice(&data);
            let crc = crc16_ccitt(&crc_input);
            hdr[10..12].copy_from_slice(&crc.to_be_bytes());
            new_buffer[new_tail..new_tail + HEADER_LEN].copy_from_slice(&hdr);
            new_buffer[new_tail + HEADER_LEN..new_tail + HEADER_LEN + data.len()]
                .copy_from_slice(&data);
            new_index.insert(file_id, (new_tail + HEADER_LEN, data.len()));
            new_tail += total;
            records_kept += 1;
        }
        self.buffer = new_buffer;
        self.index = new_index;
        self.tail = new_tail;
        Ok(CompactionStats {
            records_kept,
            bytes_used: new_tail,
            bytes_reclaimed: old_tail.saturating_sub(new_tail),
        })
    }

    fn append(&mut self, tag: Tag, file_id: u32, data: &[u8]) -> Result<(), FsError> {
        let total = HEADER_LEN + data.len();
        if self.tail + total > self.buffer.len() {
            return Err(FsError::OutOfSpace {
                capacity: self.buffer.len(),
                requested: total,
            });
        }
        let mut hdr = [0_u8; HEADER_LEN];
        hdr[0] = MAGIC_VERSION;
        hdr[1] = tag as u8;
        hdr[2..6].copy_from_slice(&file_id.to_be_bytes());
        hdr[6..10].copy_from_slice(&(data.len() as u32).to_be_bytes());
        let mut crc_input = Vec::with_capacity(HEADER_LEN - 2 + data.len());
        crc_input.extend_from_slice(&hdr[..HEADER_LEN - 2]);
        crc_input.extend_from_slice(data);
        let crc = crc16_ccitt(&crc_input);
        hdr[10..12].copy_from_slice(&crc.to_be_bytes());
        self.buffer[self.tail..self.tail + HEADER_LEN].copy_from_slice(&hdr);
        self.buffer[self.tail + HEADER_LEN..self.tail + total].copy_from_slice(data);
        match tag {
            Tag::Write => {
                self.index.insert(file_id, (self.tail + HEADER_LEN, data.len()));
            }
            Tag::Delete => {
                self.index.remove(&file_id);
            }
        }
        self.tail += total;
        Ok(())
    }

    /// Borrow the underlying journal buffer (for persistence to flash).
    #[must_use]
    pub fn buffer(&self) -> &[u8] {
        &self.buffer
    }
}

/// Outcome of inspecting a candidate record header.
enum ParseOutcome {
    Valid { tag: Tag, file_id: u32, len: usize },
    /// Header bytes form a recognised structure but the CRC failed.
    /// Boundaries can still be trusted; advance by `HEADER_LEN + len`.
    CorruptPayload { len: usize },
    /// Magic byte missing and the remaining buffer is uniformly blank.
    EndOfLog,
    /// Header byte is not recognised and the buffer is not blank;
    /// caller must enter resync.
    Desync,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn write_read_round_trip() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"hello").unwrap();
        fs.write(2, b"world").unwrap();
        assert_eq!(fs.read(1).unwrap(), b"hello");
        assert_eq!(fs.read(2).unwrap(), b"world");
        assert_eq!(fs.len(), 2);
    }

    #[test]
    fn overwrite_returns_latest() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"first").unwrap();
        fs.write(1, b"second").unwrap();
        assert_eq!(fs.read(1).unwrap(), b"second");
    }

    #[test]
    fn delete_then_read_fails() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"hello").unwrap();
        fs.delete(1).unwrap();
        assert!(matches!(fs.read(1).unwrap_err(), FsError::NotFound(1)));
    }

    #[test]
    fn out_of_space() {
        let mut fs = Filesystem::new(32);
        let big = vec![0_u8; 32];
        assert!(matches!(fs.write(1, &big).unwrap_err(), FsError::OutOfSpace { .. }));
    }

    #[test]
    fn open_recovers_index_and_skips_corrupted_record() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"alpha").unwrap();
        fs.write(2, b"bravo").unwrap();
        // Corrupt the second record's payload (not the magic byte).
        let mut buffer = fs.buffer().to_vec();
        buffer[HEADER_LEN + 5 + HEADER_LEN] ^= 0xFF;
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 1);
        assert_eq!(fs2.recovery.corrupted_records, 1);
        assert_eq!(fs2.recovery.desync_events, 0);
        assert_eq!(fs2.read(1).unwrap(), b"alpha");
        assert!(matches!(fs2.read(2).unwrap_err(), FsError::NotFound(_)));
    }

    #[test]
    fn open_treats_zeroed_tail_as_end_of_log() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"hello").unwrap();
        let buffer = fs.buffer().to_vec();
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 1);
        assert_eq!(fs2.recovery.corrupted_records, 0);
        assert_eq!(fs2.recovery.desync_events, 0);
    }

    #[test]
    fn open_treats_nand_erased_tail_as_end_of_log() {
        let mut fs = Filesystem::new(256);
        fs.write(1, b"hello").unwrap();
        let mut buffer = fs.buffer().to_vec();
        // Simulate NAND erase state (all 0xFF) for the unused tail.
        for b in &mut buffer[fs.used_bytes()..] {
            *b = 0xFF;
        }
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 1);
        assert_eq!(fs2.recovery.desync_events, 0);
    }

    #[test]
    fn open_resyncs_past_zeroed_island_with_valid_record_after() {
        // Write three records, then zero out the second record's
        // magic byte to inject a desync hole that is followed by a
        // valid record. Walker must skip the hole and recover record 3.
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"alpha").unwrap();
        fs.write(2, b"bravo").unwrap();
        fs.write(3, b"charlie").unwrap();
        let mut buffer = fs.buffer().to_vec();
        let r2_start = HEADER_LEN + b"alpha".len();
        // Punch a non-blank hole that breaks magic but is not zeroed
        // (otherwise the walker would interpret it as end-of-log).
        buffer[r2_start] = 0x55;
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 2, "records 1 and 3 recovered");
        assert!(fs2.recovery.desync_events >= 1);
        assert!(fs2.recovery.bytes_skipped >= 1);
        assert_eq!(fs2.read(1).unwrap(), b"alpha");
        assert_eq!(fs2.read(3).unwrap(), b"charlie");
        assert!(matches!(fs2.read(2).unwrap_err(), FsError::NotFound(_)));
    }

    #[test]
    fn open_rejects_unsupported_version_via_resync() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"alpha").unwrap();
        fs.write(2, b"bravo").unwrap();
        let mut buffer = fs.buffer().to_vec();
        // Bump record 2's version nibble to a future version we don't
        // understand. Magic still matches, so this is a "future format"
        // signal — walker treats it as a desync rather than blindly
        // parsing.
        let r2_start = HEADER_LEN + b"alpha".len();
        buffer[r2_start] = MAGIC_NIBBLE | 0x0F;
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 1);
        assert!(fs2.recovery.desync_events >= 1);
    }

    #[test]
    fn compact_returns_stats_and_drops_obsolete_versions() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"v1").unwrap();
        fs.write(1, b"v2_longer").unwrap();
        fs.write(2, b"keep").unwrap();
        fs.delete(1).unwrap();
        let before = fs.used_bytes();
        let stats = fs.compact().unwrap();
        let after = fs.used_bytes();
        assert!(after < before, "compaction should reclaim space: {before} → {after}");
        assert_eq!(stats.records_kept, 1);
        assert_eq!(stats.bytes_used, after);
        assert_eq!(stats.bytes_reclaimed, before - after);
        assert!(matches!(fs.read(1).unwrap_err(), FsError::NotFound(_)));
        assert_eq!(fs.read(2).unwrap(), b"keep");
    }

    #[test]
    fn compact_idempotent_when_no_garbage() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"a").unwrap();
        fs.write(2, b"bb").unwrap();
        let first = fs.compact().unwrap();
        let second = fs.compact().unwrap();
        assert_eq!(first.records_kept, 2);
        assert_eq!(second.records_kept, 2);
        assert_eq!(second.bytes_reclaimed, 0);
    }
}
