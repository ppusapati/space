/**
 * HR Service Factories
 * Typed ConnectRPC clients for employee, payroll, leave, attendance, etc.
 */
import { getApiClient } from '../client/client.js';
import { EmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/employee_pb.js';
import { PayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/payroll_pb.js';
import { LeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/leave_pb.js';
import { AttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/attendance_pb.js';
import { AppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/appraisal_pb.js';
import { ExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/expense_pb.js';
import { ExitService } from '@samavāya/proto/gen/business/hr/exit/proto/exit_pb.js';
import { RecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/recruitment_pb.js';
import { SalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/salarystructure_pb.js';
import { TrainingService } from '@samavāya/proto/gen/business/hr/training/proto/training_pb.js';
// Vertical-specific — Agriculture
import { AgricultureAppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/agriculture/appraisal_agriculture_pb.js';
import { AgricultureAttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/agriculture/attendance_agriculture_pb.js';
import { AgricultureEmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/agriculture/employee_agriculture_pb.js';
import { AgricultureExitService } from '@samavāya/proto/gen/business/hr/exit/proto/agriculture/exit_agriculture_pb.js';
import { AgricultureExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/agriculture/expense_agriculture_pb.js';
import { AgricultureLeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/agriculture/leave_agriculture_pb.js';
import { AgriculturePayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/agriculture/payroll_agriculture_pb.js';
import { AgricultureRecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/agriculture/recruitment_agriculture_pb.js';
import { AgricultureSalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/agriculture/salarystructure_agriculture_pb.js';
import { AgricultureTrainingService } from '@samavāya/proto/gen/business/hr/training/proto/agriculture/training_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionAppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/construction/appraisal_construction_pb.js';
import { ConstructionAttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/construction/attendance_construction_pb.js';
import { ConstructionEmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/construction/employee_construction_pb.js';
import { ConstructionExitService } from '@samavāya/proto/gen/business/hr/exit/proto/construction/exit_construction_pb.js';
import { ConstructionExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/construction/expense_construction_pb.js';
import { ConstructionLeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/construction/leave_construction_pb.js';
import { ConstructionPayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/construction/payroll_construction_pb.js';
import { ConstructionRecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/construction/recruitment_construction_pb.js';
import { ConstructionSalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/construction/salarystructure_construction_pb.js';
import { ConstructionTrainingService } from '@samavāya/proto/gen/business/hr/training/proto/construction/training_construction_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalAppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/mfgvertical/appraisal_mfgvertical_pb.js';
import { MfgVerticalAttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/mfgvertical/attendance_mfgvertical_pb.js';
import { MfgVerticalEmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/mfgvertical/employee_mfgvertical_pb.js';
import { MfgVerticalExitService } from '@samavāya/proto/gen/business/hr/exit/proto/mfgvertical/exit_mfgvertical_pb.js';
import { MfgVerticalExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/mfgvertical/expense_mfgvertical_pb.js';
import { MfgVerticalLeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/mfgvertical/leave_mfgvertical_pb.js';
import { MfgVerticalPayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/mfgvertical/payroll_mfgvertical_pb.js';
import { MfgVerticalRecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/mfgvertical/recruitment_mfgvertical_pb.js';
import { MfgVerticalSalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/mfgvertical/salarystructure_mfgvertical_pb.js';
import { MfgVerticalTrainingService } from '@samavāya/proto/gen/business/hr/training/proto/mfgvertical/training_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarAppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/solar/appraisal_solar_pb.js';
import { solarAttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/solar/attendance_solar_pb.js';
import { solarEmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/solar/employee_solar_pb.js';
import { solarExitService } from '@samavāya/proto/gen/business/hr/exit/proto/solar/exit_solar_pb.js';
import { solarExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/solar/expense_solar_pb.js';
import { solarLeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/solar/leave_solar_pb.js';
import { solarPayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/solar/payroll_solar_pb.js';
import { solarRecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/solar/recruitment_solar_pb.js';
import { solarSalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/solar/salarystructure_solar_pb.js';
import { solarTrainingService } from '@samavāya/proto/gen/business/hr/training/proto/solar/training_solar_pb.js';
// Vertical-specific — Water
import { WaterAppraisalService } from '@samavāya/proto/gen/business/hr/appraisal/proto/water/appraisal_water_pb.js';
import { WaterAttendanceService } from '@samavāya/proto/gen/business/hr/attendance/proto/water/attendance_water_pb.js';
import { WaterEmployeeService } from '@samavāya/proto/gen/business/hr/employee/proto/water/employee_water_pb.js';
import { WaterExitService } from '@samavāya/proto/gen/business/hr/exit/proto/water/exit_water_pb.js';
import { WaterExpenseService } from '@samavāya/proto/gen/business/hr/expense/proto/water/expense_water_pb.js';
import { WaterLeaveService } from '@samavāya/proto/gen/business/hr/leave/proto/water/leave_water_pb.js';
import { WaterPayrollService } from '@samavāya/proto/gen/business/hr/payroll/proto/water/payroll_water_pb.js';
import { WaterRecruitmentService } from '@samavāya/proto/gen/business/hr/recruitment/proto/water/recruitment_water_pb.js';
import { WaterSalaryStructureService } from '@samavāya/proto/gen/business/hr/salarystructure/proto/water/salarystructure_water_pb.js';
import { WaterTrainingService } from '@samavāya/proto/gen/business/hr/training/proto/water/training_water_pb.js';
export { EmployeeService, PayrollService, LeaveService, AttendanceService, AppraisalService, ExpenseService, ExitService, RecruitmentService, SalaryStructureService, TrainingService, };
export function getEmployeeService() {
    return getApiClient().getService(EmployeeService);
}
export function getPayrollService() {
    return getApiClient().getService(PayrollService);
}
export function getLeaveService() {
    return getApiClient().getService(LeaveService);
}
export function getAttendanceService() {
    return getApiClient().getService(AttendanceService);
}
export function getAppraisalService() {
    return getApiClient().getService(AppraisalService);
}
export function getExpenseService() {
    return getApiClient().getService(ExpenseService);
}
export function getExitService() {
    return getApiClient().getService(ExitService);
}
export function getRecruitmentService() {
    return getApiClient().getService(RecruitmentService);
}
export function getSalaryStructureService() {
    return getApiClient().getService(SalaryStructureService);
}
export function getTrainingService() {
    return getApiClient().getService(TrainingService);
}
export { AgricultureAppraisalService, ConstructionAppraisalService, MfgVerticalAppraisalService, solarAppraisalService, WaterAppraisalService, AgricultureAttendanceService, ConstructionAttendanceService, MfgVerticalAttendanceService, solarAttendanceService, WaterAttendanceService, AgricultureEmployeeService, ConstructionEmployeeService, MfgVerticalEmployeeService, solarEmployeeService, WaterEmployeeService, AgricultureExitService, ConstructionExitService, MfgVerticalExitService, solarExitService, WaterExitService, AgricultureExpenseService, ConstructionExpenseService, MfgVerticalExpenseService, solarExpenseService, WaterExpenseService, AgricultureLeaveService, ConstructionLeaveService, MfgVerticalLeaveService, solarLeaveService, WaterLeaveService, AgriculturePayrollService, ConstructionPayrollService, MfgVerticalPayrollService, solarPayrollService, WaterPayrollService, AgricultureRecruitmentService, ConstructionRecruitmentService, MfgVerticalRecruitmentService, solarRecruitmentService, WaterRecruitmentService, AgricultureSalaryStructureService, ConstructionSalaryStructureService, MfgVerticalSalaryStructureService, solarSalaryStructureService, WaterSalaryStructureService, AgricultureTrainingService, ConstructionTrainingService, MfgVerticalTrainingService, solarTrainingService, WaterTrainingService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureAppraisalService() {
    return getApiClient().getService(AgricultureAppraisalService);
}
export function getAgricultureAttendanceService() {
    return getApiClient().getService(AgricultureAttendanceService);
}
export function getAgricultureEmployeeService() {
    return getApiClient().getService(AgricultureEmployeeService);
}
export function getAgricultureExitService() {
    return getApiClient().getService(AgricultureExitService);
}
export function getAgricultureExpenseService() {
    return getApiClient().getService(AgricultureExpenseService);
}
export function getAgricultureLeaveService() {
    return getApiClient().getService(AgricultureLeaveService);
}
export function getAgriculturePayrollService() {
    return getApiClient().getService(AgriculturePayrollService);
}
export function getAgricultureRecruitmentService() {
    return getApiClient().getService(AgricultureRecruitmentService);
}
export function getAgricultureSalaryStructureService() {
    return getApiClient().getService(AgricultureSalaryStructureService);
}
export function getAgricultureTrainingService() {
    return getApiClient().getService(AgricultureTrainingService);
}
// ─── Construction Vertical Factories ───
export function getConstructionAppraisalService() {
    return getApiClient().getService(ConstructionAppraisalService);
}
export function getConstructionAttendanceService() {
    return getApiClient().getService(ConstructionAttendanceService);
}
export function getConstructionEmployeeService() {
    return getApiClient().getService(ConstructionEmployeeService);
}
export function getConstructionExitService() {
    return getApiClient().getService(ConstructionExitService);
}
export function getConstructionExpenseService() {
    return getApiClient().getService(ConstructionExpenseService);
}
export function getConstructionLeaveService() {
    return getApiClient().getService(ConstructionLeaveService);
}
export function getConstructionPayrollService() {
    return getApiClient().getService(ConstructionPayrollService);
}
export function getConstructionRecruitmentService() {
    return getApiClient().getService(ConstructionRecruitmentService);
}
export function getConstructionSalaryStructureService() {
    return getApiClient().getService(ConstructionSalaryStructureService);
}
export function getConstructionTrainingService() {
    return getApiClient().getService(ConstructionTrainingService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalAppraisalService() {
    return getApiClient().getService(MfgVerticalAppraisalService);
}
export function getMfgVerticalAttendanceService() {
    return getApiClient().getService(MfgVerticalAttendanceService);
}
export function getMfgVerticalEmployeeService() {
    return getApiClient().getService(MfgVerticalEmployeeService);
}
export function getMfgVerticalExitService() {
    return getApiClient().getService(MfgVerticalExitService);
}
export function getMfgVerticalExpenseService() {
    return getApiClient().getService(MfgVerticalExpenseService);
}
export function getMfgVerticalLeaveService() {
    return getApiClient().getService(MfgVerticalLeaveService);
}
export function getMfgVerticalPayrollService() {
    return getApiClient().getService(MfgVerticalPayrollService);
}
export function getMfgVerticalRecruitmentService() {
    return getApiClient().getService(MfgVerticalRecruitmentService);
}
export function getMfgVerticalSalaryStructureService() {
    return getApiClient().getService(MfgVerticalSalaryStructureService);
}
export function getMfgVerticalTrainingService() {
    return getApiClient().getService(MfgVerticalTrainingService);
}
// ─── Solar Vertical Factories ───
export function getSolarAppraisalService() {
    return getApiClient().getService(solarAppraisalService);
}
export function getSolarAttendanceService() {
    return getApiClient().getService(solarAttendanceService);
}
export function getSolarEmployeeService() {
    return getApiClient().getService(solarEmployeeService);
}
export function getSolarExitService() {
    return getApiClient().getService(solarExitService);
}
export function getSolarExpenseService() {
    return getApiClient().getService(solarExpenseService);
}
export function getSolarLeaveService() {
    return getApiClient().getService(solarLeaveService);
}
export function getSolarPayrollService() {
    return getApiClient().getService(solarPayrollService);
}
export function getSolarRecruitmentService() {
    return getApiClient().getService(solarRecruitmentService);
}
export function getSolarSalaryStructureService() {
    return getApiClient().getService(solarSalaryStructureService);
}
export function getSolarTrainingService() {
    return getApiClient().getService(solarTrainingService);
}
// ─── Water Vertical Factories ───
export function getWaterAppraisalService() {
    return getApiClient().getService(WaterAppraisalService);
}
export function getWaterAttendanceService() {
    return getApiClient().getService(WaterAttendanceService);
}
export function getWaterEmployeeService() {
    return getApiClient().getService(WaterEmployeeService);
}
export function getWaterExitService() {
    return getApiClient().getService(WaterExitService);
}
export function getWaterExpenseService() {
    return getApiClient().getService(WaterExpenseService);
}
export function getWaterLeaveService() {
    return getApiClient().getService(WaterLeaveService);
}
export function getWaterPayrollService() {
    return getApiClient().getService(WaterPayrollService);
}
export function getWaterRecruitmentService() {
    return getApiClient().getService(WaterRecruitmentService);
}
export function getWaterSalaryStructureService() {
    return getApiClient().getService(WaterSalaryStructureService);
}
export function getWaterTrainingService() {
    return getApiClient().getService(WaterTrainingService);
}
//# sourceMappingURL=hr.js.map