/**
 * Data domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - backup-policies: BackupDRService.ListBackupPolicies under
 *     `/data.backupdr.api.v1.`. Lives in `core/data/backupdr/`.
 *     Response uses FLAT `totalCount` — proto declares
 *     `int32 total_count = 4` with NO nested Pagination object (it has
 *     `next_page_token` for cursor pagination instead, but we ignore the
 *     cursor for the demo and rely on offset pagination from the request).
 */
import type { DomainModule } from '../index.js';

export const data: DomainModule = {
  id: 'data',
  label: 'Data',
  entities: [
    {
      slug: 'backup-policies',
      label: 'Backup Policies',
      formId: 'backup_scheduling',
      listEndpoint: '/data.backupdr.api.v1.BackupDRService/ListBackupPolicies',
      responseRowsKey: 'policies',
      responseTotalKey: 'totalCount',
      columns: ['id', 'name', 'backupType', 'targetType', 'scheduleType', 'status'],
    },
  ],
};
