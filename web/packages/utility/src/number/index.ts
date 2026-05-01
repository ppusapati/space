export const clamp = (value: number, min: number, max: number): number => {
  return Math.min(Math.max(value, min), max)
}

export const round = (value: number, precision = 0): number => {
  const factor = Math.pow(10, precision)
  return Math.round(value * factor) / factor
}

export const randomInt = (min: number, max: number): number => {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export const randomFloat = (min: number, max: number, precision = 2): number => {
  const random = Math.random() * (max - min) + min
  return round(random, precision)
}

export const lerp = (start: number, end: number, factor: number): number => {
  return start + (end - start) * factor
}

export const normalize = (value: number, min: number, max: number): number => {
  return (value - min) / (max - min)
}

export const scale = (value: number, inMin: number, inMax: number, outMin: number, outMax: number): number => {
  return ((value - inMin) * (outMax - outMin)) / (inMax - inMin) + outMin
}

export const percentage = (value: number, total: number, precision = 2): number => {
  if (total === 0) return 0
  return round((value / total) * 100, precision)
}

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

export const formatCompact = (value: number, locale = 'en-US'): string => {
  return new Intl.NumberFormat(locale, {
    notation: 'compact',
    maximumFractionDigits: 1
  }).format(value)
}

export const parseNumber = (value: string | number): number => {
  if (typeof value === 'number') return value
  
  // Remove non-numeric characters except decimal point and minus
  const cleaned = value.replace(/[^\d.-]/g, '')
  const parsed = parseFloat(cleaned)
  
  return isNaN(parsed) ? 0 : parsed
}

export const isEven = (value: number): boolean => {
  return value % 2 === 0
}

export const isOdd = (value: number): boolean => {
  return value % 2 !== 0
}

export const isPrime = (value: number): boolean => {
  if (value < 2) return false
  if (value === 2) return true
  if (value % 2 === 0) return false
  
  for (let i = 3; i <= Math.sqrt(value); i += 2) {
    if (value % i === 0) return false
  }
  
  return true
}

export const fibonacci = (n: number): number => {
  if (n <= 1) return n
  
  let a = 0, b = 1
  for (let i = 2; i <= n; i++) {
    [a, b] = [b, a + b]
  }
  
  return b
}

export const factorial = (n: number): number => {
  if (n < 0) return NaN
  if (n === 0 || n === 1) return 1
  
  let result = 1
  for (let i = 2; i <= n; i++) {
    result *= i
  }
  
  return result
}

export const gcd = (a: number, b: number): number => {
  a = Math.abs(a)
  b = Math.abs(b)
  
  while (b !== 0) {
    [a, b] = [b, a % b]
  }
  
  return a
}

export const lcm = (a: number, b: number): number => {
  return Math.abs(a * b) / gcd(a, b)
}

export const average = (numbers: number[]): number => {
  if (numbers.length === 0) return 0
  return numbers.reduce((sum, num) => sum + num, 0) / numbers.length
}

export const median = (numbers: number[]): number => {
  if (numbers.length === 0) return 0

  const sorted = [...numbers].sort((a, b) => a - b)
  const middle = Math.floor(sorted.length / 2)

  return sorted.length % 2 === 0
    ? (sorted[middle - 1]! + sorted[middle]!) / 2
    : sorted[middle]!
}

export const mode = (numbers: number[]): number[] => {
  if (numbers.length === 0) return []
  
  const frequency: Record<number, number> = {}
  let maxFreq = 0
  
  // Count frequencies
  numbers.forEach(num => {
    frequency[num] = (frequency[num] || 0) + 1
    maxFreq = Math.max(maxFreq, frequency[num])
  })
  
  // Return all numbers with max frequency
  return Object.keys(frequency)
    .filter(key => frequency[Number(key)] === maxFreq)
    .map(Number)
}

export const standardDeviation = (numbers: number[]): number => {
  if (numbers.length === 0) return 0
  
  const avg = average(numbers)
  const squaredDiffs = numbers.map(num => Math.pow(num - avg, 2))
  const avgSquaredDiff = average(squaredDiffs)
  
  return Math.sqrt(avgSquaredDiff)
}

export const variance = (numbers: number[]): number => {
  if (numbers.length === 0) return 0
  
  const avg = average(numbers)
  const squaredDiffs = numbers.map(num => Math.pow(num - avg, 2))
  
  return average(squaredDiffs)
}

export const sum = (numbers: number[]): number => {
  return numbers.reduce((total, num) => total + num, 0)
}

export const product = (numbers: number[]): number => {
  return numbers.reduce((total, num) => total * num, 1)
}

export const range = (start: number, end: number, step = 1): number[] => {
  const result: number[] = []
  
  if (step === 0) return result
  
  if (step > 0) {
    for (let i = start; i <= end; i += step) {
      result.push(i)
    }
  } else {
    for (let i = start; i >= end; i += step) {
      result.push(i)
    }
  }
  
  return result
}

export const inRange = (value: number, min: number, max: number, inclusive = true): boolean => {
  return inclusive
    ? value >= min && value <= max
    : value > min && value < max
}

export const toRadians = (degrees: number): number => {
  return degrees * (Math.PI / 180)
}

export const toDegrees = (radians: number): number => {
  return radians * (180 / Math.PI)
}

export const distance = (x1: number, y1: number, x2: number, y2: number): number => {
  return Math.sqrt(Math.pow(x2 - x1, 2) + Math.pow(y2 - y1, 2))
}