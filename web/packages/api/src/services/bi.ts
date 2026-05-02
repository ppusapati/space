/**
 * BI Service Factories
 * Typed ConnectRPC clients for BI analytics, dashboards, reports, search,
 * query engine, data connectors, NL query, real-time streaming, and embeds.
 *
 * This is the canonical BI service file. For backward compatibility,
 * insights.ts is preserved but new services should be added here.
 *
 * Cleanup history (2026-04-19):
 *   - `savedsearch` module retired on the backend (Sprint 7.T6, never wired
 *     in prod). All SavedSearch* factories removed. Folder + share
 *     functionality is now in core/bi/report.
 *   - `metasearch` module relocated to core/platform/search (Sprint 7.T7).
 *     Import paths updated here; proto package name unchanged.
 *   - `queryengine`, `nlquery`, `analytics` modules retired on the backend
 *     (all four: never wired per roadmap). Their TODO blocks below are
 *     therefore permanently resolved — the new service factories live
 *     alongside each new core/bi/* service when they're wired in
 *     (follow-up task: point at core/bi/query / core/bi/report /
 *     core/bi/presentation / core/bi/dataset once their proto bindings
 *     are regenerated on the frontend).
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

// ─── Core BI Services ───

import { BIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/bianalytics_pb.js';
import { DashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/dashboard_pb.js';

export { BIAnalyticsService, DashboardService };

export function getBIAnalyticsService(): Client<typeof BIAnalyticsService> {
  return getApiClient().getService(BIAnalyticsService);
}

export function getDashboardService(): Client<typeof DashboardService> {
  return getApiClient().getService(DashboardService);
}

// ─── Report & Execution Services ───

import { ReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/insighthub_pb.js';
import {
  ReportExecutionService,
  ReportSchedulingService,
  ReportSubscriptionService,
  ReportAlertsService,
  CacheService,
} from '@chetana/proto/gen/business/insights/insightviewer/proto/insightviewer_pb.js';

export {
  ReportService,
  ReportExecutionService,
  ReportSchedulingService,
  ReportSubscriptionService,
  ReportAlertsService,
  CacheService,
};

export function getReportService(): Client<typeof ReportService> {
  return getApiClient().getService(ReportService);
}

export function getReportExecutionService(): Client<typeof ReportExecutionService> {
  return getApiClient().getService(ReportExecutionService);
}

export function getReportSchedulingService(): Client<typeof ReportSchedulingService> {
  return getApiClient().getService(ReportSchedulingService);
}

export function getReportSubscriptionService(): Client<typeof ReportSubscriptionService> {
  return getApiClient().getService(ReportSubscriptionService);
}

export function getReportAlertsService(): Client<typeof ReportAlertsService> {
  return getApiClient().getService(ReportAlertsService);
}

export function getCacheService(): Client<typeof CacheService> {
  return getApiClient().getService(CacheService);
}

// ─── Search Services ───

import { SearchService } from '@chetana/proto/gen/core/platform/search/proto/metasearch_pb.js';

export { SearchService };

export function getSearchService(): Client<typeof SearchService> {
  return getApiClient().getService(SearchService);
}

// ─── Vertical Services ───

// Agriculture
import { AgricultureBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/agriculture/bianalytics_agriculture_pb.js';
import { AgricultureDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/agriculture/dashboard_agriculture_pb.js';
import { AgricultureReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/agriculture/insighthub_agriculture_pb.js';
import { AgricultureReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/agriculture/insightviewer_agriculture_pb.js';
import { AgricultureSearchService } from '@chetana/proto/gen/core/platform/search/proto/agriculture/metasearch_agriculture_pb.js';

export {
  AgricultureBIAnalyticsService,
  AgricultureDashboardService,
  AgricultureReportService,
  AgricultureReportViewerService,
  AgricultureSearchService,
};

export function getAgricultureBIAnalyticsService(): Client<typeof AgricultureBIAnalyticsService> {
  return getApiClient().getService(AgricultureBIAnalyticsService);
}

export function getAgricultureDashboardService(): Client<typeof AgricultureDashboardService> {
  return getApiClient().getService(AgricultureDashboardService);
}

export function getAgricultureReportService(): Client<typeof AgricultureReportService> {
  return getApiClient().getService(AgricultureReportService);
}

export function getAgricultureReportViewerService(): Client<typeof AgricultureReportViewerService> {
  return getApiClient().getService(AgricultureReportViewerService);
}

export function getAgricultureSearchService(): Client<typeof AgricultureSearchService> {
  return getApiClient().getService(AgricultureSearchService);
}

// Construction
import { ConstructionDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/construction/dashboard_construction_pb.js';

export { ConstructionDashboardService };

export function getConstructionDashboardService(): Client<typeof ConstructionDashboardService> {
  return getApiClient().getService(ConstructionDashboardService);
}

// Manufacturing (MfgVertical)
import { MfgVerticalBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/mfgvertical/bianalytics_mfgvertical_pb.js';
import { MfgVerticalDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/mfgvertical/dashboard_mfgvertical_pb.js';
import { MfgVerticalReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/mfgvertical/insighthub_mfgvertical_pb.js';
import { MfgVerticalReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/mfgvertical/insightviewer_mfgvertical_pb.js';
import { MfgVerticalSearchService } from '@chetana/proto/gen/core/platform/search/proto/mfgvertical/metasearch_mfgvertical_pb.js';

export {
  MfgVerticalBIAnalyticsService,
  MfgVerticalDashboardService,
  MfgVerticalReportService,
  MfgVerticalReportViewerService,
  MfgVerticalSearchService,
};

export function getMfgVerticalBIAnalyticsService(): Client<typeof MfgVerticalBIAnalyticsService> {
  return getApiClient().getService(MfgVerticalBIAnalyticsService);
}

export function getMfgVerticalDashboardService(): Client<typeof MfgVerticalDashboardService> {
  return getApiClient().getService(MfgVerticalDashboardService);
}

export function getMfgVerticalReportService(): Client<typeof MfgVerticalReportService> {
  return getApiClient().getService(MfgVerticalReportService);
}

export function getMfgVerticalReportViewerService(): Client<typeof MfgVerticalReportViewerService> {
  return getApiClient().getService(MfgVerticalReportViewerService);
}

export function getMfgVerticalSearchService(): Client<typeof MfgVerticalSearchService> {
  return getApiClient().getService(MfgVerticalSearchService);
}

// Solar
import { solarBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/solar/bianalytics_solar_pb.js';
import { solarDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/solar/dashboard_solar_pb.js';
import { solarReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/solar/insighthub_solar_pb.js';
import { solarReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/solar/insightviewer_solar_pb.js';
import { solarSearchService } from '@chetana/proto/gen/core/platform/search/proto/solar/metasearch_solar_pb.js';

export {
  solarBIAnalyticsService,
  solarDashboardService,
  solarReportService,
  solarReportViewerService,
  solarSearchService,
};

export function getSolarBIAnalyticsService(): Client<typeof solarBIAnalyticsService> {
  return getApiClient().getService(solarBIAnalyticsService);
}

export function getSolarDashboardService(): Client<typeof solarDashboardService> {
  return getApiClient().getService(solarDashboardService);
}

export function getSolarReportService(): Client<typeof solarReportService> {
  return getApiClient().getService(solarReportService);
}

export function getSolarReportViewerService(): Client<typeof solarReportViewerService> {
  return getApiClient().getService(solarReportViewerService);
}

export function getSolarSearchService(): Client<typeof solarSearchService> {
  return getApiClient().getService(solarSearchService);
}

// Water
import { WaterDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/water/dashboard_water_pb.js';

export { WaterDashboardService };

export function getWaterDashboardService(): Client<typeof WaterDashboardService> {
  return getApiClient().getService(WaterDashboardService);
}

// ─────────────────────────────────────────────────────────────────────────
// New core/bi/* services (Phase A.2, 2026-04-19)
// ─────────────────────────────────────────────────────────────────────────
//
// These are the canonical BI services from the backend BI consolidation
// (Sprints 3-6). Legacy services above (insighthub ReportService,
// bianalytics BIAnalyticsService, dashboard DashboardService) remain
// available during the migration window. New frontend code should use
// the factories below. Legacy factories retire per Sprint 7.T6 once
// render-parity validates.
//
// Collision note: `ReportService`, `DashboardService`, and
// `DatasetService`-style names already exist in the legacy proto
// bindings, so we import the new ones under aliased names
// (`BIDatasetService` etc.) and expose them with `getBIDatasetService`-
// style factory names. This keeps the two ecosystems disambiguated at
// the type level while the migration is in progress.

import { DatasetService as BIDatasetService } from '@chetana/proto/gen/core/bi/dataset/proto/dataset_pb.js';
import { QueryService as BIQueryService } from '@chetana/proto/gen/core/bi/query/proto/query_pb.js';
import { ReportService as BIReportService } from '@chetana/proto/gen/core/bi/report/proto/report_pb.js';
import { PresentationService as BIPresentationService } from '@chetana/proto/gen/core/bi/presentation/proto/presentation_pb.js';

export {
  BIDatasetService,
  BIQueryService,
  BIReportService,
  BIPresentationService,
};

/**
 * BI Dataset service — semantic-layer metadata (datasets, fields,
 * relationships, calculated fields, permissions). Backend: core/bi/dataset.
 */
export function getBIDatasetService(): Client<typeof BIDatasetService> {
  return getApiClient().getService(BIDatasetService);
}

/**
 * BI Query service — execute visual queries + validate + generate SQL +
 * NL query + EXPLAIN passthrough. Backend: core/bi/query.
 */
export function getBIQueryService(): Client<typeof BIQueryService> {
  return getApiClient().getService(BIQueryService);
}

/**
 * BI Report service — saved reports (designer + runs + export +
 * schedule + subscription + alert + folder). Backend: core/bi/report.
 *
 * Distinct from the legacy insighthub `ReportService` exposed above —
 * new surface with a different RPC set.
 */
export function getBIReportService(): Client<typeof BIReportService> {
  return getApiClient().getService(BIReportService);
}

/**
 * BI Presentation service — dashboards, widgets, embeds, realtime,
 * sharing, versioning. Backend: core/bi/presentation.
 */
export function getBIPresentationService(): Client<typeof BIPresentationService> {
  return getApiClient().getService(BIPresentationService);
}
