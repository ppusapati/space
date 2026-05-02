// eslint.config.js — repo-root ESLint flat config for the web/
// monorepo.
//
// TASK-P0-CI-001. Flat config because the workspace is on eslint v9+
// (per web/packages/configs/package.json). Per-package overrides go
// in their own eslint.config.js if they need them.
//
// Rules selected per design.md §10.4:
//   • TypeScript-eslint recommended      — type-aware checks
//   • Svelte recommended                 — component-level lint
//   • unused-imports                     — fail on unused imports/vars
//   • no-console                         — warn (allow in tests)
//   • REQ-CONST-013 (no fabrications)    — no-restricted-imports for
//                                          legacy brand strings
//
// Run via `pnpm -r lint`; CI runs `task lint:web`.

import tseslint from 'typescript-eslint';
import unusedImports from 'eslint-plugin-unused-imports';
import sveltePlugin from 'eslint-plugin-svelte';

export default [
  // Ignore generated artefacts and vendored code.
  {
    ignores: [
      '**/node_modules/**',
      '**/dist/**',
      '**/.svelte-kit/**',
      '**/.turbo/**',
      '**/build/**',
      '**/coverage/**',
      'web/packages/proto/src/gen/**',
    ],
  },

  // TypeScript baseline.
  ...tseslint.configs.recommendedTypeChecked,

  // Svelte components.
  ...sveltePlugin.configs['flat/recommended'],

  {
    files: ['**/*.{ts,tsx,js,svelte}'],
    plugins: {
      'unused-imports': unusedImports,
    },
    languageOptions: {
      parserOptions: {
        // Project-aware parsing; per-package tsconfig.json drives the
        // type info. `EXPERIMENTAL_useProjectService: true` keeps
        // performance acceptable on the monorepo.
        EXPERIMENTAL_useProjectService: true,
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    rules: {
      // Acceptance criterion: unused imports MUST error.
      'unused-imports/no-unused-imports': 'error',
      'unused-imports/no-unused-vars': [
        'error',
        {
          vars: 'all',
          varsIgnorePattern: '^_',
          args: 'after-used',
          argsIgnorePattern: '^_',
        },
      ],

      // Tighten the typescript-eslint defaults.
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': 'off',          // unused-imports owns this
      '@typescript-eslint/consistent-type-imports': [
        'error',
        { prefer: 'type-imports', disallowTypeAnnotations: false },
      ],
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',

      // Console allowed in tests + scripts; warn elsewhere so prod
      // code uses the structured logger.
      'no-console': ['warn', { allow: ['warn', 'error'] }],

      // REQ-CONST-013 — block legacy brand strings re-entering the tree.
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@samavāya/*', '@samavaya/*'],
              message: 'Use @chetana/<pkg> — see TASK-P0-BRAND-001 in plan/todo.md.',
            },
          ],
        },
      ],
    },
  },

  // Tests + scripts may use console freely.
  {
    files: ['**/*.{test,spec}.{ts,tsx,js}', '**/test/**', '**/scripts/**'],
    rules: {
      'no-console': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
];
