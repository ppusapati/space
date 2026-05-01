/**
 * Rich Text Editor Types
 * Types and utilities for the RichTextEditor component
 */

import type { Size, ValidationState, BaseProps, DisableableProps } from '../types';

// ============================================================================
// Types
// ============================================================================

/** Toolbar item type */
export type RichTextToolbarItem =
  | 'bold'
  | 'italic'
  | 'underline'
  | 'strike'
  | 'code'
  | 'heading1'
  | 'heading2'
  | 'heading3'
  | 'paragraph'
  | 'bulletList'
  | 'orderedList'
  | 'taskList'
  | 'blockquote'
  | 'codeBlock'
  | 'horizontalRule'
  | 'link'
  | 'image'
  | 'align-left'
  | 'align-center'
  | 'align-right'
  | 'align-justify'
  | 'subscript'
  | 'superscript'
  | 'highlight'
  | 'textColor'
  | 'undo'
  | 'redo'
  | 'clear'
  | 'divider';

/** Toolbar group */
export type RichTextToolbarGroup = RichTextToolbarItem[];

/** Output format */
export type RichTextOutputFormat = 'html' | 'json' | 'text' | 'markdown';

/** Editor mode */
export type RichTextEditorMode = 'wysiwyg' | 'source' | 'preview';

/** RichTextEditor props */
export interface RichTextEditorProps extends BaseProps, DisableableProps {
  /** Current value (HTML string or JSON) */
  value?: string;
  /** Output format */
  outputFormat?: RichTextOutputFormat;
  /** Toolbar items/groups */
  toolbar?: RichTextToolbarItem[] | RichTextToolbarGroup[];
  /** Preset toolbar configuration */
  toolbarPreset?: 'minimal' | 'basic' | 'standard' | 'full';
  /** Placeholder text */
  placeholder?: string;
  /** Label */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error text */
  errorText?: string;
  /** Size variant */
  size?: Size;
  /** Validation state */
  state?: ValidationState;
  /** Editor height */
  height?: string;
  /** Min height */
  minHeight?: string;
  /** Max height */
  maxHeight?: string;
  /** Autofocus on mount */
  autofocus?: boolean;
  /** Enable spell check */
  spellcheck?: boolean;
  /** Read-only mode */
  readonly?: boolean;
  /** Character limit */
  maxLength?: number;
  /** Show character count */
  showCount?: boolean;
  /** Enable editor menu bar */
  menuBar?: boolean;
  /** Enable floating menu */
  floatingMenu?: boolean;
  /** Enable bubble menu (selection menu) */
  bubbleMenu?: boolean;
  /** Custom extensions configuration */
  extensions?: Record<string, unknown>;
  /** Upload handler for images */
  onImageUpload?: (file: File) => Promise<string>;
  /** Link handler */
  onLinkClick?: (url: string) => void;
}

/** Editor instance interface */
export interface RichTextEditorInstance {
  /** Get HTML content */
  getHTML: () => string;
  /** Get JSON content */
  getJSON: () => Record<string, unknown>;
  /** Get plain text */
  getText: () => string;
  /** Set content */
  setContent: (content: string) => void;
  /** Clear content */
  clearContent: () => void;
  /** Focus editor */
  focus: () => void;
  /** Blur editor */
  blur: () => void;
  /** Check if empty */
  isEmpty: () => boolean;
  /** Get character count */
  getCharacterCount: () => number;
  /** Get word count */
  getWordCount: () => number;
  /** Execute command */
  executeCommand: (command: string, attrs?: Record<string, unknown>) => void;
  /** Check if command is active */
  isActive: (command: string, attrs?: Record<string, unknown>) => boolean;
  /** Check if command can execute */
  canExecute: (command: string) => boolean;
  /** Undo */
  undo: () => void;
  /** Redo */
  redo: () => void;
}

// ============================================================================
// Toolbar Presets
// ============================================================================

