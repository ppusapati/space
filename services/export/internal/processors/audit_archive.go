// audit_archive.go — Category C2 processor.
//
// Archive a sealed range of audit events into Glacier-class S3
// storage. The chetana scheduler periodically enqueues these jobs
// (per the retention policy in TASK-P1-AUDIT-002).
//
// Job payload:
//
//	{
//	  "tenant_id": "<uuid>",
//	  "range_start": "RFC3339",
//	  "range_end":   "RFC3339",
//	  "first_chain_seq": 1,
//	  "last_chain_seq":  1000
//	}
//
// Body: gzip-compressed NDJSON of the audit events in the range,
// with the chain-attestation envelope as the first line.

package processors

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/worker"
)

// AuditArchivePayload is the per-job payload shape.
type AuditArchivePayload struct {
	TenantID      string `json:"tenant_id"`
	RangeStart    string `json:"range_start"` // RFC 3339
	RangeEnd      string `json:"range_end"`
	FirstChainSeq int64  `json:"first_chain_seq"`
	LastChainSeq  int64  `json:"last_chain_seq"`
}

// NewAuditArchiveProcessor fetches the audit-svc's NDJSON export
// for the supplied range, gzip-compresses it, and returns the
// archive body. The export-svc worker then uploads to S3 with
// `Content-Encoding: gzip` + `StorageClass: GLACIER` (configured
// via the chetana cmd-layer's S3 client).
func NewAuditArchiveProcessor(src *SourceClient) worker.ProcessFunc {
	return func(ctx context.Context, job *queue.Job) (worker.ProcessOutput, error) {
		payload, err := payloadFromJob[AuditArchivePayload](job)
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		if payload.TenantID == "" || payload.RangeStart == "" || payload.RangeEnd == "" {
			return worker.ProcessOutput{}, fmt.Errorf("%w: tenant_id + range_start + range_end required", ErrPayloadInvalid)
		}
		q := url.Values{}
		q.Set("tenant_id", payload.TenantID)
		q.Set("start", payload.RangeStart)
		q.Set("end", payload.RangeEnd)
		// `attest=true` tells the audit-svc to emit the chain-
		// attestation envelope as the first line of the NDJSON
		// stream. The recipient can independently re-verify the
		// chain by fetching the same range from a live audit-svc.
		q.Set("attest", "true")
		fullURL := fmt.Sprintf("%s/v1/audit/export.json?%s", src.AuditBaseURL, q.Encode())
		body, err := src.streamBody(ctx, fullURL, "application/x-ndjson")
		if err != nil {
			return worker.ProcessOutput{}, err
		}

		gz, err := gzipBytes(body)
		if err != nil {
			return worker.ProcessOutput{}, err
		}
		filename := fmt.Sprintf("audit-%d-%d.ndjson.gz", payload.FirstChainSeq, payload.LastChainSeq)
		return worker.ProcessOutput{
			Body:        gz,
			ContentType: "application/gzip",
			Filename:    filename,
		}, nil
	}
}

var _ = (*queue.Job)(nil)
