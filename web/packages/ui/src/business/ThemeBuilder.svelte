<script context="module" lang="ts">
  export interface ThemeBuilderColors {
    primary: string;
    secondary: string;
    success: string;
    warning: string;
    error: string;
    info: string;
  }

  export interface ThemeBuilderConfig {
    name: string;
    mode: 'light' | 'dark';
    colors: ThemeBuilderColors;
    borderRadius: number;
    fontSize: number;
  }
</script>

<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import ColorPicker from '../forms/ColorPicker.svelte';
  import Switch from '../forms/Switch.svelte';
  import Slider from '../forms/Slider.svelte';

  // Props
  export let config: ThemeBuilderConfig = {
    name: 'Custom Theme',
    mode: 'light',
    colors: {
      primary: '#3b82f6',
      secondary: '#eab308',
      success: '#22c55e',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#0ea5e9',
    },
    borderRadius: 8,
    fontSize: 16,
  };

  export let presets: { id: string; name: string; colors: Partial<ThemeBuilderColors> }[] = [
    { id: 'ocean', name: 'Ocean', colors: { primary: '#0ea5e9', secondary: '#eab308' } },
    { id: 'forest', name: 'Forest', colors: { primary: '#22c55e', secondary: '#f59e0b' } },
    { id: 'sunset', name: 'Sunset', colors: { primary: '#f97316', secondary: '#8b5cf6' } },
    { id: 'midnight', name: 'Midnight', colors: { primary: '#8b5cf6', secondary: '#ec4899' } },
    { id: 'rose', name: 'Rose', colors: { primary: '#f43f5e', secondary: '#0ea5e9' } },
    { id: 'slate', name: 'Slate', colors: { primary: '#64748b', secondary: '#0ea5e9' } },
  ];

  export let showPreview: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { config: ThemeBuilderConfig };
    save: { config: ThemeBuilderConfig };
    reset: void;
    applyPreset: { presetId: string };
  }>();

  let activeTab: 'colors' | 'typography' | 'spacing' = 'colors';

  function handleColorChange(key: keyof ThemeBuilderColors, value: string) {
    config = {
      ...config,
      colors: {
        ...config.colors,
        [key]: value,
      },
    };
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function handleModeChange(mode: 'light' | 'dark') {
    config = { ...config, mode };
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function handleBorderRadiusChange(value: number) {
    config = { ...config, borderRadius: value };
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function handleFontSizeChange(value: number) {
    config = { ...config, fontSize: value };
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function applyPreset(presetId: string) {
    const preset = presets.find(p => p.id === presetId);
    if (!preset) return;

    config = {
      ...config,
      colors: {
        ...config.colors,
        ...preset.colors,
      },
    };
    dispatch('applyPreset', { presetId });
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function handleSave() {
    dispatch('save', { config });
  }

  function handleReset() {
    config = {
      name: 'Custom Theme',
      mode: 'light',
      colors: {
        primary: '#3b82f6',
        secondary: '#eab308',
        success: '#22c55e',
        warning: '#f59e0b',
        error: '#ef4444',
        info: '#0ea5e9',
      },
      borderRadius: 8,
      fontSize: 16,
    };
    dispatch('reset');
    dispatch('change', { config });
    applyPreviewStyles();
  }

  function applyPreviewStyles() {
    if (typeof document === 'undefined') return;

    const previewEl = document.getElementById('theme-builder-preview');
    if (!previewEl) return;

    previewEl.style.setProperty('--preview-primary', config.colors.primary);
    previewEl.style.setProperty('--preview-secondary', config.colors.secondary);
    previewEl.style.setProperty('--preview-success', config.colors.success);
    previewEl.style.setProperty('--preview-warning', config.colors.warning);
    previewEl.style.setProperty('--preview-error', config.colors.error);
    previewEl.style.setProperty('--preview-info', config.colors.info);
    previewEl.style.setProperty('--preview-radius', `${config.borderRadius}px`);
    previewEl.style.setProperty('--preview-font-size', `${config.fontSize}px`);
    previewEl.style.setProperty('--preview-bg', config.mode === 'light' ? '#ffffff' : '#1f2937');
    previewEl.style.setProperty('--preview-text', config.mode === 'light' ? '#1f2937' : '#f9fafb');
    previewEl.style.setProperty('--preview-border', config.mode === 'light' ? '#e5e7eb' : '#374151');
  }

  onMount(() => {
    applyPreviewStyles();
  });

  function generateColorScale(baseColor: string): string[] {
    // Simple brightness adjustment for demo - in production use a proper color library
    return [
      adjustBrightness(baseColor, 0.9),
      adjustBrightness(baseColor, 0.8),
      adjustBrightness(baseColor, 0.6),
      adjustBrightness(baseColor, 0.4),
      adjustBrightness(baseColor, 0.2),
      baseColor,
      adjustBrightness(baseColor, -0.1),
      adjustBrightness(baseColor, -0.2),
      adjustBrightness(baseColor, -0.3),
      adjustBrightness(baseColor, -0.4),
    ];
  }

  function adjustBrightness(hex: string, factor: number): string {
    const rgb = hexToRgb(hex);
    if (!rgb) return hex;

    const adjust = (c: number) => {
      if (factor > 0) {
        return Math.round(c + (255 - c) * factor);
      } else {
        return Math.round(c * (1 + factor));
      }
    };

    return rgbToHex(adjust(rgb.r), adjust(rgb.g), adjust(rgb.b));
  }

  function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
      r: parseInt(result[1]!, 16),
      g: parseInt(result[2]!, 16),
      b: parseInt(result[3]!, 16),
    } : null;
  }

  function rgbToHex(r: number, g: number, b: number): string {
    return '#' + [r, g, b].map(x => {
      const hex = Math.max(0, Math.min(255, x)).toString(16);
      return hex.length === 1 ? '0' + hex : hex;
    }).join('');
  }
</script>

<div class={cn('theme-builder', className)}>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Controls Panel -->
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-semibold text-[var(--color-text-primary)]">
          Theme Builder
        </h2>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="px-3 py-1.5 text-sm rounded-lg border border-[var(--color-border-primary)] hover:bg-[var(--color-surface-secondary)]"
            on:click={handleReset}
          >
            Reset
          </button>
          <button
            type="button"
            class="px-3 py-1.5 text-sm rounded-lg bg-[var(--color-interactive-primary)] text-white hover:bg-[var(--color-interactive-primary-hover)]"
            on:click={handleSave}
          >
            Save Theme
          </button>
        </div>
      </div>

      <!-- Theme Name -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text-primary)] mb-2">
          Theme Name
        </label>
        <input
          type="text"
          bind:value={config.name}
          class="w-full px-3 py-2 rounded-lg border border-[var(--color-border-primary)] bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]"
          on:input={() => dispatch('change', { config })}
        />
      </div>

      <!-- Mode Toggle -->
      <div class="flex items-center justify-between">
        <span class="text-sm font-medium text-[var(--color-text-primary)]">Dark Mode</span>
        <Switch
          checked={config.mode === 'dark'}
          on:change={(e) => handleModeChange(e.detail.checked ? 'dark' : 'light')}
        />
      </div>

      <!-- Presets -->
      <div>
        <label class="block text-sm font-medium text-[var(--color-text-primary)] mb-3">
          Quick Presets
        </label>
        <div class="grid grid-cols-3 gap-2">
          {#each presets as preset}
            <button
              type="button"
              class="p-3 rounded-lg border border-[var(--color-border-primary)] hover:border-[var(--color-interactive-primary)] transition-colors"
              on:click={() => applyPreset(preset.id)}
            >
              <div class="flex gap-1 mb-2">
                <div
                  class="w-4 h-4 rounded-full"
                  style="background-color: {preset.colors.primary}"
                />
                <div
                  class="w-4 h-4 rounded-full"
                  style="background-color: {preset.colors.secondary}"
                />
              </div>
              <span class="text-xs text-[var(--color-text-secondary)]">{preset.name}</span>
            </button>
          {/each}
        </div>
      </div>

      <!-- Tabs -->
      <div class="border-b border-[var(--color-border-primary)]">
        <div class="flex gap-4">
          {#each ['colors', 'typography', 'spacing'] as tab}
            <button
              type="button"
              class={cn(
                'px-4 py-2 text-sm font-medium border-b-2 -mb-px capitalize',
                activeTab === tab
                  ? 'border-[var(--color-interactive-primary)] text-[var(--color-interactive-primary)]'
                  : 'border-transparent text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
              )}
              on:click={() => activeTab = tab as typeof activeTab}
            >
              {tab}
            </button>
          {/each}
        </div>
      </div>

      <!-- Tab Content -->
      {#if activeTab === 'colors'}
        <div class="space-y-4">
          {#each Object.entries(config.colors) as [key, value]}
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-[var(--color-text-primary)] capitalize">
                {key}
              </span>
              <div class="flex items-center gap-2">
                <ColorPicker
                  {value}
                  on:change={(e) => handleColorChange(key as keyof ThemeBuilderColors, e.detail.value)}
                />
                <input
                  type="text"
                  {value}
                  class="w-24 px-2 py-1 text-sm rounded border border-[var(--color-border-primary)] bg-[var(--color-surface-primary)] text-[var(--color-text-primary)] font-mono"
                  on:input={(e) => handleColorChange(key as keyof ThemeBuilderColors, (e.target as HTMLInputElement).value)}
                />
              </div>
            </div>

            <!-- Color scale preview -->
            <div class="flex gap-0.5">
              {#each generateColorScale(value) as shade, i}
                <div
                  class="h-4 flex-1 first:rounded-l last:rounded-r"
                  style="background-color: {shade}"
                  title="{String((i + 1) * 100)}"
                />
              {/each}
            </div>
          {/each}
        </div>
      {:else if activeTab === 'typography'}
        <div class="space-y-6">
          <div>
            <label class="block text-sm font-medium text-[var(--color-text-primary)] mb-2">
              Base Font Size: {config.fontSize}px
            </label>
            <Slider
              value={config.fontSize}
              min={12}
              max={20}
              step={1}
              on:input={(e) => handleFontSizeChange(e.detail.value)}
            />
          </div>
        </div>
      {:else if activeTab === 'spacing'}
        <div class="space-y-6">
          <div>
            <label class="block text-sm font-medium text-[var(--color-text-primary)] mb-2">
              Border Radius: {config.borderRadius}px
            </label>
            <Slider
              value={config.borderRadius}
              min={0}
              max={24}
              step={2}
              on:input={(e) => handleBorderRadiusChange(e.detail.value)}
            />
          </div>
        </div>
      {/if}
    </div>

    <!-- Preview Panel -->
    {#if showPreview}
      <div
        id="theme-builder-preview"
        class="p-6 rounded-xl border border-[var(--color-border-primary)]"
        style="
          background-color: var(--preview-bg);
          color: var(--preview-text);
          font-size: var(--preview-font-size);
        "
      >
        <h3 class="text-lg font-semibold mb-4">Preview</h3>

        <!-- Buttons -->
        <div class="space-y-4">
          <div>
            <p class="text-sm opacity-70 mb-2">Buttons</p>
            <div class="flex flex-wrap gap-2">
              <button
                class="px-4 py-2 text-white font-medium"
                style="background-color: var(--preview-primary); border-radius: var(--preview-radius);"
              >
                Primary
              </button>
              <button
                class="px-4 py-2 text-white font-medium"
                style="background-color: var(--preview-secondary); border-radius: var(--preview-radius);"
              >
                Secondary
              </button>
              <button
                class="px-4 py-2 border font-medium"
                style="border-color: var(--preview-border); border-radius: var(--preview-radius);"
              >
                Outline
              </button>
            </div>
          </div>

          <!-- Status badges -->
          <div>
            <p class="text-sm opacity-70 mb-2">Status</p>
            <div class="flex flex-wrap gap-2">
              <span
                class="px-2 py-1 text-xs text-white"
                style="background-color: var(--preview-success); border-radius: var(--preview-radius);"
              >
                Success
              </span>
              <span
                class="px-2 py-1 text-xs text-white"
                style="background-color: var(--preview-warning); border-radius: var(--preview-radius);"
              >
                Warning
              </span>
              <span
                class="px-2 py-1 text-xs text-white"
                style="background-color: var(--preview-error); border-radius: var(--preview-radius);"
              >
                Error
              </span>
              <span
                class="px-2 py-1 text-xs text-white"
                style="background-color: var(--preview-info); border-radius: var(--preview-radius);"
              >
                Info
              </span>
            </div>
          </div>

          <!-- Card -->
          <div>
            <p class="text-sm opacity-70 mb-2">Card</p>
            <div
              class="p-4 border"
              style="border-color: var(--preview-border); border-radius: var(--preview-radius);"
            >
              <h4 class="font-semibold mb-2">Sample Card</h4>
              <p class="text-sm opacity-70">
                This is a preview of how cards will look with your theme settings.
              </p>
              <button
                class="mt-3 px-3 py-1.5 text-sm text-white"
                style="background-color: var(--preview-primary); border-radius: var(--preview-radius);"
              >
                Action
              </button>
            </div>
          </div>

          <!-- Input -->
          <div>
            <p class="text-sm opacity-70 mb-2">Form Input</p>
            <input
              type="text"
              placeholder="Sample input field"
              class="w-full px-3 py-2 border"
              style="
                border-color: var(--preview-border);
                border-radius: var(--preview-radius);
                background: transparent;
              "
            />
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>
