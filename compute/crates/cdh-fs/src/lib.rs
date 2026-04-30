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
/// `[tag(1)][_reserved(1)][file_id(4)][len(4)][crc(2)]`
const HEADER_LEN: usize = 12;

/// Recovery statistics produced by [`Filesystem::open`].
#[derive(Debug, Default, Clone, Copy)]
pub struct RecoveryStats {
    /// Records loaded successfully.
    pub valid_records: u32,
    /// Records skipped due to CRC failure.
    pub corrupted_records: u32,
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
    /// # Errors
    /// [`FsError::InvalidRecord`] only if the very first byte is malformed.
    pub fn open(buffer: Vec<u8>) -> Result<Self, FsError> {
        let mut fs = Self {
            tail: 0,
            buffer,
            index: BTreeMap::new(),
            recovery: RecoveryStats::default(),
        };
        let mut offset = 0;
        loop {
            if offset + HEADER_LEN > fs.buffer.len() {
                break;
            }
            // A "fully zero" header marks end-of-log.
            if fs.buffer[offset..offset + HEADER_LEN].iter().all(|&b| b == 0) {
                break;
            }
            let tag_byte = fs.buffer[offset];
            let len = u32::from_be_bytes(
                fs.buffer[offset + 6..offset + 10].try_into().unwrap(),
            ) as usize;
            let crc_received = u16::from_be_bytes(
                fs.buffer[offset + 10..offset + 12].try_into().unwrap(),
            );
            if Tag::from_byte(tag_byte).is_none() {
                fs.recovery.corrupted_records += 1;
                offset += HEADER_LEN;
                continue;
            }
            if offset + HEADER_LEN + len > fs.buffer.len() {
                fs.recovery.corrupted_records += 1;
                break;
            }
            let mut crc_input = Vec::with_capacity(HEADER_LEN - 2 + len);
            crc_input.extend_from_slice(&fs.buffer[offset..offset + HEADER_LEN - 2]);
            crc_input.extend_from_slice(
                &fs.buffer[offset + HEADER_LEN..offset + HEADER_LEN + len],
            );
            let crc_computed = crc16_ccitt(&crc_input);
            if crc_computed != crc_received {
                fs.recovery.corrupted_records += 1;
                offset += HEADER_LEN + len;
                continue;
            }
            // Update index.
            let file_id = u32::from_be_bytes(
                fs.buffer[offset + 2..offset + 6].try_into().unwrap(),
            );
            match Tag::from_byte(tag_byte).unwrap() {
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
        fs.tail = offset;
        Ok(fs)
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
    pub fn compact(&mut self) {
        let mut new_buffer = vec![0_u8; self.buffer.len()];
        let mut new_index = BTreeMap::new();
        let mut new_tail = 0_usize;
        // Snapshot live records so we can copy from the old buffer.
        let snapshot: Vec<(u32, Vec<u8>)> = self
            .index
            .iter()
            .map(|(&id, &(off, len))| (id, self.buffer[off..off + len].to_vec()))
            .collect();
        for (file_id, data) in snapshot {
            // Build header + payload.
            let total = HEADER_LEN + data.len();
            if new_tail + total > new_buffer.len() {
                // Compaction failed to fit; abort by leaving the old state intact.
                return;
            }
            let mut hdr = [0_u8; HEADER_LEN];
            hdr[0] = Tag::Write as u8;
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
        }
        self.buffer = new_buffer;
        self.index = new_index;
        self.tail = new_tail;
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
        hdr[0] = tag as u8;
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
        // Corrupt the second record's payload (after first 17 bytes).
        let buffer = fs.buffer().to_vec();
        let mut buffer = buffer;
        buffer[HEADER_LEN + 5 + HEADER_LEN] ^= 0xFF;
        let fs2 = Filesystem::open(buffer).unwrap();
        assert_eq!(fs2.recovery.valid_records, 1);
        assert_eq!(fs2.recovery.corrupted_records, 1);
        assert_eq!(fs2.read(1).unwrap(), b"alpha");
        assert!(matches!(fs2.read(2).unwrap_err(), FsError::NotFound(_)));
    }

    #[test]
    fn compact_drops_obsolete_versions() {
        let mut fs = Filesystem::new(4096);
        fs.write(1, b"v1").unwrap();
        fs.write(1, b"v2_longer").unwrap();
        fs.write(2, b"keep").unwrap();
        fs.delete(1).unwrap();
        let before = fs.used_bytes();
        fs.compact();
        let after = fs.used_bytes();
        assert!(after < before, "compaction should reclaim space: {before} → {after}");
        assert!(matches!(fs.read(1).unwrap_err(), FsError::NotFound(_)));
        assert_eq!(fs.read(2).unwrap(), b"keep");
    }
}
