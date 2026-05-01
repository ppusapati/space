/**
 * Extension Service Factories — Land Module
 * Typed ConnectRPC clients for all land extension services
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// Land services
import { ComplianceService as LandComplianceService } from '@samavāya/proto/gen/extension/land/compliance/proto/compliance_pb.js';
import { DueDiligenceService } from '@samavāya/proto/gen/extension/land/due-diligence/proto/due_diligence_pb.js';
import { FieldOpsService } from '@samavāya/proto/gen/extension/land/field-ops/proto/field_ops_pb.js';
import { GISSpatialService } from '@samavāya/proto/gen/extension/land/gis-spatial/proto/gis_spatial_pb.js';
import { LandFinanceService } from '@samavāya/proto/gen/extension/land/land-finance/proto/land_finance_pb.js';
import { LandInsightsService } from '@samavāya/proto/gen/extension/land/land-insights/proto/land_insights_pb.js';
import { LandParcelService } from '@samavāya/proto/gen/extension/land/land-parcel/proto/land_parcel_pb.js';
import { LandWorkflowOrchestratorService } from '@samavāya/proto/gen/extension/land/land-workflow-orchestrator/proto/land_workflow_pb.js';
import { LegalCaseService } from '@samavāya/proto/gen/extension/land/legal-case/proto/legal_case_pb.js';
import { NegotiationService } from '@samavāya/proto/gen/extension/land/negotiation/proto/negotiation_pb.js';
import { RiskScoringService } from '@samavāya/proto/gen/extension/land/risk-scoring/proto/risk_scoring_pb.js';
import { StakeholderService } from '@samavāya/proto/gen/extension/land/stakeholder/proto/stakeholder_pb.js';

export {
  LandComplianceService, DueDiligenceService, FieldOpsService, GISSpatialService,
  LandFinanceService, LandInsightsService, LandParcelService, LandWorkflowOrchestratorService,
  LegalCaseService, NegotiationService, RiskScoringService, StakeholderService,
};

// ─── Land Compliance ────────────────────────────────────────────────────────

/** Typed client for ComplianceService (land regulatory compliance) — aliased to avoid conflict with audit ComplianceService */
export function getLandComplianceService(): Client<typeof LandComplianceService> {
  return getApiClient().getService(LandComplianceService);
}

// ─── Due Diligence ──────────────────────────────────────────────────────────

/** Typed client for DueDiligenceService (land due diligence checks) */
export function getDueDiligenceService(): Client<typeof DueDiligenceService> {
  return getApiClient().getService(DueDiligenceService);
}

// ─── Field Ops ──────────────────────────────────────────────────────────────

/** Typed client for FieldOpsService (field operations management) */
export function getFieldOpsService(): Client<typeof FieldOpsService> {
  return getApiClient().getService(FieldOpsService);
}

// ─── GIS Spatial ────────────────────────────────────────────────────────────

/** Typed client for GISSpatialService (geographic information, spatial data) */
export function getGISSpatialService(): Client<typeof GISSpatialService> {
  return getApiClient().getService(GISSpatialService);
}

// ─── Land Finance ───────────────────────────────────────────────────────────

/** Typed client for LandFinanceService (land financial transactions) */
export function getLandFinanceService(): Client<typeof LandFinanceService> {
  return getApiClient().getService(LandFinanceService);
}

// ─── Land Insights ──────────────────────────────────────────────────────────

/** Typed client for LandInsightsService (land analytics, valuations) */
export function getLandInsightsService(): Client<typeof LandInsightsService> {
  return getApiClient().getService(LandInsightsService);
}

// ─── Land Parcel ────────────────────────────────────────────────────────────

/** Typed client for LandParcelService (land parcel management) */
export function getLandParcelService(): Client<typeof LandParcelService> {
  return getApiClient().getService(LandParcelService);
}

// ─── Land Workflow ──────────────────────────────────────────────────────────

/** Typed client for LandWorkflowOrchestratorService (land workflow orchestration) */
export function getLandWorkflowService(): Client<typeof LandWorkflowOrchestratorService> {
  return getApiClient().getService(LandWorkflowOrchestratorService);
}

// ─── Legal Case ─────────────────────────────────────────────────────────────

/** Typed client for LegalCaseService (legal case tracking) */
export function getLegalCaseService(): Client<typeof LegalCaseService> {
  return getApiClient().getService(LegalCaseService);
}

// ─── Negotiation ────────────────────────────────────────────────────────────

/** Typed client for NegotiationService (land deal negotiations) */
export function getNegotiationService(): Client<typeof NegotiationService> {
  return getApiClient().getService(NegotiationService);
}

// ─── Risk Scoring ───────────────────────────────────────────────────────────

/** Typed client for RiskScoringService (land risk assessment) */
export function getRiskScoringService(): Client<typeof RiskScoringService> {
  return getApiClient().getService(RiskScoringService);
}

// ─── Stakeholder ────────────────────────────────────────────────────────────

/** Typed client for StakeholderService (stakeholder management) */
export function getStakeholderService(): Client<typeof StakeholderService> {
  return getApiClient().getService(StakeholderService);
}
