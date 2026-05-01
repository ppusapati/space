// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// RouteOptimizationSchedulingSaga implements SAGA-SC06: Route Optimization & Scheduling
// Business Flow: InitializeRouteOptimization → CollectDeliveryRequests → AnalyzeRouteParameters → ComputeOptimalRoutes → AssignVehicles → GenerateSchedule → ValidateCapacity → PostRouteJournals → CompleteRouteOptimization
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,9
type RouteOptimizationSchedulingSaga struct {
	steps []*saga.StepDefinition
}

// NewRouteOptimizationSchedulingSaga creates a new Route Optimization & Scheduling saga handler
func NewRouteOptimizationSchedulingSaga() saga.SagaHandler {
	return &RouteOptimizationSchedulingSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initialize Route Optimization
			{
				StepNumber:    1,
				ServiceName:   "route-optimization",
				HandlerMethod: "InitializeRouteOptimization",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"routeID":          "$.input.route_id",
					"optimizationDate": "$.input.optimization_date",
					"vehicleCount":     "$.input.vehicle_count",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
				CompensationSteps: []int32{},
			},
			// Step 2: Collect Delivery Requests
			{
				StepNumber:    2,
				ServiceName:   "logistics",
				HandlerMethod: "CollectDeliveryRequests",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"routeID":            "$.input.route_id",
					"optimizationDate":   "$.input.optimization_date",
					"optimizationData":   "$.steps.1.result.optimization_data",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{102},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 3: Analyze Route Parameters
			{
				StepNumber:    3,
				ServiceName:   "route-optimization",
				HandlerMethod: "AnalyzeRouteParameters",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"routeID":           "$.input.route_id",
					"deliveryRequests":  "$.steps.2.result.delivery_requests",
					"vehicleCount":      "$.input.vehicle_count",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{103},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 4: Compute Optimal Routes
			{
				StepNumber:    4,
				ServiceName:   "route-optimization",
				HandlerMethod: "ComputeOptimalRoutes",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"routeID":             "$.input.route_id",
					"routeParameters":     "$.steps.3.result.route_parameters",
					"deliveryRequests":    "$.steps.2.result.delivery_requests",
				},
				TimeoutSeconds: 60,
				IsCritical:     true,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Assign Vehicles
			{
				StepNumber:    5,
				ServiceName:   "warehouse",
				HandlerMethod: "AssignVehicles",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"routeID":         "$.input.route_id",
					"vehicleCount":    "$.input.vehicle_count",
					"optimalRoutes":   "$.steps.4.result.optimal_routes",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Generate Schedule
			{
				StepNumber:    6,
				ServiceName:   "logistics",
				HandlerMethod: "GenerateSchedule",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"routeID":         "$.input.route_id",
					"optimizationDate": "$.input.optimization_date",
					"vehicleAssignments": "$.steps.5.result.vehicle_assignments",
					"optimalRoutes":   "$.steps.4.result.optimal_routes",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Validate Capacity
			{
				StepNumber:    7,
				ServiceName:   "route-optimization",
				HandlerMethod: "ValidateCapacity",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"routeID":           "$.input.route_id",
					"schedule":          "$.steps.6.result.schedule",
					"vehicleAssignments": "$.steps.5.result.vehicle_assignments",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{107},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 8: Post Route Journals
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostRouteJournals",
				InputMapping: map[string]string{
					"tenantID":        "$.tenantID",
					"companyID":       "$.companyID",
					"branchID":        "$.branchID",
					"routeID":         "$.input.route_id",
					"optimizationDate": "$.input.optimization_date",
					"schedule":        "$.steps.6.result.schedule",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Complete Route Optimization
			{
				StepNumber:    9,
				ServiceName:   "route-optimization",
				HandlerMethod: "CompleteRouteOptimization",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"routeID":           "$.input.route_id",
					"schedule":          "$.steps.6.result.schedule",
					"capacityValidation": "$.steps.7.result.capacity_validation",
					"journalEntries":    "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     true,
				CompensationSteps: []int32{},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// ===== COMPENSATION STEPS =====

			// Step 102: ReverseDeliveryCollection (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "logistics",
				HandlerMethod: "ReverseDeliveryCollection",
				InputMapping: map[string]string{
					"routeID":           "$.input.route_id",
					"deliveryRequests":  "$.steps.2.result.delivery_requests",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: ReverseRouteAnalysis (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "route-optimization",
				HandlerMethod: "ReverseRouteAnalysis",
				InputMapping: map[string]string{
					"routeID":         "$.input.route_id",
					"routeParameters": "$.steps.3.result.route_parameters",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseOptimalComputation (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "route-optimization",
				HandlerMethod: "ReverseOptimalComputation",
				InputMapping: map[string]string{
					"routeID":       "$.input.route_id",
					"optimalRoutes": "$.steps.4.result.optimal_routes",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseVehicleAssignment (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "warehouse",
				HandlerMethod: "ReverseVehicleAssignment",
				InputMapping: map[string]string{
					"routeID":            "$.input.route_id",
					"vehicleAssignments": "$.steps.5.result.vehicle_assignments",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseScheduleGeneration (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "logistics",
				HandlerMethod: "ReverseScheduleGeneration",
				InputMapping: map[string]string{
					"routeID": "$.input.route_id",
					"schedule": "$.steps.6.result.schedule",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReverseCapacityValidation (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "route-optimization",
				HandlerMethod: "ReverseCapacityValidation",
				InputMapping: map[string]string{
					"routeID":             "$.input.route_id",
					"capacityValidation":  "$.steps.7.result.capacity_validation",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReverseRouteJournals (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReverseRouteJournals",
				InputMapping: map[string]string{
					"routeID":        "$.input.route_id",
					"journalEntries": "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: CancelRouteOptimizationCompletion (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "route-optimization",
				HandlerMethod: "CancelRouteOptimizationCompletion",
				InputMapping: map[string]string{
					"routeID": "$.input.route_id",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *RouteOptimizationSchedulingSaga) SagaType() string {
	return "SAGA-SC06"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *RouteOptimizationSchedulingSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *RouteOptimizationSchedulingSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *RouteOptimizationSchedulingSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["route_id"] == nil {
		return errors.New("route_id is required")
	}

	routeID, ok := inputMap["route_id"].(string)
	if !ok || routeID == "" {
		return errors.New("route_id must be a non-empty string")
	}

	if inputMap["optimization_date"] == nil {
		return errors.New("optimization_date is required")
	}

	optimizationDate, ok := inputMap["optimization_date"].(string)
	if !ok || optimizationDate == "" {
		return errors.New("optimization_date must be a non-empty string")
	}

	if inputMap["vehicle_count"] == nil {
		return errors.New("vehicle_count is required")
	}

	vehicleCount, ok := inputMap["vehicle_count"].(string)
	if !ok || vehicleCount == "" {
		return errors.New("vehicle_count must be a non-empty string")
	}

	return nil
}
