//! Offline Synchronization
//!
//! This crate provides:
//! - Change tracking and delta generation
//! - Conflict detection and resolution
//! - Three-way merge for data sync
//! - Operation queue management
//! - Version vector clocks

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

/// Operation type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum OperationType {
    Create,
    Update,
    Delete,
    Patch,
}

/// Sync status
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SyncStatus {
    Pending,
    Syncing,
    Synced,
    Conflict,
    Error,
}

/// Conflict resolution strategy
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ConflictStrategy {
    /// Server version wins
    ServerWins,
    /// Client version wins
    ClientWins,
    /// Latest timestamp wins
    LastWriteWins,
    /// Merge non-conflicting fields
    MergeFields,
    /// Manual resolution required
    Manual,
}

/// Change record
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChangeRecord {
    pub id: String,
    pub entity_type: String,
    pub entity_id: String,
    pub operation: OperationType,
    pub data: Option<Value>,
    pub previous_data: Option<Value>,
    pub timestamp: String,
    pub client_id: String,
    pub version: u64,
    pub status: SyncStatus,
    pub retry_count: u32,
    pub error_message: Option<String>,
}

/// Conflict record
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConflictRecord {
    pub id: String,
    pub entity_type: String,
    pub entity_id: String,
    pub local_version: ChangeRecord,
    pub server_version: Value,
    pub base_version: Option<Value>,
    pub resolved: bool,
    pub resolution: Option<Value>,
    pub resolution_strategy: Option<ConflictStrategy>,
    pub created_at: String,
    pub resolved_at: Option<String>,
}

/// Sync result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncResult {
    pub success: bool,
    pub synced_count: u32,
    pub conflict_count: u32,
    pub error_count: u32,
    pub conflicts: Vec<ConflictRecord>,
    pub errors: Vec<String>,
}

/// Delta (changes between versions)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Delta {
    pub added: HashMap<String, Value>,
    pub modified: HashMap<String, FieldChange>,
    pub removed: Vec<String>,
}

/// Field change
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FieldChange {
    pub old_value: Value,
    pub new_value: Value,
}

/// Version vector for tracking causality
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VersionVector {
    pub versions: HashMap<String, u64>,
}

/// Initialize the offline module (called from core init)
fn offline_init() {
    console_error_panic_hook::set_once();
}

/// Generate unique change ID
#[wasm_bindgen]
pub fn generate_change_id() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};

    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_millis())
        .unwrap_or(0);

    let random: u32 = (timestamp as u32).wrapping_mul(1103515245).wrapping_add(12345);

    format!("chg_{}_{:08x}", timestamp, random)
}

/// Create change record
#[wasm_bindgen]
pub fn create_change_record(
    entity_type: &str,
    entity_id: &str,
    operation: &str,
    data: JsValue,
    previous_data: JsValue,
    client_id: &str,
    version: u64,
) -> JsValue {
    let op = match operation.to_lowercase().as_str() {
        "create" => OperationType::Create,
        "update" => OperationType::Update,
        "delete" => OperationType::Delete,
        "patch" => OperationType::Patch,
        _ => OperationType::Update,
    };

    let data_val: Option<Value> = serde_wasm_bindgen::from_value(data).ok();
    let prev_val: Option<Value> = serde_wasm_bindgen::from_value(previous_data).ok();

    let record = ChangeRecord {
        id: generate_change_id(),
        entity_type: entity_type.to_string(),
        entity_id: entity_id.to_string(),
        operation: op,
        data: data_val,
        previous_data: prev_val,
        timestamp: Utc::now().to_rfc3339(),
        client_id: client_id.to_string(),
        version,
        status: SyncStatus::Pending,
        retry_count: 0,
        error_message: None,
    };

    serde_wasm_bindgen::to_value(&record).unwrap_or(JsValue::NULL)
}

/// Calculate delta between two versions
#[wasm_bindgen]
pub fn calculate_delta(old_version: JsValue, new_version: JsValue) -> JsValue {
    let old: Value = match serde_wasm_bindgen::from_value(old_version) {
        Ok(v) => v,
        Err(_) => return JsValue::NULL,
    };

    let new: Value = match serde_wasm_bindgen::from_value(new_version) {
        Ok(v) => v,
        Err(_) => return JsValue::NULL,
    };

    let delta = calculate_delta_internal(&old, &new);
    serde_wasm_bindgen::to_value(&delta).unwrap_or(JsValue::NULL)
}

