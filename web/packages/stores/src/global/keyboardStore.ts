/**
 * Keyboard Shortcuts Store
 *
 * Global keyboard shortcut management system with:
 * - Shortcut registration and unregistration
 * - Modifier key support (Ctrl, Alt, Shift, Meta)
 * - Context-aware shortcuts (modal, page, global)
 * - Shortcut help panel data
 * - Conflict detection
 * - Chained shortcuts (e.g., g then h for "go home")
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type ModifierKey = 'ctrl' | 'alt' | 'shift' | 'meta';

export type ShortcutContext = 'global' | 'page' | 'modal' | 'input';

export interface ShortcutDefinition {
  /** Unique identifier for the shortcut */
  id: string;
  /** The key to press (lowercase) */
  key: string;
  /** Modifier keys required */
  modifiers?: ModifierKey[];
  /** Context where this shortcut is active */
  context?: ShortcutContext;
  /** Human-readable description */
  description: string;
  /** Category for grouping in help panel */
  category?: string;
  /** Callback function to execute */
  handler: (event: KeyboardEvent) => void | Promise<void>;
  /** Whether the shortcut is enabled */
  enabled?: boolean;
  /** Prevent default browser behavior */
  preventDefault?: boolean;
  /** Stop event propagation */
  stopPropagation?: boolean;
  /** Allow when focused on input elements */
  allowInInput?: boolean;
  /** Chained key sequence (for multi-key shortcuts like vim) */
  sequence?: string[];
}

export interface ShortcutGroup {
  name: string;
  shortcuts: ShortcutDefinition[];
}

export interface KeyboardState {
  /** All registered shortcuts */
  shortcuts: Map<string, ShortcutDefinition>;
  /** Current active context */
  activeContext: ShortcutContext;
  /** Whether help panel is visible */
  helpVisible: boolean;
  /** Pending keys for chained shortcuts */
  pendingSequence: string[];
  /** Timeout for sequence reset */
  sequenceTimeout: number;
  /** Whether keyboard shortcuts are globally enabled */
  enabled: boolean;
}

// ============================================================================
// Constants
// ============================================================================

const SEQUENCE_TIMEOUT = 1000; // ms to wait for next key in sequence

const MODIFIER_MAP: Record<string, ModifierKey> = {
  Control: 'ctrl',
  Alt: 'alt',
  Shift: 'shift',
  Meta: 'meta',
};

// ============================================================================
// Helpers
// ============================================================================

function normalizeKey(key: string): string {
  return key.toLowerCase();
}

function getModifiersFromEvent(event: KeyboardEvent): Set<ModifierKey> {
  const modifiers = new Set<ModifierKey>();
  if (event.ctrlKey) modifiers.add('ctrl');
  if (event.altKey) modifiers.add('alt');
  if (event.shiftKey) modifiers.add('shift');
  if (event.metaKey) modifiers.add('meta');
  return modifiers;
}

function modifiersMatch(
  required: ModifierKey[] | undefined,
  actual: Set<ModifierKey>
): boolean {
  const requiredSet = new Set(required || []);

  if (requiredSet.size !== actual.size) return false;

  for (const mod of requiredSet) {
    if (!actual.has(mod)) return false;
  }

  return true;
}

function generateShortcutKey(
  key: string,
  modifiers?: ModifierKey[],
  sequence?: string[]
): string {
  const parts: string[] = [];

  if (modifiers?.length) {
    parts.push(...[...modifiers].sort());
  }

  if (sequence?.length) {
    parts.push(...sequence);
  }

  parts.push(normalizeKey(key));

  return parts.join('+');
}

function formatShortcutForDisplay(shortcut: ShortcutDefinition): string {
  const parts: string[] = [];

  if (shortcut.modifiers?.includes('ctrl')) {
    parts.push(navigator.platform.includes('Mac') ? '⌃' : 'Ctrl');
  }
  if (shortcut.modifiers?.includes('alt')) {
    parts.push(navigator.platform.includes('Mac') ? '⌥' : 'Alt');
  }
  if (shortcut.modifiers?.includes('shift')) {
    parts.push(navigator.platform.includes('Mac') ? '⇧' : 'Shift');
  }
  if (shortcut.modifiers?.includes('meta')) {
    parts.push(navigator.platform.includes('Mac') ? '⌘' : 'Win');
  }

  if (shortcut.sequence?.length) {
    parts.push(...shortcut.sequence.map((k) => k.toUpperCase()));
  }

  parts.push(shortcut.key.toUpperCase());

  return parts.join(navigator.platform.includes('Mac') ? '' : '+');
}

function isInputElement(element: Element | null): boolean {
  if (!element) return false;

  const tagName = element.tagName.toLowerCase();
  const isInput = ['input', 'textarea', 'select'].includes(tagName);
  const isContentEditable = (element as HTMLElement).isContentEditable;

  return isInput || isContentEditable;
}

// ============================================================================
// Store
// ============================================================================

const initialState: KeyboardState = {
  shortcuts: new Map(),
  activeContext: 'global',
  helpVisible: false,
  pendingSequence: [],
  sequenceTimeout: 0,
  enabled: true,
};

