/**
 * Theme Store
 * Manages theme state and provides methods for theme customization
 */

import { writable, derived, get } from 'svelte/store';
import type {
  ThemeState,
  ThemeConfig,
  ThemeMode,
  ThemePreset,
  ColorScale,
  ThemeColors,
  BrandColors,
} from './theme.types.js';
import {
  DEFAULT_COLOR_SCALE,
  DEFAULT_NEUTRAL,
  DEFAULT_TYPOGRAPHY,
  DEFAULT_SPACING,
  DEFAULT_BORDER_RADIUS,
  DEFAULT_SHADOWS,
} from './theme.types.js';

// =============================================================================
// THEME PRESETS
// =============================================================================

const THEME_PRESETS: ThemePreset[] = [
  {
    id: 'ocean',
    name: 'Ocean Blue',
    description: 'A calm, professional blue theme',
    preview: { primary: '#0ea5e9', secondary: '#eab308', accent: '#22c55e' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#f0f9ff', 100: '#e0f2fe', 200: '#bae6fd', 300: '#7dd3fc',
            400: '#38bdf8', 500: '#0ea5e9', 600: '#0284c7', 700: '#0369a1',
            800: '#075985', 900: '#0c4a6e',
          },
          secondary: {
            50: '#fefce8', 100: '#fef9c3', 200: '#fef08a', 300: '#fde047',
            400: '#facc15', 500: '#eab308', 600: '#ca8a04', 700: '#a16207',
            800: '#854d0e', 900: '#713f12',
          },
        },
      },
    },
  },
  {
    id: 'forest',
    name: 'Forest Green',
    description: 'A natural, earthy green theme',
    preview: { primary: '#22c55e', secondary: '#f59e0b', accent: '#3b82f6' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#f0fdf4', 100: '#dcfce7', 200: '#bbf7d0', 300: '#86efac',
            400: '#4ade80', 500: '#22c55e', 600: '#16a34a', 700: '#15803d',
            800: '#166534', 900: '#14532d',
          },
          secondary: {
            50: '#fffbeb', 100: '#fef3c7', 200: '#fde68a', 300: '#fcd34d',
            400: '#fbbf24', 500: '#f59e0b', 600: '#d97706', 700: '#b45309',
            800: '#92400e', 900: '#78350f',
          },
        },
      },
    },
  },
  {
    id: 'sunset',
    name: 'Sunset Orange',
    description: 'A warm, energetic orange theme',
    preview: { primary: '#f97316', secondary: '#8b5cf6', accent: '#06b6d4' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#fff7ed', 100: '#ffedd5', 200: '#fed7aa', 300: '#fdba74',
            400: '#fb923c', 500: '#f97316', 600: '#ea580c', 700: '#c2410c',
            800: '#9a3412', 900: '#7c2d12',
          },
          secondary: {
            50: '#f5f3ff', 100: '#ede9fe', 200: '#ddd6fe', 300: '#c4b5fd',
            400: '#a78bfa', 500: '#8b5cf6', 600: '#7c3aed', 700: '#6d28d9',
            800: '#5b21b6', 900: '#4c1d95',
          },
        },
      },
    },
  },
  {
    id: 'midnight',
    name: 'Midnight Purple',
    description: 'A deep, sophisticated purple theme',
    preview: { primary: '#8b5cf6', secondary: '#ec4899', accent: '#14b8a6' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#f5f3ff', 100: '#ede9fe', 200: '#ddd6fe', 300: '#c4b5fd',
            400: '#a78bfa', 500: '#8b5cf6', 600: '#7c3aed', 700: '#6d28d9',
            800: '#5b21b6', 900: '#4c1d95',
          },
          secondary: {
            50: '#fdf2f8', 100: '#fce7f3', 200: '#fbcfe8', 300: '#f9a8d4',
            400: '#f472b6', 500: '#ec4899', 600: '#db2777', 700: '#be185d',
            800: '#9d174d', 900: '#831843',
          },
        },
      },
    },
  },
  {
    id: 'rose',
    name: 'Rose Pink',
    description: 'A soft, elegant pink theme',
    preview: { primary: '#f43f5e', secondary: '#0ea5e9', accent: '#22c55e' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#fff1f2', 100: '#ffe4e6', 200: '#fecdd3', 300: '#fda4af',
            400: '#fb7185', 500: '#f43f5e', 600: '#e11d48', 700: '#be123c',
            800: '#9f1239', 900: '#881337',
          },
          secondary: {
            50: '#f0f9ff', 100: '#e0f2fe', 200: '#bae6fd', 300: '#7dd3fc',
            400: '#38bdf8', 500: '#0ea5e9', 600: '#0284c7', 700: '#0369a1',
            800: '#075985', 900: '#0c4a6e',
          },
        },
      },
    },
  },
  {
    id: 'slate',
    name: 'Slate Gray',
    description: 'A neutral, professional gray theme',
    preview: { primary: '#64748b', secondary: '#0ea5e9', accent: '#f59e0b' },
    config: {
      colors: {
        brand: {
          primary: {
            50: '#f8fafc', 100: '#f1f5f9', 200: '#e2e8f0', 300: '#cbd5e1',
            400: '#94a3b8', 500: '#64748b', 600: '#475569', 700: '#334155',
            800: '#1e293b', 900: '#0f172a',
          },
          secondary: {
            50: '#f0f9ff', 100: '#e0f2fe', 200: '#bae6fd', 300: '#7dd3fc',
            400: '#38bdf8', 500: '#0ea5e9', 600: '#0284c7', 700: '#0369a1',
            800: '#075985', 900: '#0c4a6e',
          },
        },
      },
    },
  },
];

