/**
 * Logging Interceptor
 * Logs API requests and responses for debugging
 * @packageDocumentation
 */
import type { Interceptor } from '@connectrpc/connect';
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
/**
 * Creates a logging interceptor
 */
export declare function createLoggingInterceptor(options?: LoggingOptions): Interceptor;
/**
 * Creates a structured logger for production use
 */
export declare function createStructuredLogger(sendFn: (entries: LogEntry[]) => Promise<void>, options?: {
    batchSize?: number;
    flushInterval?: number;
}): (entry: LogEntry) => void;
//# sourceMappingURL=logging.interceptor.d.ts.map