export interface FileValidationOptions {
    maxSize?: number;
    minSize?: number;
    allowedTypes?: string[];
    allowedExtensions?: string[];
    maxFiles?: number;
}
export interface FileUploadResult {
    success: boolean;
    file?: File;
    error?: string;
    url?: string;
    id?: string;
}
export interface FileProcessingOptions {
    resize?: {
        width?: number;
        height?: number;
        quality?: number;
    };
    compress?: boolean;
    watermark?: {
        text: string;
        position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'center';
    };
}
export declare class FileValidator {
    private options;
    constructor(options?: FileValidationOptions);
    validate(file: File): {
        isValid: boolean;
        errors: string[];
    };
    validateMultiple(files: File[]): {
        isValid: boolean;
        errors: string[];
        validFiles: File[];
    };
    private formatFileSize;
    private getFileExtension;
}
export declare class FileProcessor {
    static readAsBase64(file: File): Promise<string>;
    static readAsText(file: File): Promise<string>;
    static readAsArrayBuffer(file: File): Promise<ArrayBuffer>;
    static compressImage(file: File, options?: {
        maxWidth?: number;
        maxHeight?: number;
        quality?: number;
    }): Promise<File>;
    static getFileIcon(file: File): string;
    static formatFileSize(bytes: number): string;
    static generateFileId(): string;
    static createThumbnail(file: File, size?: number): Promise<string>;
}
export declare const FileTypes: {
    IMAGES: string[];
    DOCUMENTS: string[];
    ARCHIVES: string[];
    VIDEOS: string[];
    AUDIOS: string[];
    TEXT: string[];
};
export declare const FileSizeLimits: {
    SMALL: number;
    MEDIUM: number;
    LARGE: number;
    XLARGE: number;
    XXLARGE: number;
};
//# sourceMappingURL=index.d.ts.map