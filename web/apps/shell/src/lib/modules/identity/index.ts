/**
 * Identity domain module.
 *
 * Surfaces every standalone List RPC that returns 200 across the
 * 6 identity services (access / auth / entity / pdp / tenant / user).
 *
 * 2026-04-29 (DEPLOYMENT_READINESS item 42): the AccessService 7 List
 * RPCs are now wired. Previously gated on PolicySetService + AuditService
 * signature alignment (Audit #1 deferred-work). Resolution: concrete
 * services rewritten to satisfy canonical svcInterfaces; AccessService
 * route flipped from defunct `grpc_services` group to `routes`;
 * `resolveTenantID(ctx, mc)` helper added so RPCs read JWT-verified
 * tenant id (was reading proto-context only — empty under JWT).
 *
 * 2026-04-29 (Track A.2 of identity completion): RoleService and
 * PermissionDefService under `core.identity.user.api.v1` are now wired
 * — read-side RPCs (GetByID/GetAll/GetByName/ListByRole) are backed by
 * repos; write-side RPCs return CodeUnimplemented via embedded
 * Unimplemented*Handler until the user-services-tree restoration
 * (Track A.3, ~1 day per audit doc — 17 UserService RPCs + Role/
 * Permission/PermissionDef/UserTenant write methods). Distinct from
 * AccessService/Roles: the user-package Role is the hierarchy-aware
 * role (id/name/parent_id/is_preserved/metadata), while AccessService
 * Role is the RBAC role (displayName/roleType/isSystem). Both surfaces
 * intentionally exposed.
 *
 * Wire-shape gotchas captured per entity:
 *   - tenant/ListTenant uses `items` (not `tenants`) + `totalSize`
 *     (not `totalCount`). Proto declares `repeated Tenant items = 2;
 *     int32 total_size = 3;` directly — these are non-canonical key
 *     names that ListPage's lookupPath() handles via explicit
 *     responseRowsKey + responseTotalKey.
 *   - user/ListUsers uses `users` + `totalSize` (totalCount is also
 *     not present). Proto declares `int32 total_size = 2; int32
 *     filter_size = 3; repeated User users = 4;`.
 *   - All 4 pdp Lists carry both `totalCount` (int32) AND
 *     `nextPageToken` (string). Loader reads totalCount; if a future
 *     use-case needs token pagination, extend the loader (do NOT
 *     shoehorn token state into pageOffset).
 *
 * Tenant defect history (closed 2026-04-26): ListTenant was returning
 * 500 "cannot scan unknown type (OID 630503)" because sqlc generated
 * `TenantDb interface{}` for the `identity.tenant_db_type` enum, and
 * pgx can't scan an unregistered OID into interface{}. Hand-fixed the
 * 4 generated struct fields to `string` (the enum is stored as text);
 * downstream interfaceToTenantDBType already handled the string case.
 *
 * pdp defect history (closed 2026-04-27): GetCacheStats was returning
 * 500 `cannot scan NULL into *int64` for empty-cache tenants. Fixed
 * at SQL layer with COALESCE(SUM(...), 0)::BIGINT — see
 * project_pdp_cachestats_nullable_fix.md memory.
 */
import type { DomainModule } from '../index.js';