fn calculate_delta_internal(old: &Value, new: &Value) -> Delta {
    let mut added = HashMap::new();
    let mut modified = HashMap::new();
    let mut removed = Vec::new();

    match (old, new) {
        (Value::Object(old_obj), Value::Object(new_obj)) => {
            // Find added and modified
            for (key, new_val) in new_obj {
                if let Some(old_val) = old_obj.get(key) {
                    if old_val != new_val {
                        modified.insert(
                            key.clone(),
                            FieldChange {
                                old_value: old_val.clone(),
                                new_value: new_val.clone(),
                            },
                        );
                    }
                } else {
                    added.insert(key.clone(), new_val.clone());
                }
            }

            // Find removed
            for key in old_obj.keys() {
                if !new_obj.contains_key(key) {
                    removed.push(key.clone());
                }
            }
        }
        _ => {}
    }

    Delta { added, modified, removed }
}

/// Apply delta to a value
#[wasm_bindgen]
pub fn apply_delta(base: JsValue, delta: JsValue) -> JsValue {
    let mut base_val: Value = match serde_wasm_bindgen::from_value(base) {
        Ok(v) => v,
        Err(_) => return JsValue::NULL,
    };

    let delta_val: Delta = match serde_wasm_bindgen::from_value(delta) {
        Ok(d) => d,
        Err(_) => return JsValue::NULL,
    };

    if let Value::Object(ref mut obj) = base_val {
        // Apply additions
        for (key, value) in delta_val.added {
            obj.insert(key, value);
        }

        // Apply modifications
        for (key, change) in delta_val.modified {
            obj.insert(key, change.new_value);
        }

        // Apply removals
        for key in delta_val.removed {
            obj.remove(&key);
        }
    }

    serde_wasm_bindgen::to_value(&base_val).unwrap_or(JsValue::NULL)
}

