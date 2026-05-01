<script lang="ts">
  // ─── Props ──────────────────────────────────────────────────────────────────
  interface FieldRef {
    id: string;
    label: string;
    data_type: string;
  }

  interface FunctionDef {
    name: string;
    description: string;
    category: string;
    parameters: { name: string; type: string }[];
  }

  interface Props {
    value: string;
    fields: FieldRef[];
    functions: FunctionDef[];
    placeholder?: string;
    class?: string;
    onchange?: (e: CustomEvent<{ value: string }>) => void;
    onvalidate?: (e: CustomEvent<{ isValid: boolean; errorMessage: string }>) => void;
  }

  let {
    value = $bindable(''),
    fields,
    functions: funcs,
    placeholder = 'Enter expression...',
    class: className = '',
    onchange,
    onvalidate,
  }: Props = $props();

  // ─── State ──────────────────────────────────────────────────────────────────
  let textareaRef: HTMLTextAreaElement | undefined = $state(undefined);
  let showAutocomplete = $state(false);
  let autocompleteItems = $state<{ type: 'field' | 'function'; label: string; description: string; insert: string }[]>([]);
  let autocompleteIndex = $state(0);
  let autocompletePos = $state({ top: 0, left: 0 });
  let showFunctionPicker = $state(false);
  let errors = $state<{ start: number; end: number; message: string }[]>([]);

  // ─── Derived ────────────────────────────────────────────────────────────────
  let functionCategories = $derived.by(() => {
    const cats = new Map<string, FunctionDef[]>();
    for (const fn of funcs) {
      if (!cats.has(fn.category)) cats.set(fn.category, []);
      cats.get(fn.category)!.push(fn);
    }
    return cats;
  });

  let highlightedHtml = $derived.by(() => {
    if (!value) return '';
    let html = escapeHtml(value);

    // Highlight field references [FieldName]
    html = html.replace(/\[([^\]]+)\]/g, (match, name) => {
      const found = fields.some(f => f.label === name || f.id === name);
      if (found) {
        return `<span class="bi-expr__hl-field">${match}</span>`;
      }
      return `<span class="bi-expr__hl-error">${match}</span>`;
    });

    // Highlight function names
    const funcNames = funcs.map(f => f.name).join('|');
    if (funcNames) {
      const re = new RegExp(`\\b(${funcNames})\\s*(?=\\()`, 'g');
      html = html.replace(re, '<span class="bi-expr__hl-func">$1</span>');
    }

    // Highlight strings
    html = html.replace(/"([^"]*?)"/g, '<span class="bi-expr__hl-string">"$1"</span>');
    html = html.replace(/'([^']*?)'/g, '<span class="bi-expr__hl-string">\'$1\'</span>');

    // Highlight numbers
    html = html.replace(/\b(\d+\.?\d*)\b/g, '<span class="bi-expr__hl-number">$1</span>');

    return html;
  });

  // ─── Autocomplete ──────────────────────────────────────────────────────────
  function updateAutocomplete() {
    if (!textareaRef) return;

    const pos = textareaRef.selectionStart;
    const textBefore = value.slice(0, pos);

    // Check for [ trigger (field reference)
    const bracketMatch = textBefore.match(/\[([^\]]*)$/);
    if (bracketMatch) {
      const query = bracketMatch[1].toLowerCase();
      autocompleteItems = fields
        .filter(f => f.label.toLowerCase().includes(query) || f.id.toLowerCase().includes(query))
        .map(f => ({
          type: 'field' as const,
          label: f.label,
          description: f.data_type,
          insert: `${f.label}]`,
        }));
      showAutocomplete = autocompleteItems.length > 0;
      autocompleteIndex = 0;
      positionAutocomplete();
      return;
    }

    // Check for word trigger (function name)
    const wordMatch = textBefore.match(/\b([A-Za-z_]\w*)$/);
    if (wordMatch && wordMatch[1].length >= 2) {
      const query = wordMatch[1].toLowerCase();
      autocompleteItems = funcs
        .filter(f => f.name.toLowerCase().startsWith(query))
        .map(f => ({
          type: 'function' as const,
          label: f.name,
          description: f.description,
          insert: `${f.name}(${f.parameters.map(p => p.name).join(', ')})`,
        }));
      showAutocomplete = autocompleteItems.length > 0;
      autocompleteIndex = 0;
      positionAutocomplete();
      return;
    }

    showAutocomplete = false;
  }

  function positionAutocomplete() {
    if (!textareaRef) return;
    const rect = textareaRef.getBoundingClientRect();
    // Approximate position using character count
    const lines = value.slice(0, textareaRef.selectionStart).split('\n');
    const lineNum = lines.length - 1;
    const lineHeight = 20;
    autocompletePos = {
      top: (lineNum + 1) * lineHeight + 4,
      left: Math.min(lines[lineNum].length * 7.5, rect.width - 240),
    };
  }

  function applyAutocomplete(item: typeof autocompleteItems[0]) {
    if (!textareaRef) return;
    const pos = textareaRef.selectionStart;
    const textBefore = value.slice(0, pos);

    let replaceFrom = pos;
    if (item.type === 'field') {
      const bracketPos = textBefore.lastIndexOf('[');
      if (bracketPos >= 0) replaceFrom = bracketPos + 1;
    } else {
      const match = textBefore.match(/\b([A-Za-z_]\w*)$/);
      if (match) replaceFrom = pos - match[1].length;
    }

    const before = value.slice(0, replaceFrom);
    const after = value.slice(pos);
    value = before + item.insert + after;
    showAutocomplete = false;

    // Emit change
    onchange?.(new CustomEvent('change', { detail: { value } }));
  }

  function handleKeydown(e: KeyboardEvent) {
    if (showAutocomplete) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        autocompleteIndex = (autocompleteIndex + 1) % autocompleteItems.length;
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        autocompleteIndex = (autocompleteIndex - 1 + autocompleteItems.length) % autocompleteItems.length;
      } else if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault();
        applyAutocomplete(autocompleteItems[autocompleteIndex]);
      } else if (e.key === 'Escape') {
        showAutocomplete = false;
      }
    }
  }

  function handleInput() {
    updateAutocomplete();
    onchange?.(new CustomEvent('change', { detail: { value } }));
  }

  function handleBlur() {
    // Delay to allow autocomplete click
    setTimeout(() => {
      showAutocomplete = false;
      validate();
    }, 200);
  }

  function validate() {
    const foundErrors: typeof errors = [];

    // Check for unknown field references
    const fieldRefs = value.matchAll(/\[([^\]]+)\]/g);
    for (const match of fieldRefs) {
      const name = match[1];
      const found = fields.some(f => f.label === name || f.id === name);
      if (!found) {
        foundErrors.push({
          start: match.index!,
          end: match.index! + match[0].length,
          message: `Unknown field: ${name}`,
        });
      }
    }

    errors = foundErrors;

    const isValid = foundErrors.length === 0;
    const errorMessage = foundErrors.map(e => e.message).join('; ');
    onvalidate?.(new CustomEvent('validate', { detail: { isValid, errorMessage } }));
  }

  function insertFunction(fn: FunctionDef) {
    const template = `${fn.name}(${fn.parameters.map(p => p.name).join(', ')})`;
    if (textareaRef) {
      const pos = textareaRef.selectionStart;
      value = value.slice(0, pos) + template + value.slice(pos);
      textareaRef.focus();
      const newPos = pos + template.length;
      textareaRef.setSelectionRange(newPos, newPos);
    } else {
      value += template;
    }
    onchange?.(new CustomEvent('change', { detail: { value } }));
  }

  function escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  let collapsedCategories = $state<Set<string>>(new Set());

  function toggleCategory(cat: string) {
    const next = new Set(collapsedCategories);
    if (next.has(cat)) next.delete(cat);
    else next.add(cat);
    collapsedCategories = next;
  }
