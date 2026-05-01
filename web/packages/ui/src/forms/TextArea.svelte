<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type TextAreaResize,
    type TextAreaVariant,
    textareaBaseClasses,
    textareaSizeClasses,
    textareaVariantClasses,
    textareaStateClasses,
    textareaResizeClasses,
    textareaHelperClasses,
    charCountClasses,
    calculateAutoHeight,
  } from './textarea.types';
  import type { Size, ValidationState } from '../types';

  // Props
  export let value: string = '';
  export let placeholder: string = '';
  export let size: Size = 'md';
  export let variant: TextAreaVariant = 'default';
  export let state: ValidationState = 'default';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let rows: number = 3;
  export let minRows: number = 2;
  export let maxRows: number = 10;
  export let resize: TextAreaResize = 'vertical';
  export let maxlength: number | undefined = undefined;
  export let showCount: boolean = false;
  export let autoResize: boolean = false;
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let name: string = '';
  export let id: string = uid('textarea');
  export let testId: string = '';
  export let fullWidth: boolean = true;

  let className: string = '';
  export { className as class };

  let textareaRef: HTMLTextAreaElement;

  const dispatch = createEventDispatcher<{
    input: { value: string; event: Event };
    change: { value: string; event: Event };
    focus: { event: FocusEvent };
    blur: { event: FocusEvent };
  }>();

  // Computed classes
  $: textareaClasses = cn(
    textareaBaseClasses,
    textareaSizeClasses[size],
    textareaVariantClasses[variant],
    textareaStateClasses[state === 'default' && errorText ? 'invalid' : state],
    autoResize ? 'resize-none overflow-hidden' : textareaResizeClasses[resize],
    className
  );

  $: containerClasses = cn('relative', fullWidth ? 'w-full' : 'inline-block');

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;
  $: helperClasses = textareaHelperClasses[errorText ? 'invalid' : state];

  $: charCount = value?.length || 0;
  $: charCountText = maxlength ? `${charCount}/${maxlength}` : `${charCount}`;

  // Auto-resize handler
  function handleAutoResize() {
    if (autoResize && textareaRef) {
      calculateAutoHeight(textareaRef, minRows, maxRows);
    }
  }

  // Event handlers
  function handleInput(event: Event) {
    const target = event.target as HTMLTextAreaElement;
    value = target.value;
    dispatch('input', { value, event });
    handleAutoResize();
  }

  function handleChange(event: Event) {
    const target = event.target as HTMLTextAreaElement;
    value = target.value;
    dispatch('change', { value, event });
  }

  function handleFocus(event: FocusEvent) {
    dispatch('focus', { event });
  }

  function handleBlur(event: FocusEvent) {
    dispatch('blur', { event });
  }

  onMount(() => {
    handleAutoResize();
  });

  // React to value changes for auto-resize
  $: if (autoResize && textareaRef && value !== undefined) {
    handleAutoResize();
  }
</script>

<div class={containerClasses}>
  {#if label}
    <label for={id} class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-semantic-error-500 ml-0.5" aria-hidden="true">*</span>
      {/if}
    </label>
  {/if}

  <textarea
    bind:this={textareaRef}
    {id}
    {name}
    {placeholder}
    {disabled}
    {readonly}
    {required}
    {maxlength}
    {rows}
    {value}
    class={textareaClasses}
    data-testid={testId || undefined}
    aria-invalid={state === 'invalid' || !!errorText}
    aria-describedby={displayedHelperText ? `${id}-helper` : undefined}
    aria-required={required}
    on:input={handleInput}
    on:change={handleChange}
    on:focus={handleFocus}
    on:blur={handleBlur}
  />

  <div class="flex justify-between items-start">
    {#if displayedHelperText}
      <p id="{id}-helper" class={helperClasses}>
        {displayedHelperText}
      </p>
    {:else}
      <span></span>
    {/if}

    {#if showCount}
      <span class={charCountClasses}>{charCountText}</span>
    {/if}
  </div>
</div>
