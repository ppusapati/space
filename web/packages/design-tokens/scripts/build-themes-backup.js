#!/usr/bin/env node

import StyleDictionary from 'style-dictionary'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'
import fs from 'fs'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

// Register custom formats
StyleDictionary.registerFormat({
  name: 'css/variables-no-keyframes',
  format: function({ dictionary }) {
    const tokens = dictionary.allTokens
      .filter(token => token.type !== 'keyframes')
      .map(token => `  --${token.name}: ${token.value};`)
      .join('\n')
    
    return `:root {\n${tokens}\n}`
  }
})

StyleDictionary.registerFormat({
  name: 'css/keyframes',
  format: function({ dictionary }) {
    const keyframeTokens = dictionary.allTokens
      .filter(token => token.type === 'keyframes')
    
    if (keyframeTokens.length === 0) return ''
    
    const keyframes = keyframeTokens.map(token => {
      const name = token.name.replace('animation-keyframes-', '')
      let keyframeCSS = ''
      
      if (Array.isArray(token.original.value)) {
        keyframeCSS = token.original.value.map(frame => {
          const { offset, ...properties } = frame
          const percentage = Math.round(offset * 100)
          const props = Object.entries(properties)
            .map(([key, value]) => `    ${key.replace(/([A-Z])/g, '-$1').toLowerCase()}: ${value};`)
            .join('\n')
          return `  ${percentage}% {\n${props}\n  }`
        }).join('\n')
      }
      
      return `@keyframes ${name} {\n${keyframeCSS}\n}`
    }).join('\n\n')
    
    return keyframes
  }
})

