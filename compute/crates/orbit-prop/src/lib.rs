//! Orbit propagation utilities.
//!
//! Two complementary propagators are exposed:
//!
//! * [`sgp4_prop`] — wrapper around the well-tested `sgp4` crate that
//!   accepts a NORAD two-line element set and returns position/velocity in
//!   the True-Equator-Mean-Equinox (TEME) frame at any future epoch.
//! * [`integrators`] — fixed-step RK4 and adaptive Dormand-Prince RK45
//!   integrators suitable for two-body or perturbed numerical
//!   propagation when TLE accuracy or model fidelity is insufficient.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod integrators;
pub mod sgp4_prop;

pub use integrators::{rk4, rk45_dopri};
pub use sgp4_prop::{Sgp4Error, Sgp4Propagator, StateTeme};
