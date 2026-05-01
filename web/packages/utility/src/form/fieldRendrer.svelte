<script lang="ts">
  import { addRepeatedField, removeRepeatedField, initFormData } from './formHelper';
  export let field: FormField;
  export let formData: any;
</script>

<div style="margin-bottom: 1rem;">
  <label>{field.label}</label>

  {#if field.repeated}
    {#each formData[field.name] as item, idx}
      <div style="display:flex; flex-direction: column; gap:0.25rem; border:1px solid #eee; padding:0.5rem; margin-bottom:0.5rem;">
        {#if field.type === 'object'}
          {#each field.nested as nestedField}
            <label>{nestedField.label}</label>
            <input type={nestedField.widget} bind:value={formData[field.name][idx][nestedField.name]} min={nestedField.min} max={nestedField.max} placeholder={nestedField.placeholder}/>
            {#if nestedField.help}<small style="color:gray;">{nestedField.help}</small>{/if}
            {#if nestedField.error}<small style="color:red;">{nestedField.error}</small>{/if}
          {/each}
        {:else if field.enum}
          <select bind:value={formData[field.name][idx]}>
            {#each field.enum as option}
              <option value={option}>{option}</option>
            {/each}
          </select>
        {:else if field.type === 'boolean'}
          <input type="checkbox" bind:checked={formData[field.name][idx]} />
        {:else}
          <input type={field.widget} bind:value={formData[field.name][idx]} min={field.min} max={field.max} placeholder={field.placeholder}/>
        {/if}
        <button type="button" on:click={() => removeRepeatedField(formData, field.name, idx)}>Remove</button>
      </div>
    {/each}
    <button type="button" on:click={() => addRepeatedField(formData, field)}>Add {field.label}</button>

  {:else if field.type === 'object'}
    {#each field.nested as nestedField}
      <label>{nestedField.label}</label>
      <input type={nestedField.widget} bind:value={formData[field.name][nestedField.name]} min={nestedField.min} max={nestedField.max} placeholder={nestedField.placeholder}/>
      {#if nestedField.help}<small style="color:gray;">{nestedField.help}</small>{/if}
      {#if nestedField.error}<small style="color:red;">{nestedField.error}</small>{/if}
    {/each}

  {:else if field.enum}
    <select bind:value={formData[field.name]}>
      {#each field.enum as option}
        <option value={option}>{option}</option>
      {/each}
    </select>

  {:else if field.type === 'boolean'}
    <input type="checkbox" bind:checked={formData[field.name]} />

  {:else}
    <input type={field.widget} bind:value={formData[field.name]} min={field.min} max={field.max} placeholder={field.placeholder}/>
  {/if}

  {#if field.help && !field.repeated && field.type !== 'object'}
    <small style="color:gray;">{field.help}</small>
  {/if}

  {#if field.error && !field.repeated && field.type !== 'object'}
    <small style="color:red;">{field.error}</small>
  {/if}
</div>
