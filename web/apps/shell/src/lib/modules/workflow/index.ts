/**
 * Workflow domain module.
 *
 * Surfaces every standalone List RPC across the 4 workflow services
 * that don't require a parent_id (approval / escalation / formbuilder.approval
 * / workflow). Sub-resource Lists requiring a parent_id are intentionally
 * NOT wired:
 *   - approval/ListPendingApprovals — needs approver_id
 *
 * One entity is gated on a postgres-owner migration:
 *   - formbuilder/ListForms (FormBuilder service) returns 500
 *     `column "friendly_endpoint" does not exist (SQLSTATE 42703)` until
 *     migration 000223 applies. Tracked as deferred-work in
 *     docs/BASE_DOMAIN_AUDITS.md. NOT wired here — would surface a broken
 *     entry in the FE menu.
 *
 * Every (formId, listEndpoint, responseRowsKey) triple is verified live
 * against the JWT monolith on 2026-04-29:
 *   - formId → FormService.GetFormSchema returns 200 with formDefinition
 *   - listEndpoint → returns 200 + rows array under empty-tenant context
 *   - responseRowsKey matches the proto field name of the rows array
 *
 * Form catalog reuse: the workflow module catalog has 17 forms. 7 cover
 * the 9 entities here — multiple entities share a form_id where the
 * functional area overlaps (e.g. all 3 escalation entities share
 * `escalation_policy` because they're three views into the same domain).
 * ListPage uses explicit `columns` per entity for the row shape so the
 * create-form-fields mismatch doesn't show up in list views.
 *
 * Wire-shape note: every workflow response carries both `totalCount`
 * (int32) AND `nextPageToken` (string). The loader reads totalCount
 * directly — token-based pagination would need a different loader
 * shape (loader takes `pageSize` + `pageOffset` today, not a token).
 * If a future use-case needs token pagination, extend the loader to
 * support both modes; do NOT shoehorn token state into pageOffset.
 */
import type { DomainModule } from '../index.js';

export const workflow: DomainModule = {
  id: 'workflow',
  label: 'Workflow',
  entities: [
    // ---- escalation ----
    {
      slug: 'escalation-rules',
      label: 'Escalation Rules',
      formId: 'escalation_policy',
      listEndpoint: '/workflow.escalation.api.v1.EscalationService/ListRules',
      responseRowsKey: 'rules',
      responseTotalKey: 'totalCount',
      columns: ['ruleCode', 'ruleName', 'triggerSource', 'thresholdMinutes', 'isActive'],
    },
    {
      slug: 'escalation-chains',
      label: 'Escalation Chains',
      formId: 'escalation_policy',
      listEndpoint: '/workflow.escalation.api.v1.EscalationService/ListChains',
      responseRowsKey: 'chains',
      responseTotalKey: 'totalCount',
      columns: ['chainCode', 'chainName', 'description', 'isActive'],
    },
    {
      slug: 'escalation-triggers',
      label: 'Escalation Triggers',
      formId: 'escalation_policy',
      listEndpoint: '/workflow.escalation.api.v1.EscalationService/ListTriggers',
      responseRowsKey: 'triggers',
      responseTotalKey: 'totalCount',
      columns: ['triggerNumber', 'entityType', 'source', 'status', 'currentLevel'],
    },
    // ---- approval (legacy stage-chain engine) ----
    {
      slug: 'department-levels',
      label: 'Department Approval Levels',
      formId: 'approval_routing_configuration',
      listEndpoint: '/workflow.approval.api.v1.ApprovalService/ListDepartmentLevels',
      responseRowsKey: 'levels',
      responseTotalKey: 'totalCount',
      columns: ['department', 'levelName', 'coreLevel', 'description', 'amountThreshold'],
    },
    // ---- formbuilder.approval (form-flavored approval) ----
    {
      slug: 'approval-instances',
      label: 'Approval Instances',
      formId: 'workflow_approval',
      listEndpoint: '/workflow.formbuilder.api.v1.ApprovalService/ListApprovalInstances',
      responseRowsKey: 'instances',
      responseTotalKey: 'totalCount',
      columns: ['entityType', 'entityId', 'requestedBy', 'assignedTo', 'status', 'priority'],
    },
    {
      slug: 'approval-delegates',
      label: 'Approval Delegates',
      formId: 'workflow_approval',
      listEndpoint: '/workflow.formbuilder.api.v1.ApprovalService/ListDelegates',
      responseRowsKey: 'delegates',
      responseTotalKey: 'totalCount',
      columns: ['delegatorId', 'delegateId', 'startDate', 'endDate', 'isActive', 'reason'],
    },
    // ---- workflow (BPMN engine — Layer 1) ----
    {
      slug: 'workflow-definitions',
      label: 'Workflow Definitions',
      formId: 'process_definition',
      listEndpoint: '/workflow.workflow.api.v1.WorkflowService/ListWorkflowDefinitions',
      responseRowsKey: 'definitions',
      responseTotalKey: 'totalCount',
      columns: ['key', 'version', 'name', 'category', 'status'],
    },
    {
      slug: 'workflow-executions',
      label: 'Workflow Executions',
      formId: 'workflow_execution',
      listEndpoint: '/workflow.workflow.api.v1.WorkflowService/ListWorkflowExecutions',
      responseRowsKey: 'executions',
      responseTotalKey: 'totalCount',
      columns: ['definitionKey', 'definitionVersion', 'entityType', 'status', 'startTime', 'endTime'],
    },
    {
      slug: 'workflow-tasks',
      label: 'Workflow Tasks',
      formId: 'task_assignment',
      listEndpoint: '/workflow.workflow.api.v1.WorkflowService/ListTasks',
      responseRowsKey: 'tasks',
      responseTotalKey: 'totalCount',
      columns: ['name', 'taskType', 'status', 'assigneeId', 'description'],
    },
  ],
};
