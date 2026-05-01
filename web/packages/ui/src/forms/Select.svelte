<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { Keys, isKey } from '../utils/keyboard';
  import {
    type SelectOption,
    type SelectOptionGroup,
    type SelectVariant,
    selectBaseClasses,
    selectSizeClasses,
    selectVariantClasses,
    selectStateClasses,
    dropdownPanelClasses,
    optionBaseClasses,
    optionHoverClasses,
    optionSelectedClasses,
    optionDisabledClasses,
    groupLabelClasses,
    chevronClasses,
    searchInputClasses,
    selectHelperClasses,
    filterOptions,
    getSelectedLabel,
  } from './select.types';
  import type { Size, ValidationState } from '../types';

  type T = $$Generic;

  // Props
  export let value: T | T[] | undefined = undefined;
  export let options: SelectOption<T>[] = [];
  export let groups: SelectOptionGroup<T>[] = [];
  export let placeholder: string = 'Select an option';
  export let size: Size = 'md';
  export let variant: SelectVariant = 'default';
  export let state: ValidationState = 'default';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let multiple: boolean = false;
  export let searchable: boolean = false;
  export let clearable: boolean = false;
  export let disabled: boolean = false;
  export let required: boolean = false;
  export let readonly: boolean = false;
  export let name: string = '';
  export let id: string = uid('select');
  export let testId: string = '';
  export let fullWidth: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { value: T | T[]; option?: SelectOption<T> };
    open: void;
    close: void;
    clear: void;
  }>();

  // Internal state
  let isOpen = false;
  let searchQuery = '';
  let highlightedIndex = -1;
  let containerRef: HTMLDivElement;
  let searchInputRef: HTMLInputElement;
  let listboxRef: HTMLUListElement;

  // Computed: merge options from groups
  $: allOptions = groups.length > 0
    ? groups.flatMap((g) => g.options)
    : options;

  $: filteredOptions = searchable ? filterOptions(allOptions, searchQuery) : allOptions;

  $: displayLabel = getSelectedLabel(value, allOptions, placeholder);

  $: selectClasses = cn(
    selectBaseClasses,
    selectSizeClasses[size],
    selectVariantClasses[variant],
    selectStateClasses[state === 'default' && errorText ? 'invalid' : state],
    className
  );

  $: containerClasses = cn(
    'relative',
    fullWidth ? 'w-full' : 'inline-block'
  );

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;
  $: helperClasses = selectHelperClasses[errorText ? 'invalid' : state];

  // Check if value is selected
  function isSelected(optionValue: T): boolean {
    if (multiple && Array.isArray(value)) {
      return value.includes(optionValue);
    }
    return value === optionValue;
  }

  // Toggle dropdown
  function toggleDropdown() {
    if (disabled || readonly) return;
    isOpen = !isOpen;
    if (isOpen) {
      dispatch('open');
      highlightedIndex = -1;
      if (searchable) {
        setTimeout(() => searchInputRef?.focus(), 10);
      }
    } else {
      dispatch('close');
      searchQuery = '';
    }
  }

  // Select option
  function selectOption(option: SelectOption<T>) {
    if (option.disabled) return;

    if (multiple) {
      const currentValue = (Array.isArray(value) ? value : []) as T[];
      if (currentValue.includes(option.value)) {
        value = currentValue.filter((v) => v !== option.value) as T[];
      } else {
        value = [...currentValue, option.value] as T[];
      }
    } else {
      value = option.value;
      isOpen = false;
      searchQuery = '';
    }

    dispatch('change', { value: value as T | T[], option });
  }

  // Clear selection
  function handleClear(e: MouseEvent) {
    e.stopPropagation();
    value = multiple ? [] : undefined;
    dispatch('clear');
    dispatch('change', { value: value as T | T[] });
  }

  // Keyboard navigation
  function handleKeydown(event: KeyboardEvent) {
    if (disabled) return;

    if (isKey(event, 'Escape')) {
      isOpen = false;
      searchQuery = '';
      return;
    }

    if (!isOpen) {
      if (isKey(event, 'Enter') || isKey(event, 'Space') || isKey(event, 'ArrowDown')) {
        event.preventDefault();
        toggleDropdown();
      }
      return;
    }

    switch (event.key) {
      case Keys.ArrowDown:
        event.preventDefault();
        highlightedIndex = Math.min(highlightedIndex + 1, filteredOptions.length - 1);
        scrollToHighlighted();
        break;
      case Keys.ArrowUp:
        event.preventDefault();
        highlightedIndex = Math.max(highlightedIndex - 1, 0);
        scrollToHighlighted();
        break;
      case Keys.Enter:
        event.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < filteredOptions.length) {
          selectOption(filteredOptions[highlightedIndex]!);
        }
        break;
      case Keys.Home:
        event.preventDefault();
        highlightedIndex = 0;
        scrollToHighlighted();
        break;
      case Keys.End:
        event.preventDefault();
        highlightedIndex = filteredOptions.length - 1;
        scrollToHighlighted();
        break;
    }
  }

  function scrollToHighlighted() {
    if (listboxRef && highlightedIndex >= 0) {
      const options = listboxRef.querySelectorAll('[role="option"]');
      options[highlightedIndex]?.scrollIntoView({ block: 'nearest' });
    }
  }

  // Click outside handler
  function handleClickOutside(event: MouseEvent) {
    if (containerRef && !containerRef.contains(event.target as Node)) {
      isOpen = false;
      searchQuery = '';
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });

  $: hasValue = multiple
    ? Array.isArray(value) && value.length > 0
    : value !== undefined && value !== null;
