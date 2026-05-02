/// <reference types="vite/client" />

// This brings in `import.meta.env` typings (ImportMetaEnv) so Svelte
// components in @chetana/ui that read VITE_API_URL or import.meta.env.DEV
// type-check correctly under svelte-check. Without this triple-slash
// directive, TypeScript's default ImportMeta has no `env` property and
// every reference produces an error.
//
// Custom env vars consumed by this package (extend ImportMetaEnv as new
// VITE_* vars are read by ui code):
interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
