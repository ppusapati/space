/**
 * Custom transition utilities and presets
 * These work alongside Svelte's built-in transitions
 *
 * Usage:
 * import { fadeScale, slideUp } from '@samavāya/ui';
 * <div transition:fadeScale>Content</div>
 */

import { cubicOut, cubicIn, cubicInOut } from 'svelte/easing';
import type { TransitionConfig } from 'svelte/transition';

export interface TransitionOptions {
  delay?: number;
  duration?: number;
  easing?: (t: number) => number;
}

/**
 * Fade and scale transition
 */
export function fadeScale(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;
  const transform = style.transform === 'none' ? '' : style.transform;

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      transform: ${transform} scale(${0.95 + 0.05 * t});
    `,
  };
}

/**
 * Slide up transition
 */
export function slideUp(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;
  const height = parseFloat(style.height);

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      height: ${t * height}px;
      transform: translateY(${(1 - t) * 10}px);
      overflow: hidden;
    `,
  };
}

/**
 * Slide down transition
 */
export function slideDown(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;
  const height = parseFloat(style.height);

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      height: ${t * height}px;
      transform: translateY(${(1 - t) * -10}px);
      overflow: hidden;
    `,
  };
}

/**
 * Slide from left transition
 */
export function slideLeft(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      transform: translateX(${(1 - t) * -20}px);
    `,
  };
}

/**
 * Slide from right transition
 */
export function slideRight(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      transform: translateX(${(1 - t) * 20}px);
    `,
  };
}

/**
 * Pop/bounce transition
 */
export function pop(
  node: HTMLElement,
  { delay = 0, duration = 300, easing = cubicOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;

  return {
    delay,
    duration,
    easing,
    css: (t) => {
      const scale = t < 0.5 ? t * 2.2 : 1 + (1 - t) * 0.2;
      return `
        opacity: ${Math.min(t * 2, 1) * opacity};
        transform: scale(${scale});
      `;
    },
  };
}

/**
 * Collapse transition (for accordions, etc.)
 */
export function collapse(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicInOut }: TransitionOptions = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const height = parseFloat(style.height);
  const paddingTop = parseFloat(style.paddingTop);
  const paddingBottom = parseFloat(style.paddingBottom);
  const marginTop = parseFloat(style.marginTop);
  const marginBottom = parseFloat(style.marginBottom);

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      height: ${t * height}px;
      padding-top: ${t * paddingTop}px;
      padding-bottom: ${t * paddingBottom}px;
      margin-top: ${t * marginTop}px;
      margin-bottom: ${t * marginBottom}px;
      overflow: hidden;
    `,
  };
}

/**
 * Blur transition
 */
export function blur(
  node: HTMLElement,
  { delay = 0, duration = 200, easing = cubicOut }: TransitionOptions & { amount?: number } = {}
): TransitionConfig {
  const style = getComputedStyle(node);
  const opacity = +style.opacity;
  const amount = 5;

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      opacity: ${t * opacity};
      filter: blur(${(1 - t) * amount}px);
    `,
  };
}

/**
 * Typewriter transition for text
 */
export function typewriter(
  node: HTMLElement,
  { delay = 0, duration = 50 }: { delay?: number; duration?: number } = {}
): TransitionConfig {
  const text = node.textContent || '';
  const totalDuration = duration * text.length;

  return {
    delay,
    duration: totalDuration,
    tick: (t) => {
      const i = Math.trunc(text.length * t);
      node.textContent = text.slice(0, i);
    },
  };
}

export default {
  fadeScale,
  slideUp,
  slideDown,
  slideLeft,
  slideRight,
  pop,
  collapse,
  blur,
  typewriter,
};
