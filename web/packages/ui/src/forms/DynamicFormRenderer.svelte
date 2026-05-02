<script lang="ts">
  import { writable } from 'svelte/store';
  import type { FormSchema, FormFieldConfig, FormSection } from '@chetana/core';
  import { cn } from '../utils/classnames';

  // Import all field components
  import Input from './Input.svelte';
  import TextArea from './TextArea.svelte';
  import NumberInput from './NumberInput.svelte';
  import Select from './Select.svelte';
  import Checkbox from './Checkbox.svelte';
  import Radio from './Radio.svelte';
  import Switch from './Switch.svelte';
  import DatePicker from './DatePicker.svelte';
  import DateRangePicker from './DateRangePicker.svelte';
  import TimePicker from './TimePicker.svelte';
  import ColorPicker from './ColorPicker.svelte';
  import FileUpload from './FileUpload.svelte';
  import Slider from './Slider.svelte';
  import Rating from './Rating.svelte';
  import TagInput from './TagInput.svelte';
  import Combobox from './Combobox.svelte';
  import RichTextEditor from './RichTextEditor.svelte';
  import CurrencyInput from './CurrencyInput.svelte';

  // New field components (17 created)
  import PhoneInput from './PhoneInput.svelte';
  import UrlInput from './UrlInput.svelte';
  import ImageUpload from './ImageUpload.svelte';
  import PercentageInput from './PercentageInput.svelte';
  import JsonEditor from './JsonEditor.svelte';
  import LookupField from './LookupField.svelte';
  import MultiLookupField from './MultiLookupField.svelte';
  import TreeSelector from './TreeSelector.svelte';
  import CascadeSelect from './CascadeSelect.svelte';
  import TableField from './TableField.svelte';
  import BarcodeInput from './BarcodeInput.svelte';
  import CronInput from './CronInput.svelte';
  import MonthPicker from './MonthPicker.svelte';
  import YearPicker from './YearPicker.svelte';
  import DateTimeRangeField from './DateTimeRangeField.svelte';
  import RepeaterField from './RepeaterField.svelte';
  import NestedForm from './NestedForm.svelte';
  import HiddenField from './HiddenField.svelte';
  import CustomFieldRenderer from './CustomFieldRenderer.svelte';
  import CheckboxGroup from './CheckboxGroup.svelte';
  import KeyValueEditor from './KeyValueEditor.svelte';
  import FormulaField from './FormulaField.svelte';

  // Props
  export let schema: FormSchema<Record<string, unknown>>;
  export let values: Record<string, unknown> = {};
  export let errors: Record<string, string> = {};
  export let touched: Record<string, boolean> = {};
  export let readonly: boolean = false;
  export let disabled: boolean = false;
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let submitLabel: string = 'Submit';
  export let resetLabel: string = 'Reset';
  export let showReset: boolean = true;
  export let onSubmit: ((values: Record<string, unknown>) => void | Promise<void>) | null = null;
  export let onReset: (() => void) | null = null;

  let className: string = '';
  export { className as class };

  // State
  const formValues = writable<Record<string, unknown>>(values);
  const formErrors = writable<Record<string, string>>(errors);
  const formTouched = writable<Record<string, boolean>>(touched);
  let isSubmitting = false;

  // Subscribe to external changes
  $: formValues.set(values);
  $: formErrors.set(errors);
  $: formTouched.set(touched);

  /**
   * Map field type to component
   */
  function getFieldComponent(
    fieldType: FormFieldConfig['type']
  ): any {
    const componentMap: Partial<Record<string, any>> = {
      text: Input,
      email: Input,
      password: Input,
      tel: Input,
      url: UrlInput,
      search: Input,
      textarea: TextArea,
      number: NumberInput,
      select: Select,
      checkbox: Checkbox,
      'checkbox-group': CheckboxGroup,
      radio: Radio,
      switch: Switch,
      date: DatePicker,
      datetime: DatePicker,
      time: TimePicker,
      month: MonthPicker,
      year: YearPicker,
      daterange: DateRangePicker,
      'datetime-range': DateTimeRangeField,
      color: ColorPicker,
      file: FileUpload,
      slider: Slider,
      rating: Rating,
      autocomplete: Combobox,
      'tag-input': TagInput,
      richtext: RichTextEditor,
      currency: CurrencyInput,
      phone: PhoneInput,
      image: ImageUpload,
      percent: PercentageInput,
      json: JsonEditor,
      lookup: LookupField,
      'multi-lookup': MultiLookupField,
      tree: TreeSelector,
      cascade: CascadeSelect,
      table: TableField,
      barcode: BarcodeInput,
      cron: CronInput,
      array: RepeaterField,
      object: NestedForm,
      hidden: HiddenField,
      custom: CustomFieldRenderer,
      keyvalue: KeyValueEditor,
      formula: FormulaField,
    } as const;

    return componentMap[fieldType as string] ?? Input;
  }

  /**
   * Handle field change
   */
  function handleFieldChange(fieldName: string, value: unknown) {
    $formValues[fieldName] = value;
    $formTouched[fieldName] = true;
  }

  /**
   * Handle field blur
   */
  function handleFieldBlur(fieldName: string) {
    $formTouched[fieldName] = true;
  }

  /**
   * Handle form submit
   */
  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (isSubmitting) return;

    isSubmitting = true;
    try {
      if (onSubmit) {
        await onSubmit($formValues);
      }
      // Dispatch custom event
      dispatch('submit', { values: $formValues });
    } catch (err) {
      console.error('Form submission error:', err);
      dispatch('error', { error: err });
    } finally {
      isSubmitting = false;
    }
  }

  /**
   * Handle form reset
   */
  function handleReset(e: Event) {
    e.preventDefault();
    $formValues = values;
    $formTouched = {};
    $formErrors = {};
    if (onReset) onReset();
    dispatch('reset', undefined);
  }

  /**
   * Get field by name
   */
  function getField(name: string): FormFieldConfig | undefined {
    return schema.fields.find((f) => f.name === name);
  }

  /**
   * Check if field should be visible
   */
  function isFieldVisible(field: FormFieldConfig): boolean {
    if (field.hidden) return false;
    if (field.condition) {
      return field.condition($formValues);
    }
    return true;
  }

  /**
   * Get section fields
   */
  function getSectionFields(section: FormSection): FormFieldConfig[] {
    return section.fields
      .map((fieldName) => schema.fields.find((f) => f.name === fieldName))
      .filter((f): f is FormFieldConfig => f !== undefined);
  }

  /**
   * Create event dispatcher
   */
  function dispatch(type: string, detail: unknown) {
    const event = new CustomEvent(type, { detail });
    (element as HTMLFormElement).dispatchEvent(event);
  }

  let element: HTMLElement;

  // Get sections or create default
  $: sections = schema.layout?.sections || [
    {
      id: 'default',
      title: undefined,
      fields: schema.fields.map((f) => f.name),
    },
  ];

  // Compute layout classes
  $: layoutClass = cn(
    'form-renderer',
    schema.layout?.type === 'horizontal' && 'form-horizontal',
    schema.layout?.type === 'grid' && 'form-grid',
    schema.layout?.type === 'inline' && 'form-inline'
  );

  // Compute responsive gap
  $: gapClass = ({
    sm: 'gap-2',
    md: 'gap-4',
    lg: 'gap-6',
  } as Record<string, string>)[schema.layout?.gap || 'md'] ?? 'gap-4';
