/**
 * Samavaya Offline - TypeScript Bindings
 * Offline sync and conflict resolution using WASM
 */

import { loadWasmModule, type WasmModuleName } from './loader';
import type {
  ChangeRecord,
  ConflictRecord,
  Delta,
  SyncResult,
  VersionVector,
  MergeResult,
  OperationType,
  SyncStatus,
  ConflictResolutionStrategy,
} from './types';

// Type for the raw WASM module
interface OfflineWasm {
  create_change_record: (entityType: string, entityId: string, operation: string, data: unknown, userId: string) => unknown;
  calculate_delta: (before: unknown, after: unknown) => unknown;
  apply_delta: (base: unknown, delta: unknown) => unknown;
  detect_conflict: (localChange: unknown, remoteChange: unknown) => unknown;
  resolve_conflict: (localChange: unknown, remoteChange: unknown, strategy: string) => unknown;
  three_way_merge: (base: unknown, local: unknown, remote: unknown) => unknown;
  version_vector_increment: (vector: unknown, nodeId: string) => unknown;
  version_vector_merge: (v1: unknown, v2: unknown) => unknown;
  version_vector_compare: (v1: unknown, v2: unknown) => string;
  version_vector_is_concurrent: (v1: unknown, v2: unknown) => boolean;
  generate_sync_id: () => string;
  calculate_checksum: (data: unknown) => string;
  create_sync_batch: (changes: unknown, batchSize: number) => unknown;
  apply_sync_batch: (batch: unknown, currentData: unknown) => unknown;
  get_pending_changes: (changes: unknown, lastSyncVector: unknown) => unknown;
  mark_synced: (changes: unknown, syncedIds: unknown) => unknown;
  compact_changes: (changes: unknown) => unknown;
  validate_sync_order: (changes: unknown) => unknown;
}

let wasmModule: OfflineWasm | null = null;

/**
 * Initialize the offline module
 */
async function ensureLoaded(): Promise<OfflineWasm> {
  if (!wasmModule) {
    wasmModule = await loadWasmModule<OfflineWasm>('offline' as WasmModuleName);
  }
  return wasmModule;
}

// ============================================================================
// Change Tracking Functions
// ============================================================================

/**
 * Create a change record for an entity modification
 * @param entityType - Type of entity (e.g., "invoice", "customer")
 * @param entityId - Entity identifier
 * @param operation - Operation type (create, update, delete)
 * @param data - Changed data
 * @param userId - User who made the change
 * @returns Change record
 */
export async function createChangeRecord(
  entityType: string,
  entityId: string,
  operation: OperationType,
  data: unknown,
  userId: string
): Promise<ChangeRecord> {
  const wasm = await ensureLoaded();
  return wasm.create_change_record(entityType, entityId, operation, data, userId) as ChangeRecord;
}

/**
 * Calculate delta between two states
 * @param before - State before change
 * @param after - State after change
 * @returns Delta object
 */
export async function calculateDelta(before: unknown, after: unknown): Promise<Delta> {
  const wasm = await ensureLoaded();
  return wasm.calculate_delta(before, after) as Delta;
}

/**
 * Apply delta to base state
 * @param base - Base state
 * @param delta - Delta to apply
 * @returns New state
 */
export async function applyDelta<T>(base: T, delta: Delta): Promise<T> {
  const wasm = await ensureLoaded();
  return wasm.apply_delta(base, delta) as T;
}

// ============================================================================
// Conflict Detection and Resolution
// ============================================================================

/**
 * Detect if two changes conflict
 * @param localChange - Local change record
 * @param remoteChange - Remote change record
 * @returns Conflict record if conflict exists
 */
export async function detectConflict(
  localChange: ChangeRecord,
  remoteChange: ChangeRecord
): Promise<ConflictRecord | null> {
  const wasm = await ensureLoaded();
  return wasm.detect_conflict(localChange, remoteChange) as ConflictRecord | null;
}

/**
 * Resolve a conflict using specified strategy
 * @param localChange - Local change record
 * @param remoteChange - Remote change record
 * @param strategy - Resolution strategy
 * @returns Resolved change record
 */
export async function resolveConflict(
  localChange: ChangeRecord,
  remoteChange: ChangeRecord,
  strategy: ConflictResolutionStrategy
): Promise<ChangeRecord> {
  const wasm = await ensureLoaded();
  return wasm.resolve_conflict(localChange, remoteChange, strategy) as ChangeRecord;
}

/**
 * Perform three-way merge
 * @param base - Common ancestor state
 * @param local - Local state
 * @param remote - Remote state
 * @returns Merge result
 */
export async function threeWayMerge<T>(base: T, local: T, remote: T): Promise<MergeResult<T>> {
  const wasm = await ensureLoaded();
  return wasm.three_way_merge(base, local, remote) as MergeResult<T>;
}

// ============================================================================
// Version Vector Functions
// ============================================================================

