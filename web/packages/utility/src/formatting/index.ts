export const formatCurrency = (
  amount: number,
  currency = 'USD',
  locale = 'en-US',
  options: Intl.NumberFormatOptions = {}
): string => {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    ...options
  }).format(amount)
}

export const formatPercent = (
  value: number,
  locale = 'en-US',
  options: Intl.NumberFormatOptions = {}
): string => {
  return new Intl.NumberFormat(locale, {
    style: 'percent',
    ...options
  }).format(value)
}

export const formatNumber = (
  value: number,
  locale = 'en-US',
  options: Intl.NumberFormatOptions = {}
): string => {
  return new Intl.NumberFormat(locale, options).format(value)
}

export const formatBytes = (bytes: number, decimals = 2): string => {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

export const formatDuration = (milliseconds: number): string => {
  const seconds = Math.floor(milliseconds / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    return `${days}d ${hours % 24}h ${minutes % 60}m ${seconds % 60}s`
  } else if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  } else {
    return `${seconds}s`
  }
}

export const formatTime = (
  date: Date,
  format: '12' | '24' = '12',
  includeSeconds = false
): string => {
  const options: Intl.DateTimeFormatOptions = {
    hour: 'numeric',
    minute: '2-digit',
    hour12: format === '12'
  }

  if (includeSeconds) {
    options.second = '2-digit'
  }

  return date.toLocaleTimeString('en-US', options)
}

