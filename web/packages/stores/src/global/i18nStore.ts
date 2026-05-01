/**
 * i18n Store - Internationalization and Localization
 *
 * Features:
 * - Multiple language support
 * - Lazy loading of translations
 * - Pluralization support
 * - Interpolation with variables
 * - Number/Date/Currency formatting
 * - RTL language support
 */

import { writable, derived, get } from 'svelte/store';

// ═══════════════════════════════════════════════════════════════════════════
// TYPES
// ═══════════════════════════════════════════════════════════════════════════

export type Locale = 'en' | 'hi' | 'ta' | 'te' | 'kn' | 'ml' | 'mr' | 'gu' | 'bn' | 'pa';
export type SupportedLocale = Locale;

export interface LocaleConfig {
  code: Locale;
  name: string;
  nativeName: string;
  direction: 'ltr' | 'rtl';
  dateFormat: string;
  timeFormat: string;
  currency: string;
  currencySymbol: string;
  numberFormat: {
    decimal: string;
    thousands: string;
  };
}

export type TranslationValue = string | { [key: string]: TranslationValue };
export type Translations = Record<string, TranslationValue>;
export type TranslationLoader = () => Promise<Translations>;

export interface PluralOptions {
  zero?: string;
  one?: string;
  two?: string;
  few?: string;
  many?: string;
  other: string;
}

export interface I18nState {
  locale: Locale;
  fallbackLocale: Locale;
  translations: Record<Locale, Translations>;
  loadedLocales: Set<Locale>;
  loading: boolean;
}

// ═══════════════════════════════════════════════════════════════════════════
// LOCALE CONFIGURATIONS
// ═══════════════════════════════════════════════════════════════════════════

