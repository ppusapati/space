/**
 * Auth Store
 * Handles authentication state, tokens, and session management
 */
import { type Readable } from 'svelte/store';
export interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    displayName: string;
    avatar?: string;
    phone?: string;
    locale?: string;
    timezone?: string;
    metadata?: Record<string, unknown>;
}
export interface AuthTokens {
    accessToken: string;
    refreshToken: string;
    expiresAt: Date;
    tokenType: 'Bearer';
}
export interface Permission {
    resource: string;
    action: string;
    conditions?: Record<string, unknown>;
}
export interface Role {
    id: string;
    name: string;
    displayName: string;
    permissions: Permission[];
}
export interface Session {
    id: string;
    userId: string;
    deviceInfo?: {
        userAgent: string;
        platform: string;
        browser: string;
    };
    ipAddress?: string;
    location?: string;
    createdAt: Date;
    lastActiveAt: Date;
    expiresAt: Date;
    isCurrent: boolean;
}
export interface AuthState {
    isAuthenticated: boolean;
    isLoading: boolean;
    isInitialized: boolean;
    user: User | null;
    tokens: AuthTokens | null;
    roles: Role[];
    permissions: Permission[];
    session: Session | null;
    error: AuthError | null;
}
export interface AuthError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}
export interface LoginCredentials {
    email: string;
    password: string;
    rememberMe?: boolean;
    mfaCode?: string;
}
export interface RegisterData {
    email: string;
    password: string;
    firstName: string;
    lastName: string;
    phone?: string;
}
export interface AuthStoreActions {
    login: (credentials: LoginCredentials) => Promise<void>;
    logout: () => Promise<void>;
    register: (data: RegisterData) => Promise<void>;
    refreshTokens: () => Promise<void>;
    setTokens: (tokens: AuthTokens) => void;
    clearTokens: () => void;
    setUser: (user: User) => void;
    updateUser: (updates: Partial<User>) => void;
    clearUser: () => void;
    setSession: (session: Session) => void;
    clearSession: () => void;
    setRoles: (roles: Role[]) => void;
    setPermissions: (permissions: Permission[]) => void;
    hasPermission: (resource: string, action: string) => boolean;
    hasRole: (roleName: string) => boolean;
    hasAnyRole: (roleNames: string[]) => boolean;
    hasAllRoles: (roleNames: string[]) => boolean;
    initialize: () => Promise<void>;
    reset: () => void;
    setError: (error: AuthError | null) => void;
    setLoading: (isLoading: boolean) => void;
}
export declare const authStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<AuthState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    user: Readable<User | null>;
    isAuthenticated: Readable<boolean>;
    isLoading: Readable<boolean>;
    roles: Readable<Role[]>;
    permissions: Readable<Permission[]>;
    login: (credentials: LoginCredentials) => Promise<void>;
    logout: () => Promise<void>;
    register: (data: RegisterData) => Promise<void>;
    refreshTokens: () => Promise<void>;
    setTokens: (tokens: AuthTokens) => void;
    clearTokens: () => void;
    setUser: (user: User) => void;
    updateUser: (updates: Partial<User>) => void;
    clearUser: () => void;
    setSession: (session: Session) => void;
    clearSession: () => void;
    setRoles: (roles: Role[]) => void;
    setPermissions: (permissions: Permission[]) => void;
    hasPermission: (resource: string, action: string) => boolean;
    hasRole: (roleName: string) => boolean;
    hasAnyRole: (roleNames: string[]) => boolean;
    hasAllRoles: (roleNames: string[]) => boolean;
    initialize: () => Promise<void>;
    reset: () => void;
    setError: (error: AuthError | null) => void;
    setLoading: (isLoading: boolean) => void;
};
//# sourceMappingURL=auth.store.d.ts.map