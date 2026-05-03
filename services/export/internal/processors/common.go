// Package processors implements the per-job-kind body renderers
// the export worker dispatches to.
//
// → REQ-FUNC-CMN-005, design.md §5.2.
//
// Cross-service dependency arrow
// ------------------------------
// The chetana service modules are intentionally one-way: every
// service depends on `services/packages/*` but no service
// depends on a sibling service module. The export processors
// would naturally want to import (e.g.) the gdpr SnapshotBuilder
// from services/iam/internal/gdpr, but that would create a cycle
// the moment iam grows an internal use of an export-svc helper.
//
// Resolution: the processors are PURE per-kind body renderers
// that fetch their inputs from the source service over its HTTP
// surface. The chetana cmd-layer wires a `SourceClient` that
// knows the in-cluster URL of the source service; the processor
// receives the client and invokes the right route.
//
// In single-binary dev posture (everything on localhost) the
// SourceClient hits 127.0.0.1; in production it hits the
// service-mesh DNS. Identical code path either way — tests pass
// a `httptest.Server` URL.

package processors

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/worker"
)

// SourceClient holds the in-cluster URLs the export worker uses
// to fetch source data from chetana services. cmd/export
// constructs one and passes it to every processor via the
// closures registered against the Registry.
type SourceClient struct {
	HTTP        *http.Client
	IAMBaseURL  string // e.g. http://iam.chetana.svc.cluster.local:8080
	AuditBaseURL string // e.g. http://audit.chetana.svc.cluster.local:8082
	// Bearer is the export-svc's own service-account token (long-
	// lived, scoped to the export-internal permissions). Stamped
	// onto every cross-service call so the source service's authz
	// interceptor sees a real principal.
	Bearer string
}

// NewSourceClient returns a SourceClient with sensible defaults.
func NewSourceClient(iamBase, auditBase, bearer string) *SourceClient {
	return &SourceClient{
		HTTP:         &http.Client{Timeout: 60 * time.Second},
		IAMBaseURL:   strings.TrimRight(iamBase, "/"),
		AuditBaseURL: strings.TrimRight(auditBase, "/"),
		Bearer:       bearer,
	}
}

// getJSON does a GET that decodes the response body into `out`.
// Used by the processors that consume chetana JSON read endpoints.
func (s *SourceClient) getJSON(ctx context.Context, fullURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("processors: build req: %w", err)
	}
	if s.Bearer != "" {
		req.Header.Set("Authorization", "Bearer "+s.Bearer)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("processors: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("processors: %s → %d: %s", fullURL, resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// streamBody does a GET and returns the raw response body as a
// byte slice. Used by the audit-export processors that just
// proxy the source service's already-formatted output.
func (s *SourceClient) streamBody(ctx context.Context, fullURL, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("processors: build req: %w", err)
	}
	if s.Bearer != "" {
		req.Header.Set("Authorization", "Bearer "+s.Bearer)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("processors: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("processors: %s → %d: %s", fullURL, resp.StatusCode, preview)
	}
	return io.ReadAll(resp.Body)
}

// gzipBytes returns a gzip-compressed copy of body. Used by the
// audit-archive processor for cold storage size.
func gzipBytes(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(body); err != nil {
		return nil, fmt.Errorf("processors: gzip: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("processors: gzip close: %w", err)
	}
	return buf.Bytes(), nil
}

// payloadFromJob unmarshals the per-job payload column. Returns
// ErrPayloadInvalid when the bytes don't decode.
func payloadFromJob[T any](job *queue.Job) (*T, error) {
	if len(job.Payload) == 0 {
		return nil, ErrPayloadInvalid
	}
	var out T
	if err := json.Unmarshal(job.Payload, &out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPayloadInvalid, err)
	}
	return &out, nil
}

// ProcessorFunc is the closure shape every processor exposes. The
// cmd-layer constructs each ProcessorFunc with its closed-over
// `SourceClient` and registers it against `worker.Registry`.
type ProcessorFunc = worker.ProcessFunc

// ErrPayloadInvalid is returned when a job's payload column does
// not decode into the kind-specific shape.
var ErrPayloadInvalid = errors.New("processors: invalid payload")
