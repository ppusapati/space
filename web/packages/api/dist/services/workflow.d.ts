/**
 * Workflow Service Factories
 * Typed ConnectRPC clients for workflow, approval, escalation
 */
import type { Client } from '@connectrpc/connect';
import { WorkflowService } from '@samavāya/proto/gen/core/workflow/workflow/proto/workflow_pb.js';
import { ApprovalService } from '@samavāya/proto/gen/core/workflow/approval/proto/approval_pb.js';
import { EscalationService } from '@samavāya/proto/gen/core/workflow/escalation/proto/escalation_pb.js';
import { FormBuilder } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/formbuilder_pb.js';
import { formSubmission as FormSubmissionService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/forminstance_pb.js';
import { FormStateMachineService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/form_state_machine_pb.js';
import { ApprovalService as FormBuilderApprovalService } from '@samavāya/proto/gen/core/workflow/formbuilder/proto/approval_pb.js';
export { WorkflowService, ApprovalService, EscalationService, FormBuilder, FormSubmissionService, FormStateMachineService, FormBuilderApprovalService, };
export declare function getWorkflowService(): Client<typeof WorkflowService>;
export declare function getApprovalService(): Client<typeof ApprovalService>;
export declare function getEscalationService(): Client<typeof EscalationService>;
/** Typed client for FormBuilder (form definitions, schema management) */
export declare function getFormBuilderService(): Client<typeof FormBuilder>;
/** Typed client for FormSubmissionService (form instances, submissions) */
export declare function getFormSubmissionService(): Client<typeof FormSubmissionService>;
/** Typed client for FormStateMachineService (form state transitions) */
export declare function getFormStateMachineService(): Client<typeof FormStateMachineService>;
/** Typed client for FormBuilderApprovalService (form-level approvals) — aliased to avoid conflict with workflow ApprovalService */
export declare function getFormBuilderApprovalService(): Client<typeof FormBuilderApprovalService>;
//# sourceMappingURL=workflow.d.ts.map