export const capitalize = (str: string): string => {
  if (!str) return str
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase()
}

export const camelCase = (str: string): string => {
  return str
    .replace(/(?:^\w|[A-Z]|\b\w)/g, (word, index) => 
      index === 0 ? word.toLowerCase() : word.toUpperCase()
    )
    .replace(/\s+/g, '')
}

export const kebabCase = (str: string): string => {
  return str
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/[\s_]+/g, '-')
    .toLowerCase()
}

export const snakeCase = (str: string): string => {
  return str
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .replace(/[\s-]+/g, '_')
    .toLowerCase()
}

export const pascalCase = (str: string): string => {
  return str
    .replace(/(?:^\w|[A-Z]|\b\w|\s+)/g, (match, index) => 
      +match === 0 ? '' : match.toUpperCase()
    )
}

export const truncate = (str: string, length: number, suffix = '...'): string => {
  if (str.length <= length) return str
  return str.substring(0, length - suffix.length) + suffix
}

export const slugify = (str: string): string => {
  return str
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export const removeAccents = (str: string): string => {
  return str.normalize('NFD').replace(/[\u0300-\u036f]/g, '')
}

export const mask = (str: string, maskChar = '*', visibleStart = 0, visibleEnd = 0): string => {
  if (str.length <= visibleStart + visibleEnd) {
    return str
  }
  
  const start = str.substring(0, visibleStart)
  const end = visibleEnd > 0 ? str.substring(str.length - visibleEnd) : ''
  const masked = maskChar.repeat(str.length - visibleStart - visibleEnd)
  
  return start + masked + end
}

export const ellipsis = (str: string, maxLength: number, position: 'start' | 'middle' | 'end' = 'end'): string => {
  if (str.length <= maxLength) return str
  
  const ellipsisStr = '...'
  const ellipsisLength = ellipsisStr.length
  
  switch (position) {
    case 'start':
      return ellipsisStr + str.slice(str.length - maxLength + ellipsisLength)
    case 'middle':
      const midPoint = Math.floor((maxLength - ellipsisLength) / 2)
      return str.slice(0, midPoint) + ellipsisStr + str.slice(str.length - midPoint)
    case 'end':
    default:
      return str.slice(0, maxLength - ellipsisLength) + ellipsisStr
  }
}

export const isEmail = (str: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(str)
}

export const isUrl = (str: string): boolean => {
  try {
    new URL(str)
    return true
  } catch {
    return false
  }
}

export const extractUrls = (text: string): string[] => {
  const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi
  return text.match(urlRegex) || []
}

export const extractEmails = (text: string): string[] => {
  const emailRegex = /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b/g
  return text.match(emailRegex) || []
}

export const removeHtml = (str: string): string => {
  return str.replace(/<[^>]*>/g, '')
}

export const escapeHtml = (str: string): string => {
  const htmlEscapes: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;'
  }
  
  return str.replace(/[&<>"'\/]/g, match => htmlEscapes[match] ?? match)
}

export const unescapeHtml = (str: string): string => {
  const htmlUnescapes: Record<string, string> = {
    '&amp;': '&',
    '&lt;': '<',
    '&gt;': '>',
    '&quot;': '"',
    '&#x27;': "'",
    '&#x2F;': '/'
  }
  
  return str.replace(/&(?:amp|lt|gt|quot|#x27|#x2F);/g, match => htmlUnescapes[match] ?? match)
}

export const wordCount = (str: string): number => {
  return str.trim().split(/\s+/).filter(word => word.length > 0).length
}

export const characterCount = (str: string, includeSpaces = true): number => {
  return includeSpaces ? str.length : str.replace(/\s/g, '').length
}

export const reverse = (str: string): string => {
  return str.split('').reverse().join('')
}

export const isPalindrome = (str: string): boolean => {
  const cleaned = str.toLowerCase().replace(/[^a-z0-9]/gi, '')
  return cleaned === reverse(cleaned)
}

export const randomString = (length: number, charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'): string => {
  let result = ''
  for (let i = 0; i < length; i++) {
    result += charset.charAt(Math.floor(Math.random() * charset.length))
  }
  return result
}

export const levenshteinDistance = (str1: string, str2: string): number => {
  const matrix: number[][] = Array(str2.length + 1).fill(null).map(() => Array(str1.length + 1).fill(0))

  for (let i = 0; i <= str1.length; i++) {
    matrix[0]![i] = i
  }

  for (let j = 0; j <= str2.length; j++) {
    matrix[j]![0] = j
  }

  for (let j = 1; j <= str2.length; j++) {
    for (let i = 1; i <= str1.length; i++) {
      const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1
      const row = matrix[j]!
      const prevRow = matrix[j - 1]!
      row[i] = Math.min(
        row[i - 1]! + 1, // deletion
        prevRow[i]! + 1, // insertion
        prevRow[i - 1]! + indicator // substitution
      )
    }
  }

  return matrix[str2.length]![str1.length]!
}

export const similarity = (str1: string, str2: string): number => {
  const maxLength = Math.max(str1.length, str2.length)
  if (maxLength === 0) return 1
  
  const distance = levenshteinDistance(str1, str2)
  return (maxLength - distance) / maxLength
}

export const highlight = (text: string, query: string, className = 'highlight'): string => {
  if (!query.trim()) return text
  
  const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return text.replace(regex, `<span class="${className}">$1</span>`)
}