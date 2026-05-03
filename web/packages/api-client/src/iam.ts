/**
 * iam.ts — chetana IAM client.
 *
 * Surface mirrors the cmd-layer JSON handlers RETROFIT-001 mounts:
 *
 *   POST /v1/iam/login
 *   POST /v1/iam/logout
 *   POST /v1/iam/refresh
 *   POST /v1/iam/mfa/verify
 *   POST /v1/iam/mfa/enroll
 *   POST /v1/iam/webauthn/register/begin
 *   POST /v1/iam/webauthn/register/finish
 *   POST /v1/iam/webauthn/assert/begin
 *   POST /v1/iam/webauthn/assert/finish
 *   POST /v1/iam/reset/request
 *   POST /v1/iam/reset/confirm
 *   GET  /v1/iam/sessions
 *   POST /v1/iam/sessions/{id}/revoke
 *   GET  /v1/iam/api-keys
 *   POST /v1/iam/api-keys
 *   POST /v1/iam/api-keys/{id}/revoke
 *
 * The chetana access-token shape mirrors the IAM JWT claims
 * (services/iam/internal/token/jwt.go::Claims) — stable across the
 * eventual Connect regen.
 */

import { baseURL, joinURL, request } from "./common.js";

export interface Principal {
  user_id: string;
  tenant_id: string;
  session_id: string;
  is_us_person: boolean;
  clearance_level: "public" | "internal" | "restricted" | "cui" | "itar";
  nationality: string;
  roles: string[];
  scopes: string[];
  amr: string[];
  issued_at: string;     // RFC 3339
  expires_at: string;    // RFC 3339
  jti: string;
}

export interface LoginRequest {
  email: string;
  password: string;
  /**
   * When the user has MFA enrolled, the first /login returns
   * MFA_REQUIRED. Resubmit with mfa_code populated AND the
   * mfa_session_token from the first response.
   */
  mfa_code?: string;
  mfa_session_token?: string;
}

