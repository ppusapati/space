<script lang="ts">
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { formFieldClasses } from './formfield.types';
  import type { ValidationState } from '../types';

  // Props
  export let label: string = '';
  export let labelPosition: 'top' | 'left' | 'right' = 'top';
  export let labelWidth: string = '120px';
  export let helperText: string = '';
  export let errorText: string = '';
  export let state: ValidationState = 'default';
  export let required: boolean = false;
  export let showOptional: boolean = false;
  export let fullWidth: boolean = true;
  export let id: string = uid('field');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  // Computed
  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;
  $: helperClasses = formFieldClasses.helper[errorText ? 'invalid' : state];

  $: containerClasses = cn(
    formFieldClasses.container,
    labelPosition === 'top'
      ? formFieldClasses.labelTop
      : labelPosition === 'left'
      ? formFieldClasses.labelLeft
      : formFieldClasses.labelRight,
    fullWidth && 'w-full',
    className
  );

  $: labelStyle = labelPosition !== 'top' ? `width: ${labelWidth}; flex-shrink: 0;` : '';
</script>

<div class={containerClasses} data-testid={testId || undefined}>
  {#if label}
    <label for={id} class={formFieldClasses.label} style={labelStyle}>
      {label}
      {#if required}
        <span class={formFieldClasses.labelRequired} aria-hidden="true">*</span>
      {:else if showOptional}
        <span class={formFieldClasses.labelOptional}>(optional)</span>
      {/if}
    </label>
  {/if}

  <div class={formFieldClasses.content}>
    <slot />

    {#if displayedHelperText}
      <p id="{id}-helper" class={helperClasses}>
        {displayedHelperText}
      </p>
    {/if}
  </div>
</div>
