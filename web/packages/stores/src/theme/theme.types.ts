/**
 * Theme Builder Types
 * Defines the structure for customizable themes
 */

// =============================================================================
// COLOR TYPES
// =============================================================================

export interface ColorScale {
  50: string;
  100: string;
  200: string;
  300: string;
  400: string;
  500: string;
  600: string;
  700: string;
  800: string;
  900: string;
}

export interface BrandColors {
  primary: ColorScale;
  secondary: ColorScale;
}

export interface SemanticColors {
  success: ColorScale;
  warning: ColorScale;
  error: ColorScale;
  info: ColorScale;
}

export interface NeutralColors {
  white: string;
  black: string;
  25: string;
  50: string;
  100: string;
  200: string;
  300: string;
  400: string;
  500: string;
  600: string;
  700: string;
  800: string;
  850: string;
  900: string;
}

export interface SurfaceColors {
  primary: string;
  secondary: string;
  tertiary: string;
  inverse: string;
  overlay: string;
}

export interface TextColors {
  primary: string;
  secondary: string;
  tertiary: string;
  inverse: string;
  placeholder: string;
  disabled: string;
}

export interface BorderColors {
  primary: string;
  secondary: string;
  focus: string;
  inverse: string;
}

export interface InteractiveColors {
  primary: string;
  primaryHover: string;
  primaryActive: string;
  secondary: string;
  secondaryHover: string;
  secondaryActive: string;
}

// =============================================================================
// TYPOGRAPHY TYPES
// =============================================================================

export interface FontFamily {
  sans: string;
  serif: string;
  mono: string;
}

export interface FontSize {
  xs: string;
  sm: string;
  base: string;
  lg: string;
  xl: string;
  '2xl': string;
  '3xl': string;
  '4xl': string;
  '5xl': string;
}

export interface FontWeight {
  thin: number;
  light: number;
  normal: number;
  medium: number;
  semibold: number;
  bold: number;
  extrabold: number;
}

export interface LineHeight {
  none: number;
  tight: number;
  snug: number;
  normal: number;
  relaxed: number;
  loose: number;
}

export interface Typography {
  fontFamily: FontFamily;
  fontSize: FontSize;
  fontWeight: FontWeight;
  lineHeight: LineHeight;
}

// =============================================================================
// SPACING & LAYOUT TYPES
// =============================================================================

export interface Spacing {
  0: string;
  1: string;
  2: string;
  3: string;
  4: string;
  5: string;
  6: string;
  8: string;
  10: string;
  12: string;
  16: string;
  20: string;
  24: string;
  32: string;
  40: string;
  48: string;
  56: string;
  64: string;
}

export interface BorderRadius {
  none: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
  '2xl': string;
  '3xl': string;
  full: string;
}

export interface Shadow {
  none: string;
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
  '2xl': string;
  inner: string;
}

// =============================================================================
// THEME CONFIGURATION
// =============================================================================

export type ThemeMode = 'light' | 'dark' | 'system';

export interface ThemeColors {
  brand: BrandColors;
  semantic: SemanticColors;
  neutral: NeutralColors;
  surface: SurfaceColors;
  text: TextColors;
  border: BorderColors;
  interactive: InteractiveColors;
}

export interface ThemeConfig {
  id: string;
  name: string;
  description?: string;
  mode: ThemeMode;
  colors: ThemeColors;
  typography: Typography;
  spacing: Spacing;
  borderRadius: BorderRadius;
  shadows: Shadow;
  createdAt: Date;
  updatedAt: Date;
  isDefault?: boolean;
  isSystem?: boolean;
}

export interface ThemePresetConfig {
  colors?: {
    brand?: Partial<BrandColors>;
    semantic?: Partial<SemanticColors>;
    neutral?: Partial<NeutralColors>;
    surface?: Partial<SurfaceColors>;
    text?: Partial<TextColors>;
    border?: Partial<BorderColors>;
    interactive?: Partial<InteractiveColors>;
  };
  typography?: Partial<Typography>;
  spacing?: Partial<Spacing>;
  borderRadius?: Partial<BorderRadius>;
  shadows?: Partial<Shadow>;
}

