const s = (e) => e instanceof Date ? e : new Date(e), W = (e) => {
  const t = s(e);
  return !isNaN(t.getTime());
}, k = (e, t = "en-US", n = {}) => {
  const r = s(e);
  return new Intl.DateTimeFormat(t, n).format(r);
}, H = (e, t = "en-US") => {
  const n = s(e), r = /* @__PURE__ */ new Date(), a = new Intl.RelativeTimeFormat(t, { numeric: "auto" }), o = (n.getTime() - r.getTime()) / 1e3, D = o / 60, u = D / 60, c = u / 24, g = c / 7, l = c / 30.44, h = c / 365.25;
  return Math.abs(o) < 60 ? a.format(Math.round(o), "second") : Math.abs(D) < 60 ? a.format(Math.round(D), "minute") : Math.abs(u) < 24 ? a.format(Math.round(u), "hour") : Math.abs(c) < 7 ? a.format(Math.round(c), "day") : Math.abs(g) < 4 ? a.format(Math.round(g), "week") : Math.abs(l) < 12 ? a.format(Math.round(l), "month") : a.format(Math.round(h), "year");
}, Y = (e, t) => {
  const n = new Date(s(e));
  return n.setDate(n.getDate() + t), n;
}, m = (e, t) => Y(e, t * 7), w = (e, t) => {
  const n = new Date(s(e));
  return n.setMonth(n.getMonth() + t), n;
}, y = (e, t) => {
  const n = new Date(s(e));
  return n.setFullYear(n.getFullYear() + t), n;
}, E = (e, t) => Y(e, -t), U = (e, t) => m(e, -t), p = (e, t) => w(e, -t), A = (e, t) => y(e, -t), M = (e) => {
  const t = new Date(s(e));
  return t.setHours(0, 0, 0, 0), t;
}, d = (e) => {
  const t = new Date(s(e));
  return t.setHours(23, 59, 59, 999), t;
}, f = (e, t = 0) => {
  const n = new Date(s(e)), r = n.getDay(), a = (r < t ? 7 : 0) + r - t;
  return n.setDate(n.getDate() - a), M(n);
}, L = (e, t = 0) => {
  const n = f(e, t);
  return n.setDate(n.getDate() + 6), d(n);
}, N = (e) => {
  const t = new Date(s(e));
  return t.setDate(1), M(t);
}, R = (e) => {
  const t = new Date(s(e));
  return t.setMonth(t.getMonth() + 1, 0), d(t);
}, _ = (e) => {
  const t = new Date(s(e));
  return t.setMonth(0, 1), M(t);
}, z = (e) => {
  const t = new Date(s(e));
  return t.setMonth(11, 31), d(t);
}, i = (e, t) => {
  const n = s(e), r = s(t);
  return n.getFullYear() === r.getFullYear() && n.getMonth() === r.getMonth() && n.getDate() === r.getDate();
}, C = (e, t, n = 0) => {
  const r = f(e, n), a = f(t, n);
  return r.getTime() === a.getTime();
}, v = (e, t) => {
  const n = s(e), r = s(t);
  return n.getFullYear() === r.getFullYear() && n.getMonth() === r.getMonth();
}, B = (e, t) => {
  const n = s(e), r = s(t);
  return n.getFullYear() === r.getFullYear();
}, x = (e) => i(e, /* @__PURE__ */ new Date()), G = (e) => {
  const t = /* @__PURE__ */ new Date();
  return t.setDate(t.getDate() - 1), i(e, t);
}, P = (e) => {
  const t = /* @__PURE__ */ new Date();
  return t.setDate(t.getDate() + 1), i(e, t);
}, T = (e) => {
  const t = s(e).getDay();
  return t === 0 || t === 6;
}, O = (e) => !T(e), I = (e) => e % 4 === 0 && e % 100 !== 0 || e % 400 === 0, F = (e) => {
  const t = s(e);
  return new Date(t.getFullYear(), t.getMonth() + 1, 0).getDate();
}, V = (e) => I(e) ? 366 : 365, Z = (e) => {
  const t = new Date(s(e)), n = new Date(t.getFullYear(), 0, 1), r = Math.floor((t.getTime() - n.getTime()) / (1440 * 60 * 1e3));
  return Math.ceil((r + n.getDay() + 1) / 7);
}, j = (e, t = /* @__PURE__ */ new Date()) => {
  const n = s(e), r = s(t);
  let a = r.getFullYear() - n.getFullYear();
  const o = r.getMonth() - n.getMonth();
  return (o < 0 || o === 0 && r.getDate() < n.getDate()) && a--, a;
}, b = (e, t) => {
  const n = s(e), r = s(t), a = [], o = new Date(n);
  for (; o <= r; )
    a.push(new Date(o)), o.setDate(o.getDate() + 1);
  return a;
}, S = (e, t) => b(e, t).filter(O), q = (e, t) => S(e, t).length, J = (e, t, n, r) => {
  const a = new Date(e, t, 1), o = a.getDay(), u = 1 + (n - o + 7) % 7 + (r - 1) * 7, c = F(a);
  return u > c ? null : new Date(e, t, u);
}, K = (e = /* @__PURE__ */ new Date()) => s(e).getTimezoneOffset(), Q = (e) => {
  const t = s(e);
  return new Date(t.getTime() + t.getTimezoneOffset() * 6e4);
}, X = (e) => {
  const t = s(e);
  return new Date(t.getTime() - t.getTimezoneOffset() * 6e4);
}, $ = {
  ISO: "YYYY-MM-DD",
  US: "MM/DD/YYYY",
  EU: "DD/MM/YYYY",
  SHORT: "MMM DD, YYYY",
  LONG: "MMMM DD, YYYY",
  FULL: "dddd, MMMM DD, YYYY",
  TIME_12: "h:mm A",
  TIME_24: "HH:mm",
  DATETIME_12: "MMM DD, YYYY h:mm A",
  DATETIME_24: "MMM DD, YYYY HH:mm"
};
export {
  Y as addDays,
  w as addMonths,
  m as addWeeks,
  y as addYears,
  d as endOfDay,
  R as endOfMonth,
  L as endOfWeek,
  z as endOfYear,
  k as formatDate,
  $ as formatPatterns,
  H as formatRelative,
  X as fromUTC,
  j as getAge,
  S as getBusinessDays,
  q as getBusinessDaysCount,
  b as getDateRange,
  F as getDaysInMonth,
  V as getDaysInYear,
  J as getNthWeekdayOfMonth,
  K as getTimeZoneOffset,
  Z as getWeekOfYear,
  I as isLeapYear,
  i as isSameDay,
  v as isSameMonth,
  C as isSameWeek,
  B as isSameYear,
  x as isToday,
  P as isTomorrow,
  W as isValidDate,
  O as isWeekday,
  T as isWeekend,
  G as isYesterday,
  s as parseDate,
  M as startOfDay,
  N as startOfMonth,
  f as startOfWeek,
  _ as startOfYear,
  E as subtractDays,
  p as subtractMonths,
  U as subtractWeeks,
  A as subtractYears,
  Q as toUTC
};
//# sourceMappingURL=index.js.map
