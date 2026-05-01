/**
 * Domain module registry — central index of every domain's routes
 * and menu entries.
 *
 * Each `lib/modules/<domain>/` folder exports two things:
 *
 *   routes: DomainRoutes  — declares the entities the domain owns,
 *                           each with its formId + listEndpoint shape.
 *   menu:   DomainMenu    — the items shown in the sidebar for this
 *                           domain, mapping to URLs the routes render.
 *
 * The shell consumes both: the menu populates ErpShell's sub-nav,
 * and the routes drive a generic `[domain]/[entity]` SvelteKit route
 * that picks the right form/list page based on the registry.
 *
 * Adding a new domain to the demo set: create `lib/modules/<domain>/`
 * + import its `routes` and `menu` here. No code changes elsewhere.
 */

import { masters } from './masters/index.js';
import { asset } from './asset/index.js';
import { sales } from './sales/index.js';
import { hr } from './hr/index.js';
import { finance } from './finance/index.js';
import { purchase } from './purchase/index.js';
import { manufacturing } from './manufacturing/index.js';
import { projects } from './projects/index.js';
import { inventory } from './inventory/index.js';
import { fulfillment } from './fulfillment/index.js';
import { audit } from './audit/index.js';
import { banking } from './banking/index.js';
import { budget } from './budget/index.js';
import { communication } from './communication/index.js';
import { data } from './data/index.js';
import { identity } from './identity/index.js';
import { notifications } from './notifications/index.js';
import { platform } from './platform/index.js';
import { workflow } from './workflow/index.js';

export interface EntityRoute {
  /** Path segment under the domain — e.g. "items" → /<domain>/items */
  slug: string;
  /** Display label for the menu */
  label: string;
  /** Form_id for the create + list schema (proto FormDefinition) */
  formId: string;
  /** Full list-RPC endpoint, e.g. "/masters.item.api.v1.ItemService/ListItems" */
  listEndpoint: string;
  /** Key in list response holding the rows array, e.g. "items" */
  responseRowsKey: string;
  /** Key in list response holding the total count, e.g. "totalCount". Some services don't return one — that's fine, list page falls back to rows.length. */
  responseTotalKey?: string;
  /** Column names to show in the list view — overrides schema coreFields if set */
  columns?: string[];
}

export interface DomainModule {
  /** Domain id (e.g. "masters") */
  id: string;
  /** Display label */
  label: string;
  /** Entities this domain owns — each maps to a /<domain>/<slug> route */
  entities: EntityRoute[];
}

export const DOMAIN_MODULES: Record<string, DomainModule> = {
  masters,
  asset,
  sales,
  hr,
  finance,
  purchase,
  manufacturing,
  projects,
  inventory,
  fulfillment,
  audit,
  banking,
  budget,
  communication,
  data,
  identity,
  notifications,
  platform,
  workflow,
};

/** Look up the entity config for a `(domain, slug)` pair. */
export function getEntity(domainId: string, slug: string): EntityRoute | undefined {
  return DOMAIN_MODULES[domainId]?.entities.find((e) => e.slug === slug);
}

/** Iterate all (domain, entity) pairs — useful for building a global menu. */
export function allEntities(): Array<{ domain: DomainModule; entity: EntityRoute }> {
  const out: Array<{ domain: DomainModule; entity: EntityRoute }> = [];
  for (const domain of Object.values(DOMAIN_MODULES)) {
    for (const entity of domain.entities) {
      out.push({ domain, entity });
    }
  }
  return out;
}
