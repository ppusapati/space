/**
 * Identity Service Factories
 * Typed ConnectRPC clients for auth, users, tenants, roles, access
 */
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
export { TenantService, UserService, RoleService, UserTenantService, AccessService, PDPService, AuthService, IdentityEntityService, PermissionService, PermissionDefService, TenantInternalService, };
/** Typed client for TenantService */
export function getTenantService() {
    return getApiClient().getService(TenantService);
}
/** Typed client for UserService */
export function getUserService() {
    return getApiClient().getService(UserService);
}
/** Typed client for RoleService */
export function getRoleService() {
    return getApiClient().getService(RoleService);
}
/** Typed client for UserTenantService */
export function getUserTenantService() {
    return getApiClient().getService(UserTenantService);
}
/** Typed client for AccessService */
export function getAccessService() {
    return getApiClient().getService(AccessService);
}
/** Typed client for PDPService (Policy Decision Point) */
export function getPDPService() {
    return getApiClient().getService(PDPService);
}
/** Typed client for AuthService (authentication, login, tokens) */
export function getAuthService() {
    return getApiClient().getService(AuthService);
}
/** Typed client for IdentityEntityService (entity management) */
export function getIdentityEntityService() {
    return getApiClient().getService(IdentityEntityService);
}
/** Typed client for PermissionService (permission checks) */
export function getPermissionService() {
    return getApiClient().getService(PermissionService);
}
/** Typed client for PermissionDefService (permission definitions) */
export function getPermissionDefService() {
    return getApiClient().getService(PermissionDefService);
}
/** Typed client for TenantInternalService (internal tenant operations) */
export function getTenantInternalService() {
    return getApiClient().getService(TenantInternalService);
}
//# sourceMappingURL=identity.js.map