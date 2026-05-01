<script lang="ts">
  import type { AvatarGroupProps, AvatarData } from './display.types';
  import { avatarGroupClasses, avatarSizeClasses, avatarClasses } from './display.types';
  import { cn } from '../utils';
  import Avatar from './Avatar.svelte';

  type $$Props = AvatarGroupProps;

  export let max: $$Props['max'] = 5;
  export let size: $$Props['size'] = 'md';
  export let direction: $$Props['direction'] = 'left';
  export let avatars: AvatarData[] | undefined = undefined;
  export let shape: $$Props['shape'] = 'circle';
  let className: $$Props['class'] = undefined;
  export { className as class };

  // For slot-based usage, pass total count
  export let total: number = 0;

  $: sizeStyles = avatarSizeClasses[size || 'md'];
  $: directionClass = avatarGroupClasses[direction || 'left'];
  $: shapeClass = avatarClasses[shape || 'circle'];

  // Calculate visible avatars and overflow
  $: effectiveMax = max || 5;
  $: visibleAvatars = avatars ? avatars.slice(0, effectiveMax) : [];
  $: avatarOverflow = avatars ? Math.max(0, avatars.length - effectiveMax) : 0;
  $: slotOverflow = total > effectiveMax ? total - effectiveMax : 0;
  $: overflow = avatars ? avatarOverflow : slotOverflow;
</script>

<div
  class={cn(avatarGroupClasses.container, directionClass, className)}
  role="group"
  aria-label="Avatar group"
>
  {#if avatars}
    {#each visibleAvatars as avatar, i (i)}
      <div
        class={cn(
          'ring-2 ring-neutral-white',
          shapeClass
        )}
      >
        <Avatar
          src={avatar.src}
          fallback={avatar.name}
          alt={avatar.alt || avatar.name}
          {size}
          {shape}
        />
      </div>
    {/each}
  {:else}
    <slot />
  {/if}

  {#if overflow > 0}
    <div
      class={cn(
        avatarGroupClasses.overflow,
        sizeStyles.container,
        sizeStyles.text,
        shapeClass,
        'ring-2 ring-neutral-white'
      )}
      aria-label={`+${overflow} more`}
    >
      +{overflow}
    </div>
  {/if}
</div>
