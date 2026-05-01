export interface FileValidationOptions {
  maxSize?: number // in bytes
  minSize?: number // in bytes
  allowedTypes?: string[] // MIME types
  allowedExtensions?: string[]
  maxFiles?: number
}

export interface FileUploadResult {
  success: boolean
  file?: File
  error?: string
  url?: string
  id?: string
}

export interface FileProcessingOptions {
  resize?: {
    width?: number
    height?: number
    quality?: number
  }
  compress?: boolean
  watermark?: {
    text: string
    position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'center'
  }
}

export class FileValidator {
  private options: FileValidationOptions

  constructor(options: FileValidationOptions = {}) {
    this.options = options
  }

  validate(file: File): { isValid: boolean; errors: string[] } {
    const errors: string[] = []

    // Check file size
    if (this.options.maxSize && file.size > this.options.maxSize) {
      errors.push(`File size must be less than ${this.formatFileSize(this.options.maxSize)}`)
    }

    if (this.options.minSize && file.size < this.options.minSize) {
      errors.push(`File size must be at least ${this.formatFileSize(this.options.minSize)}`)
    }

    // Check MIME type
    if (this.options.allowedTypes && !this.options.allowedTypes.includes(file.type)) {
      errors.push(`File type ${file.type} is not allowed. Allowed types: ${this.options.allowedTypes.join(', ')}`)
    }

    // Check file extension
    if (this.options.allowedExtensions) {
      const extension = this.getFileExtension(file.name)
      if (!this.options.allowedExtensions.includes(extension)) {
        errors.push(`File extension .${extension} is not allowed. Allowed extensions: ${this.options.allowedExtensions.map(ext => `.${ext}`).join(', ')}`)
      }
    }

    return {
      isValid: errors.length === 0,
      errors
    }
  }

  validateMultiple(files: File[]): { isValid: boolean; errors: string[]; validFiles: File[] } {
    const errors: string[] = []
    const validFiles: File[] = []

    // Check max files limit
    if (this.options.maxFiles && files.length > this.options.maxFiles) {
      errors.push(`Too many files. Maximum allowed: ${this.options.maxFiles}`)
      return { isValid: false, errors, validFiles }
    }

    // Validate each file
    files.forEach((file, index) => {
      const validation = this.validate(file)
      if (validation.isValid) {
        validFiles.push(file)
      } else {
        errors.push(`File ${index + 1} (${file.name}): ${validation.errors.join(', ')}`)
      }
    })

    return {
      isValid: errors.length === 0,
      errors,
      validFiles
    }
  }

  private formatFileSize(bytes: number): string {
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    if (bytes === 0) return '0 Bytes'
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
  }

  private getFileExtension(filename: string): string {
    return filename.split('.').pop()?.toLowerCase() || ''
  }
}

export class FileProcessor {
  static async readAsBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
  }

  static async readAsText(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = reject
      reader.readAsText(file)
    })
  }

  static async readAsArrayBuffer(file: File): Promise<ArrayBuffer> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(reader.result as ArrayBuffer)
      reader.onerror = reject
      reader.readAsArrayBuffer(file)
    })
  }

  static async compressImage(
    file: File, 
    options: { maxWidth?: number; maxHeight?: number; quality?: number } = {}
  ): Promise<File> {
    const { maxWidth = 1920, maxHeight = 1080, quality = 0.8 } = options

    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      const img = new Image()

      img.onload = () => {
        // Calculate new dimensions
        let { width, height } = img
        
        if (width > height) {
          if (width > maxWidth) {
            height = height * (maxWidth / width)
            width = maxWidth
          }
        } else {
          if (height > maxHeight) {
            width = width * (maxHeight / height)
            height = maxHeight
          }
        }

        canvas.width = width
        canvas.height = height

        // Draw and compress
        ctx?.drawImage(img, 0, 0, width, height)
        
        canvas.toBlob(
          (blob) => {
            if (blob) {
              const compressedFile = new File([blob], file.name, {
                type: file.type,
                lastModified: Date.now()
              })
              resolve(compressedFile)
            } else {
              reject(new Error('Failed to compress image'))
            }
          },
          file.type,
          quality
        )
      }

      img.onerror = reject
      img.src = URL.createObjectURL(file)
    })
  }

  static getFileIcon(file: File): string {
    const type = file.type.toLowerCase()
    const extension = file.name.split('.').pop()?.toLowerCase()

    // Image files
    if (type.startsWith('image/')) return '🖼️'
    
    // Document files
    if (type.includes('pdf')) return '📄'
    if (type.includes('word') || extension === 'doc' || extension === 'docx') return '📝'
    if (type.includes('excel') || extension === 'xls' || extension === 'xlsx') return '📊'
    if (type.includes('powerpoint') || extension === 'ppt' || extension === 'pptx') return '📽️'
    
    // Archive files
    if (type.includes('zip') || type.includes('rar') || type.includes('7z')) return '📦'
    
    // Video files
    if (type.startsWith('video/')) return '🎥'
    
    // Audio files
    if (type.startsWith('audio/')) return '🎵'
    
    // Code files
    if (['js', 'ts', 'jsx', 'tsx', 'html', 'css', 'json', 'xml'].includes(extension || '')) return '💻'
    
    // Text files
    if (type.startsWith('text/') || extension === 'txt') return '📄'
    
    return '📎' // Default file icon
  }

  static formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  static generateFileId(): string {
    return `file_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  static async createThumbnail(file: File, size = 150): Promise<string> {
    if (!file.type.startsWith('image/')) {
      throw new Error('File is not an image')
    }

    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      const img = new Image()

      img.onload = () => {
        canvas.width = size
        canvas.height = size

        // Calculate crop dimensions for square thumbnail
        const minDimension = Math.min(img.width, img.height)
        const startX = (img.width - minDimension) / 2
        const startY = (img.height - minDimension) / 2

        ctx?.drawImage(
          img,
          startX, startY, minDimension, minDimension,
          0, 0, size, size
        )

        resolve(canvas.toDataURL(file.type))
      }

      img.onerror = reject
      img.src = URL.createObjectURL(file)
    })
  }
}

// Common file type constants
export const FileTypes = {
  IMAGES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'],
  DOCUMENTS: [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'application/vnd.ms-powerpoint',
    'application/vnd.openxmlformats-officedocument.presentationml.presentation'
  ],
  ARCHIVES: ['application/zip', 'application/x-rar-compressed', 'application/x-7z-compressed'],
  VIDEOS: ['video/mp4', 'video/webm', 'video/ogg', 'video/quicktime'],
  AUDIOS: ['audio/mpeg', 'audio/wav', 'audio/ogg', 'audio/m4a'],
  TEXT: ['text/plain', 'text/csv', 'application/json', 'text/html', 'text/css', 'application/javascript']
}

export const FileSizeLimits = {
  SMALL: 1024 * 1024, // 1MB
  MEDIUM: 5 * 1024 * 1024, // 5MB
  LARGE: 10 * 1024 * 1024, // 10MB
  XLARGE: 50 * 1024 * 1024, // 50MB
  XXLARGE: 100 * 1024 * 1024 // 100MB
}