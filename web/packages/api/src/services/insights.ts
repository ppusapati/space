/**
 * Insights Service Factories
 * Typed ConnectRPC clients for BI analytics, dashboards, reports, search.
 *
 * Cleanup history (2026-04-19):
 *   - `savedsearch` module retired on the backend (Sprint 7.T6, never wired
 *     in prod). All SavedSearch* factories removed. Folder + share
 *     functionality migrated to core/bi/report's Folder/ReportShare tables.
 *   - `metasearch` module relocated to core/platform/search (Sprint 7.T7).
 *     Import paths updated; proto package name unchanged for wire compat.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';

import { BIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/bianalytics_pb.js';
import { DashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/dashboard_pb.js';
import { ReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/insighthub_pb.js';
import {
  ReportExecutionService,
  ReportSchedulingService,
  ReportSubscriptionService,
  ReportAlertsService,
  CacheService,
} from '@chetana/proto/gen/business/insights/insightviewer/proto/insightviewer_pb.js';
import { SearchService } from '@chetana/proto/gen/core/platform/search/proto/metasearch_pb.js';

// Vertical-specific — Agriculture
import { AgricultureBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/agriculture/bianalytics_agriculture_pb.js';
import { AgricultureDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/agriculture/dashboard_agriculture_pb.js';
import { AgricultureReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/agriculture/insighthub_agriculture_pb.js';
import { AgricultureReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/agriculture/insightviewer_agriculture_pb.js';
import { AgricultureSearchService } from '@chetana/proto/gen/core/platform/search/proto/agriculture/metasearch_agriculture_pb.js';

// Vertical-specific — Construction
import { ConstructionDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/construction/dashboard_construction_pb.js';

// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/mfgvertical/bianalytics_mfgvertical_pb.js';
import { MfgVerticalDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/mfgvertical/dashboard_mfgvertical_pb.js';
import { MfgVerticalReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/mfgvertical/insighthub_mfgvertical_pb.js';
import { MfgVerticalReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/mfgvertical/insightviewer_mfgvertical_pb.js';
import { MfgVerticalSearchService } from '@chetana/proto/gen/core/platform/search/proto/mfgvertical/metasearch_mfgvertical_pb.js';

// Vertical-specific — Solar
import { solarBIAnalyticsService } from '@chetana/proto/gen/business/insights/bianalytics/proto/solar/bianalytics_solar_pb.js';
import { solarDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/solar/dashboard_solar_pb.js';
import { solarReportService } from '@chetana/proto/gen/business/insights/insighthub/proto/solar/insighthub_solar_pb.js';
import { solarReportViewerService } from '@chetana/proto/gen/business/insights/insightviewer/proto/solar/insightviewer_solar_pb.js';
import { solarSearchService } from '@chetana/proto/gen/core/platform/search/proto/solar/metasearch_solar_pb.js';

// Vertical-specific — Water
import { WaterDashboardService } from '@chetana/proto/gen/business/insights/dashboard/proto/water/dashboard_water_pb.js';

export {
  BIAnalyticsService,
  DashboardService,
  ReportService,
  ReportExecutionService,
  ReportSchedulingService,
  ReportSubscriptionService,
  ReportAlertsService,
  CacheService,
  SearchService,
};

export function getBIAnalyticsService(): Client<typeof BIAnalyticsService> {
  return getApiClient().getService(BIAnalyticsService);
}

export function getDashboardService(): Client<typeof DashboardService> {
  return getApiClient().getService(DashboardService);
}

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

export function getSearchService(): Client<typeof SearchService> {
  return getApiClient().getService(SearchService);
}

export {
  AgricultureBIAnalyticsService,
  MfgVerticalBIAnalyticsService,
  solarBIAnalyticsService,
  AgricultureDashboardService,
  ConstructionDashboardService,
  MfgVerticalDashboardService,
  solarDashboardService,
  WaterDashboardService,
  AgricultureReportService,
  MfgVerticalReportService,
  solarReportService,
  AgricultureReportViewerService,
  MfgVerticalReportViewerService,
  solarReportViewerService,
  AgricultureSearchService,
  MfgVerticalSearchService,
  solarSearchService,
};

// ─── Agriculture Vertical Factories ───

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

// ─── Construction Vertical Factories ───

export function getConstructionDashboardService(): Client<typeof ConstructionDashboardService> {
  return getApiClient().getService(ConstructionDashboardService);
}

// ─── MfgVertical (Manufacturing) Vertical Factories ───

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

// ─── Solar Vertical Factories ───

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

// ─── Water Vertical Factories ───

export function getWaterDashboardService(): Client<typeof WaterDashboardService> {
  return getApiClient().getService(WaterDashboardService);
}
