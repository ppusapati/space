<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import {
    type RichTextToolbarItem,
    type RichTextOutputFormat,
    type RichTextEditorMode,
    type RichTextEditorInstance,
    toolbarPresets,
    richTextClasses,
    richTextSizeClasses,
    toolbarItemLabels,
    getCharacterCount,
    getWordCount,
    sanitizeHtml,
  } from './richtext.types';
  import type { Size, ValidationState } from '../types';

  // Props
  export let value: string = '';
  export let outputFormat: RichTextOutputFormat = 'html';
  export let toolbar: RichTextToolbarItem[] | undefined = undefined;
  export let toolbarPreset: 'minimal' | 'basic' | 'standard' | 'full' = 'standard';
  export let placeholder: string = '';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let height: string = '';
  export let minHeight: string = '150px';
  export let maxHeight: string = '';
  export let autofocus: boolean = false;
  export let spellcheck: boolean = true;
  export let readonly: boolean = false;
  export let disabled: boolean = false;
  export let maxLength: number | undefined = undefined;
  export let showCount: boolean = false;
  export let required: boolean = false;
  export let id: string = uid('richtext');
  export let testId: string = '';
  export let onImageUpload: ((file: File) => Promise<string>) | undefined = undefined;

  let className: string = '';
  export { className as class };

  // Refs and state
  let editorRef: HTMLDivElement;
  let isFocused = false;
  let currentMode: RichTextEditorMode = 'wysiwyg';
  let activeFormats: Set<string> = new Set();

  const dispatch = createEventDispatcher<{
    input: { value: string; html: string };
    change: { value: string; html: string };
    focus: { event: FocusEvent };
    blur: { event: FocusEvent };
    ready: { instance: RichTextEditorInstance };
  }>();

  // Get effective toolbar items
  $: effectiveToolbar = toolbar || toolbarPresets[toolbarPreset] || toolbarPresets.standard;

  // Computed classes
  $: containerClasses = cn(
    richTextClasses.container,
    isFocused && richTextClasses.containerFocused,
    disabled && richTextClasses.containerDisabled,
    state === 'invalid' || errorText ? richTextClasses.containerInvalid : '',
    state === 'valid' ? richTextClasses.containerValid : '',
    className
  );

  $: editorClasses = cn(
    richTextClasses.editor,
    richTextSizeClasses[size],
    !value && placeholder && richTextClasses.editorPlaceholder
  );

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;
  $: helperClasses = cn(
    'mt-1 text-sm',
    errorText ? 'text-semantic-error-500' : 'text-neutral-500'
  );

  $: charCount = getCharacterCount(value);
  $: wordCount = getWordCount(value);
  $: charCountText = maxLength ? `${charCount}/${maxLength}` : `${charCount} chars`;

  // Toolbar icon mapping
  const toolbarIcons: Record<RichTextToolbarItem, string> = {
    bold: 'bold',
    italic: 'italic',
    underline: 'underline',
    strike: 'strikethrough',
    code: 'code',
    heading1: 'heading-1',
    heading2: 'heading-2',
    heading3: 'heading-3',
    paragraph: 'paragraph',
    bulletList: 'list',
    orderedList: 'list-ordered',
    taskList: 'list-check',
    blockquote: 'quote',
    codeBlock: 'code-block',
    horizontalRule: 'minus',
    link: 'link',
    image: 'image',
    'align-left': 'align-left',
    'align-center': 'align-center',
    'align-right': 'align-right',
    'align-justify': 'align-justify',
    subscript: 'subscript',
    superscript: 'superscript',
    highlight: 'highlighter',
    textColor: 'palette',
    undo: 'undo',
    redo: 'redo',
    clear: 'eraser',
    divider: '',
  };

  // Command mapping
  const commandMap: Record<string, { command: string; value?: string; block?: boolean }> = {
    bold: { command: 'bold' },
    italic: { command: 'italic' },
    underline: { command: 'underline' },
    strike: { command: 'strikeThrough' },
    code: { command: 'insertHTML', value: '<code>' },
    heading1: { command: 'formatBlock', value: 'h1', block: true },
    heading2: { command: 'formatBlock', value: 'h2', block: true },
    heading3: { command: 'formatBlock', value: 'h3', block: true },
    paragraph: { command: 'formatBlock', value: 'p', block: true },
    bulletList: { command: 'insertUnorderedList' },
    orderedList: { command: 'insertOrderedList' },
    blockquote: { command: 'formatBlock', value: 'blockquote', block: true },
    horizontalRule: { command: 'insertHorizontalRule' },
    'align-left': { command: 'justifyLeft' },
    'align-center': { command: 'justifyCenter' },
    'align-right': { command: 'justifyRight' },
    'align-justify': { command: 'justifyFull' },
    subscript: { command: 'subscript' },
    superscript: { command: 'superscript' },
    undo: { command: 'undo' },
    redo: { command: 'redo' },
    clear: { command: 'removeFormat' },
  };

  // Execute editor command
  function executeCommand(item: RichTextToolbarItem) {
    if (disabled || readonly) return;

    editorRef?.focus();

    const mapping = commandMap[item];
    if (mapping) {
      if (mapping.block && mapping.value) {
        document.execCommand(mapping.command, false, mapping.value);
      } else if (mapping.value) {
        document.execCommand(mapping.command, false, mapping.value);
      } else {
        document.execCommand(mapping.command, false);
      }
      updateContent();
      updateActiveFormats();
      return;
    }

    // Special handling for specific items
    switch (item) {
      case 'link':
        insertLink();
        break;
      case 'image':
        insertImage();
        break;
      case 'codeBlock':
        insertCodeBlock();
        break;
      case 'taskList':
        insertTaskList();
        break;
      case 'highlight':
        applyHighlight();
        break;
      case 'textColor':
        applyTextColor();
        break;
    }
  }

  function insertLink() {
    const url = prompt('Enter URL:');
    if (url) {
      document.execCommand('createLink', false, url);
      updateContent();
    }
  }

  async function insertImage() {
    if (onImageUpload) {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = 'image/*';
      input.onchange = async (e) => {
        const file = (e.target as HTMLInputElement).files?.[0];
        if (file) {
          try {
            const url = await onImageUpload(file);
            document.execCommand('insertImage', false, url);
            updateContent();
          } catch (err) {
            console.error('Image upload failed:', err);
          }
        }
      };
      input.click();
    } else {
      const url = prompt('Enter image URL:');
      if (url) {
        document.execCommand('insertImage', false, url);
        updateContent();
      }
    }
  }

  function insertCodeBlock() {
    const selection = window.getSelection();
    if (selection && selection.rangeCount > 0) {
      const range = selection.getRangeAt(0);
      const pre = document.createElement('pre');
      const code = document.createElement('code');
      code.textContent = selection.toString() || 'code here';
      pre.appendChild(code);
      range.deleteContents();
      range.insertNode(pre);
      updateContent();
    }
  }

  function insertTaskList() {
    const html = '<ul class="task-list"><li><input type="checkbox" /> Task item</li></ul>';
    document.execCommand('insertHTML', false, html);
    updateContent();
  }

  function applyHighlight() {
    document.execCommand('hiliteColor', false, '#ffff00');
    updateContent();
  }

  function applyTextColor() {
    const color = prompt('Enter color (hex or name):', '#000000');
    if (color) {
      document.execCommand('foreColor', false, color);
      updateContent();
    }
  }

  // Update active formats based on selection
  function updateActiveFormats() {
    const newFormats = new Set<string>();

    if (document.queryCommandState('bold')) newFormats.add('bold');
    if (document.queryCommandState('italic')) newFormats.add('italic');
    if (document.queryCommandState('underline')) newFormats.add('underline');
    if (document.queryCommandState('strikeThrough')) newFormats.add('strike');
    if (document.queryCommandState('subscript')) newFormats.add('subscript');
    if (document.queryCommandState('superscript')) newFormats.add('superscript');
    if (document.queryCommandState('insertUnorderedList')) newFormats.add('bulletList');
    if (document.queryCommandState('insertOrderedList')) newFormats.add('orderedList');
    if (document.queryCommandState('justifyLeft')) newFormats.add('align-left');
    if (document.queryCommandState('justifyCenter')) newFormats.add('align-center');
    if (document.queryCommandState('justifyRight')) newFormats.add('align-right');
    if (document.queryCommandState('justifyFull')) newFormats.add('align-justify');

    // Check block format
    const block = document.queryCommandValue('formatBlock');
    if (block === 'h1') newFormats.add('heading1');
    if (block === 'h2') newFormats.add('heading2');
    if (block === 'h3') newFormats.add('heading3');
    if (block === 'blockquote') newFormats.add('blockquote');

    activeFormats = newFormats;
  }

  // Content update handlers
  function updateContent() {
    if (!editorRef) return;
    const html = editorRef.innerHTML;
    value = sanitizeHtml(html);
  }

  function handleInput() {
    updateContent();
    dispatch('input', { value, html: editorRef?.innerHTML || '' });
  }

  function handleFocus(event: FocusEvent) {
    isFocused = true;
    dispatch('focus', { event });
  }

  function handleBlur(event: FocusEvent) {
    isFocused = false;
    updateContent();
    dispatch('change', { value, html: editorRef?.innerHTML || '' });
    dispatch('blur', { event });
  }

  function handleKeyDown(event: KeyboardEvent) {
    // Handle keyboard shortcuts
    if (event.ctrlKey || event.metaKey) {
      switch (event.key.toLowerCase()) {
        case 'b':
          event.preventDefault();
          executeCommand('bold');
          break;
        case 'i':
          event.preventDefault();
          executeCommand('italic');
          break;
        case 'u':
          event.preventDefault();
          executeCommand('underline');
          break;
        case 'z':
          if (event.shiftKey) {
            event.preventDefault();
            executeCommand('redo');
          } else {
            event.preventDefault();
            executeCommand('undo');
          }
          break;
        case 'y':
          event.preventDefault();
          executeCommand('redo');
          break;
      }
    }
  }

  function handleSelectionChange() {
    updateActiveFormats();
  }

  function handlePaste(event: ClipboardEvent) {
    event.preventDefault();
    const text = event.clipboardData?.getData('text/plain') || '';
    document.execCommand('insertText', false, text);
    updateContent();
  }

  // Check if item is active
  function isItemActive(item: RichTextToolbarItem): boolean {
    return activeFormats.has(item);
  }

  // Editor instance methods
  const editorInstance: RichTextEditorInstance = {
    getHTML: () => editorRef?.innerHTML || '',
    getJSON: () => ({ type: 'doc', content: editorRef?.innerHTML || '' }),
    getText: () => editorRef?.innerText || '',
    setContent: (content: string) => {
      if (editorRef) {
        editorRef.innerHTML = sanitizeHtml(content);
        updateContent();
      }
    },
    clearContent: () => {
      if (editorRef) {
        editorRef.innerHTML = '';
        value = '';
      }
    },
    focus: () => editorRef?.focus(),
    blur: () => editorRef?.blur(),
    isEmpty: () => !editorRef?.innerText?.trim(),
    getCharacterCount: () => getCharacterCount(value),
    getWordCount: () => getWordCount(value),
    executeCommand: (cmd: string, attrs?: Record<string, unknown>) => {
      document.execCommand(cmd, false, attrs?.value as string);
      updateContent();
    },
    isActive: (cmd: string) => document.queryCommandState(cmd),
    canExecute: (cmd: string) => document.queryCommandEnabled(cmd),
    undo: () => executeCommand('undo'),
    redo: () => executeCommand('redo'),
  };

  // Export instance for parent access
  export function getInstance(): RichTextEditorInstance {
    return editorInstance;
  }

  onMount(() => {
    if (editorRef && value) {
      editorRef.innerHTML = sanitizeHtml(value);
    }

    document.addEventListener('selectionchange', handleSelectionChange);

    if (autofocus && editorRef) {
      editorRef.focus();
    }

    dispatch('ready', { instance: editorInstance });
  });

  onDestroy(() => {
    document.removeEventListener('selectionchange', handleSelectionChange);
  });

  // Watch for external value changes
  $: if (editorRef && value !== editorRef.innerHTML) {
    const currentHTML = editorRef.innerHTML;
    const sanitizedValue = sanitizeHtml(value);
    if (currentHTML !== sanitizedValue && !isFocused) {
      editorRef.innerHTML = sanitizedValue;
    }
  }
