<script lang="ts">
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';

  // Temporarily bypass auth - go directly to dashboard
  $effect(() => {
    if (browser) {
      goto('/dashboard');
    }
  });
</script>

<div class="loading-container">
  <div class="loading-spinner"></div>
  <p>Loading...</p>
</div>

<style>
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    gap: var(--spacing-md);
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  p {
    color: var(--color-text-secondary);
    font-size: 0.875rem;
  }
</style>
