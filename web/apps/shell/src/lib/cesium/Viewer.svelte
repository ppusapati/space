<!--
  Viewer.svelte — base Cesium viewer Svelte component.

  Used by Phase 2 (ground tracks, sky plot, AOS/LOS timeline) and
  Phase 4 visualisations. This component is the SINGLE entry
  point Cesium chunks load through; mounting it triggers the
  dynamic import declared in loader.ts. Unmounting destroys the
  Cesium viewer to release WebGL context + worker handles.

  Usage:
    <CesiumViewer
      onReady={(viewer) => { viewer.camera.flyTo({...}); }}
      ionAccessToken={import.meta.env.VITE_CESIUM_ION_TOKEN}
    />
-->
<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { loadCesium } from "./loader";

  interface Props {
    /** Optional Cesium Ion token for hosted tiles. */
    ionAccessToken?: string;
    /** Base URL for Cesium static assets. Defaults to /cesium-assets/. */
    baseUrl?: string;
    /** Called once the Viewer instance is ready. The host can
     *  drive the camera, add entities, etc. */
    onReady?: (viewer: import("@cesium/engine").Viewer) => void;
    /** Called when the viewer is destroyed (component unmount). */
    onDestroyed?: () => void;
  }

  let { ionAccessToken, baseUrl, onReady, onDestroyed }: Props = $props();

  let container: HTMLDivElement | undefined = $state();
  let viewer: import("@cesium/engine").Viewer | null = null;
  let loadError = $state<string | null>(null);

  onMount(async () => {
    if (!container) return;
    try {
      const cesium = await loadCesium({ ionAccessToken, baseUrl });
      viewer = new cesium.Viewer(container);
      onReady?.(viewer);
    } catch (err) {
      loadError = (err as Error).message ?? "Failed to load Cesium.";
    }
  });

  onDestroy(() => {
    if (viewer) {
      try {
        viewer.destroy();
      } catch {
        // best-effort
      }
      viewer = null;
    }
    onDestroyed?.();
  });
</script>

<div bind:this={container} class="cesium-viewer w-full h-full" data-testid="cesium-container">
  {#if loadError}
    <div role="alert" class="p-md text-error text-sm" data-testid="cesium-error">
      Cesium failed to load: {loadError}
    </div>
  {/if}
</div>

<style>
  /* Cesium expects its container to be sized; the parent layout
     must provide a positioned context (height + width) for the
     globe canvas to render correctly. */
  .cesium-viewer {
    position: relative;
    min-height: 320px;
  }

  /* Suppress Cesium's default credit container in chetana; the
     attribution lives in the route's footer when needed. */
  .cesium-viewer :global(.cesium-viewer-bottom) {
    display: none !important;
  }
</style>
