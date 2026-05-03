// audit_csv.go — Category C3 processor.

package processors

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/worker"
)

// AuditCSVPayload is the per-job payload shape — the chetana
// audit.SearchQuery serialised into a flat map. The chetana audit
// service's `/v1/audit/export.csv?<query>` route accepts the
// same querystring shape as `/v1/audit/search`, but streams the
// full set as a CSV with the chain-attestation envelope as the
// leading `# envelope: ...` comment line.
type AuditCSVPayload struct {
	TenantID    string `json:"tenant_id"`
	Start       string `json:"start"`        // RFC 3339
	End         string `json:"end"`
	ActorUserID string `json:"actor_user_id"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	Decision    string `json:"decision"`
	Procedure   string `json:"procedure"`
	FreeText    string `json:"free_text"`
}

// NewAuditCSVProcessor renders the audit-svc's CSV export
// stream into the export-svc body.
func NewAuditCSVProcessor(src *SourceClient) worker.ProcessFunc {
	return func(ctx context.Context, job *queue.Job) (worker.ProcessOutput, error) {
		payload, err := payloadFromJob[AuditCSVPayload](job)
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		q := buildAuditQuery(payload)
		fullURL := fmt.Sprintf("%s/v1/audit/export.csv?%s", src.AuditBaseURL, q.Encode())
		body, err := src.streamBody(ctx, fullURL, "text/csv")
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		return worker.ProcessOutput{
			Body:        body,
			ContentType: "text/csv",
			Filename:    "audit.csv",
		}, nil
	}
}

// buildAuditQuery converts the typed payload into the querystring
// the chetana audit-svc accepts. Empty fields are skipped.
func buildAuditQuery(p *AuditCSVPayload) url.Values {
	v := url.Values{}
	if p.TenantID != "" {
		v.Set("tenant_id", p.TenantID)
	}
	if p.Start != "" {
		v.Set("start", p.Start)
	}
	if p.End != "" {
		v.Set("end", p.End)
	}
	if p.ActorUserID != "" {
		v.Set("actor_user_id", p.ActorUserID)
	}
	if p.Action != "" {
		v.Set("action", p.Action)
	}
	if p.Resource != "" {
		v.Set("resource", p.Resource)
	}
	if p.Decision != "" {
		v.Set("decision", p.Decision)
	}
	if p.Procedure != "" {
		v.Set("procedure", p.Procedure)
	}
	if p.FreeText != "" {
		v.Set("free_text", p.FreeText)
	}
	return v
}

// _ keeps queue.Job referenced in case the audit-csv processor
// later needs job-specific metadata (e.g. requester id) on the
// CSV envelope.
var _ = (*queue.Job)(nil)
