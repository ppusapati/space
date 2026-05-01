/**
 * Error Interceptor
 * Standardizes error handling and notifications
 * @packageDocumentation
 */
import { toastStore } from '@samavāya/stores';
import { createApiError } from '../client/client.js';
// ============================================================================
// ERROR INTERCEPTOR
// ============================================================================
/** Error code to user-friendly message mapping */
const ERROR_MESSAGES = {
    CANCELLED: 'Request was cancelled',
    UNKNOWN: 'An unexpected error occurred',
    INVALID_ARGUMENT: 'Invalid input provided',
    DEADLINE_EXCEEDED: 'Request timed out',
    NOT_FOUND: 'Resource not found',
    ALREADY_EXISTS: 'Resource already exists',
    PERMISSION_DENIED: 'You do not have permission to perform this action',
    RESOURCE_EXHAUSTED: 'Too many requests. Please try again later',
    FAILED_PRECONDITION: 'Operation cannot be performed in current state',
    ABORTED: 'Operation was aborted',
    OUT_OF_RANGE: 'Value is out of valid range',
    UNIMPLEMENTED: 'This feature is not yet available',
    INTERNAL: 'An internal error occurred',
    UNAVAILABLE: 'Service is temporarily unavailable',
    DATA_LOSS: 'Data integrity error',
    UNAUTHENTICATED: 'Please log in to continue',
};
/** Error codes that should show a toast notification */
const NOTIFY_ERRORS = [
    'PERMISSION_DENIED',
    'NOT_FOUND',
    'ALREADY_EXISTS',
    'INTERNAL',
    'UNAVAILABLE',
];
/** Error codes that are transient and shouldn't be shown */
const SILENT_ERRORS = [
    'CANCELLED',
    'DEADLINE_EXCEEDED',
];
/**
 * Creates an error interceptor
 */
export function createErrorInterceptor(options = {}) {
    const { showToasts = true, logErrors = true, onError, suppressCodes = [], } = options;
    return (next) => async (req) => {
        try {
            return await next(req);
        }
        catch (error) {
            const apiError = normalizeError(error);
            // Log error if enabled
            if (logErrors) {
                console.error('[API Error]', {
                    code: apiError.code,
                    message: apiError.message,
                    method: req.method.name,
                    details: apiError.details,
                });
            }
            // Show toast if enabled and not suppressed
            if (showToasts &&
                !SILENT_ERRORS.includes(apiError.code) &&
                !suppressCodes.includes(apiError.code) &&
                NOTIFY_ERRORS.includes(apiError.code)) {
                toastStore.error(apiError.message, {
                    title: getErrorTitle(apiError.code),
                    duration: 5000,
                });
            }
            // Call custom error handler
            if (onError) {
                onError(apiError);
            }
            throw apiError;
        }
    };
}
// ============================================================================
// HELPER FUNCTIONS
// ============================================================================
/**
 * Normalizes various error types to ApiError
 */
function normalizeError(error) {
    // ConnectRPC error (check first as it has name === 'ConnectError')
    if (isConnectError(error)) {
        return {
            code: String(error.code),
            message: error.message || getDefaultMessage(String(error.code)),
            details: error.metadata ? Object.fromEntries(error.metadata) : undefined,
            retryable: isRetryable(String(error.code)),
        };
    }
    // Already an ApiError
    if (isApiError(error)) {
        return error;
    }
    // Standard Error
    if (error instanceof Error) {
        return createApiError('UNKNOWN', error.message);
    }
    // Unknown error
    return createApiError('UNKNOWN', 'An unexpected error occurred');
}
/**
 * Checks if error is an ApiError
 */
function isApiError(error) {
    return (typeof error === 'object' &&
        error !== null &&
        'code' in error &&
        'message' in error &&
        typeof error.code === 'string' &&
        typeof error.message === 'string');
}
/**
 * Checks if error is a ConnectRPC error
 */
function isConnectError(error) {
    return (typeof error === 'object' &&
        error !== null &&
        'code' in error &&
        'name' in error &&
        error.name === 'ConnectError');
}
/**
 * Gets the default message for an error code
 */
function getDefaultMessage(code) {
    return ERROR_MESSAGES[code] || ERROR_MESSAGES['UNKNOWN'] || 'An error occurred';
}
/**
 * Gets a user-friendly title for an error
 */
function getErrorTitle(code) {
    switch (code) {
        case 'PERMISSION_DENIED':
            return 'Access Denied';
        case 'NOT_FOUND':
            return 'Not Found';
        case 'ALREADY_EXISTS':
            return 'Duplicate';
        case 'INTERNAL':
            return 'Server Error';
        case 'UNAVAILABLE':
            return 'Service Unavailable';
        default:
            return 'Error';
    }
}
/**
 * Checks if an error code indicates a retryable error
 */
function isRetryable(code) {
    return [
        'UNAVAILABLE',
        'DEADLINE_EXCEEDED',
        'RESOURCE_EXHAUSTED',
        'ABORTED',
        'INTERNAL',
    ].includes(code);
}
//# sourceMappingURL=error.interceptor.js.map