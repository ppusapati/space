# API Client Architecture - @chetana/api

**Purpose:** Unified HTTP client for both **ConnectRPC (internal)** and **REST API (3rd party integration)**

---

## Overview

The API client provides a single, flexible interface for:
- ✅ **Internal Communication:** ConnectRPC calls to Chetana backend
- ✅ **External Integration:** REST API calls to 3rd party services
- ✅ **Cross-Cutting Concerns:** Authentication, error handling, retry logic, logging

---

## Architecture

```
@chetana/api
├── client/           (HTTP transport layer)
│   ├── client.ts     (ApiClient class for ConnectRPC)
│   ├── transport.ts  (HTTP transport setup)
│   └── index.ts      (Exports)
│
├── interceptors/     (Request/Response middleware)
│   ├── auth.interceptor.ts       (JWT token injection)
│   ├── tenant.interceptor.ts     (Multi-tenant isolation)
│   ├── error.interceptor.ts      (Error handling & mapping)
│   ├── retry.interceptor.ts      (Retry with exponential backoff)
│   ├── logging.interceptor.ts    (Request/response logging)
│   └── index.ts                  (Exports)
│
├── types/            (Type definitions)
│   ├── index.ts
│   └── [type files]
│
├── utils/            (Helper functions)
│   ├── index.ts
│   └── [utility files]
│
├── index.ts          (Main exports + initialization)
└── README.md         (This file)
```

---

## Use Cases

### Use Case 1: ConnectRPC (Internal - Recommended)

**For:** Communication with Chetana backend

```typescript
import { ApiClient } from '@chetana/api';
import { UserService } from '@chetana/backend'; // Generated from proto

const apiClient = new ApiClient();
const userService = apiClient.getService(UserService);

// Call internal RPC methods
const user = await userService.getUser({ id: '123' });
```

**Why ConnectRPC?**
- ✅ Type-safe (generated from protobuf)
- ✅ Efficient binary protocol
- ✅ Built-in streaming support
- ✅ Automatic payload serialization
- ✅ Less bandwidth
- ✅ Better performance

### Use Case 2: HTTP REST (3rd Party Integration)

**For:** Integration with external services (payment, email, etc.)

```typescript
import { createTransport, addInterceptor } from '@chetana/api';

// Create REST-specific transport
const restTransport = createTransport({
  baseUrl: 'https://api.payment-provider.com',
  protocol: 'rest', // or 'grpc-web'
});

// Make REST calls
const response = await fetch('https://api.payment-provider.com/charges', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(chargeData),
});
```

---

## Interceptors & Their Purposes

### 1. **Logging Interceptor** 🔍
Logs all requests and responses for debugging and monitoring.

```typescript
// Request
[14:30:45] POST /users/create
  Headers: { Authorization: "Bearer ...", Content-Type: "application/json" }
  Body: { name: "John", email: "john@example.com" }

// Response
[14:30:46] ✓ 201 Created
  Body: { id: "user_123", name: "John" }
  Duration: 1.2s
```

**Order:** FIRST (to capture original request)

### 2. **Retry Interceptor** ♻️
Automatically retries failed requests with exponential backoff.

```typescript
// Retry policy:
// - Max retries: 3
// - Initial delay: 1s
// - Backoff: exponential (1s → 2s → 4s)
// - Retry on: 429 (rate limit), 503 (service unavailable), network errors

Request 1: FAIL (network error)
  ↓ wait 1s
Request 2: FAIL (503 Service Unavailable)
  ↓ wait 2s
Request 3: SUCCESS ✓
```

**Order:** SECOND (before error handling)

### 3. **Error Interceptor** ❌
Maps different error types to consistent error objects.

```typescript
// Raw HTTP Error:
{
  status: 401,
  statusText: "Unauthorized",
  data: { message: "Invalid token" }
}

// Transformed Error:
{
  code: 'UNAUTHORIZED',
  message: 'Invalid token',
  status: 401,
  timestamp: '2026-02-25T14:30:46Z',
  requestId: 'req_xyz',
  details: { reason: 'expired_token' }
}
```

**Order:** THIRD (after retry, before auth/tenant)

### 4. **Tenant Interceptor** 🏢
Injects tenant ID into requests for multi-tenant isolation.

```typescript
// Adds to every request:
{
  headers: {
    'X-Tenant-ID': 'tenant_abc123'
  }
}

// Backend uses this to:
// - Filter data by tenant
// - Apply RLS (Row Level Security) policies
// - Ensure data isolation
```

**Order:** FOURTH (before auth)

### 5. **Auth Interceptor** 🔐
Injects JWT token and handles token refresh.

```typescript
// Adds to every request:
{
  headers: {
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIs...'
  }
}

// On 401 Unauthorized:
// 1. Check refresh token
// 2. Get new access token
// 3. Retry original request with new token
```

**Order:** LAST (applied after all other interceptors)

---

## Interceptor Execution Order

