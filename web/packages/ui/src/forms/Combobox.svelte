<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { clickOutside } from '../actions/clickOutside';
  import { cn } from '../utils';
  import type { ComboboxProps, ComboboxOption, HighlightSegment } from './combobox.types';
  import {
    comboboxClasses,
    comboboxOptionClasses,
    comboboxSizeClasses,
    comboboxStateClasses,
    comboboxHelperClasses,
    filterComboboxOptions,
    highlightMatches as highlightMatchesFn,
    getComboboxDisplayValue,
    debounce,
  } from './combobox.types';

  type T = $$Generic;
  type $$Props = ComboboxProps<T>;

  export let value: $$Props['value'] = undefined;
  export let options: $$Props['options'] = [];
  export let groups: $$Props['groups'] = undefined;
  export let loadOptions: $$Props['loadOptions'] = undefined;
  export let debounceMs: $$Props['debounceMs'] = 300;
  export let minChars: $$Props['minChars'] = 0;
  export let placeholder: $$Props['placeholder'] = 'Search...';
  export let size: $$Props['size'] = 'md';
  export let state: $$Props['state'] = 'default';
  export let label: $$Props['label'] = undefined;
  export let helperText: $$Props['helperText'] = undefined;
  export let errorText: $$Props['errorText'] = undefined;
  export let disabled: $$Props['disabled'] = false;
  export let readonly: $$Props['readonly'] = false;
  export let multiple: $$Props['multiple'] = false;
  export let clearable: $$Props['clearable'] = true;
  export let creatable: $$Props['creatable'] = false;
  export let createText: $$Props['createText'] = 'Create';
  export let fullWidth: $$Props['fullWidth'] = true;
  export let highlightMatches: $$Props['highlightMatches'] = true;
  export let matchType: $$Props['matchType'] = 'contains';
  export let searchFields: $$Props['searchFields'] = undefined;
  export let noResultsText: $$Props['noResultsText'] = 'No results found';
  export let loadingText: $$Props['loadingText'] = 'Loading...';
  export let maxVisibleItems: $$Props['maxVisibleItems'] = 8;
  export let cacheResults: $$Props['cacheResults'] = true;
  export let name: $$Props['name'] = undefined;
  export let id: $$Props['id'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: T | T[] | undefined;
    search: string;
    create: string;
    open: void;
    close: void;
    clear: void;
  }>();

  let inputElement: HTMLInputElement;
  let isOpen = false;
  let query = '';
  let activeIndex = -1;
  let isLoading = false;
  let asyncOptions: ComboboxOption<T>[] = [];
  let abortController: AbortController | null = null;
  let cache = new Map<string, ComboboxOption<T>[]>();

  // Flatten grouped options
  $: allOptions = groups
    ? groups.flatMap((g) => g.options)
    : options || [];

  // Merge static and async options
  $: displayOptions = loadOptions
    ? asyncOptions
    : filterComboboxOptions(allOptions, query, matchType, searchFields);

  // Check if we can create new option
  $: canCreate = creatable && query.length > 0 && !displayOptions.some(
    (opt) => opt.label.toLowerCase() === query.toLowerCase()
  );

  // Display value in input
  $: displayValue = isOpen ? query : getComboboxDisplayValue(value, allOptions, '');

  // Has value to clear
  $: hasValue = multiple
    ? Array.isArray(value) && value.length > 0
    : value !== undefined && value !== null;

  // Debounced async search
  const debouncedSearch = debounce(async (searchQuery: unknown) => {
    const searchQueryStr = searchQuery as string;
    if (!loadOptions) return;

    // Check cache
    if (cacheResults && cache.has(searchQueryStr)) {
      asyncOptions = cache.get(searchQueryStr) || [];
      isLoading = false;
      return;
    }

    // Cancel previous request
    if (abortController) {
      abortController.abort();
    }

    abortController = new AbortController();
    isLoading = true;

    try {
      const results = await loadOptions(searchQueryStr, abortController.signal);
      asyncOptions = results;
      if (cacheResults) {
        cache.set(searchQueryStr, results);
      }
    } catch (error) {
      if ((error as Error).name !== 'AbortError') {
        console.error('Combobox load error:', error);
        asyncOptions = [];
      }
    } finally {
      isLoading = false;
    }
  }, debounceMs ?? 300);

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    query = target.value;
    activeIndex = -1;

    dispatch('search', query);

    if (!isOpen) {
      openDropdown();
    }

    if (loadOptions && query.length >= (minChars ?? 0)) {
      isLoading = true;
      debouncedSearch(query);
    }
  }

  function handleFocus() {
    if (!disabled && !readonly) {
      openDropdown();
    }
  }

  function openDropdown() {
    if (!isOpen) {
      isOpen = true;
      dispatch('open');
      if (loadOptions && query.length >= (minChars ?? 0)) {
        isLoading = true;
        debouncedSearch(query);
      }
    }
  }

  function closeDropdown() {
    if (isOpen) {
      isOpen = false;
      activeIndex = -1;
      query = '';
      dispatch('close');
    }
  }

  function selectOption(option: ComboboxOption<T>) {
    if (option.disabled) return;

    if (multiple) {
      const currentValues = Array.isArray(value) ? value : [];
      const isSelected = currentValues.includes(option.value);

      if (isSelected) {
        value = currentValues.filter((v) => v !== option.value) as T[];
      } else {
        value = [...currentValues, option.value] as T[];
      }
    } else {
      value = option.value;
      closeDropdown();
    }

    dispatch('change', value);
  }

  function handleCreate() {
    if (!canCreate) return;

    dispatch('create', query);

    if (!multiple) {
      closeDropdown();
    }
  }

  function handleClear(e: Event) {
    e.stopPropagation();
    value = multiple ? [] : undefined;
    query = '';
    dispatch('clear');
    dispatch('change', value);
    inputElement?.focus();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (disabled || readonly) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        if (!isOpen) {
          openDropdown();
        } else {
          activeIndex = Math.min(activeIndex + 1, displayOptions.length - 1);
        }
        break;

      case 'ArrowUp':
        e.preventDefault();
        if (isOpen) {
          activeIndex = Math.max(activeIndex - 1, 0);
        }
        break;

      case 'Enter':
        e.preventDefault();
        if (isOpen && activeIndex >= 0 && displayOptions[activeIndex]) {
          selectOption(displayOptions[activeIndex]!);
        } else if (canCreate) {
          handleCreate();
        }
        break;

      case 'Escape':
        e.preventDefault();
        closeDropdown();
        break;

      case 'Tab':
        closeDropdown();
        break;
    }
  }

  function isSelected(option: ComboboxOption<T>): boolean {
    if (multiple && Array.isArray(value)) {
      return value.includes(option.value);
    }
    return value === option.value;
  }

  function getHighlightedLabel(label: string): HighlightSegment[] {
    if (!highlightMatches || !query) {
      return [{ text: label, isMatch: false }];
    }
    return highlightMatchesFn(label, query);
  }

  onDestroy(() => {
    if (abortController) {
      abortController.abort();
    }
  });

  $: helperMessage = state === 'invalid' && errorText ? errorText : helperText;
  $: helperClass = comboboxHelperClasses[state ?? 'default'];
