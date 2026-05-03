// audit_json.go — Category C4 processor.

package processors

import (
	"context"
	"fmt"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/worker"
)

// AuditJSONPayload mirrors AuditCSVPayload — the chetana audit-
// svc's `/v1/audit/export.json` route accepts the same query
// surface as `/v1/audit/search`.
type AuditJSONPayload = AuditCSVPayload

// NewAuditJSONProcessor renders the audit-svc's NDJSON export
// stream into the export-svc body.
func NewAuditJSONProcessor(src *SourceClient) worker.ProcessFunc {
	return func(ctx context.Context, job *queue.Job) (worker.ProcessOutput, error) {
		payload, err := payloadFromJob[AuditJSONPayload](job)
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		q := buildAuditQuery(payload)
		fullURL := fmt.Sprintf("%s/v1/audit/export.json?%s", src.AuditBaseURL, q.Encode())
		body, err := src.streamBody(ctx, fullURL, "application/x-ndjson")
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		return worker.ProcessOutput{
			Body:        body,
			ContentType: "application/x-ndjson",
			Filename:    "audit.ndjson",
		}, nil
	}
}

var _ = (*queue.Job)(nil)
