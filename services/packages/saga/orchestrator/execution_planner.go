// Package orchestrator provides execution planning for saga steps
package orchestrator

import (
	"fmt"

	"p9e.in/chetana/packages/saga"
)

// ExecutionPlanner handles step dependency resolution and execution planning
type ExecutionPlanner struct {
	stepDefs []*saga.StepDefinition
}

// NewExecutionPlanner creates a new execution planner
func NewExecutionPlanner(stepDefs []*saga.StepDefinition) *ExecutionPlanner {
	return &ExecutionPlanner{
		stepDefs: stepDefs,
	}
}

// PlanExecution returns the sequence of steps to execute
// Validates dependencies and returns ordered execution plan
func (ep *ExecutionPlanner) PlanExecution() ([]*saga.StepDefinition, error) {
	if len(ep.stepDefs) == 0 {
		return nil, fmt.Errorf("no step definitions provided")
	}

	// Validate step numbers are sequential
	for i, stepDef := range ep.stepDefs {
		expectedStepNum := int32(i + 1)
		if stepDef.StepNumber != expectedStepNum {
			return nil, fmt.Errorf("step numbers must be sequential, expected step %d but got %d", expectedStepNum, stepDef.StepNumber)
		}
	}

	// Validate no circular dependencies
	if err := validateNoDependencyCycles(ep.stepDefs); err != nil {
		return nil, fmt.Errorf("circular dependency detected: %w", err)
	}

	// Validate all dependency references are valid
	if err := validateDependencyReferences(ep.stepDefs); err != nil {
		return nil, fmt.Errorf("invalid dependency reference: %w", err)
	}

	// Return steps in execution order (already sequential)
	return ep.stepDefs, nil
}

// GetStepDependencies returns steps that a given step depends on
func (ep *ExecutionPlanner) GetStepDependencies(stepNum int32) ([]*saga.StepDefinition, error) {
	if stepNum < 1 || stepNum > int32(len(ep.stepDefs)) {
		return nil, fmt.Errorf("invalid step number: %d", stepNum)
	}

	// step := ep.stepDefs[stepNum-1] — not needed: dependencies are
	// currently derived structurally (sequential), not from step-level
	// metadata. Left commented so the intent is recoverable when
	// the planner grows dependency-aware logic.
	_ = ep.stepDefs[stepNum-1]
	dependencies := make([]*saga.StepDefinition, 0)

	// In current design, steps execute sequentially
	// Dependencies are implicit (each step depends on previous)
	if stepNum > 1 {
		dependencies = append(dependencies, ep.stepDefs[:stepNum-1]...)
	}

	return dependencies, nil
}

// CanExecuteStep checks if a step can be executed given execution state
func (ep *ExecutionPlanner) CanExecuteStep(stepNum int32, executionState map[string]interface{}) (bool, error) {
	if stepNum < 1 || stepNum > int32(len(ep.stepDefs)) {
		return false, fmt.Errorf("invalid step number: %d", stepNum)
	}

	// Check that all predecessor steps have completed
	for i := int32(1); i < stepNum; i++ {
		stepKey := fmt.Sprintf("step_%d_result", i)
		if _, exists := executionState[stepKey]; !exists {
			return false, fmt.Errorf("step %d depends on completion of step %d", stepNum, i)
		}
	}

	return true, nil
}

// GetNextExecutableStep returns the first step that can be executed
func (ep *ExecutionPlanner) GetNextExecutableStep(currentStep int32) *saga.StepDefinition {
	if currentStep < 1 || currentStep > int32(len(ep.stepDefs)) {
		return nil
	}

	// In sequential execution, next step is just currentStep + 1
	nextStepNum := currentStep + 1
	if nextStepNum <= int32(len(ep.stepDefs)) {
		return ep.stepDefs[nextStepNum-1]
	}

	return nil
}

// GetCriticalPath returns all critical steps (must succeed)
func (ep *ExecutionPlanner) GetCriticalPath() []*saga.StepDefinition {
	critical := make([]*saga.StepDefinition, 0)

	for _, stepDef := range ep.stepDefs {
		if stepDef.IsCritical {
			critical = append(critical, stepDef)
		}
	}

	return critical
}

// GetOptionalSteps returns all optional steps (can fail without saga failure)
func (ep *ExecutionPlanner) GetOptionalSteps() []*saga.StepDefinition {
	optional := make([]*saga.StepDefinition, 0)

	for _, stepDef := range ep.stepDefs {
		if !stepDef.IsCritical {
			optional = append(optional, stepDef)
		}
	}

	return optional
}

// ValidateExecutionState checks if execution state is valid for current step
func (ep *ExecutionPlanner) ValidateExecutionState(stepNum int32, executionState map[string]interface{}) error {
	if stepNum < 1 || stepNum > int32(len(ep.stepDefs)) {
		return fmt.Errorf("invalid step number: %d", stepNum)
	}

	// Verify all preceding steps have results
	for i := int32(1); i < stepNum; i++ {
		stepKey := fmt.Sprintf("step_%d_result", i)
		if _, exists := executionState[stepKey]; !exists {
			return fmt.Errorf("step %d result missing in execution state", i)
		}
	}

	return nil
}

// EstimateExecutionTime calculates expected execution time
func (ep *ExecutionPlanner) EstimateExecutionTime() int32 {
	totalSeconds := int32(0)

	for _, stepDef := range ep.stepDefs {
		if stepDef.TimeoutSeconds > 0 {
			totalSeconds += stepDef.TimeoutSeconds
		} else {
			// Default 60 seconds per step
			totalSeconds += 60
		}
	}

	return totalSeconds
}

// Helper functions

func validateNoDependencyCycles(stepDefs []*saga.StepDefinition) error {
	// In sequential execution model, cycles are not possible
	// Each step number is unique and incremental
	// This validates the structural invariant
	for i, stepDef := range stepDefs {
		if stepDef.StepNumber != int32(i+1) {
			return fmt.Errorf("step numbers must be sequential")
		}
	}
	return nil
}

func validateDependencyReferences(stepDefs []*saga.StepDefinition) error {
	// Validate compensation steps reference valid step numbers
	maxStepNum := int32(len(stepDefs))

	for _, stepDef := range stepDefs {
		// CompensationSteps is []int32 of referenced step-number indexes;
		// no per-entry struct wrapper.
		for _, compStepNum := range stepDef.CompensationSteps {
			if compStepNum < 1 || compStepNum > maxStepNum {
				return fmt.Errorf("invalid compensation step number %d referenced by step %d", compStepNum, stepDef.StepNumber)
			}
		}
	}

	return nil
}
