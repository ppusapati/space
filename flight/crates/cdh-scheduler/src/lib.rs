//! Time-tagged on-board command scheduler (SAT-FR-012).
//!
//! Maintains a priority queue of scheduled commands keyed by execution
//! epoch. Commands are pulled by [`Scheduler::poll`] when the supplied
//! "now" reaches or passes their scheduled epoch. Ties are broken
//! deterministically by insertion order via a monotonically increasing
//! sequence counter, guaranteeing FIFO behaviour for equal-time entries.

#![cfg_attr(docsrs, feature(doc_cfg))]

use std::cmp::Ordering;
use std::collections::BinaryHeap;

use thiserror::Error;

/// Errors produced by `cdh-scheduler`.
#[derive(Debug, Error, PartialEq)]
pub enum SchedulerError {
    /// Scheduler is full.
    #[error("scheduler full: capacity {0}")]
    Full(usize),
    /// Command epoch is in the past more than the configured horizon.
    #[error("command epoch {epoch} is older than now {now} by more than {horizon} ticks")]
    StaleEpoch {
        /// Provided epoch.
        epoch: u64,
        /// Current time supplied by the caller.
        now: u64,
        /// Configured staleness horizon.
        horizon: u64,
    },
}

/// A scheduled command. The opaque payload type `T` carries whatever the
/// caller wants to associate with the entry — typically a command struct
/// or an opcode/data pair.
#[derive(Debug, Clone)]
pub struct Command<T> {
    /// Execution epoch (monotonic ticks; caller-defined unit).
    pub epoch: u64,
    /// Application-defined identifier or numeric handle.
    pub id: u64,
    /// User payload.
    pub payload: T,
}

#[derive(Debug)]
struct HeapEntry<T> {
    epoch: u64,
    seq: u64,
    cmd: Command<T>,
}

impl<T> Eq for HeapEntry<T> {}
impl<T> PartialEq for HeapEntry<T> {
    fn eq(&self, other: &Self) -> bool {
        self.epoch == other.epoch && self.seq == other.seq
    }
}
impl<T> Ord for HeapEntry<T> {
    fn cmp(&self, other: &Self) -> Ordering {
        // Min-heap on epoch, then seq.
        match other.epoch.cmp(&self.epoch) {
            Ordering::Equal => other.seq.cmp(&self.seq),
            o => o,
        }
    }
}
impl<T> PartialOrd for HeapEntry<T> {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

/// Scheduler.
#[derive(Debug)]
pub struct Scheduler<T> {
    capacity: usize,
    /// Reject commands whose epoch is more than `staleness_horizon` ticks
    /// in the past relative to the caller's `now`. Use [`u64::MAX`] to
    /// disable.
    pub staleness_horizon: u64,
    heap: BinaryHeap<HeapEntry<T>>,
    next_seq: u64,
}

impl<T> Scheduler<T> {
    /// Construct a new scheduler with the given maximum number of pending
    /// commands.
    #[must_use]
    pub fn new(capacity: usize, staleness_horizon: u64) -> Self {
        Self {
            capacity,
            staleness_horizon,
            heap: BinaryHeap::with_capacity(capacity),
            next_seq: 0,
        }
    }

    /// Number of pending commands.
    #[must_use]
    pub fn len(&self) -> usize {
        self.heap.len()
    }

    /// True if the scheduler is empty.
    #[must_use]
    pub fn is_empty(&self) -> bool {
        self.heap.is_empty()
    }

    /// Insert a command.
    ///
    /// # Errors
    /// [`SchedulerError::Full`] if the scheduler is at capacity;
    /// [`SchedulerError::StaleEpoch`] if `cmd.epoch` lies more than
    /// `staleness_horizon` ticks before `now`.
    pub fn insert(&mut self, now: u64, cmd: Command<T>) -> Result<(), SchedulerError> {
        if self.heap.len() >= self.capacity {
            return Err(SchedulerError::Full(self.capacity));
        }
        if self.staleness_horizon != u64::MAX && cmd.epoch < now {
            let lag = now - cmd.epoch;
            if lag > self.staleness_horizon {
                return Err(SchedulerError::StaleEpoch {
                    epoch: cmd.epoch,
                    now,
                    horizon: self.staleness_horizon,
                });
            }
        }
        let seq = self.next_seq;
        self.next_seq = self.next_seq.checked_add(1).unwrap_or(0);
        self.heap.push(HeapEntry { epoch: cmd.epoch, seq, cmd });
        Ok(())
    }

