//go:build saga_examples

// This file is an illustrative example of a cross-module land-acquisition
// saga. It's gated by the `saga_examples` build tag because it references
// fields on StepDefinition / RetryConfiguration that don't match the
// canonical models.StepDefinition shape (StepNum → StepNumber,
// Handler → HandlerMethod, etc.). Exclude from regular builds. Fix the
// field mapping + drop the tag when the example is brought current.
// B.8 quarantine 2026-04-19.
package sagas

import (
	"context"
	"fmt"

	"p9e.in/samavaya/packages/saga"
	"p9e.in/samavaya/packages/saga/models"
)

// ============================================================================
// SAGA: Land Acquisition Saga
// Orchestrates multi-service workflow for acquiring land for renewable projects
// ============================================================================

// LandAcquisitionSaga defines the complete land acquisition process
// across government leases, right-of-way, grid interconnection, and finance
type LandAcquisitionSaga struct{}

// ============================================================================
// SAGA DEFINITION
// ============================================================================

// SagaType returns the unique saga identifier
func (s *LandAcquisitionSaga) SagaType() string {
	return "SAGA-LAND-ACQUISITION"
}

// GetStepDefinitions returns all steps in the saga workflow
func (s *LandAcquisitionSaga) GetStepDefinitions() []*saga.StepDefinition {
	return []*saga.StepDefinition{
		// Step 1: Create Government Lease
		{
			StepNum:      1,
			Name:         "CreateGovernmentLease",
			ServiceName:  "govt-lease-service",
			Handler:      "CreateGovernmentLease",
			Timeout:      30, // seconds
			RetryConfig: &saga.RetryConfiguration{
				MaxAttempts: 3,
				BackoffMs:   1000,
			},
			Compensation: "CancelLease",
		},

		// Step 2: Create Right-of-Way
		{
			StepNum:      2,
			Name:         "CreateRightOfWay",
			ServiceName:  "right-of-way-service",
			Handler:      "CreateRightOfWay",
			Timeout:      30,
			RetryConfig: &saga.RetryConfiguration{
				MaxAttempts: 3,
				BackoffMs:   1000,
			},
			Compensation: "CancelRightOfWay",
		},

		// Step 3: Submit Grid Interconnection Application
		{
			StepNum:      3,
			Name:         "SubmitGridApplication",
			ServiceName:  "grid-interconnection-service",
			Handler:      "SubmitInterconnectionApplication",
			Timeout:      60, // Longer timeout for system studies
			RetryConfig: &saga.RetryConfiguration{
				MaxAttempts: 3,
				BackoffMs:   2000,
			},
			Compensation: "CancelGridApplication",
		},

		// Step 4: Create Project Finance Structure
		{
			StepNum:      4,
			Name:         "CreateProjectFinance",
			ServiceName:  "renewable-energy-finance-service",
			Handler:      "CreateProjectFinance",
			Timeout:      30,
			RetryConfig: &saga.RetryConfiguration{
				MaxAttempts: 3,
				BackoffMs:   1000,
			},
			Compensation: "CancelProjectFinance",
		},

		// Step 5: Link Lease Asset to Finance
		{
			StepNum:      5,
			Name:         "LinkLeaseAsset",
			ServiceName:  "renewable-energy-finance-service",
			Handler:      "LinkGovernmentLeaseAsset",
			Timeout:      20,
			RetryConfig: &saga.RetryConfiguration{
				MaxAttempts: 2,
				BackoffMs:   500,
			},
			Compensation: "UnlinkLeaseAsset",
		},
	}
}

// GetStepDefinition returns a specific step definition by step number
func (s *LandAcquisitionSaga) GetStepDefinition(stepNum int) *saga.StepDefinition {
	steps := s.GetStepDefinitions()
	for _, step := range steps {
		if step.StepNum == stepNum {
			return step
		}
	}
	return nil
}

// ============================================================================
// INPUT VALIDATION
// ============================================================================

// LandAcquisitionInput defines the input for this saga
type LandAcquisitionInput struct {
	// Project Information
	ProjectID       string
	ProjectName     string
	ProjectCapacity float64 // MW

	// Government Lease Information
	MinistryName       string
	AllocationArea     float64 // hectares
	AllocationPeriod   int     // years
	RentPerHectare     float64
	SecurityDeposit    float64
	LeaseStartDate     string // ISO 8601 format
	LeaseEndDate       string // ISO 8601 format

	// Right-of-Way Information
	TransmissionLineID string
	RowLength          float64 // km
	RowWidth           float64 // meters
	AcquisitionMethod  string  // negotiation, eminent_domain, lease

	// Grid Interconnection Information
	ConnectionVoltage      float64 // kV
	PointOfInterconnection string  // Substation name

	// Finance Information
	TotalProjectCost   float64 // Currency amount
	EquityPercentage   float64 // 0-100
	DebtPercentage     float64 // 0-100
	ProjectLifespan    int     // years
}

