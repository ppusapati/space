/**
 * common.ts — error shape + fetch wrapper shared by every client.
 *
 * The chetana cmd-layer JSON handlers emit a canonical error
 * envelope:
 *
 *   { "error": "<machine_code>", "error_description": "<human>" }
 *
 * (matches the OAuth 2.1 §5.2 shape for the OAuth endpoints +
 * the per-route conventions for the rest). `request()` decodes
 * that envelope into ApiError so callers can `instanceof` test.
 */

export interface ApiErrorJSON {
  error: string;
  error_description?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly description?: string;

  constructor(status: number, code: string, description?: string) {
    super(description ? `${code}: ${description}` : code);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.description = description;
  }
}

export function isApiError(e: unknown): e is ApiError {
  return e instanceof ApiError;
}

export interface RequestInit_ extends RequestInit {
  /**
   * Bearer token to stamp into Authorization. When omitted the
   * caller is making an unauthenticated request (login + reset
   * fall in this bucket).
   */
  bearer?: string;
  /**
   * Per-call abort. Defaults to no timeout — callers compose
   * AbortController for tab-navigation cancellation.
   */
  signal?: AbortSignal;
}

/**
 * request issues a JSON RPC against the chetana cmd-layer.
 *
 * Conventions:
 *   • POST + JSON body for write ops.
 *   • GET + querystring for read ops.
 *   • Bearer auth via Authorization header.
 *   • 2xx → parsed JSON (or undefined for 204).
 *   • non-2xx → ApiError with the parsed envelope.
 */
export async function request<T = unknown>(
  url: string,
  init: RequestInit_ = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.bearer) {
    headers.set("Authorization", `Bearer ${init.bearer}`);
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(url, {
    ...init,
    headers,
    credentials: init.credentials ?? "include",
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  let body: unknown = undefined;
  if (text.length > 0) {
    try {
      body = JSON.parse(text);
    } catch {
      // Non-JSON body — surface the raw text as the description.
      throw new ApiError(res.status, `http_${res.status}`, text);
    }
  }

  if (!res.ok) {
    const env = body as ApiErrorJSON | undefined;
    throw new ApiError(
      res.status,
      env?.error ?? `http_${res.status}`,
      env?.error_description,
    );
  }

  return body as T;
}

/**
 * baseURL resolves the chetana platform host for the running
 * environment. The shell sets VITE_CHETANA_API_BASE at build time
 * (or runtime via window.__CHETANA_API_BASE__ for k8s
 * configmap-based override).
 */
export function baseURL(): string {
  const w = globalThis as unknown as {
    __CHETANA_API_BASE__?: string;
    __vite_env__?: { VITE_CHETANA_API_BASE?: string };
  };
  if (typeof window !== "undefined" && w.__CHETANA_API_BASE__) {
    return w.__CHETANA_API_BASE__;
  }
  // Vite stamps process.env.VITE_* + import.meta.env.VITE_* at
  // build time. We accept either so this module is portable to
  // the SSR + browser bundles.
  const env =
    (typeof process !== "undefined"
      ? (process.env as Record<string, string | undefined>)
      : undefined) ?? {};
  return env.VITE_CHETANA_API_BASE ?? "/";
}

/**
 * joinURL stitches a base + path safely.
 */
export function joinURL(base: string, path: string): string {
  if (base.endsWith("/") && path.startsWith("/")) return base + path.slice(1);
  if (!base.endsWith("/") && !path.startsWith("/")) return `${base}/${path}`;
  return base + path;
}
