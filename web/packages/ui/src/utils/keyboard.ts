/**
 * Keyboard navigation utilities
 */

export const Keys = {
  Enter: 'Enter',
  Space: ' ',
  Escape: 'Escape',
  Tab: 'Tab',
  ArrowUp: 'ArrowUp',
  ArrowDown: 'ArrowDown',
  ArrowLeft: 'ArrowLeft',
  ArrowRight: 'ArrowRight',
  Home: 'Home',
  End: 'End',
  PageUp: 'PageUp',
  PageDown: 'PageDown',
} as const;

export type KeyName = keyof typeof Keys;

/**
 * Check if a keyboard event matches a specific key
 */
export function isKey(event: KeyboardEvent, key: KeyName): boolean {
  return event.key === Keys[key];
}

/**
 * Check if Enter or Space was pressed (common for button activation)
 */
export function isActivationKey(event: KeyboardEvent): boolean {
  return isKey(event, 'Enter') || isKey(event, 'Space');
}

/**
 * Check if an arrow key was pressed
 */
export function isArrowKey(event: KeyboardEvent): boolean {
  return (
    isKey(event, 'ArrowUp') ||
    isKey(event, 'ArrowDown') ||
    isKey(event, 'ArrowLeft') ||
    isKey(event, 'ArrowRight')
  );
}

/**
 * Get the next index in a list based on arrow key navigation
 */
export function getNextIndex(
  event: KeyboardEvent,
  currentIndex: number,
  length: number,
  options: { loop?: boolean; vertical?: boolean } = {}
): number {
  const { loop = true, vertical = true } = options;
  const upKey = vertical ? 'ArrowUp' : 'ArrowLeft';
  const downKey = vertical ? 'ArrowDown' : 'ArrowRight';

  let nextIndex = currentIndex;

  if (event.key === Keys[upKey === 'ArrowUp' ? 'ArrowUp' : 'ArrowLeft']) {
    nextIndex = currentIndex - 1;
    if (nextIndex < 0) {
      nextIndex = loop ? length - 1 : 0;
    }
  } else if (event.key === Keys[downKey === 'ArrowDown' ? 'ArrowDown' : 'ArrowRight']) {
    nextIndex = currentIndex + 1;
    if (nextIndex >= length) {
      nextIndex = loop ? 0 : length - 1;
    }
  } else if (isKey(event, 'Home')) {
    nextIndex = 0;
  } else if (isKey(event, 'End')) {
    nextIndex = length - 1;
  }

  return nextIndex;
}

/**
 * Create a keyboard event handler for list navigation
 */
export function createListNavigationHandler(
  getCurrentIndex: () => number,
  getLength: () => number,
  onSelect: (index: number) => void,
  options: { loop?: boolean; vertical?: boolean } = {}
): (event: KeyboardEvent) => void {
  return (event: KeyboardEvent) => {
    if (!isArrowKey(event) && !isKey(event, 'Home') && !isKey(event, 'End')) {
      return;
    }

    event.preventDefault();
    const currentIndex = getCurrentIndex();
    const length = getLength();
    const nextIndex = getNextIndex(event, currentIndex, length, options);

    if (nextIndex !== currentIndex) {
      onSelect(nextIndex);
    }
  };
}