export interface LoginResponse {
  status:
    | "ok"
    | "mfa_required"
    | "bad_credentials"
    | "rate_limited"
    | "locked"
    | "internal_error";
  access_token?: string;
  access_token_expires_at?: string;
  refresh_token?: string;
  refresh_token_expires_at?: string;
  session_id?: string;
  /**
   * Set when status === 'mfa_required'. Echo back into the next
   * /login call to bind the MFA challenge to the original
   * email+password verification.
   */
  mfa_session_token?: string;
  /**
   * Set when status === 'rate_limited' or 'locked'. Seconds.
   */
  retry_after_seconds?: number;
  reason?: string;
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  return request<LoginResponse>(joinURL(baseURL(), "v1/iam/login"), {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function logout(bearer: string): Promise<void> {
  await request<void>(joinURL(baseURL(), "v1/iam/logout"), {
    method: "POST",
    bearer,
  });
}

export interface RefreshResponse {
  access_token: string;
  access_token_expires_at: string;
  refresh_token: string;
  refresh_token_expires_at: string;
}

export async function refresh(refreshToken: string): Promise<RefreshResponse> {
  return request<RefreshResponse>(joinURL(baseURL(), "v1/iam/refresh"), {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

// ----------------------------------------------------------------------
// MFA TOTP
// ----------------------------------------------------------------------

export interface MfaEnrollResponse {
  secret_base32: string;
  qr_uri: string;     // otpauth://totp/...
  backup_codes: string[];
}

export async function enrollMfa(bearer: string): Promise<MfaEnrollResponse> {
  return request<MfaEnrollResponse>(joinURL(baseURL(), "v1/iam/mfa/enroll"), {
    method: "POST",
    bearer,
  });
}

export async function verifyMfa(bearer: string, code: string): Promise<void> {
  await request<void>(joinURL(baseURL(), "v1/iam/mfa/verify"), {
    method: "POST",
    bearer,
    body: JSON.stringify({ code }),
  });
}

// ----------------------------------------------------------------------
// WebAuthn
// ----------------------------------------------------------------------

/**
 * The chetana cmd-layer wraps the underlying go-webauthn protocol
 * library; the JSON shapes returned here are the same
 * `PublicKeyCredentialCreationOptions` / `PublicKeyCredentialRequestOptions`
 * the W3C WebAuthn API expects, with the binary fields
 * base64url-encoded so the JSON envelope round-trips cleanly.
 */
export interface WebAuthnBeginResponse {
  publicKey: PublicKeyCredentialCreationOptionsJSON;
  session_token: string;
}

export interface WebAuthnAssertBeginResponse {
  publicKey: PublicKeyCredentialRequestOptionsJSON;
  session_token: string;
}

// W3C-ish JSON shapes — fields are stringified base64url.
export interface PublicKeyCredentialCreationOptionsJSON {
  challenge: string;
  rp: { id: string; name: string };
  user: { id: string; name: string; displayName: string };
  pubKeyCredParams: { type: "public-key"; alg: number }[];
  timeout?: number;
  excludeCredentials?: { id: string; type: "public-key" }[];
  authenticatorSelection?: {
    authenticatorAttachment?: "platform" | "cross-platform";
    userVerification?: "required" | "preferred" | "discouraged";
  };
  attestation?: "none" | "indirect" | "direct" | "enterprise";
}

export interface PublicKeyCredentialRequestOptionsJSON {
  challenge: string;
  timeout?: number;
  rpId: string;
  allowCredentials?: { id: string; type: "public-key" }[];
  userVerification?: "required" | "preferred" | "discouraged";
}

export async function webauthnRegisterBegin(bearer: string): Promise<WebAuthnBeginResponse> {
  return request<WebAuthnBeginResponse>(
    joinURL(baseURL(), "v1/iam/webauthn/register/begin"),
    { method: "POST", bearer },
  );
}

export async function webauthnRegisterFinish(
  bearer: string,
  sessionToken: string,
  // The browser's PublicKeyCredential serialised — caller
  // converts ArrayBuffers to base64url before posting.
  credential: unknown,
): Promise<void> {
  await request<void>(joinURL(baseURL(), "v1/iam/webauthn/register/finish"), {
    method: "POST",
    bearer,
    body: JSON.stringify({ session_token: sessionToken, credential }),
  });
}

export async function webauthnAssertBegin(email: string): Promise<WebAuthnAssertBeginResponse> {
  return request<WebAuthnAssertBeginResponse>(
    joinURL(baseURL(), "v1/iam/webauthn/assert/begin"),
    { method: "POST", body: JSON.stringify({ email }) },
  );
}

export async function webauthnAssertFinish(
  sessionToken: string,
  credential: unknown,
): Promise<LoginResponse> {
  return request<LoginResponse>(joinURL(baseURL(), "v1/iam/webauthn/assert/finish"), {
    method: "POST",
    body: JSON.stringify({ session_token: sessionToken, credential }),
  });
}

// ----------------------------------------------------------------------
// Password reset
// ----------------------------------------------------------------------

/**
 * Constant-time response per REQ-FUNC-PLT-IAM-010. The client always
 * sees `accepted: true` regardless of whether the email maps to a
 * known user — non-disclosure is enforced server-side.
 */
export async function requestPasswordReset(email: string): Promise<void> {
  await request<void>(joinURL(baseURL(), "v1/iam/reset/request"), {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export async function confirmPasswordReset(
  token: string,
  newPassword: string,
): Promise<void> {
  await request<void>(joinURL(baseURL(), "v1/iam/reset/confirm"), {
    method: "POST",
    body: JSON.stringify({ token, new_password: newPassword }),
  });
}

// ----------------------------------------------------------------------
// Sessions (REQ-FUNC-PLT-IAM-009)
// ----------------------------------------------------------------------

export interface Session {
  session_id: string;
  issued_at: string;
  last_seen_at: string;
  idle_expires_at: string;
  absolute_expires_at: string;
  client_ip: string;
  user_agent: string;
  amr: string[];
  current: boolean; // true for the session of the calling token
}

export async function listSessions(bearer: string): Promise<Session[]> {
  const r = await request<{ sessions: Session[] }>(
    joinURL(baseURL(), "v1/iam/sessions"),
    { bearer },
  );
  return r.sessions;
}

export async function revokeSession(bearer: string, sessionID: string): Promise<void> {
  await request<void>(
    joinURL(baseURL(), `v1/iam/sessions/${encodeURIComponent(sessionID)}/revoke`),
    { method: "POST", bearer },
  );
}

// ----------------------------------------------------------------------
// API keys (long-lived bearer tokens for service-to-service)
// ----------------------------------------------------------------------

export interface ApiKey {
  id: string;
  label: string;
  scopes: string[];
  created_at: string;
  expires_at: string | null;
  last_used_at: string | null;
}

export interface ApiKeyCreated extends ApiKey {
  /** The bearer value — shown to the user EXACTLY ONCE. */
  bearer: string;
}

export async function listApiKeys(bearer: string): Promise<ApiKey[]> {
  const r = await request<{ api_keys: ApiKey[] }>(
    joinURL(baseURL(), "v1/iam/api-keys"),
    { bearer },
  );
  return r.api_keys;
}

export async function createApiKey(
  bearer: string,
  label: string,
  scopes: string[],
  ttlDays?: number,
): Promise<ApiKeyCreated> {
  return request<ApiKeyCreated>(joinURL(baseURL(), "v1/iam/api-keys"), {
    method: "POST",
    bearer,
    body: JSON.stringify({ label, scopes, ttl_days: ttlDays }),
  });
}

export async function revokeApiKey(bearer: string, id: string): Promise<void> {
  await request<void>(
    joinURL(baseURL(), `v1/iam/api-keys/${encodeURIComponent(id)}/revoke`),
    { method: "POST", bearer },
  );
}
