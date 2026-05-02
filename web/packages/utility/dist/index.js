import { FormValidator as R, SchemaValidator as E, ValidationRules as I, createYupFormValidator as L, createZodFormValidator as M, getFieldErrors as j, getFirstFieldError as P, hasFieldError as H, isEmpty as Y, mapToFormErrors as N, validateWithYup as B, validateWithYupAsync as U, validateWithZod as V, validateWithZodAsync as z, yupToAsyncRule as A, yupToRule as J, zodToAsyncRule as $, zodToRule as Z } from "./validation/index.js";
import { FileProcessor as X, FileSizeLimits as q, FileTypes as G, FileValidator as K } from "./file/index.js";
import { camelCase as tt, capitalize as et, characterCount as at, ellipsis as rt, escapeHtml as ot, extractEmails as it, extractUrls as nt, highlight as st, isEmail as ct, isPalindrome as lt, isUrl as dt, kebabCase as mt, levenshteinDistance as pt, mask as ft, pascalCase as ut, randomString as yt, removeAccents as ht, removeHtml as gt, reverse as bt, similarity as xt, slugify as Tt, snakeCase as wt, truncate as Ct, unescapeHtml as vt, wordCount as Dt } from "./string/index.js";
import { addDays as Ot, addMonths as St, addWeeks as kt, addYears as Wt, endOfDay as Rt, endOfMonth as Et, endOfWeek as It, endOfYear as Lt, formatDate as Mt, formatPatterns as jt, formatRelative as Pt, fromUTC as Ht, getAge as Yt, getBusinessDays as Nt, getBusinessDaysCount as Bt, getDateRange as Ut, getDaysInMonth as Vt, getDaysInYear as zt, getNthWeekdayOfMonth as At, getTimeZoneOffset as Jt, getWeekOfYear as $t, isLeapYear as Zt, isSameDay as Qt, isSameMonth as Xt, isSameWeek as qt, isSameYear as Gt, isToday as Kt, isTomorrow as _t, isValidDate as te, isWeekday as ee, isWeekend as ae, isYesterday as re, parseDate as oe, startOfDay as ie, startOfMonth as ne, startOfWeek as se, startOfYear as ce, subtractDays as le, subtractMonths as de, subtractWeeks as me, subtractYears as pe, toUTC as fe } from "./date/index.js";
import { average as ye, clamp as he, distance as ge, factorial as be, fibonacci as xe, formatCompact as Te, gcd as we, inRange as Ce, isEven as ve, isOdd as De, isPrime as Fe, lcm as Oe, lerp as Se, median as ke, mode as We, normalize as Re, parseNumber as Ee, percentage as Ie, product as Le, randomFloat as Me, randomInt as je, range as Pe, round as He, scale as Ye, standardDeviation as Ne, sum as Be, toDegrees as Ue, toRadians as Ve, variance as ze } from "./number/index.js";
import { formatAddress as Je, formatBytes as $e, formatCSV as Ze, formatCreditCard as Qe, formatCurrency as Xe, formatDuration as qe, formatFileSize as Ge, formatHashtag as Ke, formatInitials as _e, formatJson as ta, formatList as ea, formatMarkdown as aa, formatMention as ra, formatNumber as oa, formatOrdinal as ia, formatPercent as na, formatPhoneNumber as sa, formatPlural as ca, formatSSN as la, formatTemplate as da, formatTime as ma, stripFormatting as pa } from "./formatting/index.js";
function s() {
  return typeof navigator < "u" && !!navigator.clipboard;
}
function h() {
  return s() && typeof navigator.clipboard.read == "function";
}
async function m(e, a) {
  try {
    if (s())
      await navigator.clipboard.writeText(e);
    else {
      const t = document.createElement("textarea");
      t.value = e, t.style.position = "fixed", t.style.left = "-9999px", document.body.appendChild(t), t.select(), document.execCommand("copy"), document.body.removeChild(t);
    }
    return a?.notify && p(a.message || "Copied to clipboard", a.duration), !0;
  } catch (t) {
    return console.error("Failed to copy text:", t), !1;
  }
}
async function g(e, a, t) {
  try {
    if (s() && typeof ClipboardItem < "u") {
      const r = new Blob([e], { type: "text/html" }), o = new Blob([a || d(e)], { type: "text/plain" }), i = new ClipboardItem({
        "text/html": r,
        "text/plain": o
      });
      await navigator.clipboard.write([i]);
    } else
      return m(a || d(e), t);
    return t?.notify && p(t.message || "Copied to clipboard", t.duration), !0;
  } catch (r) {
    return console.error("Failed to copy HTML:", r), m(a || d(e), t);
  }
}
async function C(e, a) {
  try {
    const t = JSON.stringify(e, null, 2);
    return m(t, a);
  } catch (t) {
    return console.error("Failed to copy JSON:", t), !1;
  }
}
async function v(e, a) {
  try {
    if (!s()) return !1;
    let t;
    e instanceof Blob ? t = e : e instanceof HTMLImageElement ? t = await y(e) : t = await (await fetch(e)).blob(), t.type !== "image/png" && (t = await T(t));
    const r = new ClipboardItem({ "image/png": t });
    return await navigator.clipboard.write([r]), a?.notify && p(a.message || "Image copied to clipboard", a.duration), !0;
  } catch (t) {
    return console.error("Failed to copy image:", t), !1;
  }
}
async function b() {
  try {
    return s() ? await navigator.clipboard.readText() : null;
  } catch (e) {
    return console.error("Failed to read clipboard:", e), null;
  }
}
async function D() {
  const e = [];
  if (!h()) {
    const a = await b();
    return a && e.push({ type: "text", data: a, mimeType: "text/plain" }), e;
  }
  try {
    const a = await navigator.clipboard.read();
    for (const t of a)
      for (const r of t.types) {
        const o = await t.getType(r);
        if (r.startsWith("text/html"))
          e.push({ type: "html", data: await o.text(), mimeType: r });
        else if (r.startsWith("text/")) {
          const i = await o.text();
          u(i) ? e.push({ type: "json", data: i, mimeType: "application/json" }) : e.push({ type: "text", data: i, mimeType: r });
        } else r.startsWith("image/") ? e.push({ type: "image", data: o, mimeType: r }) : e.push({ type: "unknown", data: o, mimeType: r });
      }
  } catch (a) {
    console.error("Failed to read clipboard:", a);
  }
  return e;
}
function x(e) {
  const a = [];
  if (!e.clipboardData) return a;
  e.clipboardData.files.length > 0 && a.push({
    type: "file",
    data: Array.from(e.clipboardData.files),
    mimeType: "application/octet-stream"
  });
  const t = e.clipboardData.getData("text/html");
  t && a.push({ type: "html", data: t, mimeType: "text/html" });
  const r = e.clipboardData.getData("text/plain");
  return r && (u(r) ? a.push({ type: "json", data: r, mimeType: "application/json" }) : a.push({ type: "text", data: r, mimeType: "text/plain" })), a;
}
function F(e, a) {
  return (t) => {
    a?.preventDefault && t.preventDefault();
    const r = x(t);
    e(r, t);
  };
}
async function O(e, a, t) {
  const r = a.map((n) => n.header).join("	"), o = e.map(
    (n) => a.map((l) => String(n[l.key] ?? "")).join("	")
  ), i = [r, ...o].join(`
`), c = `
    <table>
      <thead><tr>${a.map((n) => `<th>${f(n.header)}</th>`).join("")}</tr></thead>
      <tbody>${e.map(
    (n) => `<tr>${a.map((l) => `<td>${f(String(n[l.key] ?? ""))}</td>`).join("")}</tr>`
  ).join("")}</tbody>
    </table>
  `;
  return g(c, i, t);
}
function S(e) {
  return e.split(/\r?\n/).filter((t) => t.trim()).map((t) => t.includes("	") ? t.split("	") : w(t));
}
function d(e) {
  return new DOMParser().parseFromString(e, "text/html").body.textContent || "";
}
function f(e) {
  const a = document.createElement("div");
  return a.textContent = e, a.innerHTML;
}
function u(e) {
  try {
    return JSON.parse(e), !0;
  } catch {
    return !1;
  }
}
async function y(e) {
  const a = document.createElement("canvas");
  return a.width = e.naturalWidth, a.height = e.naturalHeight, a.getContext("2d").drawImage(e, 0, 0), new Promise((r, o) => {
    a.toBlob((i) => {
      i ? r(i) : o(new Error("Failed to convert image to blob"));
    }, "image/png");
  });
}
async function T(e) {
  const a = new Image(), t = URL.createObjectURL(e);
  return new Promise((r, o) => {
    a.onload = async () => {
      URL.revokeObjectURL(t);
      try {
        r(await y(a));
      } catch (i) {
        o(i);
      }
    }, a.onerror = () => {
      URL.revokeObjectURL(t), o(new Error("Failed to load image"));
    }, a.src = t;
  });
}
function w(e) {
  const a = [];
  let t = "", r = !1;
  for (let o = 0; o < e.length; o++) {
    const i = e[o], c = e[o + 1];
    r ? i === '"' && c === '"' ? (t += '"', o++) : i === '"' ? r = !1 : t += i : i === '"' ? r = !0 : i === "," ? (a.push(t), t = "") : t += i;
  }
  return a.push(t), a;
}
function p(e, a = 2e3) {
  if (typeof document > "u") return;
  const t = document.createElement("div");
  t.textContent = e, t.style.cssText = `
    position: fixed;
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: #333;
    color: white;
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 14px;
    z-index: 9999;
    animation: fadeIn 0.2s ease-out;
  `, document.body.appendChild(t), setTimeout(() => {
    t.style.opacity = "0", t.style.transition = "opacity 0.2s", setTimeout(() => t.remove(), 200);
  }, a);
}
export {
  X as FileProcessor,
  q as FileSizeLimits,
  G as FileTypes,
  K as FileValidator,
  R as FormValidator,
  E as SchemaValidator,
  I as ValidationRules,
  Ot as addDays,
  St as addMonths,
  kt as addWeeks,
  Wt as addYears,
  ye as average,
  tt as camelCase,
  et as capitalize,
  at as characterCount,
  he as clamp,
  g as copyHTML,
  v as copyImage,
  C as copyJSON,
  O as copyTableData,
  m as copyText,
  F as createPasteHandler,
  L as createYupFormValidator,
  M as createZodFormValidator,
  ge as distance,
  rt as ellipsis,
  Rt as endOfDay,
  Et as endOfMonth,
  It as endOfWeek,
  Lt as endOfYear,
  ot as escapeHtml,
  it as extractEmails,
  nt as extractUrls,
  be as factorial,
  xe as fibonacci,
  Je as formatAddress,
  $e as formatBytes,
  Ze as formatCSV,
  Te as formatCompact,
  Qe as formatCreditCard,
  Xe as formatCurrency,
  Mt as formatDate,
  qe as formatDuration,
  Ge as formatFileSize,
  Ke as formatHashtag,
  _e as formatInitials,
  ta as formatJson,
  ea as formatList,
  aa as formatMarkdown,
  ra as formatMention,
  oa as formatNumber,
  ia as formatOrdinal,
  jt as formatPatterns,
  na as formatPercent,
  sa as formatPhoneNumber,
  ca as formatPlural,
  Pt as formatRelative,
  la as formatSSN,
  da as formatTemplate,
  ma as formatTime,
  Ht as fromUTC,
  we as gcd,
  Yt as getAge,
  Nt as getBusinessDays,
  Bt as getBusinessDaysCount,
  Ut as getDateRange,
  Vt as getDaysInMonth,
  zt as getDaysInYear,
  j as getFieldErrors,
  P as getFirstFieldError,
  At as getNthWeekdayOfMonth,
  Jt as getTimeZoneOffset,
  $t as getWeekOfYear,
  x as handlePaste,
  H as hasFieldError,
  st as highlight,
  Ce as inRange,
  h as isClipboardReadSupported,
  s as isClipboardSupported,
  ct as isEmail,
  Y as isEmpty,
  ve as isEven,
  Zt as isLeapYear,
  De as isOdd,
  lt as isPalindrome,
  Fe as isPrime,
  Qt as isSameDay,
  Xt as isSameMonth,
  qt as isSameWeek,
  Gt as isSameYear,
  Kt as isToday,
  _t as isTomorrow,
  dt as isUrl,
  te as isValidDate,
  ee as isWeekday,
  ae as isWeekend,
  re as isYesterday,
  mt as kebabCase,
  Oe as lcm,
  Se as lerp,
  pt as levenshteinDistance,
  N as mapToFormErrors,
  ft as mask,
  ke as median,
  We as mode,
  Re as normalize,
  oe as parseDate,
  Ee as parseNumber,
  S as parseTableData,
  ut as pascalCase,
  Ie as percentage,
  Le as product,
  Me as randomFloat,
  je as randomInt,
  yt as randomString,
  Pe as range,
  D as readClipboard,
  b as readText,
  ht as removeAccents,
  gt as removeHtml,
  bt as reverse,
  He as round,
  Ye as scale,
  xt as similarity,
  Tt as slugify,
  wt as snakeCase,
  Ne as standardDeviation,
  ie as startOfDay,
  ne as startOfMonth,
  se as startOfWeek,
  ce as startOfYear,
  pa as stripFormatting,
  le as subtractDays,
  de as subtractMonths,
  me as subtractWeeks,
  pe as subtractYears,
  Be as sum,
  Ue as toDegrees,
  Ve as toRadians,
  fe as toUTC,
  Ct as truncate,
  vt as unescapeHtml,
  B as validateWithYup,
  U as validateWithYupAsync,
  V as validateWithZod,
  z as validateWithZodAsync,
  ze as variance,
  Dt as wordCount,
  A as yupToAsyncRule,
  J as yupToRule,
  $ as zodToAsyncRule,
  Z as zodToRule
};
//# sourceMappingURL=index.js.map
