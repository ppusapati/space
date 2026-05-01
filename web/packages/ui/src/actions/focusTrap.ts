/**
 * Focus Trap action - traps focus within an element
 *
 * Usage:
 * <div use:focusTrap>Modal content</div>
 * <div use:focusTrap={{ enabled: isOpen, returnFocus: true }}>Modal content</div>
 */

export interface FocusTrapOptions {
  /** Whether the trap is enabled */
  enabled?: boolean;
  /** Initial element to focus (selector or element) */
  initialFocus?: string | HTMLElement;
  /** Return focus to previous element on destroy */
  returnFocus?: boolean;
  /** Allow escape key to deactivate */
  escapeDeactivates?: boolean;
  /** Callback when escape is pressed */
  onEscape?: () => void;
}

const FOCUSABLE_SELECTORS = [
  'a[href]',
  'area[href]',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  'button:not([disabled])',
  'iframe',
  'object',
  'embed',
  '[contenteditable]',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

export function focusTrap(node: HTMLElement, options: FocusTrapOptions = {}) {
  let enabled = options.enabled !== false;
  let returnFocus = options.returnFocus !== false;
  let escapeDeactivates = options.escapeDeactivates !== false;
  let onEscape = options.onEscape;
  let previousActiveElement: HTMLElement | null = null;

  function getFocusableElements(): HTMLElement[] {
    return Array.from(node.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS)).filter(
      (el) => el.offsetParent !== null // Is visible
    );
  }

  function focusFirst() {
    const focusableElements = getFocusableElements();

    if (options.initialFocus) {
      let initialElement: HTMLElement | null = null;

      if (typeof options.initialFocus === 'string') {
        initialElement = node.querySelector(options.initialFocus);
      } else {
        initialElement = options.initialFocus;
      }

      if (initialElement && focusableElements.includes(initialElement)) {
        initialElement.focus();
        return;
      }
    }

    if (focusableElements.length > 0) {
      focusableElements[0]!.focus();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!enabled) return;

    if (event.key === 'Escape' && escapeDeactivates) {
      event.preventDefault();
      onEscape?.();
      return;
    }

    if (event.key !== 'Tab') return;

    const focusableElements = getFocusableElements();
    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0]!;
    const lastElement = focusableElements[focusableElements.length - 1]!;

    if (event.shiftKey) {
      // Shift + Tab
      if (document.activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
      }
    } else {
      // Tab
      if (document.activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    }
  }

  function handleFocusIn(event: FocusEvent) {
    if (!enabled) return;

    const target = event.target as HTMLElement;

    // If focus went outside the trap, bring it back
    if (!node.contains(target)) {
      event.preventDefault();
      focusFirst();
    }
  }

  function activate() {
    if (!enabled) return;

    previousActiveElement = document.activeElement as HTMLElement;

    // Small delay to ensure DOM is ready
    requestAnimationFrame(() => {
      focusFirst();
    });

    document.addEventListener('keydown', handleKeydown);
    document.addEventListener('focusin', handleFocusIn);
  }

  function deactivate() {
    document.removeEventListener('keydown', handleKeydown);
    document.removeEventListener('focusin', handleFocusIn);

    if (returnFocus && previousActiveElement) {
      previousActiveElement.focus();
    }
  }

  if (enabled) {
    activate();
  }

  return {
    update(newOptions: FocusTrapOptions = {}) {
      const wasEnabled = enabled;
      enabled = newOptions.enabled !== false;
      returnFocus = newOptions.returnFocus !== false;
      escapeDeactivates = newOptions.escapeDeactivates !== false;
      onEscape = newOptions.onEscape;

      if (!wasEnabled && enabled) {
        activate();
      } else if (wasEnabled && !enabled) {
        deactivate();
      }
    },
    destroy() {
      deactivate();
    },
  };
}

export default focusTrap;
