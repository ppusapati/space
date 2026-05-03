package health

import (
	"context"
	"testing"
	"time"
)

func TestNewAlerter_Defaults(t *testing.T) {
	st := &Store{}
	a, err := NewAlerter(AlerterConfig{Store: st})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if a.cfg.FlapThreshold != 3 {
		t.Errorf("flap threshold default: %d", a.cfg.FlapThreshold)
	}
	if a.cfg.FlapWindow != 10*time.Minute {
		t.Errorf("flap window default: %v", a.cfg.FlapWindow)
	}
	if a.cfg.SustainedThreshold != 5*time.Minute {
		t.Errorf("sustained default: %v", a.cfg.SustainedThreshold)
	}
}

func TestNewAlerter_NilStoreRejected(t *testing.T) {
	if _, err := NewAlerter(AlerterConfig{}); err == nil {
		t.Error("expected error for nil store")
	}
}

func TestNopNotifier_NeverErrors(t *testing.T) {
	if err := (NopNotifier{}).Notify(context.Background(), Alert{}); err != nil {
		t.Errorf("nop notifier should never err: %v", err)
	}
}

func TestCapturingNotifier_RecordsAlerts(t *testing.T) {
	c := &CapturingNotifier{}
	for i := 0; i < 3; i++ {
		_ = c.Notify(context.Background(), Alert{Service: "svc", Severity: SeverityWarn})
	}
	if len(c.Alerts) != 3 {
		t.Errorf("captured: %d", len(c.Alerts))
	}
}

func TestErrorRate(t *testing.T) {
	cases := []struct {
		errCount, succCount int64
		want                float64
	}{
		{0, 0, 0},
		{1, 1, 0.5},
		{1, 0, 1.0},
		{0, 100, 0.0},
		{25, 75, 0.25},
	}
	for _, c := range cases {
		if got := errorRate(c.errCount, c.succCount); got != c.want {
			t.Errorf("errorRate(%d,%d) = %v want %v", c.errCount, c.succCount, got, c.want)
		}
	}
}

func TestExcerpt(t *testing.T) {
	short := "short message"
	if got := excerpt(short); got != short {
		t.Errorf("short pass-through: %q", got)
	}
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'x'
	}
	got := excerpt(string(long))
	if len([]rune(got)) > 257 { // 256 + ellipsis
		t.Errorf("excerpt length: %d", len(got))
	}
}

func TestStatusConstants_Distinct(t *testing.T) {
	all := map[string]bool{
		StatusOK:       true,
		StatusDegraded: true,
		StatusDown:     true,
		StatusUnknown:  true,
	}
	if len(all) != 4 {
		t.Errorf("expected 4 distinct statuses, got %d", len(all))
	}
}

func TestSeverityConstants_Distinct(t *testing.T) {
	if SeverityWarn == SeverityPage {
		t.Error("severities must differ")
	}
}

func TestSnapshot_IsHealthy(t *testing.T) {
	if !(Snapshot{LastStatus: StatusOK}).IsHealthy() {
		t.Error("OK snapshot should be healthy")
	}
	for _, s := range []string{StatusDegraded, StatusDown, StatusUnknown} {
		if (Snapshot{LastStatus: s}).IsHealthy() {
			t.Errorf("%q should NOT be healthy", s)
		}
	}
}

func TestNewAggregator_Defaults(t *testing.T) {
	st := &Store{}
	a, err := NewAggregator(AggregatorConfig{Store: st})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if a.interval != 10*time.Second {
		t.Errorf("interval default: %v", a.interval)
	}
	if a.timeout != 5*time.Second {
		t.Errorf("timeout default: %v", a.timeout)
	}
}

func TestAggregator_RegisterAndTargetsSorted(t *testing.T) {
	a, _ := NewAggregator(AggregatorConfig{Store: &Store{}})
	a.Register("zeta", "https://z.example/ready")
	a.Register("alpha", "https://a.example/ready")
	a.Register("mu", "https://m.example/ready")
	got := a.Targets()
	if len(got) != 3 {
		t.Fatalf("targets: %d", len(got))
	}
	if got[0].Service != "alpha" || got[1].Service != "mu" || got[2].Service != "zeta" {
		t.Errorf("not sorted: %+v", got)
	}
}

func TestAggregator_RegisterReplaces(t *testing.T) {
	a, _ := NewAggregator(AggregatorConfig{Store: &Store{}})
	a.Register("svc", "https://old/ready")
	a.Register("svc", "https://new/ready")
	got := a.Targets()
	if len(got) != 1 || got[0].URL != "https://new/ready" {
		t.Errorf("re-register: %+v", got)
	}
}

func TestFallbackStatus(t *testing.T) {
	if got := fallbackStatus(""); got != StatusUnknown {
		t.Errorf("empty → %q want %q", got, StatusUnknown)
	}
	if got := fallbackStatus("ok"); got != "ok" {
		t.Errorf("non-empty pass-through: %q", got)
	}
}
