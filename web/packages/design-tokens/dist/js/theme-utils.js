/**
 * Enterprise Theme Manager for MFE Applications (JS version)
 * @typedef {'light'|'dark'|'auto'} ThemeMode
 */
export class ThemeManager {
  constructor() {
    this.currentTheme = 'auto'
    this.initializeTheme()
    this.setupSystemThemeListener()
  }

  initializeTheme() {
    const savedTheme = /** @type {ThemeMode|null} */ (localStorage.getItem('theme'))
    if (savedTheme && ['light', 'dark', 'auto'].includes(savedTheme)) {
      this.setTheme(savedTheme, false)
    } else {
      this.setTheme('auto', false)
    }
  }

  setupSystemThemeListener() {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      if (this.currentTheme === 'auto') {
        this.applyTheme('auto')
        this.notifyThemeChange()
      }
    })
  }

  applyTheme(theme) {
    const root = document.documentElement
    // Add transitioning class to prevent flash
    root.setAttribute('data-theme-switching', '')
    requestAnimationFrame(() => {
      root.removeAttribute('data-theme-switching')
    })

    root.removeAttribute('data-theme')

    switch (theme) {
      case 'light':
        root.setAttribute('data-theme', 'light')
        break
      case 'dark':
        root.setAttribute('data-theme', 'dark')
        break
      case 'auto':
        {
          const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
          root.setAttribute('data-theme', systemDark ? 'dark' : 'light')
        }
        break
    }
  }

  notifyThemeChange() {
    window.dispatchEvent(new CustomEvent('themechange', {
      detail: {
        theme: this.currentTheme,
        resolvedTheme: this.getResolvedTheme(),
        systemPreference: this.getSystemPreference()
      }
    }))
  }

  setTheme(theme, save = true) {
    this.currentTheme = theme
    this.applyTheme(theme)
    if (save) localStorage.setItem('theme', theme)
    this.notifyThemeChange()
  }

  getTheme() { return this.currentTheme }

  getResolvedTheme() {
    return this.currentTheme === 'auto' ? this.getSystemPreference() : this.currentTheme
  }

  getSystemPreference() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }

  toggleTheme() {
    const current = this.getResolvedTheme()
    this.setTheme(current === 'light' ? 'dark' : 'light')
  }

  isDark() { return this.getResolvedTheme() === 'dark' }

  getCSSVar(property) {
    return getComputedStyle(document.documentElement).getPropertyValue(property).trim()
  }
}

export const themeManager = new ThemeManager()

// Export utility functions
export const getThemeValue = (tokenPath) => {
  return themeManager.getCSSVar(`--${tokenPath.replace(/\./g, '-')}`)
}

export const isSystemDark = () => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}