export const toolbarPresets: Record<string, RichTextToolbarItem[]> = {
  minimal: ['bold', 'italic', 'link'],
  basic: ['bold', 'italic', 'underline', 'divider', 'bulletList', 'orderedList', 'divider', 'link'],
  standard: [
    'bold',
    'italic',
    'underline',
    'strike',
    'divider',
    'heading1',
    'heading2',
    'heading3',
    'divider',
    'bulletList',
    'orderedList',
    'taskList',
    'divider',
    'blockquote',
    'codeBlock',
    'horizontalRule',
    'divider',
    'link',
    'image',
    'divider',
    'undo',
    'redo',
  ],
  full: [
    'bold',
    'italic',
    'underline',
    'strike',
    'code',
    'divider',
    'heading1',
    'heading2',
    'heading3',
    'paragraph',
    'divider',
    'bulletList',
    'orderedList',
    'taskList',
    'divider',
    'blockquote',
    'codeBlock',
    'horizontalRule',
    'divider',
    'align-left',
    'align-center',
    'align-right',
    'align-justify',
    'divider',
    'subscript',
    'superscript',
    'highlight',
    'textColor',
    'divider',
    'link',
    'image',
    'divider',
    'undo',
    'redo',
    'clear',
  ],
};

// ============================================================================
// CSS Classes
// ============================================================================

export const richTextClasses = {
  container: 'rich-text-editor relative border border-neutral-300 rounded-md overflow-hidden',
  containerFocused: 'ring-2 ring-brand-primary-500 border-brand-primary-500',
  containerDisabled: 'opacity-50 cursor-not-allowed bg-neutral-50',
  containerInvalid: 'border-semantic-error-500',
  containerValid: 'border-semantic-success-500',

  toolbar: 'flex flex-wrap items-center gap-1 p-2 bg-neutral-50 border-b border-neutral-200',
  toolbarDivider: 'w-px h-6 bg-neutral-300 mx-1',
  toolbarButton:
    'p-1.5 rounded hover:bg-neutral-200 text-neutral-600 hover:text-neutral-900 transition-colors',
  toolbarButtonActive: 'bg-brand-primary-100 text-brand-primary-700 hover:bg-brand-primary-200',
  toolbarButtonDisabled: 'opacity-50 cursor-not-allowed hover:bg-transparent',

  editor: 'prose prose-sm max-w-none p-4 focus:outline-none min-h-[150px]',
  editorPlaceholder: 'text-neutral-400',

  footer: 'flex items-center justify-between px-3 py-2 bg-neutral-50 border-t border-neutral-200 text-xs text-neutral-500',

  label: 'block text-sm font-medium text-neutral-700 mb-1',
  helperText: 'mt-1 text-sm text-neutral-500',
  errorText: 'mt-1 text-sm text-semantic-error-500',
};

export const richTextSizeClasses: Record<Size, string> = {
  xs: 'text-xs',
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-lg',
  xl: 'text-xl',
};

// ============================================================================
// Toolbar Item Labels
// ============================================================================

export const toolbarItemLabels: Record<RichTextToolbarItem, string> = {
  bold: 'Bold',
  italic: 'Italic',
  underline: 'Underline',
  strike: 'Strikethrough',
  code: 'Inline Code',
  heading1: 'Heading 1',
  heading2: 'Heading 2',
  heading3: 'Heading 3',
  paragraph: 'Paragraph',
  bulletList: 'Bullet List',
  orderedList: 'Numbered List',
  taskList: 'Task List',
  blockquote: 'Blockquote',
  codeBlock: 'Code Block',
  horizontalRule: 'Horizontal Rule',
  link: 'Link',
  image: 'Image',
  'align-left': 'Align Left',
  'align-center': 'Align Center',
  'align-right': 'Align Right',
  'align-justify': 'Justify',
  subscript: 'Subscript',
  superscript: 'Superscript',
  highlight: 'Highlight',
  textColor: 'Text Color',
  undo: 'Undo',
  redo: 'Redo',
  clear: 'Clear Formatting',
  divider: '',
};

// ============================================================================
// Utilities
// ============================================================================

/** Get character count from HTML */
export function getCharacterCount(html: string): number {
  if (!html) return 0;
  const text = html.replace(/<[^>]*>/g, '');
  return text.length;
}

/** Get word count from HTML */
export function getWordCount(html: string): number {
  if (!html) return 0;
  const text = html.replace(/<[^>]*>/g, '').trim();
  if (!text) return 0;
  return text.split(/\s+/).length;
}

/** Convert HTML to plain text */
export function htmlToText(html: string): string {
  if (!html) return '';
  return html
    .replace(/<br\s*\/?>/gi, '\n')
    .replace(/<\/p>/gi, '\n')
    .replace(/<[^>]*>/g, '')
    .trim();
}

/** Sanitize HTML (basic) */
export function sanitizeHtml(html: string): string {
  if (!html) return '';
  // Remove script tags and event handlers
  return html
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    .replace(/on\w+="[^"]*"/gi, '')
    .replace(/on\w+='[^']*'/gi, '');
}
