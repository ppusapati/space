/**
 * Identity Service Factories
 * Typed ConnectRPC clients for auth, users, tenants, roles, access
 */
import type { Client } from '@connectrpc/connect';
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
export { TenantService, UserService, RoleService, UserTenantService, AccessService, PDPService, AuthService, IdentityEntityService, PermissionService, PermissionDefService, TenantInternalService, };
export type { Client };
/** Typed client for TenantService */
export declare function getTenantService(): Client<typeof TenantService>;
/** Typed client for UserService */
export declare function getUserService(): Client<typeof UserService>;
/** Typed client for RoleService */
export declare function getRoleService(): Client<typeof RoleService>;
/** Typed client for UserTenantService */
export declare function getUserTenantService(): Client<typeof UserTenantService>;
/** Typed client for AccessService */
export declare function getAccessService(): Client<typeof AccessService>;
/** Typed client for PDPService (Policy Decision Point) */
export declare function getPDPService(): Client<typeof PDPService>;
/** Typed client for AuthService (authentication, login, tokens) */
export declare function getAuthService(): Client<typeof AuthService>;
/** Typed client for IdentityEntityService (entity management) */
export declare function getIdentityEntityService(): Client<typeof IdentityEntityService>;
/** Typed client for PermissionService (permission checks) */
export declare function getPermissionService(): Client<typeof PermissionService>;
/** Typed client for PermissionDefService (permission definitions) */
export declare function getPermissionDefService(): Client<typeof PermissionDefService>;
/** Typed client for TenantInternalService (internal tenant operations) */
export declare function getTenantInternalService(): Client<typeof TenantInternalService>;
//# sourceMappingURL=identity.d.ts.map