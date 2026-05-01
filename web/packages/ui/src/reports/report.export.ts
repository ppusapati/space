/**
 * Report Export — CSV, XLSX, PDF, Print for entire reports.
 * Builds on tables/table.export.ts and adds:
 *  - Full-report PDF with KPIs + charts + tables in one document
 *  - Per-widget table exports
 *  - Chart image capture and embed into PDF
 *  - Report-level CSV/XLSX flattening all widgets
 */

import type { TableColumn } from '../tables/table.types';
import { exportToCSV, exportToXLSX, exportToPDF, printTable } from '../tables/table.export';
import type { ExportOptions, ExportFormat } from '../tables/table.export';
import {
  buildTableColumns,
  formatValue,
  computeAggregate,
  computePivot,
} from './report.logic';
import type {
  ReportVisualization,
  ReportWidget,
  ReportData,
  WidgetKPIConfig,
  WidgetTableConfig,
  ReportFieldFormat,
  KPIAggregate,
} from './report.types';

// Re-export for convenience
export type { ExportFormat } from '../tables/table.export';

// ============================================================================
// TYPES
// ============================================================================

export type ReportExportFormat = 'csv' | 'xlsx' | 'pdf' | 'print';

export interface ReportExportOptions {
  /** Report visualization schema */
  visualization: ReportVisualization;
  /** Report data */
  data: ReportData;
  /** Filename without extension */
  filename: string;
  /** Report title (used in PDF header) */
  title?: string;
  /** Report subtitle */
  subtitle?: string;
  /** Company name for header */
  companyName?: string;
  /** Chart image captures — map of widget_id → dataURL (png) */
  chartImages?: Map<string, string>;
  /** Include timestamp in filename */
  includeTimestamp?: boolean;
}

export interface WidgetExportOptions {
  /** Widget to export */
  widget: ReportWidget;
  /** Data rows */
  rows: Record<string, unknown>[];
  /** Filename without extension */
  filename: string;
  /** Format */
  format: ReportExportFormat;
  /** Title override */
  title?: string;
}

// ============================================================================
// SINGLE WIDGET EXPORT
// ============================================================================

/**
 * Export a single table widget's data.
 */
export async function exportWidget(options: WidgetExportOptions): Promise<void> {
  const { widget, rows, filename, format, title } = options;

  if (widget.widget_type === 'table' && widget.table_config) {
    const columns = buildTableColumns(widget.table_config);
    const exportOpts: ExportOptions = {
      data: rows,
      columns,
      filename,
      format: format === 'print' ? 'print' : format,
      title: title ?? widget.title,
      includeHeaders: true,
    };

    switch (format) {
      case 'csv':
        exportToCSV(exportOpts);
        break;
      case 'xlsx':
        await exportToXLSX(exportOpts);
        break;
      case 'pdf':
        await exportToPDF(exportOpts);
        break;
      case 'print':
        printTable(exportOpts);
        break;
    }
    return;
  }

  if (widget.widget_type === 'kpi_card' && widget.kpi_config) {
    // Export KPI as single-row table
    const kpi = widget.kpi_config;
    const value = computeAggregate(rows, kpi.value_field_code, kpi.aggregate);
    const columns: TableColumn[] = [
      { key: 'label', header: 'Metric' },
      { key: 'value', header: 'Value' },
    ];
    const data = [{ label: kpi.label, value: formatValue(value, kpi.format) }];
    const exportOpts: ExportOptions = { data, columns, filename, format, title: title ?? widget.title };
    switch (format) {
      case 'csv': exportToCSV(exportOpts); break;
      case 'xlsx': await exportToXLSX(exportOpts); break;
      case 'pdf': await exportToPDF(exportOpts); break;
      case 'print': printTable(exportOpts); break;
    }
    return;
  }

  // Generic fallback — export raw data rows
  if (rows.length > 0) {
    const keys = Object.keys(rows[0]);
    const columns: TableColumn[] = keys.map((k) => ({ key: k, header: k }));
    const exportOpts: ExportOptions = { data: rows, columns, filename, format, title: title ?? widget.title };
    switch (format) {
      case 'csv': exportToCSV(exportOpts); break;
      case 'xlsx': await exportToXLSX(exportOpts); break;
      case 'pdf': await exportToPDF(exportOpts); break;
      case 'print': printTable(exportOpts); break;
    }
  }
}

// ============================================================================
// FULL REPORT EXPORT — CSV / XLSX
// ============================================================================

/**
 * Export all report data as a flat CSV (all rows, all available columns).
 */
export function exportReportCSV(options: ReportExportOptions): void {
  const { data, filename, title } = options;
  const columns = dataColumnsToTableColumns(data);
  exportToCSV({
    data: data.rows,
    columns,
    filename,
    format: 'csv',
    title,
    includeHeaders: true,
  });
}