</script>

<div class="w-full">
  {#if label}
    <label for={id} class={richTextClasses.label}>
      {label}
      {#if required}
        <span class="text-semantic-error-500 ml-0.5" aria-hidden="true">*</span>
      {/if}
    </label>
  {/if}

  <div class={containerClasses} data-testid={testId || undefined}>
    <!-- Toolbar -->
    <div class={richTextClasses.toolbar} role="toolbar" aria-label="Formatting options">
      {#each effectiveToolbar as item}
        {#if item === 'divider'}
          <div class={richTextClasses.toolbarDivider}></div>
        {:else}
          <button
            type="button"
            class={cn(
              richTextClasses.toolbarButton,
              isItemActive(item) && richTextClasses.toolbarButtonActive,
              disabled && richTextClasses.toolbarButtonDisabled
            )}
            title={toolbarItemLabels[item]}
            aria-label={toolbarItemLabels[item]}
            aria-pressed={isItemActive(item)}
            disabled={disabled}
            on:click={() => executeCommand(item)}
            on:mousedown|preventDefault
          >
            <Icon name={toolbarIcons[item]} size="sm" />
          </button>
        {/if}
      {/each}
    </div>

    <!-- Editor Content -->
    <div
      bind:this={editorRef}
      {id}
      class={editorClasses}
      contenteditable={!disabled && !readonly}
      role="textbox"
      aria-multiline="true"
      aria-label={label || 'Rich text editor'}
      aria-invalid={state === 'invalid' || !!errorText}
      aria-describedby={displayedHelperText ? `${id}-helper` : undefined}
      aria-placeholder={placeholder}
      data-placeholder={placeholder}
      {spellcheck}
      style:height={height || undefined}
      style:min-height={minHeight}
      style:max-height={maxHeight || undefined}
      style:overflow-y={maxHeight ? 'auto' : undefined}
      on:input={handleInput}
      on:focus={handleFocus}
      on:blur={handleBlur}
      on:keydown={handleKeyDown}
      on:paste={handlePaste}
    ></div>

    <!-- Footer with character count -->
    {#if showCount}
      <div class={richTextClasses.footer}>
        <span>{wordCount} words</span>
        <span>{charCountText}</span>
      </div>
    {/if}
  </div>

  {#if displayedHelperText}
    <p id="{id}-helper" class={helperClasses}>
      {displayedHelperText}
    </p>
  {/if}
</div>

<style>
  /* Placeholder styling */
  [contenteditable]:empty:before {
    content: attr(data-placeholder);
    color: #9ca3af;
    pointer-events: none;
    display: block;
  }

  /* Editor prose styling */
  :global(.rich-text-editor [contenteditable]) {
    outline: none;
  }

  :global(.rich-text-editor [contenteditable] h1) {
    font-size: 2em;
    font-weight: bold;
    margin: 0.67em 0;
  }

  :global(.rich-text-editor [contenteditable] h2) {
    font-size: 1.5em;
    font-weight: bold;
    margin: 0.83em 0;
  }

  :global(.rich-text-editor [contenteditable] h3) {
    font-size: 1.17em;
    font-weight: bold;
    margin: 1em 0;
  }

  :global(.rich-text-editor [contenteditable] blockquote) {
    border-left: 4px solid #e5e7eb;
    padding-left: 1em;
    margin: 1em 0;
    color: #6b7280;
  }

  :global(.rich-text-editor [contenteditable] pre) {
    background: #f3f4f6;
    padding: 1em;
    border-radius: 0.375rem;
    overflow-x: auto;
    font-family: monospace;
  }

  :global(.rich-text-editor [contenteditable] code) {
    background: #f3f4f6;
    padding: 0.125em 0.25em;
    border-radius: 0.25rem;
    font-family: monospace;
    font-size: 0.875em;
  }

  :global(.rich-text-editor [contenteditable] pre code) {
    background: transparent;
    padding: 0;
  }

  :global(.rich-text-editor [contenteditable] ul),
  :global(.rich-text-editor [contenteditable] ol) {
    padding-left: 1.5em;
    margin: 1em 0;
  }

  :global(.rich-text-editor [contenteditable] ul) {
    list-style-type: disc;
  }

  :global(.rich-text-editor [contenteditable] ol) {
    list-style-type: decimal;
  }

  :global(.rich-text-editor [contenteditable] li) {
    margin: 0.25em 0;
  }

  :global(.rich-text-editor [contenteditable] a) {
    color: #2563eb;
    text-decoration: underline;
  }

  :global(.rich-text-editor [contenteditable] hr) {
    border: none;
    border-top: 1px solid #e5e7eb;
    margin: 1em 0;
  }

  :global(.rich-text-editor [contenteditable] img) {
    max-width: 100%;
    height: auto;
    border-radius: 0.375rem;
  }

  :global(.rich-text-editor .task-list) {
    list-style: none;
    padding-left: 0;
  }

  :global(.rich-text-editor .task-list li) {
    display: flex;
    align-items: flex-start;
    gap: 0.5em;
  }

  :global(.rich-text-editor .task-list input[type="checkbox"]) {
    margin-top: 0.25em;
  }
</style>
