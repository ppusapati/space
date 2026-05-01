/**
 * Upload Utilities
 * File upload handling with progress tracking
 * @packageDocumentation
 */
import type { UploadOptions, UploadResponse } from '../types/index.js';
/**
 * Uploads a single file
 */
export declare function uploadFile(file: File, options?: UploadOptions): Promise<UploadResponse>;
/**
 * Uploads a file with progress tracking using XMLHttpRequest
 */
export declare function uploadFileWithProgress(file: File, options?: UploadOptions): Promise<UploadResponse>;
/**
 * Uploads a large file in chunks with resumability
 */
export declare function uploadChunked(file: File, options?: UploadOptions): Promise<UploadResponse>;
/** Multiple upload result */
export interface MultipleUploadResult {
    successful: UploadResponse[];
    failed: Array<{
        file: File;
        error: Error;
    }>;
}
/**
 * Uploads multiple files with parallel processing
 */
export declare function uploadMultiple(files: File[], options?: UploadOptions & {
    concurrency?: number;
}): Promise<MultipleUploadResult>;
/**
 * Validates file before upload
 */
export declare function validateFile(file: File, options?: {
    maxSize?: number;
    allowedTypes?: string[];
    allowedExtensions?: string[];
}): {
    valid: boolean;
    error?: string;
};
//# sourceMappingURL=upload.d.ts.map