/**
 * API Types
 * Type definitions for API client
 * @packageDocumentation
 */

// ============================================================================
// API CONFIGURATION
// ============================================================================

/** API client configuration */
export interface ApiConfig {
  /** Base URL for the API */
  baseUrl: string;

  /** Default timeout in milliseconds */
  timeout?: number;

  /** Whether to include credentials */
  credentials?: RequestCredentials;

  /** Custom headers to include in all requests */
  headers?: Record<string, string>;

  /** Enable request/response logging */
  debug?: boolean;

  /** Retry configuration */
  retry?: RetryConfig;

  /** Cache configuration */
  cache?: CacheConfig;
}

/** Retry configuration */
export interface RetryConfig {
  /** Maximum number of retries */
  maxRetries: number;

  /** Initial delay between retries in ms */
  initialDelay: number;

  /** Maximum delay between retries in ms */
  maxDelay: number;

  /** Backoff multiplier */
  backoffMultiplier: number;

  /** HTTP status codes that should trigger a retry */
  retryableStatuses: number[];

  /** Whether to retry on network errors */
  retryOnNetworkError: boolean;
}

/** Cache configuration */
export interface CacheConfig {
  /** Enable caching */
  enabled: boolean;

  /** Default TTL in seconds */
  defaultTtl: number;

  /** Maximum cache size */
  maxSize: number;

  /** Cache storage type */
  storage: 'memory' | 'localStorage' | 'sessionStorage';
}

// ============================================================================
// API RESPONSE TYPES
// ============================================================================

/** Standard API response wrapper */
export interface ApiResponse<T> {
  data: T;
  meta?: ResponseMeta;
}

/** Response metadata */
export interface ResponseMeta {
  /** Request ID for tracing */
  requestId?: string;

  /** Server timestamp */
  timestamp?: string;

  /** Response duration in ms */
  duration?: number;
}

/** Paginated response */
export interface PaginatedResponse<T> {
  items: T[];
  total?: number;
  pagination: PaginationMeta;
}

/** Pagination metadata */
export interface PaginationMeta {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
}

/** List response with optional aggregations */
export interface ListResponse<T, TAggregates = Record<string, unknown>> {
  items: T[];
  pagination: PaginationMeta;
  aggregates?: TAggregates;
  filters?: AppliedFilter[];
}

/** Applied filter info */
export interface AppliedFilter {
  field: string;
  operator: string;
  value: unknown;
}

// ============================================================================
// API ERROR TYPES
// ============================================================================

/** API error */
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  field?: string;
  retryable?: boolean;
  statusCode?: number;
  requestId?: string;
}

/** Validation error */
export interface ValidationError {
  field: string;
  code: string;
  message: string;
  constraints?: Record<string, string>;
}

/** Validation errors response */
export interface ValidationErrorResponse {
  code: 'VALIDATION_ERROR';
  message: string;
  errors: ValidationError[];
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

/** Base request options */
export interface RequestOptions {
  /** Custom headers for this request */
  headers?: Record<string, string>;

  /** Request timeout override */
  timeout?: number;

  /** Abort signal */
  signal?: AbortSignal;

  /** Skip cache for this request */
  skipCache?: boolean;

  /** Cache TTL override */
  cacheTtl?: number;

  /** Skip authentication */
  skipAuth?: boolean;

  /** Custom retry config for this request */
  retry?: Partial<RetryConfig>;
}

/** List request parameters */
export interface ListParams<TFilters = Record<string, unknown>> {
  /** Page number (1-indexed) */
  page?: number;

  /** Items per page */
  pageSize?: number;

  /** Sort field */
  sortBy?: string;

  /** Sort direction */
  sortOrder?: 'asc' | 'desc';

  /** Search query */
  search?: string;

  /** Filters */
  filters?: TFilters;

  /** Include soft-deleted items */
  includeDeleted?: boolean;

  /** Fields to include in response */
  fields?: string[];

  /** Related entities to include */
  include?: string[];
}

/** Entity request parameters */
export interface EntityParams {
  /** Fields to include */
  fields?: string[];

  /** Related entities to include */
  include?: string[];
}

/** Create request */
export interface CreateRequest<T> {
  data: Omit<T, 'id' | 'createdAt' | 'updatedAt'>;
}

/** Update request */
export interface UpdateRequest<T> {
  id: string | number;
  data: Partial<Omit<T, 'id' | 'createdAt' | 'updatedAt'>>;
}

/** Bulk operation request */
export interface BulkRequest<T> {
  items: T[];
}

/** Bulk operation response */
export interface BulkResponse<T> {
  successful: T[];
  failed: Array<{
    item: T;
    error: ApiError;
  }>;
}

// ============================================================================
// STREAMING TYPES
// ============================================================================

/** Stream event */
export interface StreamEvent<T> {
  type: 'data' | 'error' | 'complete';
  data?: T;
  error?: ApiError;
}

/** Stream options */
export interface StreamOptions extends RequestOptions {
  /** Buffer size for backpressure */
  bufferSize?: number;

  /** Auto-reconnect on disconnect */
  autoReconnect?: boolean;

  /** Reconnect delay in ms */
  reconnectDelay?: number;
}

// ============================================================================
// UPLOAD TYPES
// ============================================================================

/** Upload progress */
export interface UploadProgress {
  loaded: number;
  total: number;
  percentage: number;
}

/** Upload options */
export interface UploadOptions extends RequestOptions {
  /** Progress callback */
  onProgress?: (progress: UploadProgress) => void;

  /** Chunk size for chunked uploads */
  chunkSize?: number;

  /** Enable resumable uploads */
  resumable?: boolean;
}

/** Upload response */
export interface UploadResponse {
  id: string;
  filename: string;
  mimeType: string;
  size: number;
  url: string;
  thumbnailUrl?: string;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// INTERCEPTOR TYPES
// ============================================================================

/** Request interceptor */
export interface RequestInterceptor {
  /** Unique identifier */
  id: string;

  /** Priority (higher = runs first) */
  priority?: number;

  /** Intercept function */
  intercept: (config: InterceptedRequest) => InterceptedRequest | Promise<InterceptedRequest>;
}

/** Response interceptor */
export interface ResponseInterceptor {
  /** Unique identifier */
  id: string;

  /** Priority (higher = runs first) */
  priority?: number;

  /** Success handler */
  onSuccess?: <T>(response: InterceptedResponse<T>) => InterceptedResponse<T> | Promise<InterceptedResponse<T>>;

  /** Error handler */
  onError?: (error: ApiError) => ApiError | Promise<ApiError> | null;
}

/** Intercepted request */
export interface InterceptedRequest {
  url: string;
  method: string;
  headers: Record<string, string>;
  body?: unknown;
  options: RequestOptions;
}

/** Intercepted response */
export interface InterceptedResponse<T> {
  data: T;
  status: number;
  headers: Record<string, string>;
  duration: number;
  requestId?: string;
}

// ============================================================================
// FORM SERVICE TYPES (generated from formservice.proto / formbuilder.proto)
// ============================================================================

export * from './formservice.types.js';