export const formatPhoneNumber = (phoneNumber: string, format: 'US' | 'INTERNATIONAL' = 'US'): string => {
  // Remove all non-digit characters
  const cleaned = phoneNumber.replace(/\D/g, '')

  if (format === 'US' && cleaned.length === 10) {
    // Format as (XXX) XXX-XXXX
    return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`
  } else if (format === 'US' && cleaned.length === 11 && cleaned.startsWith('1')) {
    // Format as +1 (XXX) XXX-XXXX
    return `+1 (${cleaned.slice(1, 4)}) ${cleaned.slice(4, 7)}-${cleaned.slice(7)}`
  } else if (format === 'INTERNATIONAL') {
    // Basic international format
    if (cleaned.length > 10) {
      return `+${cleaned.slice(0, -10)} ${cleaned.slice(-10, -7)} ${cleaned.slice(-7, -4)} ${cleaned.slice(-4)}`
    }
  }

  // Return original if can't format
  return phoneNumber
}

export const formatCreditCard = (cardNumber: string, separator = ' '): string => {
  // Remove all non-digit characters
  const cleaned = cardNumber.replace(/\D/g, '')
  
  // Add separator every 4 digits
  return cleaned.replace(/(.{4})/g, `$1${separator}`).trim()
}

export const formatSSN = (ssn: string, masked = false): string => {
  const cleaned = ssn.replace(/\D/g, '')
  
  if (cleaned.length !== 9) {
    return ssn // Return original if invalid
  }

  if (masked) {
    return `XXX-XX-${cleaned.slice(5)}`
  }

  return `${cleaned.slice(0, 3)}-${cleaned.slice(3, 5)}-${cleaned.slice(5)}`
}

export const formatAddress = (address: {
  street?: string
  city?: string
  state?: string
  zipCode?: string
  country?: string
}, format: 'US' | 'INTERNATIONAL' = 'US'): string => {
  const { street, city, state, zipCode, country } = address

  if (format === 'US') {
    const parts = [
      street,
      [city, state].filter(Boolean).join(', '),
      zipCode
    ].filter(Boolean)
    
    return parts.join('\n')
  } else {
    const parts = [
      street,
      city,
      [state, zipCode].filter(Boolean).join(' '),
      country
    ].filter(Boolean)
    
    return parts.join('\n')
  }
}

export const formatList = (
  items: string[],
  type: 'conjunction' | 'disjunction' = 'conjunction',
  locale = 'en-US'
): string => {
  if (items.length === 0) return ''
  if (items.length === 1) return items[0]!

  const listFormat = new Intl.ListFormat(locale, {
    style: 'long',
    type: type
  })

  return listFormat.format(items)
}

export const formatInitials = (name: string, maxInitials = 2): string => {
  const names = name.trim().split(/\s+/)
  const initials = names
    .slice(0, maxInitials)
    .map(name => name.charAt(0).toUpperCase())
    .join('')

  return initials
}

export const formatPlural = (count: number, singular: string, plural?: string): string => {
  const pluralForm = plural || `${singular}s`
  return count === 1 ? `${count} ${singular}` : `${count} ${pluralForm}`
}

export const formatOrdinal = (number: number, locale = 'en-US'): string => {
  const pr = new Intl.PluralRules(locale, { type: 'ordinal' })
  const suffixes = new Map([
    ['one', 'st'],
    ['two', 'nd'],
    ['few', 'rd'],
    ['other', 'th']
  ])
  
  const rule = pr.select(number)
  const suffix = suffixes.get(rule) || 'th'
  
  return `${number}${suffix}`
}

export const formatFileSize = (bytes: number, binary = false): string => {
  const base = binary ? 1024 : 1000
  const units = binary 
    ? ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
    : ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

  if (bytes === 0) return '0 B'

  const i = Math.floor(Math.log(bytes) / Math.log(base))
  const size = bytes / Math.pow(base, i)

  return `${size.toFixed(1)} ${units[i]}`
}

export const formatHashtag = (text: string): string => {
  return text
    .toLowerCase()
    .replace(/[^\w\s]/g, '')
    .replace(/\s+/g, '')
    .replace(/^/, '#')
}

export const formatMention = (username: string): string => {
  return `@${username.replace(/^@/, '')}`
}

export const formatTemplate = (template: string, variables: Record<string, any>): string => {
  return template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
    return variables.hasOwnProperty(key) ? String(variables[key]) : match
  })
}

export const formatJson = (obj: any, indent = 2): string => {
  try {
    return JSON.stringify(obj, null, indent)
  } catch (error) {
    return String(obj)
  }
}

export const formatCSV = (data: Record<string, any>[], delimiter = ','): string => {
  if (data.length === 0) return ''

  const headers = Object.keys(data[0]!)
  const csvRows = [headers.join(delimiter)]

  for (const row of data) {
    const values = headers.map(header => {
      const value = row[header]
      // Escape quotes and wrap in quotes if contains delimiter
      const escaped = String(value).replace(/"/g, '""')
      return escaped.includes(delimiter) || escaped.includes('\n') || escaped.includes('"')
        ? `"${escaped}"`
        : escaped
    })
    csvRows.push(values.join(delimiter))
  }

  return csvRows.join('\n')
}

export const formatMarkdown = {
  bold: (text: string) => `**${text}**`,
  italic: (text: string) => `*${text}*`,
  code: (text: string) => `\`${text}\``,
  codeBlock: (text: string, language?: string) => `\`\`\`${language || ''}\n${text}\n\`\`\``,
  link: (text: string, url: string) => `[${text}](${url})`,
  image: (alt: string, src: string) => `![${alt}](${src})`,
  heading: (text: string, level: number) => `${'#'.repeat(level)} ${text}`,
  quote: (text: string) => `> ${text}`,
  list: (items: string[], ordered = false) => {
    return items.map((item, index) => 
      ordered ? `${index + 1}. ${item}` : `- ${item}`
    ).join('\n')
  }
}

export const stripFormatting = {
  html: (text: string) => text.replace(/<[^>]*>/g, ''),
  markdown: (text: string) => text
    .replace(/[*_`~]/g, '')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/^#+\s*/gm, '')
    .replace(/^>\s*/gm, '')
    .replace(/^[-*+]\s*/gm, '')
    .replace(/^\d+\.\s*/gm, ''),
  
  whitespace: (text: string) => text
    .replace(/\s+/g, ' ')
    .trim(),
  
  nonPrintable: (text: string) => text.replace(/[^\x20-\x7E]/g, '')
}