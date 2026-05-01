const i = (t, r = "USD", e = "en-US", n = {}) => new Intl.NumberFormat(e, {
  style: "currency",
  currency: r,
  ...n
}).format(t), u = (t, r = "en-US", e = {}) => new Intl.NumberFormat(r, {
  style: "percent",
  ...e
}).format(t), f = (t, r = "en-US", e = {}) => new Intl.NumberFormat(r, e).format(t), $ = (t, r = 2) => {
  if (t === 0) return "0 Bytes";
  const e = 1024, n = r < 0 ? 0 : r, o = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"], s = Math.floor(Math.log(t) / Math.log(e));
  return parseFloat((t / Math.pow(e, s)).toFixed(n)) + " " + o[s];
}, m = (t) => {
  const r = Math.floor(t / 1e3), e = Math.floor(r / 60), n = Math.floor(e / 60), o = Math.floor(n / 24);
  return o > 0 ? `${o}d ${n % 24}h ${e % 60}m ${r % 60}s` : n > 0 ? `${n}h ${e % 60}m ${r % 60}s` : e > 0 ? `${e}m ${r % 60}s` : `${r}s`;
}, p = (t, r = "12", e = !1) => {
  const n = {
    hour: "numeric",
    minute: "2-digit",
    hour12: r === "12"
  };
  return e && (n.second = "2-digit"), t.toLocaleTimeString("en-US", n);
}, g = (t, r = "US") => {
  const e = t.replace(/\D/g, "");
  return r === "US" && e.length === 10 ? `(${e.slice(0, 3)}) ${e.slice(3, 6)}-${e.slice(6)}` : r === "US" && e.length === 11 && e.startsWith("1") ? `+1 (${e.slice(1, 4)}) ${e.slice(4, 7)}-${e.slice(7)}` : r === "INTERNATIONAL" && e.length > 10 ? `+${e.slice(0, -10)} ${e.slice(-10, -7)} ${e.slice(-7, -4)} ${e.slice(-4)}` : t;
}, h = (t, r = " ") => t.replace(/\D/g, "").replace(/(.{4})/g, `$1${r}`).trim(), d = (t, r = !1) => {
  const e = t.replace(/\D/g, "");
  return e.length !== 9 ? t : r ? `XXX-XX-${e.slice(5)}` : `${e.slice(0, 3)}-${e.slice(3, 5)}-${e.slice(5)}`;
}, B = (t, r = "US") => {
  const { street: e, city: n, state: o, zipCode: s, country: l } = t;
  return r === "US" ? [
    e,
    [n, o].filter(Boolean).join(", "),
    s
  ].filter(Boolean).join(`
`) : [
    e,
    n,
    [o, s].filter(Boolean).join(" "),
    l
  ].filter(Boolean).join(`
`);
}, S = (t, r = "conjunction", e = "en-US") => t.length === 0 ? "" : t.length === 1 ? t[0] : new Intl.ListFormat(e, {
  style: "long",
  type: r
}).format(t), w = (t, r = 2) => t.trim().split(/\s+/).slice(0, r).map((o) => o.charAt(0).toUpperCase()).join(""), M = (t, r, e) => {
  const n = e || `${r}s`;
  return t === 1 ? `${t} ${r}` : `${t} ${n}`;
}, U = (t, r = "en-US") => {
  const e = new Intl.PluralRules(r, { type: "ordinal" }), n = /* @__PURE__ */ new Map([
    ["one", "st"],
    ["two", "nd"],
    ["few", "rd"],
    ["other", "th"]
  ]), o = e.select(t), s = n.get(o) || "th";
  return `${t}${s}`;
}, F = (t, r = !1) => {
  const e = r ? 1024 : 1e3, n = r ? ["B", "KiB", "MiB", "GiB", "TiB", "PiB"] : ["B", "KB", "MB", "GB", "TB", "PB"];
  if (t === 0) return "0 B";
  const o = Math.floor(Math.log(t) / Math.log(e));
  return `${(t / Math.pow(e, o)).toFixed(1)} ${n[o]}`;
}, j = (t) => t.toLowerCase().replace(/[^\w\s]/g, "").replace(/\s+/g, "").replace(/^/, "#"), y = (t) => `@${t.replace(/^@/, "")}`, N = (t, r) => t.replace(/\{\{(\w+)\}\}/g, (e, n) => r.hasOwnProperty(n) ? String(r[n]) : e), P = (t, r = 2) => {
  try {
    return JSON.stringify(t, null, r);
  } catch {
    return String(t);
  }
}, I = (t, r = ",") => {
  if (t.length === 0) return "";
  const e = Object.keys(t[0]), n = [e.join(r)];
  for (const o of t) {
    const s = e.map((l) => {
      const a = o[l], c = String(a).replace(/"/g, '""');
      return c.includes(r) || c.includes(`
`) || c.includes('"') ? `"${c}"` : c;
    });
    n.push(s.join(r));
  }
  return n.join(`
`);
}, T = {
  bold: (t) => `**${t}**`,
  italic: (t) => `*${t}*`,
  code: (t) => `\`${t}\``,
  codeBlock: (t, r) => `\`\`\`${r || ""}
${t}
\`\`\``,
  link: (t, r) => `[${t}](${r})`,
  image: (t, r) => `![${t}](${r})`,
  heading: (t, r) => `${"#".repeat(r)} ${t}`,
  quote: (t) => `> ${t}`,
  list: (t, r = !1) => t.map(
    (e, n) => r ? `${n + 1}. ${e}` : `- ${e}`
  ).join(`
`)
}, x = {
  html: (t) => t.replace(/<[^>]*>/g, ""),
  markdown: (t) => t.replace(/[*_`~]/g, "").replace(/\[([^\]]+)\]\([^)]+\)/g, "$1").replace(/^#+\s*/gm, "").replace(/^>\s*/gm, "").replace(/^[-*+]\s*/gm, "").replace(/^\d+\.\s*/gm, ""),
  whitespace: (t) => t.replace(/\s+/g, " ").trim(),
  nonPrintable: (t) => t.replace(/[^\x20-\x7E]/g, "")
};
export {
  B as formatAddress,
  $ as formatBytes,
  I as formatCSV,
  h as formatCreditCard,
  i as formatCurrency,
  m as formatDuration,
  F as formatFileSize,
  j as formatHashtag,
  w as formatInitials,
  P as formatJson,
  S as formatList,
  T as formatMarkdown,
  y as formatMention,
  f as formatNumber,
  U as formatOrdinal,
  u as formatPercent,
  g as formatPhoneNumber,
  M as formatPlural,
  d as formatSSN,
  N as formatTemplate,
  p as formatTime,
  x as stripFormatting
};
//# sourceMappingURL=index.js.map
