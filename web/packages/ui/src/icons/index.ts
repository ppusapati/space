/**
 * @samavāya/ui/icons - SVG Icon Library
 *
 * A comprehensive collection of 400+ SVG icons that work offline.
 * All icons are bundled as path data for tree-shaking optimization.
 *
 * Usage:
 * ```svelte
 * <script>
 *   import { Icon } from '@samavāya/ui/icons';
 * </script>
 *
 * <Icon name="home" size="md" />
 * <Icon name="user" size={24} strokeWidth={1.5} />
 * <Icon name="settings" class="text-primary" label="Settings" />
 * ```
 *
 * Size presets: 'xs' (12px) | 'sm' (16px) | 'md' (20px) | 'lg' (24px) | 'xl' (32px) | '2xl' (40px)
 * Or pass a number for custom pixel size.
 */

// Main Icon component
export { default as Icon } from './Icon.svelte';

// Icon data and utilities
export { icons, getIconNames, hasIcon, type IconName } from './icons';
