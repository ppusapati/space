// Package supplychain provides saga handlers for supply chain module workflows
package supplychain

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// SupplierPerformanceCollaborationSaga implements SAGA-SC08: Supplier Collaboration & Performance
// Business Flow: InitializeEvaluation → CollectPerformanceMetrics → AssessQualityScores → AnalyzeDeliveryPerformance → EvaluateComplianceMetrics → CalculatePerformanceRating → GeneratePerformanceReport → PostPerformanceJournals → CompleteSupplierEvaluation
// Timeout: 180 seconds, Critical steps: 1,2,3,4,6,8,10
type SupplierPerformanceCollaborationSaga struct {
	steps []*saga.StepDefinition
}

// NewSupplierPerformanceCollaborationSaga creates a new Supplier Collaboration & Performance saga handler
func NewSupplierPerformanceCollaborationSaga() saga.SagaHandler {
	return &SupplierPerformanceCollaborationSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Initialize Evaluation
			{
				StepNumber:    1,
				ServiceName:   "supplier-management",
				HandlerMethod: "InitializeEvaluation",
				InputMapping: map[string]string{
					"tenantID":          "$.tenantID",
					"companyID":         "$.companyID",
					"branchID":          "$.branchID",
					"supplierID":        "$.input.supplier_id",
					"evaluationPeriod":  "$.input.evaluation_period",
					"evaluationType":    "$.input.evaluation_type",
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
			// Step 2: Collect Performance Metrics
			{
				StepNumber:    2,
				ServiceName:   "procurement",
				HandlerMethod: "CollectPerformanceMetrics",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"supplierID":          "$.input.supplier_id",
					"evaluationPeriod":    "$.input.evaluation_period",
					"evaluationData":      "$.steps.1.result.evaluation_data",
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
			// Step 3: Assess Quality Scores
			{
				StepNumber:    3,
				ServiceName:   "quality-inspection",
				HandlerMethod: "AssessQualityScores",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"supplierID":         "$.input.supplier_id",
					"evaluationPeriod":   "$.input.evaluation_period",
					"performanceMetrics": "$.steps.2.result.performance_metrics",
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
			// Step 4: Analyze Delivery Performance
			{
				StepNumber:    4,
				ServiceName:   "logistics",
				HandlerMethod: "AnalyzeDeliveryPerformance",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"supplierID":         "$.input.supplier_id",
					"evaluationPeriod":   "$.input.evaluation_period",
					"performanceMetrics": "$.steps.2.result.performance_metrics",
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
			// Step 5: Evaluate Compliance Metrics
			{
				StepNumber:    5,
				ServiceName:   "procurement",
				HandlerMethod: "EvaluateComplianceMetrics",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"supplierID":         "$.input.supplier_id",
					"evaluationPeriod":   "$.input.evaluation_period",
					"performanceMetrics": "$.steps.2.result.performance_metrics",
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
			// Step 6: Calculate Performance Rating
			{
				StepNumber:    6,
				ServiceName:   "supplier-management",
				HandlerMethod: "CalculatePerformanceRating",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"supplierID":         "$.input.supplier_id",
					"qualityScores":      "$.steps.3.result.quality_scores",
					"deliveryPerformance": "$.steps.4.result.delivery_performance",
					"complianceMetrics":  "$.steps.5.result.compliance_metrics",
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
			// Step 7: Generate Performance Report
			{
				StepNumber:    7,
				ServiceName:   "supplier-management",
				HandlerMethod: "GeneratePerformanceReport",
				InputMapping: map[string]string{
					"tenantID":         "$.tenantID",
					"companyID":        "$.companyID",
					"branchID":         "$.branchID",
					"supplierID":       "$.input.supplier_id",
					"evaluationPeriod": "$.input.evaluation_period",
					"performanceRating": "$.steps.6.result.performance_rating",
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
			// Step 8: Post Performance Journals
			{
				StepNumber:    8,
				ServiceName:   "general-ledger",
				HandlerMethod: "PostPerformanceJournals",
				InputMapping: map[string]string{
					"tenantID":           "$.tenantID",
					"companyID":          "$.companyID",
					"branchID":           "$.branchID",
					"supplierID":         "$.input.supplier_id",
					"evaluationPeriod":   "$.input.evaluation_period",
					"performanceRating":  "$.steps.6.result.performance_rating",
				},
				TimeoutSeconds: 45,
				IsCritical:     true,
				CompensationSteps: []int32{108},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 9: Complete Supplier Evaluation
			{
				StepNumber:    9,
				ServiceName:   "supplier-management",
				HandlerMethod: "CompleteSupplierEvaluation",
				InputMapping: map[string]string{
					"tenantID":            "$.tenantID",
					"companyID":           "$.companyID",
					"branchID":            "$.branchID",
					"supplierID":          "$.input.supplier_id",
					"evaluationPeriod":    "$.input.evaluation_period",
					"performanceReport":   "$.steps.7.result.performance_report",
					"journalEntries":      "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
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

			// Step 102: ReverseMetricsCollection (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseMetricsCollection",
				InputMapping: map[string]string{
					"supplierID":         "$.input.supplier_id",
					"performanceMetrics": "$.steps.2.result.performance_metrics",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 103: ReverseQualityAssessment (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "quality-inspection",
				HandlerMethod: "ReverseQualityAssessment",
				InputMapping: map[string]string{
					"supplierID":    "$.input.supplier_id",
					"qualityScores": "$.steps.3.result.quality_scores",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 104: ReverseDeliveryAnalysis (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "logistics",
				HandlerMethod: "ReverseDeliveryAnalysis",
				InputMapping: map[string]string{
					"supplierID":         "$.input.supplier_id",
					"deliveryPerformance": "$.steps.4.result.delivery_performance",
				},
				TimeoutSeconds: 60,
				IsCritical:     false,
			},
			// Step 105: ReverseComplianceEvaluation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "procurement",
				HandlerMethod: "ReverseComplianceEvaluation",
				InputMapping: map[string]string{
					"supplierID":        "$.input.supplier_id",
					"complianceMetrics": "$.steps.5.result.compliance_metrics",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 106: ReverseRatingCalculation (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "supplier-management",
				HandlerMethod: "ReverseRatingCalculation",
				InputMapping: map[string]string{
					"supplierID":        "$.input.supplier_id",
					"performanceRating": "$.steps.6.result.performance_rating",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 107: ReversePerformanceReport (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "supplier-management",
				HandlerMethod: "ReversePerformanceReport",
				InputMapping: map[string]string{
					"supplierID":       "$.input.supplier_id",
					"performanceReport": "$.steps.7.result.performance_report",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 108: ReversePerformanceJournals (compensates step 8)
			{
				StepNumber:    108,
				ServiceName:   "general-ledger",
				HandlerMethod: "ReversePerformanceJournals",
				InputMapping: map[string]string{
					"supplierID":      "$.input.supplier_id",
					"journalEntries":  "$.steps.8.result.journal_entries",
				},
				TimeoutSeconds: 45,
				IsCritical:     false,
			},
			// Step 109: CancelSupplierEvaluationCompletion (compensates step 9)
			{
				StepNumber:    109,
				ServiceName:   "supplier-management",
				HandlerMethod: "CancelSupplierEvaluationCompletion",
				InputMapping: map[string]string{
					"supplierID":       "$.input.supplier_id",
					"evaluationPeriod": "$.input.evaluation_period",
				},
				TimeoutSeconds: 30,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *SupplierPerformanceCollaborationSaga) SagaType() string {
	return "SAGA-SC08"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *SupplierPerformanceCollaborationSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *SupplierPerformanceCollaborationSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *SupplierPerformanceCollaborationSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["supplier_id"] == nil {
		return errors.New("supplier_id is required")
	}

	supplierID, ok := inputMap["supplier_id"].(string)
	if !ok || supplierID == "" {
		return errors.New("supplier_id must be a non-empty string")
	}

	if inputMap["evaluation_period"] == nil {
		return errors.New("evaluation_period is required")
	}

	evaluationPeriod, ok := inputMap["evaluation_period"].(string)
	if !ok || evaluationPeriod == "" {
		return errors.New("evaluation_period must be a non-empty string")
	}

	if inputMap["evaluation_type"] == nil {
		return errors.New("evaluation_type is required")
	}

	evaluationType, ok := inputMap["evaluation_type"].(string)
	if !ok || evaluationType == "" {
		return errors.New("evaluation_type must be a non-empty string")
	}

	return nil
}
