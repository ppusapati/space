/**
 * Table export functionality - CSV, XLSX, PDF, JSON, Print
 */

import type { TableColumn } from './table.types';
import { getNestedValue } from './table.logic';

export type ExportFormat = 'csv' | 'xlsx' | 'pdf' | 'json' | 'print';

export interface ExportOptions<T = Record<string, unknown>> {
  /** Data to export */
  data: T[];
  /** Column definitions */
  columns: TableColumn<T>[];
  /** Filename without extension */
  filename: string;
  /** Export format */
  format: ExportFormat;
  /** Include headers */
  includeHeaders?: boolean;
  /** Custom title for PDF/Print */
  title?: string;
  /** Custom subtitle */
  subtitle?: string;
  /** Include timestamp in filename */
  includeTimestamp?: boolean;
  /** Custom date format function */
  dateFormatter?: (date: Date) => string;
  /** PDF/Print orientation */
  orientation?: 'portrait' | 'landscape' | 'auto';
  /** PDF page size */
  pageSize?: 'a4' | 'letter' | 'legal';
  /** Include footer with page numbers (PDF) */
  includeFooter?: boolean;
  /** Company/Organization name for header */
  companyName?: string;
  /** Additional metadata */
  metadata?: Record<string, string>;
}

/**
 * Export data to CSV format
 */
export function exportToCSV<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): void {
  const { data, columns, filename, includeHeaders = true } = options;

  // Filter visible columns
  const visibleColumns = columns.filter(c => c.visible !== false);

  const rows: string[][] = [];

  // Add header row
  if (includeHeaders) {
    rows.push(visibleColumns.map(c => c.header));
  }

  // Add data rows
  for (const item of data) {
    const row = visibleColumns.map(column => {
      const value = getNestedValue(item, column.key);
      const formatted = column.format
        ? column.format(value, item)
        : String(value ?? '');
      // Escape quotes and wrap in quotes if contains comma, newline, or quote
      if (formatted.includes(',') || formatted.includes('\n') || formatted.includes('"')) {
        return `"${formatted.replace(/"/g, '""')}"`;
      }
      return formatted;
    });
    rows.push(row);
  }

  // Convert to CSV string
  const csvContent = rows.map(row => row.join(',')).join('\n');

  // Download
  downloadFile(csvContent, `${filename}.csv`, 'text/csv;charset=utf-8;');
}

/**
 * Export data to XLSX format
 * Requires xlsx library to be installed
 */
export async function exportToXLSX<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): Promise<void> {
  const { data, columns, filename, includeHeaders = true, title } = options;

  try {
    // Dynamic import to avoid bundling if not used
    const XLSX = await import('xlsx');

    // Filter visible columns
    const visibleColumns = columns.filter(c => c.visible !== false);

    // Prepare data for worksheet
    const wsData: unknown[][] = [];

    // Add title row if provided
    if (title) {
      wsData.push([title]);
      wsData.push([]); // Empty row
    }

    // Add header row
    if (includeHeaders) {
      wsData.push(visibleColumns.map(c => c.header));
    }

    // Add data rows
    for (const item of data) {
      const row = visibleColumns.map(column => {
        const value = getNestedValue(item, column.key);
        if (column.format) {
          return column.format(value, item);
        }
        return value;
      });
      wsData.push(row);
    }

    // Create workbook and worksheet
    const wb = XLSX.utils.book_new();
    const ws = XLSX.utils.aoa_to_sheet(wsData);

    // Set column widths
    ws['!cols'] = visibleColumns.map(column => ({
      wch: Math.max(
        column.header.length,
        column.width ? parseInt(column.width) / 8 : 15
      ),
    }));

    XLSX.utils.book_append_sheet(wb, ws, 'Data');

    // Download
    XLSX.writeFile(wb, `${filename}.xlsx`);
  } catch (error) {
    console.error('Failed to export to XLSX:', error);
    throw new Error('XLSX export failed. Make sure xlsx package is installed.');
  }
}

