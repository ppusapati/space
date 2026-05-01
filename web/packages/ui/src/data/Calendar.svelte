<script context="module" lang="ts">
  export interface CalendarEvent {
    id: string;
    title: string;
    start: Date | string;
    end?: Date | string;
    color?: string;
    allDay?: boolean;
    metadata?: Record<string, unknown>;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let date: Date = new Date();
  export let events: CalendarEvent[] = [];
  export let view: 'month' | 'week' | 'day' = 'month';
  export let showHeader: boolean = true;
  export let showWeekNumbers: boolean = false;
  export let firstDayOfWeek: 0 | 1 = 0; // 0 = Sunday, 1 = Monday
  export let minDate: Date | null = null;
  export let maxDate: Date | null = null;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    dateSelect: { date: Date };
    eventClick: { event: CalendarEvent };
    viewChange: { view: 'month' | 'week' | 'day' };
    navigate: { date: Date; direction: 'prev' | 'next' | 'today' };
  }>();

  const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  $: orderedWeekdays = firstDayOfWeek === 1
    ? [...WEEKDAYS.slice(1), WEEKDAYS[0]]
    : WEEKDAYS;

  $: currentMonth = date.getMonth();
  $: currentYear = date.getFullYear();

  $: calendarDays = generateCalendarDays(currentYear, currentMonth);

  function generateCalendarDays(year: number, month: number): Date[][] {
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);

    // Adjust for first day of week
    let startDay = firstDay.getDay() - firstDayOfWeek;
    if (startDay < 0) startDay += 7;

    const daysInMonth = lastDay.getDate();
    const weeks: Date[][] = [];
    let currentWeek: Date[] = [];

    // Add days from previous month
    const prevMonth = new Date(year, month, 0);
    const daysInPrevMonth = prevMonth.getDate();
    for (let i = startDay - 1; i >= 0; i--) {
      currentWeek.push(new Date(year, month - 1, daysInPrevMonth - i));
    }

    // Add days of current month
    for (let day = 1; day <= daysInMonth; day++) {
      currentWeek.push(new Date(year, month, day));
      if (currentWeek.length === 7) {
        weeks.push(currentWeek);
        currentWeek = [];
      }
    }

    // Add days from next month
    if (currentWeek.length > 0) {
      let nextDay = 1;
      while (currentWeek.length < 7) {
        currentWeek.push(new Date(year, month + 1, nextDay++));
      }
      weeks.push(currentWeek);
    }

    return weeks;
  }

  function getEventsForDate(day: Date): CalendarEvent[] {
    return events.filter(event => {
      const eventStart = new Date(event.start);
      const eventEnd = event.end ? new Date(event.end) : eventStart;

      const dayStart = new Date(day.getFullYear(), day.getMonth(), day.getDate());
      const dayEnd = new Date(day.getFullYear(), day.getMonth(), day.getDate(), 23, 59, 59);

      return eventStart <= dayEnd && eventEnd >= dayStart;
    });
  }

  function isToday(day: Date): boolean {
    const today = new Date();
    return day.getDate() === today.getDate() &&
           day.getMonth() === today.getMonth() &&
           day.getFullYear() === today.getFullYear();
  }

  function isCurrentMonth(day: Date): boolean {
    return day.getMonth() === currentMonth;
  }

  function isSelected(day: Date): boolean {
    return day.getDate() === date.getDate() &&
           day.getMonth() === date.getMonth() &&
           day.getFullYear() === date.getFullYear();
  }

  function isDisabled(day: Date): boolean {
    if (minDate && day < minDate) return true;
    if (maxDate && day > maxDate) return true;
    return false;
  }

  function getWeekNumber(day: Date): number {
    const firstDayOfYear = new Date(day.getFullYear(), 0, 1);
    const pastDaysOfYear = (day.getTime() - firstDayOfYear.getTime()) / 86400000;
    return Math.ceil((pastDaysOfYear + firstDayOfYear.getDay() + 1) / 7);
  }

  function navigatePrev() {
    if (view === 'month') {
      date = new Date(currentYear, currentMonth - 1, 1);
    } else if (view === 'week') {
      date = new Date(date.getTime() - 7 * 24 * 60 * 60 * 1000);
    } else {
      date = new Date(date.getTime() - 24 * 60 * 60 * 1000);
    }
    dispatch('navigate', { date, direction: 'prev' });
  }

  function navigateNext() {
    if (view === 'month') {
      date = new Date(currentYear, currentMonth + 1, 1);
    } else if (view === 'week') {
      date = new Date(date.getTime() + 7 * 24 * 60 * 60 * 1000);
    } else {
      date = new Date(date.getTime() + 24 * 60 * 60 * 1000);
    }
    dispatch('navigate', { date, direction: 'next' });
  }

  function goToToday() {
    date = new Date();
    dispatch('navigate', { date, direction: 'today' });
  }

  function selectDate(day: Date) {
    if (isDisabled(day)) return;
    date = day;
    dispatch('dateSelect', { date: day });
  }

  function handleEventClick(event: CalendarEvent, e: MouseEvent) {
    e.stopPropagation();
    dispatch('eventClick', { event });
  }

  function setView(newView: 'month' | 'week' | 'day') {
    view = newView;
    dispatch('viewChange', { view: newView });
  }

  function formatMonthYear(d: Date): string {
    return d.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  }
