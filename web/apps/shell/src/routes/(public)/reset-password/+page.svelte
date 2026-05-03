<!--
  Password reset. Two views:

    1. Request: email → POST /v1/iam/reset/request. Constant-time
       response per REQ-FUNC-PLT-IAM-010 — we ALWAYS show the
       confirmation card regardless of whether the email exists.

    2. Confirm: token (from URL ?token=…) + new password → POST
       /v1/iam/reset/confirm.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import * as iam from "@chetana/api-client/iam";

  const tokenFromURL = $derived($page.url.searchParams.get("token"));

  let email = $state("");
  let newPassword = $state("");
  let confirmPassword = $state("");
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let submitted = $state(false);

  async function requestReset(e: Event) {
    e.preventDefault();
    error = null;
    isLoading = true;
    try {
      await iam.requestPasswordReset(email);
      submitted = true;
    } catch (err) {
      // Constant-time response policy → still show the confirm
      // card on transport errors so the email-existence channel
      // stays sealed.
      submitted = true;
      console.warn("reset request failed (will show success anyway)", err);
    } finally {
      isLoading = false;
    }
  }

  async function confirmReset(e: Event) {
    e.preventDefault();
    error = null;
    if (newPassword !== confirmPassword) {
      error = "Passwords do not match.";
      return;
    }
    if (newPassword.length < 12) {
      error = "Password must be at least 12 characters.";
      return;
    }
    isLoading = true;
    try {
      await iam.confirmPasswordReset(tokenFromURL!, newPassword);
      submitted = true;
    } catch (err) {
      error = (err as Error).message ?? "Reset failed. The link may be expired.";
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Reset password — Chetana</title>
</svelte:head>

<div class="flex flex-col gap-md">
  <h2 class="text-xl font-semibold text-text-primary text-center">
    {tokenFromURL ? "Choose a new password" : "Reset password"}
  </h2>

  {#if submitted && tokenFromURL}
    <div class="px-md py-sm rounded border border-success/40 bg-success/10 text-sm text-success">
      Password updated. You can now <a href="/login" class="underline">sign in</a>.
    </div>
  {:else if submitted}
    <div class="px-md py-sm rounded border border-success/40 bg-success/10 text-sm text-success">
      If an account exists for that email, a reset link is on its way.
    </div>
    <a href="/login" class="text-center text-xs text-text-muted hover:text-text-secondary">
      Back to sign in
    </a>
  {:else if tokenFromURL}
    {#if error}
      <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
        {error}
      </div>
    {/if}
    <form onsubmit={confirmReset} class="flex flex-col gap-md">
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">New password</span>
        <input
          type="password"
          autocomplete="new-password"
          required
          bind:value={newPassword}
          class="input-field"
          data-testid="reset-new-password"
        />
      </label>
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">Confirm new password</span>
        <input
          type="password"
          autocomplete="new-password"
          required
          bind:value={confirmPassword}
          class="input-field"
          data-testid="reset-confirm-password"
        />
      </label>
      <button
        type="submit"
        disabled={isLoading}
        class="bg-primary text-on-primary py-sm rounded font-medium disabled:opacity-60"
        data-testid="reset-submit"
      >
        {isLoading ? "Updating…" : "Update password"}
      </button>
    </form>
  {:else}
    {#if error}
      <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
        {error}
      </div>
    {/if}
    <form onsubmit={requestReset} class="flex flex-col gap-md">
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">Email</span>
        <input
          type="email"
          required
          bind:value={email}
          class="input-field"
          data-testid="reset-email"
        />
      </label>
      <button
        type="submit"
        disabled={isLoading}
        class="bg-primary text-on-primary py-sm rounded font-medium disabled:opacity-60"
        data-testid="reset-request-submit"
      >
        {isLoading ? "Sending…" : "Send reset link"}
      </button>
      <a href="/login" class="text-center text-xs text-text-muted hover:text-text-secondary">
        Back to sign in
      </a>
    </form>
  {/if}
</div>

<style>
  .input-field {
    @apply px-md py-sm border border-border rounded bg-surface text-text-primary
           focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20;
  }
</style>
