# @p9e/design-tokens

Enterprise design tokens for MFE applications. Ships CSS variables, theme utilities, and UnoCSS configuration wired to the same variables for seamless light/dark theming.

## Features
- CSS variables for all tokens (combined `tokens.css`, and per-theme CSS)
- Light/Dark with auto mode via `data-theme` and `prefers-color-scheme`
- JS utilities for theme switching and reading CSS variables
- UnoCSS theme, shortcuts, and animations that read the same CSS variables
- Generated TypeScript token names/types

## Build
- One-time build
```bash
pnpm --filter @p9e/design-tokens build
```
- Watch (Windows-safe)
```bash
pnpm --filter @p9e/design-tokens dev
```
This runs nodemon to rebuild on changes in `tokens/`.

Generated outputs:
- `dist/css/`
  - `tokens.css` (MAIN: base + light + dark + keyframes)
  - `tokens-light.css`
  - `tokens-dark.css`
  - `keyframes.css`
- `dist/js/`
  - `tokens.js`
  - `theme-utils.js`
- `dist/uno/`
  - `index.ts` (MAIN: Uno exports)
  - `theme.ts`
- `dist/types/`
  - `tokens.d.ts`

## Consuming in an app
Always include the CSS so variables exist at runtime.

- Preferred (via subpath export):
```ts
import '@p9e/design-tokens/css'
```
- Or direct import:
```ts
import '@p9e/design-tokens/dist/css/tokens.css'
```

Use variables in CSS:
```css
.button {
  background: var(--color-brand-primary-500);
  color: var(--color-neutral-white);
  padding: var(--spacing-md);
}
```

## Theme switching
`tokens.css` supports:
- `:root { ... }` base and light overrides
- `[data-theme="dark"] { ... }` dark overrides
- `@media (prefers-color-scheme: dark)` for auto

Minimal toggle:
```ts
document.documentElement.setAttribute('data-theme', 'dark')
```

Use the helper utilities:
```ts
import { ThemeManager, themeManager, getThemeValue } from '@p9e/design-tokens/js/theme-utils'

// Set or toggle theme
themeManager.setTheme('dark')   // 'light' | 'dark' | 'auto'
// themeManager.toggleTheme()

// Read a variable (respects current theme)
const primary500 = getThemeValue('color.brand.primary.500') // -> var(--color-brand-primary-500)
```

## UnoCSS integration
Use the provided Uno theme and shortcuts that reference the CSS variables.

`uno.config.ts`
```ts
import { defineConfig, presetUno } from 'unocss'
import { themeConfig, componentShortcuts, animations } from '@p9e/design-tokens/uno'

export default defineConfig({
  presets: [presetUno()],
  theme: themeConfig,
  shortcuts: componentShortcuts,
})
```

In your app entry, ensure CSS is loaded:
```ts
import '@p9e/design-tokens/css'
```

Use shortcuts/utilities in markup:
```html
<button class="btn-primary">Primary</button>
<div class="p-spacing-md bg-color-brand-primary-50 text-color-neutral-900"></div>
```

## Using tokens from JS/TS
- Static token values (Style Dictionary output):
```ts
import tokens from '@p9e/design-tokens/js'
console.log(tokens)
```
- Read runtime variable value (theme-aware):
```ts
import { getThemeValue } from '@p9e/design-tokens/js/theme-utils'
const spacingMd = getThemeValue('spacing.md')
```

## Types
The package root `types` points to `dist/types/tokens.d.ts`.
You can also import subpath types if needed:
```ts
import type { DesignTokens } from '@p9e/design-tokens/types'
```

## Troubleshooting (Windows)
- Watch script uses `"pnpm.cmd run build"` under nodemon `--exec` for Windows shells.
- Ensure pnpm is available on PATH: `npm i -g pnpm` and restart terminal.

## Project structure
- `tokens/` – source token JSON files (global, layout, components, themes)
- `scripts/` – build, validate, semantic token helpers
- `dist/` – generated CSS/JS/Uno/types

## License
Internal use.
