class x {
  options;
  constructor(t = {}) {
    this.options = t;
  }
  validate(t) {
    const e = [];
    if (this.options.maxSize && t.size > this.options.maxSize && e.push(`File size must be less than ${this.formatFileSize(this.options.maxSize)}`), this.options.minSize && t.size < this.options.minSize && e.push(`File size must be at least ${this.formatFileSize(this.options.minSize)}`), this.options.allowedTypes && !this.options.allowedTypes.includes(t.type) && e.push(`File type ${t.type} is not allowed. Allowed types: ${this.options.allowedTypes.join(", ")}`), this.options.allowedExtensions) {
      const i = this.getFileExtension(t.name);
      this.options.allowedExtensions.includes(i) || e.push(`File extension .${i} is not allowed. Allowed extensions: ${this.options.allowedExtensions.map((s) => `.${s}`).join(", ")}`);
    }
    return {
      isValid: e.length === 0,
      errors: e
    };
  }
  validateMultiple(t) {
    const e = [], i = [];
    return this.options.maxFiles && t.length > this.options.maxFiles ? (e.push(`Too many files. Maximum allowed: ${this.options.maxFiles}`), { isValid: !1, errors: e, validFiles: i }) : (t.forEach((s, a) => {
      const p = this.validate(s);
      p.isValid ? i.push(s) : e.push(`File ${a + 1} (${s.name}): ${p.errors.join(", ")}`);
    }), {
      isValid: e.length === 0,
      errors: e,
      validFiles: i
    });
  }
  formatFileSize(t) {
    const e = ["Bytes", "KB", "MB", "GB"];
    if (t === 0) return "0 Bytes";
    const i = Math.floor(Math.log(t) / Math.log(1024));
    return Math.round(t / Math.pow(1024, i) * 100) / 100 + " " + e[i];
  }
  getFileExtension(t) {
    return t.split(".").pop()?.toLowerCase() || "";
  }
}
class w {
  static async readAsBase64(t) {
    return new Promise((e, i) => {
      const s = new FileReader();
      s.onload = () => e(s.result), s.onerror = i, s.readAsDataURL(t);
    });
  }
  static async readAsText(t) {
    return new Promise((e, i) => {
      const s = new FileReader();
      s.onload = () => e(s.result), s.onerror = i, s.readAsText(t);
    });
  }
  static async readAsArrayBuffer(t) {
    return new Promise((e, i) => {
      const s = new FileReader();
      s.onload = () => e(s.result), s.onerror = i, s.readAsArrayBuffer(t);
    });
  }
  static async compressImage(t, e = {}) {
    const { maxWidth: i = 1920, maxHeight: s = 1080, quality: a = 0.8 } = e;
    return new Promise((p, o) => {
      const n = document.createElement("canvas"), d = n.getContext("2d"), c = new Image();
      c.onload = () => {
        let { width: r, height: l } = c;
        r > l ? r > i && (l = l * (i / r), r = i) : l > s && (r = r * (s / l), l = s), n.width = r, n.height = l, d?.drawImage(c, 0, 0, r, l), n.toBlob(
          (m) => {
            if (m) {
              const u = new File([m], t.name, {
                type: t.type,
                lastModified: Date.now()
              });
              p(u);
            } else
              o(new Error("Failed to compress image"));
          },
          t.type,
          a
        );
      }, c.onerror = o, c.src = URL.createObjectURL(t);
    });
  }
  static getFileIcon(t) {
    const e = t.type.toLowerCase(), i = t.name.split(".").pop()?.toLowerCase();
    return e.startsWith("image/") ? "🖼️" : e.includes("pdf") ? "📄" : e.includes("word") || i === "doc" || i === "docx" ? "📝" : e.includes("excel") || i === "xls" || i === "xlsx" ? "📊" : e.includes("powerpoint") || i === "ppt" || i === "pptx" ? "📽️" : e.includes("zip") || e.includes("rar") || e.includes("7z") ? "📦" : e.startsWith("video/") ? "🎥" : e.startsWith("audio/") ? "🎵" : ["js", "ts", "jsx", "tsx", "html", "css", "json", "xml"].includes(i || "") ? "💻" : e.startsWith("text/") || i === "txt" ? "📄" : "📎";
  }
  static formatFileSize(t) {
    if (t === 0) return "0 Bytes";
    const e = 1024, i = ["Bytes", "KB", "MB", "GB", "TB"], s = Math.floor(Math.log(t) / Math.log(e));
    return parseFloat((t / Math.pow(e, s)).toFixed(2)) + " " + i[s];
  }
  static generateFileId() {
    return `file_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
  static async createThumbnail(t, e = 150) {
    if (!t.type.startsWith("image/"))
      throw new Error("File is not an image");
    return new Promise((i, s) => {
      const a = document.createElement("canvas"), p = a.getContext("2d"), o = new Image();
      o.onload = () => {
        a.width = e, a.height = e;
        const n = Math.min(o.width, o.height), d = (o.width - n) / 2, c = (o.height - n) / 2;
        p?.drawImage(
          o,
          d,
          c,
          n,
          n,
          0,
          0,
          e,
          e
        ), i(a.toDataURL(t.type));
      }, o.onerror = s, o.src = URL.createObjectURL(t);
    });
  }
}
const g = {
  IMAGES: ["image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"],
  DOCUMENTS: [
    "application/pdf",
    "application/msword",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "application/vnd.ms-excel",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    "application/vnd.ms-powerpoint",
    "application/vnd.openxmlformats-officedocument.presentationml.presentation"
  ],
  ARCHIVES: ["application/zip", "application/x-rar-compressed", "application/x-7z-compressed"],
  VIDEOS: ["video/mp4", "video/webm", "video/ogg", "video/quicktime"],
  AUDIOS: ["audio/mpeg", "audio/wav", "audio/ogg", "audio/m4a"],
  TEXT: ["text/plain", "text/csv", "application/json", "text/html", "text/css", "application/javascript"]
}, f = {
  SMALL: 1024 * 1024,
  // 1MB
  MEDIUM: 5 * 1024 * 1024,
  // 5MB
  LARGE: 10 * 1024 * 1024,
  // 10MB
  XLARGE: 50 * 1024 * 1024,
  // 50MB
  XXLARGE: 100 * 1024 * 1024
  // 100MB
};
export {
  w as FileProcessor,
  f as FileSizeLimits,
  g as FileTypes,
  x as FileValidator
};
//# sourceMappingURL=index.js.map
