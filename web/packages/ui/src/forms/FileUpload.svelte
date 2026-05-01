<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type UploadedFile,
    fileUploadSizeClasses,
    fileUploadContainerClasses,
    fileListClasses,
    previewClasses,
    formatFileSize,
    validateFile,
    createUploadedFile,
    createPreview,
  } from './fileupload.types';
  import type { Size, ValidationState } from '../types';

  // Props
  export let files: UploadedFile[] = [];
  export let accept: string = '';
  export let maxSize: number | undefined = undefined;
  export let maxFiles: number | undefined = undefined;
  export let multiple: boolean = false;
  export let dragDrop: boolean = true;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let uploadText: string = 'Drop files here or click to upload';
  export let showPreview: boolean = true;
  export let showProgress: boolean = true;
  export let disabled: boolean = false;
  export let required: boolean = false;
  export let name: string = '';
  export let id: string = uid('fileupload');
  export let testId: string = '';
  export let fullWidth: boolean = true;

  let className: string = '';
  export { className as class };

  let inputRef: HTMLInputElement;
  let isDragging = false;

  const dispatch = createEventDispatcher<{
    select: { files: UploadedFile[] };
    remove: { file: UploadedFile };
    error: { file: File; error: string };
  }>();

  // Computed classes
  $: sizeConfig = fileUploadSizeClasses[size];

  $: containerClasses = cn(
    fileUploadContainerClasses.base,
    sizeConfig.container,
    isDragging
      ? fileUploadContainerClasses.dragging
      : state === 'invalid' || errorText
      ? fileUploadContainerClasses.error
      : fileUploadContainerClasses.default,
    disabled && fileUploadContainerClasses.disabled,
    fullWidth ? 'w-full' : '',
    className
  );

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;

  // Handlers
  async function handleFiles(fileList: FileList | null) {
    if (!fileList || disabled) return;

    const newFiles: UploadedFile[] = [];
    const fileArray = Array.from(fileList);

    // Check max files limit
    if (maxFiles && files.length + fileArray.length > maxFiles) {
      dispatch('error', {
        file: fileArray[0]!,
        error: `Maximum ${maxFiles} files allowed`
      });
      return;
    }

    for (const file of fileArray) {
      const validation = validateFile(file, accept, maxSize);

      if (!validation.valid) {
        dispatch('error', { file, error: validation.error || 'Invalid file' });
        continue;
      }

      const uploadedFile = createUploadedFile(file);

      // Create preview for images
      if (showPreview && file.type.startsWith('image/')) {
        uploadedFile.preview = await createPreview(file);
      }

      newFiles.push(uploadedFile);
    }

    if (newFiles.length > 0) {
      files = multiple ? [...files, ...newFiles] : newFiles;
      dispatch('select', { files: newFiles });
    }
  }

  function handleInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    handleFiles(target.files);
    // Reset input value to allow selecting same file again
    target.value = '';
  }

  function handleDragOver(event: DragEvent) {
    if (disabled || !dragDrop) return;
    event.preventDefault();
    isDragging = true;
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault();
    isDragging = false;
  }

  function handleDrop(event: DragEvent) {
    if (disabled || !dragDrop) return;
    event.preventDefault();
    isDragging = false;
    handleFiles(event.dataTransfer?.files || null);
  }

  function handleClick() {
    if (disabled) return;
    inputRef?.click();
  }

  function removeFile(file: UploadedFile) {
    files = files.filter(f => f.id !== file.id);
    dispatch('remove', { file });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleClick();
    }
  }
</script>

<div class={cn('w-full', fullWidth && 'max-w-full')}>
  {#if label}
    <label for={id} class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-semantic-error-500 ml-0.5" aria-hidden="true">*</span>
      {/if}
    </label>
  {/if}

  <!-- Hidden file input -->
  <input
    bind:this={inputRef}
    type="file"
    {id}
    {name}
    {accept}
    {multiple}
    {disabled}
    {required}
    class="sr-only"
    data-testid={testId || undefined}
    on:change={handleInputChange}
  />

  <!-- Drop zone -->
  <div
    class={containerClasses}
    role="button"
    tabindex={disabled ? -1 : 0}
    aria-disabled={disabled}
    on:click={handleClick}
    on:keydown={handleKeydown}
    on:dragover={handleDragOver}
    on:dragleave={handleDragLeave}
    on:drop={handleDrop}
  >
    <!-- Upload icon -->
    <svg
      class={cn(sizeConfig.icon, 'text-neutral-400 mb-2')}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
      />
    </svg>

    <p class={cn(sizeConfig.text, 'text-neutral-600 font-medium')}>
      {uploadText}
    </p>

    {#if accept}
      <p class="text-xs text-neutral-400 mt-1">
        Accepted: {accept}
      </p>
    {/if}

    {#if maxSize}
      <p class="text-xs text-neutral-400">
        Max size: {formatFileSize(maxSize)}
      </p>
    {/if}
  </div>

  <!-- File list -->
  {#if files.length > 0}
    <div class={fileListClasses.container}>
      {#each files as file (file.id)}
        <div
          class={cn(
            fileListClasses.item,
            file.status === 'error' && fileListClasses.itemError
          )}
        >
          <!-- Preview / Icon -->
          {#if showPreview && file.preview}
            <div class={previewClasses.container}>
              <img src={file.preview} alt={file.name} class={previewClasses.image} />
            </div>
          {:else}
            <svg class={fileListClasses.icon} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          {/if}

          <!-- File info -->
          <div class={fileListClasses.info}>
            <p class={fileListClasses.name}>{file.name}</p>
            <p class={fileListClasses.size}>
              {formatFileSize(file.size)}
              {#if file.status === 'error' && file.error}
                <span class="text-semantic-error-500"> - {file.error}</span>
              {/if}
            </p>

            <!-- Progress bar -->
            {#if showProgress && file.status === 'uploading'}
              <div class={fileListClasses.progress}>
                <div
                  class={fileListClasses.progressBar}
                  style="width: {file.progress}%"
                ></div>
              </div>
            {/if}
          </div>

          <!-- Actions -->
          <div class={fileListClasses.actions}>
            {#if file.status === 'success'}
              <svg class="w-5 h-5 text-semantic-success-500" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            {/if}

            <button
              type="button"
              class={fileListClasses.removeBtn}
              on:click={() => removeFile(file)}
              aria-label="Remove file"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if displayedHelperText}
    <p
      id="{id}-helper"
      class={cn(
        'mt-2 text-sm',
        errorText ? 'text-semantic-error-600' : 'text-neutral-500'
      )}
    >
      {displayedHelperText}
    </p>
  {/if}
</div>
