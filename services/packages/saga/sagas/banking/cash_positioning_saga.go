// Package banking provides saga handlers for banking module workflows
package banking

import (
	"errors"

	"p9e.in/samavaya/packages/saga"
)

// CashPositioningSaga implements SAGA-B03: Cash Positioning & Forecasting workflow
// Business Flow: CollectCashData → AggregatePositions → CalculateLiquidity → ForecastCashFlow → ValidateForecasts → UpdatePositionMaster → GenerateReports → ArchivePositionData
// Timeout: 90 seconds, Critical steps: 1,2,3,5,8
type CashPositioningSaga struct {
	steps []*saga.StepDefinition
}

// NewCashPositioningSaga creates a new Cash Positioning & Forecasting saga handler
func NewCashPositioningSaga() saga.SagaHandler {
	return &CashPositioningSaga{
		steps: []*saga.StepDefinition{
			// Step 1: Collect Cash Data
			{
				StepNumber:    1,
				ServiceName:   "cash-management",
				HandlerMethod: "CollectCashData",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
				},
				TimeoutSeconds: 20,
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
			// Step 2: Aggregate Positions
			{
				StepNumber:    2,
				ServiceName:   "banking",
				HandlerMethod: "AggregatePositions",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"cashData":             "$.steps.1.result.cash_data",
				},
				TimeoutSeconds: 20,
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
			// Step 3: Calculate Liquidity
			{
				StepNumber:    3,
				ServiceName:   "cash-management",
				HandlerMethod: "CalculateLiquidity",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"aggregatedPositions":  "$.steps.2.result.aggregated_positions",
				},
				TimeoutSeconds: 20,
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
			// Step 4: Forecast Cash Flow
			{
				StepNumber:    4,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ForecastCashFlow",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"liquidityMetrics":     "$.steps.3.result.liquidity_metrics",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{104},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 5: Validate Forecasts
			{
				StepNumber:    5,
				ServiceName:   "cash-management",
				HandlerMethod: "ValidateForecasts",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"forecastData":         "$.steps.4.result.forecast_data",
				},
				TimeoutSeconds: 20,
				IsCritical:     true,
				CompensationSteps: []int32{105},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 6: Update Position Master
			{
				StepNumber:    6,
				ServiceName:   "banking",
				HandlerMethod: "UpdatePositionMaster",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"validatedForecasts":   "$.steps.5.result.validated_forecasts",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
				CompensationSteps: []int32{106},
				RetryConfig: &saga.RetryConfiguration{
					MaxRetries:        3,
					InitialBackoffMs:  1000,
					MaxBackoffMs:      30000,
					BackoffMultiplier: 2.0,
					JitterFraction:    0.1,
				},
			},
			// Step 7: Generate Reports
			{
				StepNumber:    7,
				ServiceName:   "general-ledger",
				HandlerMethod: "GenerateReports",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"validatedForecasts":   "$.steps.5.result.validated_forecasts",
				},
				TimeoutSeconds: 20,
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
			// Step 8: Archive Position Data
			{
				StepNumber:    8,
				ServiceName:   "accounts-receivable",
				HandlerMethod: "ArchivePositionData",
				InputMapping: map[string]string{
					"tenantID":             "$.tenantID",
					"companyID":            "$.companyID",
					"branchID":             "$.branchID",
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"currency":             "$.input.currency",
					"aggregatedPositions":  "$.steps.2.result.aggregated_positions",
					"reportData":           "$.steps.7.result.report_data",
				},
				TimeoutSeconds: 20,
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

			// Step 102: ReversePositionAggregation (compensates step 2)
			{
				StepNumber:    102,
				ServiceName:   "banking",
				HandlerMethod: "ReversePositionAggregation",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"aggregatedPositions":  "$.steps.2.result.aggregated_positions",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 103: ClearLiquidityCalculation (compensates step 3)
			{
				StepNumber:    103,
				ServiceName:   "cash-management",
				HandlerMethod: "ClearLiquidityCalculation",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"liquidityMetrics":     "$.steps.3.result.liquidity_metrics",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 104: ReverseCashFlowForecast (compensates step 4)
			{
				StepNumber:    104,
				ServiceName:   "accounts-payable",
				HandlerMethod: "ReverseCashFlowForecast",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"forecastData":         "$.steps.4.result.forecast_data",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 105: ClearForecastValidation (compensates step 5)
			{
				StepNumber:    105,
				ServiceName:   "cash-management",
				HandlerMethod: "ClearForecastValidation",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"validatedForecasts":   "$.steps.5.result.validated_forecasts",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 106: ReversePositionUpdate (compensates step 6)
			{
				StepNumber:    106,
				ServiceName:   "banking",
				HandlerMethod: "ReversePositionUpdate",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"validatedForecasts":   "$.steps.5.result.validated_forecasts",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
			// Step 107: ClearReports (compensates step 7)
			{
				StepNumber:    107,
				ServiceName:   "general-ledger",
				HandlerMethod: "ClearReports",
				InputMapping: map[string]string{
					"forecastPeriod":       "$.input.forecast_period",
					"consolidationLevel":   "$.input.consolidation_level",
					"reportData":           "$.steps.7.result.report_data",
				},
				TimeoutSeconds: 20,
				IsCritical:     false,
			},
		},
	}
}

// SagaType returns the saga type identifier
func (s *CashPositioningSaga) SagaType() string {
	return "SAGA-B03"
}

// GetStepDefinitions returns all step definitions (forward + compensation)
func (s *CashPositioningSaga) GetStepDefinitions() []*saga.StepDefinition {
	return s.steps
}

// GetStepDefinition returns a specific step definition by step number
func (s *CashPositioningSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	for _, step := range s.steps {
		if step.StepNumber == int32(stepNum) {
			return step
		}
	}
	return nil
}

// ValidateInput validates the saga input parameters
func (s *CashPositioningSaga) ValidateInput(input interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return errors.New("invalid input type")
	}

	if inputMap["forecast_period"] == nil {
		return errors.New("forecast_period is required")
	}

	forecastPeriod, ok := inputMap["forecast_period"].(string)
	if !ok || forecastPeriod == "" {
		return errors.New("forecast_period must be a non-empty string")
	}

	if inputMap["consolidation_level"] == nil {
		return errors.New("consolidation_level is required")
	}

	consolidationLevel, ok := inputMap["consolidation_level"].(string)
	if !ok || consolidationLevel == "" {
		return errors.New("consolidation_level must be a non-empty string")
	}

	if inputMap["currency"] == nil {
		return errors.New("currency is required")
	}

	currency, ok := inputMap["currency"].(string)
	if !ok || currency == "" {
		return errors.New("currency must be a non-empty string")
	}

	return nil
}