/**
 * Export all report data as XLSX with one sheet per table/KPI widget + one "Raw Data" sheet.
 */
export async function exportReportXLSX(options: ReportExportOptions): Promise<void> {
  const { visualization, data, filename, title } = options;

  try {
    const XLSX = await import('xlsx');
    const wb = XLSX.utils.book_new();

    // Sheet 1: Raw Data
    const rawColumns = dataColumnsToTableColumns(data);
    const rawHeader = rawColumns.map((c) => c.header);
    const rawRows = data.rows.map((row) =>
      rawColumns.map((c) => row[c.key] ?? '')
    );
    const rawSheet = XLSX.utils.aoa_to_sheet([rawHeader, ...rawRows]);
    rawSheet['!cols'] = rawColumns.map((c) => ({ wch: Math.max(c.header.length, 15) }));
    XLSX.utils.book_append_sheet(wb, rawSheet, 'Raw Data');

    // Sheet per KPI summary
    const kpiWidgets = visualization.widgets.filter(
      (w) => w.widget_type === 'kpi_card' && w.kpi_config
    );
    if (kpiWidgets.length > 0) {
      const kpiRows = kpiWidgets.map((w) => {
        const kpi = w.kpi_config!;
        const val = data.aggregates?.[kpi.value_field_code] ??
          computeAggregate(data.rows, kpi.value_field_code, kpi.aggregate);
        return [kpi.label, formatValue(val, kpi.format)];
      });
      const kpiSheet = XLSX.utils.aoa_to_sheet([['Metric', 'Value'], ...kpiRows]);
      kpiSheet['!cols'] = [{ wch: 30 }, { wch: 20 }];
      XLSX.utils.book_append_sheet(wb, kpiSheet, 'KPIs');
    }

    // Sheet per table widget
    const tableWidgets = visualization.widgets.filter(
      (w) => w.widget_type === 'table' && w.table_config
    );
    for (const tw of tableWidgets) {
      const cols = buildTableColumns(tw.table_config!);
      const header = cols.map((c) => c.header);
      const rows = data.rows.map((row) => cols.map((c) => row[c.key] ?? ''));
      const ws = XLSX.utils.aoa_to_sheet([header, ...rows]);
      ws['!cols'] = cols.map((c) => ({ wch: Math.max(c.header.length, 15) }));
      const sheetName = (tw.title || 'Table').slice(0, 31); // Excel 31 char limit
      XLSX.utils.book_append_sheet(wb, ws, sheetName);
    }

    XLSX.writeFile(wb, `${filename}.xlsx`);
  } catch (error) {
    console.error('Report XLSX export failed:', error);
    throw new Error('XLSX export failed. Make sure xlsx package is installed.');
  }
}

// ============================================================================
// FULL REPORT EXPORT — PDF
// ============================================================================

/**
 * Export full report as PDF with KPI summary, chart images, and data tables.
 */
