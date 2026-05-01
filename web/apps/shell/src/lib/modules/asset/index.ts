/**
 * Asset domain module — entity registry.
 *
 * History:
 *   - Phase 2 / item 30: 1 entity (categories) shipped.
 *   - Item 31 (DEPLOY.FE.TIER1-COVERAGE): 2nd entity (transfers) added.
 *   - **2026-04-29 DEPLOY.FE.ASSET-COVERAGE (this file):** expanded to
 *     26 entities across the 5 asset sub-modules (asset, depreciation,
 *     equipment, maintenance, vehicle). Recon probed all 29 List RPCs;
 *     26 returned 200 under empty payload + valid JWT, 3 returned 500
 *     because they require a parent_id (sub-resources). The 3 are
 *     intentionally NOT wired — exposing them as standalone menu
 *     entries would create broken UX, exactly as we did for masters
 *     sub-resources in item 34.
 *
 * Sub-resource Lists intentionally NOT wired (require parent_id):
 *   - asset.depreciation.DepreciationService/ListDepreciationEntries
 *     (needs run_id; surfaced inline under a depreciation-run detail page).
 *   - asset.equipment.EquipmentService/ListCalibrations
 *     (needs equipment_id; surfaced inline under equipment detail).
 *   - asset.vehicle.VehicleService/ListVehicleDocuments
 *     (needs vehicle_id; surfaced inline under vehicle detail).
 *
 * Wire-shape conventions (verified per proto + live-probed):
 *   - Every paginated List response carries
 *     `packages.api.v1.pagination.Pagination pagination = 3` →
 *     `responseTotalKey: 'pagination.totalCount'`.
 *   - Class-list responses (ListAssetClasses / ListEquipmentClasses /
 *     ListVehicleClasses / ListDepreciationClasses /
 *     ListMaintenanceClasses) carry no pagination field — the loader
 *     falls back to `rows.length` when totalKey is omitted.
 *   - rowsKey is always the proto's `repeated X <key> = 2;` field name
 *     in camelCase JSON form. Captured per entity below; verifying via
 *     `head -c 600` of a live List response is the canonical way to
 *     confirm a proto change later.
 *
 * Form catalog reuse:
 *   - The asset module's form catalog has 16 forms; many entities
 *     reuse a related form for column derivation (the form's schema
 *     describes the create-side fields, which usually overlap the
 *     list-side). Where the form catalog has no semantic match (e.g.
 *     class-list views, audit-trail-only entities), the closest
 *     existing form is used — Forms aren't required to render the
 *     list page (columns prop overrides the form's coreFields), but
 *     a valid formId is required for the create button to work.
 */
import type { DomainModule } from '../index.js';

