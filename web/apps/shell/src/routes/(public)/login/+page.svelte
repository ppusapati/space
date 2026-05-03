<!--
  Chetana platform login.

  Two-stage UX:

    1. email + password → POST /v1/iam/login
       • status='ok'           → store the bearer + redirect /dashboard
       • status='mfa_required' → show the TOTP entry; resubmit with mfa_code
       • status='bad_credentials' / 'rate_limited' / 'locked' → error toast
    2. mfa_code + mfa_session_token → POST /v1/iam/login again

  REQ-FUNC-PLT-IAM-001 + REQ-FUNC-PLT-IAM-004: a single login UI
  surface that handles both factors.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import * as iam from "@chetana/api-client/iam";

  let email = $state("");
  let password = $state("");
  let mfaCode = $state("");
  let mfaSessionToken = $state<string | null>(null);
  let isLoading = $state(false);
  let error = $state<string | null>(null);

  function describeStatus(status: iam.LoginResponse["status"], reason?: string): string {
    switch (status) {
      case "bad_credentials":
        return "Email or password is incorrect.";
      case "rate_limited":
        return "Too many attempts. Please try again later.";
      case "locked":
        return "Account locked. Contact your administrator.";
      case "internal_error":
        return reason || "Sign-in failed. Please try again.";
      default:
        return reason || "Sign-in failed.";
    }
  }

  async function submit(e: Event) {
    e.preventDefault();
    error = null;
    isLoading = true;
    try {
      const res = await iam.login({
        email,
        password,
        mfa_code: mfaSessionToken ? mfaCode : undefined,
        mfa_session_token: mfaSessionToken ?? undefined,
      });

      if (res.status === "mfa_required") {
        mfaSessionToken = res.mfa_session_token ?? null;
        return;
      }
      if (res.status === "ok" && res.access_token) {
        // The chetana cmd layer also sets a session cookie via Set-Cookie;
        // the bearer is mirrored into sessionStorage so the api-client
        // can stamp Authorization headers from non-cookie contexts (the
        // realtime WS opens via sub-protocol with the bearer baked in).
        sessionStorage.setItem("chetana.access_token", res.access_token);
        if (res.refresh_token) {
          sessionStorage.setItem("chetana.refresh_token", res.refresh_token);
        }
        await goto("/dashboard");
        return;
      }
      error = describeStatus(res.status, res.reason);
    } catch (err) {
      error = (err as Error).message ?? "Sign-in failed.";
    } finally {
      isLoading = false;
    }
  }

  function backToCredentials() {
    mfaSessionToken = null;
    mfaCode = "";
    error = null;
  }
</script>

<svelte:head>
  <title>Sign in — Chetana</title>
</svelte:head>

<div class="flex flex-col gap-md">
  <h2 class="text-xl font-semibold text-text-primary text-center">
    {mfaSessionToken ? "Two-factor verification" : "Sign in"}
  </h2>

  {#if error}
    <div
      role="alert"
      class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error"
    >
      {error}
    </div>
  {/if}

  <form onsubmit={submit} class="flex flex-col gap-md">
    {#if !mfaSessionToken}
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">Email</span>
        <input
          type="email"
          autocomplete="username"
          required
          bind:value={email}
          class="input-field"
          data-testid="login-email"
        />
      </label>
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">Password</span>
        <input
          type="password"
          autocomplete="current-password"
          required
          bind:value={password}
          class="input-field"
          data-testid="login-password"
        />
      </label>
    {:else}
      <p class="text-sm text-text-secondary">
        Enter the 6-digit code from your authenticator app.
      </p>
      <label class="flex flex-col gap-2xs text-sm">
        <span class="text-text-secondary">Authenticator code</span>
        <input
          type="text"
          inputmode="numeric"
          pattern="[0-9]{6}"
          maxlength="6"
          autocomplete="one-time-code"
          required
          bind:value={mfaCode}
          class="input-field tracking-widest text-center text-lg"
          data-testid="login-mfa-code"
        />
      </label>
    {/if}

    <button
      type="submit"
      disabled={isLoading}
      class="bg-primary text-on-primary py-sm rounded font-medium hover:bg-primary/90 disabled:opacity-60"
      data-testid="login-submit"
    >
      {isLoading ? "Signing in…" : mfaSessionToken ? "Verify" : "Sign in"}
    </button>

    {#if mfaSessionToken}
      <button
        type="button"
        onclick={backToCredentials}
        class="text-xs text-text-muted hover:text-text-secondary"
      >
        Back to email + password
      </button>
    {:else}
      <div class="flex justify-between text-xs text-text-muted">
        <a href="/reset-password" class="hover:text-text-secondary">Forgot password?</a>
        <a href="/login/webauthn" class="hover:text-text-secondary">Sign in with passkey</a>
      </div>
    {/if}
  </form>
</div>

<style>
  .input-field {
    @apply px-md py-sm border border-border rounded bg-surface text-text-primary
           focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20;
  }
</style>
