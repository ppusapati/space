/**
 * Masters domain module.
 *
 * Wires the 7 standalone List RPCs the masters domain currently
 * exposes — one per service, plus item-categories which is parent-id-
 * free and high-traffic enough to warrant its own menu entry. Sub-
 * resource Lists (ItemService/ListVariants, LocationService/ListZones,
 * TaxCodeService/ListTaxRates, …) are intentionally not wired:
 * they require a parent-id and have no usable empty-context view.
 *
 * Every (formId, listEndpoint, responseRowsKey) tuple is verified live
 * against the JWT monolith on 2026-04-29:
 *   - formId → FormService.GetFormSchema returns 200 with formDefinition
 *   - listEndpoint → returns 200 + rows array under empty-tenant context
 *   - responseRowsKey matches the proto field name of the rows array
 *
 * Columns are explicit per entity rather than derived from form schema
 * because the form catalog's form_id ↔ entity mapping is many-to-one
 * (one form_id covers a whole sub-area; ListPage needs row-shape
 * columns from the response, not from the form's create-shape fields).
 */
import type { DomainModule } from '../index.js';

export const masters: DomainModule = {
  id: 'masters',
  label: 'Masters',
  entities: [
    {
      slug: 'items',
      label: 'Items',
      formId: 'item_master_management',
      listEndpoint: '/masters.item.api.v1.ItemService/ListItems',
      responseRowsKey: 'items',
      responseTotalKey: 'totalCount',
      columns: ['itemCode', 'itemName', 'itemType', 'baseUomId'],
    },
    {
      slug: 'categories',
      label: 'Item Categories',
      formId: 'product_category',
      listEndpoint: '/masters.item.api.v1.ItemService/ListCategories',
      responseRowsKey: 'categories',
      responseTotalKey: 'totalCount',
      columns: ['categoryCode', 'categoryName', 'level', 'description'],
    },
    {
      slug: 'uoms',
      label: 'Units of Measure',
      formId: 'uom_conversion',
      listEndpoint: '/masters.uom.api.v1.UOMService/ListUOMs',
      responseRowsKey: 'uoms',
      responseTotalKey: 'totalCount',
      columns: ['uomCode', 'uomName', 'symbol', 'isBaseUnit'],
    },
    {
      slug: 'parties',
      label: 'Parties',
      formId: 'customer_master_management',
      listEndpoint: '/masters.party.api.v1.EntityService/ListEntities',
      responseRowsKey: 'entities',
      responseTotalKey: 'totalCount',
      columns: ['entityCode', 'displayName', 'entityType', 'status'],
    },
    {
      slug: 'locations',
      label: 'Locations',
      formId: 'warehouse_location',
      listEndpoint: '/masters.location.api.v1.LocationService/ListLocations',
      responseRowsKey: 'locations',
      responseTotalKey: 'totalCount',
      columns: ['locationCode', 'locationName', 'locationType', 'gstin'],
    },
    {
      slug: 'tax-codes',
      label: 'Tax Codes',
      formId: 'tax_configuration',
      listEndpoint: '/masters.taxcode.api.v1.TaxCodeService/ListTaxCodes',
      responseRowsKey: 'taxCodes',
      responseTotalKey: 'totalCount',
      columns: ['taxCode', 'taxName', 'taxType', 'jurisdiction'],
    },
    {
      slug: 'accounts',
      label: 'Chart of Accounts',
      formId: 'cost_center_setup',
      listEndpoint: '/masters.chartofaccounts.api.v1.ChartOfAccountsService/ListAccounts',
      responseRowsKey: 'accounts',
      responseTotalKey: 'totalCount',
      columns: ['accountCode', 'accountName', 'accountType', 'nature'],
    },
  ],
};
