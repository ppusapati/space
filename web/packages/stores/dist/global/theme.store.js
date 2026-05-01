/**
 * Theme Store
 * Handles theming, color modes, and visual preferences
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const defaultColors = {
    primary: '#3b82f6',
    secondary: '#64748b',
    accent: '#f59e0b',
    success: '#22c55e',
    warning: '#eab308',
    error: '#ef4444',
    info: '#06b6d4',
    neutral: '#6b7280',
};
const initialState = {
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
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // System preference media query
    let mediaQuery = null;
    let motionQuery = null;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const mode = derived(store, ($s) => $s.mode);
    const colorScheme = derived(store, ($s) => $s.resolvedScheme);
    const isDark = derived(store, ($s) => $s.resolvedScheme === 'dark');
    const colors = derived(store, ($s) => $s.colors);
    const density = derived(store, ($s) => $s.density);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function setMode(mode) {
        update((s) => ({ ...s, mode }));
        resolveColorScheme();
        persistPreferences();
    }
    function toggleMode() {
        const state = get(store);
        const newMode = state.mode === 'light' ? 'dark' : state.mode === 'dark' ? 'system' : 'light';
        setMode(newMode);
    }
    function setColors(colors) {
        update((s) => ({
            ...s,
            colors: { ...s.colors, ...colors },
        }));
        applyColors();
        persistPreferences();
    }
    function resetColors() {
        update((s) => ({
            ...s,
            colors: defaultColors,
        }));
        applyColors();
        persistPreferences();
    }
    function setDensity(density) {
        update((s) => ({ ...s, density }));
        applyDensity();
        persistPreferences();
    }
    function setFontSize(fontSize) {
        update((s) => ({ ...s, fontSize }));
        applyFontSize();
        persistPreferences();
    }
    function setRadius(radius) {
        update((s) => ({ ...s, radius }));
        applyRadius();
        persistPreferences();
    }
    function setReducedMotion(reducedMotion) {
        update((s) => ({ ...s, reducedMotion }));
        applyReducedMotion();
        persistPreferences();
    }
    function setHighContrast(highContrast) {
        update((s) => ({ ...s, highContrast }));
        applyHighContrast();
        persistPreferences();
    }
    // ============================================================================
    // HELPERS
    // ============================================================================
    function resolveColorScheme() {
        const state = get(store);
        let resolvedScheme;
        if (state.mode === 'system') {
            resolvedScheme = mediaQuery?.matches ? 'dark' : 'light';
        }
        else {
            resolvedScheme = state.mode;
        }
        update((s) => ({ ...s, resolvedScheme }));
        applyColorScheme(resolvedScheme);
    }
    function applyColorScheme(scheme) {
        if (typeof document === 'undefined')
            return;
        const root = document.documentElement;
        root.classList.remove('light', 'dark');
        root.classList.add(scheme);
        root.setAttribute('data-theme', scheme);
    }
    function applyColors() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        Object.entries(state.colors).forEach(([key, value]) => {
            root.style.setProperty(`--color-${key}`, value);
        });
    }
    function applyDensity() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        root.setAttribute('data-density', state.density);
        // Apply spacing based on density
        const spacingMultiplier = state.density === 'compact' ? 0.75 : state.density === 'comfortable' ? 1.25 : 1;
        root.style.setProperty('--density-multiplier', String(spacingMultiplier));
    }
    function applyFontSize() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        const fontSizeMap = {
            sm: '14px',
            md: '16px',
            lg: '18px',
        };
        root.style.setProperty('--font-size-base', fontSizeMap[state.fontSize]);
        root.setAttribute('data-font-size', state.fontSize);
    }
    function applyRadius() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        const radiusMap = {
            none: '0',
            sm: '0.25rem',
            md: '0.5rem',
            lg: '0.75rem',
            full: '9999px',
        };
        root.style.setProperty('--radius-base', radiusMap[state.radius]);
        root.setAttribute('data-radius', state.radius);
    }
    function applyReducedMotion() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        if (state.reducedMotion) {
            root.classList.add('reduce-motion');
        }
        else {
            root.classList.remove('reduce-motion');
        }
    }
    function applyHighContrast() {
        if (typeof document === 'undefined')
            return;
        const state = get(store);
        const root = document.documentElement;
        if (state.highContrast) {
            root.classList.add('high-contrast');
        }
        else {
            root.classList.remove('high-contrast');
        }
    }
    function persistPreferences() {
        if (typeof localStorage === 'undefined')
            return;
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
    function loadPreferences() {
        if (typeof localStorage === 'undefined')
            return;
        const stored = localStorage.getItem('theme_preferences');
        if (stored) {
            try {
                const preferences = JSON.parse(stored);
                update((s) => ({
                    ...s,
                    ...preferences,
                    isLoading: false,
                }));
            }
            catch {
                update((s) => ({ ...s, isLoading: false }));
            }
        }
        else {
            update((s) => ({ ...s, isLoading: false }));
        }
    }
    function applyAll() {
        resolveColorScheme();
        applyColors();
        applyDensity();
        applyFontSize();
        applyRadius();
        applyReducedMotion();
        applyHighContrast();
    }
    function initialize() {
        if (typeof window === 'undefined')
            return;
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
    function reset() {
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
