/**
 * Clipboard Utilities
 *
 * Comprehensive clipboard operations:
 * - Copy text, HTML, images
 * - Read from clipboard
 * - Paste handling
 * - File handling
 * - Format detection
 */

// ============================================================================
// Types
// ============================================================================

export type ClipboardFormat = 'text' | 'html' | 'image' | 'file' | 'json' | 'unknown';

export interface ClipboardItem {
  type: ClipboardFormat;
  data: string | Blob | File[];
  mimeType: string;
}

export interface CopyOptions {
  /** Show success notification */
  notify?: boolean;
  /** Notification message */
  message?: string;
  /** Notification duration in ms */
  duration?: number;
}

// ============================================================================
// Core Functions
// ============================================================================

/**
 * Check if Clipboard API is supported
 */
export function isClipboardSupported(): boolean {
  return typeof navigator !== 'undefined' && !!navigator.clipboard;
}

/**
 * Check if clipboard read is supported
 */
export function isClipboardReadSupported(): boolean {
  return isClipboardSupported() && typeof navigator.clipboard.read === 'function';
}

/**
 * Copy text to clipboard
 */
export async function copyText(text: string, options?: CopyOptions): Promise<boolean> {
  try {
    if (isClipboardSupported()) {
      await navigator.clipboard.writeText(text);
    } else {
      // Fallback for older browsers
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.left = '-9999px';
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
    }

    if (options?.notify) {
      showCopyNotification(options.message || 'Copied to clipboard', options.duration);
    }
    return true;
  } catch (error) {
    console.error('Failed to copy text:', error);
    return false;
  }
}

/**
 * Copy HTML content to clipboard
 */
export async function copyHTML(html: string, fallbackText?: string, options?: CopyOptions): Promise<boolean> {
  try {
    if (isClipboardSupported() && typeof ClipboardItem !== 'undefined') {
      const htmlBlob = new Blob([html], { type: 'text/html' });
      const textBlob = new Blob([fallbackText || stripHtml(html)], { type: 'text/plain' });

      const item = new ClipboardItem({
        'text/html': htmlBlob,
        'text/plain': textBlob,
      });

      await navigator.clipboard.write([item]);
    } else {
      // Fallback: copy plain text
      return copyText(fallbackText || stripHtml(html), options);
    }

    if (options?.notify) {
      showCopyNotification(options.message || 'Copied to clipboard', options.duration);
    }
    return true;
  } catch (error) {
    console.error('Failed to copy HTML:', error);
    return copyText(fallbackText || stripHtml(html), options);
  }
}

/**
 * Copy JSON data to clipboard
 */
export async function copyJSON<T>(data: T, options?: CopyOptions): Promise<boolean> {
  try {
    const jsonString = JSON.stringify(data, null, 2);
    return copyText(jsonString, options);
  } catch (error) {
    console.error('Failed to copy JSON:', error);
    return false;
  }
}

/**
 * Copy image to clipboard
 */
export async function copyImage(imageSource: string | Blob | HTMLImageElement, options?: CopyOptions): Promise<boolean> {
  try {
    if (!isClipboardSupported()) return false;

    let blob: Blob;

    if (imageSource instanceof Blob) {
      blob = imageSource;
    } else if (imageSource instanceof HTMLImageElement) {
      blob = await imageToBlob(imageSource);
    } else {
      // URL string - fetch and convert
      const response = await fetch(imageSource);
      blob = await response.blob();
    }

    // Ensure it's a PNG for clipboard compatibility
    if (blob.type !== 'image/png') {
      blob = await convertToPng(blob);
    }

    const item = new ClipboardItem({ 'image/png': blob });
    await navigator.clipboard.write([item]);

    if (options?.notify) {
      showCopyNotification(options.message || 'Image copied to clipboard', options.duration);
    }
    return true;
  } catch (error) {
    console.error('Failed to copy image:', error);
    return false;
  }
}

/**
 * Read text from clipboard
 */
export async function readText(): Promise<string | null> {
  try {
    if (isClipboardSupported()) {
      return await navigator.clipboard.readText();
    }
    return null;
  } catch (error) {
    console.error('Failed to read clipboard:', error);
    return null;
  }
}

/**
 * Read all clipboard contents
 */
export async function readClipboard(): Promise<ClipboardItem[]> {
  const items: ClipboardItem[] = [];

  if (!isClipboardReadSupported()) {
    // Fallback: try reading text only
    const text = await readText();
    if (text) {
      items.push({ type: 'text', data: text, mimeType: 'text/plain' });
    }
    return items;
  }

  try {
    const clipboardItems = await navigator.clipboard.read();

    for (const item of clipboardItems) {
      for (const type of item.types) {
        const blob = await item.getType(type);

        if (type.startsWith('text/html')) {
          items.push({ type: 'html', data: await blob.text(), mimeType: type });
        } else if (type.startsWith('text/')) {
          const text = await blob.text();
          // Try to detect JSON
          if (isValidJSON(text)) {
            items.push({ type: 'json', data: text, mimeType: 'application/json' });
          } else {
            items.push({ type: 'text', data: text, mimeType: type });
          }
        } else if (type.startsWith('image/')) {
          items.push({ type: 'image', data: blob, mimeType: type });
        } else {
          items.push({ type: 'unknown', data: blob, mimeType: type });
        }
      }
    }
  } catch (error) {
    console.error('Failed to read clipboard:', error);
  }

  return items;
}

