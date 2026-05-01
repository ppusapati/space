/**
 * Budget domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - budgets: BudgetService.ListBudgets under `/budget.budget.api.v1.`.
 *     Lives in `core/budget/budget/` (core platform). Proto exposes both
 *     `pagination: Pagination` and `total_count: int32` — the nested form
 *     is canonical here.
 */
import type { DomainModule } from '../index.js';

export const budget: DomainModule = {
  id: 'budget',
  label: 'Budget',
  entities: [
    {
      slug: 'budgets',
      label: 'Budgets',
      formId: 'form_annual_budget_creation',
      listEndpoint: '/budget.budget.api.v1.BudgetService/ListBudgets',
      responseRowsKey: 'budgets',
      responseTotalKey: 'pagination.totalCount',
      columns: ['id', 'budgetCode', 'name', 'budgetType', 'status', 'version'],
    },
  ],
};
