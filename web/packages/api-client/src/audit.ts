/**
 * audit.ts — chetana audit-svc client.
 *
 * Surface mirrors the cmd-layer JSON handlers:
 *
 *   GET  /v1/audit/search?...           → paginated search
 *   POST /v1/audit/export                → enqueue CSV/JSON export
 *
 * Search filters (REQ-FUNC-PLT-AUDIT-003) align with
 * services/audit/internal/search/query.go::Query.
 */

import { baseURL, joinURL, request } from "./common.js";

export interface AuditEvent {
  id: number;
  tenant_id: string;
  event_time: string;
  actor_user_id: string;
  actor_session_id: string;
  actor_client_ip: string;
  actor_user_agent: string;
  action: string;
  resource: string;
  decision: "allow" | "deny" | "ok" | "fail" | "info";
  reason: string;
  matched_policy_id: string;
  procedure: string;
  classification: "public" | "internal" | "restricted" | "cui" | "itar";
  metadata: Record<string, unknown>;
  // Chain attestation — present on every row but not used by the
  // UI directly; the export envelope is the consumer-facing
  // attestation.
  prev_hash?: string;
  row_hash?: string;
  chain_seq?: number;
}

export interface SearchQuery {
  start?: string;          // RFC 3339 inclusive lower bound
  end?: string;            // RFC 3339 inclusive upper bound
  actor_user_id?: string;
  action?: string;
  resource?: string;
  decision?: AuditEvent["decision"];
  procedure?: string;
  free_text?: string;      // "key=value" against metadata JSONB
  limit?: number;          // max 500
  before_time?: string;    // keyset pagination cursor
  before_id?: number;
}

export interface SearchPage {
  hits: AuditEvent[];
  next_cursor: { before_time: string; before_id: number } | null;
}

export async function search(bearer: string, q: SearchQuery): Promise<SearchPage> {
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(q)) {
    if (v === undefined || v === null || v === "") continue;
    params.set(k, String(v));
  }
  return request<SearchPage>(
    joinURL(baseURL(), `v1/audit/search?${params.toString()}`),
    { bearer },
  );
}

// ----------------------------------------------------------------------
// Export (CSV / JSON) — kicks off an export-svc job.
// ----------------------------------------------------------------------

export interface ExportRequest {
  format: "csv" | "json";
  query: SearchQuery;
}

export interface ExportSubmission {
  job_id: string;
  status: "queued" | "running" | "succeeded" | "failed" | "expired";
}

export async function submitExport(
  bearer: string,
  req: ExportRequest,
): Promise<ExportSubmission> {
  return request<ExportSubmission>(joinURL(baseURL(), "v1/audit/export"), {
    method: "POST",
    bearer,
    body: JSON.stringify(req),
  });
}
