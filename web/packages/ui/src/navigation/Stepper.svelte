<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { stepperClasses, type StepItem } from './navigation.types';
  import type { Size } from '../types';

  // Props
  export let steps: StepItem[] = [];
  export let currentStep: number = 0;
  export let orientation: 'horizontal' | 'vertical' = 'horizontal';
  export let size: Size = 'md';
  export let clickable: boolean = false;
  export let id: string = uid('stepper');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    stepClick: { step: number; item: StepItem };
  }>();

  function getStepStatus(index: number): 'completed' | 'active' | 'pending' | 'error' {
    if (steps[index]!.error) return 'error';
    if (index < currentStep) return 'completed';
    if (index === currentStep) return 'active';
    return 'pending';
  }

  function handleStepClick(index: number) {
    if (!clickable || index > currentStep) return;
    dispatch('stepClick', { step: index, item: steps[index]! });
  }

  $: containerClasses = cn(
    stepperClasses.container,
    orientation === 'vertical' && stepperClasses.containerVertical,
    className
  );
</script>

<div
  {id}
  class={containerClasses}
  data-testid={testId || undefined}
  role="list"
  aria-label="Progress"
>
  {#each steps as step, index (step.id)}
    {@const status = getStepStatus(index)}
    {@const isClickable = clickable && index <= currentStep}

    <div
      class={cn(
        stepperClasses.step,
        orientation === 'vertical' && stepperClasses.stepVertical,
        orientation === 'horizontal' && index < steps.length - 1 && 'flex-1'
      )}
    >
      <div class="flex items-center">
        <!-- Step indicator -->
        <button
          type="button"
          class={cn(
            stepperClasses.indicator,
            status === 'completed' && stepperClasses.indicatorCompleted,
            status === 'active' && stepperClasses.indicatorActive,
            status === 'pending' && stepperClasses.indicatorPending,
            status === 'error' && stepperClasses.indicatorError,
            isClickable && 'cursor-pointer hover:scale-105 transition-transform',
            !isClickable && 'cursor-default'
          )}
          disabled={!isClickable}
          on:click={() => handleStepClick(index)}
          aria-current={status === 'active' ? 'step' : undefined}
        >
          {#if status === 'completed'}
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          {:else if status === 'error'}
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
            </svg>
          {:else}
            {index + 1}
          {/if}
        </button>

        <!-- Step content -->
        <div class={stepperClasses.content}>
          <span class={stepperClasses.title}>
            {step.title}
            {#if step.optional}
              <span class="font-normal text-neutral-400 ml-1">(optional)</span>
            {/if}
          </span>
          {#if step.description}
            <p class={stepperClasses.description}>{step.description}</p>
          {/if}
        </div>
      </div>

      <!-- Connector line -->
      {#if index < steps.length - 1}
        <div
          class={cn(
            orientation === 'horizontal'
              ? stepperClasses.connector
              : stepperClasses.connectorVertical,
            index < currentStep && stepperClasses.connectorActive
          )}
        ></div>
      {/if}
    </div>
  {/each}
</div>
