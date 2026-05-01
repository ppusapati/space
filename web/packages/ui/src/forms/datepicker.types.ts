/**
 * DatePicker component types and logic
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** DatePicker mode */
export type DatePickerMode = 'single' | 'range' | 'multiple';

/** DatePicker view */
export type DatePickerView = 'days' | 'months' | 'years';

/** Date range value */
export interface DateRange {
  start: Date | null;
  end: Date | null;
}

/** DatePicker props interface */
export interface DatePickerProps extends FormElementProps {
  /** Selected date(s) */
  value?: Date | Date[] | DateRange | null;
  /** Selection mode */
  mode?: DatePickerMode;
  /** Placeholder text */
  placeholder?: string;
  /** Date format for display */
  format?: string;
  /** Size */
  size?: Size;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error message */
  errorText?: string;
  /** Minimum selectable date */
  minDate?: Date;
  /** Maximum selectable date */
  maxDate?: Date;
  /** Disabled specific dates */
  disabledDates?: Date[];
  /** First day of week (0 = Sunday, 1 = Monday) */
  firstDayOfWeek?: 0 | 1;
  /** Show week numbers */
  showWeekNumbers?: boolean;
  /** Clearable */
  clearable?: boolean;
  /** Full width */
  fullWidth?: boolean;
  /** Locale for formatting */
  locale?: string;
}

/** Calendar day info */
export interface CalendarDay {
  date: Date;
  day: number;
  isCurrentMonth: boolean;
  isToday: boolean;
  isSelected: boolean;
  isInRange: boolean;
  isRangeStart: boolean;
  isRangeEnd: boolean;
  isDisabled: boolean;
}

/** UnoCSS class mappings for datepicker sizes */
export const datepickerSizeClasses: Record<Size, string> = {
  xs: 'h-7 px-2 text-xs',
  sm: 'h-8 px-2.5 text-sm',
  md: 'h-10 px-3 text-base',
  lg: 'h-12 px-4 text-lg',
  xl: 'h-14 px-5 text-xl',
};

/** Calendar classes */
export const calendarClasses = {
  container: 'absolute z-dropdown mt-1 bg-neutral-white border border-neutral-200 rounded-lg shadow-lg p-4',
  header: 'flex items-center justify-between mb-4',
  headerButton: 'p-1 rounded hover:bg-neutral-100 text-neutral-600 hover:text-neutral-900 transition-colors',
  headerTitle: 'text-sm font-semibold text-neutral-900 cursor-pointer hover:text-brand-primary-600',
  weekdays: 'grid grid-cols-7 gap-1 mb-2',
  weekday: 'text-xs font-medium text-neutral-500 text-center py-1',
  days: 'grid grid-cols-7 gap-1',
  day: {
    base: 'w-8 h-8 flex items-center justify-center text-sm rounded-full cursor-pointer transition-colors',
    currentMonth: 'text-neutral-900 hover:bg-brand-primary-50',
    otherMonth: 'text-neutral-300',
    today: 'font-bold ring-1 ring-brand-primary-500',
    selected: 'bg-brand-primary-500 text-neutral-white hover:bg-brand-primary-600',
    inRange: 'bg-brand-primary-100 text-brand-primary-900',
    rangeStart: 'rounded-l-full bg-brand-primary-500 text-neutral-white',
    rangeEnd: 'rounded-r-full bg-brand-primary-500 text-neutral-white',
    disabled: 'text-neutral-300 cursor-not-allowed hover:bg-transparent',
  },
  months: 'grid grid-cols-3 gap-2',
  month: 'px-3 py-2 text-sm rounded hover:bg-brand-primary-50 cursor-pointer transition-colors',
  monthSelected: 'bg-brand-primary-500 text-neutral-white hover:bg-brand-primary-600',
  years: 'grid grid-cols-4 gap-2 max-h-48 overflow-auto',
  year: 'px-2 py-1 text-sm rounded hover:bg-brand-primary-50 cursor-pointer transition-colors',
  yearSelected: 'bg-brand-primary-500 text-neutral-white hover:bg-brand-primary-600',
};

/** Input trigger classes */
export const datepickerInputClasses =
  'block w-full border rounded-md transition-all duration-200 cursor-pointer ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-1 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-neutral-100 ' +
  'bg-neutral-white border-neutral-300 focus:ring-brand-primary-500 focus:border-brand-primary-500';

/** Helper functions */

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

export function getMonthName(month: number): string {
  return MONTH_NAMES[month] ?? '';
}

export function getDayNames(firstDayOfWeek: 0 | 1): string[] {
  if (firstDayOfWeek === 1) {
    return [...DAY_NAMES.slice(1), DAY_NAMES[0]!];
  }
  return DAY_NAMES;
}