export const SUPPORTED_LOCALES: Locale[] = ['en', 'hi', 'ta', 'te', 'kn', 'ml', 'mr', 'gu', 'bn', 'pa'];
export const LOCALE_CONFIGS: Record<Locale, LocaleConfig> = {
  en: {
    code: 'en',
    name: 'English',
    nativeName: 'English',
    direction: 'ltr',
    dateFormat: 'MM/DD/YYYY',
    timeFormat: 'hh:mm A',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  hi: {
    code: 'hi',
    name: 'Hindi',
    nativeName: 'हिन्दी',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  ta: {
    code: 'ta',
    name: 'Tamil',
    nativeName: 'தமிழ்',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  te: {
    code: 'te',
    name: 'Telugu',
    nativeName: 'తెలుగు',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  kn: {
    code: 'kn',
    name: 'Kannada',
    nativeName: 'ಕನ್ನಡ',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  ml: {
    code: 'ml',
    name: 'Malayalam',
    nativeName: 'മലയാളം',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  mr: {
    code: 'mr',
    name: 'Marathi',
    nativeName: 'मराठी',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  gu: {
    code: 'gu',
    name: 'Gujarati',
    nativeName: 'ગુજરાતી',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  bn: {
    code: 'bn',
    name: 'Bengali',
    nativeName: 'বাংলা',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
  pa: {
    code: 'pa',
    name: 'Punjabi',
    nativeName: 'ਪੰਜਾਬੀ',
    direction: 'ltr',
    dateFormat: 'DD/MM/YYYY',
    timeFormat: 'HH:mm',
    currency: 'INR',
    currencySymbol: '₹',
    numberFormat: { decimal: '.', thousands: ',' },
  },
};

// ═══════════════════════════════════════════════════════════════════════════
// STORE
// ═══════════════════════════════════════════════════════════════════════════

const initialState: I18nState = {
  locale: 'en',
  fallbackLocale: 'en',
  translations: {} as Record<Locale, Translations>,
  loadedLocales: new Set(['en']),
  loading: false,
};

const store = writable<I18nState>(initialState);

// Translation loaders registry
const translationLoaders: Map<Locale, TranslationLoader> = new Map();

// ═══════════════════════════════════════════════════════════════════════════
// DERIVED STORES
// ═══════════════════════════════════════════════════════════════════════════

export const locale = derived(store, ($store) => $store.locale);
export const localeConfig = derived(store, ($store) => LOCALE_CONFIGS[$store.locale]);
export const direction = derived(store, ($store) => LOCALE_CONFIGS[$store.locale].direction);
export const isRTL = derived(store, ($store) => LOCALE_CONFIGS[$store.locale].direction === 'rtl');
export const loading = derived(store, ($store) => $store.loading);

// ═══════════════════════════════════════════════════════════════════════════
// TRANSLATION FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Get nested value from object using dot notation
 */
function getNestedValue(obj: Translations, path: string): string | undefined {
  const keys = path.split('.');
  let value: TranslationValue | undefined = obj;

  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key];
    } else {
      return undefined;
    }
  }

  return typeof value === 'string' ? value : undefined;
}

/**
 * Interpolate variables in translation string
 * Supports: {variable}, {count, plural, one{...} other{...}}
 */
function interpolate(text: string, params: Record<string, unknown> = {}): string {
  // Simple variable interpolation: {name}
  let result = text.replace(/\{(\w+)\}/g, (_, key) => {
    return params[key] !== undefined ? String(params[key]) : `{${key}}`;
  });

  // Pluralization: {count, plural, =0{none} one{# item} other{# items}}
  result = result.replace(
    /\{(\w+),\s*plural,\s*(?:=0\{([^}]*)\}\s*)?(?:one\{([^}]*)\}\s*)?other\{([^}]*)\}\}/g,
    (_, key, zero, one, other) => {
      const count = Number(params[key]) || 0;
      let selected: string;

      if (count === 0 && zero !== undefined) {
        selected = zero;
      } else if (count === 1 && one !== undefined) {
        selected = one;
      } else {
        selected = other;
      }

      return selected.replace(/#/g, String(count));
    }
  );

  return result;
}

/**
 * Main translation function
 */
export function t(key: string, params: Record<string, unknown> = {}): string {
  const state = get(store);
  const { locale, fallbackLocale, translations } = state;

  // Try current locale
  let value = translations[locale] ? getNestedValue(translations[locale], key) : undefined;

  // Fallback to fallback locale
  if (value === undefined && locale !== fallbackLocale) {
    value = translations[fallbackLocale]
      ? getNestedValue(translations[fallbackLocale], key)
      : undefined;
  }

  // Return key if not found
  if (value === undefined) {
    console.warn(`[i18n] Missing translation: ${key}`);
    return key;
  }

  return interpolate(value, params);
}

/**
 * Reactive translation store
 */
export const _ = derived(store, () => t);

// ═══════════════════════════════════════════════════════════════════════════
// FORMATTING FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Format number according to locale
 */
export function formatNumber(value: number, options?: Intl.NumberFormatOptions): string {
  const state = get(store);
  return new Intl.NumberFormat(state.locale, options).format(value);
}

/**
 * Format currency according to locale
 */
export function formatCurrency(
  value: number,
  currency?: string,
  options?: Intl.NumberFormatOptions
): string {
  const state = get(store);
  const config = LOCALE_CONFIGS[state.locale];
  return new Intl.NumberFormat(state.locale, {
    style: 'currency',
    currency: currency || config.currency,
    ...options,
  }).format(value);
}

/**
 * Format date according to locale
 */
export function formatDate(
  date: Date | string | number,
  options?: Intl.DateTimeFormatOptions
): string {
  const state = get(store);
  const d = date instanceof Date ? date : new Date(date);
  return new Intl.DateTimeFormat(state.locale, options).format(d);
}

/**
 * Format time according to locale
 */
export function formatTime(
  date: Date | string | number,
  options?: Intl.DateTimeFormatOptions
): string {
  const state = get(store);
  const d = date instanceof Date ? date : new Date(date);
  return new Intl.DateTimeFormat(state.locale, {
    hour: 'numeric',
    minute: 'numeric',
    ...options,
  }).format(d);
}

/**
 * Format relative time (e.g., "2 days ago")
 */
export function formatRelativeTime(date: Date | string | number): string {
  const state = get(store);
  const d = date instanceof Date ? date : new Date(date);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  const months = Math.floor(days / 30);
  const years = Math.floor(days / 365);

  const rtf = new Intl.RelativeTimeFormat(state.locale, { numeric: 'auto' });

  if (years > 0) return rtf.format(-years, 'year');
  if (months > 0) return rtf.format(-months, 'month');
  if (days > 0) return rtf.format(-days, 'day');
  if (hours > 0) return rtf.format(-hours, 'hour');
  if (minutes > 0) return rtf.format(-minutes, 'minute');
  return rtf.format(-seconds, 'second');
}

/**
 * Format list according to locale
 */
export function formatList(items: string[], style: 'long' | 'short' | 'narrow' = 'long'): string {
  const state = get(store);
  return new Intl.ListFormat(state.locale, { style, type: 'conjunction' }).format(items);
}

// ═══════════════════════════════════════════════════════════════════════════
// STORE ACTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Register a translation loader for a locale
 */
export function registerTranslations(locale: Locale, loader: TranslationLoader): void {
  translationLoaders.set(locale, loader);
}

/**
 * Load translations for a specific locale
 */
export async function loadTranslations(locale: Locale): Promise<void> {
  const state = get(store);

  // Already loaded
  if (state.loadedLocales.has(locale)) {
    return;
  }

  const loader = translationLoaders.get(locale);
  if (!loader) {
    console.warn(`[i18n] No translation loader registered for locale: ${locale}`);
    return;
  }

  store.update((s) => ({ ...s, loading: true }));

  try {
    const translations = await loader();
    store.update((s) => ({
      ...s,
      translations: { ...s.translations, [locale]: translations },
      loadedLocales: new Set([...s.loadedLocales, locale]),
      loading: false,
    }));
  } catch (error) {
    console.error(`[i18n] Failed to load translations for locale: ${locale}`, error);
    store.update((s) => ({ ...s, loading: false }));
  }
}

/**
 * Set the current locale
 */
export async function setLocale(newLocale: Locale): Promise<void> {
  // Load translations if not loaded
  await loadTranslations(newLocale);

  store.update((s) => ({ ...s, locale: newLocale }));

  // Update HTML attributes
  if (typeof document !== 'undefined') {
    document.documentElement.lang = newLocale;
    document.documentElement.dir = LOCALE_CONFIGS[newLocale].direction;
  }

  // Persist to localStorage
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('locale', newLocale);
  }
}

/**
 * Set translations directly (for SSR or testing)
 */
export function setTranslations(locale: Locale, translations: Translations): void {
  store.update((s) => ({
    ...s,
    translations: { ...s.translations, [locale]: translations },
    loadedLocales: new Set([...s.loadedLocales, locale]),
  }));
}

/**
 * Initialize i18n with default translations
 */
export function initI18n(options: {
  defaultLocale?: Locale;
  fallbackLocale?: Locale;
  translations?: Record<Locale, Translations>;
}): void {
  const { defaultLocale = 'en', fallbackLocale = 'en', translations = {} } = options;

  store.set({
    locale: defaultLocale,
    fallbackLocale,
    translations: translations as Record<Locale, Translations>,
    loadedLocales: new Set(Object.keys(translations) as Locale[]),
    loading: false,
  });

  // Set HTML attributes
  if (typeof document !== 'undefined') {
    document.documentElement.lang = defaultLocale;
    document.documentElement.dir = LOCALE_CONFIGS[defaultLocale].direction;
  }
}

/**
 * Get available locales
 */
export function getAvailableLocales(): LocaleConfig[] {
  return Object.values(LOCALE_CONFIGS);
}

/**
 * Detect user's preferred locale
 */
export function detectLocale(): Locale {
  // Check localStorage
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('locale') as Locale;
    if (saved && saved in LOCALE_CONFIGS) {
      return saved;
    }
  }

  // Check browser language
  if (typeof navigator !== 'undefined') {
    const browserLang = navigator.language.split('-')[0] as Locale;
    if (browserLang in LOCALE_CONFIGS) {
      return browserLang;
    }
  }

  return 'en';
}

// ═══════════════════════════════════════════════════════════════════════════
// EXPORT STORE
// ═══════════════════════════════════════════════════════════════════════════

export const i18nStore = {
  subscribe: store.subscribe,
  setLocale,
  loadTranslations,
  registerTranslations,
  setTranslations,
  initI18n,
  detectLocale,
  getAvailableLocales,
  t,
  formatNumber,
  formatCurrency,
  formatDate,
  formatTime,
  formatRelativeTime,
  formatList,
};