</script>

<div class="bi-expr-editor {className}">
  <div class="bi-expr-editor__main">
    <!-- Editor area -->
    <div class="bi-expr-editor__editor">
      <!-- Syntax highlight overlay -->
      <div class="bi-expr-editor__highlight" aria-hidden="true">
        {@html highlightedHtml}&nbsp;
      </div>

      <!-- Actual textarea -->
      <textarea
        bind:this={textareaRef}
        bind:value
        {placeholder}
        class="bi-expr-editor__textarea"
        spellcheck="false"
        autocomplete="off"
        oninput={handleInput}
        onkeydown={handleKeydown}
        onblur={handleBlur}
        rows="4"
      ></textarea>

      <!-- Autocomplete dropdown -->
      {#if showAutocomplete && autocompleteItems.length > 0}
        <div
          class="bi-expr-editor__autocomplete"
          style="top: {autocompletePos.top}px; left: {autocompletePos.left}px;"
        >
          {#each autocompleteItems as item, i (item.label + item.type)}
            <button
              class="bi-expr-editor__ac-item"
              class:bi-expr-editor__ac-item--active={i === autocompleteIndex}
              onmousedown={(e: MouseEvent) => { e.preventDefault(); applyAutocomplete(item); }}
              onmouseenter={() => autocompleteIndex = i}
            >
              <span class="bi-expr-editor__ac-icon" class:bi-expr-editor__ac-icon--field={item.type === 'field'} class:bi-expr-editor__ac-icon--func={item.type === 'function'}>
                {item.type === 'field' ? '[]' : 'fn'}
              </span>
              <span class="bi-expr-editor__ac-label">{item.label}</span>
              <span class="bi-expr-editor__ac-desc">{item.description}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Validation errors -->
    {#if errors.length > 0}
      <div class="bi-expr-editor__errors">
        {#each errors as err}
          <span class="bi-expr-editor__error">{err.message}</span>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Function picker sidebar -->
  <div class="bi-expr-editor__sidebar">
    <div class="bi-expr-editor__sidebar-header">
      <button
        class="bi-expr-editor__sidebar-toggle"
        onclick={() => showFunctionPicker = !showFunctionPicker}
        aria-expanded={showFunctionPicker}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
          <path d="M4 6h16M4 12h10M4 18h14"/>
        </svg>
        Functions
      </button>
    </div>

    {#if showFunctionPicker}
      <div class="bi-expr-editor__func-list">
        {#each [...functionCategories.entries()] as [category, categoryFuncs]}
          <button
            class="bi-expr-editor__func-cat"
            onclick={() => toggleCategory(category)}
            aria-expanded={!collapsedCategories.has(category)}
          >
            <svg
              class="bi-expr-editor__func-chevron"
              class:bi-expr-editor__func-chevron--collapsed={collapsedCategories.has(category)}
              viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"
            >
              <path d="m6 9 6 6 6-6"/>
            </svg>
            {category}
          </button>

          {#if !collapsedCategories.has(category)}
            {#each categoryFuncs as fn (fn.name)}
              <button
                class="bi-expr-editor__func-item"
                onclick={() => insertFunction(fn)}
                title={fn.description}
              >
                <span class="bi-expr-editor__func-name">{fn.name}</span>
                <span class="bi-expr-editor__func-sig">
                  ({fn.parameters.map(p => p.name).join(', ')})
                </span>
              </button>
            {/each}
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .bi-expr-editor {
    display: flex;
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    background: hsl(var(--background));
    overflow: hidden;
  }

  .bi-expr-editor__main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .bi-expr-editor__editor {
    position: relative;
    flex: 1;
  }

  .bi-expr-editor__highlight,
  .bi-expr-editor__textarea {
    font-family: 'Fira Code', 'Cascadia Code', 'JetBrains Mono', monospace;
    font-size: 0.8125rem;
    line-height: 1.5;
    padding: 0.75rem;
    white-space: pre-wrap;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .bi-expr-editor__highlight {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    color: transparent;
    overflow: hidden;
  }

  :global(.bi-expr__hl-field) {
    color: hsl(var(--primary));
    font-weight: 500;
  }

  :global(.bi-expr__hl-func) {
    color: hsl(260 60% 60%);
    font-weight: 600;
  }

  :global(.bi-expr__hl-string) {
    color: hsl(140 60% 40%);
  }

  :global(.bi-expr__hl-number) {
    color: hsl(30 90% 50%);
  }

  :global(.bi-expr__hl-error) {
    color: hsl(var(--destructive));
    text-decoration: wavy underline hsl(var(--destructive));
    text-underline-offset: 3px;
  }

  .bi-expr-editor__textarea {
    position: relative;
    width: 100%;
    min-height: 6rem;
    border: none;
    background: transparent;
    color: hsl(var(--foreground));
    outline: none;
    resize: vertical;
    caret-color: hsl(var(--foreground));
    z-index: 1;
  }

  .bi-expr-editor__textarea::placeholder {
    color: hsl(var(--muted-foreground));
  }

  /* Autocomplete */
  .bi-expr-editor__autocomplete {
    position: absolute;
    z-index: 50;
    min-width: 15rem;
    max-height: 12rem;
    overflow-y: auto;
    background: hsl(var(--popover));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    box-shadow: 0 4px 16px hsl(var(--foreground) / 0.12);
    padding: 0.25rem;
  }

  .bi-expr-editor__ac-item {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    width: 100%;
    padding: 0.375rem 0.5rem;
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    font-size: 0.8125rem;
    color: hsl(var(--popover-foreground));
  }

  .bi-expr-editor__ac-item:hover,
  .bi-expr-editor__ac-item--active {
    background: hsl(var(--accent));
  }

  .bi-expr-editor__ac-icon {
    flex-shrink: 0;
    width: 1.5rem;
    text-align: center;
    font-size: 0.625rem;
    font-weight: 700;
    font-family: monospace;
    padding: 0.125rem 0;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-expr-editor__ac-icon--field {
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
  }

  .bi-expr-editor__ac-icon--func {
    background: hsl(260 60% 60% / 0.15);
    color: hsl(260 60% 60%);
  }

  .bi-expr-editor__ac-label {
    flex: 1;
    font-weight: 500;
  }

  .bi-expr-editor__ac-desc {
    font-size: 0.6875rem;
    color: hsl(var(--muted-foreground));
    max-width: 8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Errors */
  .bi-expr-editor__errors {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    padding: 0.375rem 0.75rem;
    border-top: 1px solid hsl(var(--destructive) / 0.3);
    background: hsl(var(--destructive) / 0.05);
  }

  .bi-expr-editor__error {
    font-size: 0.75rem;
    color: hsl(var(--destructive));
  }

  /* Function picker sidebar */
  .bi-expr-editor__sidebar {
    display: flex;
    flex-direction: column;
    width: auto;
    border-left: 1px solid hsl(var(--border));
    background: hsl(var(--muted));
  }

  .bi-expr-editor__sidebar-header {
    padding: 0.375rem;
    border-bottom: 1px solid hsl(var(--border));
  }

  .bi-expr-editor__sidebar-toggle {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.5rem;
    border: none;
    background: transparent;
    font-size: 0.75rem;
    font-weight: 600;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    white-space: nowrap;
  }

  .bi-expr-editor__sidebar-toggle:hover {
    background: hsl(var(--accent));
    color: hsl(var(--accent-foreground));
  }

  .bi-expr-editor__func-list {
    flex: 1;
    overflow-y: auto;
    width: 12rem;
    padding: 0.25rem;
  }

  .bi-expr-editor__func-cat {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    width: 100%;
    padding: 0.375rem 0.5rem;
    border: none;
    background: transparent;
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    text-align: left;
  }

  .bi-expr-editor__func-cat:hover {
    color: hsl(var(--foreground));
  }

  .bi-expr-editor__func-chevron {
    transition: transform 0.15s ease;
  }

  .bi-expr-editor__func-chevron--collapsed {
    transform: rotate(-90deg);
  }

  .bi-expr-editor__func-item {
    display: flex;
    align-items: baseline;
    gap: 0.125rem;
    width: 100%;
    padding: 0.25rem 0.5rem 0.25rem 1.25rem;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    text-align: left;
  }

  .bi-expr-editor__func-item:hover {
    background: hsl(var(--accent));
  }

  .bi-expr-editor__func-name {
    font-size: 0.75rem;
    font-weight: 500;
    color: hsl(260 60% 60%);
    font-family: monospace;
  }

  .bi-expr-editor__func-sig {
    font-size: 0.625rem;
    color: hsl(var(--muted-foreground));
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
