/**
 * Cesium chunk-split e2e — TASK-P1-WEB-002 acceptance #1 + #2.
 *
 * Two assertions:
 *
 *   1. The initial JS bundle (the chunks loaded BEFORE the user
 *      navigates to /dev/cesium) does NOT contain `@cesium/engine`
 *      or any of its identifying strings. We check this by
 *      collecting every script URL the browser fetches between
 *      the initial page load and the moment we navigate to the
 *      Cesium route — none of those URLs may match the
 *      cesium-engine chunk pattern.
 *
 *   2. Navigating to a Cesium-hosting route loads the chunk on
 *      demand. We assert at least one network request for a
 *      cesium-engine* chunk fires AFTER navigation begins.
 *
 * The chetana shell dynamic-imports Cesium via src/lib/cesium/
 * loader.ts; the manualChunks function in vite.config.ts splits
 * the @cesium/engine module into a stably-named chunk so the
 * regex below matches reliably across builds.
 */

import { test, expect } from "@playwright/test";
import {
  DEFAULT_USER,
  mockIamLogin,
  mockAuditSearch,
  mockExportsList,
} from "./_helpers";

test.describe("cesium chunk-split", () => {
  test.beforeEach(async ({ page }) => {
    await mockIamLogin(page);
    await mockAuditSearch(page);
    await mockExportsList(page);
  });

  test("initial bundle has no @cesium/engine; navigating loads it on demand", async ({ page }) => {
    const initialURLs: string[] = [];
    const allURLs: string[] = [];

    // Capture every JS network request from the moment the
    // browser starts loading the page.
    page.on("request", (req) => {
      const url = req.url();
      const isJS = url.endsWith(".js") || url.includes(".js?") || req.resourceType() === "script";
      if (isJS) {
        allURLs.push(url);
      }
    });

    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();
    await page.waitForURL("**/dashboard");

    // Snapshot the URLs fetched up to + including /dashboard. Any
    // JS chunk requested up to this point is part of the initial
    // bundle path the user hits before opting into Cesium.
    initialURLs.push(...allURLs);

    // Acceptance #1: none of the initial URLs reference cesium.
    const initialCesium = initialURLs.filter((u) => /cesium/i.test(u));
    expect(initialCesium, `unexpected cesium chunks in initial bundle: ${initialCesium.join(", ")}`).toEqual([]);

    // Now navigate to the Cesium-hosting route.
    const beforeNav = allURLs.length;
    await page.goto("/dev/cesium");

    // Wait for the cesium chunk to actually fetch.
    await expect.poll(
      () => allURLs.slice(beforeNav).filter((u) => /cesium-engine/i.test(u)).length,
      { timeout: 10_000, message: "expected cesium-engine* chunk to be fetched on demand" },
    ).toBeGreaterThan(0);
  });

  test("globe container renders without throwing", async ({ page }) => {
    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();
    await page.waitForURL("**/dashboard");

    await page.goto("/dev/cesium");
    await expect(page.getByTestId("cesium-host")).toBeVisible();
    // The Cesium container element should be present even before
    // the viewer is fully ready (Cesium injects a canvas + DOM
    // chrome into the host div on first frame).
    await expect(page.getByTestId("cesium-container")).toBeVisible();
  });
});
