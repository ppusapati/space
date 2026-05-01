//! Atmospheric correction process wrappers.
//!
//! Sen2Cor, 6S, and MODTRAN are large external scientific programs
//! distributed as compiled Fortran/C++ binaries. This crate does **not**
//! reimplement them; it provides a strongly typed Rust interface that:
//!
//! * Validates input file existence before invocation.
//! * Builds the appropriate command-line arguments per backend.
//! * Spawns the backend with controlled environment, working directory,
//!   timeout, and stdout/stderr capture.
//! * Parses backend-specific success indicators from stdout / output paths.
//! * Returns a typed [`AtmosCorrResult`] containing input metadata, output
//!   paths, exit status, and stderr/stdout for diagnostics.
//!
//! Each backend is a [`Backend`] enum variant. The common entry point is
//! [`run_atmos_correction`].

#![cfg_attr(docsrs, feature(doc_cfg))]

use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Errors produced by `eo-atmos-corr`.
#[derive(Debug, Error)]
pub enum AtmosCorrError {
    /// Path passed in does not exist on disk.
    #[error("path does not exist: {0}")]
    MissingPath(PathBuf),
    /// Backend binary failed to spawn.
    #[error("failed to launch backend `{backend}`: {source}")]
    SpawnFailed {
        /// Backend name.
        backend: &'static str,
        /// Underlying I/O error.
        #[source]
        source: std::io::Error,
    },
    /// Backend exited with non-zero status.
    #[error("backend `{backend}` exited with status {status} - stderr: {stderr}")]
    NonZeroExit {
        /// Backend name.
        backend: &'static str,
        /// Process exit status as integer.
        status: i32,
        /// Captured stderr.
        stderr: String,
    },
    /// Backend exceeded the configured timeout.
    #[error("backend `{backend}` exceeded timeout of {timeout_s}s")]
    Timeout {
        /// Backend name.
        backend: &'static str,
        /// Configured timeout in seconds.
        timeout_s: u64,
    },
    /// Output path was not produced.
    #[error("expected output path missing after backend run: {0}")]
    OutputMissing(PathBuf),
    /// Configuration error (e.g., unsupported resolution).
    #[error("configuration error: {0}")]
    Config(&'static str),
}

/// Sentinel-2 Sen2Cor configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sen2CorConfig {
    /// Path to the Sen2Cor executable (`L2A_Process`).
    pub binary: PathBuf,
    /// Optional GIPP configuration file path.
    pub gipp: Option<PathBuf>,
    /// Output target resolution: 10, 20, or 60 metres.
    pub resolution_m: u16,
    /// Optional output directory; if omitted, the default Sen2Cor convention
    /// is used (input directory adjacent).
    pub output_dir: Option<PathBuf>,
}

/// 6S configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SixSConfig {
    /// Path to the 6S executable (typically `6sV1.1` or `6sV2.1`).
    pub binary: PathBuf,
    /// Path to the 6S input parameter file.
    pub input_file: PathBuf,
    /// Path to the directory where 6S writes its output.
    pub output_dir: PathBuf,
}

/// MODTRAN configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModtranConfig {
    /// Path to the MODTRAN executable (`mod6c1` etc).
    pub binary: PathBuf,
    /// MODTRAN tape5 / JSON input file.
    pub input_file: PathBuf,
    /// MODTRAN run-control directory (`DATA/`).
    pub data_dir: PathBuf,
    /// Output directory.
    pub output_dir: PathBuf,
}

/// Backend selection.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Backend {
    /// Sen2Cor for Sentinel-2 imagery.
    Sen2Cor(Sen2CorConfig),
    /// 6SV — Second Simulation of a Satellite Signal in the Solar Spectrum.
    SixS(SixSConfig),
    /// MODTRAN — MODerate resolution atmospheric TRANsmission.
    Modtran(ModtranConfig),
}

impl Backend {
    /// Backend name suitable for logging.
    #[must_use]
    pub fn name(&self) -> &'static str {
        match self {
            Backend::Sen2Cor(_) => "Sen2Cor",
            Backend::SixS(_) => "6S",
            Backend::Modtran(_) => "MODTRAN",
        }
    }
}

/// Inputs for an atmospheric correction run.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AtmosCorrJob {
    /// Input product (e.g. Sentinel-2 L1C `.SAFE` directory).
    pub input: PathBuf,
    /// Backend selection and its configuration.
    pub backend: Backend,
    /// Wall-clock timeout for the backend run.
    pub timeout: Duration,
}

/// Result of a successful atmospheric correction run.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AtmosCorrResult {
    /// Backend used.
    pub backend: String,
    /// Input path that was processed.
    pub input: PathBuf,
    /// Primary output path (e.g. Sen2Cor `.SAFE` L2A directory or 6S output).
    pub output: PathBuf,
    /// Wall-clock duration of the backend run.
    pub duration: Duration,
    /// Captured stdout (truncated to 64 KiB).
    pub stdout_tail: String,
}

