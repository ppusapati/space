/**
 * Projects domain module.
 *
 * Entities (verified live 2026-04-26):
 *   - projects: ProjectService.ListProjects. Response `{base, projects,
 *     pagination}` (verified — the live probe returned `{"base":{...}}` with
 *     base.message="OK" against an empty tenant).
 *   - tasks: TaskService.ListTasks. Response `{tasks, pagination}` per
 *     project task proto.
 */
import type { DomainModule } from '../index.js';

export const projects: DomainModule = {
  id: 'projects',
  label: 'Projects',
  entities: [
    {
      slug: 'projects',
      label: 'Projects',
      formId: 'form_projects_master_setup',
      listEndpoint: '/projects.project.api.v1.ProjectService/ListProjects',
      responseRowsKey: 'projects',
      responseTotalKey: 'pagination.totalCount',
      columns: ['projectCode', 'name', 'status', 'startDate', 'plannedEndDate'],
    },
    {
      slug: 'tasks',
      label: 'Tasks',
      formId: 'milestone_tracking',
      listEndpoint: '/projects.task.api.v1.TaskService/ListTasks',
      responseRowsKey: 'tasks',
      responseTotalKey: 'pagination.totalCount',
      // Without a projectId filter, the handler returns an empty success
      // (200) with a base.message hinting that projectId is needed for
      // non-empty results. Real cross-project tenant-wide listing requires
      // a new sqlc query + repo + service method (see
      // business/projects/task/internal/handler/task_handler.go ListTasks
      // for the deferred work). The empty-state lands cleanly in
      // CrudListPage and the user can navigate to the form to create a
      // task; the proper list comes when the new query is added.
      columns: ['taskId', 'name', 'status', 'assignee', 'dueDate'],
    },
  ],
};
