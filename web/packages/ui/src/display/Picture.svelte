<script lang="ts">
  import type { PictureProps } from './display.types';
  import { pictureClasses } from './display.types';

  const {
    src,
    alt,
    width,
    height,
    loading = 'lazy',
    decoding = 'async',
    class: className = '',
    fallback = 'jpg',
    sizes,
    widths,
    fit,
    rounded,
  }: PictureProps = $props();

  // Extract base path and extension
  function getBasePath(imageSrc: string): { base: string; ext: string } {
    const lastDot = imageSrc.lastIndexOf('.');
    if (lastDot === -1) {
      return { base: imageSrc, ext: '' };
    }
    return {
      base: imageSrc.substring(0, lastDot),
      ext: imageSrc.substring(lastDot + 1),
    };
  }

  const { base, ext } = getBasePath(src);
  const hasExtension = ext.length > 0;
  const basePath = hasExtension ? base : src;
  const fallbackExt = hasExtension ? ext : fallback;

  // Generate srcset for responsive images
  function generateSrcset(format: string): string {
    if (!widths || widths.length === 0) {
      return `${basePath}.${format}`;
    }
    return widths
      .map((w) => `${basePath}-${w}w.${format} ${w}w`)
      .join(', ');
  }

  const avifSrcset = generateSrcset('avif');
  const webpSrcset = generateSrcset('webp');
  const fallbackSrcset = generateSrcset(fallbackExt);

  // Build style classes using pictureClasses
  const fitClass = fit ? pictureClasses.fit[fit] : '';
  const roundedClass = rounded ? pictureClasses.rounded[rounded] : '';
  const imgClass = [fitClass, roundedClass, className].filter(Boolean).join(' ');
</script>

<picture>
  {#if widths && widths.length > 0}
    <source type="image/avif" srcset={avifSrcset} {sizes} />
    <source type="image/webp" srcset={webpSrcset} {sizes} />
    <img
      src="{basePath}.{fallbackExt}"
      srcset={fallbackSrcset}
      {sizes}
      {alt}
      {width}
      {height}
      {loading}
      {decoding}
      class={imgClass}
    />
  {:else}
    <source type="image/avif" srcset="{basePath}.avif" />
    <source type="image/webp" srcset="{basePath}.webp" />
    <img
      src="{basePath}.{fallbackExt}"
      {alt}
      {width}
      {height}
      {loading}
      {decoding}
      class={imgClass}
    />
  {/if}
</picture>