</script>

<div class={cn('calendar rounded-lg border border-[var(--color-border-primary)] bg-[var(--color-surface-primary)]', className)}>
  {#if showHeader}
    <div class="flex items-center justify-between p-4 border-b border-[var(--color-border-primary)]">
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="p-2 rounded hover:bg-[var(--color-surface-secondary)] text-[var(--color-text-secondary)]"
          on:click={navigatePrev}
          aria-label="Previous"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <h2 class="text-lg font-semibold text-[var(--color-text-primary)] min-w-[180px] text-center">
          {formatMonthYear(date)}
        </h2>

        <button
          type="button"
          class="p-2 rounded hover:bg-[var(--color-surface-secondary)] text-[var(--color-text-secondary)]"
          on:click={navigateNext}
          aria-label="Next"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <button
          type="button"
          class="ml-2 px-3 py-1.5 text-sm rounded border border-[var(--color-border-primary)] hover:bg-[var(--color-surface-secondary)]"
          on:click={goToToday}
        >
          Today
        </button>
      </div>

      <div class="flex items-center gap-1 bg-[var(--color-surface-secondary)] rounded-lg p-1">
        {#each ['month', 'week', 'day'] as v}
          <button
            type="button"
            class={cn(
              'px-3 py-1.5 text-sm rounded capitalize',
              view === v
                ? 'bg-[var(--color-interactive-primary)] text-white'
                : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
            )}
            on:click={() => setView(v as 'month' | 'week' | 'day')}
          >
            {v}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Month View -->
  {#if view === 'month'}
    <div class="p-4">
      <!-- Weekday Headers -->
      <div class="grid grid-cols-7 gap-1 mb-2">
        {#if showWeekNumbers}
          <div class="text-xs font-medium text-[var(--color-text-tertiary)] text-center p-2">
            Wk
          </div>
        {/if}
        {#each orderedWeekdays as weekday}
          <div class="text-xs font-medium text-[var(--color-text-tertiary)] text-center p-2">
            {weekday}
          </div>
        {/each}
      </div>

      <!-- Calendar Grid -->
      {#each calendarDays as week, weekIndex}
        <div class="grid grid-cols-7 gap-1">
          {#if showWeekNumbers}
            <div class="text-xs text-[var(--color-text-tertiary)] text-center p-2">
              {getWeekNumber(week[0]!)}
            </div>
          {/if}
          {#each week as day}
            {@const dayEvents = getEventsForDate(day)}
            <button
              type="button"
              class={cn(
                'min-h-[80px] p-1 rounded text-left transition-colors',
                'hover:bg-[var(--color-surface-secondary)]',
                'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]',
                !isCurrentMonth(day) && 'opacity-40',
                isToday(day) && 'bg-[var(--color-interactive-primary)]/10',
                isSelected(day) && 'ring-2 ring-[var(--color-interactive-primary)]',
                isDisabled(day) && 'opacity-30 cursor-not-allowed'
              )}
              on:click={() => selectDate(day)}
              disabled={isDisabled(day)}
            >
              <span
                class={cn(
                  'inline-flex items-center justify-center w-7 h-7 text-sm rounded-full',
                  isToday(day) && 'bg-[var(--color-interactive-primary)] text-white font-semibold'
                )}
              >
                {day.getDate()}
              </span>

              <!-- Events -->
              {#if dayEvents.length > 0}
                <div class="mt-1 space-y-0.5">
                  {#each dayEvents.slice(0, 3) as event}
                    <button
                      type="button"
                      class="w-full text-left text-xs px-1 py-0.5 rounded truncate"
                      style="background-color: {event.color || 'var(--color-interactive-primary)'}; color: white;"
                      on:click={(e) => handleEventClick(event, e)}
                    >
                      {event.title}
                    </button>
                  {/each}
                  {#if dayEvents.length > 3}
                    <span class="text-xs text-[var(--color-text-tertiary)]">
                      +{dayEvents.length - 3} more
                    </span>
                  {/if}
                </div>
              {/if}
            </button>
          {/each}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Week/Day Views (simplified) -->
  {#if view === 'week' || view === 'day'}
    <div class="p-4 text-center text-[var(--color-text-secondary)]">
      <p>Week and Day views require additional implementation.</p>
      <p class="text-sm mt-2">Currently showing: {view} view for {date.toLocaleDateString()}</p>
    </div>
  {/if}
</div>
