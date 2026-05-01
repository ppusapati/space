<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import * as echarts from 'echarts';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import { formatNumber } from './types';
  import type { ChartAnimationConfig } from './types';
  import type { MapDataItem, MapType } from './MapChart.types';

  // Props
  export let data: MapDataItem[] = [];
  export let geoJson: object | undefined = undefined;
  export let mapType: MapType = 'scatter';
  export let title: string = '';
  export let height: string = '400px';
  export let loading: boolean = false;
  export let theme: string | object = '';
  export let colorRange: [string, string] = ['#313695', '#d73027'];
  export let showLabels: boolean = false;
  export let zoom: number = 1;
  export let center: [number, number] | undefined = undefined;
  export let animation: ChartAnimationConfig | boolean = true;

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const MAP_NAME = '__samavaya_custom_map__';

  const dispatch = createEventDispatcher();

  // Register custom GeoJSON when provided
  $: if (geoJson) {
    echarts.registerMap(MAP_NAME, geoJson as any);
  }

  // Compute value range for visual mapping
  $: values = data.map((d) => d.value);
  $: minValue = values.length > 0 ? Math.min(...values) : 0;
  $: maxValue = values.length > 0 ? Math.max(...values) : 100;

  // Build scatter/point map option
  function buildScatterOption(): EChartsOption {
    const hasGeo = !!geoJson;

    const geoConfig = hasGeo
      ? {
          geo: {
            map: MAP_NAME,
            roam: true,
            zoom,
            ...(center ? { center: [center[1], center[0]] } : {}),
            label: {
              show: showLabels,
              color: 'var(--color-text-secondary)',
              fontSize: 10,
            },
            itemStyle: {
              areaColor: 'var(--color-surface-secondary, #f3f4f6)',
              borderColor: 'var(--color-border-primary, #d1d5db)',
            },
            emphasis: {
              itemStyle: {
                areaColor: 'var(--color-surface-tertiary, #e5e7eb)',
              },
            },
          },
        }
      : {};

    const seriesData = data
      .filter((d) => d.lat != null && d.lng != null)
      .map((d) => ({
        name: d.name,
        value: [d.lng!, d.lat!, d.value],
        itemStyle: d.color ? { color: d.color } : undefined,
      }));

    return {
      ...geoConfig,
      tooltip: {
        trigger: 'item',
        backgroundColor: 'var(--color-surface-primary)',
        borderColor: 'var(--color-border-primary)',
        textStyle: { color: 'var(--color-text-primary)' },
        formatter: (params: any) => {
          const val = Array.isArray(params.value) ? params.value[2] : params.value;
          return `<strong>${params.name}</strong><br/>Value: ${formatNumber(val ?? 0)}`;
        },
      },
      visualMap: {
        min: minValue,
        max: maxValue,
        calculable: true,
        inRange: { color: [colorRange[0], colorRange[1]] },
        textStyle: { color: 'var(--color-text-secondary)' },
        right: '2%',
        bottom: '5%',
      },
      series: [
        {
          type: 'scatter',
          coordinateSystem: hasGeo ? 'geo' : 'cartesian2d',
          data: seriesData,
          symbolSize: (val: number[]) => {
            const v = val[2] ?? 0;
            const range = maxValue - minValue || 1;
            return Math.max(6, Math.min(30, ((v - minValue) / range) * 24 + 6));
          },
          encode: hasGeo ? { value: 2 } : undefined,
          label: {
            show: showLabels,
            formatter: '{b}',
            position: 'right',
            fontSize: 10,
            color: 'var(--color-text-secondary)',
          },
        },
      ],
    } as EChartsOption;
  }

  // Build heatmap/density option (uses scatter with effectScatter for emphasis)
  function buildHeatmapOption(): EChartsOption {
    const hasGeo = !!geoJson;

    const geoConfig = hasGeo
      ? {
          geo: {
            map: MAP_NAME,
            roam: true,
            zoom,
            ...(center ? { center: [center[1], center[0]] } : {}),
            label: { show: false },
            itemStyle: {
              areaColor: 'var(--color-surface-secondary, #f3f4f6)',
              borderColor: 'var(--color-border-primary, #d1d5db)',
            },
            emphasis: {
              itemStyle: {
                areaColor: 'var(--color-surface-tertiary, #e5e7eb)',
              },
            },
          },
        }
      : {};

    const seriesData = data
      .filter((d) => d.lat != null && d.lng != null)
      .map((d) => ({
        name: d.name,
        value: [d.lng!, d.lat!, d.value],
      }));

    return {
      ...geoConfig,
      tooltip: {
        trigger: 'item',
        backgroundColor: 'var(--color-surface-primary)',
        borderColor: 'var(--color-border-primary)',
        textStyle: { color: 'var(--color-text-primary)' },
        formatter: (params: any) => {
          const val = Array.isArray(params.value) ? params.value[2] : params.value;
          return `<strong>${params.name}</strong><br/>Value: ${formatNumber(val ?? 0)}`;
        },
      },
      visualMap: {
        min: minValue,
        max: maxValue,
        calculable: true,
        inRange: {
          color: [colorRange[0], colorRange[1]],
          symbolSize: [6, 30],
        },
        textStyle: { color: 'var(--color-text-secondary)' },
        right: '2%',
        bottom: '5%',
      },
      series: [
        {
          type: 'effectScatter',
          coordinateSystem: hasGeo ? 'geo' : 'cartesian2d',
          data: seriesData,
          encode: hasGeo ? { value: 2 } : undefined,
          symbolSize: (val: number[]) => {
            const v = val[2] ?? 0;
            const range = maxValue - minValue || 1;
            return Math.max(6, Math.min(40, ((v - minValue) / range) * 34 + 6));
          },
          showEffectOn: 'render',
          rippleEffect: {
            brushType: 'stroke',
            scale: 3,
          },
          label: {
            show: showLabels,
            formatter: '{b}',
            position: 'right',
            fontSize: 10,
            color: 'var(--color-text-secondary)',
          },
        },
      ],
    } as EChartsOption;
  }

  // Build choropleth/regions option
  function buildRegionsOption(): EChartsOption {
    if (!geoJson) {
      // Regions mode requires GeoJSON; fall back to empty chart with message
      return {
        title: {
          text: title || 'Map',
          subtext: 'GeoJSON data required for region map',
          left: 'center',
          top: 'center',
          textStyle: { color: 'var(--color-text-primary)' },
          subtextStyle: { color: 'var(--color-text-tertiary)', fontSize: 14 },
        },
      } as EChartsOption;
    }

    const seriesData = data.map((d) => ({
      name: d.name,
      value: d.value,
      itemStyle: d.color ? { color: d.color } : undefined,
    }));

    return {
      tooltip: {
        trigger: 'item',
        backgroundColor: 'var(--color-surface-primary)',
        borderColor: 'var(--color-border-primary)',
        textStyle: { color: 'var(--color-text-primary)' },
        formatter: (params: any) => {
          return `<strong>${params.name}</strong><br/>Value: ${formatNumber(params.value ?? 0)}`;
        },
      },
      visualMap: {
        min: minValue,
        max: maxValue,
        calculable: true,
        inRange: { color: [colorRange[0], colorRange[1]] },
        textStyle: { color: 'var(--color-text-secondary)' },
        right: '2%',
        bottom: '5%',
      },
      series: [
        {
          type: 'map',
          map: MAP_NAME,
          roam: true,
          zoom,
          ...(center ? { center: [center[1], center[0]] } : {}),
          data: seriesData,
          label: {
            show: showLabels,
            color: 'var(--color-text-primary)',
            fontSize: 10,
          },
          emphasis: {
            label: {
              show: true,
              fontWeight: 'bold',
            },
            itemStyle: {
              areaColor: 'var(--color-interactive-primary, #3b82f6)',
            },
          },
          itemStyle: {
            areaColor: 'var(--color-surface-secondary, #f3f4f6)',
            borderColor: 'var(--color-border-primary, #d1d5db)',
          },
        },
      ],
    } as EChartsOption;
  }

  // Build empty state option
  function buildEmptyOption(): EChartsOption {
    return {
      title: {
        text: title || 'Map',
        subtext: 'No data available',
        left: 'center',
        top: 'center',
        textStyle: { color: 'var(--color-text-primary)' },
        subtextStyle: { color: 'var(--color-text-tertiary)', fontSize: 14 },
      },
    } as EChartsOption;
  }

  // Compute the ECharts option based on mapType and data
  $: option = (() => {
    if (!data || data.length === 0) {
      return buildEmptyOption();
    }

    const baseTitle = title
      ? {
          title: {
            text: title,
            left: 'center',
            textStyle: { color: 'var(--color-text-primary)' },
          },
        }
      : {};

    let chartOption: EChartsOption;

    switch (mapType) {
      case 'heatmap':
        chartOption = buildHeatmapOption();
        break;
      case 'regions':
        chartOption = buildRegionsOption();
        break;
      case 'scatter':
      default:
        chartOption = buildScatterOption();
        break;
    }

    return { ...chartOption, ...baseTitle } as EChartsOption;
  })();

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }

  function handleInit(event: CustomEvent) {
    dispatch('init', event.detail);
  }

  // Expose chart methods
  export function getChart() {
    return chartRef?.getChart();
  }

  export function downloadImage(filename = 'map-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('map-chart', className)}>
  <EChart
    bind:this={chartRef}
    {option}
    {height}
    {loading}
    {theme}
    {animation}
    on:click={handleClick}
    on:init={handleInit}
  />
</div>