export async function exportReportPDF(options: ReportExportOptions): Promise<void> {
  const { visualization, data, filename, title, subtitle, companyName, chartImages } = options;

  try {
    const jsPDFModule = await import('jspdf');
    const jsPDF = jsPDFModule.default || jsPDFModule.jsPDF;
    await import('jspdf-autotable');

    const doc = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      format: 'a4',
    });

    const pageWidth = doc.internal.pageSize.getWidth();
    let y = 15;

    // ── Header ──────────────────────────────────────────────────────────
    if (companyName) {
      doc.setFontSize(10);
      doc.setTextColor(150);
      doc.text(companyName, 14, y);
      y += 6;
    }

    if (title) {
      doc.setFontSize(18);
      doc.setTextColor(30);
      doc.text(title, 14, y);
      y += 7;
    }

    if (subtitle) {
      doc.setFontSize(10);
      doc.setTextColor(100);
      doc.text(subtitle, 14, y);
      y += 5;
    }

    // Timestamp
    doc.setFontSize(8);
    doc.setTextColor(170);
    doc.text(`Generated: ${new Date().toLocaleString()}`, 14, y);
    y += 8;

    // ── KPI Cards ───────────────────────────────────────────────────────
    const kpiWidgets = visualization.widgets.filter(
      (w) => w.widget_type === 'kpi_card' && w.kpi_config && !w.hidden
    );

    if (kpiWidgets.length > 0) {
      doc.setFontSize(12);
      doc.setTextColor(50);
      doc.text('Key Metrics', 14, y);
      y += 6;

      const kpiColWidth = (pageWidth - 28) / Math.min(kpiWidgets.length, 4);

      for (let i = 0; i < kpiWidgets.length; i++) {
        const kpi = kpiWidgets[i].kpi_config!;
        const val = data.aggregates?.[kpi.value_field_code] ??
          computeAggregate(data.rows, kpi.value_field_code, kpi.aggregate);
        const formatted = formatValue(val, kpi.format);

        const x = 14 + (i % 4) * kpiColWidth;
        const row = Math.floor(i / 4);
        const ky = y + row * 18;

        // KPI box
        doc.setFillColor(248, 249, 250);
        doc.roundedRect(x, ky, kpiColWidth - 4, 15, 2, 2, 'F');

        doc.setFontSize(14);
        doc.setTextColor(30);
        doc.text(formatted, x + 4, ky + 7);

        doc.setFontSize(8);
        doc.setTextColor(120);
        doc.text(kpi.label, x + 4, ky + 12);
      }

      y += Math.ceil(kpiWidgets.length / 4) * 18 + 4;
    }

    // ── Chart Images ────────────────────────────────────────────────────
    if (chartImages && chartImages.size > 0) {
      const chartWidgets = visualization.widgets.filter(
        (w) => w.widget_type === 'chart' && !w.hidden && chartImages.has(w.widget_id)
      );

      for (const cw of chartWidgets) {
        const imgData = chartImages.get(cw.widget_id)!;

        // Check page space
        if (y + 90 > doc.internal.pageSize.getHeight()) {
          doc.addPage();
          y = 15;
        }

        doc.setFontSize(10);
        doc.setTextColor(60);
        doc.text(cw.title, 14, y);
        y += 4;

        try {
          const imgWidth = pageWidth - 28;
          const imgHeight = 80;
          doc.addImage(imgData, 'PNG', 14, y, imgWidth, imgHeight);
          y += imgHeight + 8;
        } catch {
          doc.setFontSize(8);
          doc.setTextColor(180);
          doc.text('[Chart image could not be embedded]', 14, y + 4);
          y += 12;
        }
      }
    }

    // ── Data Tables ─────────────────────────────────────────────────────
    const tableWidgets = visualization.widgets.filter(
      (w) => w.widget_type === 'table' && w.table_config && !w.hidden
    );

    for (const tw of tableWidgets) {
      // Check page space
      if (y + 30 > doc.internal.pageSize.getHeight()) {
        doc.addPage();
        y = 15;
      }

      doc.setFontSize(10);
      doc.setTextColor(60);
      doc.text(tw.title, 14, y);
      y += 4;

      const cols = buildTableColumns(tw.table_config!);
      const headers = cols.map((c) => c.header);
      const displayRows = tw.table_config!.row_limit
        ? data.rows.slice(0, tw.table_config!.row_limit)
        : data.rows;
      const body = displayRows.map((row) =>
        cols.map((c) => {
          const v = row[c.key];
          return c.format ? c.format(v, row) : String(v ?? '');
        })
      );

      (doc as any).autoTable({
        head: [headers],
        body,
        startY: y,
        styles: { fontSize: 8, cellPadding: 2 },
        headStyles: {
          fillColor: [14, 165, 233],
          textColor: 255,
          fontStyle: 'bold',
        },
        alternateRowStyles: { fillColor: [250, 250, 250] },
        margin: { left: 14, right: 14 },
        didDrawPage: (pageData: any) => {
          y = pageData.cursor?.y ?? y;
        },
      });

      y = (doc as any).lastAutoTable?.finalY ?? y;
      y += 10;
    }

    // ── Footer ──────────────────────────────────────────────────────────
    const totalPages = (doc as any).internal.getNumberOfPages();
    for (let i = 1; i <= totalPages; i++) {
      doc.setPage(i);
      doc.setFontSize(7);
      doc.setTextColor(180);
      doc.text(
        `Page ${i} of ${totalPages}`,
        pageWidth - 30,
        doc.internal.pageSize.getHeight() - 7
      );
    }

    doc.save(`${filename}.pdf`);
  } catch (error) {
    console.error('Report PDF export failed:', error);
    throw new Error('PDF export failed. Make sure jspdf and jspdf-autotable packages are installed.');
  }
}

// ============================================================================
// FULL REPORT PRINT
// ============================================================================

/**
 * Open a print-friendly window with the full report.
 */
