const f = (t, r, n) => Math.min(Math.max(t, r), n), s = (t, r = 0) => {
  const n = Math.pow(10, r);
  return Math.round(t * n) / n;
}, i = (t, r) => Math.floor(Math.random() * (r - t + 1)) + t, h = (t, r, n = 2) => {
  const e = Math.random() * (r - t) + t;
  return s(e, n);
}, l = (t, r, n) => t + (r - t) * n, d = (t, r, n) => (t - r) / (n - r), m = (t, r, n, e, o) => (t - r) * (o - e) / (n - r) + e, M = (t, r, n = 2) => r === 0 ? 0 : s(t / r * 100, n), g = (t, r = "USD", n = "en-US", e = {}) => new Intl.NumberFormat(n, {
  style: "currency",
  currency: r,
  ...e
}).format(t), p = (t, r = "en-US", n = {}) => new Intl.NumberFormat(r, n).format(t), B = (t, r = 2) => {
  if (t === 0) return "0 Bytes";
  const n = 1024, e = r < 0 ? 0 : r, o = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"], c = Math.floor(Math.log(t) / Math.log(n));
  return parseFloat((t / Math.pow(n, c)).toFixed(e)) + " " + o[c];
}, N = (t, r = "en-US") => new Intl.NumberFormat(r, {
  notation: "compact",
  maximumFractionDigits: 1
}).format(t), w = (t) => {
  if (typeof t == "number") return t;
  const r = t.replace(/[^\d.-]/g, ""), n = parseFloat(r);
  return isNaN(n) ? 0 : n;
}, F = (t) => t % 2 === 0, q = (t) => t % 2 !== 0, y = (t) => {
  if (t < 2) return !1;
  if (t === 2) return !0;
  if (t % 2 === 0) return !1;
  for (let r = 3; r <= Math.sqrt(t); r += 2)
    if (t % r === 0) return !1;
  return !0;
}, D = (t) => {
  if (t <= 1) return t;
  let r = 0, n = 1;
  for (let e = 2; e <= t; e++)
    [r, n] = [n, r + n];
  return n;
}, I = (t) => {
  if (t < 0) return NaN;
  if (t === 0 || t === 1) return 1;
  let r = 1;
  for (let n = 2; n <= t; n++)
    r *= n;
  return r;
}, u = (t, r) => {
  for (t = Math.abs(t), r = Math.abs(r); r !== 0; )
    [t, r] = [r, t % r];
  return t;
}, x = (t, r) => Math.abs(t * r) / u(t, r), a = (t) => t.length === 0 ? 0 : t.reduce((r, n) => r + n, 0) / t.length, S = (t) => {
  if (t.length === 0) return 0;
  const r = [...t].sort((e, o) => e - o), n = Math.floor(r.length / 2);
  return r.length % 2 === 0 ? (r[n - 1] + r[n]) / 2 : r[n];
}, P = (t) => {
  if (t.length === 0) return [];
  const r = {};
  let n = 0;
  return t.forEach((e) => {
    r[e] = (r[e] || 0) + 1, n = Math.max(n, r[e]);
  }), Object.keys(r).filter((e) => r[Number(e)] === n).map(Number);
}, U = (t) => {
  if (t.length === 0) return 0;
  const r = a(t), n = t.map((o) => Math.pow(o - r, 2)), e = a(n);
  return Math.sqrt(e);
}, E = (t) => {
  if (t.length === 0) return 0;
  const r = a(t), n = t.map((e) => Math.pow(e - r, 2));
  return a(n);
}, k = (t) => t.reduce((r, n) => r + n, 0), z = (t) => t.reduce((r, n) => r * n, 1), C = (t, r, n = 1) => {
  const e = [];
  if (n === 0) return e;
  if (n > 0)
    for (let o = t; o <= r; o += n)
      e.push(o);
  else
    for (let o = t; o >= r; o += n)
      e.push(o);
  return e;
}, O = (t, r, n, e = !0) => e ? t >= r && t <= n : t > r && t < n, R = (t) => t * (Math.PI / 180), j = (t) => t * (180 / Math.PI), G = (t, r, n, e) => Math.sqrt(Math.pow(n - t, 2) + Math.pow(e - r, 2));
export {
  a as average,
  f as clamp,
  G as distance,
  I as factorial,
  D as fibonacci,
  B as formatBytes,
  N as formatCompact,
  g as formatCurrency,
  p as formatNumber,
  u as gcd,
  O as inRange,
  F as isEven,
  q as isOdd,
  y as isPrime,
  x as lcm,
  l as lerp,
  S as median,
  P as mode,
  d as normalize,
  w as parseNumber,
  M as percentage,
  z as product,
  h as randomFloat,
  i as randomInt,
  C as range,
  s as round,
  m as scale,
  U as standardDeviation,
  k as sum,
  j as toDegrees,
  R as toRadians,
  E as variance
};
//# sourceMappingURL=index.js.map
