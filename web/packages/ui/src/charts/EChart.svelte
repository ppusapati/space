<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import * as echarts from 'echarts';
  import type { EChartsOption, ECharts as EChartsInstance } from 'echarts';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, ChartExportOptions, DataZoomConfig } from './types';

  // Props
  export let option: EChartsOption;
  export let theme: string | object = '';
  export let width: string = '100%';
  export let height: string = '400px';
  export let loading: boolean = false;
  export let autoResize: boolean = true;
  export let renderer: 'canvas' | 'svg' = 'canvas';
  export let notMerge: boolean = false;
  export let lazyUpdate: boolean = false;

  // Animation props
  export let animation: ChartAnimationConfig | boolean = true;

  // Data zoom props
  export let dataZoom: DataZoomConfig | DataZoomConfig[] | boolean = false;

  // Color palette
  export let colors: string[] = [];

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    init: { chart: EChartsInstance };
    click: { params: unknown };
    dblclick: { params: unknown };
    mousedown: { params: unknown };
    mousemove: { params: unknown };
    mouseup: { params: unknown };
    mouseover: { params: unknown };
    mouseout: { params: unknown };
    legendselectchanged: { params: unknown };
    datazoom: { params: unknown };
    brushselected: { params: unknown };
    rendered: void;
    finished: void;
  }>();

  let chartContainer: HTMLDivElement;
  let chart: EChartsInstance | null = null;
  let resizeObserver: ResizeObserver | null = null;

  // Build animation options
  function getAnimationOptions(): Partial<EChartsOption> {
    if (animation === false) {
      return { animation: false };
    }
    if (animation === true) {
      return { animation: true };
    }
    return {
      animation: animation.enabled !== false,
      animationDuration: animation.duration ?? 1000,
      animationEasing: animation.easing ?? 'cubicOut',
      animationDelay: animation.delay ?? 0,
      animationThreshold: animation.threshold ?? 2000,
    };
  }

  // Build data zoom options
  function getDataZoomOptions(): EChartsOption['dataZoom'] {
    if (!dataZoom) return undefined;
    if (dataZoom === true) {
      return [
        { type: 'inside', start: 0, end: 100 },
        { type: 'slider', start: 0, end: 100 },
      ];
    }
    return Array.isArray(dataZoom) ? dataZoom : [dataZoom];
  }

  // Merge options with animation and zoom
  function getMergedOption(): EChartsOption {
    const animationOpts = getAnimationOptions();
    const dataZoomOpts = getDataZoomOptions();

    return {
      ...option,
      ...animationOpts,
      ...(dataZoomOpts ? { dataZoom: dataZoomOpts } : {}),
      ...(colors.length > 0 ? { color: colors } : {}),
    };
  }

  function initChart() {
    if (!chartContainer) return;

    chart = echarts.init(chartContainer, theme, {
      renderer,
      width: 'auto',
      height: 'auto',
    });

    chart.setOption(getMergedOption(), notMerge, lazyUpdate);

    // Add event listeners
    chart.on('click', (params) => dispatch('click', { params }));
    chart.on('dblclick', (params) => dispatch('dblclick', { params }));
    chart.on('mousedown', (params) => dispatch('mousedown', { params }));
    chart.on('mousemove', (params) => dispatch('mousemove', { params }));
    chart.on('mouseup', (params) => dispatch('mouseup', { params }));
    chart.on('mouseover', (params) => dispatch('mouseover', { params }));
    chart.on('mouseout', (params) => dispatch('mouseout', { params }));
    chart.on('legendselectchanged', (params) => dispatch('legendselectchanged', { params }));
    chart.on('datazoom', (params) => dispatch('datazoom', { params }));
    chart.on('brushselected', (params) => dispatch('brushselected', { params }));
    chart.on('rendered', () => dispatch('rendered'));
    chart.on('finished', () => dispatch('finished'));

    dispatch('init', { chart });
  }

  function destroyChart() {
    if (chart) {
      chart.dispose();
      chart = null;
    }
  }

  // Update option when it changes
  $: if (chart && option) {
    chart.setOption(getMergedOption(), notMerge, lazyUpdate);
  }

  // Handle loading state
  $: if (chart) {
    if (loading) {
      chart.showLoading({
        text: 'Loading...',
        color: 'var(--color-interactive-primary)',
        textColor: 'var(--color-text-primary)',
        maskColor: 'rgba(255, 255, 255, 0.8)',
        zlevel: 0,
        fontSize: 14,
        showSpinner: true,
        spinnerRadius: 12,
        lineWidth: 3,
      });
    } else {
      chart.hideLoading();
    }
  }

  onMount(() => {
    initChart();

    if (autoResize) {
      resizeObserver = new ResizeObserver(() => {
        chart?.resize();
      });
      resizeObserver.observe(chartContainer);
    }
  });

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
    destroyChart();
  });

  // Expose methods
  export function getChart(): EChartsInstance | null {
    return chart;
  }

  export function resize() {
    chart?.resize();
  }

  export function clear() {
    chart?.clear();
  }

  export function disposeChart() {
    destroyChart();
  }

  // Export chart as image
  export function exportImage(options: ChartExportOptions = {}): string | undefined {
    if (!chart) return undefined;

    const {
      type = 'png',
      pixelRatio = 2,
      backgroundColor = '#fff',
      excludeComponents = [],
    } = options;

    return chart.getDataURL({
      type: type === 'svg' ? 'svg' : type,
      pixelRatio,
      backgroundColor,
      excludeComponents,
    });
  }

  // Download chart as image
  export function downloadImage(filename: string = 'chart', options: ChartExportOptions = {}) {
    const dataUrl = exportImage(options);
    if (!dataUrl) return;

    const link = document.createElement('a');
    link.download = `${filename}.${options.type || 'png'}`;
    link.href = dataUrl;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  // Convert to SVG (only works when renderer is 'svg')
  export function toSVG(): string | undefined {
    if (!chart || renderer !== 'svg') return undefined;
    return (chart as unknown as { renderToSVGString: () => string }).renderToSVGString?.();
  }

  // Get current option
  export function getOption(): EChartsOption | undefined {
    return chart?.getOption() as EChartsOption | undefined;
  }

  // Set option programmatically
  export function setOption(newOption: EChartsOption, merge = true) {
    chart?.setOption(newOption, !merge, lazyUpdate);
  }

  // Highlight series
  export function highlight(params: { seriesIndex?: number; seriesName?: string; dataIndex?: number }) {
    chart?.dispatchAction({ type: 'highlight', ...params });
  }

  // Downplay (remove highlight)
  export function downplay(params: { seriesIndex?: number; seriesName?: string; dataIndex?: number }) {
    chart?.dispatchAction({ type: 'downplay', ...params });
  }

  // Show tooltip programmatically
  export function showTooltip(params: { seriesIndex?: number; dataIndex?: number; x?: number; y?: number }) {
    chart?.dispatchAction({ type: 'showTip', ...params });
  }

  // Hide tooltip
  export function hideTooltip() {
    chart?.dispatchAction({ type: 'hideTip' });
  }

  // Toggle legend selection
  export function toggleLegend(name: string) {
    chart?.dispatchAction({ type: 'legendToggleSelect', name });
  }

  // Data zoom control
  export function zoomTo(start: number, end: number, dataZoomIndex = 0) {
    chart?.dispatchAction({ type: 'dataZoom', dataZoomIndex, start, end });
  }

  // Reset zoom
  export function resetZoom() {
    chart?.dispatchAction({ type: 'dataZoom', start: 0, end: 100 });
  }
</script>

<div
  bind:this={chartContainer}
  class={cn('echart-container', className)}
  style="width: {width}; height: {height};"
></div>

<style>
  .echart-container {
    position: relative;
  }
</style>
