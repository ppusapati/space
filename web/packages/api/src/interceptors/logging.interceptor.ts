/**
 * Logging Interceptor
 * Logs API requests and responses for debugging
 * @packageDocumentation
 */

import type { Interceptor } from '@connectrpc/connect';

// ============================================================================
// LOGGING CONFIGURATION
// ============================================================================

/** Log levels */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'none';

/** Logging options */
export interface LoggingOptions {
  /** Minimum log level */
  level?: LogLevel;

  /** Whether to log request bodies */
  logRequestBody?: boolean;

  /** Whether to log response bodies */
  logResponseBody?: boolean;

  /** Whether to log headers */
  logHeaders?: boolean;

  /** Whether to include timing information */
  logTiming?: boolean;

  /** Custom logger function */
  logger?: (entry: LogEntry) => void;

  /** Paths to exclude from logging */
  excludePaths?: string[];

  /** Maximum body length to log */
  maxBodyLength?: number;
}

/** Log entry */
export interface LogEntry {
  timestamp: Date;
  level: LogLevel;
  method: string;
  url: string;
  requestId?: string;
  duration?: number;
  status?: 'success' | 'error';
  error?: {
    code: string;
    message: string;
  };
  request?: {
    headers?: Record<string, string>;
    body?: unknown;
  };
  response?: {
    headers?: Record<string, string>;
    body?: unknown;
  };
}

// ============================================================================
// LOGGING INTERCEPTOR
// ============================================================================

const LOG_LEVEL_PRIORITY: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
  none: 4,
};

/**
 * Creates a logging interceptor
 */
export function createLoggingInterceptor(
  options: LoggingOptions = {}
): Interceptor {
  const {
    level = 'info',
    logRequestBody = false,
    logResponseBody = false,
    logHeaders = false,
    logTiming = true,
    logger = defaultLogger,
    excludePaths = [],
    maxBodyLength = 1000,
  } = options;

  return (next) => async (req) => {
    // Skip logging for excluded paths
    if (shouldExclude(req.url, excludePaths)) {
      return next(req);
    }

    const startTime = performance.now();
    const requestId = crypto.randomUUID();

    // Build request log entry
    const entry: LogEntry = {
      timestamp: new Date(),
      level: 'info',
      method: req.method.name,
      url: req.url,
      requestId,
    };

    // Log request details
    if (logHeaders) {
      entry.request = {
        headers: Object.fromEntries(req.header),
      };
    }

    if (logRequestBody) {
      entry.request = {
        ...entry.request,
        body: truncateBody(req.message, maxBodyLength),
      };
    }

    try {
      const response = await next(req);

      // Calculate duration
      if (logTiming) {
        entry.duration = performance.now() - startTime;
      }

      // Log response details
      entry.status = 'success';

      if (logResponseBody) {
        entry.response = {
          body: truncateBody(response.message, maxBodyLength),
        };
      }

      // Log the entry
      if (shouldLog(level, 'info')) {
        logger(entry);
      }

      return response;
    } catch (error) {
      // Calculate duration
      if (logTiming) {
        entry.duration = performance.now() - startTime;
      }

      // Log error details
      entry.level = 'error';
      entry.status = 'error';
      entry.error = {
        code: getErrorCode(error),
        message: getErrorMessage(error),
      };

      // Log the entry
      if (shouldLog(level, 'error')) {
        logger(entry);
      }

      throw error;
    }
  };
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Checks if a path should be excluded from logging
 */
function shouldExclude(url: string, excludePaths: string[]): boolean {
  const path = new URL(url).pathname;
  return excludePaths.some((pattern) => path.includes(pattern));
}

/**
 * Checks if a message should be logged based on level
 */
function shouldLog(minLevel: LogLevel, messageLevel: LogLevel): boolean {
  return LOG_LEVEL_PRIORITY[messageLevel] >= LOG_LEVEL_PRIORITY[minLevel];
}

/**
 * Truncates body for logging
 */
function truncateBody(body: unknown, maxLength: number): unknown {
  if (!body) return body;

  try {
    const str = JSON.stringify(body);
    if (str.length > maxLength) {
      return {
        _truncated: true,
        _length: str.length,
        _preview: str.slice(0, maxLength) + '...',
      };
    }
    return body;
  } catch {
    return { _type: typeof body, _error: 'Could not serialize' };
  }
}

/**
 * Gets error code from error object
 */
function getErrorCode(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'code' in error) {
    return String((error as { code: unknown }).code);
  }
  return 'UNKNOWN';
}

/**
 * Gets error message from error object
 */
function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'object' && error !== null && 'message' in error) {
    return String((error as { message: unknown }).message);
  }
  return 'Unknown error';
}

/**
 * Default logger implementation
 */
function defaultLogger(entry: LogEntry): void {
  const prefix = `[API] ${entry.method}`;
  const timing = entry.duration ? ` (${entry.duration.toFixed(2)}ms)` : '';

  if (entry.status === 'error') {
    console.error(`${prefix} FAILED${timing}`, {
      requestId: entry.requestId,
      error: entry.error,
      request: entry.request,
    });
  } else {
    console.log(`${prefix} OK${timing}`, {
      requestId: entry.requestId,
      request: entry.request,
      response: entry.response,
    });
  }
}

// ============================================================================
// STRUCTURED LOGGER
// ============================================================================

/**
 * Creates a structured logger for production use
 */
export function createStructuredLogger(
  sendFn: (entries: LogEntry[]) => Promise<void>,
  options: {
    batchSize?: number;
    flushInterval?: number;
  } = {}
): (entry: LogEntry) => void {
  const { batchSize = 10, flushInterval = 5000 } = options;

  let buffer: LogEntry[] = [];
  let flushTimeout: ReturnType<typeof setTimeout> | null = null;

  const flush = async () => {
    if (buffer.length === 0) return;

    const entries = buffer;
    buffer = [];

    try {
      await sendFn(entries);
    } catch (error) {
      console.error('Failed to send log entries:', error);
      // Re-add entries to buffer on failure
      buffer = [...entries, ...buffer];
    }
  };

  const scheduleFlush = () => {
    if (flushTimeout) return;
    flushTimeout = setTimeout(() => {
      flushTimeout = null;
      flush();
    }, flushInterval);
  };

  return (entry: LogEntry) => {
    buffer.push(entry);

    if (buffer.length >= batchSize) {
      flush();
    } else {
      scheduleFlush();
    }
  };
}
