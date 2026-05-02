<script lang="ts">
  /**
   * HiddenField — UI for `HiddenField` (a form value that's part of
   * the schema but not rendered to the user). Replaces the placeholder
   * mapping that used to route `hidden` to `Input` (which would render
   * a visible text input — defeating the field's purpose).
   *
   * Hidden fields exist for: (a) backend-managed IDs the form must
   * round-trip but the user shouldn't edit (created_by, version),
   * (b) computed values populated by FORMULA fields and submitted
   * alongside, (c) state machine markers (current_step, draft_id).
   *
   * Renders an `<input type="hidden">` so the field appears in the form's
   * native FormData when submitted via a non-AJAX form post (browser
   * accessibility tools also rely on this shape).
   */
  import type { HiddenField } from '@chetana/core';

  interface Props {
    field: HiddenField;
    value?: unknown;
  }

  let { field, value }: Props = $props();
</script>

<input type="hidden" name={field.name} value={String(value ?? '')} />
