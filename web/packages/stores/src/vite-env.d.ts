/// <reference types="vite/client" />

// This brings in `import.meta.env` typings (ImportMetaEnv) so the
// runtime stores in @chetana/stores that read VITE_API_URL etc.
// type-check correctly under svelte-check. Without this triple-slash
// directive, TypeScript's default ImportMeta has no `env` property
// and every reference produces an error.
interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
