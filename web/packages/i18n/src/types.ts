/**
 * I18n Type Definitions
 */

/** Supported locale codes */
export type Locale = 'en' | 'hi' | 'mr' | 'gu' | 'ta' | 'te' | 'kn' | 'bn' | string;

/** Flat or nested translation messages */
export type TranslationMessages = Record<string, string | Record<string, string | Record<string, string>>>;

/** A dot-notation translation key e.g. "common.save" or "agriculture.crop.label" */
export type TranslationKey = string;

/** i18n configuration */
export interface I18nConfig {
  /** Default locale */
  locale: Locale;
  /** Fallback locale when a key is missing */
  fallback?: Locale;
  /** Translation messages keyed by locale */
  messages: Record<Locale, TranslationMessages>;
}

/** Interpolation variables for translation strings */
export type InterpolationVars = Record<string, string | number>;
