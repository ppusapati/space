const i = (e) => e && e.charAt(0).toUpperCase() + e.slice(1).toLowerCase(), u = (e) => e.replace(
  /(?:^\w|[A-Z]|\b\w)/g,
  (t, n) => n === 0 ? t.toLowerCase() : t.toUpperCase()
).replace(/\s+/g, ""), p = (e) => e.replace(/([a-z0-9])([A-Z])/g, "$1-$2").replace(/[\s_]+/g, "-").toLowerCase(), h = (e) => e.replace(/([a-z0-9])([A-Z])/g, "$1_$2").replace(/[\s-]+/g, "_").toLowerCase(), m = (e) => e.replace(
  /(?:^\w|[A-Z]|\b\w|\s+)/g,
  (t, n) => +t == 0 ? "" : t.toUpperCase()
), f = (e, t, n = "...") => e.length <= t ? e : e.substring(0, t - n.length) + n, w = (e) => e.toLowerCase().trim().replace(/[^\w\s-]/g, "").replace(/[\s_-]+/g, "-").replace(/^-+|-+$/g, ""), C = (e) => e.normalize("NFD").replace(/[\u0300-\u036f]/g, ""), x = (e, t = "*", n = 0, r = 0) => {
  if (e.length <= n + r)
    return e;
  const l = e.substring(0, n), c = r > 0 ? e.substring(e.length - r) : "", a = t.repeat(e.length - n - r);
  return l + a + c;
}, d = (e, t, n = "end") => {
  if (e.length <= t) return e;
  const r = "...", l = r.length;
  switch (n) {
    case "start":
      return r + e.slice(e.length - t + l);
    case "middle":
      const c = Math.floor((t - l) / 2);
      return e.slice(0, c) + r + e.slice(e.length - c);
    default:
      return e.slice(0, t - l) + r;
  }
}, A = (e) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e), $ = (e) => {
  try {
    return new URL(e), !0;
  } catch {
    return !1;
  }
}, z = (e) => {
  const t = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi;
  return e.match(t) || [];
}, R = (e) => {
  const t = /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b/g;
  return e.match(t) || [];
}, U = (e) => e.replace(/<[^>]*>/g, ""), Z = (e) => {
  const t = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#x27;",
    "/": "&#x2F;"
  };
  return e.replace(/[&<>"'\/]/g, (n) => t[n] ?? n);
}, L = (e) => {
  const t = {
    "&amp;": "&",
    "&lt;": "<",
    "&gt;": ">",
    "&quot;": '"',
    "&#x27;": "'",
    "&#x2F;": "/"
  };
  return e.replace(/&(?:amp|lt|gt|quot|#x27|#x2F);/g, (n) => t[n] ?? n);
}, M = (e) => e.trim().split(/\s+/).filter((t) => t.length > 0).length, b = (e, t = !0) => t ? e.length : e.replace(/\s/g, "").length, o = (e) => e.split("").reverse().join(""), k = (e) => {
  const t = e.toLowerCase().replace(/[^a-z0-9]/gi, "");
  return t === o(t);
}, y = (e, t = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") => {
  let n = "";
  for (let r = 0; r < e; r++)
    n += t.charAt(Math.floor(Math.random() * t.length));
  return n;
}, g = (e, t) => {
  const n = Array(t.length + 1).fill(null).map(() => Array(e.length + 1).fill(0));
  for (let r = 0; r <= e.length; r++)
    n[0][r] = r;
  for (let r = 0; r <= t.length; r++)
    n[r][0] = r;
  for (let r = 1; r <= t.length; r++)
    for (let l = 1; l <= e.length; l++) {
      const c = e[l - 1] === t[r - 1] ? 0 : 1, a = n[r], s = n[r - 1];
      a[l] = Math.min(
        a[l - 1] + 1,
        // deletion
        s[l] + 1,
        // insertion
        s[l - 1] + c
        // substitution
      );
    }
  return n[t.length][e.length];
}, F = (e, t) => {
  const n = Math.max(e.length, t.length);
  if (n === 0) return 1;
  const r = g(e, t);
  return (n - r) / n;
}, _ = (e, t, n = "highlight") => {
  if (!t.trim()) return e;
  const r = new RegExp(`(${t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`, "gi");
  return e.replace(r, `<span class="${n}">$1</span>`);
};
export {
  u as camelCase,
  i as capitalize,
  b as characterCount,
  d as ellipsis,
  Z as escapeHtml,
  R as extractEmails,
  z as extractUrls,
  _ as highlight,
  A as isEmail,
  k as isPalindrome,
  $ as isUrl,
  p as kebabCase,
  g as levenshteinDistance,
  x as mask,
  m as pascalCase,
  y as randomString,
  C as removeAccents,
  U as removeHtml,
  o as reverse,
  F as similarity,
  w as slugify,
  h as snakeCase,
  f as truncate,
  L as unescapeHtml,
  M as wordCount
};
//# sourceMappingURL=index.js.map
