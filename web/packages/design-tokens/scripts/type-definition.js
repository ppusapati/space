#!/usr/bin/env node

import fs from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

function generateTypeScriptDeclarations() {
  console.log('📝 Generating TypeScript declarations...\n')
  
  // Ensure dist/types directory exists
  const typesDir = join(rootDir, 'dist/types')
  if (!fs.existsSync(typesDir)) {
    fs.mkdirSync(typesDir, { recursive: true })
  }
  
  // Collect all tokens from light theme (use as base structure)
  const lightTokensPath = join(rootDir, 'tokens/themes/light.json')
  const globalTokenPaths = [
    'tokens/global/colors.json',
    'tokens/global/spacing.json', 
    'tokens/global/typography.json',
    'tokens/global/shadows.json',
    'tokens/global/animations.json',
    'tokens/global/z-index.json',
    'tokens/global/border-width.json',
    'tokens/global/transitions.json',
    'tokens/layout/grid.json',
    'tokens/layout/flex.json',
    'tokens/layout/container.json',
    'tokens/components/button.json',
    'tokens/components/card.json',
    'tokens/components/modal.json',
    'tokens/components/input.json',
  ]
  
  let allTokens = {}
  
  // Collect global tokens
  globalTokenPaths.forEach(tokenPath => {
    const fullPath = join(rootDir, tokenPath)
    if (fs.existsSync(fullPath)) {
      try {
        const content = fs.readFileSync(fullPath, 'utf8')
        const tokens = JSON.parse(content)
        allTokens = { ...allTokens, ...tokens }
        console.log(`✅ Processed ${tokenPath}`)
      } catch (error) {
        console.log(`⚠️  Could not process ${tokenPath}: ${error.message}`)
      }
    }
  })
  
  // Add theme tokens (light as base structure)
  if (fs.existsSync(lightTokensPath)) {
    try {
      const content = fs.readFileSync(lightTokensPath, 'utf8')
      const themeTokens = JSON.parse(content)
      allTokens = { ...allTokens, ...themeTokens }
      console.log(`✅ Processed theme tokens`)
    } catch (error) {
      console.log(`⚠️  Could not process theme tokens: ${error.message}`)
    }
  }
  
  // Generate TypeScript interfaces
  const tokenTypes = generateTokenInterfaces(allTokens)
  
  // Generate main tokens declaration
  const mainDeclaration = generateMainDeclaration()
  
  // Generate theme utilities declaration
  const themeUtilsDeclaration = generateThemeUtilsDeclaration()
  
  // Generate UnoCSS theme declaration
  const unoThemeDeclaration = generateUnoThemeDeclaration()
  
  // Write declaration files
  fs.writeFileSync(join(typesDir, 'tokens.d.ts'), tokenTypes)
  console.log('📄 Generated tokens.d.ts')
  
  fs.writeFileSync(join(typesDir, 'index.d.ts'), mainDeclaration)
  console.log('📄 Generated index.d.ts')
  
  fs.writeFileSync(join(typesDir, 'theme-utils.d.ts'), themeUtilsDeclaration)
  console.log('📄 Generated theme-utils.d.ts')
  
  fs.writeFileSync(join(typesDir, 'uno-theme.d.ts'), unoThemeDeclaration)
  console.log('📄 Generated uno-theme.d.ts')
  
  // Generate package.json type exports
  updatePackageJsonTypes()
  
  console.log('\n🎉 TypeScript declarations generated successfully!')
  console.log('\n📋 Generated files:')
  console.log('   • dist/types/tokens.d.ts - Token interfaces')
  console.log('   • dist/types/index.d.ts - Main export types')
  console.log('   • dist/types/theme-utils.d.ts - Theme manager types')
  console.log('   • dist/types/uno-theme.d.ts - UnoCSS theme types')
  
  console.log('\n🚀 Usage in TypeScript:')
  console.log('   import { tokens } from "@p9e/design-tokens"')
  console.log('   import type { DesignTokens } from "@p9e/design-tokens/types"')
  console.log('   import { themeManager } from "@p9e/design-tokens/js/theme-utils"')
}