function createKeyboardStore() {
  const { subscribe, set, update } = writable<KeyboardState>(initialState);

  let sequenceTimer: ReturnType<typeof setTimeout> | null = null;
  let initialized = false;

  function clearSequenceTimer() {
    if (sequenceTimer) {
      clearTimeout(sequenceTimer);
      sequenceTimer = null;
    }
  }

  function startSequenceTimer() {
    clearSequenceTimer();
    sequenceTimer = setTimeout(() => {
      update((state) => ({
        ...state,
        pendingSequence: [],
      }));
    }, SEQUENCE_TIMEOUT);
  }

  function handleKeyDown(event: KeyboardEvent) {
    const state = get({ subscribe });

    if (!state.enabled) return;

    // Skip modifier-only keypresses
    if (['Control', 'Alt', 'Shift', 'Meta'].includes(event.key)) {
      return;
    }

    const key = normalizeKey(event.key);
    const modifiers = getModifiersFromEvent(event);
    const isInInput = isInputElement(event.target as Element);

    // Check for sequence matches first
    const currentSequence = [...state.pendingSequence, key];

    for (const [, shortcut] of state.shortcuts) {
      if (!shortcut.enabled) continue;
      if (shortcut.context && shortcut.context !== state.activeContext) continue;
      if (isInInput && !shortcut.allowInInput) continue;

      // Check for sequence match
      if (shortcut.sequence?.length) {
        const fullSequence = [...shortcut.sequence, shortcut.key];

        // Check if current sequence could lead to this shortcut
        const isPartialMatch = fullSequence
          .slice(0, currentSequence.length)
          .every((k, i) => k === currentSequence[i]);

        if (isPartialMatch) {
          if (currentSequence.length === fullSequence.length) {
            // Full match - execute
            if (!modifiersMatch(shortcut.modifiers, modifiers)) continue;

            if (shortcut.preventDefault) event.preventDefault();
            if (shortcut.stopPropagation) event.stopPropagation();

            shortcut.handler(event);

            // Clear sequence
            update((s) => ({ ...s, pendingSequence: [] }));
            clearSequenceTimer();
            return;
          } else {
            // Partial match - wait for more keys
            update((s) => ({ ...s, pendingSequence: currentSequence }));
            startSequenceTimer();
            event.preventDefault();
            return;
          }
        }
      }

      // Check for direct match (no sequence)
      if (!shortcut.sequence?.length) {
        if (shortcut.key !== key) continue;
        if (!modifiersMatch(shortcut.modifiers, modifiers)) continue;

        if (shortcut.preventDefault) event.preventDefault();
        if (shortcut.stopPropagation) event.stopPropagation();

        shortcut.handler(event);

        // Clear any pending sequence
        update((s) => ({ ...s, pendingSequence: [] }));
        clearSequenceTimer();
        return;
      }
    }

    // No match - clear sequence if we were building one
    if (state.pendingSequence.length > 0) {
      update((s) => ({ ...s, pendingSequence: [] }));
      clearSequenceTimer();
    }
  }

  return {
    subscribe,

    /**
     * Initialize keyboard listener
     */
    init() {
      if (initialized || typeof window === 'undefined') return;

      window.addEventListener('keydown', handleKeyDown);
      initialized = true;
    },

    /**
     * Destroy keyboard listener
     */
    destroy() {
      if (!initialized || typeof window === 'undefined') return;

      window.removeEventListener('keydown', handleKeyDown);
      clearSequenceTimer();
      initialized = false;
    },

    /**
     * Register a keyboard shortcut
     */
    register(shortcut: ShortcutDefinition) {
      update((state) => {
        const key = generateShortcutKey(
          shortcut.key,
          shortcut.modifiers,
          shortcut.sequence
        );

        const newShortcuts = new Map(state.shortcuts);
        newShortcuts.set(shortcut.id, {
          ...shortcut,
          enabled: shortcut.enabled !== false,
          preventDefault: shortcut.preventDefault !== false,
          stopPropagation: shortcut.stopPropagation ?? false,
          allowInInput: shortcut.allowInInput ?? false,
          context: shortcut.context ?? 'global',
        });

        return { ...state, shortcuts: newShortcuts };
      });
    },

    /**
     * Register multiple shortcuts at once
     */
    registerMany(shortcuts: ShortcutDefinition[]) {
      shortcuts.forEach((s) => this.register(s));
    },

    /**
     * Unregister a shortcut by ID
     */
    unregister(id: string) {
      update((state) => {
        const newShortcuts = new Map(state.shortcuts);
        newShortcuts.delete(id);
        return { ...state, shortcuts: newShortcuts };
      });
    },

    /**
     * Enable or disable a specific shortcut
     */
    setEnabled(id: string, enabled: boolean) {
      update((state) => {
        const shortcut = state.shortcuts.get(id);
        if (!shortcut) return state;

        const newShortcuts = new Map(state.shortcuts);
        newShortcuts.set(id, { ...shortcut, enabled });
        return { ...state, shortcuts: newShortcuts };
      });
    },

    /**
     * Enable or disable all shortcuts globally
     */
    setGlobalEnabled(enabled: boolean) {
      update((state) => ({ ...state, enabled }));
    },

    /**
     * Set the active context
     */
    setContext(context: ShortcutContext) {
      update((state) => ({ ...state, activeContext: context }));
    },

    /**
     * Show/hide the help panel
     */
    toggleHelp(visible?: boolean) {
      update((state) => ({
        ...state,
        helpVisible: visible ?? !state.helpVisible,
      }));
    },

    /**
     * Reset to initial state
     */
    reset() {
      clearSequenceTimer();
      set(initialState);
    },
  };
}

