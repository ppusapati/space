<script lang="ts">
  import DynamicFormRenderer from './DynamicFormRenderer.svelte';
  import type { FormSchema } from '@samavāya/core';

  // Example form schema with all 39 field types
  const exampleSchema = {
    fields: [
      // Basic Text Fields (5)
      {
        type: 'text',
        name: 'fullName',
        label: 'Full Name',
        placeholder: 'Enter your full name',
        required: true,
      },
      {
        type: 'email',
        name: 'email',
        label: 'Email Address',
        placeholder: 'your@email.com',
        required: true,
      },
      {
        type: 'password',
        name: 'password',
        label: 'Password',
        placeholder: 'Enter password',
        required: true,
      },
      {
        type: 'tel',
        name: 'phone',
        label: 'Phone Number',
      },
      {
        type: 'url',
        name: 'website',
        label: 'Website',
        placeholder: 'https://example.com',
      },

      // Number & Currency (3)
      {
        type: 'number',
        name: 'age',
        label: 'Age',
        min: 0,
        max: 120,
      },
      {
        type: 'percent',
        name: 'discount',
        label: 'Discount Percentage',
        min: 0,
        max: 100,
      },
      {
        type: 'currency',
        name: 'price',
        label: 'Price',
        currency: 'USD',
      },

      // Selection Fields (5)
      {
        type: 'select',
        name: 'country',
        label: 'Country',
        options: [
          { label: 'United States', value: 'us' },
          { label: 'India', value: 'in' },
          { label: 'United Kingdom', value: 'uk' },
        ],
      },
      {
        type: 'checkbox',
        name: 'agree',
        label: 'I agree to terms and conditions',
      },
      {
        type: 'radio',
        name: 'status',
        label: 'Status',
        options: [
          { label: 'Active', value: 'active' },
          { label: 'Inactive', value: 'inactive' },
        ],
      },
      {
        type: 'switch',
        name: 'newsletter',
        label: 'Subscribe to newsletter',
      },
      {
        type: 'autocomplete',
        name: 'tags',
        label: 'Tags',
        options: [
          { label: 'Important', value: 'important' },
          { label: 'Urgent', value: 'urgent' },
        ],
      },

      // Date/Time Fields (7)
      {
        type: 'date',
        name: 'birthDate',
        label: 'Birth Date',
      },
      {
        type: 'datetime',
        name: 'appointmentTime',
        label: 'Appointment Time',
      },
      {
        type: 'time',
        name: 'workStart',
        label: 'Work Start Time',
      },
      {
        type: 'month',
        name: 'billingMonth',
        label: 'Billing Month',
      },
      {
        type: 'year',
        name: 'graduationYear',
        label: 'Graduation Year',
      },
      {
        type: 'daterange',
        name: 'vacationDates',
        label: 'Vacation Dates',
      },
      {
        type: 'datetime-range',
        name: 'eventDates',
        label: 'Event Duration',
      },

      // File & Image (3)
      {
        type: 'file',
        name: 'document',
        label: 'Upload Document',
        accept: '.pdf,.doc',
      },
      {
        type: 'image',
        name: 'profilePicture',
        label: 'Profile Picture',
      },
      {
        type: 'barcode',
        name: 'productCode',
        label: 'Product Barcode',
      },

      // Text Areas & Rich Content (2)
      {
        type: 'textarea',
        name: 'description',
        label: 'Description',
        rows: 4,
      },
      {
        type: 'richtext',
        name: 'content',
        label: 'Content',
        minHeight: '200px',
      },

      // Data Visualization (2)
      {
        type: 'rating',
        name: 'satisfaction',
        label: 'Satisfaction Rating',
        max: 5,
      },
      {
        type: 'slider',
        name: 'volume',
        label: 'Volume',
        min: 0,
        max: 100,
      },

      // Color & Lookup (3)
      {
        type: 'color',
        name: 'favoriteColor',
        label: 'Favorite Color',
        format: 'hex',
      },
      {
        type: 'lookup',
        name: 'customer',
        label: 'Select Customer',
      },
      {
        type: 'multi-lookup',
        name: 'assignees',
        label: 'Assign To',
      },

      // Advanced Fields (9)
      {
        type: 'tag-input',
        name: 'keywords',
        label: 'Keywords',
      },
      {
        type: 'tree',
        name: 'category',
        label: 'Category Tree',
      },
      {
        type: 'cascade',
        name: 'location',
        label: 'Location',
      },
      {
        type: 'table',
        name: 'items',
        label: 'Order Items',
      },
      {
        type: 'json',
        name: 'metadata',
        label: 'Metadata (JSON)',
      },
      {
        type: 'cron',
        name: 'schedule',
        label: 'Recurring Schedule',
      },
      {
        type: 'phone',
        name: 'mobilePhone',
        label: 'Mobile Phone',
      },
    ],
    layout: {
      type: 'grid',
      columns: 2,
      gap: 'md',
      labelPosition: 'top',
    },
  };

  let formValues: Record<string, unknown> = {};
  let formErrors: Record<string, string> = {};

  function handleSubmit(values: Record<string, unknown>) {
    console.log('Form submitted:', values);
    formValues = values;
  }

  function handleReset() {
    console.log('Form reset');
    formValues = {};
    formErrors = {};
  }
</script>

<div class="p-6">
  <h1 class="mb-2 text-2xl font-bold">Dynamic Form Renderer</h1>
  <p class="mb-6 text-gray-600">All 39 field types demonstrated below</p>

  <DynamicFormRenderer
    schema={exampleSchema as FormSchema<Record<string, unknown>>}
    values={formValues}
    errors={formErrors}
    onSubmit={handleSubmit}
    onReset={handleReset}
    submitLabel="Submit Form"
    resetLabel="Clear Form"
    showReset={true}
  />

  {#if Object.keys(formValues).length > 0}
    <div class="mt-8 rounded-lg bg-gray-100 p-4">
      <h2 class="mb-4 text-lg font-bold">Form Values (JSON)</h2>
      <pre class="overflow-auto rounded bg-white p-4"><code>{JSON.stringify(
          formValues,
          null,
          2
        )}</code></pre>
    </div>
  {/if}
</div>

<style lang="postcss">
  :global(body) {
    @apply bg-white;
  }
</style>
