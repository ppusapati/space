/**
 * Theme Store
 * Handles theming, color modes, and visual preferences
 */

import { writable, derived, get, type Readable } from 'svelte/store';

// ============================================================================
// TYPES
// ============================================================================

export type ThemeMode = 'light' | 'dark' | 'system';
export type ColorScheme = 'light' | 'dark';
export type Density = 'compact' | 'default' | 'comfortable';
export type FontSize = 'sm' | 'md' | 'lg';
export type Radius = 'none' | 'sm' | 'md' | 'lg' | 'full';

export interface ThemeColors {
  primary: string;
  secondary: string;
  accent: string;
  success: string;
  warning: string;
  error: string;
  info: string;
  neutral: string;
}

export interface ThemeState {
  mode: ThemeMode;
  resolvedScheme: ColorScheme;
  colors: ThemeColors;
  density: Density;
  fontSize: FontSize;
  radius: Radius;
  reducedMotion: boolean;
  highContrast: boolean;
  isLoading: boolean;
}

// ============================================================================
// INITIAL STATE
// ============================================================================

const defaultColors: ThemeColors = {
  primary: '#3b82f6',
  secondary: '#64748b',
  accent: '#f59e0b',
  success: '#22c55e',
  warning: '#eab308',
  error: '#ef4444',
  info: '#06b6d4',
  neutral: '#6b7280',
};

const initialState: ThemeState = {
  mode: 'system',
  resolvedScheme: 'light',
  colors: defaultColors,
  density: 'default',
  fontSize: 'md',
  radius: 'md',
  reducedMotion: false,
  highContrast: false,
  isLoading: true,
};

// ============================================================================
// STORE CREATION
// ============================================================================

