/**
 * Banking domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - bank-accounts: BankingService.ListBankAccounts. Lives under
 *     `core/banking/banking/` (NOT business/banking — banking is a
 *     core/platform-style module). Proto declares both
 *     `pagination: Pagination` AND `total_count: int32` at the top level;
 *     we use `pagination.totalCount` for consistency with the canonical
 *     pattern, but `totalCount` (flat) would also work.
 */
import type { DomainModule } from '../index.js';

export const banking: DomainModule = {
  id: 'banking',
  label: 'Banking',
  entities: [
    {
      slug: 'bank-accounts',
      label: 'Bank Accounts',
      formId: 'account_reconciliation',
      listEndpoint: '/banking.banking.api.v1.BankingService/ListBankAccounts',
      responseRowsKey: 'accounts',
      responseTotalKey: 'pagination.totalCount',
      columns: ['id', 'accountNumber', 'accountName', 'bankName', 'status', 'currentBalance'],
    },
  ],
};
