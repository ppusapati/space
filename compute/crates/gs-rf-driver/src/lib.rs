//! Ground-station RF equipment driver registry (GS-FR-011).
//!
//! Models the typical ground-station RF chain — Low-Noise Amplifier
//! (LNA), Solid-State Power Amplifier (SSPA), frequency synthesizer,
//! and IF down-converter — behind a [`Transport`] trait so the same
//! drivers can be used over SNMP, Modbus/TCP, or a unit-test mock.
//!
//! Each device exposes a strongly-typed setter / getter API expressed
//! as opaque [`Register`] keys. Concrete transports translate the
//! `Register` into the wire-level OID (SNMP) or holding-register
//! address (Modbus).
//!
//! The actual SNMP and Modbus libraries that bind to network sockets
//! are deliberately not part of this crate — production deployments
//! pull in `tokio-modbus` and `snmp2` and implement [`Transport`]
//! against them. This crate provides the device-level command
//! semantics, range validation, and a [`MockTransport`] for tests.

#![cfg_attr(docsrs, feature(doc_cfg))]

use std::collections::BTreeMap;

use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Errors produced by `gs-rf-driver`.
#[derive(Debug, Error)]
pub enum DriverError {
    /// Transport-level error.
    #[error("transport error: {0}")]
    Transport(String),
    /// Out-of-range setpoint.
    #[error("parameter `{name}` value {value} is out of range {range}")]
    OutOfRange {
        /// Field name.
        name: &'static str,
        /// Offending value.
        value: f64,
        /// Admissible range.
        range: &'static str,
    },
    /// Register not present in this device's MIB / register map.
    #[error("unknown register: {0:?}")]
    UnknownRegister(Register),
}

/// Logical register identifier. The transport maps these to wire-level
/// OIDs / addresses.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Serialize, Deserialize)]
pub enum Register {
    /// Centre frequency (Hz).
    CentreFrequencyHz,
    /// Output power (dBm).
    OutputPowerDbm,
    /// LNA gain (dB).
    LnaGainDb,
    /// Mute / un-mute.
    Mute,
    /// Health / fault summary.
    HealthFlags,
}

/// Generic transport over which device registers are read / written.
pub trait Transport {
    /// Read a register and return the current 32-bit value (caller may
    /// reinterpret as float / int / boolean).
    ///
    /// # Errors
    /// Propagates implementation-specific errors as [`DriverError::Transport`].
    fn read_u64(&mut self, reg: Register) -> Result<u64, DriverError>;
    /// Write a register.
    ///
    /// # Errors
    /// As [`Transport::read_u64`].
    fn write_u64(&mut self, reg: Register, value: u64) -> Result<(), DriverError>;
}

/// In-memory transport useful for unit tests.
#[derive(Debug, Default, Clone)]
pub struct MockTransport {
    /// Current register values.
    pub registers: BTreeMap<Register, u64>,
}

impl Transport for MockTransport {
    fn read_u64(&mut self, reg: Register) -> Result<u64, DriverError> {
        self.registers.get(&reg).copied().ok_or(DriverError::UnknownRegister(reg))
    }
    fn write_u64(&mut self, reg: Register, value: u64) -> Result<(), DriverError> {
        self.registers.insert(reg, value);
        Ok(())
    }
}

/// LNA driver.
pub struct Lna<'a, T: Transport> {
    /// Transport.
    pub transport: &'a mut T,
    /// Maximum allowed gain (dB).
    pub max_gain_db: f64,
}

impl<'a, T: Transport> Lna<'a, T> {
    /// Set LNA gain.
    ///
    /// # Errors
    /// [`DriverError::OutOfRange`] if `gain_db` outside `[0, max_gain_db]`.
    pub fn set_gain_db(&mut self, gain_db: f64) -> Result<(), DriverError> {
        if !(0.0..=self.max_gain_db).contains(&gain_db) {
            return Err(DriverError::OutOfRange {
                name: "gain_db",
                value: gain_db,
                range: "[0, max_gain_db]",
            });
        }
        self.transport.write_u64(Register::LnaGainDb, gain_db.to_bits())
    }