function createThemeStore() {
  const store = writable<ThemeState>(initialState);
  const { subscribe, set, update } = store;

  // System preference media query
  let mediaQuery: MediaQueryList | null = null;
  let motionQuery: MediaQueryList | null = null;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const mode: Readable<ThemeMode> = derived(store, ($s) => $s.mode);
  const colorScheme: Readable<ColorScheme> = derived(store, ($s) => $s.resolvedScheme);
  const isDark: Readable<boolean> = derived(store, ($s) => $s.resolvedScheme === 'dark');
  const colors: Readable<ThemeColors> = derived(store, ($s) => $s.colors);
  const density: Readable<Density> = derived(store, ($s) => $s.density);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  function setMode(mode: ThemeMode): void {
    update((s) => ({ ...s, mode }));
    resolveColorScheme();
    persistPreferences();
  }

  function toggleMode(): void {
    const state = get(store);
    const newMode: ThemeMode =
      state.mode === 'light' ? 'dark' : state.mode === 'dark' ? 'system' : 'light';
    setMode(newMode);
  }

  function setColors(colors: Partial<ThemeColors>): void {
    update((s) => ({
      ...s,
      colors: { ...s.colors, ...colors },
    }));
    applyColors();
    persistPreferences();
  }

  function resetColors(): void {
    update((s) => ({
      ...s,
      colors: defaultColors,
    }));
    applyColors();
    persistPreferences();
  }

  function setDensity(density: Density): void {
    update((s) => ({ ...s, density }));
    applyDensity();
    persistPreferences();
  }

  function setFontSize(fontSize: FontSize): void {
    update((s) => ({ ...s, fontSize }));
    applyFontSize();
    persistPreferences();
  }

  function setRadius(radius: Radius): void {
    update((s) => ({ ...s, radius }));
    applyRadius();
    persistPreferences();
  }

  function setReducedMotion(reducedMotion: boolean): void {
    update((s) => ({ ...s, reducedMotion }));
    applyReducedMotion();
    persistPreferences();
  }

  function setHighContrast(highContrast: boolean): void {
    update((s) => ({ ...s, highContrast }));
    applyHighContrast();
    persistPreferences();
  }

  // ============================================================================
  // HELPERS
  // ============================================================================

  function resolveColorScheme(): void {
    const state = get(store);

    let resolvedScheme: ColorScheme;
    if (state.mode === 'system') {
      resolvedScheme = mediaQuery?.matches ? 'dark' : 'light';
    } else {
      resolvedScheme = state.mode;
    }

    update((s) => ({ ...s, resolvedScheme }));
    applyColorScheme(resolvedScheme);
  }

  function applyColorScheme(scheme: ColorScheme): void {
    if (typeof document === 'undefined') return;

    const root = document.documentElement;
    root.classList.remove('light', 'dark');
    root.classList.add(scheme);
    root.setAttribute('data-theme', scheme);
  }

  function applyColors(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    Object.entries(state.colors).forEach(([key, value]) => {
      root.style.setProperty(`--color-${key}`, value);
    });
  }

  function applyDensity(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    root.setAttribute('data-density', state.density);

    // Apply spacing based on density
    const spacingMultiplier =
      state.density === 'compact' ? 0.75 : state.density === 'comfortable' ? 1.25 : 1;
    root.style.setProperty('--density-multiplier', String(spacingMultiplier));
  }

  function applyFontSize(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    const fontSizeMap: Record<FontSize, string> = {
      sm: '14px',
      md: '16px',
      lg: '18px',
    };

    root.style.setProperty('--font-size-base', fontSizeMap[state.fontSize]);
    root.setAttribute('data-font-size', state.fontSize);
  }

  function applyRadius(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    const radiusMap: Record<Radius, string> = {
      none: '0',
      sm: '0.25rem',
      md: '0.5rem',
      lg: '0.75rem',
      full: '9999px',
    };

    root.style.setProperty('--radius-base', radiusMap[state.radius]);
    root.setAttribute('data-radius', state.radius);
  }

  function applyReducedMotion(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    if (state.reducedMotion) {
      root.classList.add('reduce-motion');
    } else {
      root.classList.remove('reduce-motion');
    }
  }

  function applyHighContrast(): void {
    if (typeof document === 'undefined') return;

    const state = get(store);
    const root = document.documentElement;

    if (state.highContrast) {
      root.classList.add('high-contrast');
    } else {
      root.classList.remove('high-contrast');
    }
  }

  function persistPreferences(): void {
    if (typeof localStorage === 'undefined') return;

    const state = get(store);
    const preferences = {
      mode: state.mode,
      colors: state.colors,
      density: state.density,
      fontSize: state.fontSize,
      radius: state.radius,
      reducedMotion: state.reducedMotion,
      highContrast: state.highContrast,
    };

    localStorage.setItem('theme_preferences', JSON.stringify(preferences));
  }

  function loadPreferences(): void {
    if (typeof localStorage === 'undefined') return;

    const stored = localStorage.getItem('theme_preferences');
    if (stored) {
      try {
        const preferences = JSON.parse(stored);
        update((s) => ({
          ...s,
          ...preferences,
          isLoading: false,
        }));
      } catch {
        update((s) => ({ ...s, isLoading: false }));
      }
    } else {
      update((s) => ({ ...s, isLoading: false }));
    }
  }

  function applyAll(): void {
    resolveColorScheme();
    applyColors();
    applyDensity();
    applyFontSize();
    applyRadius();
    applyReducedMotion();
    applyHighContrast();
  }

  function initialize(): void {
    if (typeof window === 'undefined') return;

    // Setup system preference listeners
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', () => {
      const state = get(store);
      if (state.mode === 'system') {
        resolveColorScheme();
      }
    });

    // Setup reduced motion listener
    motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    motionQuery.addEventListener('change', (e) => {
      update((s) => ({ ...s, reducedMotion: e.matches }));
      applyReducedMotion();
    });

    // Load and apply preferences
    loadPreferences();
    applyAll();
  }

  function reset(): void {
    set(initialState);
    localStorage.removeItem('theme_preferences');
    applyAll();
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    mode,
    colorScheme,
    isDark,
    colors,
    density,
    // Actions
    setMode,
    toggleMode,
    setColors,
    resetColors,
    setDensity,
    setFontSize,
    setRadius,
    setReducedMotion,
    setHighContrast,
    initialize,
    reset,
  };
}

// ============================================================================
// EXPORT
// ============================================================================

export const themeStore = createThemeStore();
export { defaultColors };
