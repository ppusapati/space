# Migration Guide: Unified Expression Package

## Overview
This document describes the consolidation of filtering, rule evaluation, and search criteria logic into a single `packages/expression` package.

## What Changed

### Before (Scattered Across Modules)
```
packages/models/search_criteria.go        → SearchCriteria, Filter, FilterOperator types
workflow/formbuilder/internal/service/rule_evaluator.go  → RuleEvaluator interface & implementation
packages/database/pgxpostgres/builder/    → SQL filter building & operators
```

### After (Unified in packages/expression)
```
packages/expression/doc.go                → Package documentation
packages/expression/types.go              → SearchCriteria, Filter, FilterOperator types
packages/expression/evaluator.go          → Evaluator interface & implementation (formerly RuleEvaluator)
packages/expression/evaluator_test.go     → Comprehensive test coverage
```

## Migration Steps for Modules

### Step 1: Update Imports

**Old:**
```go
import (
    "p9e.in/samavaya/packages/models"
    // custom RuleEvaluator in module
)

criteria := &models.SearchCriteria{...}
evaluator := NewRuleEvaluator()
```

**New:**
```go
import "p9e.in/samavaya/packages/expression"

criteria := &expression.SearchCriteria{...}
evaluator := expression.NewEvaluator()
```

### Step 2: Update Type References

| Old | New |
|-----|-----|
| `models.SearchCriteria` | `expression.SearchCriteria` |
| `models.Filter` | `expression.Filter` |
| `models.DynamicFilter` | `expression.DynamicFilter` |
| `models.FilterOperator` | `expression.FilterOperator` |
| `models.OperatorEquals` | `expression.OperatorEquals` |
| `RuleEvaluator` (custom) | `expression.Evaluator` |
| `NewRuleEvaluator()` | `expression.NewEvaluator()` |

### Step 3: Update Method Calls

**Old RuleEvaluator interface:**
```go
ruleEvaluator.EvaluateRule(rule, data) (bool, error)
ruleEvaluator.EvaluateConditionalVisibility(hiddenWhen, data) (bool, error)
ruleEvaluator.EvaluateConditionalReadonly(readonlyWhen, data) (bool, error)
ruleEvaluator.ValidateRuleExpression(rule) error
```

**New Evaluator interface (same signatures, same behavior):**
```go
evaluator := expression.NewEvaluator()
evaluator.EvaluateRule(rule, data) (bool, error)
evaluator.EvaluateConditionalVisibility(hiddenWhen, data) (bool, error)
evaluator.EvaluateConditionalReadonly(readonlyWhen, data) (bool, error)
evaluator.ValidateRuleExpression(rule) error
evaluator.EvaluateBusinessRule(rule, fieldValue, rowData) error  // NEW
```

## New Capabilities

### SearchCriteria Helper Methods
```go
criteria := &expression.SearchCriteria{
    PageSize: 20,
    Sort: []string{"created_at"},
}

criteria.GetPageSize()        // Returns 20, defaults to 20 if not set
criteria.GetPageOffset()      // Returns 0-based offset
criteria.HasSort()            // Returns true if sorting configured
criteria.HasFilters()         // Returns true if any filters configured
criteria.IsSortDescending()   // Returns true if descending sort requested
```

### Extended Evaluator Methods
```go
evaluator := expression.NewEvaluator()

// Validate rule syntax before execution
err := evaluator.ValidateRuleExpression("amount > 1000 AND status == \"approved\"")

// Evaluate business rule for field validation
err := evaluator.EvaluateBusinessRule("value > 0", fieldValue, rowData)

// Conditional field visibility
isVisible, err := evaluator.EvaluateConditionalVisibility("status == \"rejected\"", data)

// Conditional field readonly
isReadonly, err := evaluator.EvaluateConditionalReadonly("status == \"approved\"", data)
```

## Operator Support

Both SQL filters and rule expressions support these operators:

| Category | Operators |
|----------|-----------|
| Comparison | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| Collections | `IN`, `NOT IN` |
| String | `CONTAINS`, `STARTS_WITH`, `ENDS_WITH`, `MATCHES` |
| Logical | `AND`, `OR` (with proper precedence) |

## Use Cases by Module

### FormBuilder
```go
import "p9e.in/samavaya/packages/expression"

evaluator := expression.NewEvaluator()

// Grid row validation
result, err := evaluator.EvaluateRule("quantity > 0 AND unit_price > 0", rowData)

// Conditional field visibility
visible, err := evaluator.EvaluateConditionalVisibility(field.HiddenWhen, rowData)

// Conditional field readonly
readonly, err := evaluator.EvaluateConditionalReadonly(field.ReadonlyWhen, rowData)
```

### MetaSearch
```go
import "p9e.in/samavaya/packages/expression"

// Parse search criteria
criteria := &expression.SearchCriteria{
    Filters: []expression.Filter{...},
    PageSize: 20,
    Sort: []string{"name", "created_at"},
}

// Use with database queries
// SQL query builder converts filters to WHERE clauses
```

### ReportBuilder
```go
import "p9e.in/samavaya/packages/expression"

evaluator := expression.NewEvaluator()

// Complex aggregation rules
result, err := evaluator.EvaluateRule(
    "total_amount > 10000 AND region IN (US,UK,CA) AND approved == true",
    reportData,
)
```

### Authorization (Future)
```go
import "p9e.in/samavaya/packages/expression"

evaluator := expression.NewEvaluator()

// Rule-based access control
canAccess, err := evaluator.EvaluateRule(
    "role IN (admin,manager) AND department == \"Finance\"",
    userContext,
)
```

## No Breaking Changes

The unified package maintains 100% API compatibility:
- All existing methods work identically
- Return types unchanged
- Error handling unchanged
- Operator semantics unchanged

## Deprecation Timeline

- **Phase 1 (Current)**: New `packages/expression` available, old `models.SearchCriteria` still works
- **Phase 2 (Future)**: Deprecation notice added to `models.search_criteria.go`
- **Phase 3 (Future)**: Old file removed after all modules migrated

## Testing

All existing tests pass with the unified package:
- 24+ test cases for rule evaluation
- 12+ test cases for conditional visibility/readonly
- 8+ test cases for grid row validation
- 5+ test cases for SearchCriteria helper methods

Run tests:
```bash
cd packages/expression
go test -v
```

## Questions?

For questions about the migration, refer to:
- [packages/expression/doc.go](doc.go) - Detailed package documentation
- [packages/expression/evaluator_test.go](evaluator_test.go) - Usage examples in tests
- [packages/expression/types.go](types.go) - Type definitions and helper methods
