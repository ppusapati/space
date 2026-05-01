# Design Tokens Cheatsheet

> Quick reference for CSS custom properties in `@p9e/design-tokens`

## How to Use

```css
/* In CSS */
.element {
  color: var(--color-brand-primary-500);
  padding: var(--spacing-4);
  font-size: var(--typography-fontSize-base);
}
```

```js
// In JavaScript
import { tokens } from '@p9e/design-tokens/js';
const primaryColor = tokens.color.brand.primary['500'];
```

---

## Colors

### Brand Colors

| Token | CSS Variable | Value | Preview |
|-------|--------------|-------|---------|
| Primary 50 | `--color-brand-primary-50` | `#f0f9ff` | ![](https://via.placeholder.com/20/f0f9ff/f0f9ff) |
| Primary 100 | `--color-brand-primary-100` | `#e0f2fe` | ![](https://via.placeholder.com/20/e0f2fe/e0f2fe) |
| Primary 200 | `--color-brand-primary-200` | `#bae6fd` | ![](https://via.placeholder.com/20/bae6fd/bae6fd) |
| Primary 300 | `--color-brand-primary-300` | `#7dd3fc` | ![](https://via.placeholder.com/20/7dd3fc/7dd3fc) |
| Primary 400 | `--color-brand-primary-400` | `#38bdf8` | ![](https://via.placeholder.com/20/38bdf8/38bdf8) |
| **Primary 500** | `--color-brand-primary-500` | `#0ea5e9` | ![](https://via.placeholder.com/20/0ea5e9/0ea5e9) |
| Primary 600 | `--color-brand-primary-600` | `#0284c7` | ![](https://via.placeholder.com/20/0284c7/0284c7) |
| Primary 700 | `--color-brand-primary-700` | `#0369a1` | ![](https://via.placeholder.com/20/0369a1/0369a1) |
| Primary 800 | `--color-brand-primary-800` | `#075985` | ![](https://via.placeholder.com/20/075985/075985) |
| Primary 900 | `--color-brand-primary-900` | `#0c4a6e` | ![](https://via.placeholder.com/20/0c4a6e/0c4a6e) |

| Token | CSS Variable | Value | Preview |
|-------|--------------|-------|---------|
| Secondary 50 | `--color-brand-secondary-50` | `#fefce8` | ![](https://via.placeholder.com/20/fefce8/fefce8) |
| Secondary 100 | `--color-brand-secondary-100` | `#fef9c3` | ![](https://via.placeholder.com/20/fef9c3/fef9c3) |
| Secondary 200 | `--color-brand-secondary-200` | `#fef08a` | ![](https://via.placeholder.com/20/fef08a/fef08a) |
| Secondary 300 | `--color-brand-secondary-300` | `#fde047` | ![](https://via.placeholder.com/20/fde047/fde047) |
| Secondary 400 | `--color-brand-secondary-400` | `#facc15` | ![](https://via.placeholder.com/20/facc15/facc15) |
| **Secondary 500** | `--color-brand-secondary-500` | `#eab308` | ![](https://via.placeholder.com/20/eab308/eab308) |
| Secondary 600 | `--color-brand-secondary-600` | `#ca8a04` | ![](https://via.placeholder.com/20/ca8a04/ca8a04) |
| Secondary 700 | `--color-brand-secondary-700` | `#a16207` | ![](https://via.placeholder.com/20/a16207/a16207) |
| Secondary 800 | `--color-brand-secondary-800` | `#854d0e` | ![](https://via.placeholder.com/20/854d0e/854d0e) |
| Secondary 900 | `--color-brand-secondary-900` | `#713f12` | ![](https://via.placeholder.com/20/713f12/713f12) |

### Neutral Colors

| Token | CSS Variable | Value |
|-------|--------------|-------|
| White | `--color-neutral-white` | `#ffffff` |
| 25 | `--color-neutral-25` | `#fcfcfc` |
| 50 | `--color-neutral-50` | `#fafafa` |
| 100 | `--color-neutral-100` | `#f5f5f5` |
| 200 | `--color-neutral-200` | `#e5e5e5` |
| 300 | `--color-neutral-300` | `#d4d4d4` |
| 400 | `--color-neutral-400` | `#a3a3a3` |
| 500 | `--color-neutral-500` | `#737373` |
| 600 | `--color-neutral-600` | `#525252` |
| 700 | `--color-neutral-700` | `#404040` |
| 800 | `--color-neutral-800` | `#262626` |
| 850 | `--color-neutral-850` | `#1c1c1c` |
| 900 | `--color-neutral-900` | `#171717` |
| Black | `--color-neutral-black` | `#000000` |

### Semantic Colors

#### Success (Green)
| Token | CSS Variable | Value |
|-------|--------------|-------|
| 50-100 | `--color-semantic-success-{50,100}` | Light backgrounds |
| 200-400 | `--color-semantic-success-{200,300,400}` | Lighter states |
| **500** | `--color-semantic-success-500` | `#22c55e` (Main) |
| 600-900 | `--color-semantic-success-{600,700,800,900}` | Darker states |

#### Warning (Amber)
| Token | CSS Variable | Value |
|-------|--------------|-------|
| **500** | `--color-semantic-warning-500` | `#f59e0b` (Main) |

#### Error (Red)
| Token | CSS Variable | Value |
|-------|--------------|-------|
| **500** | `--color-semantic-error-500` | `#ef4444` (Main) |

#### Info (Blue)
| Token | CSS Variable | Value |
|-------|--------------|-------|
| **500** | `--color-semantic-info-500` | `#3b82f6` (Main) |

---

## Theme-Aware Colors (Recommended)

> Use these for automatic light/dark theme support

### Surface Colors
```css
--color-surface-primary     /* Main background (white/dark) */
--color-surface-secondary   /* Subtle background */
--color-surface-tertiary    /* More contrast */
--color-surface-inverse     /* Inverted surface */
--color-surface-overlay     /* Modal/drawer overlays */
```

### Text Colors
```css
--color-text-primary        /* Main text */
--color-text-secondary      /* Less emphasis */
--color-text-tertiary       /* Subtle text */
--color-text-inverse        /* On dark/light backgrounds */
--color-text-placeholder    /* Input placeholders */
--color-text-disabled       /* Disabled state */
```

### Border Colors
```css
--color-border-primary      /* Default borders */
--color-border-secondary    /* Subtle borders */
--color-border-focus        /* Focus ring (brand color) */
--color-border-inverse      /* On inverse backgrounds */
```

### Interactive Colors
```css
--color-interactive-primary         /* Button/link default */
--color-interactive-primaryHover    /* Hover state */
--color-interactive-primaryActive   /* Active/pressed */
--color-interactive-secondary       /* Secondary button */
--color-interactive-secondaryHover  /* Secondary hover */
--color-interactive-secondaryActive /* Secondary active */
```

---

## Spacing

| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| 0 | `--spacing-0` | 0px | None |
| px | `--spacing-px` | 1px | Hairline |
| 0.5 | `--spacing-0-5` | 2px | Micro |
| 1 | `--spacing-1` | 4px | Tiny |
| 1.5 | `--spacing-1-5` | 6px | Small gap |
| 2 | `--spacing-2` | 8px | Small |
| 3 | `--spacing-3` | 12px | Compact |
| **4** | `--spacing-4` | **16px** | **Base unit** |
| 5 | `--spacing-5` | 20px | Medium-small |
| 6 | `--spacing-6` | 24px | Medium |
| 8 | `--spacing-8` | 32px | Large |
| 10 | `--spacing-10` | 40px | Section gap |
| 12 | `--spacing-12` | 48px | Large section |
| 16 | `--spacing-16` | 64px | XL section |
| 20 | `--spacing-20` | 80px | Page section |
| 24 | `--spacing-24` | 96px | Large page |

**Common Pattern:**
```css
.card {
  padding: var(--spacing-4);       /* 16px */
  gap: var(--spacing-2);           /* 8px */
  margin-bottom: var(--spacing-6); /* 24px */
}
```

---

## Typography

### Font Families
```css
--typography-fontFamily-primary    /* Inter - UI text */
--typography-fontFamily-secondary  /* Georgia - Editorial */
--typography-fontFamily-mono       /* Fira Code - Code */
--typography-fontFamily-display    /* Playfair Display - Headings */
```

### Font Sizes
| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| xs | `--typography-fontSize-xs` | 12px | Captions, labels |
| sm | `--typography-fontSize-sm` | 14px | Small text |
| **base** | `--typography-fontSize-base` | **16px** | **Body text** |
| lg | `--typography-fontSize-lg` | 18px | Large body |
| xl | `--typography-fontSize-xl` | 20px | Small heading |
| 2xl | `--typography-fontSize-2xl` | 24px | H4 |
| 3xl | `--typography-fontSize-3xl` | 30px | H3 |
| 4xl | `--typography-fontSize-4xl` | 36px | H2 |
| 5xl | `--typography-fontSize-5xl` | 48px | H1 |
| 6xl | `--typography-fontSize-6xl` | 64px | Hero |
| 7xl-9xl | `--typography-fontSize-{7,8,9}xl` | 72-128px | Display |

### Font Weights
```css
--typography-fontWeight-thin       /* 100 */
--typography-fontWeight-light      /* 300 */
--typography-fontWeight-normal     /* 400 - Body text */
--typography-fontWeight-medium     /* 500 - Emphasis */
--typography-fontWeight-semibold   /* 600 - Subheadings */
--typography-fontWeight-bold       /* 700 - Headings */
--typography-fontWeight-black      /* 900 - Display */
```

### Line Heights
```css
--typography-lineHeight-none       /* 1 - Single line */
--typography-lineHeight-tight      /* 1.25 - Headings */
--typography-lineHeight-snug       /* 1.375 - Compact text */
--typography-lineHeight-normal     /* 1.5 - Body text */
--typography-lineHeight-relaxed    /* 1.625 - Readable */
--typography-lineHeight-loose      /* 2 - Very open */
```

### Letter Spacing
```css
--typography-letterSpacing-tighter /* -0.05em - Display text */
--typography-letterSpacing-tight   /* -0.025em - Headings */
--typography-letterSpacing-normal  /* 0 - Body */
--typography-letterSpacing-wide    /* 0.025em - Buttons */
--typography-letterSpacing-wider   /* 0.05em - Labels */
--typography-letterSpacing-widest  /* 0.1em - Uppercase */
```

---

## Border Radius

| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| none | `--borderRadius-none` | 0px | Sharp corners |
| sm | `--borderRadius-sm` | 2px | Subtle rounding |
| **base** | `--borderRadius-base` | **4px** | **Default** |
| md | `--borderRadius-md` | 6px | Cards |
| lg | `--borderRadius-lg` | 8px | Modals |
| xl | `--borderRadius-xl` | 12px | Large cards |
| 2xl | `--borderRadius-2xl` | 16px | Panels |
| 3xl | `--borderRadius-3xl` | 24px | Large panels |
| full | `--borderRadius-full` | 9999px | Pills/circles |

---

## Shadows

| Token | CSS Variable | Use Case |
|-------|--------------|----------|
| none | `--shadow-none` | Flat |
| xs | `--shadow-xs` | Subtle lift |
| sm | `--shadow-sm` | Cards (default) |
| base | `--shadow-base` | Elevated cards |
| md | `--shadow-md` | Dropdowns |
| lg | `--shadow-lg` | Popovers |
| xl | `--shadow-xl` | Modals |
| 2xl | `--shadow-2xl` | Large modals |
| inner | `--shadow-inner` | Inset shadow |

**Theme-aware shadows:**
```css
--shadow-card     /* Card shadow (theme-aware) */
--shadow-modal    /* Modal shadow (theme-aware) */
--shadow-dropdown /* Dropdown shadow (theme-aware) */
```

---

## Z-Index

| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| 0-50 | `--zIndex-{0,10,20,30,40,50}` | 0-50 | Local stacking |
| dropdown | `--zIndex-dropdown` | 1000 | Dropdowns |
| header | `--zIndex-header` | 1010 | App header |
| sticky | `--zIndex-sticky` | 1020 | Sticky elements |
| fixed | `--zIndex-fixed` | 1030 | Fixed elements |
| overlay | `--zIndex-overlay` | 1035 | Backdrops |
| drawer | `--zIndex-drawer` | 1038 | Side drawers |
| modal | `--zIndex-modal` | 1040 | Modals |
| popover | `--zIndex-popover` | 1050 | Popovers |
| tooltip | `--zIndex-tooltip` | 1060 | Tooltips |
| toast | `--zIndex-toast` | 1070 | Notifications |

---

## Layout

### Breakpoints

| Token | CSS Variable | Value | Description |
|-------|--------------|-------|-------------|
| xs | `--layout-breakpoint-xs` | 320px | Mobile portrait |
| sm | `--layout-breakpoint-sm` | 640px | Large mobile |
| md | `--layout-breakpoint-md` | 768px | Tablet |
| lg | `--layout-breakpoint-lg` | 1024px | Small desktop |
| xl | `--layout-breakpoint-xl` | 1280px | Desktop |
| 2xl | `--layout-breakpoint-2xl` | 1536px | Large desktop |
| 3xl | `--layout-breakpoint-3xl` | 1920px | Wide desktop |

### Container Sizes

| Token | CSS Variable | Value |
|-------|--------------|-------|
| xs | `--layout-container-xs` | 100% |
| sm | `--layout-container-sm` | 640px |
| md | `--layout-container-md` | 768px |
| lg | `--layout-container-lg` | 1024px |
| xl | `--layout-container-xl` | 1280px |
| 2xl | `--layout-container-2xl` | 1536px |

### Aspect Ratios

| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| square | `--layout-aspect-square` | 1 / 1 | Profile images |
| video | `--layout-aspect-video` | 16 / 9 | Videos |
| cinema | `--layout-aspect-cinema` | 21 / 9 | Cinematic |
| photo | `--layout-aspect-photo` | 4 / 3 | Photos |
| portrait | `--layout-aspect-portrait` | 3 / 4 | Vertical |
| landscape | `--layout-aspect-landscape` | 3 / 2 | Horizontal |
| golden | `--layout-aspect-golden` | 1.618 / 1 | Golden ratio |
| wide | `--layout-aspect-wide` | 2 / 1 | Wide banners |
| ultrawide | `--layout-aspect-ultrawide` | 32 / 9 | Ultra-wide |

### Grid System

**Columns:**
```css
--layout-grid-columns-1   /* repeat(1, minmax(0, 1fr)) */
--layout-grid-columns-2   /* repeat(2, minmax(0, 1fr)) */
--layout-grid-columns-3   /* repeat(3, minmax(0, 1fr)) */
--layout-grid-columns-4   /* repeat(4, minmax(0, 1fr)) */
--layout-grid-columns-5   /* repeat(5, minmax(0, 1fr)) */
--layout-grid-columns-6   /* repeat(6, minmax(0, 1fr)) */
--layout-grid-columns-12  /* repeat(12, minmax(0, 1fr)) */
```

**Gap:**
| Token | CSS Variable | Value |
|-------|--------------|-------|
| xs | `--layout-grid-gap-xs` | 4px |
| sm | `--layout-grid-gap-sm` | 8px |
| md | `--layout-grid-gap-md` | 16px |
| lg | `--layout-grid-gap-lg` | 24px |
| xl | `--layout-grid-gap-xl` | 32px |
| 2xl | `--layout-grid-gap-2xl` | 48px |

### Viewport Ranges

| Device | Min | Max |
|--------|-----|-----|
| Mobile | `--layout-viewport-mobile-min` (320px) | `--layout-viewport-mobile-max` (767px) |
| Tablet | `--layout-viewport-tablet-min` (768px) | `--layout-viewport-tablet-max` (1023px) |
| Desktop | `--layout-viewport-desktop-min` (1024px) | `--layout-viewport-desktop-max` (1919px) |
| Wide | `--layout-viewport-wide-min` (1920px) | - |

### Layout Z-Index (Alternative Scale)

| Token | CSS Variable | Value |
|-------|--------------|-------|
| hide | `--layout-z-index-hide` | -1 |
| base | `--layout-z-index-base` | 0 |
| docked | `--layout-z-index-docked` | 10 |
| dropdown | `--layout-z-index-dropdown` | 1000 |
| sticky | `--layout-z-index-sticky` | 1100 |
| banner | `--layout-z-index-banner` | 1200 |
| overlay | `--layout-z-index-overlay` | 1300 |
| modal | `--layout-z-index-modal` | 1400 |
| popover | `--layout-z-index-popover` | 1500 |
| skipLink | `--layout-z-index-skipLink` | 1600 |
| toast | `--layout-z-index-toast` | 1700 |
| tooltip | `--layout-z-index-tooltip` | 1800 |

**Grid Layout Example:**
```css
.grid-container {
  display: grid;
  grid-template-columns: var(--layout-grid-columns-3);
  gap: var(--layout-grid-gap-md);
}

@media (max-width: 768px) {
  .grid-container {
    grid-template-columns: var(--layout-grid-columns-1);
  }
}
```

---

## Animations & Transitions

### Durations
| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| instant | `--animation-duration-instant` | 0ms | No animation |
| immediate | `--animation-duration-immediate` | 50ms | Micro-interactions |
| **fast** | `--animation-duration-fast` | **150ms** | **Hover states** |
| moderate | `--animation-duration-moderate` | 250ms | Transitions |
| **normal** | `--animation-duration-normal` | **300ms** | **Standard** |
| deliberate | `--animation-duration-deliberate` | 400ms | Focus shifts |
| slow | `--animation-duration-slow` | 500ms | Emphasized |
| slower | `--animation-duration-slower` | 700ms | Complex |
| slowest | `--animation-duration-slowest` | 1000ms | Loading |

### Easing Functions
```css
--animation-easing-linear      /* Constant speed */
--animation-easing-ease        /* Default CSS ease */
--animation-easing-easeIn      /* cubic-bezier(0.4, 0, 1, 1) - Slow start, fast finish - exits */
--animation-easing-easeOut     /* cubic-bezier(0, 0, 0.2, 1) - Fast start, slow finish - entrances */
--animation-easing-easeInOut   /* cubic-bezier(0.4, 0, 0.2, 1) - Slow both - emphasis */
--animation-easing-sharp       /* cubic-bezier(0.4, 0, 0.6, 1) - Quick decisive */
--animation-easing-bouncing    /* cubic-bezier(0.68, -0.55, 0.265, 1.55) - Playful bounce */
--animation-easing-elastic     /* cubic-bezier(0.175, 0.885, 0.32, 1.275) - Subtle elastic */
--animation-easing-anticipate  /* cubic-bezier(0.25, 0.46, 0.45, 0.94) - Pull-back effect */
--animation-easing-overshoot   /* cubic-bezier(0.175, 0.885, 0.32, 1.275) - Goes past target */
```

### Animation Delays
| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| none | `--animation-delay-none` | 0ms | Immediate start |
| micro | `--animation-delay-micro` | 25ms | Sequencing |
| short | `--animation-delay-short` | 50ms | Staggered animations |
| medium | `--animation-delay-medium` | 100ms | Emphasis |
| long | `--animation-delay-long` | 200ms | Deliberate pause |
| extended | `--animation-delay-extended` | 300ms | Dramatic effect |

### Keyframe Animations

**Fade:**
```css
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes fadeOut { from { opacity: 1; } to { opacity: 0; } }
```

**Slide:**
```css
@keyframes slideInUp { from { transform: translateY(100%); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
@keyframes slideInDown { from { transform: translateY(-100%); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
@keyframes slideInLeft { from { transform: translateX(-100%); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
@keyframes slideInRight { from { transform: translateX(100%); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
```

**Scale:**
```css
@keyframes scaleIn { from { transform: scale(0); opacity: 0; } to { transform: scale(1); opacity: 1; } }
@keyframes scaleOut { from { transform: scale(1); opacity: 1; } to { transform: scale(0); opacity: 0; } }
```

**Effects:**
```css
@keyframes pulse { 0%, 100% { transform: scale(1); } 50% { transform: scale(1.05); } }
@keyframes bounce { 0%, 50%, 100% { transform: translateY(0); } 25% { transform: translateY(-10px); } 75% { transform: translateY(-5px); } }
@keyframes shake { 0%, 100% { transform: translateX(0); } 10%, 30%, 50%, 70%, 90% { transform: translateX(-10px); } 20%, 40%, 60%, 80% { transform: translateX(10px); } }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@keyframes wobble { 0%, 100% { transform: rotate(0deg); } 25% { transform: rotate(-5deg); } 75% { transform: rotate(5deg); } }
@keyframes heartbeat { 0%, 70%, 100% { transform: scale(1); } 14%, 42% { transform: scale(1.3); } }
```

**Loading:**
```css
@keyframes shimmer { from { background-position: -200px 0; } to { background-position: calc(200px + 100%) 0; } }
```

### Animation Presets

| Preset | Duration | Easing | Use Case |
|--------|----------|--------|----------|
| `hoverLift` | fast | easeOut | Subtle lift on hover |
| `buttonPress` | immediate | easeIn | Button press down |
| `modalEnter` | normal | easeOut | Modal entrance |
| `modalExit` | fast | easeIn | Modal exit |
| `toastSlideIn` | normal | bounce | Toast notification |
| `errorShake` | slower | linear | Error indication |
| `loadingSpinner` | slowest | linear | Infinite spinner |
| `pageTransition` | deliberate | easeInOut | Page-to-page |
| `accordionExpand` | normal | easeOut | Panel expansion |
| `progressBar` | slow | easeOut | Progress fill |
| `skeletonShimmer` | 1500ms | linear | Loading skeleton |
| `skeletonPulse` | 2000ms | easeInOut | Pulsing skeleton |

### Stagger Timing (for sequential animations)

| Context | Base Delay | Increment |
|---------|------------|-----------|
| List Items | 25ms | +25ms |
| Cards | 50ms | +50ms |
| Navigation | 50ms | +75ms |

**Stagger Example:**
```css
.list-item { animation: fadeIn var(--animation-duration-normal) var(--animation-easing-easeOut); }
.list-item:nth-child(1) { animation-delay: 25ms; }
.list-item:nth-child(2) { animation-delay: 50ms; }
.list-item:nth-child(3) { animation-delay: 75ms; }
/* ... */
```

### Reduced Motion Support
```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0ms !important;
    transition-duration: 0ms !important;
  }
}
```

### Common Transition Patterns
```css
.button {
  transition: all var(--animation-duration-fast) var(--animation-easing-easeOut);
}

.modal {
  transition:
    opacity var(--animation-duration-normal) var(--animation-easing-easeOut),
    transform var(--animation-duration-normal) var(--animation-easing-easeOut);
}

.card:hover {
  transform: translateY(-2px) scale(1.02);
  transition: transform var(--animation-duration-fast) var(--animation-easing-easeOut);
}
```

---

## Motion (Extended)

> Extended motion tokens for fine-grained animation control

### Motion Easing (Cubic Bezier)
```css
--motion-easing-linear       /* cubic-bezier(0, 0, 1, 1) */
--motion-easing-ease         /* cubic-bezier(0.25, 0.1, 0.25, 1) */
--motion-easing-ease-in      /* cubic-bezier(0.42, 0, 1, 1) */
--motion-easing-ease-out     /* cubic-bezier(0, 0, 0.58, 1) */
--motion-easing-ease-in-out  /* cubic-bezier(0.42, 0, 0.58, 1) */
--motion-easing-bounce       /* cubic-bezier(0.68, -0.55, 0.265, 1.55) */
--motion-easing-elastic      /* cubic-bezier(0.175, 0.885, 0.32, 1.275) */
--motion-easing-spring       /* cubic-bezier(0.16, 1, 0.3, 1) */
--motion-easing-smooth       /* cubic-bezier(0.4, 0, 0.2, 1) */
--motion-easing-sharp        /* cubic-bezier(0.4, 0, 0.6, 1) */
--motion-easing-emphasized   /* cubic-bezier(0.2, 0, 0, 1) */
--motion-easing-decelerated  /* cubic-bezier(0, 0, 0.2, 1) */
--motion-easing-accelerated  /* cubic-bezier(0.4, 0, 1, 1) */
--motion-easing-standard     /* cubic-bezier(0.2, 0, 0, 1) */
```

### Motion Durations
| Token | CSS Variable | Value |
|-------|--------------|-------|
| instant | `--motion-duration-instant` | 0ms |
| micro | `--motion-duration-micro` | 75ms |
| fast | `--motion-duration-fast` | 100ms |
| short | `--motion-duration-short` | 150ms |
| medium | `--motion-duration-medium` | 200ms |
| long | `--motion-duration-long` | 300ms |
| slow | `--motion-duration-slow` | 500ms |
| slower | `--motion-duration-slower` | 750ms |
| slowest | `--motion-duration-slowest` | 1000ms |

### Motion Delays
| Token | CSS Variable | Value |
|-------|--------------|-------|
| none | `--motion-delay-none` | 0ms |
| micro | `--motion-delay-micro` | 25ms |
| short | `--motion-delay-short` | 50ms |
| medium | `--motion-delay-medium` | 100ms |
| long | `--motion-delay-long` | 150ms |
| extra-long | `--motion-delay-extra-long` | 200ms |

### Motion Stagger (for list animations)
| Token | CSS Variable | Value |
|-------|--------------|-------|
| micro | `--motion-stagger-micro` | 25ms |
| short | `--motion-stagger-short` | 50ms |
| medium | `--motion-stagger-medium` | 75ms |
| long | `--motion-stagger-long` | 100ms |

---

## Gestures (Touch & Interaction)

> Tokens for touch-friendly interfaces and gesture handling

### Touch Targets
| Token | CSS Variable | Value | Use Case |
|-------|--------------|-------|----------|
| min | `--gesture-touch-target-min` | 44px | Minimum (WCAG) |
| comfortable | `--gesture-touch-target-comfortable` | 48px | Recommended |
| large | `--gesture-touch-target-large` | 56px | Important actions |
| extra-large | `--gesture-touch-target-extra-large` | 64px | Primary CTAs |

### Touch Padding
| Token | CSS Variable | Value |
|-------|--------------|-------|
| min | `--gesture-touch-padding-min` | 8px |
| comfortable | `--gesture-touch-padding-comfortable` | 12px |
| large | `--gesture-touch-padding-large` | 16px |

### Swipe Thresholds
| Token | CSS Variable | Value | Description |
|-------|--------------|-------|-------------|
| distance-min | `--gesture-swipe-distance-min` | 20px | Minimum swipe |
| distance-default | `--gesture-swipe-distance-default` | 50px | Standard swipe |
| distance-large | `--gesture-swipe-distance-large` | 100px | Large swipe |
| velocity-min | `--gesture-swipe-velocity-min` | 0.3 | Slow swipe |
| velocity-default | `--gesture-swipe-velocity-default` | 0.5 | Normal swipe |
| velocity-fast | `--gesture-swipe-velocity-fast` | 1.0 | Fast swipe |

### Swipe Resistance
| Token | CSS Variable | Value |
|-------|--------------|-------|
| light | `--gesture-swipe-resistance-light` | 0.1 |
| medium | `--gesture-swipe-resistance-medium` | 0.2 |
| strong | `--gesture-swipe-resistance-strong` | 0.4 |

### Drag Settings
```css
--gesture-drag-threshold-start   /* 5px - Start dragging */
--gesture-drag-threshold-cancel  /* 10px - Cancel threshold */
--gesture-drag-snap-distance     /* 20px - Snap to grid */
--gesture-drag-snap-force        /* 0.8 - Snap strength */
```

### Tap Timing
| Token | CSS Variable | Value | Description |
|-------|--------------|-------|-------------|
| single | `--gesture-tap-delay-single` | 0ms | Single tap |
| double | `--gesture-tap-delay-double` | 300ms | Double tap window |
| tolerance-movement | `--gesture-tap-tolerance-movement` | 10px | Movement tolerance |
| tolerance-time | `--gesture-tap-tolerance-time` | 500ms | Time tolerance |

### Long Press
| Token | CSS Variable | Value |
|-------|--------------|-------|
| short | `--gesture-press-duration-short` | 500ms |
| long | `--gesture-press-duration-long` | 800ms |
| extra-long | `--gesture-press-duration-extra-long` | 1200ms |

### Pinch/Zoom
```css
--gesture-pinch-scale-min    /* 0.5 - Minimum zoom */
--gesture-pinch-scale-max    /* 3.0 - Maximum zoom */
--gesture-pinch-scale-step   /* 0.1 - Zoom increment */
--gesture-pinch-threshold    /* 1.05 - Start threshold */
--gesture-pinch-sensitivity  /* 0.02 - Sensitivity */
```

### Scroll Physics
```css
--gesture-scroll-momentum-decay     /* 0.95 - Momentum decay rate */
--gesture-scroll-momentum-threshold /* 0.1 - Stop threshold */
--gesture-scroll-bounce-resistance  /* 0.3 - Edge bounce resistance */
--gesture-scroll-bounce-tension     /* 0.4 - Bounce tension */
```

### Gesture CSS Example
```css
.touch-target {
  min-height: var(--gesture-touch-target-comfortable);
  min-width: var(--gesture-touch-target-comfortable);
  padding: var(--gesture-touch-padding-comfortable);
}

.swipeable {
  touch-action: pan-x;
  /* Use JS to detect swipe with --gesture-swipe-distance-default threshold */
}

.draggable {
  cursor: grab;
  touch-action: none;
}
.draggable:active {
  cursor: grabbing;
}
```

---

## Border Width

| Token | CSS Variable | Value |
|-------|--------------|-------|
| 0 | `--borderWidth-0` | 0px |
| default | `--borderWidth-default` | 1px |
| 2 | `--borderWidth-2` | 2px |
| 4 | `--borderWidth-4` | 4px |
| 8 | `--borderWidth-8` | 8px |

---

## Opacity

| Token | CSS Variable | Value |
|-------|--------------|-------|
| 0 | `--opacity-0` | 0 (invisible) |
| 25 | `--opacity-25` | 0.25 |
| 50 | `--opacity-50` | 0.5 |
| 75 | `--opacity-75` | 0.75 |
| 100 | `--opacity-100` | 1 (fully visible) |

---

## Forms (Component Tokens)

> These are UnoCSS/Tailwind utility class compositions. Use directly or as reference.

### Input Field

**Sizes:**
| Size | Classes | Height |
|------|---------|--------|
| sm | `px-3 py-1.5 text-sm` | ~32px |
| md | `px-3 py-2 text-base` | ~40px |
| lg | `px-4 py-3 text-lg` | ~48px |

**States:**
| State | Border Color | Focus Ring |
|-------|--------------|------------|
| default | `border-color-neutral-300` | `ring-color-brand-primary-500` |
| error | `border-color-semantic-error-500` | `ring-color-semantic-error-500` |
| success | `border-color-semantic-success-500` | `ring-color-semantic-success-500` |

**Variants:**
| Variant | Style |
|---------|-------|
| default | `bg-color-neutral-white` |
| filled | `bg-color-neutral-50 border-transparent` |
| search | `bg-color-neutral-white pl-10` (with icon) |

**Input CSS Example:**
```css
.input {
  display: block;
  width: 100%;
  padding: var(--spacing-2) var(--spacing-3);
  font-size: var(--typography-fontSize-base);
  border: var(--borderWidth-default) solid var(--color-neutral-300);
  border-radius: var(--borderRadius-md);
  background: var(--color-neutral-white);
  transition: all var(--animation-duration-fast);
}
.input:focus {
  outline: none;
  border-color: var(--color-brand-primary-500);
  box-shadow: 0 0 0 2px var(--color-brand-primary-100);
}
.input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: var(--color-neutral-50);
}
.input--error {
  border-color: var(--color-semantic-error-500);
}
.input--error:focus {
  box-shadow: 0 0 0 2px var(--color-semantic-error-100);
}
```

### Button

**Sizes:**
| Size | Classes | Height |
|------|---------|--------|
| sm | `px-3 py-1.5 text-sm` | 32px (h-8) |
| md | `px-4 py-2 text-sm` | 40px (h-10) |
| lg | `px-6 py-3 text-base` | 48px (h-12) |

**Variants:**
| Variant | Background | Text | Border |
|---------|------------|------|--------|
| primary | `brand-primary-500` | `white` | `brand-primary-500` |
| secondary | `white` | `brand-primary-500` | `neutral-300` |
| outline | `transparent` | `brand-primary-500` | `brand-primary-500` (2px) |
| ghost | `transparent` | `brand-primary-500` | `transparent` |
| danger | `semantic-error-500` | `white` | `semantic-error-500` |
| success | `semantic-success-500` | `white` | `semantic-success-500` |
| warning | `semantic-warning-500` | `neutral-900` | `semantic-warning-500` |

**Button CSS Example:**
```css
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-4);
  font-size: var(--typography-fontSize-sm);
  font-weight: var(--typography-fontWeight-medium);
  border: var(--borderWidth-default) solid transparent;
  border-radius: var(--borderRadius-md);
  transition: all var(--animation-duration-fast) var(--animation-easing-easeOut);
}
.btn-primary {
  background: var(--color-brand-primary-500);
  color: var(--color-neutral-white);
  border-color: var(--color-brand-primary-500);
}
.btn-primary:hover {
  background: var(--color-brand-primary-600);
  border-color: var(--color-brand-primary-600);
  transform: translateY(-1px);
}
.btn-primary:active {
  background: var(--color-brand-primary-700);
  transform: translateY(0);
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

### Checkbox

**Sizes:**
| Size | Box | Icon | Label |
|------|-----|------|-------|
| sm | 16x16px (w-4 h-4) | 8x8px | text-sm |
| md | 20x20px (w-5 h-5) | 12x12px | text-base |
| lg | 24x24px (w-6 h-6) | 16x16px | text-lg |

**States:**
- Default: `border-color-neutral-300`
- Checked: `bg-color-brand-primary-500 border-color-brand-primary-500`
- Error: `border-color-semantic-error-500`
- Disabled: `opacity-50 cursor-not-allowed`

**Checkbox CSS Example:**
```css
.checkbox-box {
  width: 20px;
  height: 20px;
  border: 2px solid var(--color-neutral-300);
  border-radius: var(--borderRadius-sm);
  transition: all var(--animation-duration-fast);
}
.checkbox-input:checked + .checkbox-box {
  background: var(--color-brand-primary-500);
  border-color: var(--color-brand-primary-500);
}
.checkbox-input:focus + .checkbox-box {
  box-shadow: 0 0 0 2px var(--color-brand-primary-100);
}
```

### Radio Button

**Sizes:**
| Size | Circle | Inner Dot | Label |
|------|--------|-----------|-------|
| sm | 16x16px | 6x6px | text-sm |
| md | 20x20px | 8x8px | text-base |
| lg | 24x24px | 10x10px | text-lg |

**Radio CSS Example:**
```css
.radio-circle {
  width: 20px;
  height: 20px;
  border: 2px solid var(--color-neutral-300);
  border-radius: var(--borderRadius-full);
  transition: all var(--animation-duration-fast);
}
.radio-input:checked + .radio-circle {
  background: var(--color-brand-primary-500);
  border-color: var(--color-brand-primary-500);
}
.radio-inner {
  width: 8px;
  height: 8px;
  background: var(--color-neutral-white);
  border-radius: var(--borderRadius-full);
  opacity: 0;
}
.radio-input:checked ~ .radio-inner {
  opacity: 1;
}
```

### Form Labels & Helpers

```css
.label {
  display: block;
  font-size: var(--typography-fontSize-sm);
  font-weight: var(--typography-fontWeight-medium);
  color: var(--color-neutral-700);
  margin-bottom: var(--spacing-1);
}
.required-mark {
  color: var(--color-semantic-error-500);
  margin-left: var(--spacing-1);
}
.helper-text {
  margin-top: var(--spacing-1);
  font-size: var(--typography-fontSize-sm);
  color: var(--color-neutral-500);
}
.helper-text--error {
  color: var(--color-semantic-error-500);
}
.helper-text--success {
  color: var(--color-semantic-success-500);
}
```

### Form Group Layout

```css
.form-group {
  margin-bottom: var(--spacing-4);
}
.radio-group-vertical {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}
.radio-group-horizontal {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-4);
}
```

---

## Quick Reference: Common Patterns

### Button
```css
.button {
  font-family: var(--typography-fontFamily-primary);
  font-size: var(--typography-fontSize-sm);
  font-weight: var(--typography-fontWeight-medium);
  padding: var(--spacing-2) var(--spacing-4);
  border-radius: var(--borderRadius-md);
  background: var(--color-interactive-primary);
  color: var(--color-text-inverse);
  transition: background var(--animation-duration-fast) var(--animation-easing-easeOut);
}
.button:hover {
  background: var(--color-interactive-primaryHover);
}
```

### Card
```css
.card {
  background: var(--color-surface-primary);
  border: var(--borderWidth-default) solid var(--color-border-primary);
  border-radius: var(--borderRadius-lg);
  padding: var(--spacing-6);
  box-shadow: var(--shadow-card);
}
```

### Input
```css
.input {
  font-size: var(--typography-fontSize-base);
  padding: var(--spacing-2) var(--spacing-3);
  border: var(--borderWidth-default) solid var(--color-border-primary);
  border-radius: var(--borderRadius-md);
  background: var(--color-surface-primary);
  color: var(--color-text-primary);
}
.input::placeholder {
  color: var(--color-text-placeholder);
}
.input:focus {
  border-color: var(--color-border-focus);
  outline: none;
}
```

### Modal
```css
.modal-overlay {
  background: var(--color-surface-overlay);
  z-index: var(--zIndex-overlay);
}
.modal {
  background: var(--color-surface-primary);
  border-radius: var(--borderRadius-xl);
  box-shadow: var(--shadow-modal);
  z-index: var(--zIndex-modal);
}
```

### Text Hierarchy
```css
.heading-1 {
  font-size: var(--typography-fontSize-4xl);
  font-weight: var(--typography-fontWeight-bold);
  line-height: var(--typography-lineHeight-tight);
  color: var(--color-text-primary);
}
.body {
  font-size: var(--typography-fontSize-base);
  font-weight: var(--typography-fontWeight-normal);
  line-height: var(--typography-lineHeight-normal);
  color: var(--color-text-primary);
}
.caption {
  font-size: var(--typography-fontSize-xs);
  color: var(--color-text-secondary);
}
```

---

## Theme Switching

```js
// Set theme programmatically
document.documentElement.setAttribute('data-theme', 'dark');
document.documentElement.setAttribute('data-theme', 'light');

// Or use the theme utilities
import { setTheme, toggleTheme } from '@p9e/design-tokens/js/theme-utils';
setTheme('dark');
toggleTheme();
```

---

## Component Tokens

> Component-specific design tokens for consistent styling

### Alert

**Container:**
| Property | Value |
|----------|-------|
| Padding X | 16px |
| Padding Y | 12px |
| Border Radius | 8px |
| Border Width | 1px |
| Gap | 12px |
| Font Size | `--typography-fontSize-sm` |
| Font Weight | `--typography-fontWeight-medium` |
| Icon Size | 20px |

**Variants (Light/Dark):**
| Variant | Background | Text | Border | Icon |
|---------|------------|------|--------|------|
| Info | `semantic-info-50/900` | `semantic-info-800/200` | `semantic-info-200/700` | `semantic-info-600/400` |
| Success | `semantic-success-50/900` | `semantic-success-800/200` | `semantic-success-200/700` | `semantic-success-600/400` |
| Warning | `semantic-warning-50/900` | `semantic-warning-800/200` | `semantic-warning-200/700` | `semantic-warning-600/400` |
| Error | `semantic-error-50/900` | `semantic-error-800/200` | `semantic-error-200/700` | `semantic-error-600/400` |

**Alert CSS Example:**
```css
.alert {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid;
  font-size: var(--typography-fontSize-sm);
}
.alert--info {
  background: var(--color-semantic-info-50);
  border-color: var(--color-semantic-info-200);
  color: var(--color-semantic-info-800);
}
.alert--info .alert-icon {
  color: var(--color-semantic-info-600);
}
```

---

### Card

**Container:**
| Property | Size | Value |
|----------|------|-------|
| Border Radius | - | `--layout-radius-lg` |
| Border Width | - | `--borderWidth-default` |
| Border Color | light/dark | `neutral-200/700` |
| Background | light/dark | `white/neutral-900` |
| Padding | sm | `--spacing-3` (12px) |
| Padding | md | `--spacing-4` (16px) |
| Padding | lg | `--spacing-6` (24px) |
| Shadow | sm/md/lg | `--shadow-sm/md/lg` |

**Sections:**
| Section | Padding | Border |
|---------|---------|--------|
| Header | bottom: `--spacing-3` | `--borderWidth-default` |
| Body | y: `--spacing-3` | - |
| Footer | top: `--spacing-3` | `--borderWidth-default` |

**Card CSS Example:**
```css
.card {
  background: var(--color-surface-primary);
  border: var(--borderWidth-default) solid var(--color-border-primary);
  border-radius: var(--borderRadius-lg);
  box-shadow: var(--shadow-sm);
}
.card--md { padding: var(--spacing-4); }
.card-header {
  padding-bottom: var(--spacing-3);
  border-bottom: var(--borderWidth-default) solid var(--color-border-primary);
}
.card-footer {
  padding-top: var(--spacing-3);
  border-top: var(--borderWidth-default) solid var(--color-border-primary);
}
```

---

### Modal

**Sizes:**
| Size | Max Width |
|------|-----------|
| xs | max-w-xs |
| sm | max-w-sm |
| md | max-w-md |
| lg | max-w-lg |
| xl | max-w-xl |
| full | w-full h-full |

**Structure:**
| Part | Classes/Styles |
|------|----------------|
| Overlay | `fixed inset-0 z-50 bg-black/50 backdrop-blur-sm` |
| Dialog | `bg-white rounded-lg shadow-xl max-h-[90vh]` |
| Header | `px-6 py-4 border-b border-neutral-200` |
| Title | `text-lg font-semibold text-neutral-900` |
| Body | `flex-1 overflow-y-auto px-6 py-4` |
| Close Button | `text-neutral-400 hover:text-neutral-600 p-1 rounded` |

**Modal Type Icons:**
| Type | Background | Icon Color |
|------|------------|------------|
| Info | `semantic-info-100` | `semantic-info-500` |
| Success | `semantic-success-100` | `semantic-success-500` |
| Warning | `semantic-warning-100` | `semantic-warning-500` |
| Error | `semantic-error-100` | `semantic-error-500` |

---

### Toast / Notification

**Container:**
| Property | Value |
|----------|-------|
| Min Width | 300px |
| Max Width | 500px |
| Padding X | 16px |
| Padding Y | 12px |
| Border Radius | 8px |
| Shadow | `--shadow-lg` |

**Positioning:**
| Position | CSS |
|----------|-----|
| Top Right | `top: 16px; right: 16px;` |
| Top Left | `top: 16px; left: 16px;` |
| Top Center | `top: 16px; left: 50%; transform: translateX(-50%);` |
| Bottom Right | `bottom: 16px; right: 16px;` |
| Bottom Left | `bottom: 16px; left: 16px;` |
| Bottom Center | `bottom: 16px; left: 50%; transform: translateX(-50%);` |

**Duration:**
| Duration | Value | Use Case |
|----------|-------|----------|
| short | 3000ms | Quick info |
| medium | 5000ms | Standard |
| long | 8000ms | Important |
| persistent | 0ms | Manual dismiss |

**Animation:**
| State | Duration | Easing |
|-------|----------|--------|
| Enter | 300ms | `cubic-bezier(0.16, 1, 0.3, 1)` |
| Exit | 200ms | `cubic-bezier(0.4, 0, 1, 1)` |

**Progress Bar:**
| Property | Value |
|----------|-------|
| Height | 3px |
| Background | `neutral-200/600` |
| Fill (by type) | `semantic-{type}-500/400` |

---

### Spinner / Loader

**Sizes:**
| Size | Dimensions |
|------|------------|
| xs | 12x12px (w-3 h-3) |
| sm | 16x16px (w-4 h-4) |
| md | 24x24px (w-6 h-6) |
| lg | 32x32px (w-8 h-8) |
| xl | 48x48px (w-12 h-12) |

**Colors:**
| Color | CSS Class |
|-------|-----------|
| primary | `text-brand-primary-500` |
| secondary | `text-brand-secondary-500` |
| success | `text-semantic-success-500` |
| warning | `text-semantic-warning-500` |
| error | `text-semantic-error-500` |
| neutral | `text-neutral-500` |

**Variants:**
| Variant | Style |
|---------|-------|
| default | `animate-spin rounded-full border-2 border-current border-t-transparent` |
| dots | Three pulsing dots |
| bars | Three bouncing bars |
| pulse | Single pulsing circle |

**Spinner CSS Example:**
```css
.spinner {
  display: inline-block;
  width: 24px;
  height: 24px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
```

---

### Table

**Row Heights:**
| Density | Height |
|---------|--------|
| compact | 32px |
| default | 40px |
| comfortable | 48px |
| spacious | 56px |

**Cell Padding:**
| Size | X | Y |
|------|---|---|
| sm | 8px | 6px |
| md | 12px | 8px |
| lg | 16px | 12px |

**Header:**
| Property | Light | Dark |
|----------|-------|------|
| Background | `neutral-50` | `neutral-800` |
| Text Color | `neutral-700` | `neutral-200` |
| Font Weight | `semibold` | `semibold` |
| Border Color | `neutral-200` | `neutral-700` |

**Row States:**
| State | Light Background | Dark Background |
|-------|------------------|-----------------|
| Default | `white` | `neutral-900` |
| Hover | `neutral-50` | `neutral-800` |
| Selected | `brand-primary-50` | `brand-primary-900` |
| Striped | `neutral-25` | `neutral-850` |

**Sorting Icon:**
| State | Light Color | Dark Color |
|-------|-------------|------------|
| Inactive | `neutral-400` | `neutral-500` |
| Active | `brand-primary-600` | `brand-primary-400` |

**Pagination Button:**
| Property | Value |
|----------|-------|
| Size | 32px |
| Border Radius | 4px |
| Gap | 4px |

---

### Navigation

**Menu Item:**
| Size | Height | Padding X | Padding Y |
|------|--------|-----------|-----------|
| sm | 32px | 8px | 6px |
| md | 40px | 12px | 8px |
| lg | 48px | 16px | 12px |

**Menu Item States:**
| State | Light BG | Dark BG | Light Text | Dark Text |
|-------|----------|---------|------------|-----------|
| Default | transparent | transparent | `neutral-700` | `neutral-300` |
| Hover | `neutral-100` | `neutral-800` | `neutral-900` | `neutral-100` |
| Active | `brand-primary-50` | `brand-primary-900` | `brand-primary-700` | `brand-primary-300` |
| Disabled | transparent | transparent | `neutral-400` | `neutral-600` |

**Active Indicator:**
| Property | Value |
|----------|-------|
| Width | 3px |
| Color (light/dark) | `brand-primary-600/400` |

**Sidebar:**
| Property | Value |
|----------|-------|
| Collapsed Width | 64px |
| Expanded Width | 240px |
| Wide Width | 280px |
| Shadow | `--shadow-sm` |

**Header:**
| Property | Value |
|----------|-------|
| Height | 64px |
| Padding X | 16px |
| Padding Y | 12px |
| Shadow | `--shadow-sm` |

**Tab States:**
| State | Light BG | Light Text | Light Border |
|-------|----------|------------|--------------|
| Default | transparent | `neutral-600` | transparent |
| Hover | `neutral-100` | `neutral-800` | - |
| Active | `white` | `brand-primary-600` | `brand-primary-500` |

---

### Dropdown

**Menu:**
| Property | Value |
|----------|-------|
| Min Width | 160px |
| Max Width | 320px |
| Padding | 4px |
| Border Radius | 8px |
| Shadow | `--shadow-lg` |

**Item:**
| Property | Value |
|----------|-------|
| Padding X | `--spacing-3` |
| Padding Y | `--spacing-2` |
| Border Radius | `--borderRadius-base` |
| Font Size | `--typography-fontSize-sm` |

---

### Tooltip

| Property | Value |
|----------|-------|
| Background | `--color-surface-inverse` |
| Text Color | `--color-text-inverse` |
| Border Radius | `--borderRadius-md` |
| Padding X | `--spacing-3` |
| Padding Y | `--spacing-2` |
| Font Size | `--typography-fontSize-xs` |
| Font Weight | `--typography-fontWeight-medium` |
| Max Width | 200px |
| Shadow | `--shadow-lg` |
| Z-Index | `--zIndex-tooltip` |

---

### Badge

**Base:**
| Property | Value |
|----------|-------|
| Border Radius | `--borderRadius-full` |
| Padding X | `--spacing-2` |
| Padding Y | `--spacing-1` |
| Font Size | `--typography-fontSize-xs` |
| Font Weight | `--typography-fontWeight-medium` |
| Line Height | 1 |

**Variants:**
| Variant | Background | Text |
|---------|------------|------|
| Default | `--color-interactive-secondary` | `--color-text-primary` |
| Primary | `--color-interactive-primary` | `--color-text-inverse` |
| Success | `semantic-success-50` | `semantic-success-900` |
| Warning | `semantic-warning-50` | `semantic-warning-900` |
| Error | `semantic-error-50` | `semantic-error-900` |

---

### Notification Badge (Dot)

| Property | Value |
|----------|-------|
| Size | 8px |
| Background | `semantic-error-500/400` |
| Border | 2px solid `white/neutral-900` |
| Offset X | -2px |
| Offset Y | -2px |

**Count Badge:**
| Property | Value |
|----------|-------|
| Min Width | 16px |
| Padding X | 4px |
| Padding Y | 2px |
| Font Size | 10px |
| Font Weight | bold |
| Border Radius | 10px |

---

### Progress

| Property | Value |
|----------|-------|
| Height | 3px (default) |
| Background | `neutral-200/600` |
| Fill | `brand-primary-500` |
| Border Radius | full |

---

### List & ListItem

**List Container:**
| Property | Value |
|----------|-------|
| Width | `w-full` |

**List Variants:**
| Variant | Classes |
|---------|---------|
| simple | (no additional styles) |
| divided | `divide-y divide-color-neutral-200` |
| bordered | `border border-color-neutral-200 rounded-lg divide-y divide-color-neutral-200` |

**List Sizes:**
| Size | Font Size |
|------|-----------|
| xs | `text-xs` |
| sm | `text-sm` |
| md | `text-sm` |
| lg | `text-base` |
| xl | `text-lg` |

**ListItem:**
| State | Classes |
|-------|---------|
| base | `flex items-center gap-3` |
| clickable | `cursor-pointer hover:bg-color-neutral-50 transition-colors` |
| active | `bg-color-brand-primary-50 text-color-brand-primary-700` |
| disabled | `opacity-50 cursor-not-allowed` |

**ListItem Size Padding:**
| Size | Padding |
|------|---------|
| xs | `px-2 py-1` |
| sm | `px-3 py-1.5` |
| md | `px-4 py-2` |
| lg | `px-4 py-3` |
| xl | `px-6 py-4` |

---

### Tree

**Container:**
| Property | Classes |
|----------|---------|
| container | `w-full` |
| node | `select-none` |

**Node Content:**
| State | Classes |
|-------|---------|
| base | `flex items-center gap-1 py-1 px-2 rounded hover:bg-color-neutral-100 transition-colors` |
| selected | `bg-color-brand-primary-50 text-color-brand-primary-700 hover:bg-color-brand-primary-100` |
| disabled | `opacity-50 cursor-not-allowed hover:bg-transparent` |

**Tree Elements:**
| Element | Classes |
|---------|---------|
| expandIcon | `w-4 h-4 transition-transform shrink-0` |
| expandIcon (expanded) | `rotate-90` |
| checkbox | `mr-1` |
| icon | `w-4 h-4 shrink-0` |
| label | `truncate` |
| children | `ml-4` |
| children (with lines) | `border-l border-color-neutral-200 ml-2` |

---

### Picture

**Object Fit:**
| Fit | Class |
|-----|-------|
| contain | `object-contain` |
| cover | `object-cover` |
| fill | `object-fill` |
| none | `object-none` |
| scale-down | `object-scale-down` |

**Border Radius:**
| Rounded | Class |
|---------|-------|
| none | (no class) |
| sm | `rounded-sm` |
| md | `rounded-md` |
| lg | `rounded-lg` |
| xl | `rounded-xl` |
| full | `rounded-full` |

---

### NumberInput

**Container:**
| Property | Classes |
|----------|---------|
| container | `w-full` |
| wrapper | `flex` |

**Input:**
| State | Classes |
|-------|---------|
| base | `w-full border border-color-neutral-300 bg-color-neutral-white text-color-neutral-900 text-right font-mono tabular-nums` |
| focus | `focus:outline-none focus:ring-2 focus:ring-color-brand-primary-500` |
| disabled | `opacity-50 cursor-not-allowed` |
| with stepper | `rounded-none` |
| without stepper | `rounded-lg` |

**Input Sizes:**
| Size | Classes |
|------|---------|
| sm | `h-8 text-sm` |
| md | `h-10 text-base` |
| lg | `h-12 text-lg` |

**Stepper Buttons:**
| State | Classes |
|-------|---------|
| base | `flex items-center justify-center border border-color-neutral-300 bg-color-neutral-100 text-color-neutral-600` |
| hover | `hover:bg-color-neutral-200` |
| focus | `focus:outline-none focus:ring-2 focus:ring-inset focus:ring-color-brand-primary-500` |
| disabled | `opacity-50 cursor-not-allowed` |
| decrement | `rounded-l-lg border-r-0` |
| increment | `rounded-r-lg border-l-0` |

**Stepper Sizes:**
| Size | Dimensions |
|------|------------|
| sm | `w-6 h-8` |
| md | `w-8 h-10` |
| lg | `w-10 h-12` |

---

### CurrencyInput

**Container:**
| Property | Classes |
|----------|---------|
| container | `w-full` |
| wrapper | `flex` |

**Input:**
| State | Classes |
|-------|---------|
| base | `flex-1 rounded-r-lg border border-color-neutral-300 bg-color-neutral-white text-color-neutral-900 text-right font-mono tabular-nums px-3` |
| focus | `focus:outline-none focus:ring-2 focus:ring-color-brand-primary-500` |
| disabled | `opacity-50 cursor-not-allowed` |

**Currency Symbol:**
| Property | Classes |
|----------|---------|
| symbol | `flex items-center justify-center rounded-l-lg border border-r-0 border-color-neutral-300 bg-color-neutral-100 text-color-neutral-600 px-3` |

**Currency Selector Trigger:**
| State | Classes |
|-------|---------|
| base | `flex items-center justify-between gap-1 rounded-l-lg border border-r-0 border-color-neutral-300 bg-color-neutral-100 text-color-neutral-900` |
| hover | `hover:bg-color-neutral-200` |
| focus | `focus:outline-none focus:ring-2 focus:ring-inset focus:ring-color-brand-primary-500` |
| disabled | `opacity-50 cursor-not-allowed` |

**Currency Dropdown:**
| Property | Classes |
|----------|---------|
| dropdown | `absolute z-dropdown mt-1 w-48 max-h-60 overflow-auto bg-color-neutral-white border border-color-neutral-200 rounded-lg shadow-lg` |
| option base | `w-full px-3 py-2 text-left text-sm` |
| option hover | `hover:bg-color-neutral-100` |
| option selected | `bg-color-brand-primary-500 text-color-neutral-white` |

---

### StatCard

**Container:**
| State | Classes |
|-------|---------|
| base | `block p-5 rounded-xl bg-color-neutral-white border border-color-neutral-200` |
| clickable | `hover:border-color-brand-primary-500 hover:shadow-md transition-all` |

**Layout:**
| Element | Classes |
|---------|---------|
| header | `flex items-start justify-between` |
| content | `flex-1` |
| icon container | `p-3 rounded-lg` |

**Typography:**
| Element | Classes |
|---------|---------|
| title | `text-sm font-medium text-color-neutral-600` |
| value | `mt-2 text-2xl font-bold text-color-neutral-900 tabular-nums` |
| loading skeleton | `mt-2 h-8 w-24 bg-color-neutral-200 rounded animate-pulse` |

**Trend:**
| Element | Classes |
|---------|---------|
| container | `flex items-center gap-2 mt-2` |
| badge base | `inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium` |
| badge up | `bg-color-semantic-success-100 text-color-semantic-success-700` |
| badge down | `bg-color-semantic-error-100 text-color-semantic-error-700` |
| badge neutral | `bg-color-neutral-200 text-color-neutral-600` |
| label | `text-xs text-color-neutral-500` |

---

### ActivityFeed

**Container:**
| Property | Classes |
|----------|---------|
| container | `space-y-4` |

**Group:**
| Element | Classes |
|---------|---------|
| container | `mb-6` |
| label | `text-sm font-medium text-color-neutral-500 mb-3` |
| items | `space-y-4` |

**Item:**
| State | Classes |
|-------|---------|
| base | `flex gap-3 w-full text-left p-2 -mx-2 rounded-lg transition-colors` |
| hover | `hover:bg-color-neutral-100` |

**Avatar:**
| Element | Classes |
|---------|---------|
| container | `relative flex-shrink-0` |
| image | `w-8 h-8 rounded-full` |
| fallback | `w-8 h-8 rounded-full bg-color-brand-primary-500 flex items-center justify-center text-color-neutral-white text-sm font-medium` |

**Avatar Badge (Activity Type):**
| Type | Background Color |
|------|-----------------|
| create | `bg-color-semantic-success-500` |
| update | `bg-color-semantic-info-500` |
| delete | `bg-color-semantic-error-500` |
| comment | `bg-color-brand-primary-500` |
| assign | `bg-color-semantic-warning-500` |
| status | `bg-color-semantic-success-500` |
| custom | `bg-color-neutral-500` |

**Content:**
| Element | Classes |
|---------|---------|
| container | `flex-1 min-w-0` |
| text | `text-sm text-color-neutral-900` |
| actor | `font-medium` |
| action | `text-color-neutral-600` |
| target | `font-medium` |

**Other:**
| Element | Classes |
|---------|---------|
| timestamp | `text-xs text-color-neutral-500 mt-0.5` |
| loadMore | `w-full mt-4 py-2 text-sm text-color-brand-primary-500 hover:underline` |
| icon size | `w-4 h-4` |

---

### Breadcrumbs

**Container:**
| Property | Classes |
|----------|---------|
| base | `flex items-center flex-wrap` |
| gap sm | `gap-1` |
| gap md | `gap-2` |
| gap lg | `gap-3` |

**Item:**
| Property | Classes |
|----------|---------|
| base | `flex items-center gap-1` |

**Link:**
| State | Classes |
|-------|---------|
| base | `flex items-center gap-1 transition-colors duration-200` |
| default | `text-color-neutral-500 hover:text-color-neutral-900` |
| current | `text-color-neutral-900 font-medium cursor-default` |
| disabled | `text-color-neutral-400 cursor-not-allowed` |

**Link Sizes:**
| Size | Class |
|------|-------|
| sm | `text-sm` |
| md | `text-base` |
| lg | `text-lg` |

**Icon Sizes:**
| Size | Class |
|------|-------|
| sm | `w-3 h-3` |
| md | `w-4 h-4` |
| lg | `w-5 h-5` |

**Separator:**
| Property | Classes |
|----------|---------|
| base | `flex-shrink-0 text-color-neutral-400` |
| slash | `mx-1` |
| chevron | `mx-1` |
| arrow | `mx-2` |

---

---

## Services & Utilities

### Form Validation (Zod/Yup Integration)

**Schema Adapters:**
```typescript
import {
  zodToRule,
  yupToRule,
  validateWithZod,
  validateWithYup,
  SchemaValidator,
  createZodFormValidator,
  createYupFormValidator
} from '@samavāya/utility/validation';

// Create validator from Zod schema
const validator = SchemaValidator.fromZod(myZodSchema);
const result = validator.validate(data);
// { success: true/false, data?, errors: { path: string; message: string; }[] }

// Create validator from Yup schema
const yupValidator = SchemaValidator.fromYup(myYupSchema);

// Convert schema to ValidationRule for use with FormValidator
const rule = zodToRule(z.string().email(), 'Invalid email');
const yupRule = yupToRule(yup.string().email(), 'Invalid email');

// Quick validation
const isValid = validateWithZod(z.string().email(), 'test@example.com');
const isValidYup = validateWithYup(yup.string().email(), 'test@example.com');

// Create full FormValidator from schema
const formValidator = createZodFormValidator(zodSchema);
const yupFormValidator = createYupFormValidator(yupSchema);
```

**SchemaValidator Class:**
| Method | Description |
|--------|-------------|
| `fromZod(schema)` | Create validator from Zod schema |
| `fromYup(schema)` | Create validator from Yup schema |
| `validate(data)` | Sync validation, returns `SchemaValidationResult` |
| `validateAsync(data)` | Async validation, returns `Promise<SchemaValidationResult>` |
| `getFieldRule(path, msg?)` | Get ValidationRule for specific field path |

**SchemaValidationResult:**
```typescript
interface SchemaValidationResult<T> {
  success: boolean;
  data?: T;
  errors: { path: string; message: string; }[];
}
```

---

### Table Column Resize

**Column Resize Action:**
```svelte
<script>
  import { columnResize } from '@samavāya/ui';

  let columnWidths = { name: 200, email: 250, status: 100 };

  function handleResize(columnKey: string, width: number) {
    columnWidths = { ...columnWidths, [columnKey]: width };
  }
</script>

<th
  use:columnResize={{
    columnKey: 'name',
    minWidth: 100,
    maxWidth: 400,
    onResize: handleResize,
    onResizeStart: (key, width) => console.log('Started resizing', key),
    onResizeEnd: (key, width) => console.log('Finished resizing', key),
  }}
>
  Name
</th>
```

**ColumnResizeOptions:**
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `columnKey` | `string` | required | Unique column identifier |
| `minWidth` | `number` | `50` | Minimum column width in px |
| `maxWidth` | `number` | `Infinity` | Maximum column width in px |
| `onResizeStart` | `function` | - | `(columnKey, initialWidth) => void` |
| `onResize` | `function` | - | `(columnKey, width) => void` |
| `onResizeEnd` | `function` | - | `(columnKey, finalWidth) => void` |

**CSS Classes:**
```css
/* columnResizeClasses */
.column-resize-handle { ... }   /* Resize handle element */
.column-resizing { ... }        /* While actively resizing */
```

---

### Table Column Reorder

**Column Reorder Action:**
```svelte
<script>
  import { columnReorder, reorderColumns } from '@samavāya/ui';

  let columns = [
    { key: 'name', header: 'Name' },
    { key: 'email', header: 'Email' },
    { key: 'status', header: 'Status' },
  ];

  function handleReorder(sourceKey: string, targetKey: string, position: 'before' | 'after') {
    columns = reorderColumns(columns, sourceKey, targetKey, position);
  }
</script>

{#each columns as column}
  <th
    use:columnReorder={{
      columnKey: column.key,
      group: 'my-table',
      draggable: true,
      droppable: true,
      onReorder: handleReorder,
    }}
  >
    {column.header}
  </th>
{/each}
```

**ColumnReorderOptions:**
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `columnKey` | `string` | required | Unique column identifier |
| `group` | `string` | `'default'` | Group for restricting drag between tables |
| `draggable` | `boolean` | `true` | Whether column can be dragged |
| `droppable` | `boolean` | `true` | Whether column accepts drops |
| `onDragStart` | `function` | - | `(columnKey) => void` |
| `onDragOver` | `function` | - | `(sourceKey, targetKey) => void` |
| `onReorder` | `function` | - | `(sourceKey, targetKey, position) => void` |
| `onDragEnd` | `function` | - | `(columnKey) => void` |

**Helper Function:**
```typescript
// Reorder columns array based on drag result
const newColumns = reorderColumns(columns, sourceKey, targetKey, position);
```

**CSS Classes:**
```css
/* columnReorderClasses */
.cursor-grab { ... }         /* Draggable cursor */
.opacity-50 { ... }          /* While dragging */
.column-drag-over { ... }    /* When dragged over */
.column-drop-indicator { ... } /* Visual drop indicator */
```

---

### Modal Stack Management Service

**Programmatic Modal Opening:**
```typescript
import {
  modalStack,
  openModal,
  openDialog,
  openDrawer,
  closeModal,
  closeTopModal,
  closeAllModals
} from '@samavāya/ui';

// Open a simple modal
const result = await openModal({
  title: 'Edit User',
  size: 'md',
  content: 'Modal content here',
});
// result: { confirmed: boolean, data?: unknown }

// Open a confirmation dialog
const confirmed = await openDialog({
  title: 'Delete Item?',
  message: 'This action cannot be undone.',
  variant: 'warning',
  confirmText: 'Delete',
  cancelText: 'Cancel',
  destructive: true,
});
// result: { confirmed: true/false }

// Open a drawer
const drawerResult = await openDrawer({
  title: 'User Details',
  position: 'right',
  size: 'md',
});

// Shorthand methods
await modalStack.alert({ title: 'Info', message: 'Something happened' });
await modalStack.confirm({ title: 'Confirm', message: 'Are you sure?' });
await modalStack.warning({ title: 'Warning', message: 'Be careful!' });
await modalStack.error({ title: 'Error', message: 'Something went wrong' });
await modalStack.success({ title: 'Success', message: 'Operation completed!' });

// Close modals
closeModal(modalId);      // Close specific modal by ID
closeTopModal();          // Close topmost modal
closeAllModals();         // Close all modals

// With result
modalStack.close(id, { confirmed: true, data: { foo: 'bar' } });
modalStack.confirmTop(data);  // Confirm and close top with data
modalStack.cancelTop();       // Cancel and close top
```

**ModalStackRenderer Component:**
```svelte
<!-- Place once at app root (e.g., +layout.svelte) -->
<script>
  import { ModalStackRenderer } from '@samavāya/ui';
</script>

<slot />
<ModalStackRenderer />
```

**Modal with Custom Component:**
```typescript
import MyFormComponent from './MyFormComponent.svelte';

const result = await openModal({
  title: 'Create Item',
  component: MyFormComponent,
  componentProps: { itemId: 123 },
});

// In MyFormComponent.svelte:
// dispatch('confirm', formData) to close with data
// dispatch('cancel') to close without data
```

**Configuration Options:**

| ModalConfig | Type | Default | Description |
|-------------|------|---------|-------------|
| `title` | `string` | - | Modal title |
| `size` | `Size \| 'full'` | `'md'` | Modal size |
| `closeOnBackdrop` | `boolean` | `true` | Close on backdrop click |
| `closeOnEscape` | `boolean` | `true` | Close on Escape key |
| `showClose` | `boolean` | `true` | Show close button |
| `centered` | `boolean` | `true` | Center vertically |
| `preventScroll` | `boolean` | `true` | Prevent body scroll |
| `content` | `string` | - | Text content |
| `component` | `Component` | - | Svelte component |
| `componentProps` | `object` | - | Props for component |

| DialogConfig | Type | Default | Description |
|--------------|------|---------|-------------|
| `variant` | `'info' \| 'warning' \| 'error' \| 'success' \| 'confirm'` | `'confirm'` | Dialog variant |
| `message` | `string` | - | Dialog message |
| `confirmText` | `string` | `'Confirm'` | Confirm button text |
| `cancelText` | `string` | `'Cancel'` | Cancel button text |
| `destructive` | `boolean` | `false` | Destructive action styling |

| DrawerConfig | Type | Default | Description |
|--------------|------|---------|-------------|
| `position` | `'left' \| 'right' \| 'top' \| 'bottom'` | `'right'` | Drawer position |
| `overlay` | `boolean` | `true` | Show overlay backdrop |

**Reactive Stores:**
```typescript
import { modalStack } from '@samavāya/ui';

// Subscribe to stack changes
$: items = $modalStack.stack;
$: hasModals = $modalStack.hasOpenModals;
$: topModal = $modalStack.topModal;
$: modalCount = $modalStack.count;

// Query methods
modalStack.isOpen(id);    // Check if modal is open
modalStack.getModal(id);  // Get modal by ID
```

---

## File Locations

| Purpose | Path |
|---------|------|
| Source tokens | `packages/design-tokens/tokens/` |
| CSS output | `packages/design-tokens/dist/css/tokens.css` |
| JS output | `packages/design-tokens/dist/js/tokens.js` |
| Theme utils | `packages/design-tokens/dist/js/theme-utils.js` |
| Type defs | `packages/design-tokens/dist/types/tokens.d.ts` |
| Validation utils | `packages/utility/src/validation/` |
| Schema adapters | `packages/utility/src/validation/schema-adapters.ts` |
| Column actions | `packages/ui/src/actions/columnResize.ts`, `columnReorder.ts` |
| Modal service | `packages/ui/src/services/modal-stack.ts` |
| Modal renderer | `packages/ui/src/services/ModalStackRenderer.svelte` |
