export type ThemeMode = 'light' | 'dark' | 'auto'

export declare class ThemeManager {
  constructor()
  initializeTheme(): void
  setupSystemThemeListener(): void
  applyTheme(theme: ThemeMode): void
  notifyThemeChange(): void
  setTheme(theme: ThemeMode, save?: boolean): void
  getTheme(): ThemeMode
  getResolvedTheme(): 'light' | 'dark'
  getSystemPreference(): 'light' | 'dark'
  toggleTheme(): void
  isDark(): boolean
  getCSSVar(property: string): string
}

export declare const themeManager: ThemeManager

export declare const getThemeValue: (tokenPath: string) => string
export declare const isSystemDark: () => boolean
