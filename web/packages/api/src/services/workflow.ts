/**
 * Workflow Service Factories
 * Typed ConnectRPC clients for workflow, approval, escalation
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { WorkflowService } from '@samavāya/proto/gen/core/workflow/workflow/proto/workflow_pb.js';
import {
  ApprovalService,
  ApprovalStatus,
} from '@samavāya/proto/gen/core/workflow/approval/proto/approval_pb.js';
import type {
  ApprovalRequest,
  ApprovalStage,
  ApprovalAction,
  ListPendingApprovalsRequest,
  ListPendingApprovalsResponse,
  ApproveStageRequest,
  RejectStageRequest,
  GetApprovalHistoryRequest,
  GetApprovalHistoryResponse,
} from '@samavāya/proto/gen/core/workflow/approval/proto/approval_pb.js';
import { EscalationService } from '@samavāya/proto/gen/core/workflow/escalation/proto/escalation_pb.js';
import { FormBuilder } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/formbuilder_pb.js';
import { formSubmission as FormSubmissionService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/forminstance_pb.js';
// 2026-04-27 (Phase 5 of engine-unification, backend/docs/BASE_DOMAIN_AUDITS.md):
// FormStateMachineService removed — the unified workflow.workflow engine
// owns form lifecycles via templates + ApprovalTaskOrchestrator.
import { ApprovalService as FormBuilderApprovalService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/approval_pb.js';

export {
  WorkflowService, ApprovalService, ApprovalStatus, EscalationService,
  FormBuilder, FormSubmissionService, FormBuilderApprovalService,
};

export type {
  ApprovalRequest,
  ApprovalStage,
  ApprovalAction,
  ListPendingApprovalsRequest,
  ListPendingApprovalsResponse,
  ApproveStageRequest,
  RejectStageRequest,
  GetApprovalHistoryRequest,
  GetApprovalHistoryResponse,
};

export function getWorkflowService(): Client<typeof WorkflowService> {
  return getApiClient().getService(WorkflowService);
}

export function getApprovalService(): Client<typeof ApprovalService> {
  return getApiClient().getService(ApprovalService);
}

export function getEscalationService(): Client<typeof EscalationService> {
  return getApiClient().getService(EscalationService);
}

// ─── FormBuilder ────────────────────────────────────────────────────────────

/** Typed client for FormBuilder (form definitions, schema management) */
export function getFormBuilderService(): Client<typeof FormBuilder> {
  return getApiClient().getService(FormBuilder);
}

/** Typed client for FormSubmissionService (form instances, submissions) */
export function getFormSubmissionService(): Client<typeof FormSubmissionService> {
  return getApiClient().getService(FormSubmissionService);
}

/** Typed client for FormBuilderApprovalService (form-level approvals) — aliased to avoid conflict with workflow ApprovalService */
export function getFormBuilderApprovalService(): Client<typeof FormBuilderApprovalService> {
  return getApiClient().getService(FormBuilderApprovalService);
}
