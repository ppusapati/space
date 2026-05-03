/**
 * Shared Playwright helpers for the chetana shell e2e suite.
 *
 * The chetana cmd-layer is NOT running during e2e — every API +
 * WS route is mocked at the page level via Playwright's request
 * interception. This keeps the suite hermetic + fast.
 */

import type { Page, Route } from "@playwright/test";

export interface MockUser {
  email: string;
  password: string;
  display_name?: string;
  user_id?: string;
  tenant_id?: string;
  is_us_person?: boolean;
  clearance_level?: "public" | "internal" | "restricted" | "cui" | "itar";
  /** When true, the first /login response demands MFA. */
  mfaEnrolled?: boolean;
  /** Accepted TOTP code. Defaults to "123456". */
  mfaCode?: string;
}

export const DEFAULT_USER: MockUser = {
  email: "alice@example.com",
  password: "correct-horse-battery-staple",
  display_name: "Alice Example",
  user_id: "11111111-1111-1111-1111-111111111111",
  tenant_id: "00000000-0000-0000-0000-000000000001",
  is_us_person: true,
  clearance_level: "cui",
  mfaEnrolled: false,
  mfaCode: "123456",
};

const FAKE_TOKEN =
  "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ0ZXN0Iiwic2Vzc2lvbl9pZCI6InNlc3MtZTJlIn0.fake";

export async function mockIamLogin(page: Page, user: MockUser = DEFAULT_USER): Promise<void> {
  let mfaSessionToken: string | null = null;

  await page.route("**/v1/iam/login", async (route: Route) => {
    const body = JSON.parse(route.request().postData() ?? "{}") as {
      email?: string;
      password?: string;
      mfa_code?: string;
      mfa_session_token?: string;
    };

    if (body.email !== user.email || body.password !== user.password) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "bad_credentials" }),
      });
      return;
    }

    if (user.mfaEnrolled && !body.mfa_session_token) {
      mfaSessionToken = "mfa-token-e2e";
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          status: "mfa_required",
          mfa_session_token: mfaSessionToken,
        }),
      });
      return;
    }

    if (user.mfaEnrolled) {
      if (body.mfa_session_token !== mfaSessionToken || body.mfa_code !== user.mfaCode) {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ status: "bad_credentials" }),
        });
        return;
      }
    }

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        status: "ok",
        access_token: FAKE_TOKEN,
        access_token_expires_at: new Date(Date.now() + 15 * 60_000).toISOString(),
        refresh_token: "refresh-e2e",
        refresh_token_expires_at: new Date(Date.now() + 7 * 24 * 60 * 60_000).toISOString(),
        session_id: "sess-e2e",
      }),
    });
  });
}

export async function mockIamReset(page: Page): Promise<void> {
  await page.route("**/v1/iam/reset/request", async (route) => {
    await route.fulfill({ status: 204, body: "" });
  });
  await page.route("**/v1/iam/reset/confirm", async (route) => {
    await route.fulfill({ status: 204, body: "" });
  });
}

export async function mockAuditSearch(page: Page, totalRows = 100_000): Promise<void> {
  // Generate a stable, large fixture lazily — Playwright only
  // ever asks for one page's worth at a time.
  function pageFor(beforeID: number | null) {
    const limit = 100;
    const startID = beforeID ?? totalRows + 1;
    const hits: unknown[] = [];
    for (let i = 1; i <= limit; i++) {
      const id = startID - i;
      if (id <= 0) break;
      hits.push({
        id,
        tenant_id: DEFAULT_USER.tenant_id,
        event_time: new Date(Date.now() - id * 60_000).toISOString(),
        actor_user_id: DEFAULT_USER.user_id,
        actor_session_id: "sess-e2e",
        actor_client_ip: "10.0.0.1",
        actor_user_agent: "playwright",
        action: "iam.user.read",
        resource: `user-${id}`,
        decision: id % 17 === 0 ? "deny" : "ok",
        reason: id % 17 === 0 ? "explicit_deny" : "",
        matched_policy_id: id % 17 === 0 ? "deny-rule" : "",
        procedure: "/iam.v1.UserService/Read",
        classification: "cui",
        metadata: {},
      });
    }
    const last = hits[hits.length - 1] as { id: number; event_time: string } | undefined;
    return {
      hits,
      next_cursor:
        last && last.id > 1
          ? { before_time: last.event_time, before_id: last.id }
          : null,
    };
  }

  await page.route("**/v1/audit/search**", async (route) => {
    const url = new URL(route.request().url());
    const beforeID = url.searchParams.get("before_id");
    const body = pageFor(beforeID ? Number(beforeID) : null);
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });

  await page.route("**/v1/audit/export", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ job_id: "job-e2e-export", status: "queued" }),
    });
  });
}

export async function mockExportsList(page: Page): Promise<void> {
  let jobs = [
    {
      id: "job-1",
      kind: "audit_csv",
      status: "succeeded",
      enqueued_at: new Date(Date.now() - 600_000).toISOString(),
      completed_at: new Date(Date.now() - 580_000).toISOString(),
      presigned_url: "https://example.test/download/job-1?sig=abc",
      bytes_total: 1024 * 100,
    },
    {
      id: "job-2",
      kind: "gdpr_sar",
      status: "running",
      enqueued_at: new Date(Date.now() - 60_000).toISOString(),
    },
  ];
  await page.route("**/v1/export/jobs", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ jobs }),
    });
  });
  // Mock the WS upgrade as a no-op the browser can connect to so
  // the realtime client moves to "open" state without a real
  // gateway. We don't actually need to push events during the e2e
  // (the static fixture is enough) — the spec just asserts the
  // initial render + the rt-state-badge.
  void jobs; // referenced for potential future push fixture
}
