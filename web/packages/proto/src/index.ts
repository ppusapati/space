// @samavāya/proto — Generated TypeScript protobuf types and ConnectRPC service descriptors
//
// Generated files are in src/gen/ — do NOT edit manually.
// Regenerate with: pnpm proto:generate
//
// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// Direct imports (preferred for tree-shaking):
//
//   import { ItemService, ItemSchema, type Item } from
//     '@samavāya/proto/gen/business/masters/item/proto/item_pb.js';
//
//   import { TenantContextSchema, type TenantContext } from
//     '@samavāya/proto/gen/packages/proto/context_pb.js';
//
// Using with ConnectRPC client:
//
//   import { createClient } from '@connectrpc/connect';
//   import { ItemService } from '@samavāya/proto/gen/business/masters/item/proto/item_pb.js';
//   import { create } from '@bufbuild/protobuf';
//
//   const client = createClient(ItemService, transport);
//   const response = await client.createItem({
//     context: create(TenantContextSchema, { tenantId: '...', companyId: '...' }),
//     name: 'Widget',
//     code: 'W-001',
//   });
//
// =============================================================================
// BARREL EXPORTS — Commonly used shared types
// =============================================================================

// Shared context
export {
  TenantContextSchema,
  type TenantContext,
} from './gen/packages/proto/context_pb.js';

// Shared response types
export {
  BaseResponseSchema,
  type BaseResponse,
  type Status,
  CanonicalReason,
} from './gen/packages/proto/response_pb.js';

// Money types
export {
  MoneySchema,
  type Money,
} from './gen/packages/proto/money_pb.js';

// Pagination
export {
  PaginationSchema,
  type Pagination,
} from './gen/packages/proto/pagination_pb.js';
