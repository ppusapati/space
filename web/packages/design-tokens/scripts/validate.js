#!/usr/bin/env node

import fs from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

function validateTokens() {
  console.log('🔍 Validating design tokens...')
  
  const errors = []
  const warnings = []

  // Check directory structure
  const requiredDirs = [
    'tokens',
    'tokens/global',
    'tokens/themes'
  ]

  requiredDirs.forEach(dir => {
    const dirPath = join(rootDir, dir)
    if (!fs.existsSync(dirPath)) {
      errors.push(`Missing required directory: ${dir}`)
    }
  })

  // Check required token files
  const requiredFiles = [
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

  const tokenFiles = []
  requiredFiles.forEach(file => {
    const filePath = join(rootDir, file)
    if (!fs.existsSync(filePath)) {
      errors.push(`Missing required file: ${file}`)
    } else {
      tokenFiles.push(filePath)
    }
  })

  // Validate JSON syntax
  const allTokens = new Map()
  tokenFiles.forEach(filePath => {
    try {
      const content = fs.readFileSync(filePath, 'utf8')
      const tokens = JSON.parse(content)
      
      // Collect all token paths for reference validation
      collectTokenPaths(tokens, '', allTokens)
      
      console.log(`✅ ${filePath.replace(rootDir + '/', '')} - Valid JSON`)
    } catch (error) {
      errors.push(`Invalid JSON in ${filePath}: ${error.message}`)
    }
  })

  // Validate token references
  tokenFiles.forEach(filePath => {
    try {
      const content = fs.readFileSync(filePath, 'utf8')
      const tokens = JSON.parse(content)
      
      validateTokenReferences(tokens, '', allTokens, filePath, errors)
    } catch (error) {
      // Already handled above
    }
  })

  // Check for common token patterns
  const requiredTokenPaths = [
    'color.neutral.white',
    'color.neutral.900',
    'color.brand.primary.500',
    'spacing.4',
    'typography.fontSize.base',
    'shadow.sm'
  ]

  requiredTokenPaths.forEach(path => {
    if (!allTokens.has(path)) {
      warnings.push(`Recommended token missing: ${path}`)
    }
  })

  // Report results
  console.log('\n📊 Validation Results:')
  
  if (errors.length === 0 && warnings.length === 0) {
    console.log('✅ All tokens are valid!')
    console.log(`📈 Found ${allTokens.size} tokens total`)
    return true
  }

  if (errors.length > 0) {
    console.log('\n❌ Errors:')
    errors.forEach(error => console.log(`  • ${error}`))
  }

  if (warnings.length > 0) {
    console.log('\n⚠️  Warnings:')
    warnings.forEach(warning => console.log(`  • ${warning}`))
  }

  console.log(`\n📈 Statistics:`)
  console.log(`  • Total tokens: ${allTokens.size}`)
  console.log(`  • Errors: ${errors.length}`)
  console.log(`  • Warnings: ${warnings.length}`)

  return errors.length === 0
}

function collectTokenPaths(obj, prefix, tokenMap) {
  for (const [key, value] of Object.entries(obj)) {
    const currentPath = prefix ? `${prefix}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token
        tokenMap.set(currentPath, value)
      } else {
        // This is a nested object
        collectTokenPaths(value, currentPath, tokenMap)
      }
    }
  }
}

function validateTokenReferences(obj, prefix, allTokens, filePath, errors) {
  for (const [key, value] of Object.entries(obj)) {
    const currentPath = prefix ? `${prefix}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token, check if value is a reference
        const tokenValue = value.value
        if (typeof tokenValue === 'string' && tokenValue.startsWith('{') && tokenValue.endsWith('}')) {
          // This is a reference
          const referencePath = tokenValue.slice(1, -1) // Remove { }
          if (!allTokens.has(referencePath)) {
            errors.push(`Invalid token reference in ${filePath}: ${tokenValue} (referenced from ${currentPath})`)
          }
        }
      } else {
        // This is a nested object
        validateTokenReferences(value, currentPath, allTokens, filePath, errors)
      }
    }
  }
}

function showTokenStructure() {
  console.log('\n📁 Expected Token Structure:')
  console.log(`
tokens/
├── global/
│   ├── colors.json        (base color palette)
│   ├── spacing.json       (spacing scale)
│   ├── typography.json    (font sizes, families, weights)
│   ├── shadows.json       (shadow definitions)
│   ├── border-radius.json (border radius scale)
│   └── animations.json    (duration, easing, keyframes)
└── themes/
    ├── light.json         (light theme semantic tokens)
    └── dark.json          (dark theme semantic tokens)
  `)
}

// Main execution
const isValid = validateTokens()

if (!isValid) {
  console.log('\n💡 Tips for fixing token issues:')
  console.log('1. Ensure all referenced tokens exist in base files')
  console.log('2. Check JSON syntax with a validator')
  console.log('3. Use consistent naming conventions')
  console.log('4. Reference tokens with {path.to.token} syntax')
  
  showTokenStructure()
  process.exit(1)
} else {
  console.log('\n🎉 Ready to build tokens!')
}