export const identity: DomainModule = {
  id: 'identity',
  label: 'Identity',
  entities: [
    // ---- tenant ----
    {
      slug: 'tenants',
      label: 'Tenants',
      formId: 'access_level_assignment',
      listEndpoint: '/core.identity.tenant.api.v1.TenantService/ListTenant',
      responseRowsKey: 'items',
      responseTotalKey: 'totalSize',
      columns: ['id', 'name', 'displayName', 'region'],
    },
    // ---- entity ----
    {
      slug: 'entities',
      label: 'Entities',
      formId: 'access_level_assignment',
      listEndpoint: '/core.identity.entity.api.v1.EntityService/ListEntities',
      responseRowsKey: 'entities',
      responseTotalKey: 'totalCount',
      columns: ['id', 'name', 'type', 'status', 'createdAt'],
    },
    // ---- user ----
    {
      slug: 'users',
      label: 'Users',
      formId: 'user_provisioning',
      listEndpoint: '/core.identity.user.api.v1.UserService/ListUsers',
      responseRowsKey: 'users',
      responseTotalKey: 'totalSize',
      columns: ['username', 'fullname', 'email', 'phone', 'gender'],
    },
    // ---- pdp ----
    {
      slug: 'evaluation-strategies',
      label: 'Evaluation Strategies',
      formId: 'permission_management',
      listEndpoint: '/core.identity.pdp.api.v1.PDPService/ListEvaluationStrategies',
      responseRowsKey: 'strategies',
      responseTotalKey: 'totalCount',
      columns: ['strategyName', 'strategyType', 'combiningAlgorithm', 'isDefault', 'isActive', 'priority'],
    },
    {
      slug: 'custom-rules',
      label: 'Custom Rules',
      formId: 'permission_management',
      listEndpoint: '/core.identity.pdp.api.v1.PDPService/ListCustomRules',
      responseRowsKey: 'rules',
      responseTotalKey: 'totalCount',
      columns: ['ruleName', 'ruleType', 'description', 'executionOrder'],
    },
    {
      slug: 'test-cases',
      label: 'Test Cases',
      formId: 'permission_management',
      listEndpoint: '/core.identity.pdp.api.v1.PDPService/ListTestCases',
      responseRowsKey: 'testCases',
      responseTotalKey: 'totalCount',
      columns: ['testName', 'description', 'expectedDecision', 'isActive'],
    },
    {
      slug: 'attribute-providers',
      label: 'Attribute Providers',
      formId: 'permission_management',
      listEndpoint: '/core.identity.pdp.api.v1.PDPService/ListAttributeProviders',
      responseRowsKey: 'providers',
      responseTotalKey: 'totalCount',
      columns: ['providerName', 'providerType', 'endpointUrl', 'isActive', 'cacheTtlSeconds'],
    },
    // ---- access (7 RPCs unblocked 2026-04-29 by item 42) ----
    {
      slug: 'roles',
      label: 'Roles',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListRoles',
      responseRowsKey: 'roles',
      responseTotalKey: 'totalCount',
      columns: ['name', 'displayName', 'roleType', 'isSystem', 'isActive'],
    },
    {
      slug: 'permissions',
      label: 'Permissions',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListPermissions',
      responseRowsKey: 'permissions',
      responseTotalKey: 'totalCount',
      columns: ['name', 'displayName', 'resourceType', 'action', 'effect', 'scope'],
    },
    {
      slug: 'resources',
      label: 'Resources',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListResources',
      responseRowsKey: 'resources',
      responseTotalKey: 'totalCount',
      columns: ['resourceType', 'resourceId', 'resourceName', 'ownerId', 'visibility'],
    },
    {
      slug: 'acls',
      label: 'Access Control Lists',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListACLs',
      responseRowsKey: 'acls',
      responseTotalKey: 'totalCount',
      columns: ['resourceId', 'subjectId', 'subjectType', 'permissionId', 'effect', 'expiresAt'],
    },
    {
      slug: 'attributes',
      label: 'Attributes',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListAttributes',
      responseRowsKey: 'attributes',
      responseTotalKey: 'totalCount',
      columns: ['attributeName', 'attributeType', 'dataType', 'isRequired', 'isSystem'],
    },
    {
      slug: 'policies',
      label: 'Policies',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListPolicies',
      responseRowsKey: 'policies',
      responseTotalKey: 'totalCount',
      columns: ['policyName', 'displayName', 'policyType', 'effect', 'priority', 'isActive'],
    },
    {
      slug: 'delegations',
      label: 'Delegations',
      formId: 'permission_management',
      listEndpoint: '/identity.access.api.v1.AccessService/ListDelegations',
      responseRowsKey: 'delegations',
      responseTotalKey: 'totalCount',
      columns: ['delegatorId', 'delegateeId', 'roleId', 'delegationType', 'reason'],
    },
    // ---- user-package role + permission-def (Track A.2 — 2026-04-29) ----
    // Distinct from access/roles above: this is the hierarchy-aware Role
    // under core.identity.user.api.v1 (id/name/parentId/isPreserved).
    {
      slug: 'user-roles',
      label: 'User Roles',
      formId: 'permission_management',
      listEndpoint: '/core.identity.user.api.v1.RoleService/GetAll',
      responseRowsKey: 'roles',
      responseTotalKey: 'totalSize',
      columns: ['name', 'parentId', 'isPreserved'],
    },
    // PermissionDefService.GetAll proto returns `items` only (no total).
    // ListPage's loader infers count from rows.length when totalKey is
    // absent — leave responseTotalKey undefined.
    {
      slug: 'permission-defs',
      label: 'Permission Definitions',
      formId: 'permission_management',
      listEndpoint: '/core.identity.user.api.v1.PermissionDefService/GetAll',
      responseRowsKey: 'items',
      columns: ['name', 'namespace', 'resource', 'action', 'scope', 'version'],
    },
  ],
};
