/**
 * Insights Service Factories
 * Typed ConnectRPC clients for BI analytics, dashboards, reports, search
 */
import { getApiClient } from '../client/client.js';
import { BIAnalyticsService } from '@samavāya/proto/gen/business/insights/bianalytics/proto/bianalytics_pb.js';
import { DashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/dashboard_pb.js';
import { ReportService } from '@samavāya/proto/gen/business/insights/insighthub/proto/insighthub_pb.js';
import { ReportExecutionService, ReportSchedulingService, ReportSubscriptionService, ReportAlertsService, CacheService } from '@samavāya/proto/gen/business/insights/insightviewer/proto/insightviewer_pb.js';
import { SearchService } from '@samavāya/proto/gen/business/insights/metasearch/proto/metasearch_pb.js';
import { SavedSearchService } from '@samavāya/proto/gen/business/insights/savedsearch/proto/savedsearch_pb.js';
// Vertical-specific — Agriculture
import { AgricultureBIAnalyticsService } from '@samavāya/proto/gen/business/insights/bianalytics/proto/agriculture/bianalytics_agriculture_pb.js';
import { AgricultureDashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/agriculture/dashboard_agriculture_pb.js';
import { AgricultureReportService } from '@samavāya/proto/gen/business/insights/insighthub/proto/agriculture/insighthub_agriculture_pb.js';
import { AgricultureReportViewerService } from '@samavāya/proto/gen/business/insights/insightviewer/proto/agriculture/insightviewer_agriculture_pb.js';
import { AgricultureSearchService } from '@samavāya/proto/gen/business/insights/metasearch/proto/agriculture/metasearch_agriculture_pb.js';
import { AgricultureSavedSearchService } from '@samavāya/proto/gen/business/insights/savedsearch/proto/agriculture/savedsearch_agriculture_pb.js';
// Vertical-specific — Construction
import { ConstructionDashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/construction/dashboard_construction_pb.js';
// Vertical-specific — MfgVertical (Manufacturing)
import { MfgVerticalBIAnalyticsService } from '@samavāya/proto/gen/business/insights/bianalytics/proto/mfgvertical/bianalytics_mfgvertical_pb.js';
import { MfgVerticalDashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/mfgvertical/dashboard_mfgvertical_pb.js';
import { MfgVerticalReportService } from '@samavāya/proto/gen/business/insights/insighthub/proto/mfgvertical/insighthub_mfgvertical_pb.js';
import { MfgVerticalReportViewerService } from '@samavāya/proto/gen/business/insights/insightviewer/proto/mfgvertical/insightviewer_mfgvertical_pb.js';
import { MfgVerticalSearchService } from '@samavāya/proto/gen/business/insights/metasearch/proto/mfgvertical/metasearch_mfgvertical_pb.js';
import { MfgVerticalSavedSearchService } from '@samavāya/proto/gen/business/insights/savedsearch/proto/mfgvertical/savedsearch_mfgvertical_pb.js';
// Vertical-specific — Solar
import { solarBIAnalyticsService } from '@samavāya/proto/gen/business/insights/bianalytics/proto/solar/bianalytics_solar_pb.js';
import { solarDashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/solar/dashboard_solar_pb.js';
import { solarReportService } from '@samavāya/proto/gen/business/insights/insighthub/proto/solar/insighthub_solar_pb.js';
import { solarReportViewerService } from '@samavāya/proto/gen/business/insights/insightviewer/proto/solar/insightviewer_solar_pb.js';
import { solarSearchService } from '@samavāya/proto/gen/business/insights/metasearch/proto/solar/metasearch_solar_pb.js';
import { solarSavedSearchService } from '@samavāya/proto/gen/business/insights/savedsearch/proto/solar/savedsearch_solar_pb.js';
// Vertical-specific — Water
import { WaterDashboardService } from '@samavāya/proto/gen/business/insights/dashboard/proto/water/dashboard_water_pb.js';
export { BIAnalyticsService, DashboardService, ReportService, ReportExecutionService, ReportSchedulingService, ReportSubscriptionService, ReportAlertsService, CacheService, SearchService, SavedSearchService, };
export function getBIAnalyticsService() {
    return getApiClient().getService(BIAnalyticsService);
}
export function getDashboardService() {
    return getApiClient().getService(DashboardService);
}
export function getReportService() {
    return getApiClient().getService(ReportService);
}
export function getReportExecutionService() {
    return getApiClient().getService(ReportExecutionService);
}
export function getReportSchedulingService() {
    return getApiClient().getService(ReportSchedulingService);
}
export function getReportSubscriptionService() {
    return getApiClient().getService(ReportSubscriptionService);
}
export function getReportAlertsService() {
    return getApiClient().getService(ReportAlertsService);
}
export function getCacheService() {
    return getApiClient().getService(CacheService);
}
export function getSearchService() {
    return getApiClient().getService(SearchService);
}
export function getSavedSearchService() {
    return getApiClient().getService(SavedSearchService);
}
export { AgricultureBIAnalyticsService, MfgVerticalBIAnalyticsService, solarBIAnalyticsService, AgricultureDashboardService, ConstructionDashboardService, MfgVerticalDashboardService, solarDashboardService, WaterDashboardService, AgricultureReportService, MfgVerticalReportService, solarReportService, AgricultureReportViewerService, MfgVerticalReportViewerService, solarReportViewerService, AgricultureSearchService, MfgVerticalSearchService, solarSearchService, AgricultureSavedSearchService, MfgVerticalSavedSearchService, solarSavedSearchService, };
// ─── Agriculture Vertical Factories ───
export function getAgricultureBIAnalyticsService() {
    return getApiClient().getService(AgricultureBIAnalyticsService);
}
export function getAgricultureDashboardService() {
    return getApiClient().getService(AgricultureDashboardService);
}
export function getAgricultureReportService() {
    return getApiClient().getService(AgricultureReportService);
}
export function getAgricultureReportViewerService() {
    return getApiClient().getService(AgricultureReportViewerService);
}
export function getAgricultureSearchService() {
    return getApiClient().getService(AgricultureSearchService);
}
export function getAgricultureSavedSearchService() {
    return getApiClient().getService(AgricultureSavedSearchService);
}
// ─── Construction Vertical Factories ───
export function getConstructionDashboardService() {
    return getApiClient().getService(ConstructionDashboardService);
}
// ─── MfgVertical (Manufacturing) Vertical Factories ───
export function getMfgVerticalBIAnalyticsService() {
    return getApiClient().getService(MfgVerticalBIAnalyticsService);
}
export function getMfgVerticalDashboardService() {
    return getApiClient().getService(MfgVerticalDashboardService);
}
export function getMfgVerticalReportService() {
    return getApiClient().getService(MfgVerticalReportService);
}
export function getMfgVerticalReportViewerService() {
    return getApiClient().getService(MfgVerticalReportViewerService);
}
export function getMfgVerticalSearchService() {
    return getApiClient().getService(MfgVerticalSearchService);
}
export function getMfgVerticalSavedSearchService() {
    return getApiClient().getService(MfgVerticalSavedSearchService);
}
// ─── Solar Vertical Factories ───
export function getSolarBIAnalyticsService() {
    return getApiClient().getService(solarBIAnalyticsService);
}
export function getSolarDashboardService() {
    return getApiClient().getService(solarDashboardService);
}
export function getSolarReportService() {
    return getApiClient().getService(solarReportService);
}
export function getSolarReportViewerService() {
    return getApiClient().getService(solarReportViewerService);
}
export function getSolarSearchService() {
    return getApiClient().getService(solarSearchService);
}
export function getSolarSavedSearchService() {
    return getApiClient().getService(solarSavedSearchService);
}
// ─── Water Vertical Factories ───
export function getWaterDashboardService() {
    return getApiClient().getService(WaterDashboardService);
}
//# sourceMappingURL=insights.js.map