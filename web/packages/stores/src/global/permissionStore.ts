/**
 * Permission Store & Composable
 *
 * Role-based access control (RBAC) and permission management:
 * - Permission checking with caching
 * - Role hierarchy support
 * - Resource-based permissions
 * - Feature permissions
 * - UI permission helpers
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type PermissionAction = 'create' | 'read' | 'update' | 'delete' | 'execute' | 'approve' | 'export' | 'import' | 'manage';

export interface Permission {
  resource: string;
  action: PermissionAction;
  conditions?: Record<string, unknown>;
}

export interface Role {
  id: string;
  name: string;
  permissions: Permission[];
  inherits?: string[];
}

export interface PermissionUser {
  id: string;
  roles: string[];
  directPermissions?: Permission[];
  tenantId?: string;
  companyId?: string;
  branchId?: string;
}

export interface PermissionConfig {
  /** Available roles */
  roles: Role[];
  /** Super admin role that bypasses all checks */
  superAdminRole?: string;
  /** Cache TTL in ms (default: 5 minutes) */
  cacheTtl: number;
  /** Fetch permissions from backend */
  fetchPermissions?: () => Promise<{ roles: Role[]; user: PermissionUser }>;
}

export interface PermissionState {
  user: PermissionUser | null;
  roles: Map<string, Role>;
  cache: Map<string, { result: boolean; expiresAt: number }>;
  isLoading: boolean;
  isInitialized: boolean;
}

// ============================================================================
// Store Implementation
// ============================================================================

const DEFAULT_CONFIG: Partial<PermissionConfig> = {
  cacheTtl: 5 * 60 * 1000,
};

function createPermissionStore() {
  const { subscribe, set, update } = writable<PermissionState>({
    user: null,
    roles: new Map(),
    cache: new Map(),
    isLoading: false,
    isInitialized: false,
  });

  let config: PermissionConfig = { roles: [], ...DEFAULT_CONFIG } as PermissionConfig;

  function buildRoleMap(roles: Role[]): Map<string, Role> {
    const map = new Map<string, Role>();
    for (const role of roles) {
      map.set(role.id, role);
    }
    return map;
  }

  function getCacheKey(resource: string, action: PermissionAction, conditions?: Record<string, unknown>): string {
    return `${resource}:${action}:${JSON.stringify(conditions || {})}`;
  }

  function getAllPermissionsForRole(roleId: string, roleMap: Map<string, Role>, visited = new Set<string>()): Permission[] {
    if (visited.has(roleId)) return []; // Prevent circular inheritance
    visited.add(roleId);

    const role = roleMap.get(roleId);
    if (!role) return [];

    let permissions = [...role.permissions];

    // Add inherited permissions
    if (role.inherits) {
      for (const inheritedRoleId of role.inherits) {
        permissions = [...permissions, ...getAllPermissionsForRole(inheritedRoleId, roleMap, visited)];
      }
    }

    return permissions;
  }

  function checkPermission(
    user: PermissionUser,
    roleMap: Map<string, Role>,
    resource: string,
    action: PermissionAction,
    conditions?: Record<string, unknown>
  ): boolean {
    // Super admin bypass
    if (config.superAdminRole && user.roles.includes(config.superAdminRole)) {
      return true;
    }

    // Collect all permissions from user's roles
    let allPermissions: Permission[] = [];
    for (const roleId of user.roles) {
      allPermissions = [...allPermissions, ...getAllPermissionsForRole(roleId, roleMap)];
    }

    // Add direct permissions
    if (user.directPermissions) {
      allPermissions = [...allPermissions, ...user.directPermissions];
    }

    // Check if any permission matches
    return allPermissions.some(perm => {
      // Check resource (support wildcards)
      if (perm.resource !== '*' && perm.resource !== resource) {
        // Check for prefix wildcard (e.g., "sales:*" matches "sales:invoices")
        if (!perm.resource.endsWith(':*') || !resource.startsWith(perm.resource.slice(0, -1))) {
          return false;
        }
      }

      // Check action
      if (perm.action !== action && perm.action !== 'manage') {
        return false;
      }

      // Check conditions
      if (perm.conditions && conditions) {
        for (const [key, value] of Object.entries(perm.conditions)) {
          if (conditions[key] !== value) {
            return false;
          }
        }
      }

      return true;
    });
  }

  return {
    subscribe,

    /** Initialize with roles and user */
    init(user: PermissionUser, options: PermissionConfig) {
      config = { ...DEFAULT_CONFIG, ...options } as PermissionConfig;
      const roleMap = buildRoleMap(config.roles);
      set({ user, roles: roleMap, cache: new Map(), isLoading: false, isInitialized: true });
    },

    /** Load permissions from backend */
    async load(): Promise<void> {
      if (!config.fetchPermissions) return;

      update(s => ({ ...s, isLoading: true }));

      try {
        const { roles, user } = await config.fetchPermissions();
        config.roles = roles;
        const roleMap = buildRoleMap(roles);
        update(s => ({ ...s, user, roles: roleMap, isLoading: false, isInitialized: true, cache: new Map() }));
      } catch (error) {
        update(s => ({ ...s, isLoading: false }));
        throw error;
      }
    },

    /** Set current user */
    setUser(user: PermissionUser | null) {
      update(s => ({ ...s, user, cache: new Map() }));
    },

    /** Check if user has permission */
    can(resource: string, action: PermissionAction, conditions?: Record<string, unknown>): boolean {
      const state = get({ subscribe });
      if (!state.user) return false;

      const cacheKey = getCacheKey(resource, action, conditions);

      // Check cache
      const cached = state.cache.get(cacheKey);
      if (cached && cached.expiresAt > Date.now()) {
        return cached.result;
      }

      const result = checkPermission(state.user, state.roles, resource, action, conditions);

      // Cache result
      update(s => {
        const cache = new Map(s.cache);
        cache.set(cacheKey, { result, expiresAt: Date.now() + config.cacheTtl });
        return { ...s, cache };
      });

      return result;
    },

    /** Check multiple permissions (all must pass) */
    canAll(...checks: [string, PermissionAction, Record<string, unknown>?][]): boolean {
      return checks.every(([resource, action, conditions]) => this.can(resource, action, conditions));
    },

    /** Check multiple permissions (any must pass) */
    canAny(...checks: [string, PermissionAction, Record<string, unknown>?][]): boolean {
      return checks.some(([resource, action, conditions]) => this.can(resource, action, conditions));
    },

    /** Check if user has role */
    hasRole(roleId: string): boolean {
      const state = get({ subscribe });
      return state.user?.roles.includes(roleId) || false;
    },

    /** Check if user has any of the roles */
    hasAnyRole(...roleIds: string[]): boolean {
      const state = get({ subscribe });
      return roleIds.some(id => state.user?.roles.includes(id));
    },

    /** Get all permissions for current user */
    getAllPermissions(): Permission[] {
      const state = get({ subscribe });
      if (!state.user) return [];

      let permissions: Permission[] = [];
      for (const roleId of state.user.roles) {
        permissions = [...permissions, ...getAllPermissionsForRole(roleId, state.roles)];
      }
      if (state.user.directPermissions) {
        permissions = [...permissions, ...state.user.directPermissions];
      }
      return permissions;
    },

    /** Clear permission cache */
    clearCache() {
      update(s => ({ ...s, cache: new Map() }));
    },

    /** Reset store */
    reset() {
      set({ user: null, roles: new Map(), cache: new Map(), isLoading: false, isInitialized: false });
    },
  };
}