// ValidateInput validates the saga input
func (s *LandAcquisitionSaga) ValidateInput(input interface{}) error {
	sagaInput, ok := input.(*LandAcquisitionInput)
	if !ok {
		return fmt.Errorf("invalid input type: expected *LandAcquisitionInput")
	}

	// Validate required fields
	if sagaInput.ProjectID == "" {
		return fmt.Errorf("ProjectID is required")
	}

	if sagaInput.ProjectName == "" {
		return fmt.Errorf("ProjectName is required")
	}

	if sagaInput.ProjectCapacity <= 0 {
		return fmt.Errorf("ProjectCapacity must be greater than 0")
	}

	if sagaInput.MinistryName == "" {
		return fmt.Errorf("MinistryName is required")
	}

	if sagaInput.AllocationArea <= 0 {
		return fmt.Errorf("AllocationArea must be greater than 0")
	}

	if sagaInput.AllocationPeriod <= 0 {
		return fmt.Errorf("AllocationPeriod must be greater than 0")
	}

	if sagaInput.SecurityDeposit < 0 {
		return fmt.Errorf("SecurityDeposit cannot be negative")
	}

	if sagaInput.TotalProjectCost <= 0 {
		return fmt.Errorf("TotalProjectCost must be greater than 0")
	}

	if sagaInput.EquityPercentage < 0 || sagaInput.EquityPercentage > 100 {
		return fmt.Errorf("EquityPercentage must be between 0 and 100")
	}

	if sagaInput.DebtPercentage < 0 || sagaInput.DebtPercentage > 100 {
		return fmt.Errorf("DebtPercentage must be between 0 and 100")
	}

	if sagaInput.EquityPercentage+sagaInput.DebtPercentage != 100 {
		return fmt.Errorf("EquityPercentage + DebtPercentage must equal 100")
	}

	if sagaInput.ProjectLifespan <= 0 {
		return fmt.Errorf("ProjectLifespan must be greater than 0")
	}

	return nil
}

// ============================================================================
// SAGA BENEFITS & WORKFLOW NOTES
// ============================================================================

/*

SAGA EXECUTION WORKFLOW:

1. Orchestrator receives LandAcquisitionInput
   └─ Validates input via ValidateInput()
   └─ Creates SagaExecution record
   └─ Publishes SagaStarted event

2. Step 1: CreateGovernmentLease (govt-lease-service)
   └─ Returns: leaseID
   └─ On success: Publishes StepCompleted event
   └─ On failure: Executes CancelLease compensation

3. Step 2: CreateRightOfWay (right-of-way-service)
   └─ Uses leaseID from Step 1
   └─ Returns: rowID
   └─ On failure: Executes CancelROW compensation

4. Step 3: SubmitGridApplication (grid-interconnection-service)
   └─ Returns: applicationID
   └─ On failure: Executes CancelApplication compensation

5. Step 4: CreateProjectFinance (renewable-energy-finance-service)
   └─ Returns: projectFinanceID
   └─ On failure: Executes CancelProjectFinance compensation

6. Step 5: LinkLeaseAsset (renewable-energy-finance-service)
   └─ Links leaseID to projectFinanceID
   └─ On failure: Executes UnlinkLeaseAsset compensation

7. Saga Completion (if all steps succeed)
   └─ Publishes SagaCompleted event
   └─ Returns SagaExecution with all step results

FAILURE SCENARIO (e.g., Step 4 fails after 3 retries):
   └─ Starts compensation phase
   └─ Executes compensations in REVERSE order:
      └─ Step 3: CancelGridApplication
      └─ Step 2: CancelRightOfWay
      └─ Step 1: CancelLease
   └─ Step 4 & 5 skipped (never completed)
   └─ Publishes CompensationCompleted event
   └─ Publishes SagaFailed event
   └─ Returns failure status

BENEFITS:
   - ATOMICITY: All services succeed or all rollback
   - CONSISTENCY: ACID within services, eventual consistency across
   - ISOLATION: Each service independent, no direct coupling
   - DURABILITY: Saga state persisted, resumable on failure
   - OBSERVABILITY: Event trail at each step, correlation IDs
   - RESILIENCE: Retry logic, timeouts, circuit breakers

*/
