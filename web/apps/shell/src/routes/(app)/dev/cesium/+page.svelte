<!--
  /dev/cesium — non-production sanity route that mounts a Cesium
  viewer purely so TASK-P1-WEB-002's e2e + bundle-analyser
  inspections have a real route to drive against.

  Phase 2 + Phase 4 visualisation routes (ground tracks, sky plot,
  AOS/LOS timeline, etc.) reuse the same `<CesiumViewer />`
  component; this route exists ONLY to validate the chunk-split
  invariant today.

  Hidden behind /dev/ to make it clear it's not a customer-facing
  surface; the production route registry doesn't link to it.
-->
<script lang="ts">
  import CesiumViewer from "$lib/cesium/Viewer.svelte";

  let cesiumLoadStartedAt = $state(0);
  let cesiumReadyAt = $state(0);

  function onMounted() {
    cesiumLoadStartedAt = performance.now();
  }

  function onReady() {
    cesiumReadyAt = performance.now();
  }
</script>

<svelte:head><title>Cesium dev — Chetana</title></svelte:head>

<div class="flex flex-col gap-md h-full">
  <div>
    <h1 class="text-xl font-semibold text-text-primary">Cesium chunk-split sanity</h1>
    <p class="text-sm text-text-muted mt-2xs">
      This route is non-production. It exists so the WEB-002 e2e can verify
      the Cesium chunk loads on demand (not in the initial bundle) and that
      a globe renders successfully.
    </p>
    {#if cesiumReadyAt > 0 && cesiumLoadStartedAt > 0}
      <p class="text-xs text-text-muted">
        Ready in {(cesiumReadyAt - cesiumLoadStartedAt).toFixed(0)}ms
      </p>
    {/if}
  </div>

  <div class="flex-1 border border-border rounded overflow-hidden bg-surface" data-testid="cesium-host">
    <svelte:boundary onerror={(e) => console.error("Cesium boundary", e)}>
      <CesiumViewer onReady={() => onReady()} />
    </svelte:boundary>
  </div>
</div>

<svelte:options />

<script lang="ts" module>
  // Capture the mount-start time at module load so the timing
  // includes the dynamic-import + worker bootstrap, not just the
  // viewer constructor.
</script>
