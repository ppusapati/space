/**
 * Upload Utilities
 * File upload handling with progress tracking
 * @packageDocumentation
 */

import type {
  UploadOptions,
  UploadProgress,
  UploadResponse,
} from '../types/index.js';
import { getConfig } from '../client/transport.js';
import { getAuthProvider } from '../providers.js';

// ============================================================================
// UPLOAD CONFIGURATION
// ============================================================================

/** Default chunk size (5MB) */
const DEFAULT_CHUNK_SIZE = 5 * 1024 * 1024;

/** Upload endpoints */
const UPLOAD_ENDPOINTS = {
  single: '/api/upload',
  chunked: '/api/upload/chunked',
  init: '/api/upload/init',
  complete: '/api/upload/complete',
} as const;

// ============================================================================
// FILE UPLOAD
// ============================================================================

/**
 * Uploads a single file
 */
export async function uploadFile(
  file: File,
  options: UploadOptions = {}
): Promise<UploadResponse> {
  const config = getConfig();
  const auth = getAuthProvider();

  const formData = new FormData();
  formData.append('file', file);

  // Create abort controller for timeout/cancellation
  const controller = new AbortController();
  const signal = options.signal
    ? combineSignals(options.signal, controller.signal)
    : controller.signal;

  // Setup timeout
  let timeoutId: ReturnType<typeof setTimeout> | undefined;
  if (options.timeout) {
    timeoutId = setTimeout(() => controller.abort(), options.timeout);
  }

  try {
    const response = await fetch(`${config.baseUrl}${UPLOAD_ENDPOINTS.single}`, {
      method: 'POST',
      body: formData,
      signal,
      credentials: config.credentials,
      headers: {
        ...config.headers,
        ...options.headers,
        ...(auth.getTokens()?.accessToken && {
          Authorization: `Bearer ${auth.getTokens()?.accessToken}`,
        }),
      },
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    return await response.json();
  } finally {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
  }
}

/**
 * Uploads a file with progress tracking using XMLHttpRequest
 */
export async function uploadFileWithProgress(
  file: File,
  options: UploadOptions = {}
): Promise<UploadResponse> {
  const config = getConfig();
  const auth = getAuthProvider();

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    formData.append('file', file);

    // Setup progress handler
    if (options.onProgress) {
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const progress: UploadProgress = {
            loaded: event.loaded,
            total: event.total,
            percentage: Math.round((event.loaded / event.total) * 100),
          };
          options.onProgress?.(progress);
        }
      });
    }

    // Setup completion handler
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText));
        } catch {
          reject(new Error('Invalid response'));
        }
      } else {
        reject(new Error(`Upload failed: ${xhr.statusText}`));
      }
    });

    // Setup error handler
    xhr.addEventListener('error', () => {
      reject(new Error('Network error during upload'));
    });

    // Setup abort handler
    xhr.addEventListener('abort', () => {
      reject(new Error('Upload cancelled'));
    });

    // Handle external abort signal
    if (options.signal) {
      options.signal.addEventListener('abort', () => {
        xhr.abort();
      });
    }

    // Open and configure request
    xhr.open('POST', `${config.baseUrl}${UPLOAD_ENDPOINTS.single}`);
    xhr.withCredentials = config.credentials === 'include';

    // Set headers
    if (config.headers) {
      for (const [key, value] of Object.entries(config.headers)) {
        xhr.setRequestHeader(key, value);
      }
    }
    if (options.headers) {
      for (const [key, value] of Object.entries(options.headers)) {
        xhr.setRequestHeader(key, value);
      }
    }
    if (auth.getTokens()?.accessToken) {
      xhr.setRequestHeader('Authorization', `Bearer ${auth.getTokens()?.accessToken}`);
    }

    // Setup timeout
    if (options.timeout) {
      xhr.timeout = options.timeout;
    }

    // Send request
    xhr.send(formData);
  });
}

// ============================================================================
// CHUNKED UPLOAD
// ============================================================================

/** Chunked upload state */
interface ChunkedUploadState {
  uploadId: string;
  file: File;
  chunkSize: number;
  totalChunks: number;
  uploadedChunks: number;
  aborted: boolean;
}

/**
 * Uploads a large file in chunks with resumability
 */
