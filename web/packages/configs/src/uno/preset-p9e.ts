import { designTokensTheme, componentShortcuts, animations } from '@p9e.in/samavaya/uno'

/**
 * P9E Enterprise UnoCSS Preset
 *
 * Comprehensive utility-first preset with complete CSS framework capabilities:
 * - Complete Bootstrap-like utilities
 * - WindCSS3 responsive system
 * - UnoCSS atomic classes
 * - Enterprise patterns (SCADA, ERP, IoT)
 * - Industrial UI components
 * - Advanced accessibility features
 * - Modern CSS features and animations
 */
export function presetP9E() {
  return {
    name: '@p9e.in/preset-p9e',
    theme: designTokensTheme,
    shortcuts: [
      ...(Array.isArray(componentShortcuts) ? componentShortcuts : []),

      // ===== BOOTSTRAP-LIKE UTILITIES =====

      // Layout & Container System
      {
        'container': 'mx-auto px-4 sm:px-6 lg:px-8',
        'container-sm': 'max-w-screen-sm mx-auto px-4',
        'container-md': 'max-w-screen-md mx-auto px-4',
        'container-lg': 'max-w-screen-lg mx-auto px-4',
        'container-xl': 'max-w-screen-xl mx-auto px-4',
        'container-2xl': 'max-w-screen-2xl mx-auto px-4',
        'container-fluid': 'w-full px-4',

        // Bootstrap Grid System
        'row': 'flex flex-wrap -mx-3',
        'col': 'flex-1 px-3',
        'col-1': 'w-1/12 px-3',
        'col-2': 'w-2/12 px-3',
        'col-3': 'w-3/12 px-3',
        'col-4': 'w-4/12 px-3',
        'col-5': 'w-5/12 px-3',
        'col-6': 'w-6/12 px-3',
        'col-7': 'w-7/12 px-3',
        'col-8': 'w-8/12 px-3',
        'col-9': 'w-9/12 px-3',
        'col-10': 'w-10/12 px-3',
        'col-11': 'w-11/12 px-3',
        'col-12': 'w-full px-3',

        // Bootstrap Cards
        'card': 'bg-white border border-neutral-200 rounded-lg shadow-sm',
        'card-dark': 'bg-neutral-800 border-neutral-700 text-white',
        'card-header': 'px-6 py-4 border-b border-neutral-200',
        'card-body': 'px-6 py-4',
        'card-footer': 'px-6 py-4 border-t border-neutral-200 bg-neutral-50',
        'card-title': 'text-lg font-semibold mb-2',
        'card-text': 'text-neutral-600 mb-4',

        // Bootstrap Buttons
        'btn': 'inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md font-medium text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed',
        'btn-sm': 'px-3 py-1.5 text-xs',
        'btn-lg': 'px-6 py-3 text-base',
        'btn-primary': 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500',
        'btn-secondary': 'bg-neutral-600 text-white hover:bg-neutral-700 focus:ring-neutral-500',
        'btn-success': 'bg-semantic-success-600 text-white hover:bg-semantic-success-700 focus:ring-semantic-success-500',
        'btn-danger': 'bg-semantic-error-600 text-white hover:bg-semantic-error-700 focus:ring-semantic-error-500',
        'btn-warning': 'bg-semantic-warning-600 text-white hover:bg-semantic-warning-700 focus:ring-semantic-warning-500',
        'btn-info': 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500',
        'btn-light': 'bg-neutral-100 text-neutral-900 hover:bg-neutral-200 focus:ring-neutral-500 border-neutral-300',
        'btn-dark': 'bg-neutral-800 text-white hover:bg-neutral-900 focus:ring-neutral-600',
        'btn-outline-primary': 'border-primary-600 text-primary-600 hover:bg-primary-600 hover:text-white',
        'btn-outline-secondary': 'border-neutral-600 text-neutral-600 hover:bg-neutral-600 hover:text-white',
        'btn-link': 'text-primary-600 hover:text-primary-700 underline bg-transparent border-0 p-0',

        // Bootstrap Forms
        'form-control': 'block w-full px-3 py-2 border border-neutral-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm',
        'form-select': 'block w-full px-3 py-2 pr-10 border border-neutral-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm',
        'form-check': 'flex items-center',
        'form-check-input': 'w-4 h-4 text-primary-600 border-neutral-300 rounded focus:ring-primary-500',
        'form-label': 'block text-sm font-medium text-neutral-700 mb-1',
        'form-text': 'text-xs text-neutral-500 mt-1',
        'input-group': 'flex relative',
        'input-group-text': 'px-3 py-2 bg-neutral-50 border border-r-0 border-neutral-300 rounded-l-md text-neutral-500 text-sm',

        // Bootstrap Alerts
        'alert': 'px-4 py-3 rounded-md border',
        'alert-primary': 'bg-primary-50 border-primary-200 text-primary-700',
        'alert-secondary': 'bg-neutral-50 border-neutral-200 text-neutral-700',
        'alert-success': 'bg-semantic-success-50 border-semantic-success-200 text-semantic-success-700',
        'alert-danger': 'bg-semantic-error-50 border-semantic-error-200 text-semantic-error-700',
        'alert-warning': 'bg-semantic-warning-50 border-semantic-warning-200 text-semantic-warning-700',
        'alert-info': 'bg-blue-50 border-blue-200 text-blue-700',
        'alert-light': 'bg-neutral-50 border-neutral-200 text-neutral-600',
        'alert-dark': 'bg-neutral-800 border-neutral-700 text-neutral-200',

        // Bootstrap Badges
        'badge': 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
        'badge-primary': 'bg-primary-100 text-primary-700',
        'badge-secondary': 'bg-neutral-100 text-neutral-700',
        'badge-success': 'bg-semantic-success-100 text-semantic-success-700',
        'badge-danger': 'bg-semantic-error-100 text-semantic-error-700',
        'badge-warning': 'bg-semantic-warning-100 text-semantic-warning-700',
        'badge-info': 'bg-blue-100 text-blue-700',
        'badge-light': 'bg-neutral-100 text-neutral-600',
        'badge-dark': 'bg-neutral-700 text-neutral-100',

        // Bootstrap Navigation
        'nav': 'flex space-x-4',
        'nav-link': 'px-3 py-2 rounded-md text-sm font-medium text-neutral-600 hover:text-neutral-900 hover:bg-neutral-100',
        'nav-link-active': 'bg-primary-100 text-primary-700',
        'navbar': 'bg-white shadow-sm border-b border-neutral-200 px-4 py-3',
        'navbar-brand': 'text-xl font-semibold text-neutral-900',
        'navbar-nav': 'flex space-x-4 ml-auto',
        'navbar-dark': 'bg-neutral-800 text-white border-neutral-700',

        // Bootstrap Tables
        'table': 'min-w-full divide-y divide-neutral-200',
        'table-striped': '[&>tbody>tr:nth-child(odd)]:bg-neutral-50',
        'table-hover': '[&>tbody>tr]:hover:bg-neutral-50',
        'table-bordered': 'border border-neutral-200 [&_td]:border [&_th]:border [&_td]:border-neutral-200 [&_th]:border-neutral-200',
        'table-sm': '[&_td]:px-3 [&_td]:py-2 [&_th]:px-3 [&_th]:py-2 text-sm',
        'table-dark': 'bg-neutral-800 text-white [&_td]:border-neutral-700 [&_th]:border-neutral-700',

        // Bootstrap List Groups
        'list-group': 'divide-y divide-neutral-200 border border-neutral-200 rounded-md',
        'list-group-item': 'px-4 py-3 hover:bg-neutral-50',
        'list-group-item-action': 'cursor-pointer hover:bg-neutral-50 focus:bg-neutral-100',
        'list-group-flush': 'border-0 rounded-none [&>.list-group-item]:border-x-0 [&>.list-group-item:first-child]:border-t-0 [&>.list-group-item:last-child]:border-b-0',

        // Bootstrap Utilities
        'd-none': 'hidden',
        'd-block': 'block',
        'd-inline': 'inline',
        'd-inline-block': 'inline-block',
        'd-flex': 'flex',
        'd-grid': 'grid',
        'text-start': 'text-left',
        'text-center': 'text-center',
        'text-end': 'text-right',
        'text-muted': 'text-neutral-500',
        'text-primary': 'text-primary-600',
        'text-secondary': 'text-neutral-600',
        'text-success': 'text-semantic-success-600',
        'text-danger': 'text-semantic-error-600',
        'text-warning': 'text-semantic-warning-600',
        'text-info': 'text-blue-600',
        'text-light': 'text-neutral-300',
        'text-dark': 'text-neutral-800',

        // Bootstrap Spacing (p-*, m-*)
        'p-0': 'p-0', 'p-1': 'p-1', 'p-2': 'p-2', 'p-3': 'p-3', 'p-4': 'p-4', 'p-5': 'p-5',
        'px-0': 'px-0', 'px-1': 'px-1', 'px-2': 'px-2', 'px-3': 'px-3', 'px-4': 'px-4', 'px-5': 'px-5',
        'py-0': 'py-0', 'py-1': 'py-1', 'py-2': 'py-2', 'py-3': 'py-3', 'py-4': 'py-4', 'py-5': 'py-5',
        'pt-0': 'pt-0', 'pt-1': 'pt-1', 'pt-2': 'pt-2', 'pt-3': 'pt-3', 'pt-4': 'pt-4', 'pt-5': 'pt-5',
        'pb-0': 'pb-0', 'pb-1': 'pb-1', 'pb-2': 'pb-2', 'pb-3': 'pb-3', 'pb-4': 'pb-4', 'pb-5': 'pb-5',
        'pl-0': 'pl-0', 'pl-1': 'pl-1', 'pl-2': 'pl-2', 'pl-3': 'pl-3', 'pl-4': 'pl-4', 'pl-5': 'pl-5',
        'pr-0': 'pr-0', 'pr-1': 'pr-1', 'pr-2': 'pr-2', 'pr-3': 'pr-3', 'pr-4': 'pr-4', 'pr-5': 'pr-5',
        'm-0': 'm-0', 'm-1': 'm-1', 'm-2': 'm-2', 'm-3': 'm-3', 'm-4': 'm-4', 'm-5': 'm-5',
        'mx-0': 'mx-0', 'mx-1': 'mx-1', 'mx-2': 'mx-2', 'mx-3': 'mx-3', 'mx-4': 'mx-4', 'mx-5': 'mx-5',
        'my-0': 'my-0', 'my-1': 'my-1', 'my-2': 'my-2', 'my-3': 'my-3', 'my-4': 'my-4', 'my-5': 'my-5',
        'mt-0': 'mt-0', 'mt-1': 'mt-1', 'mt-2': 'mt-2', 'mt-3': 'mt-3', 'mt-4': 'mt-4', 'mt-5': 'mt-5',
        'mb-0': 'mb-0', 'mb-1': 'mb-1', 'mb-2': 'mb-2', 'mb-3': 'mb-3', 'mb-4': 'mb-4', 'mb-5': 'mb-5',
        'ml-0': 'ml-0', 'ml-1': 'ml-1', 'ml-2': 'ml-2', 'ml-3': 'ml-3', 'ml-4': 'ml-4', 'ml-5': 'ml-5',
        'mr-0': 'mr-0', 'mr-1': 'mr-1', 'mr-2': 'mr-2', 'mr-3': 'mr-3', 'mr-4': 'mr-4', 'mr-5': 'mr-5',

        // ===== MODERN CSS UTILITIES =====

        // Flexbox Utilities
        'flex-row': 'flex-row',
        'flex-col': 'flex-col',
        'flex-wrap': 'flex-wrap',
        'flex-nowrap': 'flex-nowrap',
        'justify-start': 'justify-start',
        'justify-center': 'justify-center',
        'justify-end': 'justify-end',
        'justify-between': 'justify-between',
        'justify-around': 'justify-around',
        'justify-evenly': 'justify-evenly',
        'items-start': 'items-start',
        'items-center': 'items-center',
        'items-end': 'items-end',
        'items-stretch': 'items-stretch',
        'items-baseline': 'items-baseline',
        'content-start': 'content-start',
        'content-center': 'content-center',
        'content-end': 'content-end',
        'content-between': 'content-between',
        'content-around': 'content-around',
        'content-evenly': 'content-evenly',
        'self-auto': 'self-auto',
        'self-start': 'self-start',
        'self-center': 'self-center',
        'self-end': 'self-end',
        'self-stretch': 'self-stretch',

        // Grid Utilities
        'grid-cols-1': 'grid-cols-1',
        'grid-cols-2': 'grid-cols-2',
        'grid-cols-3': 'grid-cols-3',
        'grid-cols-4': 'grid-cols-4',
        'grid-cols-5': 'grid-cols-5',
        'grid-cols-6': 'grid-cols-6',
        'grid-cols-7': 'grid-cols-7',
        'grid-cols-8': 'grid-cols-8',
        'grid-cols-9': 'grid-cols-9',
        'grid-cols-10': 'grid-cols-10',
        'grid-cols-11': 'grid-cols-11',
        'grid-cols-12': 'grid-cols-12',
        'col-span-1': 'col-span-1',
        'col-span-2': 'col-span-2',
        'col-span-3': 'col-span-3',
        'col-span-4': 'col-span-4',
        'col-span-5': 'col-span-5',
        'col-span-6': 'col-span-6',
        'col-span-full': 'col-span-full',
        'gap-0': 'gap-0',
        'gap-1': 'gap-1',
        'gap-2': 'gap-2',
        'gap-3': 'gap-3',
        'gap-4': 'gap-4',
        'gap-5': 'gap-5',
        'gap-6': 'gap-6',
        'gap-8': 'gap-8',
        'gap-10': 'gap-10',
        'gap-12': 'gap-12',

        // Position Utilities
        'position-static': 'static',
        'position-relative': 'relative',
        'position-absolute': 'absolute',
        'position-fixed': 'fixed',
        'position-sticky': 'sticky',
        'top-0': 'top-0',
        'right-0': 'right-0',
        'bottom-0': 'bottom-0',
        'left-0': 'left-0',
        'inset-0': 'inset-0',

        // Size Utilities
        'w-auto': 'w-auto',
        'w-full': 'w-full',
        'w-screen': 'w-screen',
        'h-auto': 'h-auto',
        'h-full': 'h-full',
        'h-screen': 'h-screen',
        'min-w-0': 'min-w-0',
        'min-w-full': 'min-w-full',
        'min-h-0': 'min-h-0',
        'min-h-full': 'min-h-full',
        'max-w-none': 'max-w-none',
        'max-w-full': 'max-w-full',
        'max-h-none': 'max-h-none',
        'max-h-full': 'max-h-full',

        // Typography
        'text-xs': 'text-xs',
        'text-sm': 'text-sm',
        'text-base': 'text-base',
        'text-lg': 'text-lg',
        'text-xl': 'text-xl',
        'text-2xl': 'text-2xl',
        'text-3xl': 'text-3xl',
        'text-4xl': 'text-4xl',
        'text-5xl': 'text-5xl',
        'font-thin': 'font-thin',
        'font-light': 'font-light',
        'font-normal': 'font-normal',
        'font-medium': 'font-medium',
        'font-semibold': 'font-semibold',
        'font-bold': 'font-bold',
        'font-extrabold': 'font-extrabold',
        'font-black': 'font-black',
        'italic': 'italic',
        'not-italic': 'not-italic',
        'uppercase': 'uppercase',
        'lowercase': 'lowercase',
        'capitalize': 'capitalize',
        'normal-case': 'normal-case',
        'underline': 'underline',
        'no-underline': 'no-underline',
        'line-through': 'line-through',

        // Border & Radius
        'border': 'border',
        'border-0': 'border-0',
        'border-t': 'border-t',
        'border-r': 'border-r',
        'border-b': 'border-b',
        'border-l': 'border-l',
        'border-solid': 'border-solid',
        'border-dashed': 'border-dashed',
        'border-dotted': 'border-dotted',
        'rounded': 'rounded',
        'rounded-none': 'rounded-none',
        'rounded-sm': 'rounded-sm',
        'rounded-md': 'rounded-md',
        'rounded-lg': 'rounded-lg',
        'rounded-xl': 'rounded-xl',
        'rounded-2xl': 'rounded-2xl',
        'rounded-3xl': 'rounded-3xl',
        'rounded-full': 'rounded-full',

        // Shadow & Effects
        'shadow': 'shadow',
        'shadow-sm': 'shadow-sm',
        'shadow-md': 'shadow-md',
        'shadow-lg': 'shadow-lg',
        'shadow-xl': 'shadow-xl',
        'shadow-2xl': 'shadow-2xl',
        'shadow-none': 'shadow-none',
        'opacity-0': 'opacity-0',
        'opacity-25': 'opacity-25',
        'opacity-50': 'opacity-50',
        'opacity-75': 'opacity-75',
        'opacity-100': 'opacity-100',

        // Transitions & Animations
        'transition': 'transition',
        'transition-none': 'transition-none',
        'transition-all': 'transition-all',
        'transition-colors': 'transition-colors',
        'transition-opacity': 'transition-opacity',
        'transition-shadow': 'transition-shadow',
        'transition-transform': 'transition-transform',
        'duration-75': 'duration-75',
        'duration-100': 'duration-100',
        'duration-150': 'duration-150',
        'duration-200': 'duration-200',
        'duration-300': 'duration-300',
        'duration-500': 'duration-500',
        'duration-700': 'duration-700',
        'duration-1000': 'duration-1000',
        'ease-linear': 'ease-linear',
        'ease-in': 'ease-in',
        'ease-out': 'ease-out',
        'ease-in-out': 'ease-in-out',

        // Transform
        'transform': 'transform',
        'transform-gpu': 'transform-gpu',
        'transform-none': 'transform-none',
        'scale-0': 'scale-0',
        'scale-50': 'scale-50',
        'scale-75': 'scale-75',
        'scale-90': 'scale-90',
        'scale-95': 'scale-95',
        'scale-100': 'scale-100',
        'scale-105': 'scale-105',
        'scale-110': 'scale-110',
        'scale-125': 'scale-125',
        'scale-150': 'scale-150',
        'rotate-0': 'rotate-0',
        'rotate-1': 'rotate-1',
        'rotate-2': 'rotate-2',
        'rotate-3': 'rotate-3',
        'rotate-6': 'rotate-6',
        'rotate-12': 'rotate-12',
        'rotate-45': 'rotate-45',
        'rotate-90': 'rotate-90',
        'rotate-180': 'rotate-180',

        // ===== ENTERPRISE PATTERNS =====

        // Enterprise Layout Patterns
        'enterprise-container': 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8',
        'enterprise-card': 'bg-white dark:bg-neutral-800 shadow-sm border border-neutral-200 dark:border-neutral-700 rounded-lg',
        'enterprise-header': 'bg-white dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700 px-6 py-4',
        'enterprise-sidebar': 'bg-neutral-50 dark:bg-neutral-900 border-r border-neutral-200 dark:border-neutral-700 w-64 h-full',
        'enterprise-main': 'flex-1 overflow-auto p-6',
        'enterprise-layout': 'min-h-screen bg-neutral-50 dark:bg-neutral-900',

        // Data Display Patterns
        'data-grid': 'grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
        'metrics-grid': 'grid gap-6 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
        'dashboard-layout': 'grid gap-6 grid-cols-1 lg:grid-cols-3',
        'analytics-panel': 'bg-white dark:bg-neutral-800 p-6 rounded-lg shadow-sm border border-neutral-200 dark:border-neutral-700',
        'kpi-card': 'bg-gradient-to-r from-primary-500 to-primary-600 text-white p-6 rounded-lg shadow-lg',
        'stat-card': 'text-center p-4 bg-white dark:bg-neutral-800 rounded-lg border border-neutral-200 dark:border-neutral-700',

        // SCADA Status Patterns
        'status-indicator': 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
        'status-online': 'bg-semantic-success-50 text-semantic-success-700 border border-semantic-success-200',
        'status-offline': 'bg-neutral-100 text-neutral-600 border border-neutral-300',
        'status-warning': 'bg-semantic-warning-50 text-semantic-warning-700 border border-semantic-warning-200',
        'status-critical': 'bg-semantic-error-50 text-semantic-error-700 border border-semantic-error-200',
        'status-maintenance': 'bg-blue-50 text-blue-700 border border-blue-200',
        'scada-panel': 'bg-neutral-900 text-neutral-100 p-6 rounded-lg font-mono',
        'monitoring-display': 'bg-black text-green-400 p-4 rounded border-2 border-green-500 font-mono',

        // ERP Interface Patterns
        'erp-toolbar': 'flex items-center justify-between px-4 py-2 bg-neutral-50 dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700',
        'erp-form-section': 'bg-white dark:bg-neutral-800 p-6 rounded-lg border border-neutral-200 dark:border-neutral-700 space-y-4',
        'erp-field-group': 'grid grid-cols-1 md:grid-cols-2 gap-4',
        'erp-table': 'min-w-full divide-y divide-neutral-200 dark:divide-neutral-700',
        'erp-table-header': 'bg-neutral-50 dark:bg-neutral-900',
        'erp-table-cell': 'px-6 py-4 whitespace-nowrap text-sm text-neutral-900 dark:text-neutral-100',
        'erp-sidebar': 'w-64 bg-white dark:bg-neutral-800 border-r border-neutral-200 dark:border-neutral-700 h-full',
        'workflow-step': 'flex items-center justify-center w-8 h-8 rounded-full bg-primary-600 text-white text-sm font-semibold',
        'process-flow': 'flex items-center space-x-4 p-4 bg-neutral-50 dark:bg-neutral-800 rounded-lg',

        // IoT Device Patterns
        'device-card': 'bg-white dark:bg-neutral-800 p-4 rounded-lg border border-neutral-200 dark:border-neutral-700 hover:shadow-md transition-shadow',
        'device-header': 'flex items-center justify-between mb-4',
        'device-metric': 'flex flex-col items-center p-4 bg-neutral-50 dark:bg-neutral-900 rounded-lg',
        'device-reading': 'text-2xl font-bold text-neutral-900 dark:text-neutral-100',
        'device-label': 'text-sm text-neutral-600 dark:text-neutral-400 mt-1',
        'sensor-grid': 'grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
        'iot-dashboard': 'grid gap-6 grid-cols-1 lg:grid-cols-3 xl:grid-cols-4',
        'telemetry-panel': 'bg-neutral-800 text-neutral-100 p-4 rounded-lg font-mono text-sm',

        // Industrial UI Patterns
        'control-panel': 'bg-neutral-900 dark:bg-black p-6 rounded-lg border-2 border-neutral-700',
        'control-button': 'flex items-center justify-center w-12 h-12 rounded-lg border-2 font-semibold text-sm transition-all',
        'emergency-stop': 'bg-semantic-error-600 hover:bg-semantic-error-700 border-semantic-error-700 text-white',
        'start-button': 'bg-semantic-success-600 hover:bg-semantic-success-700 border-semantic-success-700 text-white',
        'stop-button': 'bg-semantic-warning-600 hover:bg-semantic-warning-700 border-semantic-warning-700 text-white',
        'industrial-gauge': 'relative w-32 h-32 rounded-full border-8 border-neutral-300',
        'machine-status': 'inline-flex items-center px-3 py-1 rounded-full text-sm font-medium',
        'safety-alert': 'bg-red-100 border-l-4 border-red-500 text-red-700 p-4',

        // High Contrast & Accessibility
        'high-contrast': 'contrast-125 saturate-150',
        'focus-enterprise': 'focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:outline-none',
        'accessible-focus': 'focus:ring-4 focus:ring-primary-300 focus:ring-offset-2 focus:outline-none',
        'screen-reader': 'sr-only',
        'skip-link': 'sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 bg-white p-2 rounded border-2 border-primary-500 z-50',

        // Data Visualization
        'chart-container': 'bg-white dark:bg-neutral-800 p-4 rounded-lg border border-neutral-200 dark:border-neutral-700',
        'legend-item': 'flex items-center space-x-2 text-sm text-neutral-600 dark:text-neutral-400',
        'chart-title': 'text-lg font-semibold mb-4 text-center',
        'chart-axis': 'text-xs text-neutral-500',
        'data-point': 'w-3 h-3 rounded-full',
        'trend-up': 'text-semantic-success-600',
        'trend-down': 'text-semantic-error-600',
        'trend-flat': 'text-neutral-500',

        // Advanced Layout Patterns
        'dashboard-grid': 'grid gap-6 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5',
        'sidebar-layout': 'flex min-h-screen',
        'content-area': 'flex-1 flex flex-col overflow-hidden',
        'header-area': 'bg-white dark:bg-neutral-800 shadow-sm border-b border-neutral-200 dark:border-neutral-700',
        'main-content': 'flex-1 overflow-y-auto p-6',
        'floating-panel': 'fixed top-4 right-4 bg-white dark:bg-neutral-800 p-4 rounded-lg shadow-lg border border-neutral-200 dark:border-neutral-700 z-50',

        // Modern Component Patterns
        'glass-effect': 'backdrop-blur-sm bg-white/75 dark:bg-neutral-800/75',
        'neumorphism': 'bg-neutral-100 shadow-[8px_8px_16px_#d1d9e6,-8px_-8px_16px_#ffffff]',
        'gradient-primary': 'bg-gradient-to-r from-primary-500 to-primary-600',
        'gradient-success': 'bg-gradient-to-r from-semantic-success-500 to-semantic-success-600',
        'gradient-warning': 'bg-gradient-to-r from-semantic-warning-500 to-semantic-warning-600',
        'gradient-danger': 'bg-gradient-to-r from-semantic-error-500 to-semantic-error-600',
      }
    ],
    rules: [
      // ===== BOOTSTRAP & WINDCSS3 DYNAMIC UTILITIES =====

      // Dynamic Spacing (Bootstrap-like p-*, m-*)
      [/^p-(\d+)$/, ([, d]: string[]) => ({ padding: `${parseInt(d) * 0.25}rem` })],
      [/^px-(\d+)$/, ([, d]: string[]) => ({ 'padding-left': `${parseInt(d) * 0.25}rem`, 'padding-right': `${parseInt(d) * 0.25}rem` })],
      [/^py-(\d+)$/, ([, d]: string[]) => ({ 'padding-top': `${parseInt(d) * 0.25}rem`, 'padding-bottom': `${parseInt(d) * 0.25}rem` })],
      [/^pt-(\d+)$/, ([, d]: string[]) => ({ 'padding-top': `${parseInt(d) * 0.25}rem` })],
      [/^pr-(\d+)$/, ([, d]: string[]) => ({ 'padding-right': `${parseInt(d) * 0.25}rem` })],
      [/^pb-(\d+)$/, ([, d]: string[]) => ({ 'padding-bottom': `${parseInt(d) * 0.25}rem` })],
      [/^pl-(\d+)$/, ([, d]: string[]) => ({ 'padding-left': `${parseInt(d) * 0.25}rem` })],
      [/^m-(\d+)$/, ([, d]: string[]) => ({ margin: `${parseInt(d) * 0.25}rem` })],
      [/^mx-(\d+)$/, ([, d]: string[]) => ({ 'margin-left': `${parseInt(d) * 0.25}rem`, 'margin-right': `${parseInt(d) * 0.25}rem` })],
      [/^my-(\d+)$/, ([, d]: string[]) => ({ 'margin-top': `${parseInt(d) * 0.25}rem`, 'margin-bottom': `${parseInt(d) * 0.25}rem` })],
      [/^mt-(\d+)$/, ([, d]: string[]) => ({ 'margin-top': `${parseInt(d) * 0.25}rem` })],
      [/^mr-(\d+)$/, ([, d]: string[]) => ({ 'margin-right': `${parseInt(d) * 0.25}rem` })],
      [/^mb-(\d+)$/, ([, d]: string[]) => ({ 'margin-bottom': `${parseInt(d) * 0.25}rem` })],
      [/^ml-(\d+)$/, ([, d]: string[]) => ({ 'margin-left': `${parseInt(d) * 0.25}rem` })],

      // Dynamic Sizing
      [/^w-(\d+)$/, ([, d]: string[]) => ({ width: `${parseInt(d) * 0.25}rem` })],
      [/^h-(\d+)$/, ([, d]: string[]) => ({ height: `${parseInt(d) * 0.25}rem` })],
      [/^w-(\d+)\/(\d+)$/, ([, n, d]: string[]) => ({ width: `${(parseInt(n) / parseInt(d)) * 100}%` })],
      [/^h-(\d+)\/(\d+)$/, ([, n, d]: string[]) => ({ height: `${(parseInt(n) / parseInt(d)) * 100}%` })],
      [/^min-w-(\d+)$/, ([, d]: string[]) => ({ 'min-width': `${parseInt(d) * 0.25}rem` })],
      [/^min-h-(\d+)$/, ([, d]: string[]) => ({ 'min-height': `${parseInt(d) * 0.25}rem` })],
      [/^max-w-(\d+)$/, ([, d]: string[]) => ({ 'max-width': `${parseInt(d) * 0.25}rem` })],
      [/^max-h-(\d+)$/, ([, d]: string[]) => ({ 'max-height': `${parseInt(d) * 0.25}rem` })],

      // Dynamic Typography
      [/^text-(\d+)px$/, ([, s]: string[]) => ({ 'font-size': `${s}px` })],
      [/^text-(\d+)$/, ([, s]: string[]) => {
        const sizes = ['xs', 'sm', 'base', 'lg', 'xl', '2xl', '3xl', '4xl', '5xl', '6xl', '7xl', '8xl', '9xl']
        const size = sizes[parseInt(s)] || 'base'
        return { 'font-size': `var(--text-${size})` }
      }],
      [/^leading-(\d+)$/, ([, h]: string[]) => ({ 'line-height': parseInt(h) * 0.25 })],
      [/^tracking-(\d+)$/, ([, s]: string[]) => ({ 'letter-spacing': `${parseInt(s) * 0.025}em` })],

      // Dynamic Colors with Shades
      [/^text-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ color: `var(--color-${color}-${shade})` })],
      [/^bg-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ 'background-color': `var(--color-${color}-${shade})` })],
      [/^border-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ 'border-color': `var(--color-${color}-${shade})` })],

      // Dynamic Flexbox
      [/^flex-(\d+)$/, ([, n]: string[]) => ({ flex: `${n} ${n} 0%` })],
      [/^order-(\d+)$/, ([, n]: string[]) => ({ order: n })],
      [/^grow-(\d+)$/, ([, n]: string[]) => ({ 'flex-grow': n })],
      [/^shrink-(\d+)$/, ([, n]: string[]) => ({ 'flex-shrink': n })],

      // Dynamic Grid
      [/^grid-cols-(\d+)$/, ([, n]: string[]) => ({ 'grid-template-columns': `repeat(${n}, minmax(0, 1fr))` })],
      [/^grid-rows-(\d+)$/, ([, n]: string[]) => ({ 'grid-template-rows': `repeat(${n}, minmax(0, 1fr))` })],
      [/^col-span-(\d+)$/, ([, n]: string[]) => ({ 'grid-column': `span ${n} / span ${n}` })],
      [/^row-span-(\d+)$/, ([, n]: string[]) => ({ 'grid-row': `span ${n} / span ${n}` })],
      [/^col-start-(\d+)$/, ([, n]: string[]) => ({ 'grid-column-start': n })],
      [/^col-end-(\d+)$/, ([, n]: string[]) => ({ 'grid-column-end': n })],
      [/^row-start-(\d+)$/, ([, n]: string[]) => ({ 'grid-row-start': n })],
      [/^row-end-(\d+)$/, ([, n]: string[]) => ({ 'grid-row-end': n })],

      // Dynamic Gap
      [/^gap-(\d+)$/, ([, s]: string[]) => ({ gap: `${parseInt(s) * 0.25}rem` })],
      [/^gap-x-(\d+)$/, ([, s]: string[]) => ({ 'column-gap': `${parseInt(s) * 0.25}rem` })],
      [/^gap-y-(\d+)$/, ([, s]: string[]) => ({ 'row-gap': `${parseInt(s) * 0.25}rem` })],

      // Dynamic Border Radius
      [/^rounded-(\d+)$/, ([, s]: string[]) => ({ 'border-radius': `${parseInt(s) * 0.125}rem` })],
      [/^rounded-t-(\d+)$/, ([, s]: string[]) => ({ 'border-top-left-radius': `${parseInt(s) * 0.125}rem`, 'border-top-right-radius': `${parseInt(s) * 0.125}rem` })],
      [/^rounded-r-(\d+)$/, ([, s]: string[]) => ({ 'border-top-right-radius': `${parseInt(s) * 0.125}rem`, 'border-bottom-right-radius': `${parseInt(s) * 0.125}rem` })],
      [/^rounded-b-(\d+)$/, ([, s]: string[]) => ({ 'border-bottom-left-radius': `${parseInt(s) * 0.125}rem`, 'border-bottom-right-radius': `${parseInt(s) * 0.125}rem` })],
      [/^rounded-l-(\d+)$/, ([, s]: string[]) => ({ 'border-top-left-radius': `${parseInt(s) * 0.125}rem`, 'border-bottom-left-radius': `${parseInt(s) * 0.125}rem` })],

      // Dynamic Shadows
      [/^shadow-(\d+)$/, ([, s]: string[]) => {
        const shadows = {
          '0': 'none',
          '1': '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
          '2': '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
          '3': '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
          '4': '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
          '5': '0 25px 50px -12px rgb(0 0 0 / 0.25)'
        } as const
        return { 'box-shadow': shadows[s as keyof typeof shadows] || shadows['1'] }
      }],

      // Dynamic Z-Index
      [/^z-(\d+)$/, ([, n]: string[]) => ({ 'z-index': n })],

      // Dynamic Opacity
      [/^opacity-(\d+)$/, ([, n]: string[]) => ({ opacity: (parseInt(n) / 100).toString() })],

      // Dynamic Transforms
      [/^scale-(\d+)$/, ([, n]: string[]) => ({ transform: `scale(${parseInt(n) / 100})` })],
      [/^rotate-(\d+)$/, ([, n]: string[]) => ({ transform: `rotate(${n}deg)` })],
      [/^translate-x-(\d+)$/, ([, n]: string[]) => ({ transform: `translateX(${parseInt(n) * 0.25}rem)` })],
      [/^translate-y-(\d+)$/, ([, n]: string[]) => ({ transform: `translateY(${parseInt(n) * 0.25}rem)` })],
      [/^-translate-x-(\d+)$/, ([, n]: string[]) => ({ transform: `translateX(-${parseInt(n) * 0.25}rem)` })],
      [/^-translate-y-(\d+)$/, ([, n]: string[]) => ({ transform: `translateY(-${parseInt(n) * 0.25}rem)` })],
      [/^skew-x-(\d+)$/, ([, n]: string[]) => ({ transform: `skewX(${n}deg)` })],
      [/^skew-y-(\d+)$/, ([, n]: string[]) => ({ transform: `skewY(${n}deg)` })],

      // Dynamic Positions
      [/^top-(\d+)$/, ([, n]: string[]) => ({ top: `${parseInt(n) * 0.25}rem` })],
      [/^right-(\d+)$/, ([, n]: string[]) => ({ right: `${parseInt(n) * 0.25}rem` })],
      [/^bottom-(\d+)$/, ([, n]: string[]) => ({ bottom: `${parseInt(n) * 0.25}rem` })],
      [/^left-(\d+)$/, ([, n]: string[]) => ({ left: `${parseInt(n) * 0.25}rem` })],
      [/^inset-(\d+)$/, ([, n]: string[]) => ({ top: `${parseInt(n) * 0.25}rem`, right: `${parseInt(n) * 0.25}rem`, bottom: `${parseInt(n) * 0.25}rem`, left: `${parseInt(n) * 0.25}rem` })],

      // Dynamic Gradients
      [/^bg-gradient-to-(\w+)$/, ([, direction]: string[]) => {
        const directions = {
          't': 'to top',
          'tr': 'to top right',
          'r': 'to right',
          'br': 'to bottom right',
          'b': 'to bottom',
          'bl': 'to bottom left',
          'l': 'to left',
          'tl': 'to top left'
        } as const
        const dir = directions[direction as keyof typeof directions] || 'to right'
        return { 'background-image': `linear-gradient(${dir}, var(--un-gradient-stops))` }
      }],
      [/^from-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ '--un-gradient-from': `var(--color-${color}-${shade})`, '--un-gradient-to': 'rgb(255 255 255 / 0)', '--un-gradient-stops': 'var(--un-gradient-from), var(--un-gradient-to)' })],
      [/^via-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ '--un-gradient-to': 'rgb(255 255 255 / 0)', '--un-gradient-stops': `var(--un-gradient-from), var(--color-${color}-${shade}), var(--un-gradient-to)` })],
      [/^to-(\w+)-(\d+)$/, ([, color, shade]: string[]) => ({ '--un-gradient-to': `var(--color-${color}-${shade})` })],

      // ===== ENTERPRISE DYNAMIC UTILITIES =====

      // SCADA System Rules
      [/^scada-status-(.+)$/, ([, status]: string[]) => {
        const statusColors = {
          'running': 'var(--color-semantic-success-500)',
          'stopped': 'var(--color-neutral-500)',
          'error': 'var(--color-semantic-error-500)',
          'warning': 'var(--color-semantic-warning-500)',
          'maintenance': 'var(--color-blue-500)',
          'offline': 'var(--color-neutral-400)'
        } as const

        const color = statusColors[status as keyof typeof statusColors] || statusColors['stopped']
        return {
          'background-color': color,
          'box-shadow': `0 0 8px ${color}40`,
        }
      }],

      // IoT Device Signal Strength
      [/^signal-strength-([1-5])$/, ([, strength]: string[]) => {
        const bars = parseInt(strength, 10)
        const opacity = Math.min(bars / 5, 1)
        return {
          'background': `linear-gradient(to right, var(--color-semantic-success-500) ${bars * 20}%, var(--color-neutral-300) ${bars * 20}%)`,
          'opacity': opacity.toString()
        }
      }],

      // Industrial Gauge Styles
      [/^gauge-value-(\d+)$/, ([, value]: string[]) => {
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
      [/^priority-(low|medium|high|critical)$/, ([, priority]: string[]) => {
        const priorityColors = {
          'low': 'var(--color-blue-500)',
          'medium': 'var(--color-semantic-warning-500)',
          'high': 'var(--color-orange-500)',
          'critical': 'var(--color-semantic-error-500)'
        } as const

        const color = priorityColors[priority as keyof typeof priorityColors]
        return {
          'border-left': `4px solid ${color}`,
          'background-color': `${color}10`
        }
      }],

      // Temperature Indicators
      [/^temp-([0-9]+)$/, ([, temp]: string[]) => {
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
      [/^data-age-(fresh|stale|old)$/, ([, age]: string[]) => {
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
        } as const

        return ageStyles[age as keyof typeof ageStyles]
      }],

      // Progress Bars with Enterprise Colors
      [/^progress-(\d+)$/, ([, value]: string[]) => {
        const percentage = Math.min(parseInt(value, 10), 100)
        return {
          'background': `linear-gradient(to right, var(--color-primary-500) ${percentage}%, var(--color-neutral-200) ${percentage}%)`,
          'height': '0.5rem',
          'border-radius': '0.25rem'
        }
      }],

      // Advanced Enterprise Features
      [/^alert-(\w+)$/, ([, type]: string[]) => {
        const alertStyles = {
          'info': { 'background-color': 'var(--color-blue-50)', 'border-color': 'var(--color-blue-200)', 'color': 'var(--color-blue-700)' },
          'success': { 'background-color': 'var(--color-semantic-success-50)', 'border-color': 'var(--color-semantic-success-200)', 'color': 'var(--color-semantic-success-700)' },
          'warning': { 'background-color': 'var(--color-semantic-warning-50)', 'border-color': 'var(--color-semantic-warning-200)', 'color': 'var(--color-semantic-warning-700)' },
          'error': { 'background-color': 'var(--color-semantic-error-50)', 'border-color': 'var(--color-semantic-error-200)', 'color': 'var(--color-semantic-error-700)' }
        } as const
        return alertStyles[type as keyof typeof alertStyles] || alertStyles['info']
      }],

      // Dynamic Button Variants
      [/^btn-(\w+)$/, ([, variant]: string[]) => {
        const btnStyles = {
          'primary': { 'background-color': 'var(--color-primary-600)', 'color': 'white' },
          'secondary': { 'background-color': 'var(--color-neutral-600)', 'color': 'white' },
          'success': { 'background-color': 'var(--color-semantic-success-600)', 'color': 'white' },
          'danger': { 'background-color': 'var(--color-semantic-error-600)', 'color': 'white' },
          'warning': { 'background-color': 'var(--color-semantic-warning-600)', 'color': 'white' },
          'info': { 'background-color': 'var(--color-blue-600)', 'color': 'white' },
          'light': { 'background-color': 'var(--color-neutral-100)', 'color': 'var(--color-neutral-900)' },
          'dark': { 'background-color': 'var(--color-neutral-800)', 'color': 'white' }
        } as const
        return btnStyles[variant as keyof typeof btnStyles] || btnStyles['primary']
      }],

      // Device Status with Animation
      [/^device-status-(\w+)$/, ([, status]: string[]) => {
        const deviceStyles = {
          'online': { 'background-color': 'var(--color-semantic-success-500)', 'animation': 'device-pulse 2s ease-in-out infinite' },
          'offline': { 'background-color': 'var(--color-neutral-400)' },
          'error': { 'background-color': 'var(--color-semantic-error-500)', 'animation': 'device-pulse 1s ease-in-out infinite' },
          'maintenance': { 'background-color': 'var(--color-semantic-warning-500)' }
        } as const
        return deviceStyles[status as keyof typeof deviceStyles] || deviceStyles['offline']
      }],

      // Responsive Breakpoints
      [/^(sm|md|lg|xl|2xl):(.+)$/, ([, breakpoint, utility]: string[]) => {
        const breakpoints = {
          'sm': '640px',
          'md': '768px',
          'lg': '1024px',
          'xl': '1280px',
          '2xl': '1536px'
        } as const
        return {
          [`@media (min-width: ${breakpoints[breakpoint as keyof typeof breakpoints]})`]: {
            [utility]: utility
          }
        }
      }]
    ],
    variants: [
      // Enterprise-specific pseudo-variants
      (matcher: string) => {
        if (!matcher.startsWith('enterprise:')) return matcher
        return {
          matcher: matcher.slice(11),
          selector: (input: string) => `[data-enterprise-mode="true"] ${input}`,
        }
      },

      // SCADA Mode variant
      (matcher: string) => {
        if (!matcher.startsWith('scada:')) return matcher
        return {
          matcher: matcher.slice(6),
          selector: (input: string) => `[data-scada-mode="true"] ${input}`,
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

          /* ===== PRINT STYLES ===== */
          @media print {
            /* Reset and Base */
            *,
            *::before,
            *::after {
              background: transparent !important;
              color: #000 !important;
              box-shadow: none !important;
              text-shadow: none !important;
            }

            /* Page Setup */
            @page {
              margin: 2cm;
              size: A4;
            }

            @page :first {
              margin-top: 3cm;
            }

            html, body {
              width: 100%;
              height: auto;
              font-size: 12pt;
              line-height: 1.5;
              background: #fff !important;
            }

            /* Typography */
            h1, h2, h3, h4, h5, h6 {
              page-break-after: avoid;
              page-break-inside: avoid;
            }

            h1 { font-size: 24pt; }
            h2 { font-size: 20pt; }
            h3 { font-size: 16pt; }
            h4 { font-size: 14pt; }

            p, li, blockquote {
              orphans: 3;
              widows: 3;
            }

            /* Links */
            a, a:visited {
              text-decoration: underline;
              color: #000 !important;
            }

            a[href^="http"]::after {
              content: " (" attr(href) ")";
              font-size: 10pt;
              font-weight: normal;
            }

            a[href^="#"]::after,
            a[href^="javascript:"]::after {
              content: "";
            }

            /* Images */
            img {
              max-width: 100% !important;
              page-break-inside: avoid;
            }

            /* Tables */
            table {
              border-collapse: collapse !important;
              width: 100%;
            }

            table, th, td {
              border: 1px solid #000 !important;
            }

            th, td {
              padding: 8pt;
              text-align: left;
            }

            thead {
              display: table-header-group;
            }

            tr {
              page-break-inside: avoid;
            }

            /* Code Blocks */
            pre, code {
              font-family: 'Courier New', monospace;
              font-size: 10pt;
              border: 1px solid #ccc;
              padding: 4pt;
            }

            pre {
              white-space: pre-wrap;
              word-wrap: break-word;
              page-break-inside: avoid;
            }

            /* Hide UI Elements */
            .no-print,
            .print-hide,
            nav,
            aside,
            .sidebar,
            .navigation,
            .navbar,
            .footer,
            .header-nav,
            button:not(.print-show),
            .btn:not(.print-show),
            .modal,
            .drawer,
            .toast,
            .notification,
            .tooltip,
            .popover,
            .dropdown-menu,
            .command-palette,
            input[type="search"],
            .search-bar,
            .filters,
            .pagination,
            .breadcrumb,
            video,
            audio,
            iframe,
            .chart-interactive,
            .scroll-to-top {
              display: none !important;
            }

            /* Show Print Elements */
            .print-only,
            .print-show {
              display: block !important;
            }

            /* Page Breaks */
            .page-break-before {
              page-break-before: always;
            }

            .page-break-after {
              page-break-after: always;
            }

            .avoid-break {
              page-break-inside: avoid;
            }

            /* ERP Specific Print Styles */
            .invoice-header,
            .report-header,
            .document-header {
              border-bottom: 2pt solid #000;
              padding-bottom: 12pt;
              margin-bottom: 24pt;
            }

            .invoice-footer,
            .report-footer,
            .document-footer {
              border-top: 1pt solid #000;
              padding-top: 12pt;
              margin-top: 24pt;
              position: fixed;
              bottom: 0;
              left: 0;
              right: 0;
            }

            .company-logo {
              max-height: 60pt;
              max-width: 200pt;
            }

            .document-title {
              font-size: 20pt;
              font-weight: bold;
              text-align: center;
              margin: 24pt 0;
            }

            .signature-line {
              border-top: 1pt solid #000;
              width: 200pt;
              margin-top: 48pt;
              padding-top: 4pt;
            }

            .stamp-area {
              border: 1pt dashed #666;
              width: 120pt;
              height: 120pt;
              text-align: center;
              padding-top: 48pt;
              color: #666 !important;
            }

            /* Financial Tables */
            .financial-table {
              font-family: 'Courier New', monospace;
            }

            .financial-table td:last-child,
            .financial-table th:last-child {
              text-align: right;
            }

            .total-row {
              font-weight: bold;
              border-top: 2pt solid #000 !important;
            }

            .subtotal-row {
              border-top: 1pt solid #000 !important;
            }

            /* Watermark */
            .print-watermark {
              position: fixed;
              top: 50%;
              left: 50%;
              transform: translate(-50%, -50%) rotate(-45deg);
              font-size: 72pt;
              color: #eee !important;
              z-index: -1;
              opacity: 0.3;
            }

            /* QR Code and Barcodes */
            .qr-code,
            .barcode {
              max-width: 100pt;
              page-break-inside: avoid;
            }
          }

          /* Print Utilities */
          .print-color-adjust {
            print-color-adjust: exact;
            -webkit-print-color-adjust: exact;
          }

          .print-bg {
            print-color-adjust: exact;
            -webkit-print-color-adjust: exact;
          }
        `
      }
    ]
  }
}