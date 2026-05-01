# Expression Package - Unified Filtering & Rule Evaluation

## Quick Start

### Import
```go
import "p9e.in/samavaya/packages/expression"
```

### Create Evaluator
```go
evaluator := expression.NewEvaluator()
```

### Evaluate Rules
```go
// Simple rule
result, err := evaluator.EvaluateRule("amount > 1000", data)

// Complex rule with AND/OR
result, err := evaluator.EvaluateRule(
    "status == \"approved\" AND amount > 1000 OR priority == \"high\"",
    data,
)

// Validate syntax before execution
err := evaluator.ValidateRuleExpression(ruleString)
```

### Conditional Field Logic
```go
// Check if field should be visible
isVisible, err := evaluator.EvaluateConditionalVisibility(
    "status == \"rejected\"",  // hidden_when condition
    rowData,
)

// Check if field should be readonly
isReadonly, err := evaluator.EvaluateConditionalReadonly(
    "status == \"approved\"",  // readonly_when condition
    rowData,
)
```

### Search Criteria
```go
criteria := &expression.SearchCriteria{
    Filters: []expression.Filter{
        {
            Field:    "status",
            Operator: expression.OperatorEquals,
            Value:    "active",
        },
    },
    PageSize:   20,
    PageOffset: 0,
    Sort:       []string{"created_at"},
}

// Use helper methods
pageSize := criteria.GetPageSize()  // 20
if criteria.HasSort() {
    // Apply sorting
}
if criteria.HasFilters() {
    // Apply filters
}
```

## Operators

### Comparison
- `==` (equal)
- `!=` (not equal)
- `<` (less than)
- `<=` (less than or equal)
- `>` (greater than)
- `>=` (greater than or equal)

### Collections
- `IN` (value in list)
- `NOT IN` (value not in list)

### String
- `CONTAINS` (substring search)
- `STARTS_WITH` (prefix match)
- `ENDS_WITH` (suffix match)
- `MATCHES` (regex pattern)

### Logical
- `AND` (higher precedence)
- `OR` (lower precedence)

## Examples

### Grid Row Validation
```go
evaluator := expression.NewEvaluator()

// Validate quantity and price
rule := "quantity > 0 AND unit_price > 0"
result, err := evaluator.EvaluateRule(rule, rowData)
if !result {
    // Row validation failed
}
```

### Complex Filtering
```go
// Multiple conditions with precedence
rule := "status == \"pending\" AND (amount > 1000 OR priority == \"high\")"
result, err := evaluator.EvaluateRule(rule, data)
```

### Email Validation Rule
```go
rule := `email MATCHES ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
result, err := evaluator.EvaluateRule(rule, fieldData)
```

### Search with Multiple Filters
```go
criteria := &expression.SearchCriteria{
    Filters: []expression.Filter{
        {Field: "status", Operator: expression.OperatorEquals, Value: "active"},
        {Field: "region", Operator: expression.OperatorIn, Value: []interface{}{"US", "UK", "CA"}},
        {Field: "amount", Operator: expression.OperatorGreaterThan, Value: 1000},
    },
    PageSize: 50,
    Sort:     []string{"created_at", "name"},
}
```

## Filter Operators

For SQL query building, use FilterOperator constants:

```go
const (
    OperatorEquals     FilterOperator = "$eq"
    OperatorNotEquals  FilterOperator = "$neq"
    OperatorContains   FilterOperator = "$contains"
    OperatorStartsWith FilterOperator = "$starts_with"
    OperatorEndsWith   FilterOperator = "$ends_with"
    OperatorIn         FilterOperator = "$in"
    OperatorNotIn      FilterOperator = "$nin"
    OperatorIsNull     FilterOperator = "$null"
    OperatorIsNotNull  FilterOperator = "$nnull"
    OperatorIsEmpty    FilterOperator = "$empty"
    OperatorIsNotEmpty FilterOperator = "$nempty"
    OperatorLike       FilterOperator = "$like"

    OperatorGreaterThan       FilterOperator = "$gt"
    OperatorGreaterThanEquals FilterOperator = "$gte"
    OperatorLessThan          FilterOperator = "$lt"
    OperatorLessThanEquals    FilterOperator = "$lte"
)
```

## Data Type Support

The evaluator automatically converts types for comparison:

```go
// Works with integers
evaluator.EvaluateRule("amount > 1000", map[string]interface{}{"amount": 1500})

// Works with floats
evaluator.EvaluateRule("price > 99.99", map[string]interface{}{"price": 150.00})

// Works with strings
evaluator.EvaluateRule("status == \"approved\"", map[string]interface{}{"status": "approved"})

// Works with mixed types (auto-conversion)
evaluator.EvaluateRule("count > 5", map[string]interface{}{"count": "10"})  // "10" → 10
```

## Use Cases

### FormBuilder
- Grid row validation with cross-column rules
- Conditional field visibility/readonly
- Business rule enforcement

### MetaSearch
- Advanced filter expressions
- Dynamic filter generation
- Search criteria building

### ReportBuilder
- Complex aggregation rules
- Filter expressions in reports
- Conditional calculations

### Authorization (Future)
- Rule-based access control
- Permission evaluation
- Role-based conditions

## Error Handling

```go
// Missing field
result, err := evaluator.EvaluateRule("status == \"approved\"", map[string]interface{}{})
// err: "field not found: status"

// Invalid regex
result, err := evaluator.EvaluateRule("email MATCHES [invalid(regex", data)
// err: "invalid regex pattern: ..."

// Invalid syntax
err := evaluator.ValidateRuleExpression("status == \"approved\"(")
// err: "unbalanced parentheses in rule: ..."
```

## Testing

Run tests:
```bash
cd packages/expression
go test -v
go test -cover
```

Test examples in `evaluator_test.go`:
- Simple operators (==, !=, <, >, <=, >=)
- Collections (IN, NOT IN)
- String operations (CONTAINS, STARTS_WITH, ENDS_WITH, MATCHES)
- Logical operations (AND, OR with proper precedence)
- Complex nested conditions
- Conditional visibility/readonly
- Grid row validation scenarios
- SearchCriteria helper methods

## Migration from Old Approach

If you previously had a custom `RuleEvaluator`:

**Old:**
```go
evaluator := NewRuleEvaluator()
```

**New:**
```go
evaluator := expression.NewEvaluator()
```

All method signatures remain identical - it's a drop-in replacement!

See [MIGRATION.md](MIGRATION.md) for detailed migration guide.
