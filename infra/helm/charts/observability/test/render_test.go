// Package observability_test renders the observability subchart
// and asserts the produced ConfigMaps carry the chetana
// dashboards + scrape config — REQ-NFR-OBS-003 / TASK-P1-OBS-001
// acceptance #1 + #2.
//
// The test invokes `helm template` directly against the chart
// directory so the assertion exercises exactly what `helm
// upgrade` would render at deploy time.

package observability_test

import (
	"os/exec"
	"strings"
	"testing"
)

// renderChart shells out to `helm template`. Skips when helm is
// not on PATH so the test stays usable in dev environments
// without the helm binary installed.
func renderChart(t *testing.T, args ...string) string {
	t.Helper()
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not on PATH; skipping observability render test")
	}
	full := append([]string{"template", "test", "..", "--namespace", "monitoring"}, args...)
	out, err := exec.Command("helm", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, out)
	}
	return string(out)
}

// REQ-NFR-OBS-003 acceptance #1: Grafana dashboards are mounted
// via the provisioning ConfigMap.
func TestRender_GrafanaDashboardsConfigMap(t *testing.T) {
	out := renderChart(t)
	for _, dash := range []string{"iam.json", "audit.json", "realtime-gw.json", "notify.json", "export.json"} {
		if !strings.Contains(out, dash+":") {
			t.Errorf("dashboard %q not found in rendered ConfigMap", dash)
		}
	}
	// The dashboards.yaml provider file must point Grafana at the
	// mount path that the ConfigMap is projected onto.
	if !strings.Contains(out, "/etc/grafana/provisioning/dashboards") {
		t.Error("dashboards provider path missing from rendered template")
	}
}

func TestRender_GrafanaDatasourceConfigMap(t *testing.T) {
	out := renderChart(t)
	if !strings.Contains(out, "prometheus.yaml") {
		t.Error("prometheus datasource entry missing from rendered template")
	}
	// Datasource UID is what the dashboards reference.
	if !strings.Contains(out, "uid: prometheus") {
		t.Error("datasource UID 'prometheus' missing")
	}
}

// REQ-NFR-OBS-003 acceptance #2: Prometheus scrape config is
// rendered into a ConfigMap that targets the chetana :9090
// metrics ports.
func TestRender_PrometheusScrapeConfigMap(t *testing.T) {
	out := renderChart(t)
	if !strings.Contains(out, "prometheus.yml") {
		t.Error("prometheus.yml key missing from rendered ConfigMap")
	}
	for _, target := range []string{
		`iam:9090`, `platform:9091`, `audit:9092`,
		`notify:9093`, `export:9094`, `scheduler:9095`, `realtime-gw:9096`,
	} {
		if !strings.Contains(out, target) {
			t.Errorf("scrape target %q missing from rendered template", target)
		}
	}
	if !strings.Contains(out, "kubernetes_sd_configs") {
		t.Error("in-cluster service discovery (kubernetes_sd_configs) missing")
	}
}

func TestRender_FullnameTemplated(t *testing.T) {
	out := renderChart(t)
	// The `test` release name should show up in the rendered
	// metadata.name fields via the standard fullname helper.
	if !strings.Contains(out, "test-observability-grafana-dashboards") &&
		!strings.Contains(out, "test-grafana-dashboards") {
		t.Error("fullname helper did not produce the expected resource name")
	}
}
