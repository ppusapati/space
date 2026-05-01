
import fs from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

function implementSemanticTokens() {
  console.log('🏗️  Implementing semantic token architecture...\n')
  
  // Step 1: Create proper theme files without theme.light nesting
  console.log('1. Restructuring theme files:')
  createSemanticThemeFiles()
  
  // Step 2: Fix component files to use semantic references
  console.log('\n2. Updating component files to use semantic tokens:')
  fixComponentReferences()
  
  // Step 3: Remove old problematic files
  console.log('\n3. Cleaning up old files:')
  cleanupOldFiles()
  
  console.log('\n🎉 Semantic token architecture implemented!')
  console.log('\n📋 What changed:')
  console.log('   ✅ Theme files now define semantic tokens directly')
  console.log('   ✅ Components reference semantic tokens (theme-agnostic)')
  console.log('   ✅ Build process automatically resolves themes')
  console.log('   ✅ CSS variables automatically switch themes')
  
  console.log('\n🚀 Next steps:')
  console.log('   1. Run: pnpm run debug:issues (verify no conflicts)')  
  console.log('   2. Run: pnpm run build:tokens (test build)')
  console.log('   3. Components now work with both light and dark themes!')
}

function createSemanticThemeFiles() {
  // Light Theme - Semantic token definitions
  const lightTheme = {
    "color": {
      "surface": {
        "primary": { "value": "{color.neutral.white}", "type": "color" },
        "secondary": { "value": "{color.neutral.50}", "type": "color" },
        "tertiary": { "value": "{color.neutral.100}", "type": "color" },
        "inverse": { "value": "{color.neutral.900}", "type": "color" },
        "overlay": { "value": "rgba(0, 0, 0, 0.5)", "type": "color" }
      },
      "text": {
        "primary": { "value": "{color.neutral.900}", "type": "color" },
        "secondary": { "value": "{color.neutral.700}", "type": "color" },
        "tertiary": { "value": "{color.neutral.500}", "type": "color" },
        "inverse": { "value": "{color.neutral.white}", "type": "color" },
        "placeholder": { "value": "{color.neutral.400}", "type": "color" },
        "disabled": { "value": "{color.neutral.300}", "type": "color" }
      },
      "border": {
        "primary": { "value": "{color.neutral.200}", "type": "color" },
        "secondary": { "value": "{color.neutral.300}", "type": "color" },
        "focus": { "value": "{color.brand.primary.500}", "type": "color" },
        "inverse": { "value": "{color.neutral.700}", "type": "color" }
      },
      "interactive": {
        "primary": { "value": "{color.brand.primary.500}", "type": "color" },
        "primaryHover": { "value": "{color.brand.primary.600}", "type": "color" },
        "primaryActive": { "value": "{color.brand.primary.700}", "type": "color" },
        "secondary": { "value": "{color.neutral.100}", "type": "color" },
        "secondaryHover": { "value": "{color.neutral.200}", "type": "color" },
        "secondaryActive": { "value": "{color.neutral.300}", "type": "color" }
      }
    },
    "shadow": {
      "card": { "value": "{shadow.sm}", "type": "shadow" },
      "modal": { "value": "{shadow.xl}", "type": "shadow" },
      "dropdown": { "value": "{shadow.lg}", "type": "shadow" }
    }
  }

  // Dark Theme - Different semantic token values
  const darkTheme = {
    "color": {
      "surface": {
        "primary": { "value": "{color.neutral.900}", "type": "color" },
        "secondary": { "value": "{color.neutral.800}", "type": "color" },
        "tertiary": { "value": "{color.neutral.700}", "type": "color" },
        "inverse": { "value": "{color.neutral.50}", "type": "color" },
        "overlay": { "value": "rgba(0, 0, 0, 0.7)", "type": "color" }
      },
      "text": {
        "primary": { "value": "{color.neutral.50}", "type": "color" },
        "secondary": { "value": "{color.neutral.300}", "type": "color" },
        "tertiary": { "value": "{color.neutral.400}", "type": "color" },
        "inverse": { "value": "{color.neutral.900}", "type": "color" },
        "placeholder": { "value": "{color.neutral.500}", "type": "color" },
        "disabled": { "value": "{color.neutral.600}", "type": "color" }
      },
      "border": {
        "primary": { "value": "{color.neutral.700}", "type": "color" },
        "secondary": { "value": "{color.neutral.600}", "type": "color" },
        "focus": { "value": "{color.brand.primary.400}", "type": "color" },
        "inverse": { "value": "{color.neutral.300}", "type": "color" }
      },
      "interactive": {
        "primary": { "value": "{color.brand.primary.400}", "type": "color" },
        "primaryHover": { "value": "{color.brand.primary.300}", "type": "color" },
        "primaryActive": { "value": "{color.brand.primary.200}", "type": "color" },
        "secondary": { "value": "{color.neutral.700}", "type": "color" },
        "secondaryHover": { "value": "{color.neutral.600}", "type": "color" },
        "secondaryActive": { "value": "{color.neutral.500}", "type": "color" }
      }
    },
    "shadow": {
      "card": { "value": "0 1px 3px 0 rgba(0, 0, 0, 0.3), 0 1px 2px -1px rgba(0, 0, 0, 0.3)", "type": "shadow" },
      "modal": { "value": "0 25px 50px -12px rgba(0, 0, 0, 0.5)", "type": "shadow" },
      "dropdown": { "value": "0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 8px 10px -6px rgba(0, 0, 0, 0.3)", "type": "shadow" }
    }
  }

  // Write theme files
  const lightPath = join(rootDir, 'tokens/themes/light.json')
  const darkPath = join(rootDir, 'tokens/themes/dark.json')
  
  fs.writeFileSync(lightPath, JSON.stringify(lightTheme, null, 2))
  console.log('   ✅ Updated light.json with semantic tokens')
  
  fs.writeFileSync(darkPath, JSON.stringify(darkTheme, null, 2))
  console.log('   ✅ Updated dark.json with semantic tokens')
}

