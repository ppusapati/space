/**
 * HR domain module.
 *
 * Entities:
 *   - employees: HR employee master. Form_id `form_hr_employee_onboarding`
 *     is the canonical onboarding form (see backend
 *     core/workflow/formbuilder/cmd/generate_forms/configs/hr/employee_onboarding.yaml).
 *     List endpoint /hr.employee.api.v1.EmployeeService/ListEmployees
 *     returns `{employees: [...], pagination: {totalCount: N}}` — the
 *     responseTotalKey uses a dotted path because the count lives in
 *     the nested Pagination message (matches the canonical
 *     packages/api/v1/pagination.Pagination shape used by HR + Finance).
 */
import type { DomainModule } from '../index.js';

export const hr: DomainModule = {
  id: 'hr',
  label: 'HR',
  entities: [
    {
      slug: 'employees',
      label: 'Employees',
      formId: 'form_hr_employee_onboarding',
      listEndpoint: '/hr.employee.api.v1.EmployeeService/ListEmployees',
      responseRowsKey: 'employees',
      responseTotalKey: 'pagination.totalCount',
      columns: ['employeeCode', 'firstName', 'lastName', 'email', 'status'],
    },
  ],
};
