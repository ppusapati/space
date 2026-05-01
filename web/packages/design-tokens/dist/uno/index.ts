import type { Theme } from '@unocss/preset-uno'
import { designTokensTheme, componentShortcuts, animations } from './theme'

export { designTokensTheme, componentShortcuts, animations }

// The theme already uses CSS variables, so we can export it directly
export const themeConfig: Theme = designTokensTheme

export default themeConfig
