import { defineConfig, presetUno } from 'unocss';
import { designTokensTheme, componentShortcuts, animations } from '@p9e.in/samavaya/uno';

export default defineConfig({
  presets: [
    presetUno(),
  ],
  theme: {
    ...designTokensTheme,
    animation: animations,
  },
  shortcuts: componentShortcuts,
});