</script>

<div class={containerClasses} bind:this={containerRef}>
  {#if label}
    <label for={id} class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-semantic-error-500 ml-0.5" aria-hidden="true">*</span>
      {/if}
    </label>
  {/if}

  <div class="relative">
    <!-- Hidden native select for form submission -->
    <select
      {id}
      {name}
      {required}
      {disabled}
      {multiple}
      class="sr-only"
      tabindex="-1"
      aria-hidden="true"
    >
      {#if multiple && Array.isArray(value)}
        {#each value as v}
          <option value={v} selected>{v}</option>
        {/each}
      {:else if value !== undefined}
        <option value={value} selected>{value}</option>
      {/if}
    </select>

    <!-- Custom select button -->
    <button
      type="button"
      class={selectClasses}
      data-testid={testId || undefined}
      aria-haspopup="listbox"
      aria-expanded={isOpen}
      aria-labelledby={label ? `${id}-label` : undefined}
      aria-describedby={displayedHelperText ? `${id}-helper` : undefined}
      aria-invalid={state === 'invalid' || !!errorText}
      {disabled}
      on:click={toggleDropdown}
      on:keydown={handleKeydown}
    >
      <span class={cn('block truncate text-left', !hasValue && 'text-neutral-400')}>
        {displayLabel}
      </span>

      <span class={chevronClasses}>
        {#if clearable && hasValue && !disabled}
          <button
            type="button"
            class="mr-2 hover:text-neutral-600 pointer-events-auto"
            on:click={handleClear}
            aria-label="Clear selection"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
        <svg
          class="w-5 h-5 transition-transform duration-200"
          class:rotate-180={isOpen}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </span>
    </button>

    <!-- Dropdown panel -->
    {#if isOpen}
      <div class={dropdownPanelClasses} role="listbox" aria-multiselectable={multiple}>
        {#if searchable}
          <input
            bind:this={searchInputRef}
            type="text"
            class={searchInputClasses}
            placeholder="Search..."
            bind:value={searchQuery}
            on:keydown={handleKeydown}
          />
        {/if}

        <ul bind:this={listboxRef} class="py-1">
          {#if groups.length > 0}
            {#each groups as group}
              <li class={groupLabelClasses}>{group.label}</li>
              {#each filterOptions(group.options, searchQuery) as option, i}
                {@const globalIndex = allOptions.indexOf(option)}
                <li
                  role="option"
                  aria-selected={isSelected(option.value)}
                  aria-disabled={option.disabled}
                  class={cn(
                    optionBaseClasses,
                    !option.disabled && optionHoverClasses,
                    isSelected(option.value) && optionSelectedClasses,
                    option.disabled && optionDisabledClasses,
                    highlightedIndex === globalIndex && 'bg-brand-primary-50'
                  )}
                  on:click={() => selectOption(option)}
                  on:mouseenter={() => (highlightedIndex = globalIndex)}
                >
                  <span class="block truncate">{option.label}</span>
                  {#if isSelected(option.value)}
                    <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-brand-primary-600">
                      <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                      </svg>
                    </span>
                  {/if}
                </li>
              {/each}
            {/each}
          {:else}
            {#each filteredOptions as option, i}
              <li
                role="option"
                aria-selected={isSelected(option.value)}
                aria-disabled={option.disabled}
                class={cn(
                  optionBaseClasses,
                  !option.disabled && optionHoverClasses,
                  isSelected(option.value) && optionSelectedClasses,
                  option.disabled && optionDisabledClasses,
                  highlightedIndex === i && 'bg-brand-primary-50'
                )}
                on:click={() => selectOption(option)}
                on:mouseenter={() => (highlightedIndex = i)}
              >
                <span class="block truncate">{option.label}</span>
                {#if isSelected(option.value)}
                  <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-brand-primary-600">
                    <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </span>
                {/if}
              </li>
            {/each}
          {/if}

          {#if filteredOptions.length === 0}
            <li class="py-2 px-3 text-neutral-500 text-sm">No options found</li>
          {/if}
        </ul>
      </div>
    {/if}
  </div>

  {#if displayedHelperText}
    <p id="{id}-helper" class={helperClasses}>
      {displayedHelperText}
    </p>
  {/if}
</div>
