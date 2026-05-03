// gdpr_sar.go — Category C1 processor.
//
// REQ-FUNC-PLT-IAM-011 / Article 15 (right of access).
//
// Job payload:
//
//	{ "user_id": "<uuid>" }
//
// Body: NDJSON (one JSON object per line). Line 1 = user
// snapshot; subsequent lines = sub-rows (sessions, refresh
// tokens, mfa enrolment status, webauthn credentials, oauth auth
// codes, audit events keyed by actor_user_id). The chetana IAM
// service's `/v1/iam/gdpr/snapshot/{user_id}` route returns the
// already-built `gdpr.Snapshot` JSON; we project that into NDJSON.

package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/worker"
)

// GDPRSARPayload is the per-job payload shape.
type GDPRSARPayload struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// NewGDPRSARProcessor returns a ProcessorFunc that fetches the
// IAM-side gdpr.Snapshot via HTTP and renders it as NDJSON.
func NewGDPRSARProcessor(src *SourceClient) worker.ProcessFunc {
	return func(ctx context.Context, job *queue.Job) (worker.ProcessOutput, error) {
		payload, err := payloadFromJob[GDPRSARPayload](job)
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		if payload.UserID == "" {
			return worker.ProcessOutput{}, fmt.Errorf("%w: empty user_id", ErrPayloadInvalid)
		}

		// IAM route: /v1/iam/gdpr/snapshot/{user_id} → JSON of
		// the gdpr.Snapshot struct.
		url := fmt.Sprintf("%s/v1/iam/gdpr/snapshot/%s", src.IAMBaseURL, payload.UserID)
		var snap map[string]any
		if err := src.getJSON(ctx, url, &snap); err != nil {
			return worker.ProcessOutput{}, err
		}

		// Project the snapshot into NDJSON. The first line is the
		// envelope itself; subsequent lines spread the array
		// fields so a streaming consumer can pull them one at a
		// time without holding the whole document in memory.
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)

		// Top-level envelope (everything except the array fields).
		envelope := map[string]any{
			"kind":         "gdpr_sar",
			"job_id":       job.ID,
			"user_id":      payload.UserID,
			"tenant_id":    payload.TenantID,
			"generated_at": snap["generated_at"],
			"user":         snap["user"],
			"mfa":          snap["mfa"],
		}
		if err := enc.Encode(envelope); err != nil {
			return worker.ProcessOutput{}, fmt.Errorf("processors: encode envelope: %w", err)
		}

		// Spread the array sub-rows.
		for _, k := range []string{"sessions", "oauth_auth_codes", "webauthn_credentials"} {
			rows, _ := snap[k].([]any)
			for _, row := range rows {
				if err := enc.Encode(map[string]any{"section": k, "row": row}); err != nil {
					return worker.ProcessOutput{}, fmt.Errorf("processors: encode row: %w", err)
				}
			}
		}

		return worker.ProcessOutput{
			Body:        buf.Bytes(),
			ContentType: "application/x-ndjson",
			Filename:    "sar.ndjson",
		}, nil
	}
}
