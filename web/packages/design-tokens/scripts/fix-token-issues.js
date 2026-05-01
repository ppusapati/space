#!/usr/bin/env node

import fs from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

function fixTokenIssues() {
  console.log('🔧 Fixing token issues...\n')
  
  // Fix #1: Remove semantic token files that are causing collisions
  const semanticFiles = [
    'tokens/themes/semantic-light.json',
    'tokens/themes/semantic-dark.json'
  ]
  
  console.log('1. Removing duplicate semantic token files:')
  semanticFiles.forEach(file => {
    const filePath = join(rootDir, file)
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath)
      console.log(`   ✅ Removed ${file}`)
    } else {
      console.log(`   ⏭️  ${file} not found (already removed)`)
    }
  })
  
  // Fix #2: Fix the broken spacing reference in modal.json
  console.log('\n2. Fixing broken spacing references:')
  const modalPath = join(rootDir, 'tokens/components/modal.json')
  
  if (fs.existsSync(modalPath)) {
    try {
      const content = fs.readFileSync(modalPath, 'utf8')
      let modalTokens = JSON.parse(content)
      
      // Fix the broken spacing reference
      let hasChanges = false
      const fixes = fixSpacingReferences(modalTokens, '')
      
      if (fixes.length > 0) {
        fixes.forEach(fix => {
          console.log(`   🔧 ${fix}`)
        })
        
        fs.writeFileSync(modalPath, JSON.stringify(modalTokens, null, 2))
        console.log(`   ✅ Updated ${modalPath}`)
        hasChanges = true
      }
      
      if (!hasChanges) {
        console.log(`   ℹ️  No spacing fixes needed in modal.json`)
      }
      
    } catch (error) {
      console.log(`   ❌ Error fixing modal.json: ${error.message}`)
    }
  } else {
    console.log(`   ⚠️  modal.json not found`)
  }
  
  // Fix #3: Add missing tokens that are commonly referenced
  console.log('\n3. Adding missing commonly referenced tokens:')
  addMissingTokens()
  
  // Fix #4: Validate all remaining references
  console.log('\n4. Validating remaining token references:')
  validateAllReferences()
  
  console.log('\n🎉 Token fixes completed!')
  console.log('💡 Next steps:')
  console.log('   1. Run: pnpm run debug:issues (to verify fixes)')
  console.log('   2. Run: pnpm run build:tokens (to test build)')
}

