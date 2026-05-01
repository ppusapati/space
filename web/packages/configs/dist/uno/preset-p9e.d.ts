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
export declare function presetP9E(): {
    name: string;
    theme: import("@unocss/preset-uno").Theme;
    shortcuts: any[];
    rules: ((RegExp | (([, d]: string[]) => {
        padding: string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'padding-top': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'padding-right': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'padding-bottom': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'padding-left': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        margin: string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'margin-top': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'margin-right': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'margin-bottom': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'margin-left': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        width: string;
    }))[] | (RegExp | (([, d]: string[]) => {
        height: string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'min-width': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'min-height': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'max-width': string;
    }))[] | (RegExp | (([, d]: string[]) => {
        'max-height': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'font-size': string;
    }))[] | (RegExp | (([, h]: string[]) => {
        'line-height': number;
    }))[] | (RegExp | (([, s]: string[]) => {
        'letter-spacing': string;
    }))[] | (RegExp | (([, color, shade]: string[]) => {
        color: string;
    }))[] | (RegExp | (([, color, shade]: string[]) => {
        'background-color': string;
    }))[] | (RegExp | (([, color, shade]: string[]) => {
        'border-color': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        flex: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        order: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'flex-grow': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'flex-shrink': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-template-columns': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-template-rows': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-column': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-row': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-column-start': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-column-end': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-row-start': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        'grid-row-end': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        gap: string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'column-gap': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'row-gap': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'border-radius': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'border-top-left-radius': string;
        'border-top-right-radius': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'border-top-right-radius': string;
        'border-bottom-right-radius': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'border-bottom-left-radius': string;
        'border-bottom-right-radius': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'border-top-left-radius': string;
        'border-bottom-left-radius': string;
    }))[] | (RegExp | (([, s]: string[]) => {
        'box-shadow': "none" | "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)" | "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)" | "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)" | "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)" | "0 25px 50px -12px rgb(0 0 0 / 0.25)";
    }))[] | (RegExp | (([, n]: string[]) => {
        'z-index': string;
    }))[] | (RegExp | (([, n]: string[]) => {
        opacity: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        transform: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        top: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        right: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        bottom: string;
    }))[] | (RegExp | (([, n]: string[]) => {
        left: string;
    }))[] | (RegExp | (([, direction]: string[]) => {
        'background-image': string;
    }))[] | (RegExp | (([, color, shade]: string[]) => {
        '--un-gradient-to': string;
    }))[] | (RegExp | (([, age]: string[]) => {
        readonly 'background-color': "var(--color-semantic-success-50)";
        readonly 'border-color': "var(--color-semantic-success-200)";
        readonly color: "var(--color-semantic-success-700)";
    } | {
        readonly 'background-color': "var(--color-semantic-warning-50)";
        readonly 'border-color': "var(--color-semantic-warning-200)";
        readonly color: "var(--color-semantic-warning-700)";
    } | {
        readonly 'background-color': "var(--color-neutral-100)";
        readonly 'border-color': "var(--color-neutral-300)";
        readonly color: "var(--color-neutral-600)";
    }))[] | (RegExp | (([, type]: string[]) => {
        readonly 'background-color': "var(--color-blue-50)";
        readonly 'border-color': "var(--color-blue-200)";
        readonly color: "var(--color-blue-700)";
    } | {
        readonly 'background-color': "var(--color-semantic-success-50)";
        readonly 'border-color': "var(--color-semantic-success-200)";
        readonly color: "var(--color-semantic-success-700)";
    } | {
        readonly 'background-color': "var(--color-semantic-warning-50)";
        readonly 'border-color': "var(--color-semantic-warning-200)";
        readonly color: "var(--color-semantic-warning-700)";
    } | {
        readonly 'background-color': "var(--color-semantic-error-50)";
        readonly 'border-color': "var(--color-semantic-error-200)";
        readonly color: "var(--color-semantic-error-700)";
    }))[] | (RegExp | (([, variant]: string[]) => {
        readonly 'background-color': "var(--color-primary-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-neutral-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-semantic-success-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-semantic-error-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-semantic-warning-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-blue-600)";
        readonly color: "white";
    } | {
        readonly 'background-color': "var(--color-neutral-100)";
        readonly color: "var(--color-neutral-900)";
    } | {
        readonly 'background-color': "var(--color-neutral-800)";
        readonly color: "white";
    }))[] | (RegExp | (([, status]: string[]) => {
        readonly 'background-color': "var(--color-semantic-success-500)";
        readonly animation: "device-pulse 2s ease-in-out infinite";
    } | {
        readonly 'background-color': "var(--color-neutral-400)";
    } | {
        readonly 'background-color': "var(--color-semantic-error-500)";
        readonly animation: "device-pulse 1s ease-in-out infinite";
    } | {
        readonly 'background-color': "var(--color-semantic-warning-500)";
    }))[] | (RegExp | (([, breakpoint, utility]: string[]) => {
        [x: string]: {
            [x: string]: string;
        };
    }))[])[];
    variants: ((matcher: string) => string | {
        matcher: string;
        selector: (input: string) => string;
    })[];
    preflights: {
        getCSS: () => string;
    }[];
};
//# sourceMappingURL=preset-p9e.d.ts.map