function fixComponentReferences() {
  // Updated modal.json with semantic references
  const fixedModal = {
    "component": {
      "modal": {
        "overlay": {
          "backgroundColor": { "value": "{color.surface.overlay}", "type": "color" },
          "backdropFilter": { "value": "blur(4px)", "type": "string" },
          "zIndex": { "value": "{zIndex.modal}", "type": "number" }
        },
        "content": {
          "backgroundColor": { "value": "{color.surface.primary}", "type": "color" },
          "borderRadius": { "value": "{borderRadius.xl}", "type": "dimension" },
          "boxShadow": { "value": "{shadow.modal}", "type": "shadow" },
          "padding": { "value": "{spacing.6}", "type": "dimension" },
          "maxWidth": { "value": "{layout.container.md}", "type": "dimension" },
          "maxHeight": { "value": "90vh", "type": "dimension" },
          "width": { "value": "90vw", "type": "dimension" }
        }
      },
      "tooltip": {
        "backgroundColor": { "value": "{color.surface.inverse}", "type": "color" },
        "color": { "value": "{color.text.inverse}", "type": "color" },
        "borderRadius": { "value": "{borderRadius.md}", "type": "dimension" },
        "paddingX": { "value": "{spacing.3}", "type": "dimension" },
        "paddingY": { "value": "{spacing.2}", "type": "dimension" },
        "fontSize": { "value": "{typography.fontSize.xs}", "type": "dimension" },
        "fontWeight": { "value": "{typography.fontWeight.medium}", "type": "fontWeight" },
        "boxShadow": { "value": "{shadow.lg}", "type": "shadow" },
        "zIndex": { "value": "{zIndex.tooltip}", "type": "number" },
        "maxWidth": { "value": "200px", "type": "dimension" }
      },
      "dropdown": {
        "backgroundColor": { "value": "{color.surface.primary}", "type": "color" },
        "borderColor": { "value": "{color.border.primary}", "type": "color" },
        "borderRadius": { "value": "{borderRadius.lg}", "type": "dimension" },
        "borderWidth": { "value": "{borderWidth.default}", "type": "dimension" },
        "boxShadow": { "value": "{shadow.dropdown}", "type": "shadow" },
        "padding": { "value": "{spacing.1}", "type": "dimension" },
        "zIndex": { "value": "{zIndex.dropdown}", "type": "number" },
        "minWidth": { "value": "180px", "type": "dimension" },
        "item": {
          "paddingX": { "value": "{spacing.3}", "type": "dimension" },
          "paddingY": { "value": "{spacing.2}", "type": "dimension" },
          "borderRadius": { "value": "{borderRadius.base}", "type": "dimension" },
          "fontSize": { "value": "{typography.fontSize.sm}", "type": "dimension" },
          "color": { "value": "{color.text.primary}", "type": "color" },
          "cursor": { "value": "pointer", "type": "string" },
          "transition": { "value": "all {transition.duration.fast} {transition.easing.easeOut}", "type": "string" },
          "hover": {
            "backgroundColor": { "value": "{color.interactive.secondaryHover}", "type": "color" }
          },
          "active": {
            "backgroundColor": { "value": "{color.interactive.primaryActive}", "type": "color" },
            "color": { "value": "{color.text.inverse}", "type": "color" }
          }
        }
      },
      "badge": {
        "backgroundColor": { "value": "{color.interactive.secondary}", "type": "color" },
        "color": { "value": "{color.text.primary}", "type": "color" },
        "borderRadius": { "value": "{borderRadius.full}", "type": "dimension" },
        "paddingX": { "value": "{spacing.2}", "type": "dimension" },
        "paddingY": { "value": "{spacing.1}", "type": "dimension" },
        "fontSize": { "value": "{typography.fontSize.xs}", "type": "dimension" },
        "fontWeight": { "value": "{typography.fontWeight.medium}", "type": "fontWeight" },
        "lineHeight": { "value": "{typography.lineHeight.none}", "type": "number" },
        "variants": {
          "primary": {
            "backgroundColor": { "value": "{color.interactive.primary}", "type": "color" },
            "color": { "value": "{color.text.inverse}", "type": "color" }
          },
          "success": {
            "backgroundColor": { "value": "{color.semantic.success.50}", "type": "color" },
            "color": { "value": "{color.semantic.success.900}", "type": "color" }
          },
          "warning": {
            "backgroundColor": { "value": "{color.semantic.warning.50}", "type": "color" },
            "color": { "value": "{color.semantic.warning.900}", "type": "color" }
          },
          "error": {
            "backgroundColor": { "value": "{color.semantic.error.50}", "type": "color" },
            "color": { "value": "{color.semantic.error.900}", "type": "color" }
          }
        }
      }
    }
  }

  // Write fixed modal
  const modalPath = join(rootDir, 'tokens/components/modal.json')
  fs.writeFileSync(modalPath, JSON.stringify(fixedModal, null, 2))
  console.log('   ✅ Updated modal.json to use semantic tokens')

  // Fix other component files if they exist
  const componentDir = join(rootDir, 'tokens/components')
  if (fs.existsSync(componentDir)) {
    const componentFiles = fs.readdirSync(componentDir).filter(f => f.endsWith('.json') && f !== 'modal.json')
    
    componentFiles.forEach(file => {
      const filePath = join(componentDir, file)
      try {
        const content = fs.readFileSync(filePath, 'utf8')
        let tokens = JSON.parse(content)
        
        // Fix theme.light references
        const fixedContent = JSON.stringify(tokens, null, 2)
          .replace(/\{theme\.light\.color\./g, '{color.')
          .replace(/\{theme\.light\.shadow\./g, '{shadow.')
          .replace(/\{theme\.dark\.color\./g, '{color.')
          .replace(/\{theme\.dark\.shadow\./g, '{shadow.')
        
        if (fixedContent !== JSON.stringify(tokens, null, 2)) {
          fs.writeFileSync(filePath, fixedContent)
          console.log(`   ✅ Fixed theme references in ${file}`)
        } else {
          console.log(`   ⏭️  No theme references to fix in ${file}`)
        }
      } catch (error) {
        console.log(`   ⚠️  Could not process ${file}: ${error.message}`)
      }
    })
  }
}

function cleanupOldFiles() {
  const filesToRemove = [
    'tokens/themes/semantic-light.json',
    'tokens/themes/semantic-dark.json'
  ]
  
  filesToRemove.forEach(file => {
    const filePath = join(rootDir, file)
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath)
      console.log(`   ✅ Removed ${file}`)
    } else {
      console.log(`   ⏭️  ${file} not found (already removed)`)
    }
  })
}

// Add missing zIndex and other tokens that components need
function addMissingTokens() {
  console.log('\n4. Adding missing tokens:')
  
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
}

// // Run the implementation
implementSemanticTokens()
addMissingTokens()