/**
 * Export data to PDF format
 * Requires jspdf and jspdf-autotable libraries to be installed
 */
export async function exportToPDF<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): Promise<void> {
  const { data, columns, filename, title } = options;

  try {
    // Dynamic imports
    const jsPDFModule = await import('jspdf');
    const jsPDF = jsPDFModule.default || jsPDFModule.jsPDF;
    await import('jspdf-autotable');

    // Filter visible columns
    const visibleColumns = columns.filter(c => c.visible !== false);

    // Create PDF document
    const doc = new jsPDF({
      orientation: visibleColumns.length > 6 ? 'landscape' : 'portrait',
      unit: 'mm',
      format: 'a4',
    });

    // Add title if provided
    if (title) {
      doc.setFontSize(16);
      doc.text(title, 14, 15);
    }

    // Prepare table data
    const headers = visibleColumns.map(c => c.header);
    const body = data.map(item =>
      visibleColumns.map(column => {
        const value = getNestedValue(item, column.key);
        return column.format
          ? column.format(value, item)
          : String(value ?? '');
      })
    );

    // Add table using autotable
    (doc as any).autoTable({
      head: [headers],
      body,
      startY: title ? 25 : 15,
      styles: {
        fontSize: 9,
        cellPadding: 3,
      },
      headStyles: {
        fillColor: [14, 165, 233], // brand-primary-500
        textColor: 255,
        fontStyle: 'bold',
      },
      alternateRowStyles: {
        fillColor: [250, 250, 250],
      },
      margin: { top: 15 },
    });

    // Download
    doc.save(`${filename}.pdf`);
  } catch (error) {
    console.error('Failed to export to PDF:', error);
    throw new Error('PDF export failed. Make sure jspdf and jspdf-autotable packages are installed.');
  }
}

/**
 * Export data to JSON format
 */
export function exportToJSON<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): void {
  const { data, columns, filename, includeTimestamp } = options;

  // Filter visible columns
  const visibleColumns = columns.filter(c => c.visible !== false);

  // Transform data to include only visible columns with formatted values
  const exportData = data.map(item => {
    const row: Record<string, unknown> = {};
    for (const column of visibleColumns) {
      const value = getNestedValue(item, column.key);
      row[column.key as string] = column.format ? column.format(value, item) : value;
    }
    return row;
  });

  const jsonContent = JSON.stringify(exportData, null, 2);
  const finalFilename = includeTimestamp
    ? `${filename}_${formatTimestamp(new Date())}.json`
    : `${filename}.json`;

  downloadFile(jsonContent, finalFilename, 'application/json');
}

/**
 * Print table data
 */
