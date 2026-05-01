<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { Keys, isKey } from '../utils/keyboard';
  import {
    type DatePickerMode,
    type DatePickerView,
    type DateRange,
    type CalendarDay,
    datepickerSizeClasses,
    datepickerInputClasses,
    calendarClasses,
    getMonthName,
    getDayNames,
    getCalendarDays,
    formatDate,
  } from './datepicker.types';
  import type { Size, ValidationState } from '../types';

  // Props
  export let value: Date | Date[] | DateRange | null = null;
  export let mode: DatePickerMode = 'single';
  export let placeholder: string = 'Select date';
  export let format: string = 'MM/DD/YYYY';
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let minDate: Date | undefined = undefined;
  export let maxDate: Date | undefined = undefined;
  export let disabledDates: Date[] = [];
  export let firstDayOfWeek: 0 | 1 = 0;
  export let showWeekNumbers: boolean = false;
  export let clearable: boolean = false;
  export let disabled: boolean = false;
  export let required: boolean = false;
  export let name: string = '';
  export let id: string = uid('datepicker');
  export let testId: string = '';
  export let fullWidth: boolean = true;
  export let locale: string = 'en-US';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { value: Date | Date[] | DateRange | null };
    open: void;
    close: void;
    clear: void;
  }>();

  // Internal state
  let isOpen = false;
  let currentView: DatePickerView = 'days';
  let viewDate = new Date();
  let containerRef: HTMLDivElement;

  // For range selection
  let rangeStartDate: Date | null = null;

  // Computed
  $: currentMonth = viewDate.getMonth();
  $: currentYear = viewDate.getFullYear();
  $: dayNames = getDayNames(firstDayOfWeek);
  $: calendarDays = getCalendarDays(
    currentYear,
    currentMonth,
    firstDayOfWeek,
    mode === 'range' ? (value as DateRange) || { start: rangeStartDate, end: null } : value,
    minDate,
    maxDate,
    disabledDates
  );

  $: displayValue = formatDate(value, format, locale);

  $: inputClasses = cn(
    datepickerInputClasses,
    datepickerSizeClasses[size],
    className
  );

  $: containerClasses = cn('relative', fullWidth ? 'w-full' : 'inline-block');

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;

  // Navigation
  function prevMonth() {
    viewDate = new Date(currentYear, currentMonth - 1, 1);
  }

  function nextMonth() {
    viewDate = new Date(currentYear, currentMonth + 1, 1);
  }

  function prevYear() {
    viewDate = new Date(currentYear - 1, currentMonth, 1);
  }

  function nextYear() {
    viewDate = new Date(currentYear + 1, currentMonth, 1);
  }

  function goToToday() {
    viewDate = new Date();
    currentView = 'days';
  }

  // View switching
  function switchToMonths() {
    currentView = 'months';
  }

  function switchToYears() {
    currentView = 'years';
  }

  function selectMonth(month: number) {
    viewDate = new Date(currentYear, month, 1);
    currentView = 'days';
  }

  function selectYear(year: number) {
    viewDate = new Date(year, currentMonth, 1);
    currentView = 'months';
  }

  // Date selection
  function selectDate(day: CalendarDay) {
    if (day.isDisabled) return;

    const selectedDate = day.date;

    if (mode === 'single') {
      value = selectedDate;
      isOpen = false;
      dispatch('change', { value });
    } else if (mode === 'multiple') {
      const currentDates = (value as Date[]) || [];
      const existingIndex = currentDates.findIndex(
        d => d.toDateString() === selectedDate.toDateString()
      );

      if (existingIndex >= 0) {
        value = [...currentDates.slice(0, existingIndex), ...currentDates.slice(existingIndex + 1)];
      } else {
        value = [...currentDates, selectedDate];
      }
      dispatch('change', { value });
    } else if (mode === 'range') {
      if (!rangeStartDate) {
        rangeStartDate = selectedDate;
        value = { start: selectedDate, end: null };
      } else {
        const start = selectedDate < rangeStartDate ? selectedDate : rangeStartDate;
        const end = selectedDate < rangeStartDate ? rangeStartDate : selectedDate;
        value = { start, end };
        rangeStartDate = null;
        isOpen = false;
        dispatch('change', { value });
      }
    }
  }

  // Toggle calendar
  function toggleCalendar() {
    if (disabled) return;
    isOpen = !isOpen;
    if (isOpen) {
      dispatch('open');
      // Reset view to current value's month/year
      if (value instanceof Date) {
        viewDate = new Date(value);
      } else if (mode === 'range' && value && 'start' in value && value.start) {
        viewDate = new Date(value.start);
      }
      currentView = 'days';
    } else {
      dispatch('close');
    }
  }

  // Clear value
  function handleClear(e: MouseEvent) {
    e.stopPropagation();
    if (mode === 'range') {
      value = { start: null, end: null };
      rangeStartDate = null;
    } else if (mode === 'multiple') {
      value = [];
    } else {
      value = null;
    }
    dispatch('clear');
    dispatch('change', { value });
  }

  // Keyboard navigation
  function handleKeydown(event: KeyboardEvent) {
    if (disabled) return;

    if (isKey(event, 'Escape')) {
      isOpen = false;
      return;
    }

    if (!isOpen && (isKey(event, 'Enter') || isKey(event, 'Space') || isKey(event, 'ArrowDown'))) {
      event.preventDefault();
      toggleCalendar();
    }
  }

  // Click outside handler
  function handleClickOutside(event: MouseEvent) {
    if (containerRef && !containerRef.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });

  // Generate years for year picker
  $: yearRange = Array.from({ length: 12 }, (_, i) => currentYear - 5 + i);

  $: hasValue = mode === 'range'
    ? value && 'start' in value && (value.start || value.end)
    : mode === 'multiple'
    ? Array.isArray(value) && value.length > 0
    : value !== null;
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

  <!-- Hidden input for form submission -->
  <input
    type="hidden"
    {id}
    {name}
    value={displayValue}
    {required}
  />

  <!-- Custom trigger -->
  <button
    type="button"
    class={inputClasses}
    data-testid={testId || undefined}
    aria-haspopup="dialog"
    aria-expanded={isOpen}
    aria-invalid={state === 'invalid' || !!errorText}
    {disabled}
    on:click={toggleCalendar}
    on:keydown={handleKeydown}
  >
    <span class="flex items-center justify-between">
      <span class={cn('block truncate text-left', !hasValue && 'text-neutral-400')}>
        {displayValue || placeholder}
      </span>
      <span class="flex items-center gap-2">
        {#if clearable && hasValue && !disabled}
          <button
            type="button"
            class="hover:text-neutral-600 text-neutral-400"
            on:click={handleClear}
            aria-label="Clear date"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
        <svg class="w-5 h-5 text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </span>
    </span>
  </button>

  <!-- Calendar dropdown -->
  {#if isOpen}
    <div class={calendarClasses.container} role="dialog" aria-modal="true" aria-label="Date picker">
      <!-- Header -->
      <div class={calendarClasses.header}>
        <button
          type="button"
          class={calendarClasses.headerButton}
          on:click={currentView === 'years' ? () => viewDate = new Date(currentYear - 12, currentMonth, 1) : prevMonth}
          aria-label="Previous"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <div class="flex items-center gap-1">
          {#if currentView === 'days'}
            <button
              type="button"
              class={calendarClasses.headerTitle}
              on:click={switchToMonths}
            >
              {getMonthName(currentMonth)}
            </button>
            <button
              type="button"
              class={calendarClasses.headerTitle}
              on:click={switchToYears}
            >
              {currentYear}
            </button>
          {:else if currentView === 'months'}
            <button
              type="button"
              class={calendarClasses.headerTitle}
              on:click={switchToYears}
            >
              {currentYear}
            </button>
          {:else}
            <span class="text-sm font-semibold text-neutral-900">
              {yearRange[0]} - {yearRange[yearRange.length - 1]}
            </span>
          {/if}
        </div>

        <button
          type="button"
          class={calendarClasses.headerButton}
          on:click={currentView === 'years' ? () => viewDate = new Date(currentYear + 12, currentMonth, 1) : nextMonth}
          aria-label="Next"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <!-- Days view -->
      {#if currentView === 'days'}
        <div class={calendarClasses.weekdays}>
          {#each dayNames as day}
            <div class={calendarClasses.weekday}>{day}</div>
          {/each}
        </div>

        <div class={calendarClasses.days}>
          {#each calendarDays as day}
            <button
              type="button"
              class={cn(
                calendarClasses.day.base,
                day.isCurrentMonth ? calendarClasses.day.currentMonth : calendarClasses.day.otherMonth,
                day.isToday && calendarClasses.day.today,
                day.isSelected && calendarClasses.day.selected,
                day.isInRange && calendarClasses.day.inRange,
                day.isRangeStart && calendarClasses.day.rangeStart,
                day.isRangeEnd && calendarClasses.day.rangeEnd,
                day.isDisabled && calendarClasses.day.disabled
              )}
              disabled={day.isDisabled}
              on:click={() => selectDate(day)}
              aria-label={day.date.toDateString()}
              aria-selected={day.isSelected}
            >
              {day.day}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Months view -->
      {#if currentView === 'months'}
        <div class={calendarClasses.months}>
          {#each Array(12) as _, i}
            <button
              type="button"
              class={cn(
                calendarClasses.month,
                currentMonth === i && calendarClasses.monthSelected
              )}
              on:click={() => selectMonth(i)}
            >
              {getMonthName(i).slice(0, 3)}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Years view -->
      {#if currentView === 'years'}
        <div class={calendarClasses.years}>
          {#each yearRange as year}
            <button
              type="button"
              class={cn(
                calendarClasses.year,
                currentYear === year && calendarClasses.yearSelected
              )}
              on:click={() => selectYear(year)}
            >
              {year}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Footer -->
      <div class="mt-4 pt-3 border-t border-neutral-200 flex justify-between">
        <button
          type="button"
          class="text-sm text-brand-primary-600 hover:text-brand-primary-700"
          on:click={goToToday}
        >
          Today
        </button>
        {#if mode === 'range' && rangeStartDate}
          <span class="text-sm text-neutral-500">
            Select end date
          </span>
        {/if}
      </div>
    </div>
  {/if}

  {#if displayedHelperText}
    <p
      id="{id}-helper"
      class={cn(
        'mt-1 text-sm',
        errorText ? 'text-semantic-error-600' : 'text-neutral-500'
      )}
    >
      {displayedHelperText}
    </p>
  {/if}
</div>
