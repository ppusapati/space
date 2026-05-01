<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size, ValidationState } from '../types';

  interface CronInputProps {
    value?: string;
    label?: string;
    helperText?: string;
    errorText?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: Size;
    state?: ValidationState;
    name?: string;
    id?: string;
  }

  export let value: string = '0 0 * * *';
  export let label: string = '';
  export let helperText: string = 'Format: minute hour day month day-of-week';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('cron');

  let className: string = '';
  export { className as class };

  let minute = '0';
  let hour = '0';
  let dayOfMonth = '*';
  let month = '*';
  let dayOfWeek = '*';
  let cronText = value;
  let isValid = true;
  let preset = 'custom';

  const dispatch = createEventDispatcher<{
    change: string;
    blur: void;
    focus: void;
  }>();

  const presets = [
    { value: '0 0 * * *', label: 'Every day at midnight' },
    { value: '0 0 * * 0', label: 'Every Sunday at midnight' },
    { value: '0 9 * * 1-5', label: 'Weekdays at 9 AM' },
    { value: '*/5 * * * *', label: 'Every 5 minutes' },
    { value: '0 * * * *', label: 'Every hour' },
    { value: '0 0 1 * *', label: 'First day of month' },
    { value: '0 0 * * *', label: 'Daily' },
    { value: 'custom', label: 'Custom' },
  ];

  const stateClasses = {
    default: 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500',
    success: 'border-green-500 focus:border-green-600 focus:ring-green-500',
    error: 'border-red-500 focus:border-red-600 focus:ring-red-500',
    warning: 'border-yellow-500 focus:border-yellow-600 focus:ring-yellow-500',
  };

  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-4 py-3 text-lg',
  };

  function validateCron(cron: string): boolean {
    const parts = cron.trim().split(/\s+/);
    if (parts.length !== 5) return false;

    const validRanges = [
      { min: 0, max: 59 }, // minute
      { min: 0, max: 23 }, // hour
      { min: 1, max: 31 }, // day
      { min: 1, max: 12 }, // month
      { min: 0, max: 6 }, // day of week
    ];

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]!;
      const range = validRanges[i]!;
      if (part === '*') continue;
      if (part === '?') continue;

      // Check ranges
      if (part.includes('-')) {
        const rangeParts = part.split('-').map(Number);
        const start = rangeParts[0]!;
        const end = rangeParts[1]!;
        if (isNaN(start) || isNaN(end)) return false;
        if (start < range.min || end > range.max) return false;
      } else if (part.includes(',')) {
        const values = part.split(',').map(Number);
        if (values.some((v) => isNaN(v) || v < range.min || v > range.max)) {
          return false;
        }
      } else if (part.includes('/')) {
        // Ignore step validation for simplicity
        continue;
      } else {
        const num = Number(part);
        if (isNaN(num) || num < range.min || num > range.max) {
          return false;
        }
      }
    }

    return true;
  }

  function updateCronFromFields() {
    cronText = `${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`;
    isValid = validateCron(cronText);
    value = cronText;
    dispatch('change', value);
  }

  function handlePresetChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    preset = target.value;

    if (preset !== 'custom') {
      value = preset;
      cronText = preset;
      isValid = true;
      dispatch('change', value);
    }
  }

  function handleCronTextChange(e: Event) {
    const target = e.target as HTMLInputElement;
    cronText = target.value;
    isValid = validateCron(cronText);

    if (isValid) {
      const parts = cronText.split(/\s+/);
      minute = parts[0] ?? minute;
      hour = parts[1] ?? hour;
      dayOfMonth = parts[2] ?? dayOfMonth;
      month = parts[3] ?? month;
      dayOfWeek = parts[4] ?? dayOfWeek;
      value = cronText;
      dispatch('change', value);
    }
  }

  function handleMinuteChange(e: Event) {
    minute = (e.target as HTMLInputElement).value;
    updateCronFromFields();
  }

  function handleHourChange(e: Event) {
    hour = (e.target as HTMLInputElement).value;
    updateCronFromFields();
  }

  function handleDayChange(e: Event) {
    dayOfMonth = (e.target as HTMLInputElement).value;
    updateCronFromFields();
  }

  function handleMonthChange(e: Event) {
    month = (e.target as HTMLInputElement).value;
    updateCronFromFields();
  }

  function handleDayOfWeekChange(e: Event) {
    dayOfWeek = (e.target as HTMLInputElement).value;
    updateCronFromFields();
  }

  function getReadableSchedule(): string {
    if (!isValid) return 'Invalid cron expression';

    const parts = cronText.split(/\s+/);
    const minute = parts[0] ?? '*';
    const hour = parts[1] ?? '*';
    const day = parts[2] ?? '*';
    const month = parts[3] ?? '*';
    const dow = parts[4] ?? '*';

    if (cronText === '0 0 * * *') return 'Every day at midnight';
    if (cronText === '0 0 * * 0') return 'Every Sunday at midnight';
    if (cronText === '0 9 * * 1-5') return 'Weekdays at 9 AM';
    if (cronText === '*/5 * * * *') return 'Every 5 minutes';

    return `Minute: ${minute}, Hour: ${hour}, Day: ${day}, Month: ${month}, Day of Week: ${dow}`;
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-red-500 ml-1">*</span>
      {/if}
    </label>
  {/if}

  <div class="space-y-3">
    <!-- Preset Selection -->
    <div>
      <label class="block text-xs font-medium text-neutral-600 mb-1">Quick Presets</label>
      <select
        value={preset}
        on:change={handlePresetChange}
        disabled={disabled || readonly}
        class="w-full px-3 py-2 border border-neutral-300 rounded-md text-sm disabled:opacity-50"
      >
        {#each presets as p}
          <option value={p.value}>{p.label}</option>
        {/each}
      </select>
    </div>

    <!-- Cron Text Input -->
    <div>
      <label for={id} class="block text-xs font-medium text-neutral-600 mb-1">Cron Expression</label>
      <input
        {id}
        {name}
        {disabled}
        {readonly}
        type="text"
        value={cronText}
        on:change={handleCronTextChange}
        on:input={handleCronTextChange}
        placeholder="0 0 * * *"
        class={cn(
          'w-full px-3 py-2 border rounded-md text-sm font-mono',
          isValid ? 'border-neutral-300' : 'border-red-500',
          disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
        )}
      />
      <p class="mt-1 text-xs text-neutral-500">minute hour day month day-of-week (use * for any, ? for no specific)</p>
    </div>

    <!-- Field Inputs -->
    {#if preset === 'custom'}
      <div class="grid grid-cols-5 gap-2">
        <div>
          <label class="block text-xs font-medium text-neutral-600 mb-1">Minute</label>
          <input
            type="text"
            value={minute}
            on:change={handleMinuteChange}
            placeholder="0"
            disabled={disabled || readonly}
            class="w-full px-2 py-1 border border-neutral-300 rounded text-sm disabled:opacity-50"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-neutral-600 mb-1">Hour</label>
          <input
            type="text"
            value={hour}
            on:change={handleHourChange}
            placeholder="0"
            disabled={disabled || readonly}
            class="w-full px-2 py-1 border border-neutral-300 rounded text-sm disabled:opacity-50"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-neutral-600 mb-1">Day</label>
          <input
            type="text"
            value={dayOfMonth}
            on:change={handleDayChange}
            placeholder="*"
            disabled={disabled || readonly}
            class="w-full px-2 py-1 border border-neutral-300 rounded text-sm disabled:opacity-50"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-neutral-600 mb-1">Month</label>
          <input
            type="text"
            value={month}
            on:change={handleMonthChange}
            placeholder="*"
            disabled={disabled || readonly}
            class="w-full px-2 py-1 border border-neutral-300 rounded text-sm disabled:opacity-50"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-neutral-600 mb-1">Day of Week</label>
          <input
            type="text"
            value={dayOfWeek}
            on:change={handleDayOfWeekChange}
            placeholder="*"
            disabled={disabled || readonly}
            class="w-full px-2 py-1 border border-neutral-300 rounded text-sm disabled:opacity-50"
          />
        </div>
      </div>
    {/if}

    <!-- Schedule Preview -->
    <div
      class={cn(
        'p-2 rounded text-sm',
        isValid ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'
      )}
    >
      {getReadableSchedule()}
    </div>
  </div>

  {#if errorText}
    <p class="mt-2 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-2 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
