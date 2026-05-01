class u {
  fields = /* @__PURE__ */ new Map();
  addField(t, s) {
    this.fields.set(t, s);
  }
  removeField(t) {
    this.fields.delete(t);
  }
  async validateField(t, s) {
    const r = this.fields.get(t);
    if (!r)
      return { isValid: !0, errors: [] };
    const i = [];
    if (r.required && this.isEmpty(s) && i.push("This field is required"), this.isEmpty(s) && !r.required)
      return { isValid: !0, errors: [] };
    for (const a of r.rules)
      try {
        if (!await a.validate(s)) {
          const l = typeof a.message == "function" ? a.message(s) : a.message;
          i.push(l);
        }
      } catch {
        i.push("Validation error occurred");
      }
    return {
      isValid: i.length === 0,
      errors: i
    };
  }
  async validateForm(t) {
    const s = {};
    for (const [r, i] of this.fields) {
      const a = t[r];
      s[r] = await this.validateField(r, a);
    }
    return s;
  }
  isEmpty(t) {
    return t == null || t === "" || Array.isArray(t) && t.length === 0;
  }
}
const f = {
  required: (e = "This field is required") => ({
    validate: (t) => !o(t),
    message: e
  }),
  email: (e = "Please enter a valid email address") => ({
    validate: (t) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(t),
    message: e
  }),
  minLength: (e, t) => ({
    validate: (s) => s.length >= e,
    message: t || `Must be at least ${e} characters long`
  }),
  maxLength: (e, t) => ({
    validate: (s) => s.length <= e,
    message: t || `Must be no more than ${e} characters long`
  }),
  pattern: (e, t = "Invalid format") => ({
    validate: (s) => e.test(s),
    message: t
  }),
  number: (e = "Must be a valid number") => ({
    validate: (t) => !isNaN(Number(t)),
    message: e
  }),
  min: (e, t) => ({
    validate: (s) => Number(s) >= e,
    message: t || `Must be at least ${e}`
  }),
  max: (e, t) => ({
    validate: (s) => Number(s) <= e,
    message: t || `Must be no more than ${e}`
  }),
  phone: (e = "Please enter a valid phone number") => ({
    validate: (t) => /^[\+]?[1-9][\d]{0,15}$/.test(t.replace(/[\s\-\(\)]/g, "")),
    message: e
  }),
  url: (e = "Please enter a valid URL") => ({
    validate: (t) => {
      try {
        return new URL(t), !0;
      } catch {
        return !1;
      }
    },
    message: e
  }),
  password: (e = {}, t) => ({
    validate: (s) => {
      const {
        minLength: r = 8,
        requireUppercase: i = !0,
        requireLowercase: a = !0,
        requireNumbers: n = !0,
        requireSymbols: l = !1
      } = e;
      return !(s.length < r || i && !/[A-Z]/.test(s) || a && !/[a-z]/.test(s) || n && !/\d/.test(s) || l && !/[^A-Za-z0-9]/.test(s));
    },
    message: t || "Password does not meet requirements"
  }),
  match: (e, t, s) => ({
    validate: (r) => r === t(),
    message: s || `Must match ${e}`
  }),
  fileSize: (e, t) => ({
    validate: (s) => s.size <= e,
    message: t || `File size must be less than ${d(e)}`
  }),
  fileType: (e, t) => ({
    validate: (s) => e.includes(s.type),
    message: t || `File type must be one of: ${e.join(", ")}`
  }),
  async: (e, t = "Validation failed") => ({
    validate: e,
    message: t
  })
};
function o(e) {
  return e == null || e === "" || Array.isArray(e) && e.length === 0;
}
function d(e, t = 2) {
  if (e === 0) return "0 Bytes";
  const s = 1024, r = t < 0 ? 0 : t, i = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"], a = Math.floor(Math.log(e) / Math.log(s));
  return parseFloat((e / Math.pow(s, a)).toFixed(r)) + " " + i[a];
}
export {
  u as FormValidator,
  f as ValidationRules,
  d as formatBytes,
  o as isEmpty
};
//# sourceMappingURL=index.js.map
