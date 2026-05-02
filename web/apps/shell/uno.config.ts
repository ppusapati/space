import { defineConfig, presetUno } from 'unocss';
import { designTokensTheme, componentShortcuts, animations } from '@@chetana/uno';

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