// =============================================================================
// DEFAULT THEME
// =============================================================================

function createDefaultTheme(mode: ThemeMode = 'light'): ThemeConfig {
  const isLight = mode === 'light';

  return {
    id: 'default',
    name: 'Default Theme',
    description: 'The default samavāya theme',
    mode,
    colors: {
      brand: {
        primary: DEFAULT_COLOR_SCALE,
        secondary: {
          50: '#fefce8', 100: '#fef9c3', 200: '#fef08a', 300: '#fde047',
          400: '#facc15', 500: '#eab308', 600: '#ca8a04', 700: '#a16207',
          800: '#854d0e', 900: '#713f12',
        },
      },
      semantic: {
        success: {
          50: '#f0fdf4', 100: '#dcfce7', 200: '#bbf7d0', 300: '#86efac',
          400: '#4ade80', 500: '#22c55e', 600: '#16a34a', 700: '#15803d',
          800: '#166534', 900: '#14532d',
        },
        warning: {
          50: '#fffbeb', 100: '#fef3c7', 200: '#fde68a', 300: '#fcd34d',
          400: '#fbbf24', 500: '#f59e0b', 600: '#d97706', 700: '#b45309',
          800: '#92400e', 900: '#78350f',
        },
        error: {
          50: '#fef2f2', 100: '#fee2e2', 200: '#fecaca', 300: '#fca5a5',
          400: '#f87171', 500: '#ef4444', 600: '#dc2626', 700: '#b91c1c',
          800: '#991b1b', 900: '#7f1d1d',
        },
        info: {
          50: '#eff6ff', 100: '#dbeafe', 200: '#bfdbfe', 300: '#93c5fd',
          400: '#60a5fa', 500: '#3b82f6', 600: '#2563eb', 700: '#1d4ed8',
          800: '#1e40af', 900: '#1e3a8a',
        },
      },
      neutral: DEFAULT_NEUTRAL,
      surface: {
        primary: isLight ? DEFAULT_NEUTRAL.white : DEFAULT_NEUTRAL[900],
        secondary: isLight ? DEFAULT_NEUTRAL[50] : DEFAULT_NEUTRAL[850],
        tertiary: isLight ? DEFAULT_NEUTRAL[100] : DEFAULT_NEUTRAL[800],
        inverse: isLight ? DEFAULT_NEUTRAL[900] : DEFAULT_NEUTRAL.white,
        overlay: 'rgba(0, 0, 0, 0.5)',
      },
      text: {
        primary: isLight ? DEFAULT_NEUTRAL[900] : DEFAULT_NEUTRAL[50],
        secondary: isLight ? DEFAULT_NEUTRAL[700] : DEFAULT_NEUTRAL[300],
        tertiary: isLight ? DEFAULT_NEUTRAL[500] : DEFAULT_NEUTRAL[400],
        inverse: isLight ? DEFAULT_NEUTRAL.white : DEFAULT_NEUTRAL[900],
        placeholder: DEFAULT_NEUTRAL[400],
        disabled: isLight ? DEFAULT_NEUTRAL[300] : DEFAULT_NEUTRAL[600],
      },
      border: {
        primary: isLight ? DEFAULT_NEUTRAL[200] : DEFAULT_NEUTRAL[700],
        secondary: isLight ? DEFAULT_NEUTRAL[300] : DEFAULT_NEUTRAL[600],
        focus: DEFAULT_COLOR_SCALE[500],
        inverse: isLight ? DEFAULT_NEUTRAL[700] : DEFAULT_NEUTRAL[200],
      },
      interactive: {
        primary: DEFAULT_COLOR_SCALE[500],
        primaryHover: DEFAULT_COLOR_SCALE[600],
        primaryActive: DEFAULT_COLOR_SCALE[700],
        secondary: isLight ? DEFAULT_NEUTRAL[100] : DEFAULT_NEUTRAL[800],
        secondaryHover: isLight ? DEFAULT_NEUTRAL[200] : DEFAULT_NEUTRAL[700],
        secondaryActive: isLight ? DEFAULT_NEUTRAL[300] : DEFAULT_NEUTRAL[600],
      },
    },
    typography: DEFAULT_TYPOGRAPHY,
    spacing: DEFAULT_SPACING,
    borderRadius: DEFAULT_BORDER_RADIUS,
    shadows: DEFAULT_SHADOWS,
    createdAt: new Date(),
    updatedAt: new Date(),
    isDefault: true,
    isSystem: true,
  };
}