    /// Peek at the next-due command without removing it.
    #[must_use]
    pub fn peek(&self) -> Option<&Command<T>> {
        self.heap.peek().map(|e| &e.cmd)
    }

    /// Pop a command if its epoch is `<= now`. Returns `None` when no
    /// command is yet due (or the queue is empty).
    pub fn poll(&mut self, now: u64) -> Option<Command<T>> {
        if self.heap.peek().is_some_and(|top| top.epoch <= now) {
            self.heap.pop().map(|e| e.cmd)
        } else {
            None
        }
    }

    /// Drain all commands due at or before `now`, preserving FIFO order
    /// for equal-epoch commands.
    pub fn poll_all(&mut self, now: u64) -> Vec<Command<T>> {
        let mut out = Vec::new();
        while let Some(c) = self.poll(now) {
            out.push(c);
        }
        out
    }

    /// Remove a command by `id`. Returns `true` iff a command was
    /// removed.
    pub fn cancel(&mut self, id: u64) -> bool {
        let mut kept: Vec<HeapEntry<T>> = self.heap.drain().collect();
        let before = kept.len();
        kept.retain(|e| e.cmd.id != id);
        let removed = kept.len() < before;
        for e in kept {
            self.heap.push(e);
        }
        removed
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cmd(epoch: u64, id: u64) -> Command<u32> {
        Command { epoch, id, payload: id as u32 }
    }

    #[test]
    fn pop_in_epoch_order() {
        let mut s = Scheduler::new(16, u64::MAX);
        s.insert(0, cmd(30, 3)).unwrap();
        s.insert(0, cmd(10, 1)).unwrap();
        s.insert(0, cmd(20, 2)).unwrap();
        let v = s.poll_all(50);
        assert_eq!(v.iter().map(|c| c.id).collect::<Vec<_>>(), vec![1, 2, 3]);
    }

    #[test]
    fn ties_break_by_insertion_order() {
        let mut s = Scheduler::new(16, u64::MAX);
        for i in 0..5_u64 {
            s.insert(0, cmd(100, i)).unwrap();
        }
        let v = s.poll_all(100);
        assert_eq!(v.iter().map(|c| c.id).collect::<Vec<_>>(), vec![0, 1, 2, 3, 4]);
    }

    #[test]
    fn future_command_is_not_polled() {
        let mut s = Scheduler::new(16, u64::MAX);
        s.insert(0, cmd(100, 1)).unwrap();
        assert!(s.poll(50).is_none());
        assert_eq!(s.poll(100).unwrap().id, 1);
    }

    #[test]
    fn capacity_enforced() {
        let mut s = Scheduler::new(2, u64::MAX);
        s.insert(0, cmd(1, 1)).unwrap();
        s.insert(0, cmd(2, 2)).unwrap();
        assert!(matches!(s.insert(0, cmd(3, 3)).unwrap_err(), SchedulerError::Full(2)));
    }

    #[test]
    fn stale_epoch_rejected() {
        let mut s = Scheduler::new(16, 5);
        let err = s.insert(100, cmd(50, 1)).unwrap_err();
        assert!(matches!(err, SchedulerError::StaleEpoch { .. }));
    }

    #[test]
    fn cancel_removes_specific_id() {
        let mut s = Scheduler::new(16, u64::MAX);
        s.insert(0, cmd(10, 1)).unwrap();
        s.insert(0, cmd(20, 2)).unwrap();
        s.insert(0, cmd(30, 3)).unwrap();
        assert!(s.cancel(2));
        let v = s.poll_all(100);
        assert_eq!(v.iter().map(|c| c.id).collect::<Vec<_>>(), vec![1, 3]);
    }
}
