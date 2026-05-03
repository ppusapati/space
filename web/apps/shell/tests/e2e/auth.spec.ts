/**
 * Auth e2e — REQ-FUNC-PLT-IAM-001 + REQ-FUNC-PLT-IAM-004.
 * Acceptance #1 of TASK-P1-WEB-001: Login → MFA → land on default
 * route works under Playwright.
 */

import { test, expect } from "@playwright/test";
import {
  DEFAULT_USER,
  mockIamLogin,
  mockIamReset,
  mockAuditSearch,
  mockExportsList,
} from "./_helpers";

test.describe("auth", () => {
  test("password-only login lands on /dashboard", async ({ page }) => {
    await mockIamLogin(page, { ...DEFAULT_USER, mfaEnrolled: false });
    await mockAuditSearch(page);
    await mockExportsList(page);

    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();

    await page.waitForURL("**/dashboard");
    expect(page.url()).toContain("/dashboard");
  });

  test("MFA-required login surfaces the TOTP entry then continues", async ({ page }) => {
    await mockIamLogin(page, { ...DEFAULT_USER, mfaEnrolled: true });
    await mockAuditSearch(page);
    await mockExportsList(page);

    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill(DEFAULT_USER.password);
    await page.getByTestId("login-submit").click();

    await expect(page.getByTestId("login-mfa-code")).toBeVisible();
    await page.getByTestId("login-mfa-code").fill(DEFAULT_USER.mfaCode!);
    await page.getByTestId("login-submit").click();

    await page.waitForURL("**/dashboard");
  });

  test("bad credentials surfaces an error toast", async ({ page }) => {
    await mockIamLogin(page);

    await page.goto("/login");
    await page.getByTestId("login-email").fill(DEFAULT_USER.email);
    await page.getByTestId("login-password").fill("wrong-password");
    await page.getByTestId("login-submit").click();

    await expect(page.getByRole("alert")).toContainText(/incorrect/i);
  });

  test("password reset request shows the constant-time confirmation", async ({ page }) => {
    await mockIamReset(page);

    await page.goto("/reset-password");
    await page.getByTestId("reset-email").fill("ghost@example.com");
    await page.getByTestId("reset-request-submit").click();

    await expect(page.getByText(/reset link is on its way/i)).toBeVisible();
  });

  test("password reset confirm with token updates the password", async ({ page }) => {
    await mockIamReset(page);

    await page.goto("/reset-password?token=test-token");
    await page.getByTestId("reset-new-password").fill("a-strong-password-12345");
    await page.getByTestId("reset-confirm-password").fill("a-strong-password-12345");
    await page.getByTestId("reset-submit").click();

    await expect(page.getByText(/password updated/i)).toBeVisible();
  });
});
