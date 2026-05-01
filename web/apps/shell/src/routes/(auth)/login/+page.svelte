<script lang="ts">
  import { authStore } from '@samavāya/stores';
  import { goto } from '$app/navigation';

  let email = $state('');
  let password = $state('');
  let rememberMe = $state(false);
  let isLoading = $state(false);
  let error = $state<string | null>(null);

  /**
   * Extract a user-presentable error string from whatever the auth store
   * throws. Three possible shapes:
   *   1. native Error — has `.message`
   *   2. ApiError plain object from packages/api — has `.code` + `.message`
   *      (thrown verbatim by the error.interceptor)
   *   3. anything else — defensive fallback string
   * Without this, ApiError plain objects coerce to "[object Object]" via
   * template-string interpolation, which is what the user actually saw.
   */
  function describeError(err: unknown): string {
    if (err instanceof Error) return err.message;
    if (typeof err === 'object' && err !== null) {
      const e = err as { code?: string; message?: string };
      if (typeof e.message === 'string' && e.message.length > 0) return e.message;
      if (typeof e.code === 'string' && e.code.length > 0) return e.code;
    }
    if (typeof err === 'string' && err.length > 0) return err;
    return 'Login failed. Please try again.';
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = null;
    isLoading = true;

    try {
      await authStore.login({ email, password, rememberMe });
      goto('/dashboard');
    } catch (err) {
      error = describeError(err);
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Login - samavāya ERP</title>
</svelte:head>

<div class="w-full">
  <h2 class="text-2xl font-semibold text-text mb-xs">Welcome back</h2>
  <p class="text-text-secondary mb-xl">Sign in to your account</p>

  <form onsubmit={handleSubmit} class="flex flex-col gap-lg">
    {#if error}
      <div class="form-error" role="alert">
        {error}
      </div>
    {/if}

    <div class="flex flex-col gap-xs">
      <label for="email" class="flex justify-between items-center text-sm font-medium text-text">Email</label>
      <input
        type="email"
        id="email"
        bind:value={email}
        required
        autocomplete="email"
        class="form-input"
        placeholder="you@company.com"
        disabled={isLoading}
      />
    </div>

    <div class="flex flex-col gap-xs">
      <label for="password" class="flex justify-between items-center text-sm font-medium text-text">
        Password
        <a href="/forgot-password" class="text-xs font-normal text-primary">Forgot password?</a>
      </label>
      <input
        type="password"
        id="password"
        bind:value={password}
        required
        autocomplete="current-password"
        class="form-input"
        placeholder="Enter your password"
        disabled={isLoading}
      />
    </div>

    <div class="flex items-center">
      <label class="flex items-center gap-sm text-sm text-text-secondary cursor-pointer">
        <input type="checkbox" bind:checked={rememberMe} disabled={isLoading} class="form-checkbox" />
        <span>Remember me</span>
      </label>
    </div>

    <button type="submit" class="btn btn-primary w-full" disabled={isLoading}>
      {#if isLoading}
        <span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
        Signing in...
      {:else}
        Sign in
      {/if}
    </button>
  </form>

  <p class="text-center mt-xl text-sm text-text-secondary">
    Don't have an account? <a href="/register" class="text-primary font-medium">Sign up</a>
  </p>
</div>