function fixSpacingReferences(obj, path) {
  const fixes = []
  
  for (const [key, value] of Object.entries(obj)) {
    const currentPath = path ? `${path}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token, check if it has multi-value spacing
        const tokenValue = value.value
        if (typeof tokenValue === 'string') {
          // Fix multi-value spacing references like "{spacing.2} {spacing.3}"
          if (tokenValue.includes('} {')) {
            // This is a composite value, fix it
            if (key === 'padding' || key === 'margin') {
              // For padding/margin, use a single spacing value or create proper composite
              if (tokenValue === '{spacing.2} {spacing.3}') {
                value.value = '{spacing.3}' // Use the larger value
                fixes.push(`Fixed ${currentPath}: ${tokenValue} → {spacing.3}`)
              }
            }
          }
          
          // Fix references to non-existent spacing tokens
          if (tokenValue.includes('{spacing.0.5}')) {
            value.value = tokenValue.replace('{spacing.0.5}', '{spacing.1}')
            fixes.push(`Fixed ${currentPath}: spacing.0.5 → spacing.1`)
          }
        }
      } else {
        // Recurse into nested objects
        const nestedFixes = fixSpacingReferences(value, currentPath)
        fixes.push(...nestedFixes)
      }
    }
  }
  
  return fixes
}

function addMissingTokens() {
  // Add missing zIndex tokens
  const missingTokens = {
    'tokens/global/z-index.json': {
      "zIndex": {
        "base": { "value": "1", "type": "number" },
        "dropdown": { "value": "1000", "type": "number" },
        "modal": { "value": "1050", "type": "number" },
        "tooltip": { "value": "1070", "type": "number" },
        "overlay": { "value": "1040", "type": "number" }
      }
    },
    
    'tokens/global/border-width.json': {
      "borderWidth": {
        "none": { "value": "0", "type": "dimension" },
        "default": { "value": "1px", "type": "dimension" },
        "thick": { "value": "2px", "type": "dimension" }
      }
    },
    
    'tokens/global/transitions.json': {
      "transition": {
        "duration": {
          "fast": { "value": "150ms", "type": "duration" },
          "normal": { "value": "300ms", "type": "duration" },
          "slow": { "value": "500ms", "type": "duration" }
        },
        "easing": {
          "easeOut": { "value": "ease-out", "type": "cubicBezier" },
          "easeIn": { "value": "ease-in", "type": "cubicBezier" },
          "easeInOut": { "value": "ease-in-out", "type": "cubicBezier" }
        }
      }
    }
  }
  
  Object.entries(missingTokens).forEach(([filePath, tokens]) => {
    const fullPath = join(rootDir, filePath)
    
    if (!fs.existsSync(fullPath)) {
      // Ensure directory exists
      const dir = dirname(fullPath)
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true })
      }
      
      fs.writeFileSync(fullPath, JSON.stringify(tokens, null, 2))
      console.log(`   ✅ Created ${filePath}`)
    } else {
      console.log(`   ⏭️  ${filePath} already exists`)
    }
  })
  
  // Add missing layout tokens
  const layoutPath = join(rootDir, 'tokens/layout/container.json')
  if (!fs.existsSync(layoutPath)) {
    const layoutTokens = {
      "layout": {
        "container": {
          "xs": { "value": "20rem", "type": "dimension" },
          "sm": { "value": "24rem", "type": "dimension" },
          "md": { "value": "28rem", "type": "dimension" },
          "lg": { "value": "32rem", "type": "dimension" },
          "xl": { "value": "36rem", "type": "dimension" },
          "2xl": { "value": "42rem", "type": "dimension" },
          "full": { "value": "100%", "type": "dimension" }
        }
      }
    }
    
    fs.writeFileSync(layoutPath, JSON.stringify(layoutTokens, null, 2))
    console.log(`   ✅ Created layout/container.json`)
  }
  
  // Fix typography line height
  const typographyPath = join(rootDir, 'tokens/global/typography.json')
  if (fs.existsSync(typographyPath)) {
    try {
      const content = fs.readFileSync(typographyPath, 'utf8')
      const typography = JSON.parse(content)
      
      if (!typography.typography.lineHeight) {
        typography.typography.lineHeight = {
          "none": { "value": "1", "type": "number" },
          "tight": { "value": "1.25", "type": "number" },
          "normal": { "value": "1.5", "type": "number" },
          "relaxed": { "value": "1.625", "type": "number" }
        }
        
        fs.writeFileSync(typographyPath, JSON.stringify(typography, null, 2))
        console.log(`   ✅ Added line height tokens to typography.json`)
      }
    } catch (error) {
      console.log(`   ⚠️  Could not update typography.json: ${error.message}`)
    }
  }
}

function validateAllReferences() {
  // Re-run basic validation to see if we fixed the major issues
  const tokenDirs = ['tokens/global', 'tokens/themes', 'tokens/layout', 'tokens/components']
  const allTokens = new Set()
  const references = []
  
  // Collect all tokens
  tokenDirs.forEach(dir => {
    const dirPath = join(rootDir, dir)
    if (!fs.existsSync(dirPath)) return
    
    const files = fs.readdirSync(dirPath).filter(f => f.endsWith('.json'))
    files.forEach(file => {
      const filePath = join(dirPath, file)
      try {
        const content = fs.readFileSync(filePath, 'utf8')
        const tokens = JSON.parse(content)
        
        extractTokenPaths(tokens, '', allTokens, references)
      } catch (error) {
        console.log(`   ⚠️  Could not validate ${file}: ${error.message}`)
      }
    })
  })
  
  // Check references
  const brokenRefs = references.filter(ref => !allTokens.has(ref.path))
  
  if (brokenRefs.length === 0) {
    console.log(`   ✅ All token references are valid`)
  } else {
    console.log(`   ⚠️  Still have ${brokenRefs.length} broken references:`)
    brokenRefs.slice(0, 5).forEach(ref => {
      console.log(`      🔗 {${ref.path}} in ${ref.file}`)
    })
    if (brokenRefs.length > 5) {
      console.log(`      ... and ${brokenRefs.length - 5} more`)
    }
  }
  
  console.log(`   📊 Total tokens: ${allTokens.size}`)
  console.log(`   📊 Total references: ${references.length}`)
}

function extractTokenPaths(obj, prefix, allTokens, references, sourceFile = '') {
  for (const [key, value] of Object.entries(obj)) {
    const currentPath = prefix ? `${prefix}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token
        allTokens.add(currentPath)
        
        // Check if value is a reference
        const tokenValue = value.value
        if (typeof tokenValue === 'string' && tokenValue.startsWith('{') && tokenValue.endsWith('}')) {
          const referencePath = tokenValue.slice(1, -1)
          references.push({ path: referencePath, file: sourceFile })
        }
      } else {
        // This is a nested object
        extractTokenPaths(value, currentPath, allTokens, references, sourceFile)
      }
    }
  }
}

// Run the fixes
fixTokenIssues()