/**
 * HR Service Factories
 * Typed ConnectRPC clients for employee, payroll, leave, attendance, etc.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { EmployeeService } from '@chetana/proto/gen/business/hr/employee/proto/employee_pb.js';
import { PayrollService } from '@chetana/proto/gen/business/hr/payroll/proto/payroll_pb.js';
import { LeaveService } from '@chetana/proto/gen/business/hr/leave/proto/leave_pb.js';
import { AttendanceService } from '@chetana/proto/gen/business/hr/attendance/proto/attendance_pb.js';
import { AppraisalService } from '@chetana/proto/gen/business/hr/appraisal/proto/appraisal_pb.js';
import { ExpenseService } from '@chetana/proto/gen/business/hr/expense/proto/expense_pb.js';
import { ExitService } from '@chetana/proto/gen/business/hr/exit/proto/exit_pb.js';
import { RecruitmentService } from '@chetana/proto/gen/business/hr/recruitment/proto/recruitment_pb.js';
import { SalaryStructureService } from '@chetana/proto/gen/business/hr/salarystructure/proto/salarystructure_pb.js';
import { TrainingService } from '@chetana/proto/gen/business/hr/training/proto/training_pb.js';

// Vertical-specific — Agriculture
// Appraisal agriculture vertical retired in Phase F.5.8 — see config/class_registry/appraisal.yaml.
// Attendance agriculture vertical retired in Phase F.5.2 — see config/class_registry/attendance.yaml.
// Employee agriculture vertical retired in Phase F.5.1 — see config/class_registry/employee.yaml.
// Exit agriculture vertical retired in Phase F.5.9 — see config/class_registry/exit.yaml.
// Expense agriculture vertical retired in Phase F.5.10 — see config/class_registry/expense.yaml.
// Leave agriculture vertical retired in Phase F.5.3 — see config/class_registry/leave.yaml.
// Payroll agriculture vertical retired in Phase F.5.4 — see config/class_registry/payroll.yaml.
// Recruitment agriculture vertical retired in Phase F.5.6 — see config/class_registry/recruitment.yaml.
// SalaryStructure agriculture vertical retired in Phase F.5.5 — see config/class_registry/salarystructure.yaml.
// Training agriculture vertical retired in Phase F.5.7 — see config/class_registry/training.yaml.
// Vertical-specific — Construction
// Appraisal construction vertical retired in Phase F.5.8 — see config/class_registry/appraisal.yaml.
// Attendance construction vertical retired in Phase F.5.2 — see config/class_registry/attendance.yaml.
// Employee construction vertical retired in Phase F.5.1 — see config/class_registry/employee.yaml.
// Exit construction vertical retired in Phase F.5.9 — see config/class_registry/exit.yaml.
// Expense construction vertical retired in Phase F.5.10 — see config/class_registry/expense.yaml.
// Leave construction vertical retired in Phase F.5.3 — see config/class_registry/leave.yaml.
// Payroll construction vertical retired in Phase F.5.4 — see config/class_registry/payroll.yaml.
// Recruitment construction vertical retired in Phase F.5.6 — see config/class_registry/recruitment.yaml.
// SalaryStructure construction vertical retired in Phase F.5.5 — see config/class_registry/salarystructure.yaml.
// Training construction vertical retired in Phase F.5.7 — see config/class_registry/training.yaml.
// Vertical-specific — MfgVertical (Manufacturing)
// Appraisal mfgvertical retired in Phase F.5.8 — see config/class_registry/appraisal.yaml.
// Attendance mfgvertical retired in Phase F.5.2 — see config/class_registry/attendance.yaml.
// Employee mfgvertical retired in Phase F.5.1 — see config/class_registry/employee.yaml.
// Exit mfgvertical retired in Phase F.5.9 — see config/class_registry/exit.yaml.
// Expense mfgvertical retired in Phase F.5.10 — see config/class_registry/expense.yaml.
// Leave mfgvertical retired in Phase F.5.3 — see config/class_registry/leave.yaml.
// Payroll mfgvertical retired in Phase F.5.4 — see config/class_registry/payroll.yaml.
// Recruitment mfgvertical retired in Phase F.5.6 — see config/class_registry/recruitment.yaml.
// SalaryStructure mfgvertical retired in Phase F.5.5 — see config/class_registry/salarystructure.yaml.
// Training mfgvertical retired in Phase F.5.7 — see config/class_registry/training.yaml.
// Vertical-specific — Solar
// Appraisal solar vertical retired in Phase F.5.8 — see config/class_registry/appraisal.yaml.
// Attendance solar vertical retired in Phase F.5.2 — see config/class_registry/attendance.yaml.
// Employee solar vertical retired in Phase F.5.1 — see config/class_registry/employee.yaml.
// Exit solar vertical retired in Phase F.5.9 — see config/class_registry/exit.yaml.
// Expense solar vertical retired in Phase F.5.10 — see config/class_registry/expense.yaml.
// Leave solar vertical retired in Phase F.5.3 — see config/class_registry/leave.yaml.
// Payroll solar vertical retired in Phase F.5.4 — see config/class_registry/payroll.yaml.
// Recruitment solar vertical retired in Phase F.5.6 — see config/class_registry/recruitment.yaml.
// SalaryStructure solar vertical retired in Phase F.5.5 — see config/class_registry/salarystructure.yaml.
// Training solar vertical retired in Phase F.5.7 — see config/class_registry/training.yaml.
// Vertical-specific — Water
// Appraisal water vertical retired in Phase F.5.8 — see config/class_registry/appraisal.yaml.
// Attendance water vertical retired in Phase F.5.2 — see config/class_registry/attendance.yaml.
// Employee water vertical retired in Phase F.5.1 — see config/class_registry/employee.yaml.
// Exit water vertical retired in Phase F.5.9 — see config/class_registry/exit.yaml.
// Expense water vertical retired in Phase F.5.10 — see config/class_registry/expense.yaml.
// Leave water vertical retired in Phase F.5.3 — see config/class_registry/leave.yaml.
// Payroll water vertical retired in Phase F.5.4 — see config/class_registry/payroll.yaml.
// Recruitment water vertical retired in Phase F.5.6 — see config/class_registry/recruitment.yaml.
// SalaryStructure water vertical retired in Phase F.5.5 — see config/class_registry/salarystructure.yaml.
// Training water vertical retired in Phase F.5.7 — see config/class_registry/training.yaml.

export {
  EmployeeService, PayrollService, LeaveService, AttendanceService,
  AppraisalService, ExpenseService, ExitService, RecruitmentService,
  SalaryStructureService, TrainingService,
};

export function getEmployeeService(): Client<typeof EmployeeService> {
  return getApiClient().getService(EmployeeService);
}

export function getPayrollService(): Client<typeof PayrollService> {
  return getApiClient().getService(PayrollService);
}

export function getLeaveService(): Client<typeof LeaveService> {
  return getApiClient().getService(LeaveService);
}

export function getAttendanceService(): Client<typeof AttendanceService> {
  return getApiClient().getService(AttendanceService);
}

export function getAppraisalService(): Client<typeof AppraisalService> {
  return getApiClient().getService(AppraisalService);
}

export function getExpenseService(): Client<typeof ExpenseService> {
  return getApiClient().getService(ExpenseService);
}

export function getExitService(): Client<typeof ExitService> {
  return getApiClient().getService(ExitService);
}

export function getRecruitmentService(): Client<typeof RecruitmentService> {
  return getApiClient().getService(RecruitmentService);
}

export function getSalaryStructureService(): Client<typeof SalaryStructureService> {
  return getApiClient().getService(SalaryStructureService);
}

export function getTrainingService(): Client<typeof TrainingService> {
  return getApiClient().getService(TrainingService);
}


export {
  // Appraisal vertical re-exports retired in Phase F.5.8 — callers use AppraisalService + class attributes.
  // Attendance vertical re-exports retired in Phase F.5.2 — callers use AttendanceService + class attributes.
  // Employee vertical re-exports retired in Phase F.5.1 — callers use EmployeeService + class attributes.
  // Exit vertical re-exports retired in Phase F.5.9 — callers use ExitService + class attributes.
  // Expense vertical re-exports retired in Phase F.5.10 — callers use ExpenseService + class attributes.
  // Leave vertical re-exports retired in Phase F.5.3 — callers use LeaveService + class attributes.
  // Payroll vertical re-exports retired in Phase F.5.4 — callers use PayrollService + class attributes.
  // Recruitment vertical re-exports retired in Phase F.5.6 — callers use RecruitmentService + class attributes.
  // SalaryStructure vertical re-exports retired in Phase F.5.5 — callers use SalaryStructureService + class attributes.
  // Training vertical re-exports retired in Phase F.5.7 — callers use TrainingService + class attributes.
};

// ─── Agriculture Vertical Factories ───

// getAgricultureAppraisalService retired in Phase F.5.8 — callers use getAppraisalService()
// with class="agri_seasonal_review".

// getAgricultureAttendanceService retired in Phase F.5.2 — callers use getAttendanceService()
// with class="agri_daily_wage_muster".

// getAgricultureEmployeeService retired in Phase F.5.1 — callers use getEmployeeService()
// with class="agri_field_worker".

// getAgricultureExitService retired in Phase F.5.9 — callers use getExitService()
// with class="agri_seasonal_end".

// getAgricultureExpenseService retired in Phase F.5.10 — callers use getExpenseService()
// with class="agri_field_expense".

// getAgricultureLeaveService retired in Phase F.5.3 — callers use getLeaveService()
// with class="agri_seasonal_off".

// getAgriculturePayrollService retired in Phase F.5.4 — callers use getPayrollService()
// with class="agri_daily_wage_settlement".

// getAgricultureRecruitmentService retired in Phase F.5.6 — callers use getRecruitmentService()
// with class="agri_seasonal_hiring".

// getAgricultureSalaryStructureService retired in Phase F.5.5 — callers use getSalaryStructureService()
// with class="agri_daily_wage".

// getAgricultureTrainingService retired in Phase F.5.7 — callers use getTrainingService()
// with class="agri_extension".

// ─── Construction Vertical Factories ───

// getConstructionAppraisalService retired in Phase F.5.8 — callers use getAppraisalService()
// with class="construction_project_review".

// getConstructionAttendanceService retired in Phase F.5.2 — callers use getAttendanceService()
// with class="construction_site_punch".

// getConstructionEmployeeService retired in Phase F.5.1 — callers use getEmployeeService()
// with class="construction_site_worker".

// getConstructionExitService retired in Phase F.5.9 — callers use getExitService()
// with class="construction_site_demobilisation".

// getConstructionExpenseService retired in Phase F.5.10 — callers use getExpenseService()
// with class="construction_site_expense".

// getConstructionLeaveService retired in Phase F.5.3 — callers use getLeaveService()
// with class="construction_site_off".

// getConstructionPayrollService retired in Phase F.5.4 — callers use getPayrollService()
// with class="construction_site_payroll".

// getConstructionRecruitmentService retired in Phase F.5.6 — callers use getRecruitmentService()
// with class="construction_site_hiring".

// getConstructionSalaryStructureService retired in Phase F.5.5 — callers use getSalaryStructureService()
// with class="construction_site_wage".

// getConstructionTrainingService retired in Phase F.5.7 — callers use getTrainingService()
// with class="construction_site_safety_induction".

// ─── MfgVertical (Manufacturing) Vertical Factories ───

// getMfgVerticalAppraisalService retired in Phase F.5.8 — callers use getAppraisalService()
// with class="manufacturing_skill_review".

// getMfgVerticalAttendanceService retired in Phase F.5.2 — callers use getAttendanceService()
// with class="manufacturing_shopfloor_shift".

// getMfgVerticalEmployeeService retired in Phase F.5.1 — callers use getEmployeeService()
// with class="manufacturing_operator".

// getMfgVerticalExitService retired in Phase F.5.9 — callers use getExitService()
// with class="manufacturing_voluntary_retirement".

// getMfgVerticalExpenseService retired in Phase F.5.10 — callers use getExpenseService()
// with class="manufacturing_plant_expense".

// getMfgVerticalLeaveService retired in Phase F.5.3 — callers use getLeaveService()
// with class="manufacturing_shutdown_leave".

// getMfgVerticalPayrollService retired in Phase F.5.4 — callers use getPayrollService()
// with class="manufacturing_shift_payroll".

// getMfgVerticalRecruitmentService retired in Phase F.5.6 — callers use getRecruitmentService()
// with class="manufacturing_operator_hiring".

// getMfgVerticalSalaryStructureService retired in Phase F.5.5 — callers use getSalaryStructureService()
// with class="manufacturing_piece_rate" or "narrow_graded_manufacturing".

// getMfgVerticalTrainingService retired in Phase F.5.7 — callers use getTrainingService()
// with class="manufacturing_skill_matrix".

// ─── Solar Vertical Factories ───

// getSolarAppraisalService retired in Phase F.5.8 — callers use getAppraisalService()
// with class="solar_om_kpi_review".

// getSolarAttendanceService retired in Phase F.5.2 — callers use getAttendanceService()
// with class="solar_om_shift".

// getSolarEmployeeService retired in Phase F.5.1 — callers use getEmployeeService()
// with class="solar_om_technician".

// getSolarExitService retired in Phase F.5.9 — callers use getExitService()
// with class="solar_om_contract_end".

// getSolarExpenseService retired in Phase F.5.10 — callers use getExpenseService()
// with class="solar_site_expense".

// getSolarLeaveService retired in Phase F.5.3 — callers use getLeaveService()
// with class="solar_plant_rotational_off".

// getSolarPayrollService retired in Phase F.5.4 — callers use getPayrollService()
// with class="solar_om_payroll".

// getSolarRecruitmentService retired in Phase F.5.6 — callers use getRecruitmentService()
// with class="solar_om_hiring".

// getSolarSalaryStructureService retired in Phase F.5.5 — callers use getSalaryStructureService()
// with class="solar_om_structure".

// getSolarTrainingService retired in Phase F.5.7 — callers use getTrainingService()
// with class="solar_electrical_safety".

// ─── Water Vertical Factories ───

// getWaterAppraisalService retired in Phase F.5.8 — callers use getAppraisalService()
// with class="water_utility_kpi_review".

// getWaterAttendanceService retired in Phase F.5.2 — callers use getAttendanceService()
// with class="water_utility_control_room".

// getWaterEmployeeService retired in Phase F.5.1 — callers use getEmployeeService()
// with class="water_utility_operator".

// getWaterExitService retired in Phase F.5.9 — callers use getExitService()
// with class="water_utility_exit".

// getWaterExpenseService retired in Phase F.5.10 — callers use getExpenseService()
// with class="water_utility_expense".

// getWaterLeaveService retired in Phase F.5.3 — callers use getLeaveService()
// with class="water_utility_rotational_off".

// getWaterPayrollService retired in Phase F.5.4 — callers use getPayrollService()
// with class="water_utility_payroll".

// getWaterRecruitmentService retired in Phase F.5.6 — callers use getRecruitmentService()
// with class="water_utility_hiring".

// getWaterSalaryStructureService retired in Phase F.5.5 — callers use getSalaryStructureService()
// with class="water_utility_structure".

// getWaterTrainingService retired in Phase F.5.7 — callers use getTrainingService()
// with class="water_utility_quality_training".