</script>

<form bind:this={element} on:submit={handleSubmit} class={cn(layoutClass, gapClass, className)}>
  <!-- Render sections -->
  {#each sections as section (section.id)}
    {#if getSectionFields(section).some((f) => isFieldVisible(f))}
      <fieldset class="form-section">
        {#if section.title}
          <legend class="section-title">{section.title}</legend>
        {/if}
        {#if section.description}
          <p class="section-description">{section.description}</p>
        {/if}

        <div
          class={cn(
            'section-fields',
            schema.layout?.type === 'grid' &&
              `grid-cols-${section.columns || schema.layout?.columns || 1}`
          )}
        >
          <!-- Render fields in section -->
          {#each getSectionFields(section) as field (field.name)}
            {#if isFieldVisible(field)}
              <div class="form-group">
                {#if field.label && field.type !== 'hidden' && field.type !== 'checkbox' && field.type !== 'switch'}
                  <label for={field.name} class="form-label">
                    {field.label}
                    {#if field.required}
                      <span class="required">*</span>
                    {/if}
                  </label>
                {/if}

                <div class="form-control">
                  <!-- Dynamic field rendering -->
                  <svelte:component
                    this={getFieldComponent(field.type)}
                    bind:value={$formValues[field.name]}
                    name={field.name}
                    label={field.label}
                    placeholder={field.placeholder}
                    disabled={disabled || field.disabled || readonly}
                    readonly={readonly || field.readonly}
                    required={field.required}
                    size={field.type === 'hidden' ? undefined : size}
                    {...(field.type === 'select' || field.type === 'autocomplete'
                      ? { options: (field as any).options }
                      : {})}
                    {...(field.type === 'number'
                      ? {
                          min: (field as any).min,
                          max: (field as any).max,
                          step: (field as any).step,
                        }
                      : {})}
                    {...(field.type === 'textarea'
                      ? { rows: (field as any).rows }
                      : {})}
                    {...(field.type === 'date' ||
                    field.type === 'datetime' ||
                    field.type === 'time'
                      ? { format: (field as any).format }
                      : {})}
                    {...(field.type === 'file'
                      ? {
                          accept: (field as any).accept,
                          multiple: (field as any).multiple,
                          maxSize: (field as any).maxSize,
                        }
                      : {})}
                    {...(field.type === 'slider'
                      ? {
                          min: (field as any).min,
                          max: (field as any).max,
                          step: (field as any).step,
                        }
                      : {})}
                    {...(field.type === 'rating'
                      ? { max: (field as any).max }
                      : {})}
                    on:change={(e: any) => handleFieldChange(field.name, e.detail ?? (e.target as HTMLInputElement | null)?.value)}
                    on:blur={() => handleFieldBlur(field.name)}
                  />
                </div>

                <!-- Helper text -->
                {#if field.helperText}
                  <p class="helper-text">{field.helperText}</p>
                {/if}

                <!-- Error message -->
                {#if $formErrors[field.name]}
                  <p class="error-text">{$formErrors[field.name]}</p>
                {/if}
              </div>
            {/if}
          {/each}
        </div>
      </fieldset>
    {/if}
  {/each}

  <!-- Form actions -->
  <div class="form-actions">
    <button
      type="submit"
      disabled={isSubmitting || disabled}
      class="btn btn-primary"
      aria-busy={isSubmitting}
    >
      {isSubmitting ? 'Submitting...' : submitLabel}
    </button>

    {#if showReset}
      <button type="reset" disabled={disabled} class="btn btn-secondary" on:click={handleReset}>
        {resetLabel}
      </button>
    {/if}
  </div>
</form>

<style lang="postcss">
  :global(.form-renderer) {
    @apply flex flex-col;
  }

  :global(.form-horizontal) {
    @apply flex-row;
  }

  :global(.form-grid) {
    @apply grid;
  }

  :global(.form-inline) {
    @apply flex flex-row flex-wrap items-center;
  }

  :global(.form-section) {
    @apply border-0 p-0;
  }

  :global(.section-title) {
    @apply mb-3 text-lg font-semibold;
  }

  :global(.section-description) {
    @apply mb-4 text-sm text-gray-500;
  }

  :global(.section-fields) {
    @apply flex flex-col gap-4;
  }

  :global(.grid-cols-1) {
    @apply grid grid-cols-1;
  }

  :global(.grid-cols-2) {
    @apply grid grid-cols-2;
  }

  :global(.grid-cols-3) {
    @apply grid grid-cols-3;
  }

  :global(.form-group) {
    @apply flex flex-col;
  }

  :global(.form-label) {
    @apply mb-2 block text-sm font-medium text-gray-700;
  }

  :global(.required) {
    @apply ml-1 text-red-500;
  }

  :global(.form-control) {
    @apply relative;
  }

  :global(.helper-text) {
    @apply mt-1 text-xs text-gray-500;
  }

  :global(.error-text) {
    @apply mt-1 text-xs text-red-500;
  }

  :global(.form-actions) {
    @apply mt-6 flex gap-3;
  }

  :global(.btn) {
    @apply rounded-md px-4 py-2 font-medium transition-colors;
  }

  :global(.btn-primary) {
    @apply bg-blue-600 text-white hover:bg-blue-700 disabled:bg-gray-400;
  }

  :global(.btn-secondary) {
    @apply border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 disabled:bg-gray-100;
  }
</style>