    /// Read the LNA gain.
    ///
    /// # Errors
    /// As [`Transport::read_u64`].
    pub fn gain_db(&mut self) -> Result<f64, DriverError> {
        let raw = self.transport.read_u64(Register::LnaGainDb)?;
        Ok(f64::from_bits(raw))
    }
}

/// SSPA driver.
pub struct Sspa<'a, T: Transport> {
    /// Transport.
    pub transport: &'a mut T,
    /// Maximum output power (dBm).
    pub max_output_dbm: f64,
}

impl<'a, T: Transport> Sspa<'a, T> {
    /// Set output power.
    ///
    /// # Errors
    /// [`DriverError::OutOfRange`] if outside `[-30, max_output_dbm]`.
    pub fn set_output_dbm(&mut self, dbm: f64) -> Result<(), DriverError> {
        if !(-30.0..=self.max_output_dbm).contains(&dbm) {
            return Err(DriverError::OutOfRange {
                name: "output_dbm",
                value: dbm,
                range: "[-30, max_output_dbm]",
            });
        }
        self.transport.write_u64(Register::OutputPowerDbm, dbm.to_bits())
    }

    /// Mute / un-mute the amplifier.
    ///
    /// # Errors
    /// As [`Transport::write_u64`].
    pub fn set_mute(&mut self, mute: bool) -> Result<(), DriverError> {
        self.transport.write_u64(Register::Mute, u64::from(mute))
    }
}

/// Frequency synthesizer driver.
pub struct Synthesizer<'a, T: Transport> {
    /// Transport.
    pub transport: &'a mut T,
    /// Minimum tuneable frequency (Hz).
    pub min_hz: f64,
    /// Maximum tuneable frequency (Hz).
    pub max_hz: f64,
    /// Step size (Hz).
    pub step_hz: f64,
}

impl<'a, T: Transport> Synthesizer<'a, T> {
    /// Set the centre frequency.
    ///
    /// # Errors
    /// [`DriverError::OutOfRange`] if outside the configured range or
    /// not a multiple of `step_hz`.
    pub fn set_centre_hz(&mut self, hz: f64) -> Result<(), DriverError> {
        if !(self.min_hz..=self.max_hz).contains(&hz) {
            return Err(DriverError::OutOfRange {
                name: "frequency_hz",
                value: hz,
                range: "[min_hz, max_hz]",
            });
        }
        if (hz % self.step_hz).abs() > 1e-3 {
            return Err(DriverError::OutOfRange {
                name: "frequency_hz",
                value: hz,
                range: "multiple of step_hz",
            });
        }
        self.transport.write_u64(Register::CentreFrequencyHz, hz as u64)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn mock() -> MockTransport {
        MockTransport::default()
    }

    #[test]
    fn lna_set_get_round_trip() {
        let mut t = mock();
        {
            let mut l = Lna { transport: &mut t, max_gain_db: 50.0 };
            l.set_gain_db(35.0).unwrap();
            assert!((l.gain_db().unwrap() - 35.0).abs() < 1e-9);
        }
    }

    #[test]
    fn lna_rejects_out_of_range_gain() {
        let mut t = mock();
        let mut l = Lna { transport: &mut t, max_gain_db: 30.0 };
        assert!(l.set_gain_db(40.0).is_err());
        assert!(l.set_gain_db(-1.0).is_err());
    }

    #[test]
    fn sspa_mute_writes_register() {
        let mut t = mock();
        {
            let mut s = Sspa { transport: &mut t, max_output_dbm: 50.0 };
            s.set_mute(true).unwrap();
        }
        assert_eq!(t.read_u64(Register::Mute).unwrap(), 1);
    }

    #[test]
    fn synth_rejects_off_grid_frequency() {
        let mut t = mock();
        let mut s = Synthesizer {
            transport: &mut t,
            min_hz: 1e9,
            max_hz: 2e9,
            step_hz: 1_000.0,
        };
        // 1.000 000 5 GHz is half a kHz off the 1 kHz grid.
        assert!(s.set_centre_hz(1_000_000_500.5).is_err());
        // On grid → ok.
        s.set_centre_hz(1_500_000_000.0).unwrap();
    }
}