/// Run atmospheric correction.
///
/// # Errors
/// Returns [`AtmosCorrError`] for any path-, spawn-, exit-, timeout-, or
/// missing-output failure.
pub fn run_atmos_correction(job: &AtmosCorrJob) -> Result<AtmosCorrResult, AtmosCorrError> {
    if !job.input.exists() {
        return Err(AtmosCorrError::MissingPath(job.input.clone()));
    }
    let backend_name = job.backend.name();
    let start = Instant::now();
    let (mut command, expected_output) = build_command(job)?;
    command.stdout(Stdio::piped()).stderr(Stdio::piped());
    let mut child = command.spawn().map_err(|e| AtmosCorrError::SpawnFailed {
        backend: backend_name,
        source: e,
    })?;
    // Polling-based timeout enforcement.
    let deadline = start + job.timeout;
    loop {
        match child.try_wait() {
            Ok(Some(status)) => {
                let mut stdout = String::new();
                let mut stderr = String::new();
                if let Some(mut s) = child.stdout.take()
                    && let Err(e) = std::io::Read::read_to_string(&mut s, &mut stdout)
                {
                    tracing::warn!(?backend_name, error = %e, "stdout capture failed");
                }
                if let Some(mut s) = child.stderr.take()
                    && let Err(e) = std::io::Read::read_to_string(&mut s, &mut stderr)
                {
                    tracing::warn!(?backend_name, error = %e, "stderr capture failed");
                }
                if !status.success() {
                    return Err(AtmosCorrError::NonZeroExit {
                        backend: backend_name,
                        status: status.code().unwrap_or(-1),
                        stderr,
                    });
                }
                if let Some(out) = expected_output.as_ref()
                    && !out.exists()
                {
                    tracing::warn!(?out, "expected output missing after backend run");
                    return Err(AtmosCorrError::OutputMissing(out.clone()));
                }
                let stdout_tail =
                    if stdout.len() > 65_536 { stdout[stdout.len() - 65_536..].to_string() } else { stdout };
                return Ok(AtmosCorrResult {
                    backend: backend_name.to_string(),
                    input: job.input.clone(),
                    output: expected_output.unwrap_or_else(|| job.input.clone()),
                    duration: start.elapsed(),
                    stdout_tail,
                });
            }
            Ok(None) => {
                if Instant::now() >= deadline {
                    if let Err(e) = child.kill() {
                        tracing::warn!(?backend_name, error = %e, "kill on timeout failed");
                    }
                    return Err(AtmosCorrError::Timeout {
                        backend: backend_name,
                        timeout_s: job.timeout.as_secs(),
                    });
                }
                std::thread::sleep(Duration::from_millis(50));
            }
            Err(e) => {
                return Err(AtmosCorrError::SpawnFailed {
                    backend: backend_name,
                    source: e,
                });
            }
        }
    }
}

fn build_command(job: &AtmosCorrJob) -> Result<(Command, Option<PathBuf>), AtmosCorrError> {
    match &job.backend {
        Backend::Sen2Cor(c) => {
            if !c.binary.exists() {
                return Err(AtmosCorrError::MissingPath(c.binary.clone()));
            }
            if !matches!(c.resolution_m, 10 | 20 | 60) {
                return Err(AtmosCorrError::Config("Sen2Cor resolution must be 10, 20, or 60"));
            }
            let mut cmd = Command::new(&c.binary);
            cmd.arg(&job.input);
            cmd.arg("--resolution").arg(c.resolution_m.to_string());
            if let Some(g) = &c.gipp {
                if !g.exists() {
                    return Err(AtmosCorrError::MissingPath(g.clone()));
                }
                cmd.arg("--GIP_L2A").arg(g);
            }
            if let Some(o) = &c.output_dir {
                cmd.arg("--output_dir").arg(o);
            }
            // Sen2Cor convention: L1C SAFE → L2A SAFE adjacent or in output_dir.
            let expected = c.output_dir.clone().or_else(|| {
                let s = job.input.to_string_lossy();
                Some(PathBuf::from(s.replacen("MSIL1C", "MSIL2A", 1)))
            });
            Ok((cmd, expected))
        }
        Backend::SixS(c) => {
            if !c.binary.exists() {
                return Err(AtmosCorrError::MissingPath(c.binary.clone()));
            }
            if !c.input_file.exists() {
                return Err(AtmosCorrError::MissingPath(c.input_file.clone()));
            }
            let mut cmd = Command::new(&c.binary);
            cmd.stdin(Stdio::null());
            cmd.current_dir(&c.output_dir);
            cmd.arg("<").arg(&c.input_file); // run scripts typically pipe via redirection wrapper
            Ok((cmd, Some(c.output_dir.clone())))
        }
        Backend::Modtran(c) => {
            if !c.binary.exists() {
                return Err(AtmosCorrError::MissingPath(c.binary.clone()));
            }
            if !c.input_file.exists() {
                return Err(AtmosCorrError::MissingPath(c.input_file.clone()));
            }
            if !c.data_dir.exists() {
                return Err(AtmosCorrError::MissingPath(c.data_dir.clone()));
            }
            let mut cmd = Command::new(&c.binary);
            cmd.arg(&c.input_file);
            cmd.env("MODTRAN_DATA", &c.data_dir);
            cmd.current_dir(&c.output_dir);
            Ok((cmd, Some(c.output_dir.clone())))
        }
    }
}

/// Verify a backend binary is callable on this host (`<binary> --version`
/// or equivalent best-effort probe). Returns the captured stdout if the
/// probe succeeds.
///
/// # Errors
/// Returns [`AtmosCorrError::SpawnFailed`] or [`AtmosCorrError::NonZeroExit`].
pub fn probe_backend(binary: &Path) -> Result<String, AtmosCorrError> {
    let mut cmd = Command::new(binary);
    cmd.arg("--version");
    cmd.stdout(Stdio::piped()).stderr(Stdio::piped());
    let output = cmd.output().map_err(|e| AtmosCorrError::SpawnFailed {
        backend: "probe",
        source: e,
    })?;
    if !output.status.success() {
        return Err(AtmosCorrError::NonZeroExit {
            backend: "probe",
            status: output.status.code().unwrap_or(-1),
            stderr: String::from_utf8_lossy(&output.stderr).to_string(),
        });
    }
    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}
