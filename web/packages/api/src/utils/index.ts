/**
 * Utilities Module - Export all utility functions
 * @packageDocumentation
 */

// Query utilities
export {
  QueryBuilder,
  createQueryBuilder,
  calculatePagination,
  getOffset,
  getPage,
  generatePageNumbers,
  serializeFilters,
  deserializeFilters,
  type QueryBuilderOptions,
} from './query.js';

// Cache utilities
export {
  ApiCache,
  generateCacheKey,
  getApiCache,
  resetApiCache,
} from './cache.js';

// Upload utilities
export {
  uploadFile,
  uploadFileWithProgress,
  uploadChunked,
  uploadMultiple,
  validateFile,
  type MultipleUploadResult,
} from './upload.js';