export function printTable<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): void {
  const { data, columns, title, subtitle, companyName, orientation = 'auto' } = options;

  // Filter visible columns
  const visibleColumns = columns.filter(c => c.visible !== false);
  const effectiveOrientation = orientation === 'auto'
    ? (visibleColumns.length > 6 ? 'landscape' : 'portrait')
    : orientation;

  // Build print content
  const printWindow = window.open('', '_blank');
  if (!printWindow) {
    alert('Please allow popups for printing');
    return;
  }

  const tableRows = data.map(item =>
    `<tr>${visibleColumns.map(column => {
      const value = getNestedValue(item, column.key);
      const formatted = column.format ? column.format(value, item) : String(value ?? '');
      return `<td>${escapeHtml(formatted)}</td>`;
    }).join('')}</tr>`
  ).join('');

  const html = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>${title || 'Print'}</title>
      <style>
        @page { size: ${effectiveOrientation}; margin: 1cm; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-size: 12px; }
        .header { text-align: center; margin-bottom: 20px; }
        .company { font-size: 18px; font-weight: bold; margin-bottom: 5px; }
        .title { font-size: 16px; font-weight: 600; margin-bottom: 3px; }
        .subtitle { font-size: 12px; color: #666; }
        .timestamp { font-size: 10px; color: #999; margin-top: 10px; }
        table { width: 100%; border-collapse: collapse; margin-top: 10px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f5f5f5; font-weight: 600; }
        tr:nth-child(even) { background-color: #fafafa; }
        @media print {
          .no-print { display: none; }
          table { page-break-inside: auto; }
          tr { page-break-inside: avoid; }
        }
      </style>
    </head>
    <body>
      <div class="header">
        ${companyName ? `<div class="company">${escapeHtml(companyName)}</div>` : ''}
        ${title ? `<div class="title">${escapeHtml(title)}</div>` : ''}
        ${subtitle ? `<div class="subtitle">${escapeHtml(subtitle)}</div>` : ''}
        <div class="timestamp">Generated on ${new Date().toLocaleString()}</div>
      </div>
      <table>
        <thead>
          <tr>${visibleColumns.map(c => `<th>${escapeHtml(c.header)}</th>`).join('')}</tr>
        </thead>
        <tbody>${tableRows}</tbody>
      </table>
      <script>window.onload = function() { window.print(); }</script>
    </body>
    </html>
  `;

  printWindow.document.write(html);
  printWindow.document.close();
}

/**
 * Export data to specified format
 */
export async function exportData<T extends Record<string, unknown>>(
  options: ExportOptions<T>
): Promise<void> {
  switch (options.format) {
    case 'csv':
      exportToCSV(options);
      break;
    case 'xlsx':
      await exportToXLSX(options);
      break;
    case 'pdf':
      await exportToPDF(options);
      break;
    case 'json':
      exportToJSON(options);
      break;
    case 'print':
      printTable(options);
      break;
    default:
      throw new Error(`Unsupported export format: ${options.format}`);
  }
}

/**
 * Helper function to download file
 */
function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Get export format label
 */
export function getExportFormatLabel(format: ExportFormat): string {
  switch (format) {
    case 'csv':
      return 'CSV';
    case 'xlsx':
      return 'Excel (XLSX)';
    case 'pdf':
      return 'PDF';
    default:
      return format.toUpperCase();
  }
}

/**
 * Get export format icon (returns SVG path)
 */
export function getExportFormatIcon(format: ExportFormat): string {
  switch (format) {
    case 'csv':
      return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
    case 'xlsx':
      return 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
    case 'pdf':
      return 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z';
    case 'json':
      return 'M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7c-2 0-3 1-3 3zm5 2h2m-2 4h6m-6 4h4';
    case 'print':
      return 'M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z';
    default:
      return 'M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
  }
}

/**
 * Format timestamp for filenames
 */
function formatTimestamp(date: Date): string {
  return date.toISOString().replace(/[:.]/g, '-').slice(0, 19);
}

/**
 * Escape HTML entities
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Get all available export formats
 */
export function getAvailableFormats(): { format: ExportFormat; label: string; icon: string }[] {
  return [
    { format: 'csv', label: 'CSV', icon: getExportFormatIcon('csv') },
    { format: 'xlsx', label: 'Excel (XLSX)', icon: getExportFormatIcon('xlsx') },
    { format: 'pdf', label: 'PDF', icon: getExportFormatIcon('pdf') },
    { format: 'json', label: 'JSON', icon: getExportFormatIcon('json') },
    { format: 'print', label: 'Print', icon: getExportFormatIcon('print') },
  ];
}

/**
 * Create export handler with predefined options
 */
export function createExportHandler<T extends Record<string, unknown>>(
  baseOptions: Omit<ExportOptions<T>, 'format' | 'data'>
) {
  return {
    csv: (data: T[]) => exportData({ ...baseOptions, data, format: 'csv' }),
    xlsx: (data: T[]) => exportData({ ...baseOptions, data, format: 'xlsx' }),
    pdf: (data: T[]) => exportData({ ...baseOptions, data, format: 'pdf' }),
    json: (data: T[]) => exportData({ ...baseOptions, data, format: 'json' }),
    print: (data: T[]) => exportData({ ...baseOptions, data, format: 'print' }),
  };
}