export const asset: DomainModule = {
  id: 'asset',
  label: 'Assets',
  entities: [
    // ============================================================
    // asset.asset.AssetService — core asset register + lifecycle
    // ============================================================
    {
      slug: 'assets',
      label: 'Assets',
      formId: 'asset_master',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListAssets',
      responseRowsKey: 'assets',
      responseTotalKey: 'pagination.totalCount',
      columns: ['assetCode', 'assetName', 'category', 'status', 'bookValue', 'location'],
    },
    {
      slug: 'categories',
      label: 'Categories',
      formId: 'asset_audit_trail',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListCategories',
      responseRowsKey: 'categories',
      // ListCategoriesResponse has no pagination wrapper — rows.length used.
      columns: ['categoryCode', 'categoryName', 'assetType', 'isActive'],
    },
    {
      slug: 'transfers',
      label: 'Transfers',
      formId: 'asset_transfer',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListTransfers',
      responseRowsKey: 'transfers',
      responseTotalKey: 'pagination.totalCount',
      columns: ['transferNumber', 'status', 'fromLocation', 'toLocation', 'transferDate'],
    },
    {
      slug: 'disposals',
      label: 'Disposals',
      formId: 'asset_disposal',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListDisposals',
      responseRowsKey: 'disposals',
      responseTotalKey: 'pagination.totalCount',
      columns: ['disposalNumber', 'assetId', 'disposalType', 'disposalDate', 'disposalValue', 'status'],
    },
    {
      slug: 'revaluations',
      label: 'Revaluations',
      formId: 'asset_impairment',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListRevaluations',
      responseRowsKey: 'revaluations',
      responseTotalKey: 'pagination.totalCount',
      columns: ['revaluationNumber', 'assetId', 'previousValue', 'newValue', 'revaluationDate'],
    },
    {
      slug: 'asset-classes',
      label: 'Asset Classes',
      formId: 'asset_master',
      listEndpoint: '/asset.asset.api.v1.AssetService/ListAssetClasses',
      responseRowsKey: 'classes',
      // ListAssetClassesResponse has no pagination — rows.length used.
      columns: ['name', 'displayName', 'description'],
    },

    // ============================================================
    // asset.depreciation.DepreciationService — depreciation engine
    // ============================================================
    {
      slug: 'depreciation-setups',
      label: 'Depreciation Setups',
      formId: 'asset_depreciation_schedule',
      listEndpoint: '/asset.depreciation.api.v1.DepreciationService/ListDepreciationSetups',
      responseRowsKey: 'setups',
      responseTotalKey: 'pagination.totalCount',
      columns: ['assetId', 'depreciationType', 'method', 'acquisitionCost', 'salvageValue', 'usefulLifeMonths'],
    },
    {
      slug: 'depreciation-schedules',
      label: 'Schedules',
      formId: 'asset_depreciation_schedule',
      listEndpoint: '/asset.depreciation.api.v1.DepreciationService/ListSchedules',
      responseRowsKey: 'schedules',
      responseTotalKey: 'pagination.totalCount',
      columns: ['assetId', 'periodNumber', 'periodStart', 'periodEnd', 'openingValue', 'depreciationAmount'],
    },
    {
      slug: 'depreciation-runs',
      label: 'Depreciation Runs',
      formId: 'asset_depreciation_schedule',
      listEndpoint: '/asset.depreciation.api.v1.DepreciationService/ListDepreciationRuns',
      responseRowsKey: 'runs',
      responseTotalKey: 'pagination.totalCount',
      columns: ['runNumber', 'runName', 'depreciationType', 'periodStart', 'periodEnd', 'runDate', 'status'],
    },
    {
      slug: 'book-tax-differences',
      label: 'Book-Tax Differences',
      formId: 'asset_depreciation_reversal',
      listEndpoint: '/asset.depreciation.api.v1.DepreciationService/ListBookTaxDifferences',
      responseRowsKey: 'differences',
      responseTotalKey: 'pagination.totalCount',
      columns: ['assetId', 'financialYear', 'bookDepreciation', 'taxDepreciation', 'difference', 'cumulativeDifference'],
    },
    {
      slug: 'depreciation-classes',
      label: 'Depreciation Classes',
      formId: 'asset_depreciation_schedule',
      listEndpoint: '/asset.depreciation.api.v1.DepreciationService/ListDepreciationClasses',
      responseRowsKey: 'classes',
      // No pagination wrapper.
      columns: ['name', 'displayName', 'description'],
    },

    // ============================================================
    // asset.equipment.EquipmentService — equipment register
    // ============================================================
    {
      slug: 'equipment',
      label: 'Equipment',
      formId: 'equipment_warranty',
      listEndpoint: '/asset.equipment.api.v1.EquipmentService/ListEquipment',
      responseRowsKey: 'equipment',
      responseTotalKey: 'pagination.totalCount',
      columns: ['equipmentCode', 'equipmentName', 'manufacturer', 'modelNumber', 'serialNumber', 'status'],
    },
    {
      slug: 'equipment-categories',
      label: 'Equipment Categories',
      formId: 'equipment_warranty',
      listEndpoint: '/asset.equipment.api.v1.EquipmentService/ListCategories',
      responseRowsKey: 'categories',
      responseTotalKey: 'pagination.totalCount',
      columns: ['categoryCode', 'categoryName', 'parentId', 'level', 'isLeaf'],
    },
    {
      slug: 'certifications',
      label: 'Certifications',
      formId: 'equipment_calibration',
      listEndpoint: '/asset.equipment.api.v1.EquipmentService/ListCertifications',
      responseRowsKey: 'certifications',
      responseTotalKey: 'pagination.totalCount',
      columns: ['equipmentId', 'certificationType', 'certificationNumber', 'issuingAuthority', 'issueDate', 'expiryDate', 'status'],
    },
    {
      slug: 'equipment-classes',
      label: 'Equipment Classes',
      formId: 'equipment_warranty',
      listEndpoint: '/asset.equipment.api.v1.EquipmentService/ListEquipmentClasses',
      responseRowsKey: 'classes',
      columns: ['name', 'displayName', 'description'],
    },

    // ============================================================
    // asset.maintenance.MaintenanceService — work-order engine
    // ============================================================
    {
      slug: 'maintenance-requests',
      label: 'Maintenance Requests',
      formId: 'asset_maintenance_schedule',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListMaintenanceRequests',
      responseRowsKey: 'requests',
      responseTotalKey: 'pagination.totalCount',
      columns: ['requestNumber', 'assetId', 'equipmentId', 'requestType', 'priority', 'status'],
    },
    {
      slug: 'work-orders',
      label: 'Work Orders',
      formId: 'asset_maintenance_schedule',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListWorkOrders',
      responseRowsKey: 'workOrders',
      responseTotalKey: 'pagination.totalCount',
      columns: ['woNumber', 'assetId', 'equipmentId', 'woType', 'priority', 'status', 'scheduledDate'],
    },
    {
      slug: 'pm-schedules',
      label: 'PM Schedules',
      formId: 'asset_maintenance_schedule',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListPMSchedules',
      responseRowsKey: 'schedules',
      responseTotalKey: 'pagination.totalCount',
      columns: ['scheduleName', 'assetId', 'equipmentId', 'frequency', 'intervalValue', 'nextDueDate'],
    },
    {
      slug: 'spare-parts',
      label: 'Spare Parts',
      formId: 'asset_spare_parts',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListSpareParts',
      responseRowsKey: 'spareParts',
      responseTotalKey: 'pagination.totalCount',
      columns: ['partNumber', 'partName', 'category', 'minStock', 'currentStock', 'unitCost'],
    },
    {
      slug: 'checklist-templates',
      label: 'Checklist Templates',
      formId: 'asset_audit_trail',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListChecklistTemplates',
      responseRowsKey: 'templates',
      responseTotalKey: 'pagination.totalCount',
      columns: ['templateName', 'description', 'category', 'isActive'],
    },
    {
      slug: 'maintenance-classes',
      label: 'Maintenance Classes',
      formId: 'asset_maintenance_schedule',
      listEndpoint: '/asset.maintenance.api.v1.MaintenanceService/ListMaintenanceClasses',
      responseRowsKey: 'classes',
      columns: ['name', 'displayName', 'description'],
    },

    // ============================================================
    // asset.vehicle.VehicleService — fleet
    // ============================================================
    {
      slug: 'vehicles',
      label: 'Vehicles',
      formId: 'vehicle_fleet_management',
      listEndpoint: '/asset.vehicle.api.v1.VehicleService/ListVehicles',
      responseRowsKey: 'vehicles',
      responseTotalKey: 'pagination.totalCount',
      columns: ['vehicleCode', 'registrationNumber', 'vehicleType', 'make', 'model', 'year', 'status'],
    },
    {
      slug: 'trips',
      label: 'Trips',
      formId: 'vehicle_fleet_management',
      listEndpoint: '/asset.vehicle.api.v1.VehicleService/ListTrips',
      responseRowsKey: 'trips',
      responseTotalKey: 'pagination.totalCount',
      columns: ['tripNumber', 'vehicleId', 'driverId', 'tripType', 'purpose', 'startLocation', 'endLocation', 'status'],
    },
    {
      slug: 'fuel-entries',
      label: 'Fuel Entries',
      formId: 'vehicle_fuel_consumption',
      listEndpoint: '/asset.vehicle.api.v1.VehicleService/ListFuelEntries',
      responseRowsKey: 'entries',
      responseTotalKey: 'pagination.totalCount',
      columns: ['vehicleId', 'entryDate', 'fuelType', 'quantityLiters', 'ratePerLiter', 'totalAmount', 'odometerReading'],
    },
    {
      slug: 'driver-assignments',
      label: 'Driver Assignments',
      formId: 'vehicle_fleet_management',
      listEndpoint: '/asset.vehicle.api.v1.VehicleService/ListDriverAssignments',
      responseRowsKey: 'assignments',
      responseTotalKey: 'pagination.totalCount',
      columns: ['vehicleId', 'driverId', 'assignmentType', 'effectiveFrom', 'effectiveTo', 'status'],
    },
    {
      slug: 'vehicle-classes',
      label: 'Vehicle Classes',
      formId: 'vehicle_fleet_management',
      listEndpoint: '/asset.vehicle.api.v1.VehicleService/ListVehicleClasses',
      responseRowsKey: 'classes',
      columns: ['name', 'displayName', 'description'],
    },
  ],
};
