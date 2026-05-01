<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type RadioOption,
    radioSizeClasses,
    radioColorClasses,
    radioBaseClasses,
    radioGroupClasses,
    radioItemClasses,
    radioLabelClasses,
    radioLabelDisabledClasses,
    radioDescriptionClasses,
  } from './radio.types';
  import type { Size, ColorVariant } from '../types';

  // Props
  export let value: string | undefined = undefined;
  export let options: RadioOption[] = [];
  export let name: string = uid('radio-group');
  export let size: Size = 'md';
  export let color: ColorVariant = 'primary';
  export let orientation: 'horizontal' | 'vertical' = 'vertical';
  export let labelPosition: 'left' | 'right' = 'right';
  export let disabled: boolean = false;
  export let required: boolean = false;
  export let id: string = uid('radio');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { value: string; option: RadioOption; event: Event };
  }>();

  // Computed classes
  $: sizeConfig = radioSizeClasses[size];
  $: colorConfig = radioColorClasses[color];
  $: groupClasses = cn(radioGroupClasses[orientation], className);

  function getRadioClasses(isSelected: boolean, isDisabled: boolean): string {
    return cn(
      radioBaseClasses,
      sizeConfig.radio,
      colorConfig.focus,
      isSelected && colorConfig.selected,
      isDisabled && 'opacity-50 cursor-not-allowed'
    );
  }

  function getLabelClasses(isDisabled: boolean): string {
    return cn(
      radioLabelClasses,
      sizeConfig.label,
      isDisabled && radioLabelDisabledClasses
    );
  }

  function handleChange(option: RadioOption, event: Event) {
    if (option.disabled || disabled) return;
    value = option.value;
    dispatch('change', { value: option.value, option, event });
  }

  function handleKeydown(option: RadioOption, event: KeyboardEvent) {
    if (event.key === ' ' || event.key === 'Enter') {
      event.preventDefault();
      handleChange(option, event);
    }
  }
</script>

<div
  class={groupClasses}
  role="radiogroup"
  aria-required={required}
  data-testid={testId || undefined}
>
  {#each options as option, index}
    {@const isSelected = value === option.value}
    {@const isDisabled = disabled || option.disabled}
    {@const optionId = `${id}-${index}`}

    <div class={cn(radioItemClasses, labelPosition === 'left' && 'flex-row-reverse')}>
      <div class="relative">
        <input
          type="radio"
          id={optionId}
          {name}
          value={option.value}
          checked={isSelected}
          disabled={isDisabled}
          {required}
          class="sr-only peer"
          on:change={(e) => handleChange(option, e)}
        />
        <div
          class={getRadioClasses(isSelected, isDisabled ?? false)}
          role="radio"
          tabindex={isDisabled ? -1 : 0}
          aria-checked={isSelected}
          aria-disabled={isDisabled}
          on:click={() => handleChange(option, new Event('click'))}
          on:keydown={(e) => handleKeydown(option, e)}
        >
          {#if isSelected}
            <div class={cn(sizeConfig.dot, colorConfig.dot, 'rounded-full')}></div>
          {/if}
        </div>
      </div>

      {#if option.label || option.description}
        <div class="flex flex-col">
          {#if option.label}
            <label for={optionId} class={getLabelClasses(isDisabled ?? false)}>
              {option.label}
            </label>
          {/if}
          {#if option.description}
            <span class={radioDescriptionClasses}>
              {option.description}
            </span>
          {/if}
        </div>
      {/if}
    </div>
  {/each}
</div>