StyleDictionary.registerFormat({
  name: 'custom/uno-theme',
  format: function({ dictionary }) {
    const colorTokens = dictionary.allTokens
      .filter(token => token.type === 'color')
      .map(token => {
        const path = token.path.join('-')
        return `    '${path}': '${token.value}'`
      })
      .join(',\n')
    
    const spacingTokens = dictionary.allTokens
      .filter(token => token.path[0] === 'spacing')
      .map(token => {
        const key = token.path[1]
        return `    '${key}': '${token.value}'`
      })
      .join(',\n')
    
    const fontSizeTokens = dictionary.allTokens
      .filter(token => token.path[0] === 'typography' && token.path[1] === 'fontSize')
      .map(token => {
        const key = token.path[2]
        return `    '${key}': '${token.value}'`
      })
      .join(',\n')
    
    const fontFamilyTokens = dictionary.allTokens
      .filter(token => token.path[0] === 'typography' && token.path[1] === 'fontFamily')
      .map(token => {
        const key = token.path[2]
        const families = token.value.split(',').map(f => f.trim().replace(/['"]/g, ''))
        return `    '${key}': ${JSON.stringify(families)}`
      })
      .join(',\n')
    
    const borderRadiusTokens = dictionary.allTokens
      .filter(token => token.path[0] === 'borderRadius')
      .map(token => {
        const key = token.path[1]
        return `    '${key}': '${token.value}'`
      })
      .join(',\n')
    
    const shadowTokens = dictionary.allTokens
      .filter(token => token.path[0] === 'shadow')
      .map(token => {
        const key = token.path[1]
        return `    '${key}': '${token.value}'`
      })
      .join(',\n')

    const animationUtilities = `
// Animation utilities
export const animations = {
  // Durations
  ${dictionary.allTokens
    .filter(token => token.path[0] === 'animation' && token.path[1] === 'duration')
    .map(token => `${token.path[2]}: '${token.value}'`)
    .join(',\n  ')},
  
  // Easings  
  ${dictionary.allTokens
    .filter(token => token.path[0] === 'animation' && token.path[1] === 'easing')
    .map(token => `${token.path[2]}: '${token.value}'`)
    .join(',\n  ')},
  
  // Keyframes
  ${dictionary.allTokens
    .filter(token => token.path[0] === 'animation' && token.path[1] === 'keyframes')
    .map(token => `'${token.path[2]}': '${token.path[2]}'`)
    .join(',\n  ')}
}`

    return `import type { Theme } from '@unocss/preset-uno'

export const designTokensTheme: Theme = {
  colors: {
${colorTokens}
  },
  spacing: {
${spacingTokens}
  },
  fontSize: {
${fontSizeTokens}
  },
  fontFamily: {
${fontFamilyTokens}
  },
  borderRadius: {
${borderRadiusTokens}
  },
  boxShadow: {
${shadowTokens}
  }
}

${animationUtilities}

export const componentShortcuts = {
  // Base component shortcuts
  'btn-primary': 'bg-color-brand-primary-500 text-color-neutral-white border-color-brand-primary-500 rounded-borderRadius-md px-spacing-4 py-spacing-2 text-typography-fontSize-sm font-medium shadow-shadow-sm transition-all duration-150 ease-out hover:bg-color-brand-primary-600 hover:border-color-brand-primary-600 hover:-translate-y-px active:bg-color-brand-primary-700 active:border-color-brand-primary-700 active:translate-y-0 disabled:bg-color-neutral-300 disabled:border-color-neutral-300 disabled:text-color-neutral-500 disabled:cursor-not-allowed disabled:opacity-50',
  
  'btn-secondary': 'bg-color-neutral-white text-color-brand-primary-600 border-color-neutral-300 rounded-borderRadius-md px-spacing-4 py-spacing-2 text-typography-fontSize-sm font-medium shadow-shadow-sm transition-all duration-150 ease-out hover:bg-color-neutral-50 hover:border-color-brand-primary-300 hover:-translate-y-px',
  
  'input-base': 'bg-color-neutral-white border-color-neutral-300 rounded-borderRadius-md border text-color-neutral-900 text-typography-fontSize-sm px-spacing-3 py-spacing-2 h-40 transition-all duration-150 ease-out placeholder:text-color-neutral-500 focus:border-color-brand-primary-500 focus:outline-none',
  
  'card-base': 'bg-color-neutral-white border-color-neutral-200 rounded-borderRadius-lg border shadow-shadow-sm p-spacing-6 transition-all duration-300 ease-out hover:shadow-shadow-md hover:-translate-y-0.5',
  
  // ERP specific
  'data-grid-header': 'bg-color-neutral-100 text-color-neutral-900 font-semibold text-typography-fontSize-sm border-b border-color-neutral-200 px-spacing-4 py-spacing-3',
  'data-grid-cell': 'text-color-neutral-800 text-typography-fontSize-sm px-spacing-4 py-spacing-3 border-b border-color-neutral-100',
  
  // SCADA specific
  'status-indicator-active': 'bg-color-semantic-success-500 w-spacing-3 h-spacing-3 rounded-full',
  'status-indicator-inactive': 'bg-color-neutral-400 w-spacing-3 h-spacing-3 rounded-full',
  'alarm-critical': 'bg-color-semantic-error-50 border-color-semantic-error-500 text-color-semantic-error-900 border-l-4 px-spacing-4 py-spacing-3'
}

export default designTokensTheme
`
  }
})

async function buildTokens() {
  console.log('🎨 Building design tokens...')
  
  // Ensure dist directories exist
  const distDirs = ['dist/css', 'dist/js', 'dist/uno']
  distDirs.forEach(dir => {
    const fullPath = join(rootDir, dir)
    if (!fs.existsSync(fullPath)) {
      fs.mkdirSync(fullPath, { recursive: true })
    }
  })

  // Check if token directories exist and have JSON files
  const requiredTokenDirs = [
    { path: 'tokens/global', required: true, description: 'Global base tokens' },
    { path: 'tokens/themes', required: true, description: 'Theme tokens' }
  ]
  
  const optionalTokenDirs = [
    { path: 'tokens/layout', required: false, description: 'Layout tokens' },
    { path: 'tokens/components', required: false, description: 'Component tokens' }
  ]
  
  let hasAllRequired = true
  
  // Check required directories
  for (const { path, description } of requiredTokenDirs) {
    const fullPath = join(rootDir, path)
    
    if (!fs.existsSync(fullPath)) {
      console.error(`❌ Missing required directory: ${path} (${description})`)
      hasAllRequired = false
      continue
    }
    
    // Check for JSON files in the directory
    try {
      const files = fs.readdirSync(fullPath)
      const jsonFiles = files.filter(f => f.endsWith('.json'))
      
      if (jsonFiles.length === 0) {
        console.error(`❌ Directory ${path} exists but contains no JSON files`)
        hasAllRequired = false
      } else {
        console.log(`✅ Found ${jsonFiles.length} JSON files in ${path}: ${jsonFiles.join(', ')}`)
      }
    } catch (error) {
      console.error(`❌ Error reading directory ${path}: ${error.message}`)
      hasAllRequired = false
    }
  }
  
  // Check optional directories (just warn, don't fail)
  for (const { path, description } of optionalTokenDirs) {
    const fullPath = join(rootDir, path)
    
    if (fs.existsSync(fullPath)) {
      try {
        const files = fs.readdirSync(fullPath)
        const jsonFiles = files.filter(f => f.endsWith('.json'))
        
        if (jsonFiles.length > 0) {
          console.log(`✅ Found ${jsonFiles.length} JSON files in ${path}: ${jsonFiles.join(', ')}`)
        } else {
          console.log(`⚠️  Optional directory ${path} exists but contains no JSON files`)
        }
      } catch (error) {
        console.warn(`⚠️  Error reading optional directory ${path}: ${error.message}`)
      }
    } else {
      console.log(`ℹ️  Optional directory ${path} not found (${description})`)
    }
  }
  
  if (!hasAllRequired) {
    console.log('\n🚀 Quick setup: Run "npm run setup" to create missing token files')
    console.log('\nRequired structure:')
    console.log(`
tokens/
├── global/           ← Base design tokens
│   ├── colors.json
│   ├── spacing.json
│   ├── typography.json
│   ├── shadows.json
│   └── animations.json
└── themes/           ← Theme-specific tokens  
    ├── light.json
    └── dark.json
    `)
    return
  }

  try {
    // Build light theme
    console.log('🌟 Building light theme...')
    
    // Dynamically build source patterns based on what exists
    const sourcePatterns = [
      'tokens/global/**/*.json',
      'tokens/themes/light.json'
    ]
    
    // Add optional directories if they exist and have JSON files
    const optionalDirs = ['tokens/layout', 'tokens/components']
    optionalDirs.forEach(dir => {
      const fullPath = join(rootDir, dir)
      if (fs.existsSync(fullPath)) {
        try {
          const files = fs.readdirSync(fullPath).filter(f => f.endsWith('.json'))
          if (files.length > 0) {
            sourcePatterns.push(`${dir}/**/*.json`)
            console.log(`📁 Including optional directory: ${dir}`)
          }
        } catch (error) {
          console.warn(`⚠️  Skipping ${dir}: ${error.message}`)
        }
      }
    })
    
    console.log(`📋 Source patterns: ${sourcePatterns.join(', ')}`)
    
    const lightConfig = {
      source: sourcePatterns,
      platforms: {
        css: {
          transformGroup: 'css',
          buildPath: 'dist/css/',
          files: [{
            destination: 'tokens-light.css',
            format: 'css/variables-no-keyframes'
          }]
        },
        'css-keyframes': {
          transformGroup: 'css',
          buildPath: 'dist/css/',
          files: [{
            destination: 'keyframes.css',
            format: 'css/keyframes'
          }]
        },
        js: {
          transformGroup: 'js',
          buildPath: 'dist/js/',
          files: [{
            destination: 'tokens.js',
            format: 'javascript/es6'
          }]
        },
        uno: {
          transformGroup: 'js',
          buildPath: 'dist/uno/',
          files: [{
            destination: 'theme.ts',
            format: 'custom/uno-theme'
          }]
        }
      }
    }

    const lightSD = new StyleDictionary(lightConfig)
    await lightSD.buildAllPlatforms()

    // Build dark theme
    console.log('🌙 Building dark theme...')
    
    // Build dark theme source patterns
    const darkSourcePatterns = [
      'tokens/global/**/*.json',
      'tokens/themes/dark.json'
    ]
    
    // Add optional directories for dark theme too
    optionalDirs.forEach(dir => {
      const fullPath = join(rootDir, dir)
      if (fs.existsSync(fullPath)) {
        try {
          const files = fs.readdirSync(fullPath).filter(f => f.endsWith('.json'))
          if (files.length > 0) {
            darkSourcePatterns.push(`${dir}/**/*.json`)
          }
        } catch (error) {
          // Already warned above
        }
      }
    })
    
    const darkConfig = {
      source: darkSourcePatterns,
      platforms: {
        css: {
          transformGroup: 'css',
          buildPath: 'dist/css/',
          files: [{
            destination: 'tokens-dark.css',
            format: 'css/variables-no-keyframes'
          }]
        }
      }
    }

    const darkSD = new StyleDictionary(darkConfig)
    await darkSD.buildAllPlatforms()

    // Create combined CSS file
    console.log('🔗 Creating combined CSS...')
    const lightCSSPath = join(rootDir, 'dist/css/tokens-light.css')
    const darkCSSPath = join(rootDir, 'dist/css/tokens-dark.css')
    const keyframesPath = join(rootDir, 'dist/css/keyframes.css')

    if (fs.existsSync(lightCSSPath) && fs.existsSync(darkCSSPath)) {
      const lightCSS = fs.readFileSync(lightCSSPath, 'utf8')
      const darkCSS = fs.readFileSync(darkCSSPath, 'utf8')
      const keyframesCSS = fs.existsSync(keyframesPath) ? fs.readFileSync(keyframesPath, 'utf8') : ''

      const combinedCSS = `/* 
 * Enterprise Design Tokens - Light & Dark Themes
 * Automatically switches based on system preference or data-theme attribute
 */

/* Animation keyframes (theme-independent) */
${keyframesCSS}

/* Light theme (default) */
${lightCSS}

/* Dark theme via data attribute */
[data-theme="dark"] {
${darkCSS.replace(':root {', '').replace(/}([^}]*)$/, '')}
}

/* Auto dark theme based on system preference */
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
${darkCSS.replace(':root {', '').replace(/}([^}]*)$/, '')}
  }
}

/* Smooth theme transitions */
*,
*::before,
*::after {
  transition: background-color 0.3s ease, color 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
}

/* Disable transitions during theme switching */
[data-theme-switching] *,
[data-theme-switching] *::before,
[data-theme-switching] *::after {
  transition: none !important;
}
`

      fs.writeFileSync(join(rootDir, 'dist/css/tokens.css'), combinedCSS)
      console.log('✅ Created tokens.css (combined)')
    }

    // Create UnoCSS theme configuration
    console.log('🎯 Creating UnoCSS theme configuration...')
    const themeFilePath = join(rootDir, 'dist/uno/theme.ts')
    
    if (fs.existsSync(themeFilePath)) {
      const combinedUnoTheme = `import type { Theme } from '@unocss/preset-uno'

// Import theme and shortcuts
import { designTokensTheme, componentShortcuts } from './theme'

export { designTokensTheme, componentShortcuts }

// Theme configuration for UnoCSS with CSS variables (auto-switching)
export const themeConfig: Theme = {
  colors: {
    // Use CSS variables so they automatically switch with the CSS
    ...Object.keys(designTokensTheme.colors || {}).reduce((acc, key) => {
      acc[key] = \`var(--\${key.replace(/([A-Z])/g, '-$1').toLowerCase()})\`
      return acc
    }, {} as Record<string, string>)
  },
  spacing: designTokensTheme.spacing,
  fontSize: designTokensTheme.fontSize,
  fontFamily: designTokensTheme.fontFamily,
  borderRadius: designTokensTheme.borderRadius,
  boxShadow: {
    // Use CSS variables for shadows too
    ...Object.keys(designTokensTheme.boxShadow || {}).reduce((acc, key) => {
      acc[key] = \`var(--shadow-\${key})\`
      return acc
    }, {} as Record<string, string>)
  }
}

export default themeConfig
`

      fs.writeFileSync(join(rootDir, 'dist/uno/index.ts'), combinedUnoTheme)
      console.log('✅ Created index.ts (UnoCSS config)')
    }

    // Create theme utilities
    console.log('🛠️ Creating theme utilities...')
    const themeUtils = `/**
 * Enterprise Theme Manager for MFE Applications
 */

export type ThemeMode = 'light' | 'dark' | 'auto'

export class ThemeManager {
  private currentTheme: ThemeMode = 'auto'
  
  constructor() {
    this.initializeTheme()
    this.setupSystemThemeListener()
  }
  
  private initializeTheme() {
    const savedTheme = localStorage.getItem('theme') as ThemeMode
    if (savedTheme && ['light', 'dark', 'auto'].includes(savedTheme)) {
      this.setTheme(savedTheme, false)
    } else {
      this.setTheme('auto', false)
    }
  }
  
  private setupSystemThemeListener() {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      if (this.currentTheme === 'auto') {
        this.applyTheme('auto')
        this.notifyThemeChange()
      }
    })
  }
  
  private applyTheme(theme: ThemeMode) {
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
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        root.setAttribute('data-theme', systemDark ? 'dark' : 'light')
        break
    }
  }
  
  private notifyThemeChange() {
    window.dispatchEvent(new CustomEvent('themechange', { 
      detail: { 
        theme: this.currentTheme, 
        resolvedTheme: this.getResolvedTheme(),
        systemPreference: this.getSystemPreference()
      }
    }))
  }
  
  setTheme(theme: ThemeMode, save = true) {
    this.currentTheme = theme
    this.applyTheme(theme)
    
    if (save) {
      localStorage.setItem('theme', theme)
    }
    
    this.notifyThemeChange()
  }
  
  getTheme() { return this.currentTheme }
  
  getResolvedTheme(): 'light' | 'dark' {
    return this.currentTheme === 'auto' ? this.getSystemPreference() : this.currentTheme
  }
  
  getSystemPreference(): 'light' | 'dark' {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  
  toggleTheme() {
    const current = this.getResolvedTheme()
    this.setTheme(current === 'light' ? 'dark' : 'light')
  }
  
  isDark() { return this.getResolvedTheme() === 'dark' }
  
  getCSSVar(property: string): string {
    return getComputedStyle(document.documentElement).getPropertyValue(property).trim()
  }
}

export const themeManager = new ThemeManager()

// Export utility functions
export const getThemeValue = (tokenPath: string): string => {
  return themeManager.getCSSVar(\`--\${tokenPath.replace(/\\./g, '-')}\`)
}

export const isSystemDark = (): boolean => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}
`

    fs.writeFileSync(join(rootDir, 'dist/js/theme-utils.js'), themeUtils)
    console.log('✅ Created theme-utils.js')

    // Generate TypeScript declarations
    console.log('📝 Generating TypeScript declarations...')
    try {
      const { execSync } = await import('child_process')
      execSync('node scripts/type-definition.js', { cwd: rootDir, stdio: 'inherit' })
      console.log('✅ Generated TypeScript declarations')
    } catch (error) {
      console.log('⚠️  Could not generate TypeScript declarations automatically')
      console.log('💡 Run manually: npm run types:generate')
    }

    console.log('✅ Design tokens built successfully!')
    console.log(`
📦 Generated files:
├── dist/css/
│   ├── tokens.css (🎯 MAIN FILE - combined themes + animations)
│   ├── tokens-light.css (light theme only)  
│   ├── tokens-dark.css (dark theme only)
│   └── keyframes.css (CSS animations)
├── dist/js/
│   ├── tokens.js (JavaScript tokens)
│   └── theme-utils.js (theme switching utilities)
└── dist/uno/
    ├── index.ts (🎯 MAIN FILE - UnoCSS config)
    └── theme.ts (light UnoCSS theme + shortcuts)

🎨 Ready for your MFEs! Use tokens.css and index.ts
`)

  } catch (error) {
    console.error('❌ Build failed:', error.message)
    console.log('\n🔍 Troubleshooting tips:')
    console.log('1. Check that all token files exist in tokens/ directory')
    console.log('2. Verify JSON syntax in all .json files')
    console.log('3. Ensure token references like {color.brand.primary.500} are valid')
    console.log('4. Check that semantic token files reference the correct base tokens')
  }
}

buildTokens().catch(console.error)