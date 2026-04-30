//! Integration tests that exercise the wrapper plumbing using `/bin/sh` and
//! `/bin/true`/`/bin/false` as stand-ins for real backend binaries. These
//! verify the spawn / exit-code / timeout / output-validation paths end to
//! end without requiring Sen2Cor / 6S / MODTRAN to be installed.

use std::os::unix::fs::PermissionsExt;
use std::path::PathBuf;
use std::time::Duration;

use eo_atmos_corr::{
    AtmosCorrError, AtmosCorrJob, Backend, Sen2CorConfig, run_atmos_correction,
};

fn write_script(dir: &std::path::Path, name: &str, body: &str) -> PathBuf {
    let p = dir.join(name);
    std::fs::write(&p, body).unwrap();
    let mut perms = std::fs::metadata(&p).unwrap().permissions();
    perms.set_mode(0o755);
    std::fs::set_permissions(&p, perms).unwrap();
    p
}

#[test]
fn missing_input_returns_error() {
    let tmp = tempfile::tempdir().unwrap();
    let bin = write_script(tmp.path(), "fake.sh", "#!/bin/sh\nexit 0\n");
    let job = AtmosCorrJob {
        input: PathBuf::from("/nonexistent/input"),
        backend: Backend::Sen2Cor(Sen2CorConfig {
            binary: bin,
            gipp: None,
            resolution_m: 10,
            output_dir: None,
        }),
        timeout: Duration::from_secs(5),
    };
    assert!(matches!(run_atmos_correction(&job).unwrap_err(), AtmosCorrError::MissingPath(_)));
}

#[test]
fn non_zero_exit_propagates() {
    let tmp = tempfile::tempdir().unwrap();
    let input_dir = tmp.path().join("S2A_MSIL1C_demo.SAFE");
    std::fs::create_dir(&input_dir).unwrap();
    let bin = write_script(tmp.path(), "fail.sh", "#!/bin/sh\necho oops 1>&2\nexit 7\n");
    let job = AtmosCorrJob {
        input: input_dir,
        backend: Backend::Sen2Cor(Sen2CorConfig {
            binary: bin,
            gipp: None,
            resolution_m: 10,
            output_dir: None,
        }),
        timeout: Duration::from_secs(5),
    };
    let err = run_atmos_correction(&job).unwrap_err();
    match err {
        AtmosCorrError::NonZeroExit { status, .. } => assert_eq!(status, 7),
        other => panic!("unexpected error: {other:?}"),
    }
}

#[test]
fn successful_run_with_output_dir() {
    let tmp = tempfile::tempdir().unwrap();
    let input_dir = tmp.path().join("S2A_MSIL1C_demo.SAFE");
    std::fs::create_dir(&input_dir).unwrap();
    let output_dir = tmp.path().join("out");
    std::fs::create_dir(&output_dir).unwrap();
    let bin = write_script(tmp.path(), "ok.sh", "#!/bin/sh\necho running\nexit 0\n");
    let job = AtmosCorrJob {
        input: input_dir,
        backend: Backend::Sen2Cor(Sen2CorConfig {
            binary: bin,
            gipp: None,
            resolution_m: 20,
            output_dir: Some(output_dir.clone()),
        }),
        timeout: Duration::from_secs(5),
    };
    let result = run_atmos_correction(&job).unwrap();
    assert_eq!(result.backend, "Sen2Cor");
    assert!(result.stdout_tail.contains("running"));
    assert_eq!(result.output, output_dir);
}

#[test]
fn timeout_kills_long_running_backend() {
    let tmp = tempfile::tempdir().unwrap();
    let input_dir = tmp.path().join("S2A_MSIL1C_demo.SAFE");
    std::fs::create_dir(&input_dir).unwrap();
    let bin = write_script(tmp.path(), "sleep.sh", "#!/bin/sh\nsleep 60\n");
    let job = AtmosCorrJob {
        input: input_dir,
        backend: Backend::Sen2Cor(Sen2CorConfig {
            binary: bin,
            gipp: None,
            resolution_m: 10,
            output_dir: None,
        }),
        timeout: Duration::from_millis(200),
    };
    let err = run_atmos_correction(&job).unwrap_err();
    assert!(matches!(err, AtmosCorrError::Timeout { .. }));
}

#[test]
fn rejects_invalid_resolution() {
    let tmp = tempfile::tempdir().unwrap();
    let input_dir = tmp.path().join("S2A_MSIL1C_demo.SAFE");
    std::fs::create_dir(&input_dir).unwrap();
    let bin = write_script(tmp.path(), "ok.sh", "#!/bin/sh\nexit 0\n");
    let job = AtmosCorrJob {
        input: input_dir,
        backend: Backend::Sen2Cor(Sen2CorConfig {
            binary: bin,
            gipp: None,
            resolution_m: 30,
            output_dir: None,
        }),
        timeout: Duration::from_secs(5),
    };
    assert!(matches!(run_atmos_correction(&job).unwrap_err(), AtmosCorrError::Config(_)));
}
