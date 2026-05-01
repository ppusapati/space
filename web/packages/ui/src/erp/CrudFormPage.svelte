<script lang="ts">
  import type { FormSchema, FormFieldConfig } from '@samavāya/core';
  import DynamicFormRenderer from '../forms/DynamicFormRenderer.svelte';

  /** Page title */
  export let title: string;
  /** Optional subtitle */
  export let subtitle: string = '';
  /** Mode: create, edit, or view */
  export let mode: 'create' | 'edit' | 'view' = 'create';
  /** Form schema definition */
  export let schema: FormSchema<Record<string, unknown>>;
  /** Initial/current form values */
  export let values: Record<string, unknown> = {};
  /** Validation errors */
  export let errors: Record<string, string> = {};
  /** Loading state */
  export let isLoading: boolean = false;
  /** Submitting state */
  export let isSubmitting: boolean = false;
  /** Error message */
  export let error: string | null = null;
  /** Submit button label */
  export let submitLabel: string | undefined = undefined;
  /** Cancel URL (for back navigation) */
  export let cancelHref: string = '';
  /** Show delete button in edit mode */
  export let showDelete: boolean = false;

  /** Submit handler */
  export let onSubmit: ((values: Record<string, unknown>) => void | Promise<void>) | null = null;
  /** Cancel handler (alternative to cancelHref) */
  export let onCancel: (() => void) | null = null;
  /** Delete handler */
  export let onDelete: (() => void | Promise<void>) | null = null;

  let isDeleting = false;

  const defaultSubmitLabel = mode === 'create' ? 'Create' : mode === 'edit' ? 'Save Changes' : '';

  async function handleSubmit(formValues: Record<string, unknown>) {
    if (onSubmit) {
      await onSubmit(formValues);
    }
  }

  async function handleDelete() {
    if (!onDelete) return;
    if (!confirm('Are you sure you want to delete this? This action cannot be undone.')) return;
    isDeleting = true;
    try {
      await onDelete();
    } finally {
      isDeleting = false;
    }
  }
</script>

<div class="crud-form-page">
  <header class="crud-form-header">
    <div class="crud-form-header-left">
      {#if cancelHref}
        <a href={cancelHref} class="crud-back-link" aria-label="Go back">&larr;</a>
      {:else if onCancel}
        <button type="button" class="crud-back-link" on:click={onCancel} aria-label="Go back">&larr;</button>
      {/if}
      <div>
        <h1 class="crud-form-title">{title}</h1>
        {#if subtitle}
          <p class="crud-form-subtitle">{subtitle}</p>
        {/if}
      </div>
    </div>
    <div class="crud-form-header-right">
      {#if mode === 'edit' && showDelete && onDelete}
        <button
          type="button"
          class="crud-delete-btn"
          disabled={isDeleting}
          on:click={handleDelete}
        >
          {isDeleting ? 'Deleting...' : 'Delete'}
        </button>
      {/if}
    </div>
  </header>

  {#if error}
    <div class="crud-form-error" role="alert">
      <p>{error}</p>
    </div>
  {/if}

  {#if isLoading}
    <div class="crud-form-loading">
      <p>Loading...</p>
    </div>
  {:else}
    <div class="crud-form-body">
      <DynamicFormRenderer
        {schema}
        {values}
        {errors}
        readonly={mode === 'view'}
        disabled={isSubmitting}
        submitLabel={submitLabel || defaultSubmitLabel}
        showReset={mode !== 'view'}
        resetLabel="Cancel"
        onSubmit={handleSubmit}
        onReset={onCancel}
      />
    </div>
  {/if}
</div>

<style>
  .crud-form-page {
    max-width: 900px;
  }

  .crud-form-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .crud-form-header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .crud-back-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border-radius: var(--radius-md, 0.375rem);
    border: 1px solid var(--color-border, #e5e7eb);
    background: var(--color-background, #fff);
    color: var(--color-foreground, #111);
    text-decoration: none;
    font-size: 1.125rem;
    cursor: pointer;
  }

  .crud-back-link:hover {
    background: var(--color-muted, #f3f4f6);
  }

  .crud-form-title {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
  }

  .crud-form-subtitle {
    font-size: 0.875rem;
    color: var(--color-muted-foreground, #6b7280);
    margin: 0.25rem 0 0 0;
  }

  .crud-delete-btn {
    padding: 0.5rem 1rem;
    border-radius: var(--radius-md, 0.375rem);
    border: 1px solid var(--color-destructive, #ef4444);
    background: transparent;
    color: var(--color-destructive, #ef4444);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
  }

  .crud-delete-btn:hover {
    background: var(--color-destructive, #ef4444);
    color: var(--color-destructive-foreground, #fff);
  }

  .crud-delete-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .crud-form-error {
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
    border-radius: var(--radius-md, 0.375rem);
    background: var(--color-destructive-muted, #fef2f2);
    color: var(--color-destructive, #ef4444);
    border: 1px solid var(--color-destructive, #ef4444);
    font-size: 0.875rem;
  }

  .crud-form-error p {
    margin: 0;
  }

  .crud-form-loading {
    text-align: center;
    padding: 3rem;
    background-color: var(--color-card, #fff);
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-lg, 0.5rem);
  }

  .crud-form-body {
    background-color: var(--color-card, #fff);
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-lg, 0.5rem);
    padding: 1.5rem;
  }
</style>