export const permissionStore = createPermissionStore();

// ============================================================================
// Derived Stores
// ============================================================================

export const currentUser = derived(permissionStore, $s => $s.user);
export const isPermissionLoading = derived(permissionStore, $s => $s.isLoading);
export const isPermissionInitialized = derived(permissionStore, $s => $s.isInitialized);

// ============================================================================
// usePermission Composable (Svelte 5)
// ============================================================================

/**
 * Svelte 5 composable for permission checking
 * Usage:
 * ```svelte
 * <script>
 *   const permission = usePermission();
 *
 *   // Check permissions reactively
 *   $: canEdit = permission.can('invoices', 'update');
 *   $: canDelete = permission.can('invoices', 'delete');
 * </script>
 *
 * {#if canEdit}
 *   <button>Edit</button>
 * {/if}
 * ```
 */
export function usePermission() {
  return {
    subscribe: permissionStore.subscribe,
    can: permissionStore.can.bind(permissionStore),
    canAll: permissionStore.canAll.bind(permissionStore),
    canAny: permissionStore.canAny.bind(permissionStore),
    hasRole: permissionStore.hasRole.bind(permissionStore),
    hasAnyRole: permissionStore.hasAnyRole.bind(permissionStore),
    getAllPermissions: permissionStore.getAllPermissions.bind(permissionStore),
  };
}

// ============================================================================
// Permission Helpers
// ============================================================================

/**
 * Create a permission guard for routes
 */
export function createPermissionGuard(
  resource: string,
  action: PermissionAction,
  redirectTo = '/unauthorized'
): () => string | null {
  return () => {
    if (permissionStore.can(resource, action)) {
      return null; // Allow access
    }
    return redirectTo; // Redirect
  };
}

/**
 * Permission check for use in load functions
 */
export function requirePermission(
  resource: string,
  action: PermissionAction,
  conditions?: Record<string, unknown>
): void {
  if (!permissionStore.can(resource, action, conditions)) {
    throw new Error(`Permission denied: ${action} on ${resource}`);
  }
}

/**
 * Create resource-specific permission helpers
 */
export function createResourcePermissions(resource: string) {
  return {
    canCreate: () => permissionStore.can(resource, 'create'),
    canRead: () => permissionStore.can(resource, 'read'),
    canUpdate: () => permissionStore.can(resource, 'update'),
    canDelete: () => permissionStore.can(resource, 'delete'),
    canExport: () => permissionStore.can(resource, 'export'),
    canImport: () => permissionStore.can(resource, 'import'),
    canApprove: () => permissionStore.can(resource, 'approve'),
    canManage: () => permissionStore.can(resource, 'manage'),
  };
}
