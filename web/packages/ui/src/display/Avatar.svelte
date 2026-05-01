<script lang="ts">
  import type { AvatarProps } from './display.types';
  import {
    avatarClasses,
    avatarSizeClasses,
    avatarStatusClasses,
    avatarDynamicColors,
  } from './display.types';
  import { cn } from '../utils';

  type $$Props = AvatarProps;

  export let src: $$Props['src'] = undefined;
  export let alt: $$Props['alt'] = '';
  export let fallback: $$Props['fallback'] = undefined;
  export let size: $$Props['size'] = 'md';
  export let shape: $$Props['shape'] = 'circle';
  export let status: $$Props['status'] = undefined;
  export let ring: $$Props['ring'] = false;
  export let showBadge: $$Props['showBadge'] = false;
  export let badgeContent: $$Props['badgeContent'] = undefined;
  export let dynamicColor: $$Props['dynamicColor'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: sizeStyles = avatarSizeClasses[size || 'md'];
  $: shapeClass = avatarClasses[shape || 'circle'];
  $: statusClass = status ? avatarStatusClasses[status] : '';

  let imageError = false;

  function handleError() {
    imageError = true;
  }

  // Generate initials from fallback text
  $: initials = fallback
    ? fallback
        .split(' ')
        .map((word) => word[0])
        .filter(Boolean)
        .join('')
        .toUpperCase()
        .slice(0, 2)
    : '?';

  // Generate consistent color from fallback text
  function getDynamicColorClass(text: string | undefined): string {
    if (!text) return avatarDynamicColors[0] ?? '';
    let hash = 0;
    for (let i = 0; i < text.length; i++) {
      hash = text.charCodeAt(i) + ((hash << 5) - hash);
    }
    return avatarDynamicColors[Math.abs(hash) % avatarDynamicColors.length] ?? '';
  }

  $: dynamicColorClass = getDynamicColorClass(fallback);
  $: showImage = src && !imageError;
  $: useDynamicColor = dynamicColor && !showImage && fallback;

  // Format badge content
  $: formattedBadge =
    typeof badgeContent === 'number' && badgeContent > 99
      ? '99+'
      : badgeContent;
</script>

<div
  class={cn(
    avatarClasses.container,
    sizeStyles.container,
    shapeClass,
    ring && avatarClasses.ring,
    className
  )}
  role="img"
  aria-label={alt || fallback || 'Avatar'}
>
  {#if showImage}
    <img
      {src}
      {alt}
      class={cn(avatarClasses.image, shapeClass)}
      on:error={handleError}
    />
  {:else}
    <div
      class={cn(
        useDynamicColor ? avatarClasses.fallbackDynamic : avatarClasses.fallback,
        useDynamicColor && dynamicColorClass,
        sizeStyles.text,
        shapeClass,
        'w-full h-full'
      )}
    >
      {initials}
    </div>
  {/if}

  {#if status}
    <span
      class={cn(avatarClasses.status, sizeStyles.status, statusClass)}
      aria-label={`Status: ${status}`}
    />
  {/if}

  {#if showBadge && badgeContent !== undefined}
    <span
      class={cn(avatarClasses.badge, sizeStyles.badge)}
      aria-label={`Badge: ${badgeContent}`}
    >
      {formattedBadge}
    </span>
  {/if}
</div>