function generateTokenInterfaces(tokens, prefix = '', depth = 0) {
  let types = ''
  
  if (depth === 0) {
    types += `// Auto-generated TypeScript declarations for design tokens
// Generated on ${new Date().toISOString()}

export interface DesignTokens {\n`
  }
  
  for (const [key, value] of Object.entries(tokens)) {
    const currentPath = prefix ? `${prefix}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token with a value
        const tokenType = getTokenTypeForTS(value.type, value.value)
        const indent = '  '.repeat(depth + 1)
        types += `${indent}${key}: ${tokenType};\n`
      } else {
        // This is a nested object
        const indent = '  '.repeat(depth + 1)
        types += `${indent}${key}: {\n`
        types += generateTokenInterfaces(value, currentPath, depth + 1)
        types += `${indent}};\n`
      }
    }
  }
  
  if (depth === 0) {
    types += '}\n\n'
    
    // Add helper types
    types += `// Helper types for token paths
export type TokenPath = string;
export type TokenValue = string | number;

// Theme types
export type ThemeMode = 'light' | 'dark' | 'auto';

// Token categories
export type ColorToken = string;
export type SpacingToken = string;
export type TypographyToken = string;
export type ShadowToken = string;
`
  }
  
  return types
}

function getTokenTypeForTS(tokenType, value) {
  switch (tokenType) {
    case 'color':
      return 'string'
    case 'dimension':
    case 'spacing':
    case 'fontSize':
    case 'borderRadius':
      return 'string'
    case 'number':
      return 'number'
    case 'shadow':
      return 'string'
    case 'fontFamily':
      return 'string'
    case 'fontWeight':
      return 'string | number'
    case 'duration':
    case 'cubicBezier':
      return 'string'
    case 'keyframes':
      return 'string'
    default:
      // Infer from value
      if (typeof value === 'number') return 'number'
      return 'string'
  }
}

function generateMainDeclaration() {
  return `// Main export declarations
import type { DesignTokens } from './tokens';

declare const tokens: DesignTokens;

export { tokens };
export type { DesignTokens } from './tokens';
export type { ThemeMode, TokenPath, TokenValue } from './tokens';

// CSS export
export declare const css: string;

// Default export
declare const _default: DesignTokens;
export default _default;
`
}

function generateThemeUtilsDeclaration() {
  return `// Theme utilities declarations
export type ThemeMode = 'light' | 'dark' | 'auto';

export interface ThemeChangeEvent {
  theme: ThemeMode;
  resolvedTheme: 'light' | 'dark';
  systemPreference: 'light' | 'dark';
}

export declare class ThemeManager {
  constructor();
  
  setTheme(theme: ThemeMode, save?: boolean): void;
  getTheme(): ThemeMode;
  getResolvedTheme(): 'light' | 'dark';
  getSystemPreference(): 'light' | 'dark';
  toggleTheme(): void;
  isDark(): boolean;
  getCSSVar(property: string): string;
}

export declare const themeManager: ThemeManager;

export declare function getThemeValue(tokenPath: string): string;
export declare function isSystemDark(): boolean;

// React hook (if React is available)
export interface UseThemeReturn {
  theme: ThemeMode;
  resolvedTheme: 'light' | 'dark';
  systemPreference: 'light' | 'dark';
  isDark: boolean;
  setTheme: (theme: ThemeMode) => void;
  toggleTheme: () => void;
  getCSSVar: (property: string) => string;
}

export declare function useTheme(): UseThemeReturn;
`
}

function generateUnoThemeDeclaration() {
  return `// UnoCSS theme declarations
import type { Theme } from '@unocss/preset-uno';

export declare const designTokensTheme: Theme;
export declare const themeConfig: Theme;
export declare const componentShortcuts: Record<string, string>;
export declare const animationShortcuts: Record<string, string>;

export interface AnimationUtilities {
  [key: string]: string;
}

export declare const animations: AnimationUtilities;

export default designTokensTheme;
`
}

function updatePackageJsonTypes() {
  const packageJsonPath = join(rootDir, 'package.json')
  
  if (fs.existsSync(packageJsonPath)) {
    try {
      const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'))
      
      // Update exports with types
      packageJson.exports = {
        ".": {
          "import": "./dist/js/tokens.js",
          "types": "./dist/types/index.d.ts"
        },
        "./types": "./dist/types/tokens.d.ts",
        "./css": "./dist/css/tokens.css",
        "./css/light": "./dist/css/tokens-light.css", 
        "./css/dark": "./dist/css/tokens-dark.css",
        "./css/keyframes": "./dist/css/keyframes.css",
        "./js": "./dist/js/tokens.js",
        "./js/theme-utils": {
          "import": "./dist/js/theme-utils.js",  
          "types": "./dist/types/theme-utils.d.ts"
        },
        "./uno": {
          "import": "./dist/uno/index.ts",
          "types": "./dist/types/uno-theme.d.ts"
        },
        "./uno/theme": "./dist/uno/theme.ts"
      }
      
      // Update main types field
      packageJson.types = "./dist/types/index.d.ts"
      
      fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2))
      console.log('📦 Updated package.json with TypeScript exports')
      
    } catch (error) {
      console.log(`⚠️  Could not update package.json: ${error.message}`)
    }
  }
}

// Run the generator
generateTypeScriptDeclarations()