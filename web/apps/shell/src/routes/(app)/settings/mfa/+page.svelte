<!--
  Settings → MFA. REQ-FUNC-PLT-IAM-004 + REQ-FUNC-PLT-IAM-005.

  Two enrollment paths:

    • TOTP (authenticator app): /v1/iam/mfa/enroll → server returns
      base32 secret + otpauth:// QR URI + 10 backup codes. User
      confirms enrollment by submitting a 6-digit code.

    • WebAuthn (passkey / security key): /v1/iam/webauthn/register/begin
      → navigator.credentials.create() → /finish.

  REQ-FUNC-PLT-IAM-004: WebAuthn registration uses platform
  authenticator on supporting browsers (acceptance #4 of WEB-001).
-->
<script lang="ts">
  import { onMount } from "svelte";
  import * as iam from "@chetana/api-client/iam";

  // TOTP enrollment state
  let totpEnrollment = $state<iam.MfaEnrollResponse | null>(null);
  let totpCode = $state("");
  let isEnrolling = $state(false);
  let isVerifying = $state(false);
  let error = $state<string | null>(null);
  let success = $state<string | null>(null);

  // WebAuthn registration state
  let isRegisteringWebAuthn = $state(false);
  let webauthnSupported = $state(false);

  function bearer(): string {
    return sessionStorage.getItem("chetana.access_token") ?? "";
  }

  function b64urlToBuf(s: string): ArrayBuffer {
    const pad = "=".repeat((4 - (s.length % 4)) % 4);
    const b = atob(s.replace(/-/g, "+").replace(/_/g, "/") + pad);
    const buf = new Uint8Array(b.length);
    for (let i = 0; i < b.length; i++) buf[i] = b.charCodeAt(i);
    return buf.buffer;
  }

  function bufToB64url(buf: ArrayBuffer): string {
    const bytes = new Uint8Array(buf);
    let s = "";
    for (let i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
    return btoa(s).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  }

  async function startTotpEnroll() {
    error = null;
    success = null;
    isEnrolling = true;
    try {
      totpEnrollment = await iam.enrollMfa(bearer());
    } catch (err) {
      error = (err as Error).message ?? "Failed to start enrollment.";
    } finally {
      isEnrolling = false;
    }
  }

  async function verifyTotp(e: Event) {
    e.preventDefault();
    error = null;
    isVerifying = true;
    try {
      await iam.verifyMfa(bearer(), totpCode);
      success = "TOTP enrollment confirmed. MFA is now active.";
      totpEnrollment = null;
      totpCode = "";
    } catch (err) {
      error = (err as Error).message ?? "Verification failed.";
    } finally {
      isVerifying = false;
    }
  }

  async function registerWebAuthn() {
    error = null;
    success = null;
    isRegisteringWebAuthn = true;
    try {
      const begin = await iam.webauthnRegisterBegin(bearer());
      const publicKey: PublicKeyCredentialCreationOptions = {
        ...begin.publicKey,
        challenge: b64urlToBuf(begin.publicKey.challenge),
        user: {
          ...begin.publicKey.user,
          id: b64urlToBuf(begin.publicKey.user.id),
        },
        excludeCredentials: begin.publicKey.excludeCredentials?.map((c) => ({
          id: b64urlToBuf(c.id),
          type: c.type,
        })),
      };
      const cred = (await navigator.credentials.create({ publicKey })) as PublicKeyCredential | null;
      if (!cred) throw new Error("No credential returned by the authenticator.");
      const resp = cred.response as AuthenticatorAttestationResponse;
      const credentialJSON = {
        id: cred.id,
        rawId: bufToB64url(cred.rawId),
        type: cred.type,
        response: {
          attestationObject: bufToB64url(resp.attestationObject),
          clientDataJSON: bufToB64url(resp.clientDataJSON),
        },
      };
      await iam.webauthnRegisterFinish(bearer(), begin.session_token, credentialJSON);
      success = "Passkey registered.";
    } catch (err) {
      error = (err as Error).message ?? "Passkey registration failed.";
    } finally {
      isRegisteringWebAuthn = false;
    }
  }

  onMount(() => {
    webauthnSupported = typeof window !== "undefined" && !!window.PublicKeyCredential;
  });
</script>

<svelte:head><title>MFA — Chetana</title></svelte:head>

<div class="flex flex-col gap-lg max-w-3xl">
  <div>
    <h1 class="text-xl font-semibold text-text-primary">Multi-factor authentication</h1>
    <p class="text-sm text-text-muted mt-2xs">
      Add a second factor to your sign-in. We strongly recommend either a
      passkey (TouchID / Windows Hello / a hardware security key) or an
      authenticator app.
    </p>
  </div>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}
  {#if success}
    <div class="px-md py-sm rounded border border-success/40 bg-success/10 text-sm text-success">
      {success}
    </div>
  {/if}

  <!-- WebAuthn / passkey -->
  <section class="border border-border rounded p-md bg-surface flex flex-col gap-md">
    <h2 class="text-sm font-semibold text-text-secondary">Passkey</h2>
    {#if !webauthnSupported}
      <p class="text-sm text-text-muted">
        Your browser does not support WebAuthn. Use an authenticator app below.
      </p>
    {:else}
      <p class="text-sm text-text-secondary">
        Use your device's built-in authenticator (TouchID, Windows Hello) or a
        hardware security key.
      </p>
      <button
        type="button"
        disabled={isRegisteringWebAuthn}
        onclick={registerWebAuthn}
        class="self-start px-md py-xs text-sm rounded bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-60"
        data-testid="webauthn-register"
      >
        {isRegisteringWebAuthn ? "Waiting for authenticator…" : "Register passkey"}
      </button>
    {/if}
  </section>

  <!-- TOTP -->
  <section class="border border-border rounded p-md bg-surface flex flex-col gap-md">
    <h2 class="text-sm font-semibold text-text-secondary">Authenticator app (TOTP)</h2>
    {#if !totpEnrollment}
      <p class="text-sm text-text-secondary">
        Use Google Authenticator, Authy, 1Password, or another RFC 6238 client.
      </p>
      <button
        type="button"
        disabled={isEnrolling}
        onclick={startTotpEnroll}
        class="self-start px-md py-xs text-sm rounded bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-60"
        data-testid="totp-enroll"
      >
        {isEnrolling ? "Generating…" : "Start enrollment"}
      </button>
    {:else}
      <p class="text-sm text-text-secondary">
        Scan the QR or enter the secret manually, then submit a 6-digit code to confirm.
      </p>
      <div class="flex flex-col gap-sm">
        <code class="text-xs break-all px-md py-sm bg-background border border-border rounded" data-testid="totp-secret">
          {totpEnrollment.secret_base32}
        </code>
        <details class="text-xs text-text-muted">
          <summary>Backup codes (save these in a password manager)</summary>
          <pre class="mt-xs px-md py-sm bg-background border border-border rounded">{totpEnrollment.backup_codes.join("\n")}</pre>
        </details>
      </div>
      <form onsubmit={verifyTotp} class="flex gap-sm items-end">
        <label class="flex flex-col gap-2xs text-sm flex-1 max-w-xs">
          <span class="text-text-secondary">Code</span>
          <input
            type="text"
            inputmode="numeric"
            pattern="[0-9]{6}"
            maxlength="6"
            required
            bind:value={totpCode}
            class="input-field tracking-widest text-center"
            data-testid="totp-code"
          />
        </label>
        <button
          type="submit"
          disabled={isVerifying}
          class="px-md py-sm text-sm rounded bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-60"
          data-testid="totp-verify"
        >
          {isVerifying ? "Verifying…" : "Confirm"}
        </button>
      </form>
    {/if}
  </section>
</div>

<style>
  .input-field {
    @apply px-md py-sm border border-border rounded bg-surface text-text-primary
           focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20;
  }
</style>
