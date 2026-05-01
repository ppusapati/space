/**
 * Generate unique IDs for components
 */

let counter = 0;

/**
 * Generates a unique ID with an optional prefix
 */
export function uid(prefix = 'samavāya'): string {
  counter += 1;
  return `${prefix}-${counter}-${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * Reset the counter (useful for testing)
 */
export function resetUidCounter(): void {
  counter = 0;
}
