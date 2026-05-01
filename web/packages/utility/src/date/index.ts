export type DateInput = Date | string | number

export const parseDate = (input: DateInput): Date => {
  if (input instanceof Date) return input
  return new Date(input)
}

export const isValidDate = (date: DateInput): boolean => {
  const parsedDate = parseDate(date)
  return !isNaN(parsedDate.getTime())
}

export const formatDate = (
  date: DateInput,
  locale = 'en-US',
  options: Intl.DateTimeFormatOptions = {}
): string => {
  const parsedDate = parseDate(date)
  return new Intl.DateTimeFormat(locale, options).format(parsedDate)
}

export const formatRelative = (date: DateInput, locale = 'en-US'): string => {
  const parsedDate = parseDate(date)
  const now = new Date()
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' })
  
  const diffInSeconds = (parsedDate.getTime() - now.getTime()) / 1000
  const diffInMinutes = diffInSeconds / 60
  const diffInHours = diffInMinutes / 60
  const diffInDays = diffInHours / 24
  const diffInWeeks = diffInDays / 7
  const diffInMonths = diffInDays / 30.44 // Average days per month
  const diffInYears = diffInDays / 365.25 // Average days per year
  
  if (Math.abs(diffInSeconds) < 60) {
    return rtf.format(Math.round(diffInSeconds), 'second')
  } else if (Math.abs(diffInMinutes) < 60) {
    return rtf.format(Math.round(diffInMinutes), 'minute')
  } else if (Math.abs(diffInHours) < 24) {
    return rtf.format(Math.round(diffInHours), 'hour')
  } else if (Math.abs(diffInDays) < 7) {
    return rtf.format(Math.round(diffInDays), 'day')
  } else if (Math.abs(diffInWeeks) < 4) {
    return rtf.format(Math.round(diffInWeeks), 'week')
  } else if (Math.abs(diffInMonths) < 12) {
    return rtf.format(Math.round(diffInMonths), 'month')
  } else {
    return rtf.format(Math.round(diffInYears), 'year')
  }
}

export const addDays = (date: DateInput, days: number): Date => {
  const result = new Date(parseDate(date))
  result.setDate(result.getDate() + days)
  return result
}

export const addWeeks = (date: DateInput, weeks: number): Date => {
  return addDays(date, weeks * 7)
}

export const addMonths = (date: DateInput, months: number): Date => {
  const result = new Date(parseDate(date))
  result.setMonth(result.getMonth() + months)
  return result
}

export const addYears = (date: DateInput, years: number): Date => {
  const result = new Date(parseDate(date))
  result.setFullYear(result.getFullYear() + years)
  return result
}

export const subtractDays = (date: DateInput, days: number): Date => {
  return addDays(date, -days)
}

export const subtractWeeks = (date: DateInput, weeks: number): Date => {
  return addWeeks(date, -weeks)
}

export const subtractMonths = (date: DateInput, months: number): Date => {
  return addMonths(date, -months)
}

export const subtractYears = (date: DateInput, years: number): Date => {
  return addYears(date, -years)
}

export const startOfDay = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setHours(0, 0, 0, 0)
  return result
}

export const endOfDay = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setHours(23, 59, 59, 999)
  return result
}

export const startOfWeek = (date: DateInput, weekStartsOn = 0): Date => {
  const result = new Date(parseDate(date))
  const day = result.getDay()
  const diff = (day < weekStartsOn ? 7 : 0) + day - weekStartsOn
  
  result.setDate(result.getDate() - diff)
  return startOfDay(result)
}

export const endOfWeek = (date: DateInput, weekStartsOn = 0): Date => {
  const result = startOfWeek(date, weekStartsOn)
  result.setDate(result.getDate() + 6)
  return endOfDay(result)
}

export const startOfMonth = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setDate(1)
  return startOfDay(result)
}

export const endOfMonth = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setMonth(result.getMonth() + 1, 0)
  return endOfDay(result)
}

export const startOfYear = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setMonth(0, 1)
  return startOfDay(result)
}

export const endOfYear = (date: DateInput): Date => {
  const result = new Date(parseDate(date))
  result.setMonth(11, 31)
  return endOfDay(result)
}

export const isSameDay = (date1: DateInput, date2: DateInput): boolean => {
  const d1 = parseDate(date1)
  const d2 = parseDate(date2)
  
  return d1.getFullYear() === d2.getFullYear() &&
         d1.getMonth() === d2.getMonth() &&
         d1.getDate() === d2.getDate()
}

export const isSameWeek = (date1: DateInput, date2: DateInput, weekStartsOn = 0): boolean => {
  const start1 = startOfWeek(date1, weekStartsOn)
  const start2 = startOfWeek(date2, weekStartsOn)
  return start1.getTime() === start2.getTime()
}

