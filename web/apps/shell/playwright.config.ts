/**
 * Playwright configuration for the chetana shell.
 *
 * Two operating modes:
 *
 *   • CI / pre-commit: webServer launches `vite preview` against
 *     the built bundle so the e2e suite asserts the production
 *     behaviour, not the dev-server's HMR-instrumented variant.
 *
 *   • local dev: developers can target a running `pnpm dev`
 *     server by setting CHETANA_E2E_BASE_URL — the webServer
 *     block is skipped when that var is set.
 *
 * The chetana cmd-layer routes the e2e specs hit are mocked at
 * the network layer via `page.route(...)` (see specs/_helpers.ts).
 * That keeps the suite hermetic — the chetana platform services
 * don't have to be running locally.
 */

import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.CHETANA_E2E_BASE_URL ?? "http://127.0.0.1:4173";

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : [["list"]],
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
  ],
  webServer: process.env.CHETANA_E2E_BASE_URL
    ? undefined
    : {
        command: "pnpm run build && pnpm run preview --port 4173 --strictPort",
        url: baseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      },
});
