/**
 * @samavāya/i18n
 * Internationalization for Samavāya ERP
 *
 * Supports multiple languages across 4 verticals:
 *   - agriculture, manufacturing, water, construction
 *
 * @example
 * import { t, setLocale, getLocale, createI18n } from '@samavāya/i18n';
 *
 * @packageDocumentation
 */

export { createI18n, t, setLocale, getLocale, locale, messages } from './i18n';
export type { I18nConfig, Locale, TranslationMessages, TranslationKey } from './types';
