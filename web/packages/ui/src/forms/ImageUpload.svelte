<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface ImageUploadProps {
    value?: string;
    label?: string;
    helperText?: string;
    errorText?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: Size;
    state?: ValidationState;
    name?: string;
    id?: string;
    accept?: string;
    maxSize?: number; // in bytes
    multiple?: boolean;
    preview?: boolean;
  }

  // Props
  export let value: string | string[] = '';
  export let label: string = '';
  export let helperText: string = 'PNG, JPG, GIF up to 5MB';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('image-upload');
  export let accept: string = 'image/*';
  export let maxSize: number = 5 * 1024 * 1024; // 5MB default
  export let multiple: boolean = false;
  export let preview: boolean = true;

  let className: string = '';
  export { className as class };

  let isDragging = false;
  let uploadError = '';

  const dispatch = createEventDispatcher<{
    change: string | string[];
    upload: { files: File[] };
    error: string;
  }>();

  const sizeClasses = {
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  };

  function getImagePreviewUrl(file: File): Promise<string> {
    return new Promise((resolve) => {
      const reader = new FileReader();
      reader.onload = (e) => resolve(e.target?.result as string);
      reader.readAsDataURL(file);
    });
  }

  async function handleFiles(files: FileList | null) {
    if (!files || files.length === 0) return;

    uploadError = '';
    const validFiles: File[] = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i]!;

      // Validate file type
      if (!file.type.startsWith('image/')) {
        uploadError = 'Please upload image files only';
        dispatch('error', uploadError);
        return;
      }

      // Validate file size
      if (file.size > maxSize) {
        uploadError = `File size exceeds ${maxSize / 1024 / 1024}MB limit`;
        dispatch('error', uploadError);
        return;
      }

      validFiles.push(file);

      // Generate preview if enabled
      if (preview) {
        const previewUrl = await getImagePreviewUrl(file);
        if (multiple) {
          if (Array.isArray(value)) {
            value = [...value, previewUrl];
          } else {
            value = [previewUrl];
          }
        } else {
          value = previewUrl;
        }
      }
    }

    if (validFiles.length > 0) {
      dispatch('change', value);
      dispatch('upload', { files: validFiles });
    }
  }

  function handleFileInput(e: Event) {
    const target = e.target as HTMLInputElement;
    handleFiles(target.files);
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    handleFiles(e.dataTransfer?.files || null);
  }

  function removeImage(index: number) {
    if (Array.isArray(value)) {
      value = value.filter((_, i) => i !== index);
    } else {
      value = '';
    }
    dispatch('change', value);
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label class="block text-sm font-medium text-neutral-700 mb-2">
      {label}
      {#if required}
        <span class="text-red-500 ml-1">*</span>
      {/if}
    </label>
  {/if}

  <div
    on:dragover={handleDragOver}
    on:dragleave={handleDragLeave}
    on:drop={handleDrop}
    class={cn(
      'border-2 border-dashed rounded-lg transition-colors cursor-pointer',
      sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
      isDragging ? 'border-primary-500 bg-primary-50' : 'border-neutral-300 bg-neutral-50',
      disabled && 'opacity-50 cursor-not-allowed'
    )}
  >
    <input
      {id}
      {name}
      {accept}
      {multiple}
      {disabled}
      type="file"
      on:change={handleFileInput}
      class="hidden"
    />

    <label for={id} class="flex flex-col items-center justify-center cursor-pointer">
      <Icon name="image" size="lg" class="text-neutral-400 mb-2" />
      <p class="text-sm font-medium text-neutral-700">Drop images here or click to upload</p>
      <p class="text-xs text-neutral-500 mt-1">{helperText}</p>
    </label>
  </div>

  {#if preview && value}
    <div class="mt-4">
      {#if Array.isArray(value)}
        <div class="grid grid-cols-3 gap-4">
          {#each value as image, idx}
            <div class="relative group">
              <img
                src={image}
                alt="Preview {idx + 1}"
                class="w-full h-24 object-cover rounded-lg border border-neutral-200"
              />
              <button
                type="button"
                on:click={() => removeImage(idx)}
                disabled={disabled || readonly}
                class="absolute top-1 right-1 bg-red-500 text-white rounded-full p-1 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          {/each}
        </div>
      {:else if value}
        <div class="relative inline-block">
          <img
            src={value}
            alt="Preview"
            class="w-32 h-32 object-cover rounded-lg border border-neutral-200"
          />
          <button
            type="button"
            on:click={() => (value = '')}
            disabled={disabled || readonly}
            class="absolute top-1 right-1 bg-red-500 text-white rounded-full p-1"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>
      {/if}
    </div>
  {/if}

  {#if uploadError}
    <p class="mt-2 text-sm text-red-500">{uploadError}</p>
  {:else if errorText}
    <p class="mt-2 text-sm text-red-500">{errorText}</p>
  {:else if helperText && !preview}
    <p class="mt-2 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
