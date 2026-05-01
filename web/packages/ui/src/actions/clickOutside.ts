/**
 * Click Outside action - triggers callback when clicking outside the element
 *
 * Usage:
 * <div use:clickOutside={() => isOpen = false}>Content</div>
 * <div use:clickOutside={{ handler: () => isOpen = false, enabled: isOpen }}>Content</div>
 */

export interface ClickOutsideOptions {
  /** Callback when click outside occurs */
  handler: (event: MouseEvent) => void;
  /** Whether the action is enabled */
  enabled?: boolean;
  /** Elements to exclude from triggering the handler */
  exclude?: (HTMLElement | string)[];
}

export function clickOutside(
  node: HTMLElement,
  options: ClickOutsideOptions | ((event: MouseEvent) => void)
) {
  let handler: (event: MouseEvent) => void;
  let enabled = true;
  let excludeElements: HTMLElement[] = [];

  function resolveExcludes(exclude?: (HTMLElement | string)[]) {
    if (!exclude) return [];
    return exclude
      .map((item) => {
        if (typeof item === 'string') {
          return document.querySelector(item) as HTMLElement;
        }
        return item;
      })
      .filter(Boolean) as HTMLElement[];
  }

  function parseOptions(opts: ClickOutsideOptions | ((event: MouseEvent) => void)) {
    if (typeof opts === 'function') {
      handler = opts;
      enabled = true;
      excludeElements = [];
    } else {
      handler = opts.handler;
      enabled = opts.enabled !== false;
      excludeElements = resolveExcludes(opts.exclude);
    }
  }

  function handleClick(event: MouseEvent) {
    if (!enabled) return;

    const target = event.target as HTMLElement;

    // Check if click is inside the node
    if (node.contains(target)) return;

    // Check if click is on excluded elements
    for (const excludeEl of excludeElements) {
      if (excludeEl && excludeEl.contains(target)) return;
    }

    handler(event);
  }

  parseOptions(options);

  // Use setTimeout to prevent immediate triggering
  setTimeout(() => {
    document.addEventListener('click', handleClick, true);
  }, 0);

  return {
    update(newOptions: ClickOutsideOptions | ((event: MouseEvent) => void)) {
      parseOptions(newOptions);
    },
    destroy() {
      document.removeEventListener('click', handleClick, true);
    },
  };
}

export default clickOutside;