export function printReport(options: ReportExportOptions): void {
  const { visualization, data, title, subtitle, companyName, chartImages } = options;

  const printWindow = window.open('', '_blank');
  if (!printWindow) {
    alert('Please allow popups for printing');
    return;
  }

  let html = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>${title ?? 'Report'}</title>
      <style>
        @page { size: landscape; margin: 1cm; }
        * { box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-size: 11px; color: #222; }
        .header { text-align: center; margin-bottom: 16px; }
        .company { font-size: 14px; font-weight: bold; color: #666; }
        .title { font-size: 20px; font-weight: 700; margin: 4px 0; }
        .subtitle { font-size: 12px; color: #888; }
        .timestamp { font-size: 9px; color: #aaa; margin-top: 6px; }
        .kpi-row { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
        .kpi-box { flex: 1; min-width: 140px; padding: 10px 14px; background: #f8f9fa; border-radius: 6px; border: 1px solid #e9ecef; }
        .kpi-value { font-size: 22px; font-weight: 700; color: #111; }
        .kpi-label { font-size: 10px; color: #888; margin-top: 2px; }
        .section-title { font-size: 13px; font-weight: 600; margin: 14px 0 6px; color: #444; }
        .chart-img { max-width: 100%; margin-bottom: 14px; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 14px; }
        th, td { border: 1px solid #ddd; padding: 6px 8px; text-align: left; }
        th { background: #f0f4f8; font-weight: 600; font-size: 10px; }
        tr:nth-child(even) { background: #fafafa; }
        @media print {
          .no-print { display: none; }
          table { page-break-inside: auto; }
          tr { page-break-inside: avoid; }
        }
      </style>
    </head>
    <body>
      <div class="header">
        ${companyName ? `<div class="company">${esc(companyName)}</div>` : ''}
        ${title ? `<div class="title">${esc(title)}</div>` : ''}
        ${subtitle ? `<div class="subtitle">${esc(subtitle)}</div>` : ''}
        <div class="timestamp">Generated: ${new Date().toLocaleString()}</div>
      </div>
  `;

  // KPIs
  const kpiWidgets = visualization.widgets.filter(
    (w) => w.widget_type === 'kpi_card' && w.kpi_config && !w.hidden
  );
  if (kpiWidgets.length > 0) {
    html += `<div class="kpi-row">`;
    for (const w of kpiWidgets) {
      const kpi = w.kpi_config!;
      const val = data.aggregates?.[kpi.value_field_code] ??
        computeAggregate(data.rows, kpi.value_field_code, kpi.aggregate);
      const formatted = formatValue(val, kpi.format);
      html += `<div class="kpi-box"><div class="kpi-value">${esc(formatted)}</div><div class="kpi-label">${esc(kpi.label)}</div></div>`;
    }
    html += `</div>`;
  }

  // Charts
  if (chartImages) {
    const chartWidgets = visualization.widgets.filter(
      (w) => w.widget_type === 'chart' && !w.hidden && chartImages.has(w.widget_id)
    );
    for (const cw of chartWidgets) {
      html += `<div class="section-title">${esc(cw.title)}</div>`;
      html += `<img class="chart-img" src="${chartImages.get(cw.widget_id)!}" />`;
    }
  }

  // Tables
  const tableWidgets = visualization.widgets.filter(
    (w) => w.widget_type === 'table' && w.table_config && !w.hidden
  );
  for (const tw of tableWidgets) {
    const cols = buildTableColumns(tw.table_config!);
    html += `<div class="section-title">${esc(tw.title)}</div>`;
    html += `<table><thead><tr>${cols.map((c) => `<th>${esc(c.header)}</th>`).join('')}</tr></thead><tbody>`;

    const displayRows = tw.table_config!.row_limit
      ? data.rows.slice(0, tw.table_config!.row_limit)
      : data.rows;
    for (const row of displayRows) {
      html += `<tr>${cols.map((c) => {
        const v = row[c.key];
        const formatted = c.format ? c.format(v, row) : String(v ?? '');
        return `<td>${esc(formatted)}</td>`;
      }).join('')}</tr>`;
    }
    html += `</tbody></table>`;
  }

  html += `<script>window.onload = function() { window.print(); }<\/script></body></html>`;

  printWindow.document.write(html);
  printWindow.document.close();
}

// ============================================================================
// UNIFIED EXPORT API
// ============================================================================

/**
 * Export the full report in any supported format.
 */
export async function exportReport(
  format: ReportExportFormat,
  options: ReportExportOptions
): Promise<void> {
  switch (format) {
    case 'csv':
      exportReportCSV(options);
      break;
    case 'xlsx':
      await exportReportXLSX(options);
      break;
    case 'pdf':
      await exportReportPDF(options);
      break;
    case 'print':
      printReport(options);
      break;
    default:
      throw new Error(`Unsupported report export format: ${format}`);
  }
}

// ============================================================================
// HELPERS
// ============================================================================

/** Convert ReportData columns to TableColumn[] for export functions */
function dataColumnsToTableColumns(data: ReportData): TableColumn[] {
  return (data.columns ?? []).map((col) => ({
    key: col.field_code,
    header: col.label || col.field_code,
  }));
}

/** HTML-escape for print */
function esc(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