</script>

<div
  class={cn(comboboxClasses.container, fullWidth && 'w-full', className)}
  use:clickOutside={closeDropdown}
>
  {#if label}
    <label
      for={id}
      class="block text-sm font-medium text-neutral-700 mb-1"
    >
      {label}
    </label>
  {/if}

  <div class={comboboxClasses.inputWrapper}>
    <input
      bind:this={inputElement}
      type="text"
      {id}
      {name}
      {disabled}
      {readonly}
      {placeholder}
      value={displayValue}
      autocomplete="off"
      role="combobox"
      aria-expanded={isOpen}
      aria-haspopup="listbox"
      aria-autocomplete="list"
      aria-controls="combobox-listbox"
      aria-activedescendant={activeIndex >= 0 ? `option-${activeIndex}` : undefined}
      class={cn(
        comboboxClasses.input,
        comboboxSizeClasses[size ?? 'md'],
        comboboxStateClasses[state ?? 'default'],
        clearable && hasValue && 'pr-16'
      )}
      on:input={handleInput}
      on:focus={handleFocus}
      on:keydown={handleKeydown}
    />

    {#if clearable && hasValue && !disabled && !readonly}
      <button
        type="button"
        class={comboboxClasses.clearButton}
        on:click={handleClear}
        tabindex="-1"
        aria-label="Clear selection"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}

    <span class={cn(comboboxClasses.chevron, isOpen && comboboxClasses.chevronOpen)}>
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </span>
  </div>

  {#if isOpen}
    <div
      class={comboboxClasses.dropdown}
      style="max-height: {maxVisibleItems ? maxVisibleItems * 40 : 240}px;"
      role="listbox"
      id="combobox-listbox"
      aria-label="Options"
    >
      {#if isLoading}
        <div class={comboboxClasses.loadingWrapper}>
          <svg class="w-5 h-5 animate-spin text-brand-primary-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <span class="ml-2 text-sm text-neutral-500">{loadingText}</span>
        </div>
      {:else if displayOptions.length === 0 && !canCreate}
        <div class={comboboxClasses.noResults}>{noResultsText}</div>
      {:else}
        <ul class={comboboxClasses.optionList}>
          {#if groups && !loadOptions}
            {#each groups as group}
              <li class={comboboxClasses.groupLabel}>{group.label}</li>
              {#each filterComboboxOptions(group.options, query, matchType, searchFields) as option, index}
                {@const isActive = activeIndex === allOptions.indexOf(option)}
                <li
                  id="option-{allOptions.indexOf(option)}"
                  role="option"
                  aria-selected={isSelected(option)}
                  class={cn(
                    comboboxOptionClasses.base,
                    comboboxOptionClasses.hover,
                    isActive && comboboxOptionClasses.active,
                    isSelected(option) && comboboxOptionClasses.selected,
                    option.disabled && comboboxOptionClasses.disabled
                  )}
                  on:click={() => selectOption(option)}
                  on:mouseenter={() => activeIndex = allOptions.indexOf(option)}
                >
                  {#if option.icon}
                    <img src={option.icon} alt="" class={comboboxOptionClasses.icon} />
                  {/if}
                  <div class={comboboxOptionClasses.content}>
                    <span class={comboboxOptionClasses.label}>
                      {#if highlightMatches && query}
                        {#each getHighlightedLabel(option.label) as segment}
                          {#if segment.isMatch}
                            <mark class={comboboxOptionClasses.highlight}>{segment.text}</mark>
                          {:else}
                            {segment.text}
                          {/if}
                        {/each}
                      {:else}
                        {option.label}
                      {/if}
                    </span>
                    {#if option.description}
                      <span class={comboboxOptionClasses.description}>{option.description}</span>
                    {/if}
                  </div>
                  {#if isSelected(option)}
                    <svg class={comboboxOptionClasses.checkmark} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  {/if}
                </li>
              {/each}
            {/each}
          {:else}
            {#each displayOptions as option, index}
              {@const isActive = activeIndex === index}
              <li
                id="option-{index}"
                role="option"
                aria-selected={isSelected(option)}
                class={cn(
                  comboboxOptionClasses.base,
                  comboboxOptionClasses.hover,
                  isActive && comboboxOptionClasses.active,
                  isSelected(option) && comboboxOptionClasses.selected,
                  option.disabled && comboboxOptionClasses.disabled
                )}
                on:click={() => selectOption(option)}
                on:mouseenter={() => activeIndex = index}
              >
                {#if option.icon}
                  <img src={option.icon} alt="" class={comboboxOptionClasses.icon} />
                {/if}
                <div class={comboboxOptionClasses.content}>
                  <span class={comboboxOptionClasses.label}>
                    {#if highlightMatches && query}
                      {#each getHighlightedLabel(option.label) as segment}
                        {#if segment.isMatch}
                          <mark class={comboboxOptionClasses.highlight}>{segment.text}</mark>
                        {:else}
                          {segment.text}
                        {/if}
                      {/each}
                    {:else}
                      {option.label}
                    {/if}
                  </span>
                  {#if option.description}
                    <span class={comboboxOptionClasses.description}>{option.description}</span>
                  {/if}
                </div>
                {#if isSelected(option)}
                  <svg class={comboboxOptionClasses.checkmark} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                {/if}
              </li>
            {/each}
          {/if}

          {#if canCreate}
            <li
              class={comboboxClasses.createOption}
              on:click={handleCreate}
              role="option"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              {createText} "{query}"
            </li>
          {/if}
        </ul>
      {/if}
    </div>
  {/if}

  {#if helperMessage}
    <p class={helperClass}>{helperMessage}</p>
  {/if}
</div>
