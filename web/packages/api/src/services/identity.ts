/**
 * Identity Service Factories
 * Typed ConnectRPC clients for auth, users, tenants, roles, access
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// Service descriptors
import { TenantService, TenantInternalService } from '@samavāya/proto/gen/core/identity/tenant/proto/tenant_pb.js';
import { UserService } from '@samavāya/proto/gen/core/identity/user/proto/user_pb.js';
import { RoleService } from '@samavāya/proto/gen/core/identity/user/proto/role_pb.js';
import { UserTenantService } from '@samavāya/proto/gen/core/identity/user/proto/user_tenant_pb.js';
import { AccessService } from '@samavāya/proto/gen/core/identity/access/proto/access_pb.js';
import { PDPService } from '@samavāya/proto/gen/core/identity/pdp/proto/pdp_pb.js';
import { AuthService } from '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js';
import { EntityService as IdentityEntityService } from '@samavāya/proto/gen/core/identity/entity/proto/entity_pb.js';
import { PermissionService } from '@samavāya/proto/gen/core/identity/user/proto/permission_pb.js';
import { PermissionDefService } from '@samavāya/proto/gen/core/identity/user/proto/permissiondef_pb.js';

// Re-export service descriptors for direct use
export {
  TenantService, UserService, RoleService, UserTenantService, AccessService, PDPService,
  AuthService, IdentityEntityService, PermissionService, PermissionDefService, TenantInternalService,
};

// Re-export key types consumers will need
export type { Client };

/** Typed client for TenantService */
export function getTenantService(): Client<typeof TenantService> {
  return getApiClient().getService(TenantService);
}

/** Typed client for UserService */
export function getUserService(): Client<typeof UserService> {
  return getApiClient().getService(UserService);
}

/** Typed client for RoleService */
export function getRoleService(): Client<typeof RoleService> {
  return getApiClient().getService(RoleService);
}

/** Typed client for UserTenantService */
export function getUserTenantService(): Client<typeof UserTenantService> {
  return getApiClient().getService(UserTenantService);
}

/** Typed client for AccessService */
export function getAccessService(): Client<typeof AccessService> {
  return getApiClient().getService(AccessService);
}

/** Typed client for PDPService (Policy Decision Point) */
export function getPDPService(): Client<typeof PDPService> {
  return getApiClient().getService(PDPService);
}

/** Typed client for AuthService (authentication, login, tokens) */
export function getAuthService(): Client<typeof AuthService> {
  return getApiClient().getService(AuthService);
}

/** Typed client for IdentityEntityService (entity management) */
export function getIdentityEntityService(): Client<typeof IdentityEntityService> {
  return getApiClient().getService(IdentityEntityService);
}

/** Typed client for PermissionService (permission checks) */
export function getPermissionService(): Client<typeof PermissionService> {
  return getApiClient().getService(PermissionService);
}

/** Typed client for PermissionDefService (permission definitions) */
export function getPermissionDefService(): Client<typeof PermissionDefService> {
  return getApiClient().getService(PermissionDefService);
}

/** Typed client for TenantInternalService (internal tenant operations) */
export function getTenantInternalService(): Client<typeof TenantInternalService> {
  return getApiClient().getService(TenantInternalService);
}
