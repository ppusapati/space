/**
 * @chetana/api-client — chetana platform API surface (browser + node).
 *
 * Three sub-modules:
 *   • ./iam      — login / MFA / WebAuthn / sessions / API keys / reset
 *   • ./audit    — search / paginate / trigger CSV/JSON export
 *   • ./realtime — WebSocket client with auto-reconnect, backoff,
 *                   topic subscription manager, typed close-code
 *                   handlers
 *
 * Why hand-written rather than Connect-generated:
 *
 *   The chetana platform's `iam.proto` / `audit.proto` /
 *   `realtime.proto` are blocked on OQ-004 (BSR auth). Until that
 *   unblocks, the Go cmd-layer mounts plain JSON + WS handlers on
 *   srv.Mux (per TASK-P1-WIRING-RETROFIT-001), and this package
 *   talks to those routes via fetch + WebSocket. When the proto
 *   regen unblocks, generated Connect clients drop in alongside
 *   without changing the call sites — every export is named so the
 *   swap is mechanical.
 */

export * as iam from "./iam.js";
export * as audit from "./audit.js";
export * as realtime from "./realtime.js";

// Common error shape every client throws.
export {
  ApiError,
  isApiError,
  type ApiErrorJSON,
} from "./common.js";
