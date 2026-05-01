package contract

// The harness normalises every port's concrete ClassSummary and
// ClassDefinition types into these canonical shapes before comparing
// them. Callers supply small projection functions (one-liners) from
// their port types into these canonical structs. This keeps the
// harness free of per-port type imports and makes drift between the
// two adapters the ONLY thing a comparison can detect.

// ClassSummary is the canonical compact form used during comparison.
type ClassSummary struct {
	Name         string
	Label        string
	Description  string
	Industries   []string
	HasProcesses bool
}

// ClassDefinition is the canonical full schema used during comparison.
type ClassDefinition struct {
	Domain           string
	Name             string
	Label            string
	Description      string
	Industries       []string
	Attributes       map[string]AttributeDefinition
	ComplianceChecks []string
	CapacityMetrics  []string
	Processes        []string
}

// AttributeDefinition is one attribute's declared rules — canonical shape.
type AttributeDefinition struct {
	Kind        string
	Required    bool
	Min         *float64
	Max         *float64
	Values      []string
	Lookup      string
	Pattern     string
	Description string
}
