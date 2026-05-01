import { defineConfig, presetUno } from 'unocss';
import { theme } from './theme';

export const presetP9e = () => {
  return {
    name: 'p9e-preset',
    theme: {
      colors: theme.colors,
      spacing: theme.spacing,
      // Add other theme properties as needed
    },
    // Add any custom UnoCSS rules here
  };
};

export const unoConfig = defineConfig({
  presets: [
    presetUno(),
    presetP9e(),
  ],
});

export default unoConfig;
