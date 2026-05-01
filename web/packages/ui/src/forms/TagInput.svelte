<script context="module" lang="ts">
  export interface Tag {
    id: string;
    label: string;
    color?: string;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: Tag[] = [];
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let placeholder: string = 'Add tag...';
  export let maxTags: number = 0; // 0 = unlimited
  export let allowDuplicates: boolean = false;
  export let suggestions: string[] = [];
  export let allowCreate: boolean = true;
  export let name: string = '';
  export let id: string = uid('tag');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    add: { tag: Tag };
    remove: { tag: Tag };
    change: { value: Tag[] };
  }>();

  let inputValue = '';
  let inputRef: HTMLInputElement;
  let showSuggestions = false;
  let selectedSuggestionIndex = -1;

  // Size configurations
  const sizeConfig = {
    sm: { container: 'min-h-8 text-sm', tag: 'text-xs px-2 py-0.5', input: 'text-sm' },
    md: { container: 'min-h-10 text-base', tag: 'text-sm px-2.5 py-1', input: 'text-base' },
    lg: { container: 'min-h-12 text-lg', tag: 'text-base px-3 py-1.5', input: 'text-lg' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  $: filteredSuggestions = suggestions.filter(s => {
    const lowerInput = inputValue.toLowerCase();
    const lowerSuggestion = s.toLowerCase();
    const alreadyAdded = value.some(t => t.label.toLowerCase() === lowerSuggestion);
    return lowerSuggestion.includes(lowerInput) && (!alreadyAdded || allowDuplicates);
  });

  function generateId(): string {
    return `tag-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  function addTag(label: string) {
    const trimmed = label.trim();
    if (!trimmed) return;

    // Check max tags
    if (maxTags > 0 && value.length >= maxTags) return;

    // Check duplicates
    if (!allowDuplicates && value.some(t => t.label.toLowerCase() === trimmed.toLowerCase())) {
      return;
    }

    const newTag: Tag = {
      id: generateId(),
      label: trimmed,
    };

    value = [...value, newTag];
    inputValue = '';
    showSuggestions = false;
    selectedSuggestionIndex = -1;

    dispatch('add', { tag: newTag });
    dispatch('change', { value });
  }

  function removeTag(tag: Tag) {
    if (disabled) return;

    value = value.filter(t => t.id !== tag.id);
    dispatch('remove', { tag });
    dispatch('change', { value });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      if (selectedSuggestionIndex >= 0 && filteredSuggestions.length > 0) {
        addTag(filteredSuggestions[selectedSuggestionIndex]!);
      } else if (allowCreate && inputValue.trim()) {
        addTag(inputValue);
      }
    } else if (event.key === 'Backspace' && !inputValue && value.length > 0) {
      removeTag(value[value.length - 1]!);
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      if (filteredSuggestions.length > 0) {
        selectedSuggestionIndex = Math.min(selectedSuggestionIndex + 1, filteredSuggestions.length - 1);
      }
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      selectedSuggestionIndex = Math.max(selectedSuggestionIndex - 1, -1);
    } else if (event.key === 'Escape') {
      showSuggestions = false;
      selectedSuggestionIndex = -1;
    } else if (event.key === ',' || event.key === 'Tab') {
      if (inputValue.trim()) {
        event.preventDefault();
        addTag(inputValue);
      }
    }
  }

  function handleInput() {
    showSuggestions = true;
    selectedSuggestionIndex = -1;
  }

  function handleFocus() {
    showSuggestions = true;
  }

  function handleBlur() {
    // Delay to allow click on suggestion
    setTimeout(() => {
      showSuggestions = false;
    }, 150);
  }

  function selectSuggestion(suggestion: string) {
    addTag(suggestion);
    inputRef?.focus();
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label
      for={id}
      class="block mb-2 font-medium text-[var(--color-text-primary)]"
    >
      {label}
    </label>
  {/if}

  <div class="relative">
    <div
      class={cn(
        'flex flex-wrap items-center gap-2 p-2 rounded-lg',
        'border border-[var(--color-border-primary)]',
        'bg-[var(--color-surface-primary)]',
        'focus-within:ring-2 focus-within:ring-[var(--color-interactive-primary)]',
        config.container,
        disabled && 'opacity-50 cursor-not-allowed'
      )}
    >
      {#each value as tag (tag.id)}
        <span
          class={cn(
            'inline-flex items-center gap-1 rounded-full',
            'bg-[var(--color-interactive-primary)] text-white',
            config.tag
          )}
          style={tag.color ? `background-color: ${tag.color}` : ''}
        >
          {tag.label}
          {#if !disabled}
            <button
              type="button"
              class="hover:bg-white/20 rounded-full p-0.5"
              on:click={() => removeTag(tag)}
              aria-label="Remove {tag.label}"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          {/if}
        </span>
      {/each}

      {#if !disabled && (maxTags === 0 || value.length < maxTags)}
        <input
          bind:this={inputRef}
          type="text"
          {id}
          {name}
          bind:value={inputValue}
          {placeholder}
          {disabled}
          class={cn(
            'flex-1 min-w-[120px] bg-transparent border-none',
            'focus:outline-none',
            config.input
          )}
          data-testid={testId || undefined}
          on:keydown={handleKeydown}
          on:input={handleInput}
          on:focus={handleFocus}
          on:blur={handleBlur}
        />
      {/if}
    </div>

    {#if showSuggestions && filteredSuggestions.length > 0 && !disabled}
      <div
        class="absolute z-50 mt-1 w-full max-h-48 overflow-auto bg-[var(--color-surface-primary)] border border-[var(--color-border-primary)] rounded-lg shadow-lg"
      >
        {#each filteredSuggestions as suggestion, i}
          <button
            type="button"
            class={cn(
              'w-full px-3 py-2 text-left text-sm',
              'hover:bg-[var(--color-surface-secondary)]',
              'focus:outline-none focus:bg-[var(--color-surface-secondary)]',
              i === selectedSuggestionIndex && 'bg-[var(--color-surface-secondary)]'
            )}
            on:click={() => selectSuggestion(suggestion)}
          >
            {suggestion}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  {#if maxTags > 0}
    <div class="mt-1 text-xs text-[var(--color-text-tertiary)]">
      {value.length} / {maxTags} tags
    </div>
  {/if}
</div>
