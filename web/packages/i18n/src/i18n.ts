/**
 * Core i18n implementation using Svelte stores
 */
import { writable, derived, get } from 'svelte/store';
import type { Locale, TranslationMessages, I18nConfig, InterpolationVars, TranslationKey } from './types';

// ============================================================================
// STORES
// ============================================================================

/** Current active locale */
export const locale = writable<Locale>('en');

/** All loaded translation messages */
export const messages = writable<Record<Locale, TranslationMessages>>({} as Record<Locale, TranslationMessages>);

/** Fallback locale */
let _fallback: Locale = 'en';

// ============================================================================
// CORE FUNCTIONS
// ============================================================================

/** Get the current locale value */
export function getLocale(): Locale {
  return get(locale);
}

/** Set the active locale */
export function setLocale(newLocale: Locale): void {
  locale.set(newLocale);
}

/**
 * Resolve a dot-notation key against a messages object.
 * e.g. "common.save" → messages["common"]["save"]
 */
function resolve(msgs: TranslationMessages, key: TranslationKey): string | undefined {
  const parts = key.split('.');
  let current: TranslationMessages | string = msgs;
  for (const part of parts) {
    if (typeof current !== 'object' || current === null) return undefined;
    current = (current as Record<string, TranslationMessages | string>)[part];
  }
  return typeof current === 'string' ? current : undefined;
}

/**
 * Interpolate variables into a translation string.
 * e.g. "Hello {{name}}" + { name: "World" } → "Hello World"
 */
function interpolate(template: string, vars?: InterpolationVars): string {
  if (!vars) return template;
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => String(vars[key] ?? `{{${key}}}`));
}

/**
 * Translate a key to the current locale, with optional variable interpolation.
 *
 * @example
 * t('common.save')                     // "Save"
 * t('common.greeting', { name: 'Raj' }) // "Hello, Raj!"
 * t('agriculture.crop.wheat')           // "Wheat"
 */
export function t(key: TranslationKey, vars?: InterpolationVars): string {
  const currentLocale = get(locale);
  const allMessages = get(messages);

  // Try current locale
  const localeMessages = allMessages[currentLocale];
  if (localeMessages) {
    const value = resolve(localeMessages, key);
    if (value !== undefined) return interpolate(value, vars);
  }

  // Try fallback locale
  if (_fallback && _fallback !== currentLocale) {
    const fallbackMessages = allMessages[_fallback];
    if (fallbackMessages) {
      const value = resolve(fallbackMessages, key);
      if (value !== undefined) return interpolate(value, vars);
    }
  }

  // Return key as last resort
  return key;
}

// ============================================================================
// INITIALIZER
// ============================================================================

/**
 * Initialize the i18n system with config.
 * Call once at app startup.
 *
 * @example
 * import { createI18n } from '@samavāya/i18n';
 * import en from '@samavāya/i18n/locales/en.json';
 * import hi from '@samavāya/i18n/locales/hi.json';
 *
 * createI18n({
 *   locale: 'en',
 *   fallback: 'en',
 *   messages: { en, hi },
 * });
 */
export function createI18n(config: I18nConfig): void {
  _fallback = config.fallback ?? 'en';
  messages.set(config.messages);
  locale.set(config.locale);
}