export async function uploadChunked(
  file: File,
  options: UploadOptions = {}
): Promise<UploadResponse> {
  const config = getConfig();
  const auth = getAuthProvider();
  const chunkSize = options.chunkSize ?? DEFAULT_CHUNK_SIZE;
  const totalChunks = Math.ceil(file.size / chunkSize);

  // Initialize upload
  const initResponse = await fetch(`${config.baseUrl}${UPLOAD_ENDPOINTS.init}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(auth.getTokens()?.accessToken && {
        Authorization: `Bearer ${auth.getTokens()?.accessToken}`,
      }),
    },
    body: JSON.stringify({
      filename: file.name,
      size: file.size,
      mimeType: file.type,
      totalChunks,
    }),
    signal: options.signal,
    credentials: config.credentials,
  });

  if (!initResponse.ok) {
    throw new Error(`Failed to initialize upload: ${initResponse.statusText}`);
  }

  const { uploadId } = (await initResponse.json()) as { uploadId: string };

  // Upload state
  const state: ChunkedUploadState = {
    uploadId,
    file,
    chunkSize,
    totalChunks,
    uploadedChunks: 0,
    aborted: false,
  };

  // Handle abort
  if (options.signal) {
    options.signal.addEventListener('abort', () => {
      state.aborted = true;
    });
  }

  // Upload chunks
  for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
    if (state.aborted) {
      throw new Error('Upload cancelled');
    }

    const start = chunkIndex * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    await uploadChunk(state, chunkIndex, chunk, options);

    state.uploadedChunks++;

    // Report progress
    if (options.onProgress) {
      options.onProgress({
        loaded: Math.min(end, file.size),
        total: file.size,
        percentage: Math.round(((chunkIndex + 1) / totalChunks) * 100),
      });
    }
  }

  // Complete upload
  const completeResponse = await fetch(
    `${config.baseUrl}${UPLOAD_ENDPOINTS.complete}`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(auth.getTokens()?.accessToken && {
          Authorization: `Bearer ${auth.getTokens()?.accessToken}`,
        }),
      },
      body: JSON.stringify({ uploadId }),
      signal: options.signal,
      credentials: config.credentials,
    }
  );

  if (!completeResponse.ok) {
    throw new Error(`Failed to complete upload: ${completeResponse.statusText}`);
  }

  return completeResponse.json();
}

/**
 * Uploads a single chunk
 */
async function uploadChunk(
  state: ChunkedUploadState,
  chunkIndex: number,
  chunk: Blob,
  options: UploadOptions
): Promise<void> {
  const config = getConfig();
  const auth = getAuthProvider();

  const formData = new FormData();
  formData.append('chunk', chunk);
  formData.append('uploadId', state.uploadId);
  formData.append('chunkIndex', String(chunkIndex));

  const response = await fetch(`${config.baseUrl}${UPLOAD_ENDPOINTS.chunked}`, {
    method: 'POST',
    body: formData,
    headers: {
      ...(auth.getTokens()?.accessToken && {
        Authorization: `Bearer ${auth.getTokens()?.accessToken}`,
      }),
    },
    signal: options.signal,
    credentials: config.credentials,
  });

  if (!response.ok) {
    throw new Error(`Failed to upload chunk ${chunkIndex}: ${response.statusText}`);
  }
}

// ============================================================================
// MULTIPLE FILES UPLOAD
// ============================================================================

/** Multiple upload result */
export interface MultipleUploadResult {
  successful: UploadResponse[];
  failed: Array<{ file: File; error: Error }>;
}

/**
 * Uploads multiple files with parallel processing
 */
export async function uploadMultiple(
  files: File[],
  options: UploadOptions & { concurrency?: number } = {}
): Promise<MultipleUploadResult> {
  const { concurrency = 3, onProgress, ...uploadOptions } = options;

  const result: MultipleUploadResult = {
    successful: [],
    failed: [],
  };

  // Track overall progress
  const totalSize = files.reduce((sum, file) => sum + file.size, 0);
  const fileProgress = new Map<string, number>();

  const updateProgress = () => {
    if (onProgress) {
      let loaded = 0;
      for (const progress of fileProgress.values()) {
        loaded += progress;
      }
      onProgress({
        loaded,
        total: totalSize,
        percentage: Math.round((loaded / totalSize) * 100),
      });
    }
  };

  // Process files in batches
  const queue = [...files];
  const active: Promise<void>[] = [];

  while (queue.length > 0 || active.length > 0) {
    // Start new uploads up to concurrency limit
    while (queue.length > 0 && active.length < concurrency) {
      const file = queue.shift()!;
      const fileId = `${file.name}-${file.size}`;

      const uploadPromise = uploadFileWithProgress(file, {
        ...uploadOptions,
        onProgress: (progress) => {
          fileProgress.set(fileId, progress.loaded);
          updateProgress();
        },
      })
        .then((response) => {
          result.successful.push(response);
        })
        .catch((error) => {
          result.failed.push({
            file,
            error: error instanceof Error ? error : new Error(String(error)),
          });
        })
        .finally(() => {
          const index = active.indexOf(uploadPromise);
          if (index > -1) {
            active.splice(index, 1);
          }
        });

      active.push(uploadPromise);
    }

    // Wait for at least one upload to complete
    if (active.length > 0) {
      await Promise.race(active);
    }
  }

  return result;
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Combines multiple abort signals
 */
function combineSignals(
  ...signals: AbortSignal[]
): AbortSignal {
  const controller = new AbortController();

  for (const signal of signals) {
    if (signal.aborted) {
      controller.abort();
      break;
    }
    signal.addEventListener('abort', () => controller.abort());
  }

  return controller.signal;
}

/**
 * Validates file before upload
 */
export function validateFile(
  file: File,
  options: {
    maxSize?: number;
    allowedTypes?: string[];
    allowedExtensions?: string[];
  } = {}
): { valid: boolean; error?: string } {
  const { maxSize, allowedTypes, allowedExtensions } = options;

  // Check size
  if (maxSize && file.size > maxSize) {
    return {
      valid: false,
      error: `File size exceeds maximum allowed (${formatBytes(maxSize)})`,
    };
  }

  // Check MIME type
  if (allowedTypes && allowedTypes.length > 0) {
    const isAllowed = allowedTypes.some((type) => {
      if (type.endsWith('/*')) {
        return file.type.startsWith(type.slice(0, -1));
      }
      return file.type === type;
    });

    if (!isAllowed) {
      return {
        valid: false,
        error: `File type "${file.type}" is not allowed`,
      };
    }
  }

  // Check extension
  if (allowedExtensions && allowedExtensions.length > 0) {
    const extension = file.name.split('.').pop()?.toLowerCase();
    if (!extension || !allowedExtensions.includes(`.${extension}`)) {
      return {
        valid: false,
        error: `File extension ".${extension}" is not allowed`,
      };
    }
  }

  return { valid: true };
}

/**
 * Formats bytes to human readable string
 */
function formatBytes(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }

  return `${value.toFixed(1)} ${units[unitIndex]}`;
}
