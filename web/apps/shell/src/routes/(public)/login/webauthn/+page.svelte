<!--
  WebAuthn passkey sign-in. REQ-FUNC-PLT-IAM-005.

  Flow:
    1. User enters email.
    2. POST /v1/iam/webauthn/assert/begin → server returns
       PublicKeyCredentialRequestOptions + a session_token binding
       this challenge to the email.
    3. navigator.credentials.get(publicKey) prompts the platform
       authenticator (TouchID, Windows Hello, security key).
    4. POST /v1/iam/webauthn/assert/finish with the resulting
       credential. The server runs the W3C-conformant signature
       check + clone-detection branch (see services/iam/internal/webauthn).
       On success, returns a LoginResponse with the access_token.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import * as iam from "@chetana/api-client/iam";

  let email = $state("");
  let isLoading = $state(false);
  let error = $state<string | null>(null);

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

  async function submit(e: Event) {
    e.preventDefault();
    error = null;
    isLoading = true;
    try {
      const begin = await iam.webauthnAssertBegin(email);
      const publicKey: PublicKeyCredentialRequestOptions = {
        ...begin.publicKey,
        challenge: b64urlToBuf(begin.publicKey.challenge),
        allowCredentials: begin.publicKey.allowCredentials?.map((c) => ({
          id: b64urlToBuf(c.id),
          type: c.type,
        })),
      };
      const cred = (await navigator.credentials.get({ publicKey })) as PublicKeyCredential | null;
      if (!cred) throw new Error("No credential returned by the authenticator.");
      const resp = cred.response as AuthenticatorAssertionResponse;
      const credentialJSON = {
        id: cred.id,
        rawId: bufToB64url(cred.rawId),
        type: cred.type,
        response: {
          authenticatorData: bufToB64url(resp.authenticatorData),
          clientDataJSON: bufToB64url(resp.clientDataJSON),
          signature: bufToB64url(resp.signature),
          userHandle: resp.userHandle ? bufToB64url(resp.userHandle) : null,
        },
      };
      const finish = await iam.webauthnAssertFinish(begin.session_token, credentialJSON);
      if (finish.status === "ok" && finish.access_token) {
        sessionStorage.setItem("chetana.access_token", finish.access_token);
        if (finish.refresh_token) {
          sessionStorage.setItem("chetana.refresh_token", finish.refresh_token);
        }
        await goto("/dashboard");
        return;
      }
      error = finish.reason ?? "Sign-in failed.";
    } catch (err) {
      error = (err as Error).message ?? "Passkey sign-in failed.";
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Sign in with passkey — Chetana</title>
</svelte:head>

<div class="flex flex-col gap-md">
  <h2 class="text-xl font-semibold text-text-primary text-center">
    Sign in with passkey
  </h2>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}

  <form onsubmit={submit} class="flex flex-col gap-md">
    <label class="flex flex-col gap-2xs text-sm">
      <span class="text-text-secondary">Email</span>
      <input type="email" required bind:value={email} class="input-field" data-testid="webauthn-email" />
    </label>
    <button
      type="submit"
      disabled={isLoading}
      class="bg-primary text-on-primary py-sm rounded font-medium disabled:opacity-60"
      data-testid="webauthn-submit"
    >
      {isLoading ? "Waiting for authenticator…" : "Continue with passkey"}
    </button>
    <a href="/login" class="text-center text-xs text-text-muted hover:text-text-secondary">
      Use email + password instead
    </a>
  </form>
</div>

<style>
  .input-field {
    @apply px-md py-sm border border-border rounded bg-surface text-text-primary
           focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20;
  }
</style>