export function getCalendarDays(
  year: number,
  month: number,
  firstDayOfWeek: 0 | 1,
  selectedDate: Date | Date[] | DateRange | null | undefined,
  minDate?: Date,
  maxDate?: Date,
  disabledDates?: Date[]
): CalendarDay[] {
  const days: CalendarDay[] = [];
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  // Calculate starting day
  let startDay = firstDay.getDay() - firstDayOfWeek;
  if (startDay < 0) startDay += 7;

  // Add days from previous month
  const prevMonthLastDay = new Date(year, month, 0).getDate();
  for (let i = startDay - 1; i >= 0; i--) {
    const date = new Date(year, month - 1, prevMonthLastDay - i);
    days.push(createCalendarDay(date, false, today, selectedDate, minDate, maxDate, disabledDates));
  }

  // Add days from current month
  for (let day = 1; day <= lastDay.getDate(); day++) {
    const date = new Date(year, month, day);
    days.push(createCalendarDay(date, true, today, selectedDate, minDate, maxDate, disabledDates));
  }

  // Add days from next month
  const remainingDays = 42 - days.length; // 6 rows * 7 days
  for (let day = 1; day <= remainingDays; day++) {
    const date = new Date(year, month + 1, day);
    days.push(createCalendarDay(date, false, today, selectedDate, minDate, maxDate, disabledDates));
  }

  return days;
}

function createCalendarDay(
  date: Date,
  isCurrentMonth: boolean,
  today: Date,
  selectedDate: Date | Date[] | DateRange | null | undefined,
  minDate?: Date,
  maxDate?: Date,
  disabledDates?: Date[]
): CalendarDay {
  const dateTime = date.getTime();
  const isToday = date.toDateString() === today.toDateString();

  let isSelected = false;
  let isInRange = false;
  let isRangeStart = false;
  let isRangeEnd = false;

  if (selectedDate) {
    if (selectedDate instanceof Date) {
      isSelected = date.toDateString() === selectedDate.toDateString();
    } else if (Array.isArray(selectedDate)) {
      isSelected = selectedDate.some(d => d.toDateString() === date.toDateString());
    } else if ('start' in selectedDate && 'end' in selectedDate) {
      const range = selectedDate as DateRange;
      if (range.start) {
        isRangeStart = date.toDateString() === range.start.toDateString();
        isSelected = isRangeStart;
      }
      if (range.end) {
        isRangeEnd = date.toDateString() === range.end.toDateString();
        isSelected = isSelected || isRangeEnd;
      }
      if (range.start && range.end) {
        isInRange = dateTime > range.start.getTime() && dateTime < range.end.getTime();
      }
    }
  }

  let isDisabled = false;
  if (minDate && dateTime < minDate.getTime()) isDisabled = true;
  if (maxDate && dateTime > maxDate.getTime()) isDisabled = true;
  if (disabledDates?.some(d => d.toDateString() === date.toDateString())) isDisabled = true;

  return {
    date,
    day: date.getDate(),
    isCurrentMonth,
    isToday,
    isSelected,
    isInRange,
    isRangeStart,
    isRangeEnd,
    isDisabled,
  };
}

export function formatDate(
  date: Date | Date[] | DateRange | null | undefined,
  format: string = 'MM/DD/YYYY',
  locale: string = 'en-US'
): string {
  if (!date) return '';

  if (date instanceof Date) {
    return formatSingleDate(date, format, locale);
  }

  if (Array.isArray(date)) {
    return date.map(d => formatSingleDate(d, format, locale)).join(', ');
  }

  if ('start' in date && 'end' in date) {
    const range = date as DateRange;
    const start = range.start ? formatSingleDate(range.start, format, locale) : '';
    const end = range.end ? formatSingleDate(range.end, format, locale) : '';
    if (start && end) return `${start} - ${end}`;
    return start || end;
  }

  return '';
}

function formatSingleDate(date: Date, format: string, locale: string): string {
  const options: Intl.DateTimeFormatOptions = {};

  if (format.includes('YYYY')) options.year = 'numeric';
  else if (format.includes('YY')) options.year = '2-digit';

  if (format.includes('MMMM')) options.month = 'long';
  else if (format.includes('MMM')) options.month = 'short';
  else if (format.includes('MM')) options.month = '2-digit';

  if (format.includes('DD')) options.day = '2-digit';
  else if (format.includes('D')) options.day = 'numeric';

  return new Intl.DateTimeFormat(locale, options).format(date);
}

export function parseDate(value: string, format: string = 'MM/DD/YYYY'): Date | null {
  if (!value) return null;

  const parts = value.split(/[\/\-\.]/);
  if (parts.length !== 3) return null;

  let year: number, month: number, day: number;

  if (format.startsWith('DD')) {
    day = parseInt(parts[0]!, 10);
    month = parseInt(parts[1]!, 10) - 1;
    year = parseInt(parts[2]!, 10);
  } else if (format.startsWith('YYYY')) {
    year = parseInt(parts[0]!, 10);
    month = parseInt(parts[1]!, 10) - 1;
    day = parseInt(parts[2]!, 10);
  } else {
    month = parseInt(parts[0]!, 10) - 1;
    day = parseInt(parts[1]!, 10);
    year = parseInt(parts[2]!, 10);
  }

  if (year < 100) year += 2000;

  const date = new Date(year, month, day);
  if (isNaN(date.getTime())) return null;

  return date;
}
