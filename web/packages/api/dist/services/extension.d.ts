/**
 * Extension Service Factories — Land Module
 * Typed ConnectRPC clients for all land extension services
 */
import type { Client } from '@connectrpc/connect';
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
export { LandComplianceService, DueDiligenceService, FieldOpsService, GISSpatialService, LandFinanceService, LandInsightsService, LandParcelService, LandWorkflowOrchestratorService, LegalCaseService, NegotiationService, RiskScoringService, StakeholderService, };
/** Typed client for ComplianceService (land regulatory compliance) — aliased to avoid conflict with audit ComplianceService */
export declare function getLandComplianceService(): Client<typeof LandComplianceService>;
/** Typed client for DueDiligenceService (land due diligence checks) */
export declare function getDueDiligenceService(): Client<typeof DueDiligenceService>;
/** Typed client for FieldOpsService (field operations management) */
export declare function getFieldOpsService(): Client<typeof FieldOpsService>;
/** Typed client for GISSpatialService (geographic information, spatial data) */
export declare function getGISSpatialService(): Client<typeof GISSpatialService>;
/** Typed client for LandFinanceService (land financial transactions) */
export declare function getLandFinanceService(): Client<typeof LandFinanceService>;
/** Typed client for LandInsightsService (land analytics, valuations) */
export declare function getLandInsightsService(): Client<typeof LandInsightsService>;
/** Typed client for LandParcelService (land parcel management) */
export declare function getLandParcelService(): Client<typeof LandParcelService>;
/** Typed client for LandWorkflowOrchestratorService (land workflow orchestration) */
export declare function getLandWorkflowService(): Client<typeof LandWorkflowOrchestratorService>;
/** Typed client for LegalCaseService (legal case tracking) */
export declare function getLegalCaseService(): Client<typeof LegalCaseService>;
/** Typed client for NegotiationService (land deal negotiations) */
export declare function getNegotiationService(): Client<typeof NegotiationService>;
/** Typed client for RiskScoringService (land risk assessment) */
export declare function getRiskScoringService(): Client<typeof RiskScoringService>;
/** Typed client for StakeholderService (stakeholder management) */
export declare function getStakeholderService(): Client<typeof StakeholderService>;
//# sourceMappingURL=extension.d.ts.map