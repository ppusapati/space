/**
 * Exports e2e — REQ-FUNC-CMN-005.
 * Acceptance #3 of TASK-P1-WEB-001: Export UI surfaces job
 * progress via WS push (no polling).
 *
 * The chetana cmd layer is not running, so this spec covers the
 * client-side guarantees only:
 *   • the page lists jobs from /v1/export/jobs (mocked).
 *   • the realtime client opens (or attempts to) without
 *     polling /v1/export/jobs after the initial fetch.
 *   • a "Download" link is rendered for succeeded jobs.
 *
 * The server-side WS-push end-to-end test belongs in the
 * services/realtime-gw integration suite, not here.
 */

import { test, expect } from "@playwright/test";
import {
  DEFAULT_USER,
  mockIamLogin,
  mockAuditSearch,
  mockExportsList,
} from "./_helpers";

test.describe("exports", () => {
  test.beforeEach(async ({ page }) => {
    await mockIamLogin(page);
    await mockAuditSearch(page);
    await mockExportsList(page);
    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();
    await page.waitForURL("**/dashboard");
  });

  test("lists existing jobs + renders Download for succeeded ones", async ({ page }) => {
    let jobsListCalls = 0;
    await page.route("**/v1/export/jobs", async (route) => {
      jobsListCalls++;
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          jobs: [
            {
              id: "job-1",
              kind: "audit_csv",
              status: "succeeded",
              enqueued_at: new Date(Date.now() - 600_000).toISOString(),
              completed_at: new Date(Date.now() - 580_000).toISOString(),
              presigned_url: "https://example.test/download/job-1",
              bytes_total: 102_400,
            },
          ],
        }),
      });
    });

    await page.goto("/exports");
    await expect(page.getByTestId("exports-list")).toBeVisible();
    await expect(page.getByTestId("export-status-job-1")).toContainText("succeeded");
    await expect(page.getByTestId("export-download-job-1")).toHaveAttribute(
      "href",
      "https://example.test/download/job-1",
    );

    // Wait long enough that any naive polling implementation
    // would have fired a second request. The chetana page uses
    // WS push only, so the call count must remain 1.
    await page.waitForTimeout(2_000);
    expect(jobsListCalls).toBe(1);
  });

  test("realtime state badge is present", async ({ page }) => {
    await page.goto("/exports");
    await expect(page.getByTestId("rt-state-badge")).toBeVisible();
  });
});
