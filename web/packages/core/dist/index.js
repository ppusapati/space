import { writable as M, derived as H, get as u } from "svelte/store";
const pe = {
  xs: 0,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  "2xl": 1536
};
function ye(R) {
  const { config: h, initialItems: w = [], initialFilters: $, initialSort: N, initialPageSize: K = 10, fetchData: p, onError: m } = R, I = M(w), A = M({
    isLoading: !1,
    isRefreshing: !1,
    isEmpty: w.length === 0,
    hasError: !1
  }), E = M($ ?? {}), S = M([]), x = M(""), f = M(
    N ?? h.defaultSort ?? { column: null, direction: null }
  ), g = M({
    page: 1,
    pageSize: K,
    total: w.length,
    totalPages: Math.ceil(w.length / K),
    hasNext: w.length > K,
    hasPrevious: !1
  }), v = M({
    selectedItems: [],
    selectedKeys: /* @__PURE__ */ new Set(),
    isAllSelected: !1,
    isIndeterminate: !1
  }), C = H(
    [I, E, S, x, f],
    ([e, r, c, L, J]) => {
      let D = [...e];
      if (L && h.searchable && h.searchFields) {
        const B = L.toLowerCase();
        D = D.filter(
          (W) => h.searchFields.some((U) => {
            const T = W[U];
            return T != null && String(T).toLowerCase().includes(B);
          })
        );
      }
      for (const B of c)
        D = D.filter((W) => {
          const U = W[B.field];
          return j(U, B.operator, B.value, B.secondValue);
        });
      return J.column && J.direction && D.sort((B, W) => {
        const U = B[J.column], T = W[J.column];
        let ie = 0;
        return U == null && T == null ? ie = 0 : U == null ? ie = 1 : T == null ? ie = -1 : typeof U == "string" && typeof T == "string" ? ie = U.localeCompare(T) : typeof U == "number" && typeof T == "number" ? ie = U - T : U instanceof Date && T instanceof Date ? ie = U.getTime() - T.getTime() : ie = String(U).localeCompare(String(T)), J.direction === "desc" ? -ie : ie;
      }), D;
    }
  ), z = H([C, g], ([e, r]) => {
    if (!h.paginated) return e;
    const c = (r.page - 1) * r.pageSize, L = c + r.pageSize;
    return e.slice(c, L);
  });
  function F(e) {
    return typeof h.itemKey == "function" ? h.itemKey(e) : e[h.itemKey];
  }
  function q(e) {
    g.update((r) => {
      const c = Math.ceil(e / r.pageSize);
      return {
        ...r,
        total: e,
        totalPages: c,
        hasNext: r.page < c,
        hasPrevious: r.page > 1
      };
    });
  }
  function j(e, r, c, L) {
    switch (r) {
      case "equals":
        return e === c;
      case "notEquals":
        return e !== c;
      case "contains":
        return String(e).toLowerCase().includes(String(c).toLowerCase());
      case "notContains":
        return !String(e).toLowerCase().includes(String(c).toLowerCase());
      case "startsWith":
        return String(e).toLowerCase().startsWith(String(c).toLowerCase());
      case "endsWith":
        return String(e).toLowerCase().endsWith(String(c).toLowerCase());
      case "gt":
        return Number(e) > Number(c);
      case "gte":
        return Number(e) >= Number(c);
      case "lt":
        return Number(e) < Number(c);
      case "lte":
        return Number(e) <= Number(c);
      case "between":
        return Number(e) >= Number(c) && Number(e) <= Number(L);
      case "in":
        return Array.isArray(c) && c.includes(e);
      case "notIn":
        return Array.isArray(c) && !c.includes(e);
      case "isEmpty":
        return e == null || e === "";
      case "isNotEmpty":
        return e != null && e !== "";
      case "isNull":
        return e == null;
      case "isNotNull":
        return e != null;
      default:
        return !0;
    }
  }
  async function Y() {
    if (p) {
      A.update((e) => ({ ...e, isLoading: !0, hasError: !1, error: void 0 }));
      try {
        const e = u(E), r = u(S), c = u(x), L = u(f), J = u(g), D = await p({
          filters: e,
          activeFilters: r,
          searchQuery: c,
          sort: L,
          pagination: { page: J.page, pageSize: J.pageSize }
        });
        I.set(D.items), q(D.total), A.update((B) => ({
          ...B,
          isLoading: !1,
          isEmpty: D.items.length === 0,
          lastUpdated: /* @__PURE__ */ new Date()
        }));
      } catch (e) {
        const r = {
          code: "LOAD_ERROR",
          message: e instanceof Error ? e.message : "Failed to load data",
          retryable: !0
        };
        A.update((c) => ({ ...c, isLoading: !1, hasError: !0, error: r })), m == null || m(r);
      }
    }
  }
  async function Q() {
    g.update((e) => ({ ...e, page: 1 })), await Y();
  }
  async function ne() {
    A.update((e) => ({ ...e, isRefreshing: !0 })), await Y(), A.update((e) => ({ ...e, isRefreshing: !1 }));
  }
  function re(e, r) {
    E.update((c) => ({ ...c, [e]: r })), g.update((c) => ({ ...c, page: 1 }));
  }
  function ae(e) {
    E.update((r) => ({ ...r, ...e })), g.update((r) => ({ ...r, page: 1 }));
  }
  function se(e) {
    E.update((r) => {
      const c = { ...r };
      return delete c[e], c;
    }), g.update((r) => ({ ...r, page: 1 }));
  }
  function o() {
    E.set({}), S.set([]), g.update((e) => ({ ...e, page: 1 }));
  }
  function d(e) {
    S.update((r) => {
      const c = r.findIndex((L) => L.field === e.field);
      if (c >= 0) {
        const L = [...r];
        return L[c] = e, L;
      }
      return [...r, e];
    }), g.update((r) => ({ ...r, page: 1 }));
  }
  function O(e) {
    S.update((r) => r.filter((c) => c.field !== e)), g.update((r) => ({ ...r, page: 1 }));
  }
  function P(e) {
    x.set(e), g.update((r) => ({ ...r, page: 1 }));
  }
  function G() {
    x.set(""), g.update((e) => ({ ...e, page: 1 }));
  }
  function Z(e, r) {
    f.set({ column: e, direction: r });
  }
  function X(e) {
    f.update((r) => r.column !== e ? { column: e, direction: "asc" } : r.direction === "asc" ? { column: e, direction: "desc" } : { column: null, direction: null });
  }
  function V() {
    f.set({ column: null, direction: null });
  }
  function te(e) {
    g.update((r) => ({
      ...r,
      page: Math.max(1, Math.min(e, r.totalPages)),
      hasNext: e < r.totalPages,
      hasPrevious: e > 1
    }));
  }
  function oe(e) {
    g.update((r) => {
      const c = Math.ceil(r.total / e), L = Math.min(r.page, c);
      return {
        ...r,
        pageSize: e,
        page: L,
        totalPages: c,
        hasNext: L < c,
        hasPrevious: L > 1
      };
    });
  }
  function ue() {
    g.update((e) => {
      if (!e.hasNext) return e;
      const r = e.page + 1;
      return {
        ...e,
        page: r,
        hasNext: r < e.totalPages,
        hasPrevious: !0
      };
    });
  }
  function ce() {
    g.update((e) => {
      if (!e.hasPrevious) return e;
      const r = e.page - 1;
      return {
        ...e,
        page: r,
        hasNext: !0,
        hasPrevious: r > 1
      };
    });
  }
  function le() {
    te(1);
  }
  function fe() {
    const e = u(g);
    te(e.totalPages);
  }
  function s(e) {
    const r = F(e);
    v.update((c) => {
      if (c.selectedKeys.has(r)) return c;
      const L = new Set(c.selectedKeys);
      L.add(r);
      const J = h.multiSelect ? [...c.selectedItems, e] : [e], D = u(I), B = L.size === D.length;
      return {
        selectedItems: J,
        selectedKeys: L,
        isAllSelected: B,
        isIndeterminate: !B && L.size > 0
      };
    });
  }
  function a(e) {
    const r = F(e);
    v.update((c) => {
      if (!c.selectedKeys.has(r)) return c;
      const L = new Set(c.selectedKeys);
      return L.delete(r), {
        selectedItems: c.selectedItems.filter((D) => F(D) !== r),
        selectedKeys: L,
        isAllSelected: !1,
        isIndeterminate: L.size > 0
      };
    });
  }
  function t(e) {
    const r = F(e);
    u(v).selectedKeys.has(r) ? a(e) : s(e);
  }
  function y() {
    const e = u(I);
    v.set({
      selectedItems: [...e],
      selectedKeys: new Set(e.map(F)),
      isAllSelected: !0,
      isIndeterminate: !1
    });
  }
  function b() {
    v.set({
      selectedItems: [],
      selectedKeys: /* @__PURE__ */ new Set(),
      isAllSelected: !1,
      isIndeterminate: !1
    });
  }
  function n(e, r) {
    const c = u(I), L = Math.min(e, r), J = Math.max(e, r), D = c.slice(L, J + 1);
    v.update((B) => {
      const W = new Set(B.selectedKeys);
      return D.forEach((T) => W.add(F(T))), {
        selectedItems: c.filter((T) => W.has(F(T))),
        selectedKeys: W,
        isAllSelected: W.size === c.length,
        isIndeterminate: W.size > 0 && W.size < c.length
      };
    });
  }
  function i(e) {
    const r = F(e);
    return u(v).selectedKeys.has(r);
  }
  async function l(e, r) {
    const c = r != null && r.includeSelection ? u(v).selectedItems : u(C), L = r != null && r.columns ? h.columns.filter((D) => r.columns.includes(String(D.key))) : h.columns.filter((D) => D.visible !== !1), J = c.map((D) => {
      const B = {};
      for (const W of L) {
        const U = D[W.key];
        B[W.header] = W.format ? W.format(U, D, 0) : U;
      }
      return B;
    });
    switch (e) {
      case "csv":
        k(J, L.map((D) => D.header), r);
        break;
      case "json":
        _(J, r);
        break;
      default:
        console.warn(`Export format ${e} not implemented`);
    }
  }
  function k(e, r, c) {
    const L = (c == null ? void 0 : c.delimiter) ?? ",", J = (c == null ? void 0 : c.quoteChar) ?? '"', D = (U) => {
      const T = U == null ? "" : String(U);
      return T.includes(L) || T.includes(J) || T.includes(`
`) ? `${J}${T.replace(new RegExp(J, "g"), J + J)}${J}` : T;
    }, B = [];
    (c == null ? void 0 : c.includeHeaders) !== !1 && B.push(r.map(D).join(L));
    for (const U of e)
      B.push(r.map((T) => D(U[T])).join(L));
    const W = new Blob([B.join(`
`)], { type: "text/csv;charset=utf-8;" });
    ee(W, `${(c == null ? void 0 : c.filename) ?? "export"}.csv`);
  }
  function _(e, r) {
    const c = new Blob([JSON.stringify(e, null, 2)], { type: "application/json" });
    ee(c, `${(r == null ? void 0 : r.filename) ?? "export"}.json`);
  }
  function ee(e, r) {
    const c = URL.createObjectURL(e), L = document.createElement("a");
    L.href = c, L.download = r, document.body.appendChild(L), L.click(), document.body.removeChild(L), URL.revokeObjectURL(c);
  }
  return {
    // Stores
    items: I,
    filteredItems: C,
    displayItems: z,
    state: A,
    filters: E,
    activeFilters: S,
    searchQuery: x,
    sort: f,
    pagination: g,
    selection: v,
    // Methods
    load: Y,
    reload: Q,
    refresh: ne,
    // Filter methods
    setFilter: re,
    setFilters: ae,
    clearFilter: se,
    clearAllFilters: o,
    addActiveFilter: d,
    removeActiveFilter: O,
    // Search methods
    search: P,
    clearSearch: G,
    // Sort methods
    setSort: Z,
    toggleSort: X,
    clearSort: V,
    // Pagination methods
    setPage: te,
    setPageSize: oe,
    nextPage: ue,
    prevPage: ce,
    goToFirst: le,
    goToLast: fe,
    // Selection methods
    selectItem: s,
    deselectItem: a,
    toggleItem: t,
    selectAll: y,
    deselectAll: b,
    selectRange: n,
    isSelected: i,
    // Export
    exportData: l,
    // Utility
    getItemKey: F
  };
}
function he(R) {
  var b;
  const {
    config: h,
    sections: w = [],
    actions: $ = {},
    fetchEntity: N,
    saveEntity: K,
    deleteEntity: p,
    onError: m,
    onSave: I,
    onDelete: A
  } = R;
  let E = null;
  const S = M(null), x = M(null), f = M({
    isLoading: !1,
    isSaving: !1,
    isDeleting: !1,
    isDirty: !1,
    isValid: !0,
    hasError: !1
  }), g = M("view"), v = M(w), C = M((b = w[0]) == null ? void 0 : b.id), z = M({}), F = M({}), q = H([S, x], ([n, i]) => !n || !i ? !1 : JSON.stringify(n) !== JSON.stringify(i)), j = H(F, (n) => Object.keys(n).length === 0), Y = H([S, x], ([n, i]) => {
    if (!n || !i) return {};
    const l = {};
    for (const k of Object.keys(n))
      JSON.stringify(n[k]) !== JSON.stringify(i[k]) && (l[k] = n[k]);
    return l;
  });
  q.subscribe((n) => {
    f.update((i) => ({ ...i, isDirty: n }));
  }), j.subscribe((n) => {
    f.update((i) => ({ ...i, isValid: n }));
  });
  async function Q(n) {
    if (!N) {
      console.warn("fetchEntity not provided");
      return;
    }
    E = n, f.update((i) => ({ ...i, isLoading: !0, hasError: !1, error: void 0 }));
    try {
      const i = await N(n);
      S.set(i), x.set(structuredClone(i)), F.set({}), f.update((l) => ({
        ...l,
        isLoading: !1,
        lastModified: /* @__PURE__ */ new Date()
      }));
    } catch (i) {
      const l = {
        code: "LOAD_ERROR",
        message: i instanceof Error ? i.message : "Failed to load entity",
        retryable: !0
      };
      f.update((k) => ({ ...k, isLoading: !1, hasError: !0, error: l })), m == null || m(l);
    }
  }
  async function ne() {
    E != null && await Q(E);
  }
  async function re() {
    await ne();
  }
  function ae(n) {
    g.set(n);
  }
  function se() {
    h.editable && g.set("edit");
  }
  function o() {
    if (u(q)) {
      const i = u(x);
      i && S.set(structuredClone(i));
    }
    F.set({}), g.set("view");
  }
  function d(n) {
    E = null, S.set(n ?? {}), x.set(n ?? {}), F.set({}), g.set("create");
  }
  async function O() {
    if (!K) {
      console.warn("saveEntity not provided");
      return;
    }
    const n = u(S);
    if (!(!n || !await oe())) {
      f.update((l) => ({ ...l, isSaving: !0 }));
      try {
        const l = await K(n);
        S.set(l), x.set(structuredClone(l)), u(g) === "create" && (E = l[h.entityKey]), f.update((k) => ({
          ...k,
          isSaving: !1,
          lastSaved: /* @__PURE__ */ new Date()
        })), I == null || I(l), g.set("view");
      } catch (l) {
        const k = {
          code: "SAVE_ERROR",
          message: l instanceof Error ? l.message : "Failed to save entity",
          retryable: !0
        };
        f.update((_) => ({ ..._, isSaving: !1, hasError: !0, error: k })), m == null || m(k);
      }
    }
  }
  async function P() {
    if (!p) {
      console.warn("deleteEntity not provided");
      return;
    }
    const n = u(S);
    if (n) {
      f.update((i) => ({ ...i, isDeleting: !0 }));
      try {
        await p(n), f.update((i) => ({ ...i, isDeleting: !1 })), A == null || A(n), S.set(null), x.set(null), E = null;
      } catch (i) {
        const l = {
          code: "DELETE_ERROR",
          message: i instanceof Error ? i.message : "Failed to delete entity",
          retryable: !0
        };
        f.update((k) => ({ ...k, isDeleting: !1, hasError: !0, error: l })), m == null || m(l);
      }
    }
  }
  async function G() {
    const n = u(S);
    if (!n) return null;
    const i = structuredClone(n);
    return delete i[h.entityKey], d(i), i;
  }
  function Z(n, i) {
    S.update((l) => l && { ...l, [n]: i });
  }
  function X(n) {
    const i = u(x);
    i && (S.update((l) => l && { ...l, [n]: i[n] }), F.update((l) => {
      const k = { ...l };
      return delete k[n], k;
    }));
  }
  function V() {
    const n = u(x);
    n && S.set(structuredClone(n)), F.set({});
  }
  function te() {
    return u(Y);
  }
  async function oe() {
    const n = u(S), i = u(v);
    if (!n) return !1;
    const l = {};
    for (const k of i)
      for (const _ of k.fields) {
        if (_.validation) {
          const ee = n[_.key];
          for (const e of _.validation)
            try {
              if (!await e.validate(ee, n)) {
                l[_.key] = typeof e.message == "function" ? e.message(ee, n) : e.message;
                break;
              }
            } catch {
              l[_.key] = "Validation failed";
              break;
            }
        }
        if (_.required && (typeof _.required == "function" ? _.required(n) : _.required)) {
          const e = n[_.key];
          (e == null || e === "") && (l[_.key] = `${_.label} is required`);
        }
      }
    return F.set(l), Object.keys(l).length === 0;
  }
  async function ue(n) {
    const i = u(S), l = u(v);
    if (!i) return null;
    let k = null;
    for (const ee of l)
      if (k = ee.fields.find((e) => e.key === n), k) break;
    if (!k) return null;
    const _ = i[n];
    if (k.required && (typeof k.required == "function" ? k.required(i) : k.required) && (_ == null || _ === "")) {
      const e = `${k.label} is required`;
      return ce(n, e), e;
    }
    if (k.validation)
      for (const ee of k.validation)
        try {
          if (!await ee.validate(_, i)) {
            const r = typeof ee.message == "function" ? ee.message(_, i) : ee.message;
            return ce(n, r), r;
          }
        } catch {
          const e = "Validation failed";
          return ce(n, e), e;
        }
    return ce(n, null), null;
  }
  function ce(n, i) {
    F.update((l) => {
      if (i === null) {
        const k = { ...l };
        return delete k[n], k;
      }
      return { ...l, [n]: i };
    });
  }
  function le() {
    F.set({});
  }
  async function fe(n) {
    const i = $[n];
    if (!i) {
      console.warn(`Action ${n} not found`);
      return;
    }
    const l = u(S);
    if (l) {
      if (typeof i.disabled == "function") {
        if (i.disabled(l)) return;
      } else if (i.disabled)
        return;
      await i.handler(l);
    }
  }
  async function s(n, i) {
    z.update((l) => ({
      ...l,
      [n]: { data: null, isLoading: !0, hasError: !1 }
    }));
    try {
      const l = await i();
      z.update((k) => ({
        ...k,
        [n]: { data: l, isLoading: !1, hasError: !1, lastLoaded: /* @__PURE__ */ new Date() }
      }));
    } catch (l) {
      z.update((k) => ({
        ...k,
        [n]: {
          data: null,
          isLoading: !1,
          hasError: !0,
          error: {
            code: "LOAD_ERROR",
            message: l instanceof Error ? l.message : "Failed to load related data"
          }
        }
      }));
    }
  }
  async function a(n, i) {
    await s(n, i);
  }
  function t(n) {
    C.set(n);
  }
  function y(n) {
    v.update(
      (i) => i.map(
        (l) => l.id === n ? { ...l, defaultCollapsed: !l.defaultCollapsed } : l
      )
    );
  }
  return {
    // Stores
    entity: S,
    originalEntity: x,
    state: f,
    mode: g,
    sections: v,
    activeSectionId: C,
    relatedData: z,
    validationErrors: F,
    // Derived
    isDirty: q,
    isValid: j,
    changedFields: Y,
    // Methods - Data loading
    load: Q,
    reload: ne,
    refresh: re,
    // Methods - Mode switching
    setMode: ae,
    edit: se,
    view: o,
    create: d,
    // Methods - Entity operations
    save: O,
    delete: P,
    duplicate: G,
    // Methods - Field operations
    setFieldValue: Z,
    resetField: X,
    resetAll: V,
    getChangedFields: te,
    // Methods - Validation
    validate: oe,
    validateField: ue,
    setFieldError: ce,
    clearErrors: le,
    // Methods - Actions
    executeAction: fe,
    // Methods - Related data
    loadRelatedData: s,
    refreshRelatedData: a,
    // Methods - Sections
    setActiveSection: t,
    toggleSection: y
  };
}
function be(R) {
  const {
    initialValues: h,
    validation: w,
    validateOnChange: $ = !1,
    validateOnBlur: N = !0,
    validateOnMount: K = !1,
    onSubmit: p,
    onReset: m,
    onChange: I,
    onError: A
  } = R, E = M(structuredClone(h)), S = M(structuredClone(h)), x = M({}), f = M({}), g = M({}), v = M("idle");
  let C = !1;
  const z = H(x, (s) => Object.keys(s).length === 0), F = H(g, (s) => Object.values(s).some(Boolean)), q = H(v, (s) => s === "submitting"), j = M(!1);
  function Y(s, a) {
    E.update((y) => ({ ...y, [s]: a }));
    const t = u(S);
    g.update((y) => ({
      ...y,
      [s]: JSON.stringify(a) !== JSON.stringify(t[s])
    })), $ && !C && se(s), I == null || I(u(E));
  }
  function Q(s, a) {
    x.update((t) => {
      if (a === null) {
        const y = { ...t };
        return delete y[s], y;
      }
      return { ...t, [s]: a };
    });
  }
  function ne(s, a = !0) {
    f.update((t) => ({ ...t, [s]: a }));
  }
  function re(s) {
    E.update((t) => ({ ...t, ...s }));
    const a = u(S);
    g.update((t) => {
      const y = { ...t };
      for (const b of Object.keys(s))
        y[b] = JSON.stringify(s[b]) !== JSON.stringify(a[b]);
      return y;
    }), I == null || I(u(E));
  }
  function ae(s) {
    x.update((a) => ({ ...a, ...s }));
  }
  async function se(s) {
    var y, b;
    if (!w) return null;
    C = !0;
    const a = u(E), t = a[s];
    if ((y = w.schema) != null && y[s]) {
      const n = w.schema[s], i = await o(s, t, n, a);
      if (i)
        return Q(s, i), C = !1, i;
    }
    if ((b = w.rules) != null && b[s]) {
      const n = w.rules[s];
      for (const i of n)
        try {
          if (!await i.validate(t, a)) {
            const k = typeof i.message == "function" ? i.message(t) : i.message;
            return Q(s, k), C = !1, k;
          }
        } catch {
          const l = "Validation error";
          return Q(s, l), C = !1, l;
        }
    }
    return Q(s, null), C = !1, null;
  }
  async function o(s, a, t, y) {
    if (t.required && (a == null || a === "" || Array.isArray(a) && a.length === 0))
      return typeof t.required == "string" ? t.required : `${String(s)} is required`;
    if (a == null || a === "") return null;
    if (t.min !== void 0 && typeof a == "number") {
      const b = typeof t.min == "number" ? t.min : t.min.value, n = typeof t.min == "number" ? `Must be at least ${b}` : t.min.message;
      if (a < b) return n;
    }
    if (t.max !== void 0 && typeof a == "number") {
      const b = typeof t.max == "number" ? t.max : t.max.value, n = typeof t.max == "number" ? `Must be at most ${b}` : t.max.message;
      if (a > b) return n;
    }
    if (t.minLength !== void 0 && typeof a == "string") {
      const b = typeof t.minLength == "number" ? t.minLength : t.minLength.value, n = typeof t.minLength == "number" ? `Must be at least ${b} characters` : t.minLength.message;
      if (a.length < b) return n;
    }
    if (t.maxLength !== void 0 && typeof a == "string") {
      const b = typeof t.maxLength == "number" ? t.maxLength : t.maxLength.value, n = typeof t.maxLength == "number" ? `Must be at most ${b} characters` : t.maxLength.message;
      if (a.length > b) return n;
    }
    if (t.pattern && typeof a == "string") {
      const b = t.pattern instanceof RegExp ? t.pattern : t.pattern.value, n = t.pattern instanceof RegExp ? "Invalid format" : t.pattern.message;
      if (!b.test(a)) return n;
    }
    if (t.email && typeof a == "string") {
      const b = /^[^\s@]+@[^\s@]+\.[^\s@]+$/, n = typeof t.email == "string" ? t.email : "Invalid email address";
      if (!b.test(a)) return n;
    }
    if (t.url && typeof a == "string")
      try {
        new URL(a);
      } catch {
        return typeof t.url == "string" ? t.url : "Invalid URL";
      }
    if (t.validate) {
      const b = await t.validate(a, y);
      if (typeof b == "string") return b;
      if (b === !1) return `${String(s)} is invalid`;
    }
    return null;
  }
  async function d() {
    j.set(!0);
    const s = u(E), a = {};
    if (w) {
      if (w.schema)
        for (const y of Object.keys(w.schema)) {
          const b = w.schema[y];
          if (b) {
            const n = await o(
              y,
              s[y],
              b,
              s
            );
            n && (a[y] = n);
          }
        }
      if (w.rules)
        for (const y of Object.keys(w.rules)) {
          if (a[y]) continue;
          const b = w.rules[y];
          if (b)
            for (const n of b)
              try {
                if (!await n.validate(
                  s[y],
                  s
                )) {
                  a[y] = typeof n.message == "function" ? n.message(s[y]) : n.message;
                  break;
                }
              } catch {
                a[y] = "Validation error";
                break;
              }
        }
    }
    x.set(a), j.set(!1);
    const t = Object.keys(a).length === 0;
    return t || A == null || A(a), t;
  }
  function O(s) {
    const a = u(S);
    E.update((t) => ({ ...t, [s]: a[s] })), x.update((t) => {
      const y = { ...t };
      return delete y[s], y;
    }), f.update((t) => {
      const y = { ...t };
      return delete y[s], y;
    }), g.update((t) => {
      const y = { ...t };
      return delete y[s], y;
    });
  }
  function P(s) {
    const a = s ?? u(S);
    E.set(structuredClone(a)), s && S.set(structuredClone(s)), x.set({}), f.set({}), g.set({}), v.set("idle"), m == null || m();
  }
  async function G() {
    const s = u(E), a = {};
    for (const y of Object.keys(s))
      a[y] = !0;
    if (f.set(a), !await d()) {
      v.set("error");
      return;
    }
    v.set("submitting");
    try {
      await (p == null ? void 0 : p(s)), v.set("success");
    } catch (y) {
      throw v.set("error"), y;
    }
  }
  function Z(s) {
    const a = s.target, { name: t, type: y } = a;
    let b;
    y === "checkbox" ? b = a.checked : y === "number" || y === "range" ? b = a.value === "" ? null : Number(a.value) : y === "file" ? b = a.files : b = a.value, Y(t, b);
  }
  function X(s) {
    const a = s.target, { name: t } = a;
    ne(t, !0), N && se(t);
  }
  async function V(s) {
    s == null || s.preventDefault(), await G();
  }
  function te(s) {
    s == null || s.preventDefault(), P();
  }
  function oe(s) {
    return {
      name: s,
      value: u(E)[s],
      onChange: (a) => Y(s, a),
      onBlur: () => {
        ne(s, !0), N && se(s);
      }
    };
  }
  function ue(s) {
    const a = u(x), t = u(f), y = u(g), b = u(j);
    return {
      touched: t[s] ?? !1,
      dirty: y[s] ?? !1,
      error: a[s] ?? null,
      valid: !a[s],
      validating: b
    };
  }
  function ce(s) {
    const a = u(E), t = u(x), y = u(f), b = u(g), n = t[s] ?? null, i = y[s] ?? !1;
    let l = "default";
    return i && (n ? l = "invalid" : b[s] && (l = "valid")), {
      value: a[s],
      touched: i,
      dirty: b[s] ?? !1,
      error: n,
      validationState: l
    };
  }
  function le(s, a) {
    E.update((t) => s in t ? t : { ...t, [s]: a }), S.update((t) => s in t ? t : { ...t, [s]: a });
  }
  function fe(s) {
    E.update((a) => {
      const t = { ...a };
      return delete t[s], t;
    }), x.update((a) => {
      const t = { ...a };
      return delete t[s], t;
    }), f.update((a) => {
      const t = { ...a };
      return delete t[s], t;
    }), g.update((a) => {
      const t = { ...a };
      return delete t[s], t;
    });
  }
  return K && d(), {
    // Stores
    values: E,
    initialValues: S,
    errors: x,
    touched: f,
    dirty: g,
    status: v,
    // Derived
    isValid: z,
    isDirty: F,
    isSubmitting: q,
    isValidating: { subscribe: j.subscribe },
    // Field methods
    setFieldValue: Y,
    setFieldError: Q,
    setFieldTouched: ne,
    setValues: re,
    setErrors: ae,
    // Validation methods
    validateField: se,
    validateForm: d,
    // Reset methods
    resetField: O,
    resetForm: P,
    // Submit methods
    submitForm: G,
    // Handlers
    handleChange: Z,
    handleBlur: X,
    handleSubmit: V,
    handleReset: te,
    // Field helpers
    getFieldProps: oe,
    getFieldMeta: ue,
    getFieldState: ce,
    // Dynamic fields
    registerField: le,
    unregisterField: fe
  };
}
let ge = 0;
function we(R = {}) {
  const { id: h = `modal-${++ge}`, config: w, onOpen: $, onClose: N, onSubmit: K, onCancel: p } = R, m = M(!1), I = M(null), A = M(null), E = M({
    id: h,
    title: "",
    size: "md",
    closable: !0,
    closeOnEscape: !0,
    closeOnOverlay: !0,
    preventClose: !1,
    showHeader: !0,
    showFooter: !0,
    ...w
  });
  function S(F) {
    I.set(F ?? null), A.set(null), m.set(!0), $ == null || $(F);
    const q = u(E);
    if (q.closeOnEscape) {
      const j = (Q) => {
        Q.key === "Escape" && u(m) && !q.preventClose && g();
      };
      document.addEventListener("keydown", j);
      const Y = m.subscribe((Q) => {
        Q || (document.removeEventListener("keydown", j), Y());
      });
    }
  }
  function x() {
    u(E).preventClose || (m.set(!1), N == null || N());
  }
  function f(F) {
    A.set(F), K == null || K(F), x();
  }
  function g() {
    A.set(null), p == null || p(), x();
  }
  function v() {
    u(m) ? g() : S();
  }
  function C(F) {
    E.update((q) => ({ ...q, ...F }));
  }
  function z(F) {
    I.set(F);
  }
  return {
    isOpen: m,
    data: I,
    config: E,
    result: A,
    open: S,
    close: x,
    submit: f,
    cancel: g,
    toggle: v,
    updateConfig: C,
    setData: z
  };
}
function Se() {
  const R = M(/* @__PURE__ */ new Map()), h = M([]), w = /* @__PURE__ */ new Map(), $ = H([R, h], ([f, g]) => {
    if (g.length === 0) return null;
    const v = g[g.length - 1];
    return v ? f.get(v) ?? null : null;
  }), N = H(h, (f) => f.length > 0);
  function K(f, g, v) {
    return new Promise((C) => {
      const z = {
        id: f,
        config: g,
        data: v ?? null,
        isOpen: !0,
        result: null
      };
      R.update((q) => {
        const j = new Map(q);
        return j.set(f, z), j;
      }), h.update((q) => [...q, f]), w.set(f, C);
      const F = (q) => {
        if (q.key === "Escape" && !g.preventClose) {
          const j = u(h);
          j[j.length - 1] === f && p(f, null);
        }
      };
      if (g.closeOnEscape) {
        document.addEventListener("keydown", F);
        const q = () => {
          document.removeEventListener("keydown", F);
        }, j = () => {
          u(R).has(f) || q();
        };
        R.subscribe(j);
      }
    });
  }
  function p(f, g) {
    const v = w.get(f);
    R.update((C) => {
      const z = new Map(C);
      return z.delete(f), z;
    }), h.update((C) => C.filter((z) => z !== f)), w.delete(f), v == null || v(g);
  }
  function m(f) {
    p(f, null);
  }
  function I() {
    const f = u(h);
    for (const g of f)
      p(g, null);
  }
  function A() {
    const f = u(h), g = f[f.length - 1];
    g && p(g, null);
  }
  function E(f) {
    return u(R).has(f);
  }
  function S(f) {
    return u(R).get(f);
  }
  function x(f, g) {
    R.update((v) => {
      const C = v.get(f);
      if (C) {
        const z = new Map(v);
        return z.set(f, { ...C, data: g }), z;
      }
      return v;
    });
  }
  return {
    modals: { subscribe: R.subscribe },
    activeModal: $,
    isAnyOpen: N,
    modalStack: { subscribe: h.subscribe },
    open: K,
    close: m,
    closeAll: I,
    closeTop: A,
    isOpen: E,
    getModal: S,
    updateData: x
  };
}
function ve(R) {
  return async (h) => await R.open("confirmation", {
    id: "confirmation",
    title: h.title,
    size: "sm",
    closable: !0,
    closeOnEscape: !0,
    closeOnOverlay: !1,
    showHeader: !0,
    showFooter: !0
  }, h) ?? !1;
}
function Ee(R) {
  return async (h) => {
    await R.open("alert", {
      id: "alert",
      title: h.title,
      size: "sm",
      closable: !0,
      closeOnEscape: !0,
      closeOnOverlay: !0,
      showHeader: !0,
      showFooter: !0
    }, h);
  };
}
function Oe(R) {
  return async (h) => await R.open("prompt", {
    id: "prompt",
    title: h.title,
    size: "sm",
    closable: !0,
    closeOnEscape: !0,
    closeOnOverlay: !1,
    showHeader: !0,
    showFooter: !0
  }, h);
}
function ke(R = {}) {
  const {
    initialPage: h = 1,
    initialPageSize: w = 10,
    total: $ = 0,
    pageSizeOptions: N = [10, 25, 50, 100],
    onChange: K
  } = R, p = M(h), m = M(w), I = M($), A = H(
    [I, m],
    ([o, d]) => Math.max(1, Math.ceil(o / d))
  ), E = H([p, A], ([o, d]) => o < d), S = H(p, (o) => o > 1), x = H([p, m], ([o, d]) => (o - 1) * d), f = H(
    [p, m, I],
    ([o, d, O]) => Math.min(o * d, O)
  ), g = H([p, A], ([o, d]) => {
    const O = [];
    for (let P = 1; P <= d; P++)
      O.push(P);
    return O;
  }), v = H(
    [p, m, I, A, E, S],
    ([o, d, O, P, G, Z]) => ({
      page: o,
      pageSize: d,
      total: O,
      totalPages: P,
      hasNext: G,
      hasPrevious: Z
    })
  );
  let C = !0;
  v.subscribe((o) => {
    C || K == null || K(o), C = !1;
  }), I.subscribe((o) => {
    const d = u(p), O = u(m), P = Math.max(1, Math.ceil(o / O));
    d > P && p.set(P);
  }), m.subscribe(() => {
    C || p.set(1);
  });
  function z(o) {
    const d = u(A);
    p.set(Math.max(1, Math.min(o, d)));
  }
  function F(o) {
    (N.includes(o) || N.length === 0) && m.set(o);
  }
  function q(o) {
    I.set(Math.max(0, o));
  }
  function j() {
    u(E) && p.update((o) => o + 1);
  }
  function Y() {
    u(S) && p.update((o) => o - 1);
  }
  function Q() {
    p.set(1);
  }
  function ne() {
    p.set(u(A));
  }
  function re() {
    p.set(h), m.set(w), I.set($);
  }
  function ae(o) {
    const d = u(x), O = u(m);
    return o.slice(d, d + O);
  }
  function se(o = 7) {
    const d = u(p), O = u(A);
    if (O <= o)
      return Array.from({ length: O }, (V, te) => te + 1);
    const P = Math.floor(o / 2);
    let G = Math.max(1, d - P), Z = Math.min(O, d + P);
    d <= P ? Z = o : d >= O - P && (G = O - o + 1);
    const X = [];
    G > 1 && (X.push(1), G > 2 && X.push(-1));
    for (let V = G; V <= Z; V++)
      X.push(V);
    return Z < O && (Z < O - 1 && X.push(-1), X.push(O)), X;
  }
  return {
    // State
    page: p,
    pageSize: m,
    total: I,
    // Derived
    pagination: v,
    totalPages: A,
    hasNext: E,
    hasPrevious: S,
    pageRange: g,
    startIndex: x,
    endIndex: f,
    // Methods
    setPage: z,
    setPageSize: F,
    setTotal: q,
    nextPage: j,
    prevPage: Y,
    goToFirst: Q,
    goToLast: ne,
    reset: re,
    // Utility
    getPageItems: ae,
    getVisiblePages: se
  };
}
function de(R = {}) {
  const {
    initialSelection: h = [],
    getKey: w = (o) => typeof o == "object" && o !== null && "id" in o ? o.id : JSON.stringify(o),
    multiSelect: $ = !0,
    maxSelection: N,
    onChange: K
  } = R, p = M(h), m = M(
    new Set(h.map(w))
  ), I = M([]), A = { subscribe: p.subscribe }, E = { subscribe: m.subscribe }, S = H(p, (o) => o.length), x = H(
    [p, I],
    ([o, d]) => d.length > 0 && o.length === d.length
  ), f = H(
    [p, I],
    ([o, d]) => o.length > 0 && o.length < d.length
  ), g = H(
    [p, m, x, f],
    ([o, d, O, P]) => ({
      selectedItems: o,
      selectedKeys: d,
      isAllSelected: O,
      isIndeterminate: P
    })
  );
  let v = !0;
  g.subscribe((o) => {
    v || K == null || K(o), v = !1;
  });
  function C(o) {
    const d = w(o), O = u(m);
    O.has(d) || N && O.size >= N || ($ ? (m.update((P) => {
      const G = new Set(P);
      return G.add(d), G;
    }), p.update((P) => [...P, o])) : (m.set(/* @__PURE__ */ new Set([d])), p.set([o])));
  }
  function z(o) {
    const d = w(o);
    m.update((O) => {
      const P = new Set(O);
      return P.delete(d), P;
    }), p.update((O) => O.filter((P) => w(P) !== d));
  }
  function F(o) {
    const d = w(o);
    u(m).has(d) ? z(o) : C(o);
  }
  function q(o) {
    if (!$) return;
    I.set(o);
    let d = o;
    N && o.length > N && (d = o.slice(0, N)), p.set([...d]), m.set(new Set(d.map(w)));
  }
  function j() {
    p.set([]), m.set(/* @__PURE__ */ new Set());
  }
  function Y(o, d, O) {
    if (!$) return;
    const P = Math.min(d, O), G = Math.max(d, O), Z = o.slice(P, G + 1);
    let X;
    if (N) {
      const V = u(p), te = N - V.length;
      X = [...V, ...Z.slice(0, te)];
    } else {
      const V = u(p), te = u(m);
      X = [...V];
      for (const oe of Z) {
        const ue = w(oe);
        te.has(ue) || X.push(oe);
      }
    }
    p.set(X), m.set(new Set(X.map(w)));
  }
  function Q(o) {
    let d = o;
    if (N && o.length > N && (d = o.slice(0, N)), !$ && d.length > 1) {
      const O = d[0];
      O !== void 0 && (d = [O]);
    }
    p.set([...d]), m.set(new Set(d.map(w)));
  }
  function ne(o) {
    const d = w(o);
    return u(m).has(d);
  }
  function re() {
    return new Set(u(m));
  }
  function ae() {
    return [...u(p)];
  }
  function se() {
    p.set([...h]), m.set(new Set(h.map(w)));
  }
  return {
    // State
    selection: g,
    selectedItems: A,
    selectedKeys: E,
    isAllSelected: x,
    isIndeterminate: f,
    selectedCount: S,
    // Methods
    select: C,
    deselect: z,
    toggle: F,
    selectAll: q,
    deselectAll: j,
    selectRange: Y,
    setSelection: Q,
    isSelected: ne,
    getSelectedKeys: re,
    getSelectedItems: ae,
    reset: se
  };
}
function Le(R) {
  const { getRowKey: h, multiSelect: w = !0, onChange: $ } = R;
  return de({
    getKey: h,
    multiSelect: w,
    onChange: (N) => $ == null ? void 0 : $(N.selectedItems)
  });
}
function Fe(R) {
  const { options: h, initialSelected: w = [], getValue: $, maxSelected: N, onChange: K } = R, p = de({
    initialSelection: w,
    getKey: $ ?? ((m) => typeof m == "object" && m !== null && "value" in m ? m.value : JSON.stringify(m)),
    multiSelect: !0,
    maxSelection: N,
    onChange: (m) => K == null ? void 0 : K(m.selectedItems)
  });
  return {
    ...p,
    options: h,
    toggleAll: () => {
      u(p.isAllSelected) ? p.deselectAll() : p.selectAll(h);
    }
  };
}
export {
  pe as BREAKPOINTS,
  Ee as useAlert,
  Fe as useCheckboxGroup,
  ve as useConfirmation,
  he as useDetail,
  be as useForm,
  ye as useList,
  we as useModal,
  Se as useModalManager,
  ke as usePagination,
  Oe as usePrompt,
  Le as useRowSelection,
  de as useSelection
};
