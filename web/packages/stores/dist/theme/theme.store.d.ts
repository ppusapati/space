/**
 * Theme Store
 * Manages theme state and provides methods for theme customization
 */
import type { ThemeState, ThemeConfig, ThemeMode, ThemePreset } from './theme.types.js';
export declare const themeStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<ThemeState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    initialize: () => void;
    setMode: (mode: ThemeMode) => void;
    toggleMode: () => void;
    applyPreset: (presetId: string) => void;
    createTheme: (config: Partial<ThemeConfig>) => void;
    updateTheme: (themeId: string, updates: Partial<ThemeConfig>) => void;
    deleteTheme: (themeId: string) => void;
    setTheme: (themeId: string) => void;
    setCustomCss: (css: string) => void;
    exportTheme: (themeId: string) => string | null;
    importTheme: (json: string) => boolean;
    reset: () => void;
};
export declare const currentTheme: import("svelte/store").Readable<ThemeConfig>;
export declare const themeMode: import("svelte/store").Readable<ThemeMode>;
export declare const themePresets: import("svelte/store").Readable<ThemePreset[]>;
export declare const customThemes: import("svelte/store").Readable<ThemeConfig[]>;
//# sourceMappingURL=theme.store.d.ts.map