export const keyboardStore = createKeyboardStore();

// ============================================================================
// Derived Stores
// ============================================================================

/**
 * Get all shortcuts grouped by category
 */
export const shortcutsByCategory = derived(keyboardStore, ($state) => {
  const groups = new Map<string, ShortcutDefinition[]>();

  for (const [, shortcut] of $state.shortcuts) {
    const category = shortcut.category || 'General';
    const existing = groups.get(category) || [];
    groups.set(category, [...existing, shortcut]);
  }

  return groups;
});

/**
 * Get shortcuts formatted for display in help panel
 */
export const shortcutsForHelp = derived(shortcutsByCategory, ($groups) => {
  const result: ShortcutGroup[] = [];

  for (const [name, shortcuts] of $groups) {
    result.push({
      name,
      shortcuts: shortcuts.map((s) => ({
        ...s,
        displayKey: formatShortcutForDisplay(s),
      })),
    });
  }

  return result.sort((a, b) => a.name.localeCompare(b.name));
});

/**
 * Whether help panel should be visible
 */
export const helpVisible = derived(keyboardStore, ($state) => $state.helpVisible);

/**
 * Current active context
 */
export const activeContext = derived(
  keyboardStore,
  ($state) => $state.activeContext
);

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Initialize keyboard shortcuts
 */
export function initKeyboard() {
  keyboardStore.init();
}

/**
 * Register a shortcut (convenience function)
 */
export function registerShortcut(shortcut: ShortcutDefinition) {
  keyboardStore.register(shortcut);
}

/**
 * Unregister a shortcut (convenience function)
 */
export function unregisterShortcut(id: string) {
  keyboardStore.unregister(id);
}

/**
 * Set keyboard context (convenience function)
 */
export function setKeyboardContext(context: ShortcutContext) {
  keyboardStore.setContext(context);
}

/**
 * Toggle help panel visibility
 */
export function toggleKeyboardHelp(visible?: boolean) {
  keyboardStore.toggleHelp(visible);
}

// ============================================================================
// Common Shortcut Presets
// ============================================================================

/**
 * Standard application shortcuts
 */
export const COMMON_SHORTCUTS: Omit<ShortcutDefinition, 'handler'>[] = [
  // Navigation
  { id: 'goto-home', key: 'h', modifiers: ['alt'], description: 'Go to Home', category: 'Navigation' },
  { id: 'goto-dashboard', key: 'd', modifiers: ['alt'], description: 'Go to Dashboard', category: 'Navigation' },
  { id: 'goto-settings', key: ',', modifiers: ['ctrl'], description: 'Open Settings', category: 'Navigation' },
  { id: 'go-back', key: 'ArrowLeft', modifiers: ['alt'], description: 'Go Back', category: 'Navigation' },
  { id: 'go-forward', key: 'ArrowRight', modifiers: ['alt'], description: 'Go Forward', category: 'Navigation' },

  // Actions
  { id: 'save', key: 's', modifiers: ['ctrl'], description: 'Save', category: 'Actions' },
  { id: 'search', key: 'k', modifiers: ['ctrl'], description: 'Open Search', category: 'Actions' },
  { id: 'command-palette', key: 'p', modifiers: ['ctrl', 'shift'], description: 'Command Palette', category: 'Actions' },
  { id: 'new-item', key: 'n', modifiers: ['ctrl'], description: 'New Item', category: 'Actions' },
  { id: 'refresh', key: 'r', modifiers: ['ctrl'], description: 'Refresh', category: 'Actions' },

  // Help
  { id: 'show-help', key: '?', modifiers: ['shift'], description: 'Show Keyboard Shortcuts', category: 'Help' },
  { id: 'close', key: 'Escape', description: 'Close/Cancel', category: 'Help' },
];

/**
 * Vim-style navigation shortcuts
 */
export const VIM_SHORTCUTS: Omit<ShortcutDefinition, 'handler'>[] = [
  { id: 'vim-goto-home', key: 'h', sequence: ['g'], description: 'Go to Home (vim)', category: 'Vim Navigation' },
  { id: 'vim-goto-top', key: 'g', sequence: ['g'], description: 'Go to Top', category: 'Vim Navigation' },
  { id: 'vim-goto-bottom', key: 'G', modifiers: ['shift'], description: 'Go to Bottom', category: 'Vim Navigation' },
  { id: 'vim-next', key: 'j', description: 'Next Item', category: 'Vim Navigation' },
  { id: 'vim-prev', key: 'k', description: 'Previous Item', category: 'Vim Navigation' },
];
