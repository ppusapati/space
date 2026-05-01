<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';
  import type { ChartAnimationConfig, TooltipConfig, SankeyNode, SankeyLink } from './types';

  // Props
  export let nodes: SankeyNode[] = [];
  export let links: SankeyLink[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showTooltip: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;

  // Enhanced props
  export let animation: ChartAnimationConfig | boolean = true;
  export let tooltip: TooltipConfig = {};
  export let colors: string[] = [];
  export let orient: 'horizontal' | 'vertical' = 'horizontal';
  export let nodeWidth: number = 20;
  export let nodeGap: number = 8;
  export let nodeAlign: 'left' | 'right' | 'justify' = 'justify';
  export let draggable: boolean = true;
  export let layoutIterations: number = 32;
  export let emphasis: 'allEdges' | 'adjacency' = 'adjacency';
  export let lineStyle: 'source' | 'target' | 'gradient' = 'gradient';
  export let curveness: number = 0.5;

  let className: string = '';
  export { className as class };

  let chartRef: EChart;

  const dispatch = createEventDispatcher();

  $: option = {
    title: title ? {
      text: title,
      subtext: subtitle,
      left: 'center',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      subtextStyle: {
        color: 'var(--color-text-secondary)',
      },
    } : undefined,
    tooltip: showTooltip || tooltip.show !== false ? {
      trigger: 'item',
      triggerOn: 'mousemove',
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
      formatter: tooltip.formatter,
    } : undefined,
    ...(colors.length > 0 ? { color: colors } : {}),
    series: [{
      type: 'sankey',
      left: '5%',
      top: title ? 60 : 20,
      right: '5%',
      bottom: 20,
      orient,
      nodeWidth,
      nodeGap,
      nodeAlign,
      draggable,
      layoutIterations,
      emphasis: {
        focus: emphasis,
      },
      data: nodes.map(n => ({
        name: n.name,
        value: n.value,
        itemStyle: n.itemStyle,
      })),
      links: links.map(l => ({
        source: l.source,
        target: l.target,
        value: l.value,
        lineStyle: l.lineStyle,
      })),
      lineStyle: {
        color: lineStyle,
        curveness,
      },
      label: {
        show: true,
        position: orient === 'horizontal' ? 'right' : 'bottom',
        color: 'var(--color-text-primary)',
        fontSize: 12,
      },
      itemStyle: {
        borderColor: '#fff',
        borderWidth: 1,
      },
    }],
  } as EChartsOption;

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

  export function downloadImage(filename = 'sankey-chart', options = {}) {
    chartRef?.downloadImage(filename, options);
  }

  export function exportImage(options = {}) {
    return chartRef?.exportImage(options);
  }
</script>

<div class={cn('sankey-chart', className)}>
  <EChart
    bind:this={chartRef}
    {option}
    {height}
    {loading}
    {animation}
    on:click={handleClick}
    on:init={handleInit}
  />
</div>