```
Request Flow (Client → Server):
┌─────────────────────────────────────────────┐
│ 1. Logging Interceptor                      │ Capture original request
│    - Log method, URL, headers, body         │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ 2. Retry Interceptor                        │ Prepare retry logic
│    - Set up retry policy                    │
│    - Attach to request context              │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ 3. Error Interceptor                        │ Ready error handler
│    - Attach error transformer               │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ 4. Tenant Interceptor                       │ Add tenant header
│    - headers['X-Tenant-ID'] = tenant_id     │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ 5. Auth Interceptor                         │ Add auth token
│    - headers['Authorization'] = token       │
└─────────────────────────────────────────────┘
                     ↓
         [HTTP Request Sent to Server]
                     ↓
Response Flow (Server → Client):
┌─────────────────────────────────────────────┐
│ Auth Interceptor (Response)                 │ Handle 401, refresh token
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ Tenant Interceptor (Response)               │ Validate tenant in response
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ Error Interceptor (Response)                │ Transform errors
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ Retry Interceptor (Response)                │ Check retry eligibility
│    - If retriable: restart request          │
│    - Else: pass to logging                  │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│ Logging Interceptor (Response)              │ Log response
│    - Log status, headers, body              │
└─────────────────────────────────────────────┘
                     ↓
         [Response returned to caller]
```

---

## Configuration

### Basic Initialization

```typescript
import { initializeApi } from '@chetana/api';

initializeApi({
  baseUrl: 'https://api.chetana.example.com',
  withAuth: true,
  withTenant: true,
  withErrorHandling: true,
  withRetry: true,
  withLogging: false, // Enable only in development
});
```

### Advanced Configuration

```typescript
import { initializeApi } from '@chetana/api';

initializeApi({
  // Base URL for internal ConnectRPC
  baseUrl: 'https://api.chetana.example.com',

  // Protocol configuration
  protocol: 'grpc-web', // or 'connect' for newer versions

  // Timeout settings
  timeout: 30000, // 30 seconds

  // Retry configuration
  retry: {
    maxRetries: 3,
    backoff: 'exponential',
    initialDelayMs: 1000,
    maxDelayMs: 32000,
    retryableStatuses: [408, 429, 500, 502, 503, 504],
  },

  // Auth configuration
  auth: {
    tokenProvider: () => localStorage.getItem('token'),
    refreshEndpoint: '/auth/refresh',
  },

  // Tenant configuration
  tenant: {
    tenantProvider: () => localStorage.getItem('tenant_id'),
  },

  // Debug mode
  debug: process.env.NODE_ENV === 'development',

  // Interceptor flags
  withAuth: true,
  withTenant: true,
  withErrorHandling: true,
  withRetry: true,
  withLogging: process.env.NODE_ENV === 'development',
});
```

---

## Usage Examples

### ConnectRPC (Internal)

```typescript
import { ApiClient } from '@chetana/api';
import { SalesService } from '@chetana/backend';

const apiClient = new ApiClient();
const salesService = apiClient.getService(SalesService);

// Create order
const order = await salesService.createOrder({
  customerId: 'cust_123',
  items: [
    { productId: 'prod_456', quantity: 2 },
  ],
});

console.log(`Order created: ${order.id}`);
```

### 3rd Party Integration (REST)

```typescript
import fetch from 'node-fetch';

// Payment provider API (REST)
async function processPayment(chargeData) {
  const response = await fetch('https://api.stripe.example.com/charges', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${process.env.STRIPE_API_KEY}`,
    },
    body: JSON.stringify(chargeData),
  });

  if (!response.ok) {
    throw new Error(`Payment failed: ${response.statusText}`);
  }

  return response.json();
}
```

### With Error Handling

```typescript
import { ApiClient } from '@chetana/api';

try {
  const user = await apiClient.call(UserService, 'getUser', { id: '123' });
} catch (error) {
  if (error.code === 'UNAUTHORIZED') {
    // Redirect to login
    window.location.href = '/login';
  } else if (error.code === 'NOT_FOUND') {
    // Show 404 message
    console.error('User not found');
  } else {
    // Generic error handling
    console.error(`Error: ${error.message}`);
  }
}
```

---

## Why Multiple Interceptors?

Each interceptor serves a **specific, single responsibility**:

| Interceptor | Purpose | Runs | Impact |
|---|---|---|---|
| **Logging** | Debug & monitor | Every request | Development only |
| **Retry** | Resilience | Failed requests | Automatic recovery |
| **Error** | Consistency | Error responses | Standardized errors |
| **Tenant** | Security | Every request | Data isolation |
| **Auth** | Identity | Every request | Authentication |

This **separation of concerns** means:
- ✅ Each interceptor is testable independently
- ✅ Can enable/disable individual interceptors
- ✅ Easy to add new interceptors (e.g., caching, rate limiting)
- ✅ Clear responsibility for each piece
- ✅ Reusable across different API clients

---

## Best Practices

### ✅ DO
- Initialize API client early in app startup
- Use ConnectRPC for internal communication
- Use REST only for 3rd party integration
- Enable retry for resilience
- Enable error handling for consistency
- Use tenant interceptor for multi-tenant isolation
- Log in development, disable in production

### ❌ DON'T
- Manually handle 401 errors (auth interceptor does it)
- Retry requests manually (retry interceptor handles it)
- Transform errors differently in different places
- Forget to set tenant ID in multi-tenant scenarios
- Log sensitive data (PII, tokens, passwords)

---

## See Also

- [ConnectRPC Documentation](https://connectrpc.com)
- [Protobuf Documentation](https://developers.google.com/protocol-buffers)
- Backend docs on RPC services
- Auth documentation
- Error handling guide

---

**Status:** ✅ DOCUMENTED
**Last Updated:** February 25, 2026
