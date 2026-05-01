/**
 * Projects Service Factories
 * Typed ConnectRPC clients for project, task, BOQ, billing, costing, etc.
 */
import { getApiClient } from '../client/client.js';
import { ProjectService } from '@samavāya/proto/gen/business/projects/project/proto/project_pb.js';
import { TaskService } from '@samavāya/proto/gen/business/projects/task/proto/task_pb.js';
import { BOQService } from '@samavāya/proto/gen/business/projects/boq/proto/boq_pb.js';
import { ProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/progressbilling_pb.js';
import { ProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/projectcosting_pb.js';
import { SubContractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/subcontractor_pb.js';
import { TimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/timesheet_pb.js';
// Vertical-specific — Agriculture
import { AgricultureBOQService } from '@samavāya/proto/gen/business/projects/boq/proto/agriculture/boq_agriculture_pb.js';
import { AgricultureProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/agriculture/progressbilling_agriculture_pb.js';
import { AgricultureProjectService } from '@samavāya/proto/gen/business/projects/project/proto/agriculture/project_agriculture_pb.js';
import { AgricultureProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/agriculture/projectcosting_agriculture_pb.js';
import { AgricultureSubContractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/agriculture/subcontractor_agriculture_pb.js';
import { AgricultureTaskService } from '@samavāya/proto/gen/business/projects/task/proto/agriculture/task_agriculture_pb.js';
import { AgricultureTimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/agriculture/timesheet_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionBOQService } from '@samavāya/proto/gen/business/projects/boq/proto/construction/boq_construction_pb.js';
import { ConstructionProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/construction/progressbilling_construction_pb.js';
import { ConstructionProjectService } from '@samavāya/proto/gen/business/projects/project/proto/construction/project_construction_pb.js';
import { ConstructionProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/construction/projectcosting_construction_pb.js';
import { ConstructionSubcontractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/construction/subcontractor_construction_pb.js';
import { ConstructionTaskService } from '@samavāya/proto/gen/business/projects/task/proto/construction/task_construction_pb.js';
import { ConstructionTimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/construction/timesheet_construction_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalBOQService } from '@samavāya/proto/gen/business/projects/boq/proto/mfgvertical/boq_mfgvertical_pb.js';
import { MfgVerticalProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/mfgvertical/progressbilling_mfgvertical_pb.js';
import { MfgVerticalProjectService } from '@samavāya/proto/gen/business/projects/project/proto/mfgvertical/project_mfgvertical_pb.js';
import { MfgVerticalProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/mfgvertical/projectcosting_mfgvertical_pb.js';
import { MfgVerticalSubContractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/mfgvertical/subcontractor_mfgvertical_pb.js';
import { MfgVerticalTaskService } from '@samavāya/proto/gen/business/projects/task/proto/mfgvertical/task_mfgvertical_pb.js';
import { MfgVerticalTimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/mfgvertical/timesheet_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarBOQService } from '@samavāya/proto/gen/business/projects/boq/proto/solar/boq_solar_pb.js';
import { solarProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/solar/progressbilling_solar_pb.js';
import { solarProjectService } from '@samavāya/proto/gen/business/projects/project/proto/solar/project_solar_pb.js';
import { solarProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/solar/projectcosting_solar_pb.js';
import { solarSubContractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/solar/subcontractor_solar_pb.js';
import { solarTaskService } from '@samavāya/proto/gen/business/projects/task/proto/solar/task_solar_pb.js';
import { solarTimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/solar/timesheet_solar_pb.js';
// Vertical-specific — Water
import { WaterBOQService } from '@samavāya/proto/gen/business/projects/boq/proto/water/boq_water_pb.js';
import { WaterProgressBillingService } from '@samavāya/proto/gen/business/projects/progressbilling/proto/water/progressbilling_water_pb.js';
import { WaterProjectService } from '@samavāya/proto/gen/business/projects/project/proto/water/project_water_pb.js';
import { WaterProjectCostingService } from '@samavāya/proto/gen/business/projects/projectcosting/proto/water/projectcosting_water_pb.js';
import { WaterSubcontractorService } from '@samavāya/proto/gen/business/projects/subcontractor/proto/water/subcontractor_water_pb.js';
import { WaterTaskService } from '@samavāya/proto/gen/business/projects/task/proto/water/task_water_pb.js';
import { WaterTimesheetService } from '@samavāya/proto/gen/business/projects/timesheet/proto/water/timesheet_water_pb.js';
export { ProjectService, TaskService, BOQService, ProgressBillingService, ProjectCostingService, SubContractorService, TimesheetService, };
export function getProjectService() {
    return getApiClient().getService(ProjectService);
}
export function getTaskService() {
    return getApiClient().getService(TaskService);
}
export function getBOQService() {
    return getApiClient().getService(BOQService);
}
export function getProgressBillingService() {
    return getApiClient().getService(ProgressBillingService);
}
export function getProjectCostingService() {
    return getApiClient().getService(ProjectCostingService);
}
export function getSubContractorService() {
    return getApiClient().getService(SubContractorService);
}
export function getTimesheetService() {
    return getApiClient().getService(TimesheetService);
}
export { AgricultureBOQService, ConstructionBOQService, MfgVerticalBOQService, solarBOQService, WaterBOQService, AgricultureProgressBillingService, ConstructionProgressBillingService, MfgVerticalProgressBillingService, solarProgressBillingService, WaterProgressBillingService, AgricultureProjectService, ConstructionProjectService, MfgVerticalProjectService, solarProjectService, WaterProjectService, AgricultureProjectCostingService, ConstructionProjectCostingService, MfgVerticalProjectCostingService, solarProjectCostingService, WaterProjectCostingService, AgricultureSubContractorService, ConstructionSubcontractorService, MfgVerticalSubContractorService, solarSubContractorService, WaterSubcontractorService, AgricultureTaskService, ConstructionTaskService, MfgVerticalTaskService, solarTaskService, WaterTaskService, AgricultureTimesheetService, ConstructionTimesheetService, MfgVerticalTimesheetService, solarTimesheetService, WaterTimesheetService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureBOQService() {
    return getApiClient().getService(AgricultureBOQService);
}
export function getAgricultureProgressBillingService() {
    return getApiClient().getService(AgricultureProgressBillingService);
}
export function getAgricultureProjectService() {
    return getApiClient().getService(AgricultureProjectService);
}
export function getAgricultureProjectCostingService() {
    return getApiClient().getService(AgricultureProjectCostingService);
}
export function getAgricultureSubContractorService() {
    return getApiClient().getService(AgricultureSubContractorService);
}
export function getAgricultureTaskService() {
    return getApiClient().getService(AgricultureTaskService);
}
export function getAgricultureTimesheetService() {
    return getApiClient().getService(AgricultureTimesheetService);
}
// ─── Construction Vertical Factories ───
export function getConstructionBOQService() {
    return getApiClient().getService(ConstructionBOQService);
}
export function getConstructionProgressBillingService() {
    return getApiClient().getService(ConstructionProgressBillingService);
}
export function getConstructionProjectService() {
    return getApiClient().getService(ConstructionProjectService);
}
export function getConstructionProjectCostingService() {
    return getApiClient().getService(ConstructionProjectCostingService);
}
export function getConstructionSubcontractorService() {
    return getApiClient().getService(ConstructionSubcontractorService);
}
export function getConstructionTaskService() {
    return getApiClient().getService(ConstructionTaskService);
}
export function getConstructionTimesheetService() {
    return getApiClient().getService(ConstructionTimesheetService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalBOQService() {
    return getApiClient().getService(MfgVerticalBOQService);
}
export function getMfgVerticalProgressBillingService() {
    return getApiClient().getService(MfgVerticalProgressBillingService);
}
export function getMfgVerticalProjectService() {
    return getApiClient().getService(MfgVerticalProjectService);
}
export function getMfgVerticalProjectCostingService() {
    return getApiClient().getService(MfgVerticalProjectCostingService);
}
export function getMfgVerticalSubContractorService() {
    return getApiClient().getService(MfgVerticalSubContractorService);
}
export function getMfgVerticalTaskService() {
    return getApiClient().getService(MfgVerticalTaskService);
}
export function getMfgVerticalTimesheetService() {
    return getApiClient().getService(MfgVerticalTimesheetService);
}
// ─── Solar Vertical Factories ───
export function getSolarBOQService() {
    return getApiClient().getService(solarBOQService);
}
export function getSolarProgressBillingService() {
    return getApiClient().getService(solarProgressBillingService);
}
export function getSolarProjectService() {
    return getApiClient().getService(solarProjectService);
}
export function getSolarProjectCostingService() {
    return getApiClient().getService(solarProjectCostingService);
}
export function getSolarSubContractorService() {
    return getApiClient().getService(solarSubContractorService);
}
export function getSolarTaskService() {
    return getApiClient().getService(solarTaskService);
}
export function getSolarTimesheetService() {
    return getApiClient().getService(solarTimesheetService);
}
// ─── Water Vertical Factories ───
export function getWaterBOQService() {
    return getApiClient().getService(WaterBOQService);
}
export function getWaterProgressBillingService() {
    return getApiClient().getService(WaterProgressBillingService);
}
export function getWaterProjectService() {
    return getApiClient().getService(WaterProjectService);
}
export function getWaterProjectCostingService() {
    return getApiClient().getService(WaterProjectCostingService);
}
export function getWaterSubcontractorService() {
    return getApiClient().getService(WaterSubcontractorService);
}
export function getWaterTaskService() {
    return getApiClient().getService(WaterTaskService);
}
export function getWaterTimesheetService() {
    return getApiClient().getService(WaterTimesheetService);
}
//# sourceMappingURL=projects.js.map