<script lang="ts">
  import { onMount } from 'svelte';
  import { cn } from '../utils';

  interface Props {
    /** Custom fallback content when error occurs */
    fallback?: import('svelte').Snippet<[{ error: Error; reset: () => void }]>;
    /** Called when an error is caught */
    onError?: (error: Error, errorInfo: { componentStack?: string }) => void;
    /** Show detailed error info (only in development) */
    showDetails?: boolean;
    /** Custom class for the error container */
    class?: string;
    /** Children content */
    children?: import('svelte').Snippet;
  }

  const {
    fallback,
    onError,
    showDetails = false,
    class: className = '',
    children,
  }: Props = $props();

  let error = $state<Error | null>(null);
  let errorInfo = $state<{ componentStack?: string } | null>(null);
  let hasError = $state(false);

  function handleError(e: ErrorEvent | PromiseRejectionEvent) {
    let caughtError: Error;

    if (e instanceof ErrorEvent) {
      caughtError = e.error || new Error(e.message);
    } else {
      caughtError = e.reason instanceof Error ? e.reason : new Error(String(e.reason));
    }

    error = caughtError;
    errorInfo = { componentStack: caughtError.stack };
    hasError = true;

    onError?.(caughtError, { componentStack: caughtError.stack });
  }

  function reset() {
    error = null;
    errorInfo = null;
    hasError = false;
  }

  onMount(() => {
    // Catch unhandled errors
    const errorHandler = (e: ErrorEvent) => {
      e.preventDefault();
      handleError(e);
    };

    // Catch unhandled promise rejections
    const rejectionHandler = (e: PromiseRejectionEvent) => {
      e.preventDefault();
      handleError(e);
    };

    window.addEventListener('error', errorHandler);
    window.addEventListener('unhandledrejection', rejectionHandler);

    return () => {
      window.removeEventListener('error', errorHandler);
      window.removeEventListener('unhandledrejection', rejectionHandler);
    };
  });

  const containerClasses = 'flex flex-col items-center justify-center min-h-48 p-6 rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950';
  const iconClasses = 'w-12 h-12 text-red-500 dark:text-red-400 mb-4';
  const titleClasses = 'text-lg font-semibold text-red-700 dark:text-red-300 mb-2';
  const messageClasses = 'text-sm text-red-600 dark:text-red-400 text-center max-w-md mb-4';
  const detailsClasses = 'w-full max-w-2xl p-4 mt-4 text-xs font-mono bg-red-100 dark:bg-red-900 rounded border border-red-200 dark:border-red-800 overflow-auto max-h-48';
  const buttonClasses = 'px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2';
</script>

{#if hasError && error}
  {#if fallback}
    {@render fallback({ error, reset })}
  {:else}
    <div class={cn(containerClasses, className)} role="alert">
      <svg
        class={iconClasses}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>

      <h2 class={titleClasses}>Something went wrong</h2>

      <p class={messageClasses}>
        {error.message || 'An unexpected error occurred. Please try again.'}
      </p>

      <button
        type="button"
        class={buttonClasses}
        onclick={reset}
      >
        Try again
      </button>

      {#if showDetails && errorInfo?.componentStack}
        <details class={detailsClasses}>
          <summary class="cursor-pointer font-semibold mb-2">Error Details</summary>
          <pre class="whitespace-pre-wrap break-words text-red-800 dark:text-red-200">{error.name}: {error.message}

Stack Trace:
{errorInfo.componentStack}</pre>
        </details>
      {/if}
    </div>
  {/if}
{:else}
  {#if children}
    {@render children()}
  {/if}
{/if}
