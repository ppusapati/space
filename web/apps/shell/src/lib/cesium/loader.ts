/**
 * loader.ts — dynamic-import wrapper around @cesium/engine.
 *
 * REQ-CONST-002 / TASK-P1-WEB-002 acceptance #1: the initial JS
 * bundle must NOT contain @cesium/engine. Every Cesium use site
 * routes through `loadCesium()` so esbuild + vite split the
 * import into the manual `cesium-engine` + `cesium-widgets`
 * chunks declared in vite.config.ts.
 *
 * The loader is idempotent: subsequent calls return the cached
 * module set so multiple Viewer mounts on the same page don't
 * re-fetch the chunk.
 *
 * CESIUM_BASE_URL setup
 * ---------------------
 * Cesium expects to load workers + assets (textures, fonts, the
 * 1.5MB Stk.js) from a known base URL. The chetana shell copies
 * the static assets into /cesium-assets/ at build time (see
 * vite.config.ts's viteStaticCopy) and stamps
 * `window.CESIUM_BASE_URL` once on first load.
 */

interface CesiumModules {
  Viewer: typeof import("@cesium/engine").Viewer;
  Cartesian3: typeof import("@cesium/engine").Cartesian3;
  Math: typeof import("@cesium/engine").Math;
  Ion: typeof import("@cesium/engine").Ion;
  // Add re-exports as new use sites land. Keeping this surface
  // narrow forces every Cesium API consumed by the chetana shell
  // to flow through this single dynamic-import.
}

let cached: Promise<CesiumModules> | null = null;

/**
 * loadCesium dynamically imports @cesium/engine and resolves the
 * module set the chetana shell uses. Idempotent.
 *
 * Optional `ionAccessToken` configures Cesium Ion (default
 * imagery, terrain). chetana defaults to NO Ion token + locally-
 * served imagery; pass a token only if you want Ion's hosted
 * tiles.
 */
export function loadCesium(opts?: { ionAccessToken?: string; baseUrl?: string }): Promise<CesiumModules> {
  if (cached) return cached;

  // Stamp the base URL BEFORE the import so Cesium's worker
  // bootstrap reads it on first reference.
  if (typeof window !== "undefined") {
    const w = window as unknown as { CESIUM_BASE_URL?: string };
    w.CESIUM_BASE_URL = opts?.baseUrl ?? "/cesium-assets/";
  }

  cached = import("@cesium/engine").then((mod) => {
    if (opts?.ionAccessToken) {
      mod.Ion.defaultAccessToken = opts.ionAccessToken;
    }
    return {
      Viewer: mod.Viewer,
      Cartesian3: mod.Cartesian3,
      Math: mod.Math,
      Ion: mod.Ion,
    };
  });

  return cached;
}

/**
 * preloadCesium kicks off the chunk fetch without waiting on the
 * promise. Useful for routes that know they'll mount a Cesium
 * viewer soon (e.g. hover-prefetch on a "Globe" nav link).
 */
export function preloadCesium(): void {
  void loadCesium();
}
