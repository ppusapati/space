<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { Size } from '../types';

  // Props
  export let src: string = '';
  export let alt: string = '';
  export let name: string = '';
  export let size: Size = 'md';
  export let shape: 'circle' | 'square' = 'circle';
  export let status: 'online' | 'offline' | 'away' | 'busy' | 'none' = 'none';
  export let showBadge: boolean = false;
  export let badgeContent: string | number = '';

  let className: string = '';
  export { className as class };

  let imageError = false;

  const sizeConfig = {
    sm: { avatar: 'w-8 h-8 text-xs', status: 'w-2 h-2', badge: 'w-4 h-4 text-[8px]' },
    md: { avatar: 'w-10 h-10 text-sm', status: 'w-2.5 h-2.5', badge: 'w-5 h-5 text-[10px]' },
    lg: { avatar: 'w-12 h-12 text-base', status: 'w-3 h-3', badge: 'w-6 h-6 text-xs' },
  };

  const statusColors = {
    online: 'bg-[var(--color-success)]',
    offline: 'bg-[var(--color-neutral-400)]',
    away: 'bg-[var(--color-warning)]',
    busy: 'bg-[var(--color-error)]',
    none: '',
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
  $: initials = getInitials(name || alt);
  $: showImage = src && !imageError;

  function getInitials(text: string): string {
    if (!text) return '?';
    const words = text.trim().split(/\s+/);
    if (words.length === 1) {
      return words[0]!.charAt(0).toUpperCase();
    }
    return (words[0]!.charAt(0) + words[words.length - 1]!.charAt(0)).toUpperCase();
  }

  function getBackgroundColor(text: string): string {
    if (!text) return 'var(--color-neutral-400)';

    const colors = [
      'var(--color-interactive-primary)',
      'var(--color-success)',
      'var(--color-warning)',
      'var(--color-error)',
      'var(--color-info)',
    ];

    let hash = 0;
    for (let i = 0; i < text.length; i++) {
      hash = text.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length]!;
  }

  function handleImageError() {
    imageError = true;
  }
</script>

<div class={cn('relative inline-flex', className)}>
  <div
    class={cn(
      'flex items-center justify-center overflow-hidden',
      shape === 'circle' ? 'rounded-full' : 'rounded-lg',
      config.avatar
    )}
    style={!showImage ? `background-color: ${getBackgroundColor(name || alt)}` : ''}
  >
    {#if showImage}
      <img
        {src}
        alt={alt || name}
        class="w-full h-full object-cover"
        on:error={handleImageError}
      />
    {:else}
      <span class="font-medium text-white">
        {initials}
      </span>
    {/if}
  </div>

  <!-- Status Indicator -->
  {#if status !== 'none'}
    <span
      class={cn(
        'absolute bottom-0 right-0 rounded-full ring-2 ring-[var(--color-surface-primary)]',
        config.status,
        statusColors[status]
      )}
    />
  {/if}

  <!-- Badge -->
  {#if showBadge && badgeContent}
    <span
      class={cn(
        'absolute -top-1 -right-1 flex items-center justify-center rounded-full',
        'bg-[var(--color-error)] text-white font-medium',
        config.badge
      )}
    >
      {typeof badgeContent === 'number' && badgeContent > 99 ? '99+' : badgeContent}
    </span>
  {/if}
</div>
