/**
 * Theme Builder Types
 * Defines the structure for customizable themes
 */
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
export interface ThemeState {
    currentThemeId: string;
    mode: ThemeMode;
    themes: ThemeConfig[];
    presets: ThemePreset[];
    customCss: string;
    isLoading: boolean;
    error: string | null;
}
export interface ThemeBuilderState {
    activeTab: 'colors' | 'typography' | 'spacing' | 'shadows' | 'preview';
    editingTheme: ThemeConfig | null;
    isDirty: boolean;
    previewMode: 'light' | 'dark' | 'split';
    colorPickerOpen: boolean;
    selectedColorPath: string | null;
}
export declare const DEFAULT_COLOR_SCALE: ColorScale;
export declare const DEFAULT_NEUTRAL: NeutralColors;
export declare const DEFAULT_TYPOGRAPHY: Typography;
export declare const DEFAULT_SPACING: Spacing;
export declare const DEFAULT_BORDER_RADIUS: BorderRadius;
export declare const DEFAULT_SHADOWS: Shadow;
//# sourceMappingURL=theme.types.d.ts.map