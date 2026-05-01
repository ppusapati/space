/**
 * Samavaya WASM Module Loader
 * Handles lazy loading and initialization of WASM modules
 */

type WasmModule = Record<string, unknown>;

interface ModuleCache {
  module: WasmModule | null;
  loading: Promise<WasmModule> | null;
  initialized: boolean;
}

const moduleCache: Map<string, ModuleCache> = new Map();

/**
 * Available WASM modules
 */
export type WasmModuleName =
  | 'core'
  | 'tax-engine'
  | 'validation'
  | 'barcode'
  | 'ledger'
  | 'pricing'
  | 'payroll'
  | 'bom'
  | 'depreciation'
  | 'compliance'
  | 'crypto'
  | 'i18n'
  | 'offline';

/**
 * Module paths relative to pkg directory
 */
const MODULE_PATHS: Record<WasmModuleName, string> = {
  core: '../pkg/core/samavaya_core',
  'tax-engine': '../pkg/tax-engine/samavaya_tax_engine',
  validation: '../pkg/validation/samavaya_validation',
  barcode: '../pkg/barcode/samavaya_barcode',
  ledger: '../pkg/ledger/samavaya_ledger',
  pricing: '../pkg/pricing/samavaya_pricing',
  payroll: '../pkg/payroll/samavaya_payroll',
  bom: '../pkg/bom/samavaya_bom',
  depreciation: '../pkg/depreciation/samavaya_depreciation',
  compliance: '../pkg/compliance/samavaya_compliance',
  crypto: '../pkg/crypto/samavaya_crypto',
  i18n: '../pkg/i18n/samavaya_i18n',
  offline: '../pkg/offline/samavaya_offline',
};

/**
 * Check if WASM is supported in the current environment
 */
export function isWasmSupported(): boolean {
  try {
    if (typeof WebAssembly === 'object' &&
        typeof WebAssembly.instantiate === 'function') {
      const module = new WebAssembly.Module(
        Uint8Array.of(0x0, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00)
      );
      if (module instanceof WebAssembly.Module) {
        return new WebAssembly.Instance(module) instanceof WebAssembly.Instance;
      }
    }
  } catch {
    // WASM not supported
  }
  return false;
}

/**
 * Load a WASM module by name
 * @param moduleName - Name of the module to load
 * @returns Promise resolving to the loaded module
 */
export async function loadWasmModule<T = WasmModule>(
  moduleName: WasmModuleName
): Promise<T> {
  // Check cache
  let cache = moduleCache.get(moduleName);

  if (cache?.module && cache.initialized) {
    return cache.module as T;
  }

  if (cache?.loading) {
    return cache.loading as Promise<T>;
  }

  // Initialize cache entry
  cache = {
    module: null,
    loading: null,
    initialized: false,
  };
  moduleCache.set(moduleName, cache);

  // Start loading
  const modulePath = MODULE_PATHS[moduleName];
  if (!modulePath) {
    throw new Error(`Unknown WASM module: ${moduleName}`);
  }

  cache.loading = (async () => {
    try {
      // Dynamic import of the WASM module
      const wasmModule = await import(/* @vite-ignore */ modulePath);

      // Call init function if available (wasm-bindgen modules have this)
      if (typeof wasmModule.default === 'function') {
        await wasmModule.default();
      }

      cache!.module = wasmModule;
      cache!.initialized = true;
      cache!.loading = null;

      return wasmModule;
    } catch (error) {
      cache!.loading = null;
      throw new Error(
        `Failed to load WASM module '${moduleName}': ${error instanceof Error ? error.message : String(error)}`
      );
    }
  })();

  return cache.loading as Promise<T>;
}

/**
 * Preload multiple WASM modules
 * @param moduleNames - Array of module names to preload
 */
export async function preloadWasmModules(
  moduleNames: WasmModuleName[]
): Promise<void> {
  await Promise.all(moduleNames.map((name) => loadWasmModule(name)));
}

/**
 * Check if a module is loaded and initialized
 * @param moduleName - Name of the module to check
 */
export function isModuleLoaded(moduleName: WasmModuleName): boolean {
  const cache = moduleCache.get(moduleName);
  return cache?.initialized ?? false;
}

/**
 * Unload a module from cache (for memory management)
 * @param moduleName - Name of the module to unload
 */
export function unloadModule(moduleName: WasmModuleName): void {
  moduleCache.delete(moduleName);
}

/**
 * Unload all modules from cache
 */
export function unloadAllModules(): void {
  moduleCache.clear();
}

/**
 * Get list of loaded modules
 */
export function getLoadedModules(): WasmModuleName[] {
  const loaded: WasmModuleName[] = [];
  for (const [name, cache] of moduleCache.entries()) {
    if (cache.initialized) {
      loaded.push(name as WasmModuleName);
    }
  }
  return loaded;
}