/// Detect conflicts between local and server versions
#[wasm_bindgen]
pub fn detect_conflict(
    local: JsValue,
    server: JsValue,
    base: JsValue,
) -> JsValue {
    let local_val: Value = serde_wasm_bindgen::from_value(local).unwrap_or(Value::Null);
    let server_val: Value = serde_wasm_bindgen::from_value(server).unwrap_or(Value::Null);
    let base_val: Option<Value> = serde_wasm_bindgen::from_value(base).ok();

    // Calculate deltas from base
    let local_delta = base_val.as_ref()
        .map(|b| calculate_delta_internal(b, &local_val));
    let server_delta = base_val.as_ref()
        .map(|b| calculate_delta_internal(b, &server_val));

    let has_conflict = match (&local_delta, &server_delta) {
        (Some(ld), Some(sd)) => {
            // Check for overlapping modifications
            for key in ld.modified.keys() {
                if sd.modified.contains_key(key) || sd.removed.contains(key) {
                    return serde_wasm_bindgen::to_value(&serde_json::json!({
                        "has_conflict": true,
                        "conflicting_fields": [key],
                        "local_delta": ld,
                        "server_delta": sd
                    })).unwrap_or(JsValue::NULL);
                }
            }

            // Check if local adds conflict with server
            for key in ld.added.keys() {
                if sd.added.contains_key(key) {
                    return serde_wasm_bindgen::to_value(&serde_json::json!({
                        "has_conflict": true,
                        "conflicting_fields": [key],
                        "local_delta": ld,
                        "server_delta": sd
                    })).unwrap_or(JsValue::NULL);
                }
            }

            false
        }
        _ => local_val != server_val,
    };

    let result = serde_json::json!({
        "has_conflict": has_conflict,
        "conflicting_fields": [],
        "local_delta": local_delta,
        "server_delta": server_delta
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Resolve conflict using specified strategy
#[wasm_bindgen]
pub fn resolve_conflict(
    local: JsValue,
    server: JsValue,
    base: JsValue,
    strategy: &str,
) -> JsValue {
    let local_val: Value = serde_wasm_bindgen::from_value(local).unwrap_or(Value::Null);
    let server_val: Value = serde_wasm_bindgen::from_value(server).unwrap_or(Value::Null);
    let base_val: Option<Value> = serde_wasm_bindgen::from_value(base).ok();

    let resolved = match strategy.to_lowercase().as_str() {
        "server_wins" => server_val,
        "client_wins" => local_val,
        "last_write_wins" => {
            // Compare timestamps if available
            let local_ts = extract_timestamp(&local_val);
            let server_ts = extract_timestamp(&server_val);

            match (local_ts, server_ts) {
                (Some(l), Some(s)) if l > s => local_val,
                _ => server_val,
            }
        }
        "merge_fields" | "merge" => {
            merge_values(&local_val, &server_val, base_val.as_ref())
        }
        _ => server_val, // Default to server wins
    };

    serde_wasm_bindgen::to_value(&resolved).unwrap_or(JsValue::NULL)
}

fn extract_timestamp(value: &Value) -> Option<String> {
    value.as_object()
        .and_then(|obj| obj.get("updated_at").or_else(|| obj.get("modified_at")))
        .and_then(|v| v.as_str())
        .map(|s| s.to_string())
}

fn merge_values(local: &Value, server: &Value, base: Option<&Value>) -> Value {
    match (local, server) {
        (Value::Object(local_obj), Value::Object(server_obj)) => {
            let mut merged = server_obj.clone();

            // Get base for comparison
            let base_obj = base.and_then(|b| b.as_object());

            for (key, local_val) in local_obj {
                let server_val = server_obj.get(key);
                let base_val = base_obj.and_then(|b| b.get(key));

                match (server_val, base_val) {
                    // Server doesn't have this field - use local
                    (None, _) => {
                        merged.insert(key.clone(), local_val.clone());
                    }
                    // Field unchanged on server from base - use local
                    (Some(sv), Some(bv)) if sv == bv => {
                        merged.insert(key.clone(), local_val.clone());
                    }
                    // Field unchanged on local from base - keep server
                    (Some(_), Some(bv)) if local_val == bv => {
                        // Keep server value (already in merged)
                    }
                    // Both changed - recursive merge for objects, server wins for primitives
                    (Some(sv), _) => {
                        if local_val.is_object() && sv.is_object() {
                            merged.insert(key.clone(), merge_values(local_val, sv, base_val));
                        }
                        // For primitives, server value is already in merged
                    }
                }
            }

            Value::Object(merged)
        }
        // For non-objects, prefer server
        _ => server.clone(),
    }
}

/// Three-way merge
#[wasm_bindgen]
pub fn three_way_merge(
    base: JsValue,
    local: JsValue,
    server: JsValue,
) -> JsValue {
    let base_val: Value = serde_wasm_bindgen::from_value(base).unwrap_or(Value::Null);
    let local_val: Value = serde_wasm_bindgen::from_value(local).unwrap_or(Value::Null);
    let server_val: Value = serde_wasm_bindgen::from_value(server).unwrap_or(Value::Null);

    let merged = merge_values(&local_val, &server_val, Some(&base_val));

    let local_delta = calculate_delta_internal(&base_val, &local_val);
    let server_delta = calculate_delta_internal(&base_val, &server_val);

    // Check for actual conflicts (same field modified differently)
    let mut conflicts: Vec<String> = Vec::new();
    for key in local_delta.modified.keys() {
        if let Some(server_change) = server_delta.modified.get(key) {
            if let Some(local_change) = local_delta.modified.get(key) {
                if local_change.new_value != server_change.new_value {
                    conflicts.push(key.clone());
                }
            }
        }
    }

    let result = serde_json::json!({
        "merged": merged,
        "has_conflicts": !conflicts.is_empty(),
        "conflicting_fields": conflicts,
        "local_changes": local_delta,
        "server_changes": server_delta
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Version vector operations
impl VersionVector {
    pub fn new() -> Self {
        VersionVector {
            versions: HashMap::new(),
        }
    }

    pub fn increment(&mut self, node_id: &str) {
        let counter = self.versions.entry(node_id.to_string()).or_insert(0);
        *counter += 1;
    }

    pub fn merge(&mut self, other: &VersionVector) {
        for (node_id, version) in &other.versions {
            let current = self.versions.entry(node_id.clone()).or_insert(0);
            *current = (*current).max(*version);
        }
    }

    pub fn happens_before(&self, other: &VersionVector) -> bool {
        let mut dominated = false;

        for (node_id, version) in &self.versions {
            match other.versions.get(node_id) {
                Some(other_version) if version > other_version => return false,
                Some(other_version) if version < other_version => dominated = true,
                None if *version > 0 => return false,
                _ => {}
            }
        }

        for (node_id, version) in &other.versions {
            if !self.versions.contains_key(node_id) && *version > 0 {
                dominated = true;
            }
        }

        dominated
    }

    pub fn concurrent(&self, other: &VersionVector) -> bool {
        !self.happens_before(other) && !other.happens_before(self)
    }
}

/// Create version vector
#[wasm_bindgen]
pub fn create_version_vector() -> JsValue {
    let vv = VersionVector::new();
    serde_wasm_bindgen::to_value(&vv).unwrap_or(JsValue::NULL)
}

/// Increment version vector
#[wasm_bindgen]
pub fn increment_version_vector(vv: JsValue, node_id: &str) -> JsValue {
    let mut version_vector: VersionVector = match serde_wasm_bindgen::from_value(vv) {
        Ok(v) => v,
        Err(_) => VersionVector::new(),
    };

    version_vector.increment(node_id);
    serde_wasm_bindgen::to_value(&version_vector).unwrap_or(JsValue::NULL)
}

/// Merge version vectors
#[wasm_bindgen]
pub fn merge_version_vectors(vv1: JsValue, vv2: JsValue) -> JsValue {
    let mut v1: VersionVector = serde_wasm_bindgen::from_value(vv1).unwrap_or_else(|_| VersionVector::new());
    let v2: VersionVector = serde_wasm_bindgen::from_value(vv2).unwrap_or_else(|_| VersionVector::new());

    v1.merge(&v2);
    serde_wasm_bindgen::to_value(&v1).unwrap_or(JsValue::NULL)
}

/// Compare version vectors
#[wasm_bindgen]
pub fn compare_version_vectors(vv1: JsValue, vv2: JsValue) -> JsValue {
    let v1: VersionVector = serde_wasm_bindgen::from_value(vv1).unwrap_or_else(|_| VersionVector::new());
    let v2: VersionVector = serde_wasm_bindgen::from_value(vv2).unwrap_or_else(|_| VersionVector::new());

    let result = serde_json::json!({
        "v1_happens_before_v2": v1.happens_before(&v2),
        "v2_happens_before_v1": v2.happens_before(&v1),
        "concurrent": v1.concurrent(&v2),
        "equal": !v1.happens_before(&v2) && !v2.happens_before(&v1) && !v1.concurrent(&v2)
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

/// Sort changes for optimal sync order
#[wasm_bindgen]
pub fn sort_changes_for_sync(changes: JsValue) -> JsValue {
    let mut change_list: Vec<ChangeRecord> = match serde_wasm_bindgen::from_value(changes) {
        Ok(c) => c,
        Err(_) => return JsValue::NULL,
    };

    // Sort by: 1. Creates first, 2. Updates, 3. Deletes last
    // Within same type, sort by timestamp
    change_list.sort_by(|a, b| {
        let type_order = |op: &OperationType| match op {
            OperationType::Create => 0,
            OperationType::Update | OperationType::Patch => 1,
            OperationType::Delete => 2,
        };

        let a_order = type_order(&a.operation);
        let b_order = type_order(&b.operation);

        if a_order != b_order {
            a_order.cmp(&b_order)
        } else {
            a.timestamp.cmp(&b.timestamp)
        }
    });

    serde_wasm_bindgen::to_value(&change_list).unwrap_or(JsValue::NULL)
}

/// Calculate sync statistics
#[wasm_bindgen]
pub fn calculate_sync_stats(changes: JsValue) -> JsValue {
    let change_list: Vec<ChangeRecord> = match serde_wasm_bindgen::from_value(changes) {
        Ok(c) => c,
        Err(_) => return JsValue::NULL,
    };

    let total = change_list.len();
    let pending = change_list.iter().filter(|c| c.status == SyncStatus::Pending).count();
    let synced = change_list.iter().filter(|c| c.status == SyncStatus::Synced).count();
    let conflicts = change_list.iter().filter(|c| c.status == SyncStatus::Conflict).count();
    let errors = change_list.iter().filter(|c| c.status == SyncStatus::Error).count();

    let creates = change_list.iter().filter(|c| c.operation == OperationType::Create).count();
    let updates = change_list.iter().filter(|c| c.operation == OperationType::Update || c.operation == OperationType::Patch).count();
    let deletes = change_list.iter().filter(|c| c.operation == OperationType::Delete).count();

    let result = serde_json::json!({
        "total": total,
        "by_status": {
            "pending": pending,
            "synced": synced,
            "conflicts": conflicts,
            "errors": errors
        },
        "by_operation": {
            "creates": creates,
            "updates": updates,
            "deletes": deletes
        },
        "sync_progress": if total > 0 {
            (synced as f64 / total as f64 * 100.0).round()
        } else {
            100.0
        }
    });

    serde_wasm_bindgen::to_value(&result).unwrap_or(JsValue::NULL)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_delta_calculation() {
        let old = serde_json::json!({"name": "John", "age": 30});
        let new = serde_json::json!({"name": "Jane", "age": 30, "city": "NYC"});

        let delta = calculate_delta_internal(&old, &new);

        assert!(delta.modified.contains_key("name"));
        assert!(delta.added.contains_key("city"));
        assert!(delta.removed.is_empty());
    }

    #[test]
    fn test_version_vector() {
        let mut v1 = VersionVector::new();
        v1.increment("node1");
        v1.increment("node1");

        let mut v2 = VersionVector::new();
        v2.increment("node1");
        v2.increment("node2");

        assert!(v1.concurrent(&v2));
    }

    #[test]
    fn test_merge_values() {
        let base = serde_json::json!({"a": 1, "b": 2});
        let local = serde_json::json!({"a": 10, "b": 2, "c": 3});
        let server = serde_json::json!({"a": 1, "b": 20});

        let merged = merge_values(&local, &server, Some(&base));

        // a: local changed to 10, server unchanged -> local wins (10)
        // b: local unchanged, server changed to 20 -> server wins (20)
        // c: added by local -> local (3)
        assert_eq!(merged.get("a"), Some(&serde_json::json!(10)));
        assert_eq!(merged.get("b"), Some(&serde_json::json!(20)));
        assert_eq!(merged.get("c"), Some(&serde_json::json!(3)));
    }
}
