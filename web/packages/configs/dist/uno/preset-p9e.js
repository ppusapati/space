import { designTokensTheme, componentShortcuts, animations } from '@p9e.in/samavaya/uno'

/**
 * P9E Enterprise UnoCSS Preset
 *
 * Comprehensive preset for enterprise applications with:
 * - SCADA system styling utilities
 * - ERP interface patterns
 * - IoT device status indicators
 * - Enterprise data visualization
 * - Industrial UI patterns
 * - High-contrast accessibility
 */
export function presetP9E() {
  return {
    name: '@p9e.in/preset-p9e',
    theme: designTokensTheme,
    shortcuts: [
      // Import component shortcuts from design tokens
      componentShortcuts,

      // Enterprise Layout Patterns
      {
        'enterprise-container': 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8',
        'enterprise-card': 'bg-white dark:bg-neutral-800 shadow-sm border border-neutral-200 dark:border-neutral-700 rounded-lg',
        'enterprise-header': 'bg-white dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700 px-6 py-4',
        'enterprise-sidebar': 'bg-neutral-50 dark:bg-neutral-900 border-r border-neutral-200 dark:border-neutral-700 w-64 h-full',

        // Data Display Patterns
        'data-grid': 'grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
        'metrics-grid': 'grid gap-6 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
        'dashboard-layout': 'grid gap-6 grid-cols-1 lg:grid-cols-3',

        // SCADA Status Patterns
        'status-indicator': 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
        'status-online': 'bg-success-50 text-success-700 border border-success-200',
        'status-offline': 'bg-neutral-100 text-neutral-600 border border-neutral-300',
        'status-warning': 'bg-warning-50 text-warning-700 border border-warning-200',
        'status-critical': 'bg-error-50 text-error-700 border border-error-200',
        'status-maintenance': 'bg-blue-50 text-blue-700 border border-blue-200',

        // ERP Interface Patterns
        'erp-toolbar': 'flex items-center justify-between px-4 py-2 bg-neutral-50 dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700',
        'erp-form-section': 'bg-white dark:bg-neutral-800 p-6 rounded-lg border border-neutral-200 dark:border-neutral-700 space-y-4',
        'erp-field-group': 'grid grid-cols-1 md:grid-cols-2 gap-4',
        'erp-table': 'min-w-full divide-y divide-neutral-200 dark:divide-neutral-700',
        'erp-table-header': 'bg-neutral-50 dark:bg-neutral-900',
        'erp-table-cell': 'px-6 py-4 whitespace-nowrap text-sm text-neutral-900 dark:text-neutral-100',

        // IoT Device Patterns
        'device-card': 'bg-white dark:bg-neutral-800 p-4 rounded-lg border border-neutral-200 dark:border-neutral-700 hover:shadow-md transition-shadow',
        'device-header': 'flex items-center justify-between mb-4',
        'device-metric': 'flex flex-col items-center p-4 bg-neutral-50 dark:bg-neutral-900 rounded-lg',
        'device-reading': 'text-2xl font-bold text-neutral-900 dark:text-neutral-100',
        'device-label': 'text-sm text-neutral-600 dark:text-neutral-400 mt-1',

        // Industrial UI Patterns
        'control-panel': 'bg-neutral-900 dark:bg-black p-6 rounded-lg border-2 border-neutral-700',
        'control-button': 'flex items-center justify-center w-12 h-12 rounded-lg border-2 font-semibold text-sm transition-all',
        'emergency-stop': 'bg-error-600 hover:bg-error-700 border-error-700 text-white',
        'start-button': 'bg-success-600 hover:bg-success-700 border-success-700 text-white',
        'stop-button': 'bg-warning-600 hover:bg-warning-700 border-warning-700 text-white',

        // High Contrast & Accessibility
        'high-contrast': 'contrast-125 saturate-150',
        'focus-enterprise': 'focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:outline-none',

        // Data Visualization Base
        'chart-container': 'bg-white dark:bg-neutral-800 p-4 rounded-lg border border-neutral-200 dark:border-neutral-700',
        'legend-item': 'flex items-center space-x-2 text-sm text-neutral-600 dark:text-neutral-400',
      }
    ],
    rules: [
      // SCADA System Rules
      [/^scada-status-(.+)$/, ([, status]) => {
        const statusColors = {
          'running': 'var(--color-semantic-success-500)',
          'stopped': 'var(--color-neutral-500)',
          'error': 'var(--color-semantic-error-500)',
          'warning': 'var(--color-semantic-warning-500)',
          'maintenance': 'var(--color-blue-500)',
          'offline': 'var(--color-neutral-400)'
        }

        const color = statusColors[status] || statusColors['stopped']
        return {
          'background-color': color,
          'box-shadow': `0 0 8px ${color}40`,
        }
      }],

      // IoT Device Signal Strength
      [/^signal-strength-([1-5])$/, ([, strength]) => {
        const bars = parseInt(strength, 10)
        const opacity = Math.min(bars / 5, 1)
        return {
          'background': `linear-gradient(to right, var(--color-semantic-success-500) ${bars * 20}%, var(--color-neutral-300) ${bars * 20}%)`,
          'opacity': opacity.toString()
        }
      }],

      // Industrial Gauge Styles
      [/^gauge-value-(\d+)$/, ([, value]) => {
        const percentage = Math.min(parseInt(value, 10), 100)
        let color = 'var(--color-semantic-success-500)'

        if (percentage > 80) color = 'var(--color-semantic-error-500)'
        else if (percentage > 60) color = 'var(--color-semantic-warning-500)'

        return {
          'background': `conic-gradient(${color} ${percentage * 3.6}deg, var(--color-neutral-200) ${percentage * 3.6}deg)`,
          'border-radius': '50%'
        }
      }],

      // ERP Priority Levels
      [/^priority-(low|medium|high|critical)$/, ([, priority]) => {
        const priorityColors = {
          'low': 'var(--color-blue-500)',
          'medium': 'var(--color-semantic-warning-500)',
          'high': 'var(--color-orange-500)',
          'critical': 'var(--color-semantic-error-500)'
        }

        const color = priorityColors[priority]
        return {
          'border-left': `4px solid ${color}`,
          'background-color': `${color}10`
        }
      }],

      // Temperature Indicators
      [/^temp-([0-9]+)$/, ([, temp]) => {
        const temperature = parseInt(temp, 10)
        let color = 'var(--color-blue-500)' // Cold

        if (temperature > 80) color = 'var(--color-semantic-error-500)' // Hot
        else if (temperature > 60) color = 'var(--color-orange-500)' // Warm
        else if (temperature > 40) color = 'var(--color-semantic-warning-500)' // Moderate
        else if (temperature > 20) color = 'var(--color-semantic-success-500)' // Normal

        return {
          'background-color': color,
          'color': temperature > 40 ? 'white' : 'var(--color-neutral-900)'
        }
      }],

      // Data Freshness Indicators
      [/^data-age-(fresh|stale|old)$/, ([, age]) => {
        const ageStyles = {
          'fresh': {
            'background-color': 'var(--color-semantic-success-50)',
            'border-color': 'var(--color-semantic-success-200)',
            'color': 'var(--color-semantic-success-700)'
          },
          'stale': {
            'background-color': 'var(--color-semantic-warning-50)',
            'border-color': 'var(--color-semantic-warning-200)',
            'color': 'var(--color-semantic-warning-700)'
          },
          'old': {
            'background-color': 'var(--color-neutral-100)',
            'border-color': 'var(--color-neutral-300)',
            'color': 'var(--color-neutral-600)'
          }
        }

        return ageStyles[age]
      }],

      // Progress Bars with Enterprise Colors
      [/^progress-(\d+)$/, ([, value]) => {
        const percentage = Math.min(parseInt(value, 10), 100)
        return {
          'background': `linear-gradient(to right, var(--color-primary-500) ${percentage}%, var(--color-neutral-200) ${percentage}%)`,
          'height': '0.5rem',
          'border-radius': '0.25rem'
        }
      }]
    ],
    variants: [
      // Enterprise-specific pseudo-variants
      (matcher) => {
        if (!matcher.startsWith('enterprise:')) return matcher
        return {
          matcher: matcher.slice(11),
          selector: '[data-enterprise-mode="true"] &',
        }
      },

      // SCADA Mode variant
      (matcher) => {
        if (!matcher.startsWith('scada:')) return matcher
        return {
          matcher: matcher.slice(6),
          selector: '[data-scada-mode="true"] &',
        }
      }
    ],
    preflights: [
      {
        getCSS: () => `
          /* Enterprise Application Base Styles */
          [data-enterprise-mode="true"] {
            --enterprise-font-mono: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
            --enterprise-transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
          }

          /* SCADA System Base Styles */
          [data-scada-mode="true"] {
            --scada-bg-primary: var(--color-neutral-900);
            --scada-bg-secondary: var(--color-neutral-800);
            --scada-text-primary: var(--color-neutral-100);
            --scada-accent: var(--color-primary-400);
            font-family: var(--enterprise-font-mono);
            background-color: var(--scada-bg-primary);
            color: var(--scada-text-primary);
          }

          /* IoT Device Cards Animation */
          @keyframes device-pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.7; }
          }

          .device-online {
            animation: device-pulse 2s ease-in-out infinite;
          }

          /* Industrial Gauge Animation */
          @keyframes gauge-rotate {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
          }

          .gauge-loading {
            animation: gauge-rotate 2s linear infinite;
          }

          /* High Contrast Mode */
          @media (prefers-contrast: high) {
            :root {
              --color-neutral-50: #ffffff;
              --color-neutral-900: #000000;
              filter: contrast(1.2);
            }
          }

          /* Reduced Motion */
          @media (prefers-reduced-motion: reduce) {
            *, *::before, *::after {
              animation-duration: 0.01ms !important;
              animation-iteration-count: 1 !important;
              transition-duration: 0.01ms !important;
            }
          }
        `
      }
    ]
  }
}