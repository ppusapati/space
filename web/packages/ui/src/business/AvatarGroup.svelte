<script context="module" lang="ts">
  export interface AvatarData {
    src?: string;
    name: string;
    alt?: string;
  }
</script>

<script lang="ts">
  import { cn } from '../utils/classnames';
  import Avatar from './Avatar.svelte';
  import type { Size } from '../types';

  // Props
  export let avatars: AvatarData[] = [];
  export let max: number = 4;
  export let size: Size = 'md';
  export let shape: 'circle' | 'square' = 'circle';

  let className: string = '';
  export { className as class };

  $: visibleAvatars = avatars.slice(0, max);
  $: remainingCount = avatars.length - max;
  $: showOverflow = remainingCount > 0;

  const sizeConfig = {
    sm: { overlap: '-ml-2', avatar: 'w-8 h-8 text-xs' },
    md: { overlap: '-ml-3', avatar: 'w-10 h-10 text-sm' },
    lg: { overlap: '-ml-4', avatar: 'w-12 h-12 text-base' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
</script>

<div class={cn('flex items-center', className)}>
  {#each visibleAvatars as avatar, i (i)}
    <div
      class={cn(
        'ring-2 ring-[var(--color-surface-primary)]',
        shape === 'circle' ? 'rounded-full' : 'rounded-lg',
        i > 0 && config.overlap
      )}
    >
      <Avatar
        src={avatar.src}
        name={avatar.name}
        alt={avatar.alt || avatar.name}
        {size}
        {shape}
      />
    </div>
  {/each}

  {#if showOverflow}
    <div
      class={cn(
        'flex items-center justify-center',
        'ring-2 ring-[var(--color-surface-primary)]',
        'bg-[var(--color-surface-tertiary)] text-[var(--color-text-secondary)]',
        'font-medium',
        shape === 'circle' ? 'rounded-full' : 'rounded-lg',
        config.avatar,
        config.overlap
      )}
    >
      +{remainingCount}
    </div>
  {/if}
</div>
