/**
 * Audit & Compliance domain module.
 *
 * Entities (verified live 2026-04-26 against the JWT monolith):
 *   - compliance-rules: ComplianceService.ListComplianceRules. Mounted at
 *     `/core.audit.compliance.api.v1.ComplianceService/...` (note the
 *     `core.audit.` package prefix — audit is a core platform module,
 *     not a business domain). Response shape `{base, rules,
 *     pagination: {totalCount}}` — proto-confirmed via
 *     core/audit/compliance/proto/compliance.proto.
 */
import type { DomainModule } from '../index.js';

export const audit: DomainModule = {
  id: 'audit',
  label: 'Audit & Compliance',
  entities: [
    {
      slug: 'compliance-rules',
      label: 'Compliance Rules',
      formId: 'access_log_report',
      listEndpoint: '/core.audit.compliance.api.v1.ComplianceService/ListComplianceRules',
      responseRowsKey: 'rules',
      responseTotalKey: 'pagination.totalCount',
      columns: ['id', 'ruleCode', 'name', 'status', 'severity'],
    },
  ],
};