export const isSameMonth = (date1: DateInput, date2: DateInput): boolean => {
  const d1 = parseDate(date1)
  const d2 = parseDate(date2)
  
  return d1.getFullYear() === d2.getFullYear() &&
         d1.getMonth() === d2.getMonth()
}

export const isSameYear = (date1: DateInput, date2: DateInput): boolean => {
  const d1 = parseDate(date1)
  const d2 = parseDate(date2)
  
  return d1.getFullYear() === d2.getFullYear()
}

export const isToday = (date: DateInput): boolean => {
  return isSameDay(date, new Date())
}

export const isYesterday = (date: DateInput): boolean => {
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  return isSameDay(date, yesterday)
}

export const isTomorrow = (date: DateInput): boolean => {
  const tomorrow = new Date()
  tomorrow.setDate(tomorrow.getDate() + 1)
  return isSameDay(date, tomorrow)
}

export const isWeekend = (date: DateInput): boolean => {
  const day = parseDate(date).getDay()
  return day === 0 || day === 6 // Sunday or Saturday
}

export const isWeekday = (date: DateInput): boolean => {
  return !isWeekend(date)
}

export const isLeapYear = (year: number): boolean => {
  return (year % 4 === 0 && year % 100 !== 0) || (year % 400 === 0)
}

export const getDaysInMonth = (date: DateInput): number => {
  const d = parseDate(date)
  return new Date(d.getFullYear(), d.getMonth() + 1, 0).getDate()
}

export const getDaysInYear = (year: number): number => {
  return isLeapYear(year) ? 366 : 365
}

export const getWeekOfYear = (date: DateInput): number => {
  const d = new Date(parseDate(date))
  const yearStart = new Date(d.getFullYear(), 0, 1)
  const daysSinceYearStart = Math.floor((d.getTime() - yearStart.getTime()) / (24 * 60 * 60 * 1000))
  
  return Math.ceil((daysSinceYearStart + yearStart.getDay() + 1) / 7)
}

export const getAge = (birthDate: DateInput, referenceDate: DateInput = new Date()): number => {
  const birth = parseDate(birthDate)
  const reference = parseDate(referenceDate)
  
  let age = reference.getFullYear() - birth.getFullYear()
  const monthDiff = reference.getMonth() - birth.getMonth()
  
  if (monthDiff < 0 || (monthDiff === 0 && reference.getDate() < birth.getDate())) {
    age--
  }
  
  return age
}

export const getDateRange = (startDate: DateInput, endDate: DateInput): Date[] => {
  const start = parseDate(startDate)
  const end = parseDate(endDate)
  const dates: Date[] = []
  
  const current = new Date(start)
  while (current <= end) {
    dates.push(new Date(current))
    current.setDate(current.getDate() + 1)
  }
  
  return dates
}

export const getBusinessDays = (startDate: DateInput, endDate: DateInput): Date[] => {
  return getDateRange(startDate, endDate).filter(isWeekday)
}

export const getBusinessDaysCount = (startDate: DateInput, endDate: DateInput): number => {
  return getBusinessDays(startDate, endDate).length
}

export const getNthWeekdayOfMonth = (year: number, month: number, weekday: number, n: number): Date | null => {
  const firstDay = new Date(year, month, 1)
  const firstWeekday = firstDay.getDay()
  
  // Calculate the date of the first occurrence of the weekday
  let firstOccurrence = 1 + (weekday - firstWeekday + 7) % 7
  
  // Calculate the date of the nth occurrence
  const nthOccurrence = firstOccurrence + (n - 1) * 7
  
  // Check if the nth occurrence exists in this month
  const daysInMonth = getDaysInMonth(firstDay)
  if (nthOccurrence > daysInMonth) {
    return null
  }
  
  return new Date(year, month, nthOccurrence)
}

export const getTimeZoneOffset = (date: DateInput = new Date()): number => {
  return parseDate(date).getTimezoneOffset()
}

export const toUTC = (date: DateInput): Date => {
  const d = parseDate(date)
  return new Date(d.getTime() + d.getTimezoneOffset() * 60000)
}

export const fromUTC = (date: DateInput): Date => {
  const d = parseDate(date)
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000)
}

// Common date format patterns
export const formatPatterns = {
  ISO: 'YYYY-MM-DD',
  US: 'MM/DD/YYYY',
  EU: 'DD/MM/YYYY',
  SHORT: 'MMM DD, YYYY',
  LONG: 'MMMM DD, YYYY',
  FULL: 'dddd, MMMM DD, YYYY',
  TIME_12: 'h:mm A',
  TIME_24: 'HH:mm',
  DATETIME_12: 'MMM DD, YYYY h:mm A',
  DATETIME_24: 'MMM DD, YYYY HH:mm'
}