<script lang="ts">
  import { icons, type IconName } from './icons';

  interface Props {
    /** Icon name from the icons collection */
    name: IconName;
    /** Size preset or custom pixel size */
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | number;
    /** Stroke width (default: 2) */
    strokeWidth?: number;
    /** Additional CSS classes */
    class?: string;
    /** Accessible label for screen readers */
    label?: string;
  }

  const {
    name,
    size = 'md',
    strokeWidth = 2,
    class: className = '',
    label,
  }: Props = $props();

  const sizeMap: Record<string, string> = {
    xs: '12',
    sm: '16',
    md: '20',
    lg: '24',
    xl: '32',
    '2xl': '40',
  };

  const computedSize = $derived(
    typeof size === 'number' ? String(size) : sizeMap[size] || '20'
  );

  const iconPath = $derived(icons[name] || '');
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={computedSize}
  height={computedSize}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={strokeWidth}
  stroke-linecap="round"
  stroke-linejoin="round"
  class="inline-block flex-shrink-0 {className}"
  aria-hidden={!label}
  aria-label={label}
  role={label ? 'img' : undefined}
>
  {@html iconPath}
</svg>
