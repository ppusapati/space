<script context="module" lang="ts">
  export interface TreemapNode {
    name: string;
    value: number;
    children?: TreemapNode[];
    itemStyle?: {
      color?: string;
    };
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { EChartsOption } from 'echarts';
  import EChart from './EChart.svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let data: TreemapNode[] = [];
  export let title: string = '';
  export let subtitle: string = '';
  export let showTooltip: boolean = true;
  export let showBreadcrumb: boolean = true;
  export let height: string = '400px';
  export let loading: boolean = false;
  export let roam: boolean | 'scale' | 'move' = false;
  export let leafDepth: number | null = null;

  let className: string = '';
  export { className as class };

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
    tooltip: showTooltip ? {
      formatter: (info: any) => {
        const value = info.value;
        const treePathInfo = info.treePathInfo;
        const treePath: string[] = [];

        for (let i = 1; i < treePathInfo.length; i++) {
          treePath.push(treePathInfo[i].name);
        }

        return [
          '<div class="tooltip-title">' + treePath.join(' / ') + '</div>',
          'Value: ' + value,
        ].join('');
      },
      backgroundColor: 'var(--color-surface-primary)',
      borderColor: 'var(--color-border-primary)',
      textStyle: {
        color: 'var(--color-text-primary)',
      },
    } : undefined,
    series: [
      {
        name: title || 'Treemap',
        type: 'treemap',
        roam: roam,
        leafDepth: leafDepth,
        data: data,
        breadcrumb: showBreadcrumb ? {
          show: true,
          top: title ? 50 : 10,
          left: 'center',
          itemStyle: {
            color: 'var(--color-surface-secondary)',
            borderColor: 'var(--color-border-primary)',
            textStyle: {
              color: 'var(--color-text-primary)',
            },
          },
        } : {
          show: false,
        },
        upperLabel: {
          show: true,
          height: 30,
          color: 'var(--color-text-primary)',
        },
        label: {
          show: true,
          formatter: '{b}',
          color: 'var(--color-text-primary)',
        },
        itemStyle: {
          borderColor: 'var(--color-surface-primary)',
          borderWidth: 2,
          gapWidth: 2,
        },
        levels: [
          {
            itemStyle: {
              borderColor: 'var(--color-border-primary)',
              borderWidth: 0,
              gapWidth: 5,
            },
            upperLabel: {
              show: false,
            },
          },
          {
            itemStyle: {
              borderColor: 'var(--color-border-secondary)',
              borderWidth: 5,
              gapWidth: 2,
            },
            emphasis: {
              itemStyle: {
                borderColor: 'var(--color-interactive-primary)',
              },
            },
          },
          {
            colorSaturation: [0.35, 0.5],
            itemStyle: {
              borderWidth: 3,
              gapWidth: 2,
              borderColorSaturation: 0.6,
            },
          },
        ],
      },
    ],
  } as EChartsOption;

  function handleClick(event: CustomEvent) {
    dispatch('click', event.detail);
  }
</script>

<div class={cn('treemap-chart', className)}>
  <EChart {option} {height} {loading} on:click={handleClick} />
</div>
