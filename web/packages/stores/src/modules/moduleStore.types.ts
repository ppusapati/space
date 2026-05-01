/**
 * Type declarations for moduleStore.
 *
 * Why a separate file: `moduleStore.svelte.ts` is a Svelte 5 rune
 * module compiled by `svelte.compileModule()`, which calls
 * `acorn-typescript` with a strict subset of TS. Top-level
 * `interface` declarations crash the parser with
 *   `Unexpected token (19:7) [plugin vite-plugin-svelte-module:optimize-svelte]`.
 * Parameter / return type annotations and generic call expressions
 * (`$state<T>(…)`) are accepted, but `interface ... { }` is not.
 *
 * Splitting types into a plain `.ts` sibling keeps the rune file
 * minimal-typed (only annotations + generic-on-rune) while letting
 * importers consume the types from `@samavāya/stores/modules` via
 * the index barrel.
 *
 * 2026-04-29 — extracted to unstick the FE dev server boot.
 */

/** API module summary (mirrors formservice.proto ModuleSummary) */
export interface ApiModuleSummary {
  moduleId: string;
  label: string;
  formCount: number;
}

/** API form summary (mirrors formservice.proto FormSummary) */
export interface ApiFormSummary {
  formId: string;
  title: string;
  description: string;
  friendlyEndpoint: string;
  rpcEndpoint: string;
  moduleId: string;
  version: string;
}

export interface ModuleStoreState {
  /** All available modules */
  modules: ApiModuleSummary[];
  /** Currently selected module ID */
  selectedModuleId: string | null;
  /** Forms for the currently selected module */
  forms: ApiFormSummary[];
  /** Whether modules are being loaded */
  isLoadingModules: boolean;
  /** Whether forms are being loaded */
  isLoadingForms: boolean;
  /** Error from last module fetch */
  moduleError: string | null;
  /** Error from last forms fetch */
  formError: string | null;
  /** Whether the data came from API (true) or static fallback (false) */
  isApiDriven: boolean;
}
