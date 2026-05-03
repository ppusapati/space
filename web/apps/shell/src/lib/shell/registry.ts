/**
 * registry.ts — single source of truth for the chetana-platform
 * left-nav. Adding a new route = adding one entry here.
 *
 * Acceptance #5 of TASK-P1-WEB-001: the route registry remains
 * the single source of truth for `(app)/[domain]/[entity]/+page.svelte`.
 */

export interface ChetanaRoute {
  id: string;
  /** Absolute URL path (matches a SvelteKit route under (app)). */
  path: string;
  /** Left-nav label. */
  label: string;
  /** Lucide-style icon identifier rendered as a glyph. The shell
   *  uses a small inline glyph map rather than dragging in an
   *  icon dependency for the chetana surface. */
  icon: string;
  /** Optional permission (for visibility filtering once authz is
   *  wired into the shell — currently every authenticated user
   *  sees every route; the server-side handler enforces the real
   *  authz). */
  permission?: string;
  /** Section heading the entry sits under. */
  section?: "Operations" | "Settings";
}

export const chetanaRouteRegistry: readonly ChetanaRoute[] = [
  { id: "dashboard", path: "/dashboard", label: "Dashboard", icon: "▣", section: "Operations" },
  { id: "audit", path: "/audit", label: "Audit log", icon: "📜", section: "Operations" },
  { id: "exports", path: "/exports", label: "Exports", icon: "⬇", section: "Operations" },
  // Settings
  { id: "settings-sessions", path: "/settings/sessions", label: "Sessions", icon: "🖥", section: "Settings" },
  { id: "settings-api-keys", path: "/settings/api-keys", label: "API keys", icon: "🔑", section: "Settings" },
  { id: "settings-mfa", path: "/settings/mfa", label: "MFA", icon: "🛡", section: "Settings" },
];
