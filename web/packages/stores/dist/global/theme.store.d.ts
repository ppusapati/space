/**
 * Theme Store
 * Handles theming, color modes, and visual preferences
 */
import { type Readable } from 'svelte/store';
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
declare const defaultColors: ThemeColors;
export declare const themeStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<ThemeState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    mode: Readable<ThemeMode>;
    colorScheme: Readable<ColorScheme>;
    isDark: Readable<boolean>;
    colors: Readable<ThemeColors>;
    density: Readable<Density>;
    setMode: (mode: ThemeMode) => void;
    toggleMode: () => void;
    setColors: (colors: Partial<ThemeColors>) => void;
    resetColors: () => void;
    setDensity: (density: Density) => void;
    setFontSize: (fontSize: FontSize) => void;
    setRadius: (radius: Radius) => void;
    setReducedMotion: (reducedMotion: boolean) => void;
    setHighContrast: (highContrast: boolean) => void;
    initialize: () => void;
    reset: () => void;
};
export { defaultColors };
//# sourceMappingURL=theme.store.d.ts.map