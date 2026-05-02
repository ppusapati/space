import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import UnoCSS from 'unocss/vite';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const unoCss = UnoCSS() as any;

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Resolve a single hoisted tslib so all transitive consumers (echarts,
// zrender, etc.) share one version AND get pure native-ESM with named
// exports. Why this exact target file:
//
//   tslib has THREE variants in its package:
//     • tslib.js          — UMD/CJS (legacy default).
//     • tslib.es6.js      — native ESM with `export function __extends`
//                            and friends. NO default export.
//     • modules/index.js  — re-exports tslib.js via `import tslib from`
//                            then `const { __extends } = tslib;` then
//                            named re-exports. Looks like ESM but pulls
//                            CJS through the legacy default-import
//                            interop, which esbuild compiles to
//                            `import_tslib.default` — that path is what
//                            crashes at runtime when echarts' bundled
//                            consumer does the same destructure pattern
//                            and `default` is undefined.
//
// Aliasing the bare specifier `tslib` to `tslib.es6.js` makes every
// consumer (including echarts' nested copy via the alias matching the
// bare specifier inside echarts' source) resolve to the same
// truly-ESM file with NAMED exports. esbuild then emits direct
// `__extends` references with no `.default` indirection, fixing the
// runtime crash for the entire dep graph.
const hoistedTslib = path.resolve(
  __dirname,
  '../../node_modules/tslib/tslib.es6.js',
);

export default defineConfig({
  plugins: [
    unoCss,
    sveltekit(),
  ],

  resolve: {
    alias: [
      // Force every `import * from "tslib"` (and the nested copies in
      // echarts/zrender) to resolve to the single hoisted version.
      // 2026-04-29 — see comment block above for the runtime defect
      // this prevents.
      { find: /^tslib$/, replacement: hoistedTslib },
    ],
  },

  server: {
    // 2026-04-29: switched from 5173 → 6060 to avoid IPv6 collision with
    // a parallel project (UGCL web) listening on [::1]:5173. strictPort
    // is true so a busy port fails loudly during demo runs instead of
    // silently picking 5174 / 5175 / etc and leaving the user wondering
    // why the .env API URL doesn't line up. Override per-host via
    // VITE_DEV_PORT if 6060 is also taken.
    port: Number(process.env.VITE_DEV_PORT ?? 6060),
    strictPort: true,
    host: true,
  },

  preview: {
    port: 4173,
    strictPort: false,
  },

  optimizeDeps: {
    // 2026-04-29: handling for `@chetana/stores`.
    //
    // Vite's optimizer pre-bundles linked workspace deps via esbuild +
    // the @sveltejs/vite-plugin-svelte module plugin. The module
    // plugin's filter is `/\.svelte\.[jt]s$/` and it pipes matching
    // files straight to `svelte.compileModule()` — which accepts ONLY
    // a JS-with-runes subset (no `interface`, no `import type`, no
    // `: ReturnType` annotations). vitePreprocess() is NOT applied
    // on this path; it only runs on `.svelte` components.
    //
    // `packages/stores/src/modules/moduleStore.svelte.ts` is a real
    // rune store (`let state = $state<T>(...)`) but the file has TS
    // type annotations throughout. Because rune syntax is required,
    // the file has to keep its `.svelte.ts` extension; it can't be
    // renamed to plain `.ts` (svelte would reject the `$state` rune).
    //
    // The clean fix: exclude `@chetana/stores` from pre-bundling.
    // SvelteKit's plugin (loaded via sveltekit() above) processes the
    // package through `ssr.noExternal`, which correctly handles BOTH
    // type annotations AND rune syntax via vitePreprocess. We lose
    // the small startup-time win of pre-bundling but gain a working
    // dev server.
    //
    // UI + core remain in `include` because they genuinely benefit
    // from pre-bundling and contain no `.svelte.ts` rune modules.
    include: ['@chetana/ui', '@chetana/core'],
    exclude: ['@chetana/stores'],
  },

  ssr: {
    noExternal: [/^p9e.in/chetana\//, /^@p9e\.in\//],
  },

  build: {
    target: 'esnext',
  },
});
