/**
 * Projects Service Factories — Phase F.10 consolidated surface.
 *
 * Phase F.10 (2026-04-22) collapsed the prior vertical subtrees
 * (agriculture / construction / constructionvertical / mfgvertical /
 * solar / water / workvertical) into the unified classregistry
 * taxonomy on each host service:
 *   - project:         construction_project, it_services_project,
 *                      manufacturing_capex, agri_development,
 *                      solar_plant_epc, water_treatment_epc, …
 *   - boq:             civil_works, mep_electrical, mep_plumbing,
 *                      finishing, landscaping, structural_steel,
 *                      road_infra, solar_plant_boq, water_plant_boq,
 *                      manufacturing_plant_boq, agri_farm_development,
 *                      site_demolition
 *   - task:            design_task, procurement_task, execution_task,
 *                      inspection_task, approval_task, milestone,
 *                      construction_task, solar_task, water_task,
 *                      mfg_task, agri_task
 *   - timesheet:       billable_hours, non_billable_hours, overtime,
 *                      on_call, training_hours, site_hours_construction,
 *                      plant_hours_solar, plant_hours_water,
 *                      plant_hours_mfg, field_hours_agri
 *   - projectcosting:  labour_cost, material_cost, equipment_cost,
 *                      subcontract_cost, overhead_cost, travel_cost,
 *                      contingency, construction_site_cost,
 *                      solar_plant_cost, water_utility_cost,
 *                      mfg_project_cost, agri_project_cost
 *   - subcontractor:   civil_subcontractor, mep_subcontractor,
 *                      specialist_subcontractor, labour_subcontractor,
 *                      equipment_rental, solar_installation_subcontractor,
 *                      water_plant_subcontractor,
 *                      mfg_fabrication_subcontractor,
 *                      agri_field_subcontractor
 *   - progressbilling: time_and_material, fixed_price_milestone,
 *                      percent_complete, retention_release, lumpsum,
 *                      unit_rate_billing, cost_plus,
 *                      construction_progress, solar_om_progress,
 *                      water_utility_progress, mfg_project_progress,
 *                      agri_project_progress
 *
 * Vertical-specific factories (AgricultureProjectService,
 * ConstructionBOQService, SolarTimesheetService, WaterSubContractorService,
 * MfgVerticalTaskService, etc.) are RETIRED — callers must now use
 * the base service client and pass `class` on Create/Update and as a
 * ListX filter. Consolidated *class introspection RPCs are available:
 *   - ListProjectClasses / GetProjectClassSchema
 *   - ListBOQClasses / GetBOQClassSchema
 *   - ListTaskClasses / GetTaskClassSchema
 *   - ListTimesheetClasses / GetTimesheetClassSchema
 *   - ListProjectCostingClasses / GetProjectCostingClassSchema
 *   - ListSubcontractorClasses / GetSubcontractorClassSchema
 *   - ListProgressBillingClasses / GetProgressBillingClassSchema
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { ProjectService } from '@chetana/proto/gen/business/projects/project/proto/project_pb.js';
import { TaskService } from '@chetana/proto/gen/business/projects/task/proto/task_pb.js';
import { BOQService } from '@chetana/proto/gen/business/projects/boq/proto/boq_pb.js';
import { ProgressBillingService } from '@chetana/proto/gen/business/projects/progressbilling/proto/progressbilling_pb.js';
import { ProjectCostingService } from '@chetana/proto/gen/business/projects/projectcosting/proto/projectcosting_pb.js';
import { SubContractorService } from '@chetana/proto/gen/business/projects/subcontractor/proto/subcontractor_pb.js';
import { TimesheetService } from '@chetana/proto/gen/business/projects/timesheet/proto/timesheet_pb.js';

// ─────────────────────────────────────────────────────────────────────
// Base-service factories (the seven host services)
// ─────────────────────────────────────────────────────────────────────

export function getProjectService(): Client<typeof ProjectService> {
    return getApiClient().getService(ProjectService);
}

export function getTaskService(): Client<typeof TaskService> {
    return getApiClient().getService(TaskService);
}

export function getBOQService(): Client<typeof BOQService> {
    return getApiClient().getService(BOQService);
}

export function getProgressBillingService(): Client<typeof ProgressBillingService> {
    return getApiClient().getService(ProgressBillingService);
}

export function getProjectCostingService(): Client<typeof ProjectCostingService> {
    return getApiClient().getService(ProjectCostingService);
}

export function getSubContractorService(): Client<typeof SubContractorService> {
    return getApiClient().getService(SubContractorService);
}

export function getTimesheetService(): Client<typeof TimesheetService> {
    return getApiClient().getService(TimesheetService);
}

// ─────────────────────────────────────────────────────────────────────
// Re-exports for typed request / response messages
// ─────────────────────────────────────────────────────────────────────

export {
    ProjectService,
    TaskService,
    BOQService,
    ProgressBillingService,
    ProjectCostingService,
    SubContractorService,
    TimesheetService,
};
