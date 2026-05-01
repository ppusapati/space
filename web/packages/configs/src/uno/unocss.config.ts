// packages/shared-configs/uno.config.ts
import { defineConfig, presetIcons, presetWind3 } from 'unocss'
import type { Preset } from 'unocss'
import { designTokensTheme, componentShortcuts } from '@p9e.in/samavaya/uno'
import { presetP9E } from './preset-p9e.js'

export const createUnoConfig = (customShortcuts = {}) => defineConfig({
    presets: [
        presetWind3(),
        presetIcons({
            scale: 1.2,
            cdn: 'https://esm.sh/',
        }),
        presetP9E() as unknown as Preset
    ],
  theme: designTokensTheme,

  shortcuts: {
    ...componentShortcuts,
    ...customShortcuts
  },
  
  // Note: Design tokens CSS should be imported directly in your app entry point
  // instead of being handled in UnoCSS preflights to avoid path resolution issues
  
  // Content sources for all MFEs
  content: {
    filesystem: [
      'packages/*/src/**/*.{vue,js,ts,jsx,tsx}',
      'apps/*/src/**/*.{vue,js,ts,jsx,tsx}'
    ]
  },
  
  // Enterprise-specific rules
  rules: [
    // Dynamic theme switching
    ['theme-primary', { 'color-scheme': 'light' }],
    ['theme-dark', { 'color-scheme': 'dark' }],
    
    // SCADA specific utility classes
    [/^status-(.+)$/, ([, status]) => {
      const statusColors = {
        'running': 'var(--color-semantic-success-500)',
        'stopped': 'var(--color-neutral-500)',
        'error': 'var(--color-semantic-error-500)',
        'warning': 'var(--color-semantic-warning-500)'
      }
      return {
        'background-color': statusColors[status as keyof typeof statusColors] || statusColors['stopped']
      }
    }],
    
    // ERP specific data grid utilities
    [/^grid-col-(.+)$/, ([, width]) => ({
      'grid-template-columns': `repeat(${width}, minmax(0, 1fr))`
    })]
  ]
})