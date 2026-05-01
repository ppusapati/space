/**
 * Finance domain module.
 *
 * Entities:
 *   - journal-entries: GL journal entries. Form_id
 *     `form_finance_journal_entry_detailed` is the canonical authoring form
 *     (see backend core/workflow/formbuilder/cmd/generate_forms/configs/finance/journal_entry_form.yaml).
 *     List endpoint /finance.journal.api.v1.JournalEntryService/ListJournalEntries
 *     returns `{entries: [...], pagination: {totalCount: N}}` — the rows
 *     key is `entries` (NOT `journalEntries`) and the count is nested
 *     under `pagination`. ListPage's lookupPath helper handles the
 *     dotted responseTotalKey.
 */
import type { DomainModule } from '../index.js';

export const finance: DomainModule = {
  id: 'finance',
  label: 'Finance',
  entities: [
    {
      slug: 'journal-entries',
      label: 'Journal Entries',
      formId: 'form_finance_journal_entry_detailed',
      listEndpoint: '/finance.journal.api.v1.JournalEntryService/ListJournalEntries',
      responseRowsKey: 'entries',
      responseTotalKey: 'pagination.totalCount',
      columns: ['entryNumber', 'entryDate', 'narration', 'totalDebit', 'totalCredit', 'status'],
    },
  ],
};
