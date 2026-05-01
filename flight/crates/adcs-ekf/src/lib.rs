//! Extended Kalman Filter for spacecraft attitude determination and other
//! linearised state estimation problems.
//!
//! Two filters are provided:
//!
//! * [`Ekf`] — a generic discrete-time linearised EKF parameterised by the
//!   user's [`MotionModel`] and [`MeasurementModel`] traits. Suitable for
//!   any nonlinear dynamics, e.g. SAT-FR-002 attitude estimation with
//!   non-quaternion state representations or other on-board fusion tasks.
//! * [`MultiplicativeEkf`] — a quaternion-based MEKF that estimates the
//!   3-axis attitude error and gyroscope bias `(δθ, b_g)`, propagates the
//!   quaternion non-linearly via gyroscope integration, and resets the
//!   error state after every measurement update. This is the canonical
//!   filter described in *Markley & Crassidis (2014)* "Fundamentals of
//!   Spacecraft Attitude Determination and Control".
//!
//! Both filters use [`nalgebra`] dynamic matrices (`DVector` / `DMatrix`)
//! so users do not need to fix the state size at compile time.

#![cfg_attr(docsrs, feature(doc_cfg))]

pub mod generic;
pub mod mekf;

pub use generic::{Ekf, EkfError, MeasurementModel, MotionModel};
pub use mekf::{MekfError, MekfState, MultiplicativeEkf, VectorObservation};
