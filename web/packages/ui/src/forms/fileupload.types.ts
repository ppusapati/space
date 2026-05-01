/**
 * FileUpload component types and logic
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** Uploaded file info */
export interface UploadedFile {
  id: string;
  file: File;
  name: string;
  size: number;
  type: string;
  progress: number;
  status: 'pending' | 'uploading' | 'success' | 'error';
  error?: string;
  preview?: string;
}

/** FileUpload props interface */
export interface FileUploadProps extends FormElementProps {
  /** Current files */
  files?: UploadedFile[];
  /** Accept file types (e.g., "image/*,.pdf") */
  accept?: string;
  /** Maximum file size in bytes */
  maxSize?: number;
  /** Maximum number of files */
  maxFiles?: number;
  /** Allow multiple files */
  multiple?: boolean;
  /** Enable drag and drop */
  dragDrop?: boolean;
  /** Size */
  size?: Size;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error message */
  errorText?: string;
  /** Custom upload text */
  uploadText?: string;
  /** Show file previews */
  showPreview?: boolean;
  /** Show progress during upload */
  showProgress?: boolean;
  /** Full width */
  fullWidth?: boolean;
}

/** UnoCSS class mappings for file upload sizes */
export const fileUploadSizeClasses: Record<Size, { container: string; text: string; icon: string }> = {
  xs: { container: 'p-3', text: 'text-xs', icon: 'w-6 h-6' },
  sm: { container: 'p-4', text: 'text-sm', icon: 'w-8 h-8' },
  md: { container: 'p-6', text: 'text-base', icon: 'w-10 h-10' },
  lg: { container: 'p-8', text: 'text-lg', icon: 'w-12 h-12' },
  xl: { container: 'p-10', text: 'text-xl', icon: 'w-14 h-14' },
};

/** File upload container classes */
export const fileUploadContainerClasses = {
  base: 'border-2 border-dashed rounded-lg transition-all duration-200 cursor-pointer ' +
    'flex flex-col items-center justify-center text-center ' +
    'hover:border-brand-primary-400 hover:bg-brand-primary-50',
  default: 'border-neutral-300 bg-neutral-50',
  dragging: 'border-brand-primary-500 bg-brand-primary-50',
  error: 'border-semantic-error-500 bg-semantic-error-50',
  disabled: 'opacity-50 cursor-not-allowed hover:border-neutral-300 hover:bg-neutral-50',
};

/** File list classes */
export const fileListClasses = {
  container: 'mt-4 space-y-2',
  item: 'flex items-center gap-3 p-3 bg-neutral-50 rounded-lg',
  itemError: 'bg-semantic-error-50 border border-semantic-error-200',
  icon: 'w-8 h-8 text-neutral-400',
  info: 'flex-1 min-w-0',
  name: 'text-sm font-medium text-neutral-900 truncate',
  size: 'text-xs text-neutral-500',
  actions: 'flex items-center gap-2',
  removeBtn: 'p-1 text-neutral-400 hover:text-semantic-error-500 rounded hover:bg-neutral-100',
  progress: 'w-full h-1 bg-neutral-200 rounded-full overflow-hidden mt-1',
  progressBar: 'h-full bg-brand-primary-500 transition-all duration-300',
};

/** Preview classes */
export const previewClasses = {
  container: 'w-12 h-12 rounded overflow-hidden bg-neutral-100 flex items-center justify-center',
  image: 'w-full h-full object-cover',
};

/** Helper functions */

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function validateFile(
  file: File,
  accept?: string,
  maxSize?: number
): { valid: boolean; error?: string } {
  // Check file type
  if (accept) {
    const acceptedTypes = accept.split(',').map(t => t.trim());
    const isValidType = acceptedTypes.some(type => {
      if (type.startsWith('.')) {
        // Extension check
        return file.name.toLowerCase().endsWith(type.toLowerCase());
      } else if (type.endsWith('/*')) {
        // MIME type wildcard
        const baseType = type.slice(0, -2);
        return file.type.startsWith(baseType);
      } else {
        // Exact MIME type
        return file.type === type;
      }
    });

    if (!isValidType) {
      return { valid: false, error: `File type not allowed. Accepted: ${accept}` };
    }
  }

  // Check file size
  if (maxSize && file.size > maxSize) {
    return { valid: false, error: `File too large. Maximum size: ${formatFileSize(maxSize)}` };
  }

  return { valid: true };
}

export function createUploadedFile(file: File): UploadedFile {
  const id = `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;

  return {
    id,
    file,
    name: file.name,
    size: file.size,
    type: file.type,
    progress: 0,
    status: 'pending',
  };
}

export function getFileIcon(type: string): string {
  if (type.startsWith('image/')) return 'image';
  if (type.startsWith('video/')) return 'video';
  if (type.startsWith('audio/')) return 'audio';
  if (type === 'application/pdf') return 'pdf';
  if (type.includes('spreadsheet') || type.includes('excel')) return 'spreadsheet';
  if (type.includes('document') || type.includes('word')) return 'document';
  if (type.includes('zip') || type.includes('compressed')) return 'archive';
  return 'file';
}

export function createPreview(file: File): Promise<string | undefined> {
  return new Promise((resolve) => {
    if (!file.type.startsWith('image/')) {
      resolve(undefined);
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target?.result as string);
    reader.onerror = () => resolve(undefined);
    reader.readAsDataURL(file);
  });
}
