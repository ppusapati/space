// Validation - export everything
export * from './validation'

// File utilities
export * from './file'

// String utilities
export * from './string'

// Date utilities
export * from './date'

// Number utilities - exclude duplicates (formatting versions take precedence)
export {
  clamp,
  round,
  randomInt,
  randomFloat,
  lerp,
  normalize,
  scale,
  percentage,
  formatCompact,
  parseNumber,
  isEven,
  isOdd,
  isPrime,
  fibonacci,
  factorial,
  gcd,
  lcm,
  average,
  median,
  mode,
  standardDeviation,
  variance,
  sum,
  product,
  range,
  inRange,
  toRadians,
  toDegrees,
  distance
} from './number'

// Formatting - these are the canonical versions of formatBytes, formatCurrency, formatNumber
export {
  formatCurrency,
  formatPercent,
  formatNumber,
  formatBytes,
  formatDuration,
  formatTime,
  formatPhoneNumber,
  formatCreditCard,
  formatSSN,
  formatAddress,
  formatList,
  formatInitials,
  formatPlural,
  formatOrdinal,
  formatFileSize,
  formatHashtag,
  formatMention,
  formatTemplate,
  formatJson,
  formatCSV,
  formatMarkdown,
  stripFormatting
} from './formatting'

// Clipboard utilities
export {
  isClipboardSupported,
  isClipboardReadSupported,
  copyText,
  copyHTML,
  copyJSON,
  copyImage,
  readText,
  readClipboard,
  handlePaste,
  createPasteHandler,
  copyTableData,
  parseTableData,
} from './clipboard'
export type { ClipboardFormat, ClipboardItem, CopyOptions } from './clipboard'
