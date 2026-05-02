function p(r, e) {
  return {
    validate: (t) => r.safeParse(t).success,
    message: e || "Validation failed"
  };
}
function g(r, e) {
  return {
    validate: async (t) => (await r.safeParseAsync(t)).success,
    message: e || "Validation failed"
  };
}
function l(r, e) {
  const t = r.safeParse(e);
  if (t.success)
    return {
      isValid: !0,
      errors: [],
      fieldErrors: {},
      data: t.data
    };
  const s = [], i = {};
  for (const a of t.error.errors) {
    const n = a.path.join(".");
    s.push(a.message), n && (i[n] || (i[n] = []), i[n].push(a.message));
  }
  return {
    isValid: !1,
    errors: s,
    fieldErrors: i
  };
}
async function c(r, e) {
  const t = await r.safeParseAsync(e);
  if (t.success)
    return {
      isValid: !0,
      errors: [],
      fieldErrors: {},
      data: t.data
    };
  const s = [], i = {};
  for (const a of t.error.errors) {
    const n = a.path.join(".");
    s.push(a.message), n && (i[n] || (i[n] = []), i[n].push(a.message));
  }
  return {
    isValid: !1,
    errors: s,
    fieldErrors: i
  };
}
function y(r, e) {
  return {
    validate: (t) => {
      try {
        return r.validateSync(t), !0;
      } catch {
        return !1;
      }
    },
    message: e || "Validation failed"
  };
}
function v(r, e) {
  return {
    validate: async (t) => r.isValid(t),
    message: e || "Validation failed"
  };
}
function d(r, e) {
  try {
    const t = r.validateSync(e, { abortEarly: !1 });
    return {
      isValid: !0,
      errors: [],
      fieldErrors: {},
      data: t
    };
  } catch (t) {
    const s = t, i = s.errors || [], a = {};
    if (s.inner)
      for (const n of s.inner)
        n.path && (a[n.path] || (a[n.path] = []), a[n.path].push(...n.errors));
    return {
      isValid: !1,
      errors: i,
      fieldErrors: a
    };
  }
}
async function f(r, e) {
  try {
    const t = await r.validate(e, { abortEarly: !1 });
    return {
      isValid: !0,
      errors: [],
      fieldErrors: {},
      data: t
    };
  } catch (t) {
    const s = t, i = s.errors || [], a = {};
    if (s.inner)
      for (const n of s.inner)
        n.path && (a[n.path] || (a[n.path] = []), a[n.path].push(...n.errors));
    return {
      isValid: !1,
      errors: i,
      fieldErrors: a
    };
  }
}
class u {
  schema;
  type;
  constructor(e, t) {
    this.schema = e, this.type = t;
  }
  /**
   * Create validator from Zod schema
   */
  static fromZod(e) {
    return new u(e, "zod");
  }
  /**
   * Create validator from Yup schema
   */
  static fromYup(e) {
    return new u(e, "yup");
  }
  /**
   * Validate data synchronously
   */
  validate(e) {
    return this.type === "zod" ? l(this.schema, e) : d(this.schema, e);
  }
  /**
   * Validate data asynchronously
   */
  async validateAsync(e) {
    return this.type === "zod" ? c(this.schema, e) : f(this.schema, e);
  }
  /**
   * Get a ValidationRule for a specific field path
   */
  getFieldRule(e, t) {
    return {
      validate: (s) => !this.validate({ [e]: s }).fieldErrors[e]?.length,
      message: t || `${e} is invalid`
    };
  }
  /**
   * Convert schema to ValidationRule array for use with FormValidator
   */
  toValidationRules() {
    return [{
      validate: (e) => this.validate(e).isValid,
      message: "Form validation failed"
    }];
  }
}
function E(r) {
  return {
    validate: (e) => l(r, e),
    validateAsync: (e) => c(r, e),
    validateField: (e, t, s) => {
      const i = s ? { ...s, [e]: t } : { [e]: t }, a = l(r, i);
      return {
        isValid: !a.fieldErrors[e]?.length,
        errors: a.fieldErrors[e] || []
      };
    }
  };
}
function V(r) {
  return {
    validate: (e) => d(r, e),
    validateAsync: (e) => f(r, e),
    validateField: (e, t, s) => {
      const i = s ? { ...s, [e]: t } : { [e]: t }, a = d(r, i);
      return {
        isValid: !a.fieldErrors[e]?.length,
        errors: a.fieldErrors[e] || []
      };
    }
  };
}
function F(r, e) {
  return r.fieldErrors[e] || [];
}
function b(r, e) {
  return (r.fieldErrors[e]?.length || 0) > 0;
}
function A(r, e) {
  return r.fieldErrors[e]?.[0];
}
function R(r) {
  const e = {};
  for (const [t, s] of Object.entries(r.fieldErrors))
    s.length > 0 && (e[t] = s[0]);
  return e;
}
class w {
  fields = /* @__PURE__ */ new Map();
  addField(e, t) {
    this.fields.set(e, t);
  }
  removeField(e) {
    this.fields.delete(e);
  }
  async validateField(e, t) {
    const s = this.fields.get(e);
    if (!s)
      return { isValid: !0, errors: [] };
    const i = [];
    if (s.required && this.isEmpty(t) && i.push("This field is required"), this.isEmpty(t) && !s.required)
      return { isValid: !0, errors: [] };
    for (const a of s.rules)
      try {
        if (!await a.validate(t)) {
          const o = typeof a.message == "function" ? a.message(t) : a.message;
          i.push(o);
        }
      } catch {
        i.push("Validation error occurred");
      }
    return {
      isValid: i.length === 0,
      errors: i
    };
  }
  async validateForm(e) {
    const t = {};
    for (const [s, i] of this.fields) {
      const a = e[s];
      t[s] = await this.validateField(s, a);
    }
    return t;
  }
  isEmpty(e) {
    return e == null || e === "" || Array.isArray(e) && e.length === 0;
  }
}
const M = {
  required: (r = "This field is required") => ({
    validate: (e) => !h(e),
    message: r
  }),
  email: (r = "Please enter a valid email address") => ({
    validate: (e) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e),
    message: r
  }),
  minLength: (r, e) => ({
    validate: (t) => t.length >= r,
    message: e || `Must be at least ${r} characters long`
  }),
  maxLength: (r, e) => ({
    validate: (t) => t.length <= r,
    message: e || `Must be no more than ${r} characters long`
  }),
  pattern: (r, e = "Invalid format") => ({
    validate: (t) => r.test(t),
    message: e
  }),
  number: (r = "Must be a valid number") => ({
    validate: (e) => !isNaN(Number(e)),
    message: r
  }),
  min: (r, e) => ({
    validate: (t) => Number(t) >= r,
    message: e || `Must be at least ${r}`
  }),
  max: (r, e) => ({
    validate: (t) => Number(t) <= r,
    message: e || `Must be no more than ${r}`
  }),
  phone: (r = "Please enter a valid phone number") => ({
    validate: (e) => /^[\+]?[1-9][\d]{0,15}$/.test(e.replace(/[\s\-\(\)]/g, "")),
    message: r
  }),
  url: (r = "Please enter a valid URL") => ({
    validate: (e) => {
      try {
        return new URL(e), !0;
      } catch {
        return !1;
      }
    },
    message: r
  }),
  password: (r = {}, e) => ({
    validate: (t) => {
      const {
        minLength: s = 8,
        requireUppercase: i = !0,
        requireLowercase: a = !0,
        requireNumbers: n = !0,
        requireSymbols: o = !1
      } = r;
      return !(t.length < s || i && !/[A-Z]/.test(t) || a && !/[a-z]/.test(t) || n && !/\d/.test(t) || o && !/[^A-Za-z0-9]/.test(t));
    },
    message: e || "Password does not meet requirements"
  }),
  match: (r, e, t) => ({
    validate: (s) => s === e(),
    message: t || `Must match ${r}`
  }),
  fileSize: (r, e) => ({
    validate: (t) => t.size <= r,
    message: e || `File size must be less than ${m(r)}`
  }),
  fileType: (r, e) => ({
    validate: (t) => r.includes(t.type),
    message: e || `File type must be one of: ${r.join(", ")}`
  }),
  async: (r, e = "Validation failed") => ({
    validate: r,
    message: e
  })
};
function h(r) {
  return r == null || r === "" || Array.isArray(r) && r.length === 0;
}
function m(r, e = 2) {
  if (r === 0) return "0 Bytes";
  const t = 1024, s = e < 0 ? 0 : e, i = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"], a = Math.floor(Math.log(r) / Math.log(t));
  return parseFloat((r / Math.pow(t, a)).toFixed(s)) + " " + i[a];
}
export {
  w as FormValidator,
  u as SchemaValidator,
  M as ValidationRules,
  V as createYupFormValidator,
  E as createZodFormValidator,
  m as formatBytes,
  F as getFieldErrors,
  A as getFirstFieldError,
  b as hasFieldError,
  h as isEmpty,
  R as mapToFormErrors,
  d as validateWithYup,
  f as validateWithYupAsync,
  l as validateWithZod,
  c as validateWithZodAsync,
  v as yupToAsyncRule,
  y as yupToRule,
  g as zodToAsyncRule,
  p as zodToRule
};
//# sourceMappingURL=index.js.map
