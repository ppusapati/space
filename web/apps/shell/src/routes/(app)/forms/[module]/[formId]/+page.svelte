<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { FormPage } from '@chetana/ui';

  const moduleId = $derived($page.params.module ?? '');
  const formId = $derived($page.params.formId ?? '');

  function handleSuccess(entityId: string) {
    goto(`/forms/${moduleId}/${formId}?created=${entityId}`);
  }

  function handleError(errorMsg: string) {
    console.error('[forms/[module]/[formId]] Submission error:', errorMsg);
  }
</script>

{#if formId}
  <div class="form-wrapper">
    <div class="toolbar">
      <a class="secondary-link" href={`/forms/${moduleId}/${formId}/submissions`}>
        View submission history
      </a>
    </div>
    <FormPage
      {formId}
      mode="create"
      cancelHref={`/forms/${moduleId}`}
      onSuccess={handleSuccess}
      onError={handleError}
    />
  </div>
{/if}

<style>
  .form-wrapper {
    position: relative;
  }

  .toolbar {
    display: flex;
    justify-content: flex-end;
    padding: 0.5rem 1.5rem 0;
    max-width: 1100px;
    margin: 0 auto;
  }

  .secondary-link {
    font-size: 0.85rem;
    color: var(--color-accent, #2563eb);
    text-decoration: none;
  }

  .secondary-link:hover {
    text-decoration: underline;
  }
</style>
