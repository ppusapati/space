//go:build helm

// Package helm_test drives `helm template` against the
// _chetana-service library chart to assert:
//   1. happy-path values render without error and emit one of every
//      template (Deployment, Service, ServiceAccount, HPA, PDB,
//      NetworkPolicy);
//   2. negative cases are rejected by values.schema.json — missing
//      hpa.enabled / pdb.minAvailable / networkPolicy.ingress fail
//      Helm rendering with a schema-violation error.
//
// Build tag `helm` is set so go test ./... in workspaces without the
// helm binary remains green; CI runs:
//
//   go test -tags=helm ./helm/...
//
// from inside the chetana-defense or services workspace, with helm
// installed on the runner image.
package helm_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// chartRoot is the absolute path to the _chetana-service library chart
// resolved from this test file. The walk-up is deterministic — the
// test is always invoked from services/packages/helm/.
func chartRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// services/packages/helm  ->  ../../../infra/helm/charts/_chetana-service
	return filepath.Clean(filepath.Join(wd, "..", "..", "..", "infra", "helm", "charts", "_chetana-service"))
}

// requireHelm skips the test when the helm binary is not on PATH so
// developer workstations without helm installed remain unbroken.
func requireHelm(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("helm")
	if err != nil {
		t.Skipf("helm not on PATH; skipping (CI provides helm): %v", err)
	}
	return bin
}

// helmTemplate runs `helm template` against the example consumer
// chart, optionally overriding individual values. The returned
// (stdout, err) tuple is the raw helm output.
func helmTemplate(t *testing.T, helmBin, chartDir, valuesPath string, sets ...string) (string, error) {
	t.Helper()
	args := []string{"template", "release", chartDir}
	if valuesPath != "" {
		args = append(args, "-f", valuesPath)
	}
	for _, s := range sets {
		args = append(args, "--set", s)
	}
	cmd := exec.Command(helmBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

// dependencyUpdate must run once before the example consumer can
// resolve its chetana-service dependency. We invoke it lazily from
// each test so a clean checkout works without manual setup.
func ensureDeps(t *testing.T, helmBin, exampleDir string) {
	t.Helper()
	cmd := exec.Command(helmBin, "dependency", "update", exampleDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("helm dependency update: %v\nstderr: %s", err, stderr.String())
	}
}

// TestHelmTemplate_HappyPath verifies the consumer chart renders all
// six resource kinds when given a complete values file. Acceptance
// criterion #1.
func TestHelmTemplate_HappyPath(t *testing.T) {
	helmBin := requireHelm(t)
	chartDir := chartRoot(t)
	exampleDir := filepath.Join(chartDir, "test", "example-consumer")
	ensureDeps(t, helmBin, exampleDir)

	out, err := helmTemplate(t, helmBin, exampleDir, filepath.Join(exampleDir, "values.yaml"))
	if err != nil {
		t.Fatalf("helm template happy path failed: %v\noutput:\n%s", err, out)
	}

	wantKinds := []string{
		"kind: Deployment",
		"kind: Service",
		"kind: ServiceAccount",
		"kind: HorizontalPodAutoscaler",
		"kind: PodDisruptionBudget",
		"kind: NetworkPolicy",
	}
	for _, want := range wantKinds {
		if !strings.Contains(out, want) {
			t.Errorf("rendered chart missing %q", want)
		}
	}

	// Sanity: the chetana_region label MUST appear (REQ-NFR-SCALE-003).
	if !strings.Contains(out, `chetana.p9e.in/region: "us-gov-east-1"`) {
		t.Error("rendered chart missing chetana.p9e.in/region label")
	}
}

// TestHelmTemplate_NetworkPolicy_DefaultsToDeny verifies acceptance
// criterion #3: when networkPolicy.ingress is empty the rendered
// NetworkPolicy is `ingress: []` (Kubernetes default-deny semantics).
func TestHelmTemplate_NetworkPolicy_DefaultsToDeny(t *testing.T) {
	helmBin := requireHelm(t)
	chartDir := chartRoot(t)
	exampleDir := filepath.Join(chartDir, "test", "example-consumer")
	ensureDeps(t, helmBin, exampleDir)

	// Override networkPolicy.ingress=[] via --set-json; this collapses
	// to the default-deny path in the template.
	out, err := helmTemplate(t, helmBin, exampleDir,
		filepath.Join(exampleDir, "values.yaml"),
		"networkPolicy.ingress=null", // helm interprets `null` as nil
	)
	if err != nil {
		t.Fatalf("helm template default-deny: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "ingress: []") {
		t.Error("expected `ingress: []` (default-deny) in NetworkPolicy when ingress is empty")
	}
}

// TestHelmTemplate_RejectsMissingHPA verifies the schema fails
// rendering when hpa is absent. Acceptance criterion #2.
func TestHelmTemplate_RejectsMissingHPA(t *testing.T) {
	helmBin := requireHelm(t)
	chartDir := chartRoot(t)
	exampleDir := filepath.Join(chartDir, "test", "example-consumer")
	ensureDeps(t, helmBin, exampleDir)

	// Pass an alternate values file with no hpa block.
	tmp, err := os.CreateTemp(t.TempDir(), "no-hpa-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer tmp.Close()
	const noHPA = `
service:
  name: example
  port: 8080
  metricsPort: 9090
image:
  repository: ghcr.io/p9e/example
  tag: v0
region: us-gov-east-1
pdb:
  minAvailable: 1
networkPolicy:
  ingress: []
`
	if _, err := tmp.WriteString(noHPA); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()

	out, err := helmTemplate(t, helmBin, exampleDir, tmp.Name())
	if err == nil {
		t.Fatalf("expected schema rejection; got success:\n%s", out)
	}
	// Helm wraps schema errors in stderr; just look for the field name.
	if !strings.Contains(out, "hpa") {
		t.Errorf("expected error to mention 'hpa'; got: %s", out)
	}
}

// TestHelmTemplate_RejectsMissingPDB mirrors the previous test for the
// pdb block. Same mechanism, different missing field.
func TestHelmTemplate_RejectsMissingPDB(t *testing.T) {
	helmBin := requireHelm(t)
	chartDir := chartRoot(t)
	exampleDir := filepath.Join(chartDir, "test", "example-consumer")
	ensureDeps(t, helmBin, exampleDir)

	tmp, err := os.CreateTemp(t.TempDir(), "no-pdb-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer tmp.Close()
	const noPDB = `
service:
  name: example
  port: 8080
  metricsPort: 9090
image:
  repository: ghcr.io/p9e/example
  tag: v0
region: us-gov-east-1
hpa:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
networkPolicy:
  ingress: []
`
	if _, err := tmp.WriteString(noPDB); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()

	out, err := helmTemplate(t, helmBin, exampleDir, tmp.Name())
	if err == nil {
		t.Fatalf("expected schema rejection; got success:\n%s", out)
	}
	if !strings.Contains(out, "pdb") {
		t.Errorf("expected error to mention 'pdb'; got: %s", out)
	}
}

// TestHelmLint_LibraryChart runs `helm lint` against the library chart
// directly. The library is type=library so `helm lint` skips most
// install-time checks but still validates Chart.yaml and template
// syntax.
func TestHelmLint_LibraryChart(t *testing.T) {
	helmBin := requireHelm(t)
	chartDir := chartRoot(t)
	cmd := exec.Command(helmBin, "lint", chartDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// helm lint exits non-zero on any ERROR; warnings are tolerated.
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			t.Fatalf("helm lint failed:\nstdout: %s\nstderr: %s", stdout.String(), stderr.String())
		}
		t.Fatalf("helm lint exec: %v", err)
	}
}
