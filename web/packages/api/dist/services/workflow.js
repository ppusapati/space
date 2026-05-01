/**
 * Workflow Service Factories
 * Typed ConnectRPC clients for workflow, approval, escalation
 */
import { getApiClient } from '../client/client.js';
import { WorkflowService } from '@samavāya/proto/gen/core/workflow/workflow/proto/workflow_pb.js';
import { ApprovalService } from '@samavāya/proto/gen/core/workflow/approval/proto/approval_pb.js';
import { EscalationService } from '@samavāya/proto/gen/core/workflow/escalation/proto/escalation_pb.js';
import { FormBuilder } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/formbuilder_pb.js';
import { formSubmission as FormSubmissionService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/forminstance_pb.js';
import { FormStateMachineService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/form_state_machine_pb.js';
import { ApprovalService as FormBuilderApprovalService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/approval_pb.js';
export { WorkflowService, ApprovalService, EscalationService, FormBuilder, FormSubmissionService, FormStateMachineService, FormBuilderApprovalService, };
export function getWorkflowService() {
    return getApiClient().getService(WorkflowService);
}
export function getApprovalService() {
    return getApiClient().getService(ApprovalService);
}
export function getEscalationService() {
    return getApiClient().getService(EscalationService);
}
// ─── FormBuilder ────────────────────────────────────────────────────────────
/** Typed client for FormBuilder (form definitions, schema management) */
export function getFormBuilderService() {
    return getApiClient().getService(FormBuilder);
}
/** Typed client for FormSubmissionService (form instances, submissions) */
export function getFormSubmissionService() {
    return getApiClient().getService(FormSubmissionService);
}
/** Typed client for FormStateMachineService (form state transitions) */
export function getFormStateMachineService() {
    return getApiClient().getService(FormStateMachineService);
}
/** Typed client for FormBuilderApprovalService (form-level approvals) — aliased to avoid conflict with workflow ApprovalService */
export function getFormBuilderApprovalService() {
    return getApiClient().getService(FormBuilderApprovalService);
}
//# sourceMappingURL=workflow.js.map