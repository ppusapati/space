#!/usr/bin/env node

import fs from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const rootDir = join(__dirname, '..')

function debugTokenIssues() {
  console.log('🐛 Debugging token collisions and reference errors...\n')
  
  // Collect all tokens from all files
  const allTokens = new Map()
  const tokenSources = new Map() // Track where each token comes from
  const references = new Map() // Track all references
  const errors = []
  
  // Read all token files
  const tokenDirs = [
    'tokens/global',
    'tokens/themes', 
    'tokens/layout',
    'tokens/components'
  ]
  
  tokenDirs.forEach(dir => {
    const dirPath = join(rootDir, dir)
    if (!fs.existsSync(dirPath)) return
    
    console.log(`📁 Processing ${dir}:`)
    
    const files = fs.readdirSync(dirPath).filter(f => f.endsWith('.json'))
    files.forEach(file => {
      const filePath = join(dirPath, file)
      const relativePath = `${dir}/${file}`
      
      try {
        const content = fs.readFileSync(filePath, 'utf8')
        const tokens = JSON.parse(content)
        
        console.log(`  📄 ${file}`)
        
        // Extract tokens and check for collisions
        extractTokens(tokens, '', allTokens, tokenSources, references, relativePath)
        
      } catch (error) {
        console.log(`  ❌ Error reading ${file}: ${error.message}`)
        errors.push(`Error in ${relativePath}: ${error.message}`)
      }
    })
    console.log('')
  })
  
  // Check for token collisions
  console.log('🔍 Checking for token collisions:')
  const collisions = new Map()
  
  tokenSources.forEach((sources, tokenPath) => {
    if (sources.length > 1) {
      collisions.set(tokenPath, sources)
    }
  })
  
  if (collisions.size > 0) {
    console.log(`❌ Found ${collisions.size} token collisions:`)
    collisions.forEach((sources, tokenPath) => {
      console.log(`  🔥 ${tokenPath}:`)
      sources.forEach(source => {
        console.log(`     - ${source}`)
      })
    })
  } else {
    console.log('✅ No token collisions found')
  }
  
  console.log('')
  
  // Check for broken references
  console.log('🔍 Checking token references:')
  const brokenRefs = []
  
  references.forEach((sourceFile, refPath) => {
    if (!allTokens.has(refPath)) {
      brokenRefs.push({ ref: refPath, source: sourceFile })
    }
  })
  
  if (brokenRefs.length > 0) {
    console.log(`❌ Found ${brokenRefs.length} broken references:`)
    brokenRefs.forEach(({ ref, source }) => {
      console.log(`  🔗 {${ref}} referenced in ${source}`)
      
      // Suggest similar tokens
      const similar = findSimilarTokens(ref, allTokens)
      if (similar.length > 0) {
        console.log(`     💡 Did you mean: ${similar.slice(0, 3).join(', ')}?`)
      }
    })
  } else {
    console.log('✅ All token references are valid')
  }
  
  console.log('')
  
  // Show token statistics
  console.log('📊 Token Statistics:')
  console.log(`  Total tokens: ${allTokens.size}`)
  console.log(`  Token collisions: ${collisions.size}`)
  console.log(`  Broken references: ${brokenRefs.length}`)
  console.log(`  Files processed: ${tokenSources.size > 0 ? [...new Set([...tokenSources.values()].flat())].length : 0}`)
  
  // Show token breakdown by type
  const tokensByType = new Map()
  allTokens.forEach((token, path) => {
    const type = token.type || 'unknown'
    if (!tokensByType.has(type)) {
      tokensByType.set(type, 0)
    }
    tokensByType.set(type, tokensByType.get(type) + 1)
  })
  
  console.log('\n📋 Tokens by type:')
  tokensByType.forEach((count, type) => {
    console.log(`  ${type}: ${count}`)
  })
  
  // Recommendations
  console.log('\n💡 Recommendations:')
  
  if (collisions.size > 0) {
    console.log('  1. Fix token collisions by:')
    console.log('     - Renaming duplicate tokens')
    console.log('     - Moving tokens to appropriate files')
    console.log('     - Using more specific token names')
  }
  
  if (brokenRefs.length > 0) {
    console.log('  2. Fix broken references by:')
    console.log('     - Checking token names for typos')
    console.log('     - Ensuring referenced tokens exist')
    console.log('     - Updating reference paths')
  }
  
  if (collisions.size === 0 && brokenRefs.length === 0) {
    console.log('  🎉 Token structure looks good!')
    console.log('  ✅ Ready to build tokens')
  }
  
  return collisions.size === 0 && brokenRefs.length === 0
}

function extractTokens(obj, prefix, allTokens, tokenSources, references, sourceFile) {
  for (const [key, value] of Object.entries(obj)) {
    const currentPath = prefix ? `${prefix}.${key}` : key
    
    if (value && typeof value === 'object') {
      if (value.value !== undefined) {
        // This is a token
        if (allTokens.has(currentPath)) {
          // Collision detected
          if (!tokenSources.has(currentPath)) {
            tokenSources.set(currentPath, [])
          }
          if (!tokenSources.get(currentPath).includes(sourceFile)) {
            tokenSources.get(currentPath).push(sourceFile)
          }
        } else {
          allTokens.set(currentPath, value)
          tokenSources.set(currentPath, [sourceFile])
        }
        
        // Check if value is a reference
        const tokenValue = value.value
        if (typeof tokenValue === 'string' && tokenValue.startsWith('{') && tokenValue.endsWith('}')) {
          const referencePath = tokenValue.slice(1, -1)
          references.set(referencePath, sourceFile)
        }
        
      } else {
        // This is a nested object
        extractTokens(value, currentPath, allTokens, tokenSources, references, sourceFile)
      }
    }
  }
}

function findSimilarTokens(target, allTokens) {
  const targetParts = target.split('.')
  const similar = []
  
  allTokens.forEach((token, path) => {
    const pathParts = path.split('.')
    
    // Check for partial matches
    let matchScore = 0
    targetParts.forEach(part => {
      if (pathParts.includes(part)) {
        matchScore++
      }
    })
    
    if (matchScore > 0) {
      similar.push({ path, score: matchScore })
    }
  })
  
  return similar
    .sort((a, b) => b.score - a.score)
    .map(item => item.path)
}

// Show file contents for debugging
function showFileContents() {
  console.log('\n📄 Current token file contents:')
  
  const files = [
    'tokens/themes/light.json',
    'tokens/themes/dark.json',
    'tokens/themes/semantic-light.json',
    'tokens/themes/semantic-dark.json'
  ]
  
  files.forEach(file => {
    const filePath = join(rootDir, file)
    if (fs.existsSync(filePath)) {
      try {
        const content = fs.readFileSync(filePath, 'utf8')
        const tokens = JSON.parse(content)
        console.log(`\n📁 ${file}:`)
        console.log(JSON.stringify(tokens, null, 2).substring(0, 500) + '...')
      } catch (error) {
        console.log(`\n❌ Error reading ${file}: ${error.message}`)
      }
    }
  })
}

// Main execution
const isValid = debugTokenIssues()

if (!isValid) {
  console.log('\n🔧 To see more details about your current files:')
  console.log('   Run: node debug-tokens.js --show-files')
  
  if (process.argv.includes('--show-files')) {
    showFileContents()
  }
}