export interface ThemePreset {
  id: string;
  name: string;
  description: string;
  preview: {
    primary: string;
    secondary: string;
    accent: string;
  };
  config: ThemePresetConfig;
}

// =============================================================================
// THEME STORE STATE
// =============================================================================

export interface ThemeState {
  currentThemeId: string;
  mode: ThemeMode;
  themes: ThemeConfig[];
  presets: ThemePreset[];
  customCss: string;
  isLoading: boolean;
  error: string | null;
}

// =============================================================================
// THEME BUILDER STATE
// =============================================================================

export interface ThemeBuilderState {
  activeTab: 'colors' | 'typography' | 'spacing' | 'shadows' | 'preview';
  editingTheme: ThemeConfig | null;
  isDirty: boolean;
  previewMode: 'light' | 'dark' | 'split';
  colorPickerOpen: boolean;
  selectedColorPath: string | null;
}

// =============================================================================
// DEFAULT VALUES
// =============================================================================

export const DEFAULT_COLOR_SCALE: ColorScale = {
  50: '#f0f9ff',
  100: '#e0f2fe',
  200: '#bae6fd',
  300: '#7dd3fc',
  400: '#38bdf8',
  500: '#0ea5e9',
  600: '#0284c7',
  700: '#0369a1',
  800: '#075985',
  900: '#0c4a6e',
};

export const DEFAULT_NEUTRAL: NeutralColors = {
  white: '#ffffff',
  black: '#000000',
  25: '#fcfcfc',
  50: '#fafafa',
  100: '#f5f5f5',
  200: '#e5e5e5',
  300: '#d4d4d4',
  400: '#a3a3a3',
  500: '#737373',
  600: '#525252',
  700: '#404040',
  800: '#262626',
  850: '#1c1c1c',
  900: '#171717',
};

export const DEFAULT_TYPOGRAPHY: Typography = {
  fontFamily: {
    sans: 'Inter, system-ui, -apple-system, sans-serif',
    serif: 'Georgia, Cambria, serif',
    mono: 'Roboto Mono, Menlo, Monaco, monospace',
  },
  fontSize: {
    xs: '0.75rem',
    sm: '0.875rem',
    base: '1rem',
    lg: '1.125rem',
    xl: '1.25rem',
    '2xl': '1.5rem',
    '3xl': '1.875rem',
    '4xl': '2.25rem',
    '5xl': '3rem',
  },
  fontWeight: {
    thin: 100,
    light: 300,
    normal: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },
  lineHeight: {
    none: 1,
    tight: 1.25,
    snug: 1.375,
    normal: 1.5,
    relaxed: 1.625,
    loose: 2,
  },
};

export const DEFAULT_SPACING: Spacing = {
  0: '0',
  1: '0.25rem',
  2: '0.5rem',
  3: '0.75rem',
  4: '1rem',
  5: '1.25rem',
  6: '1.5rem',
  8: '2rem',
  10: '2.5rem',
  12: '3rem',
  16: '4rem',
  20: '5rem',
  24: '6rem',
  32: '8rem',
  40: '10rem',
  48: '12rem',
  56: '14rem',
  64: '16rem',
};

export const DEFAULT_BORDER_RADIUS: BorderRadius = {
  none: '0',
  sm: '0.125rem',
  md: '0.375rem',
  lg: '0.5rem',
  xl: '0.75rem',
  '2xl': '1rem',
  '3xl': '1.5rem',
  full: '9999px',
};

export const DEFAULT_SHADOWS: Shadow = {
  none: 'none',
  xs: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  sm: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.25)',
  inner: 'inset 0 2px 4px 0 rgb(0 0 0 / 0.05)',
};