/**
 * Handle paste event and extract content
 */
export function handlePaste(event: ClipboardEvent): ClipboardItem[] {
  const items: ClipboardItem[] = [];

  if (!event.clipboardData) return items;

  // Check for files
  if (event.clipboardData.files.length > 0) {
    items.push({
      type: 'file',
      data: Array.from(event.clipboardData.files),
      mimeType: 'application/octet-stream',
    });
  }

  // Check for HTML
  const html = event.clipboardData.getData('text/html');
  if (html) {
    items.push({ type: 'html', data: html, mimeType: 'text/html' });
  }

  // Check for text
  const text = event.clipboardData.getData('text/plain');
  if (text) {
    if (isValidJSON(text)) {
      items.push({ type: 'json', data: text, mimeType: 'application/json' });
    } else {
      items.push({ type: 'text', data: text, mimeType: 'text/plain' });
    }
  }

  return items;
}

/**
 * Create a paste handler for a specific element
 */
export function createPasteHandler(
  callback: (items: ClipboardItem[], event: ClipboardEvent) => void,
  options?: { preventDefault?: boolean }
): (event: ClipboardEvent) => void {
  return (event: ClipboardEvent) => {
    if (options?.preventDefault) event.preventDefault();
    const items = handlePaste(event);
    callback(items, event);
  };
}

// ============================================================================
// Table Data Utilities
// ============================================================================

/**
 * Copy table data to clipboard (for spreadsheet paste)
 */
export async function copyTableData<T extends Record<string, unknown>>(
  data: T[],
  columns: { key: keyof T; header: string }[],
  options?: CopyOptions
): Promise<boolean> {
  // Create TSV format for spreadsheet compatibility
  const headers = columns.map(c => c.header).join('\t');
  const rows = data.map(row =>
    columns.map(c => String(row[c.key] ?? '')).join('\t')
  );
  const tsv = [headers, ...rows].join('\n');

  // Also create HTML table for rich paste
  const html = `
    <table>
      <thead><tr>${columns.map(c => `<th>${escapeHtml(c.header)}</th>`).join('')}</tr></thead>
      <tbody>${data.map(row =>
        `<tr>${columns.map(c => `<td>${escapeHtml(String(row[c.key] ?? ''))}</td>`).join('')}</tr>`
      ).join('')}</tbody>
    </table>
  `;

  return copyHTML(html, tsv, options);
}

/**
 * Parse pasted table data (from spreadsheet)
 */
export function parseTableData(text: string): string[][] {
  // Split by newlines, handling different line endings
  const lines = text.split(/\r?\n/).filter(line => line.trim());

  // Split each line by tabs (TSV) or detect delimiter
  return lines.map(line => {
    // Check if TSV
    if (line.includes('\t')) {
      return line.split('\t');
    }
    // Check if CSV with quoted values
    return parseCSVLine(line);
  });
}

// ============================================================================
// Helper Functions
// ============================================================================

function stripHtml(html: string): string {
  const doc = new DOMParser().parseFromString(html, 'text/html');
  return doc.body.textContent || '';
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function isValidJSON(text: string): boolean {
  try {
    JSON.parse(text);
    return true;
  } catch {
    return false;
  }
}

async function imageToBlob(img: HTMLImageElement): Promise<Blob> {
  const canvas = document.createElement('canvas');
  canvas.width = img.naturalWidth;
  canvas.height = img.naturalHeight;
  const ctx = canvas.getContext('2d')!;
  ctx.drawImage(img, 0, 0);

  return new Promise((resolve, reject) => {
    canvas.toBlob(blob => {
      if (blob) resolve(blob);
      else reject(new Error('Failed to convert image to blob'));
    }, 'image/png');
  });
}

async function convertToPng(blob: Blob): Promise<Blob> {
  const img = new Image();
  const url = URL.createObjectURL(blob);

  return new Promise((resolve, reject) => {
    img.onload = async () => {
      URL.revokeObjectURL(url);
      try {
        resolve(await imageToBlob(img));
      } catch (e) {
        reject(e);
      }
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Failed to load image'));
    };
    img.src = url;
  });
}

function parseCSVLine(line: string): string[] {
  const result: string[] = [];
  let current = '';
  let inQuotes = false;

  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    const nextChar = line[i + 1];

    if (inQuotes) {
      if (char === '"' && nextChar === '"') {
        current += '"';
        i++; // Skip next quote
      } else if (char === '"') {
        inQuotes = false;
      } else {
        current += char;
      }
    } else {
      if (char === '"') {
        inQuotes = true;
      } else if (char === ',') {
        result.push(current);
        current = '';
      } else {
        current += char;
      }
    }
  }

  result.push(current);
  return result;
}

function showCopyNotification(message: string, duration = 2000) {
  // Simple notification - can be replaced with your app's notification system
  if (typeof document === 'undefined') return;

  const notification = document.createElement('div');
  notification.textContent = message;
  notification.style.cssText = `
    position: fixed;
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: #333;
    color: white;
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 14px;
    z-index: 9999;
    animation: fadeIn 0.2s ease-out;
  `;

  document.body.appendChild(notification);

  setTimeout(() => {
    notification.style.opacity = '0';
    notification.style.transition = 'opacity 0.2s';
    setTimeout(() => notification.remove(), 200);
  }, duration);
}
