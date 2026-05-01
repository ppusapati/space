/**
 * Utility for conditionally joining class names together
 * Supports strings, arrays, and objects with boolean values
 */

type ClassValue = string | undefined | null | false | ClassObject | ClassValue[];
type ClassObject = { [key: string]: boolean | undefined | null };

export function cn(...classes: ClassValue[]): string {
  const result: string[] = [];

  for (const cls of classes) {
    if (!cls) continue;

    if (typeof cls === 'string') {
      result.push(cls);
    } else if (Array.isArray(cls)) {
      const nested = cn(...cls);
      if (nested) result.push(nested);
    } else if (typeof cls === 'object') {
      for (const [key, value] of Object.entries(cls)) {
        if (value) result.push(key);
      }
    }
  }

  return result.join(' ');
}

/**
 * Creates a class string based on component variant
 */
export function variantClasses<T extends string>(
  baseClasses: string,
  variants: Record<T, string>,
  variant: T
): string {
  return cn(baseClasses, variants[variant]);
}

/**
 * Creates size-based classes
 */
export function sizeClasses(
  size: 'xs' | 'sm' | 'md' | 'lg' | 'xl',
  sizeMap: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', string>
): string {
  return sizeMap[size];
}