/**
 * Increment version vector for a node
 * @param vector - Current version vector
 * @param nodeId - Node identifier
 * @returns Updated version vector
 */
export async function versionVectorIncrement(
  vector: VersionVector,
  nodeId: string
): Promise<VersionVector> {
  const wasm = await ensureLoaded();
  return wasm.version_vector_increment(vector, nodeId) as VersionVector;
}

/**
 * Merge two version vectors
 * @param v1 - First version vector
 * @param v2 - Second version vector
 * @returns Merged version vector
 */
export async function versionVectorMerge(
  v1: VersionVector,
  v2: VersionVector
): Promise<VersionVector> {
  const wasm = await ensureLoaded();
  return wasm.version_vector_merge(v1, v2) as VersionVector;
}

/**
 * Compare two version vectors
 * @param v1 - First version vector
 * @param v2 - Second version vector
 * @returns Comparison result (before, after, equal, concurrent)
 */
export async function versionVectorCompare(
  v1: VersionVector,
  v2: VersionVector
): Promise<'before' | 'after' | 'equal' | 'concurrent'> {
  const wasm = await ensureLoaded();
  return wasm.version_vector_compare(v1, v2) as 'before' | 'after' | 'equal' | 'concurrent';
}

/**
 * Check if two version vectors are concurrent
 * @param v1 - First version vector
 * @param v2 - Second version vector
 * @returns Whether versions are concurrent
 */
export async function versionVectorIsConcurrent(
  v1: VersionVector,
  v2: VersionVector
): Promise<boolean> {
  const wasm = await ensureLoaded();
  return wasm.version_vector_is_concurrent(v1, v2);
}

// ============================================================================
// Sync Utilities
// ============================================================================

/**
 * Generate a unique sync ID
 * @returns Sync ID
 */
export async function generateSyncId(): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.generate_sync_id();
}

/**
 * Calculate checksum for data integrity
 * @param data - Data to checksum
 * @returns Checksum string
 */
export async function calculateChecksum(data: unknown): Promise<string> {
  const wasm = await ensureLoaded();
  return wasm.calculate_checksum(data);
}

/**
 * Create sync batches from changes
 * @param changes - Array of change records
 * @param batchSize - Maximum batch size
 * @returns Array of batches
 */
export async function createSyncBatch(
  changes: ChangeRecord[],
  batchSize: number
): Promise<Array<{ batchId: string; changes: ChangeRecord[]; checksum: string }>> {
  const wasm = await ensureLoaded();
  return wasm.create_sync_batch(changes, batchSize) as Array<{
    batchId: string;
    changes: ChangeRecord[];
    checksum: string;
  }>;
}

/**
 * Apply a sync batch to current data
 * @param batch - Sync batch
 * @param currentData - Current data state
 * @returns Updated data and sync result
 */
export async function applySyncBatch<T>(
  batch: { batchId: string; changes: ChangeRecord[]; checksum: string },
  currentData: T
): Promise<{ data: T; result: SyncResult }> {
  const wasm = await ensureLoaded();
  return wasm.apply_sync_batch(batch, currentData) as { data: T; result: SyncResult };
}

/**
 * Get pending changes since last sync
 * @param changes - All change records
 * @param lastSyncVector - Version vector of last sync
 * @returns Pending changes
 */
export async function getPendingChanges(
  changes: ChangeRecord[],
  lastSyncVector: VersionVector
): Promise<ChangeRecord[]> {
  const wasm = await ensureLoaded();
  return wasm.get_pending_changes(changes, lastSyncVector) as ChangeRecord[];
}

/**
 * Mark changes as synced
 * @param changes - Change records
 * @param syncedIds - IDs of synced changes
 * @returns Updated change records
 */
export async function markSynced(
  changes: ChangeRecord[],
  syncedIds: string[]
): Promise<ChangeRecord[]> {
  const wasm = await ensureLoaded();
  return wasm.mark_synced(changes, syncedIds) as ChangeRecord[];
}

/**
 * Compact changes by removing superseded operations
 * @param changes - Change records
 * @returns Compacted changes
 */
export async function compactChanges(changes: ChangeRecord[]): Promise<ChangeRecord[]> {
  const wasm = await ensureLoaded();
  return wasm.compact_changes(changes) as ChangeRecord[];
}

/**
 * Validate sync order of changes
 * @param changes - Change records
 * @returns Validation result
 */
export async function validateSyncOrder(
  changes: ChangeRecord[]
): Promise<{ valid: boolean; errors: string[]; reorderedChanges?: ChangeRecord[] }> {
  const wasm = await ensureLoaded();
  return wasm.validate_sync_order(changes) as {
    valid: boolean;
    errors: string[];
    reorderedChanges?: ChangeRecord[];
  };
}

// ============================================================================
// Convenience Exports
// ============================================================================

export { type ChangeRecord, type ConflictRecord, type Delta, type SyncResult, type VersionVector, type MergeResult };
