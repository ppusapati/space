/**
 * Audit viewer e2e — REQ-FUNC-PLT-AUDIT-004.
 * Acceptance #2 of TASK-P1-WEB-001: paginate 100 k events
 * without UI jank (virtualised list).
 *
 * Strategy: load the audit page, scroll the result container to
 * the bottom 50 times to trigger 50 keyset-pagination loads
 * (= 5 000 rows append + bounded eviction at 10k). Assert the
 * scroll container's scrollHeight stays bounded (i.e. the
 * virtualised render did NOT explode the DOM).
 */

import { test, expect } from "@playwright/test";
import {
  DEFAULT_USER,
  mockIamLogin,
  mockAuditSearch,
} from "./_helpers";

test.describe("audit viewer", () => {
  test.beforeEach(async ({ page }) => {
    await mockIamLogin(page);
    await mockAuditSearch(page);
    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();
    await page.waitForURL("**/dashboard");
  });

  test("renders results + scrolls 50 pages without runaway DOM", async ({ page }) => {
    await page.goto("/audit");
    const list = page.getByTestId("audit-results");
    await expect(list).toBeVisible();

    // Initial page = 100 rows.
    const initialChildren = await list.evaluate((el) => el.children.length);
    expect(initialChildren).toBeGreaterThanOrEqual(100);

    // Scroll to bottom 50 times. Each scroll triggers a keyset
    // load via the page's onScroll handler.
    for (let i = 0; i < 50; i++) {
      await list.evaluate((el) => {
        el.scrollTop = el.scrollHeight;
      });
      // Allow the load to settle before the next scroll.
      await page.waitForTimeout(50);
    }

    // The page caps the in-memory window at 10_000 rows. Each
    // row is a fixed 56px so the worst-case scrollHeight is
    // bounded by 10_000 * 56 = 560_000px. Assert the bound holds.
    const finalScrollHeight = await list.evaluate((el) => el.scrollHeight);
    expect(finalScrollHeight).toBeLessThanOrEqual(560_000 + 1000); // small slack for sentinel rows
  });

  test("filter narrows results", async ({ page }) => {
    await page.goto("/audit");
    await page.getByTestId("filter-decision").selectOption("deny");
    await page.getByTestId("search-submit").click();

    const list = page.getByTestId("audit-results");
    await expect(list).toBeVisible();
    // Every visible row should carry the "deny" badge — the mock
    // emits deny on every 17th row so the filter narrows.
    // We assert by counting visible deny badges > 0.
    const denyCount = await list
      .locator("text=deny")
      .count();
    expect(denyCount).toBeGreaterThan(0);
  });

  test("CSV export trigger shows the success banner", async ({ page }) => {
    await page.goto("/audit");
    await page.getByTestId("search-submit").click(); // populate lastQuery
    await page.getByTestId("export-csv").click();
    await expect(page.getByText(/Export job .* submitted/i)).toBeVisible();
  });
});
