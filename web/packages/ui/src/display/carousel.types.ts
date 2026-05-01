/**
 * Carousel and Gallery component types
 */

import type { Size, BaseProps } from '../types';

/** Carousel image/slide item */
export interface CarouselItem {
  /** Unique identifier */
  id: string | number;
  /** Image source URL */
  src: string;
  /** Alt text for accessibility */
  alt: string;
  /** Optional caption text */
  caption?: string;
  /** Optional thumbnail URL (defaults to src if not provided) */
  thumbnail?: string;
  /** Optional link URL */
  href?: string;
  /** Additional custom data */
  data?: Record<string, unknown>;
}

/** Carousel transition type */
export type CarouselTransition = 'slide' | 'fade' | 'none';

/** Carousel props */
export interface CarouselProps extends BaseProps {
  /** Array of carousel items */
  items: CarouselItem[];
  /** Current active index */
  currentIndex?: number;
  /** Enable auto-play */
  autoPlay?: boolean;
  /** Auto-play interval in milliseconds */
  interval?: number;
  /** Pause auto-play on hover */
  pauseOnHover?: boolean;
  /** Show navigation controls (prev/next arrows) */
  showControls?: boolean;
  /** Show indicator dots */
  showIndicators?: boolean;
  /** Indicator position */
  indicatorPosition?: 'bottom' | 'top';
  /** Loop back to start/end */
  loop?: boolean;
  /** Transition type */
  transition?: CarouselTransition;
  /** Transition duration in ms */
  transitionDuration?: number;
  /** Show captions */
  showCaptions?: boolean;
  /** Caption position */
  captionPosition?: 'bottom' | 'overlay';
  /** Aspect ratio (e.g., '16/9', '4/3', '1/1') */
  aspectRatio?: string;
  /** Object fit for images */
  fit?: 'contain' | 'cover' | 'fill';
  /** Enable keyboard navigation */
  keyboard?: boolean;
  /** Enable touch/swipe navigation */
  touch?: boolean;
}

/** Gallery layout type */
export type GalleryLayout = 'grid' | 'masonry' | 'justified';

/** Gallery props */
export interface GalleryProps extends BaseProps {
  /** Array of gallery items */
  items: CarouselItem[];
  /** Number of columns */
  columns?: number | { xs?: number; sm?: number; md?: number; lg?: number; xl?: number };
  /** Gap between items */
  gap?: Size;
  /** Layout type */
  layout?: GalleryLayout;
  /** Aspect ratio for grid items (e.g., '1/1', '4/3') */
  aspectRatio?: string;
  /** Enable lightbox on click */
  lightbox?: boolean;
  /** Object fit for images */
  fit?: 'contain' | 'cover' | 'fill';
  /** Border radius */
  rounded?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
  /** Hoverable effect */
  hoverable?: boolean;
}

/** Lightbox props */
export interface LightboxProps extends BaseProps {
  /** Whether lightbox is open */
  open: boolean;
  /** Array of items */
  items: CarouselItem[];
  /** Starting index */
  startIndex?: number;
  /** Show captions */
  showCaptions?: boolean;
  /** Show thumbnails strip */
  showThumbnails?: boolean;
  /** Show counter (1/10) */
  showCounter?: boolean;
  /** Allow close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Allow close on escape key */
  closeOnEscape?: boolean;
  /** Enable zoom */
  zoomable?: boolean;
  /** Enable download button */
  downloadable?: boolean;
}

/** Carousel classes */
export const carouselClasses = {
  container: 'relative overflow-hidden',
  viewport: 'relative w-full',
  track: 'flex transition-transform',
  slide: 'flex-shrink-0 w-full',
  image: 'w-full h-full',
  controls: {
    base: 'absolute top-1/2 -translate-y-1/2 z-10',
    button: 'flex items-center justify-center w-10 h-10 rounded-full bg-neutral-white/80 text-neutral-700 hover:bg-neutral-white hover:text-neutral-900 shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed',
    prev: 'left-3',
    next: 'right-3',
  },
  indicators: {
    container: 'absolute left-1/2 -translate-x-1/2 z-10 flex items-center gap-2',
    bottom: 'bottom-4',
    top: 'top-4',
    dot: 'w-2 h-2 rounded-full bg-neutral-white/60 hover:bg-neutral-white/80 transition-colors cursor-pointer',
    dotActive: 'bg-neutral-white w-3',
  },
  caption: {
    bottom: 'absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-neutral-black/60 to-transparent text-neutral-white',
    overlay: 'absolute bottom-0 left-0 right-0 p-4 bg-neutral-black/50 text-neutral-white',
    text: 'text-sm md:text-base',
  },
};

/** Gallery classes */
export const galleryClasses = {
  container: 'w-full',
  grid: 'grid',
  item: 'overflow-hidden cursor-pointer',
  itemHoverable: 'transition-transform hover:scale-[1.02] hover:shadow-lg',
  image: 'w-full h-full transition-transform',
  imageHover: 'group-hover:scale-110',
};

export const galleryGapClasses: Record<Size, string> = {
  xs: 'gap-1',
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
  xl: 'gap-8',
};

export const galleryRoundedClasses = {
  none: '',
  sm: 'rounded-sm',
  md: 'rounded-md',
  lg: 'rounded-lg',
  xl: 'rounded-xl',
};

/** Lightbox classes */
export const lightboxClasses = {
  overlay: 'fixed inset-0 z-modal bg-neutral-black/95',
  container: 'fixed inset-0 z-modal flex flex-col',
  header: 'flex items-center justify-between px-4 py-3 text-neutral-white',
  counter: 'text-sm font-medium',
  actions: 'flex items-center gap-2',
  actionButton: 'p-2 rounded-lg text-neutral-white/80 hover:text-neutral-white hover:bg-neutral-white/10 transition-colors',
  main: 'flex-1 flex items-center justify-center relative overflow-hidden px-14',
  image: 'max-w-full max-h-full object-contain',
  nav: {
    button: 'absolute top-1/2 -translate-y-1/2 z-10 flex items-center justify-center w-12 h-12 rounded-full bg-neutral-white/10 text-neutral-white hover:bg-neutral-white/20 transition-colors',
    prev: 'left-2',
    next: 'right-2',
  },
  caption: 'text-center text-neutral-white py-3 px-4',
  thumbnails: {
    container: 'flex items-center justify-center gap-2 py-3 px-4 overflow-x-auto',
    item: 'w-16 h-12 rounded overflow-hidden cursor-pointer opacity-60 hover:opacity-80 transition-opacity flex-shrink-0',
    itemActive: 'opacity-100 ring-2 ring-neutral-white',
    image: 'w-full h-full object-cover',
  },
  closeButton: 'p-2 rounded-lg text-neutral-white/80 hover:text-neutral-white hover:bg-neutral-white/10 transition-colors',
};