// =============================================================================
// INITIAL STATE
// =============================================================================

const initialState: ThemeState = {
  currentThemeId: 'default',
  mode: 'light',
  themes: [createDefaultTheme('light'), createDefaultTheme('dark')],
  presets: THEME_PRESETS,
  customCss: '',
  isLoading: false,
  error: null,
};

// =============================================================================
// STORE CREATION
// =============================================================================

function createThemeStore() {
  const { subscribe, set, update } = writable<ThemeState>(initialState);

  // Detect system preference
  function getSystemMode(): 'light' | 'dark' {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return 'light';
  }

  // Generate CSS variables from theme config
  function generateCssVariables(theme: ThemeConfig): string {
    const vars: string[] = [];

    // Brand colors
    Object.entries(theme.colors.brand).forEach(([brandName, scale]) => {
      Object.entries(scale as ColorScale).forEach(([shade, value]) => {
        vars.push(`--color-${brandName}-${shade}: ${value};`);
      });
    });

    // Semantic colors
    Object.entries(theme.colors.semantic).forEach(([semantic, scale]) => {
      Object.entries(scale as ColorScale).forEach(([shade, value]) => {
        vars.push(`--color-${semantic}-${shade}: ${value};`);
      });
    });

    // Neutral colors
    Object.entries(theme.colors.neutral).forEach(([shade, value]) => {
      vars.push(`--color-neutral-${shade}: ${value};`);
    });

    // Surface colors
    Object.entries(theme.colors.surface).forEach(([name, value]) => {
      vars.push(`--color-surface-${name}: ${value};`);
    });

    // Text colors
    Object.entries(theme.colors.text).forEach(([name, value]) => {
      vars.push(`--color-text-${name}: ${value};`);
    });

    // Border colors
    Object.entries(theme.colors.border).forEach(([name, value]) => {
      vars.push(`--color-border-${name}: ${value};`);
    });

    // Interactive colors
    Object.entries(theme.colors.interactive).forEach(([name, value]) => {
      vars.push(`--color-interactive-${name}: ${value};`);
    });

    // Typography
    Object.entries(theme.typography.fontFamily).forEach(([name, value]) => {
      vars.push(`--font-family-${name}: ${value};`);
    });
    Object.entries(theme.typography.fontSize).forEach(([name, value]) => {
      vars.push(`--font-size-${name}: ${value};`);
    });
    Object.entries(theme.typography.fontWeight).forEach(([name, value]) => {
      vars.push(`--font-weight-${name}: ${value};`);
    });
    Object.entries(theme.typography.lineHeight).forEach(([name, value]) => {
      vars.push(`--line-height-${name}: ${value};`);
    });

    // Spacing
    Object.entries(theme.spacing).forEach(([name, value]) => {
      vars.push(`--spacing-${name}: ${value};`);
    });

    // Border radius
    Object.entries(theme.borderRadius).forEach(([name, value]) => {
      vars.push(`--radius-${name}: ${value};`);
    });

    // Shadows
    Object.entries(theme.shadows).forEach(([name, value]) => {
      vars.push(`--shadow-${name}: ${value};`);
    });

    return `:root {\n  ${vars.join('\n  ')}\n}`;
  }

  // Apply theme to document
  function applyTheme(theme: ThemeConfig) {
    if (typeof document === 'undefined') return;

    const css = generateCssVariables(theme);

    let styleEl = document.getElementById('samavāya-theme-vars');
    if (!styleEl) {
      styleEl = document.createElement('style');
      styleEl.id = 'samavāya-theme-vars';
      document.head.appendChild(styleEl);
    }
    styleEl.textContent = css;

    // Set data attribute for theme mode
    document.documentElement.setAttribute('data-theme', theme.mode);
    document.documentElement.classList.toggle('dark', theme.mode === 'dark');
  }

  return {
    subscribe,

    // Initialize theme from storage or system preference
    initialize: () => {
      update(state => {
        const savedThemeId = typeof localStorage !== 'undefined'
          ? localStorage.getItem('samavāya-theme-id')
          : null;
        const savedMode = typeof localStorage !== 'undefined'
          ? localStorage.getItem('samavāya-theme-mode') as ThemeMode
          : null;

        const mode = savedMode || (state.mode === 'system' ? getSystemMode() : state.mode);
        const themeId = savedThemeId || 'default';
        const theme = state.themes.find((t: ThemeConfig) => t.id === themeId && t.mode === mode)
          || state.themes.find((t: ThemeConfig) => t.mode === mode)
          || createDefaultTheme(mode as 'light' | 'dark');

        applyTheme(theme);

        return {
          ...state,
          currentThemeId: theme.id,
          mode: savedMode || state.mode,
        };
      });
    },

    // Set theme mode
    setMode: (mode: ThemeMode) => {
      update(state => {
        const effectiveMode = mode === 'system' ? getSystemMode() : mode;
        const theme = state.themes.find((t: ThemeConfig) => t.id === state.currentThemeId && t.mode === effectiveMode)
          || state.themes.find((t: ThemeConfig) => t.mode === effectiveMode)
          || createDefaultTheme(effectiveMode);

        applyTheme(theme);

        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('samavāya-theme-mode', mode);
        }

        return { ...state, mode };
      });
    },

    // Toggle between light and dark
    toggleMode: () => {
      update(state => {
        const newMode = state.mode === 'light' ? 'dark' : 'light';
        const theme = state.themes.find((t: ThemeConfig) => t.id === state.currentThemeId && t.mode === newMode)
          || state.themes.find((t: ThemeConfig) => t.mode === newMode)
          || createDefaultTheme(newMode);

        applyTheme(theme);

        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('samavāya-theme-mode', newMode);
        }

        return { ...state, mode: newMode };
      });
    },

    // Apply a preset
    applyPreset: (presetId: string) => {
      update(state => {
        const preset = state.presets.find((p: ThemePreset) => p.id === presetId);
        if (!preset) return state;

        const baseTheme = createDefaultTheme(state.mode === 'system' ? getSystemMode() : state.mode as 'light' | 'dark');
        const presetColors = preset.config.colors;

        // Merge preset brand colors with defaults
        const mergedBrand: BrandColors = {
          primary: { ...baseTheme.colors.brand.primary, ...presetColors?.brand?.primary },
          secondary: { ...baseTheme.colors.brand.secondary, ...presetColors?.brand?.secondary },
        };

        const newTheme: ThemeConfig = {
          ...baseTheme,
          id: `custom-${Date.now()}`,
          name: preset.name,
          description: preset.description,
          colors: {
            ...baseTheme.colors,
            brand: mergedBrand,
          },
          isDefault: false,
          isSystem: false,
          createdAt: new Date(),
          updatedAt: new Date(),
        };

        applyTheme(newTheme);

        return {
          ...state,
          currentThemeId: newTheme.id,
          themes: [...state.themes, newTheme],
        };
      });
    },

    // Create a new custom theme
    createTheme: (config: Partial<ThemeConfig>) => {
      update(state => {
        const baseTheme = createDefaultTheme(config.mode || 'light');
        const newTheme: ThemeConfig = {
          ...baseTheme,
          ...config,
          id: `custom-${Date.now()}`,
          isDefault: false,
          isSystem: false,
          createdAt: new Date(),
          updatedAt: new Date(),
        };

        return {
          ...state,
          themes: [...state.themes, newTheme],
        };
      });
    },

    // Update an existing theme
    updateTheme: (themeId: string, updates: Partial<ThemeConfig>) => {
      update(state => {
        const themes = state.themes.map((theme: ThemeConfig) => {
          if (theme.id === themeId) {
            const updated = { ...theme, ...updates, updatedAt: new Date() };
            if (theme.id === state.currentThemeId) {
              applyTheme(updated);
            }
            return updated;
          }
          return theme;
        });

        return { ...state, themes };
      });
    },

    // Delete a custom theme
    deleteTheme: (themeId: string) => {
      update(state => {
        const theme = state.themes.find((t: ThemeConfig) => t.id === themeId);
        if (!theme || theme.isSystem) return state;

        const themes = state.themes.filter((t: ThemeConfig) => t.id !== themeId);
        const currentThemeId = state.currentThemeId === themeId ? 'default' : state.currentThemeId;

        if (state.currentThemeId === themeId) {
          const defaultTheme = themes.find((t: ThemeConfig) => t.id === 'default' && t.mode === state.mode)
            || themes[0];
          if (defaultTheme) applyTheme(defaultTheme);
        }

        return { ...state, themes, currentThemeId };
      });
    },

    // Set active theme
    setTheme: (themeId: string) => {
      update(state => {
        const theme = state.themes.find((t: ThemeConfig) => t.id === themeId);
        if (!theme) return state;

        applyTheme(theme);

        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('samavāya-theme-id', themeId);
        }

        return { ...state, currentThemeId: themeId };
      });
    },

    // Set custom CSS
    setCustomCss: (css: string) => {
      update(state => {
        if (typeof document !== 'undefined') {
          let styleEl = document.getElementById('samavāya-custom-css');
          if (!styleEl) {
            styleEl = document.createElement('style');
            styleEl.id = 'samavāya-custom-css';
            document.head.appendChild(styleEl);
          }
          styleEl.textContent = css;
        }
        return { ...state, customCss: css };
      });
    },

    // Export theme as JSON
    exportTheme: (themeId: string): string | null => {
      const state = get({ subscribe });
      const theme = state.themes.find((t: ThemeConfig) => t.id === themeId);
      if (!theme) return null;
      return JSON.stringify(theme, null, 2);
    },

    // Import theme from JSON
    importTheme: (json: string) => {
      try {
        const imported = JSON.parse(json) as ThemeConfig;
        update(state => ({
          ...state,
          themes: [
            ...state.themes,
            {
              ...imported,
              id: `imported-${Date.now()}`,
              isSystem: false,
              createdAt: new Date(),
              updatedAt: new Date(),
            },
          ],
        }));
        return true;
      } catch {
        return false;
      }
    },

    // Reset to default theme
    reset: () => {
      set(initialState);
      const defaultTheme = createDefaultTheme('light');
      applyTheme(defaultTheme);
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem('samavāya-theme-id');
        localStorage.removeItem('samavāya-theme-mode');
      }
    },
  };
}

// =============================================================================
// EXPORTS
// =============================================================================

export const themeStore = createThemeStore();

// Derived stores for convenience
export const currentTheme = derived(themeStore, ($store: ThemeState) => {
  const effectiveMode = $store.mode === 'system'
    ? (typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : $store.mode;
  return $store.themes.find((t: ThemeConfig) => t.id === $store.currentThemeId && t.mode === effectiveMode)
    || $store.themes.find((t: ThemeConfig) => t.mode === effectiveMode)
    || createDefaultTheme(effectiveMode as 'light' | 'dark');
});

export const themeMode = derived(themeStore, ($store: ThemeState) => $store.mode);
export const themePresets = derived(themeStore, ($store: ThemeState) => $store.presets);
export const customThemes = derived(themeStore, ($store: ThemeState) =>
  $store.themes.filter((t: ThemeConfig) => !t.isSystem)
);
