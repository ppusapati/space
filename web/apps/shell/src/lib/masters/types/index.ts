/**
 * Masters Module Types
 * Types for customers, vendors, items, and other master data
 * @packageDocumentation
 */

// ============================================================================
// COMMON TYPES
// ============================================================================

export type EntityStatus = 'active' | 'inactive' | 'archived';

export interface Address {
  id?: string;
  type: 'billing' | 'shipping' | 'office' | 'warehouse' | 'other';
  label?: string;
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  isDefault?: boolean;
}

export interface Contact {
  id?: string;
  type: 'primary' | 'billing' | 'sales' | 'support' | 'other';
  name: string;
  title?: string;
  email?: string;
  phone?: string;
  mobile?: string;
  isDefault?: boolean;
}

export interface BankAccount {
  id?: string;
  bankName: string;
  accountName: string;
  accountNumber: string;
  routingNumber?: string;
  swiftCode?: string;
  iban?: string;
  isDefault?: boolean;
}

export interface TaxInfo {
  taxId?: string;
  taxType?: string;
  exemptionNumber?: string;
  isExempt?: boolean;
}

// ============================================================================
// CUSTOMER TYPES
// ============================================================================

export type CustomerType = 'individual' | 'company' | 'government' | 'nonprofit';

export interface Customer {
  id: string;
  code: string;
  type: CustomerType;
  name: string;
  displayName: string;
  email?: string;
  phone?: string;
  website?: string;
  status: EntityStatus;

  // Organization
  parentId?: string;
  parent?: Customer;

  // Addresses and Contacts
  addresses: Address[];
  contacts: Contact[];

  // Financial
  creditLimit?: number;
  creditBalance?: number;
  paymentTerms?: string;
  currency: string;
  priceListId?: string;
  discountPercent?: number;
  taxInfo?: TaxInfo;
  bankAccounts: BankAccount[];

  // Classification
  categoryId?: string;
  category?: CustomerCategory;
  tags: string[];

  // Sales
  salesRepId?: string;
  salesRep?: { id: string; displayName: string };
  territoryId?: string;

  // Metadata
  notes?: string;
  customFields?: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
  createdBy?: string;
  updatedBy?: string;
}

export interface CustomerCategory {
  id: string;
  code: string;
  name: string;
  description?: string;
  parentId?: string;
  color?: string;
}

export interface CustomerFilters {
  status?: EntityStatus;
  type?: CustomerType;
  categoryId?: string;
  salesRepId?: string;
  hasCreditBalance?: boolean;
}

export interface CreateCustomerRequest {
  code?: string;
  type: CustomerType;
  name: string;
  displayName?: string;
  email?: string;
  phone?: string;
  website?: string;
  addresses?: Address[];
  contacts?: Contact[];
  creditLimit?: number;
  paymentTerms?: string;
  currency?: string;
  priceListId?: string;
  discountPercent?: number;
  taxInfo?: TaxInfo;
  categoryId?: string;
  tags?: string[];
  salesRepId?: string;
  notes?: string;
  customFields?: Record<string, unknown>;
}

export interface UpdateCustomerRequest extends Partial<CreateCustomerRequest> {
  status?: EntityStatus;
}

// ============================================================================
// VENDOR TYPES
// ============================================================================

export type VendorType = 'supplier' | 'contractor' | 'service' | 'manufacturer' | 'distributor';

export interface Vendor {
  id: string;
  code: string;
  type: VendorType;
  name: string;
  displayName: string;
  email?: string;
  phone?: string;
  website?: string;
  status: EntityStatus;

  // Organization
  parentId?: string;
  parent?: Vendor;

  // Addresses and Contacts
  addresses: Address[];
  contacts: Contact[];

  // Financial
  creditLimit?: number;
  paymentTerms?: string;
  currency: string;
  taxInfo?: TaxInfo;
  bankAccounts: BankAccount[];

  // Classification
  categoryId?: string;
  category?: VendorCategory;
  tags: string[];

  // Purchasing
  leadTime?: number;
  minimumOrderValue?: number;
  rating?: number;

  // Metadata
  notes?: string;
  customFields?: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
  createdBy?: string;
  updatedBy?: string;
}

export interface VendorCategory {
  id: string;
  code: string;
  name: string;
  description?: string;
  parentId?: string;
  color?: string;
}

export interface VendorFilters {
  status?: EntityStatus;
  type?: VendorType;
  categoryId?: string;
  minRating?: number;
}

export interface CreateVendorRequest {
  code?: string;
  type: VendorType;
  name: string;
  displayName?: string;
  email?: string;
  phone?: string;
  website?: string;
  addresses?: Address[];
  contacts?: Contact[];
  paymentTerms?: string;
  currency?: string;
  taxInfo?: TaxInfo;
  categoryId?: string;
  tags?: string[];
  leadTime?: number;
  minimumOrderValue?: number;
  notes?: string;
  customFields?: Record<string, unknown>;
}

export interface UpdateVendorRequest extends Partial<CreateVendorRequest> {
  status?: EntityStatus;
}

// ============================================================================
// ITEM TYPES
// ============================================================================

export type ItemType = 'goods' | 'service' | 'bundle' | 'virtual';
export type ItemTrackingType = 'none' | 'batch' | 'serial';
export type UnitOfMeasure = 'each' | 'kg' | 'g' | 'lb' | 'oz' | 'l' | 'ml' | 'm' | 'cm' | 'in' | 'ft' | 'box' | 'case' | 'pallet';

export interface Item {
  id: string;
  code: string;
  sku: string;
  barcode?: string;
  type: ItemType;
  name: string;
  description?: string;
  status: EntityStatus;

  // Classification
  categoryId?: string;
  category?: ItemCategory;
  brandId?: string;
  brand?: { id: string; name: string };
  tags: string[];

  // Pricing
  costPrice: number;
  sellingPrice: number;
  currency: string;
  taxRate?: number;
  taxCategoryId?: string;
  priceListPrices?: { priceListId: string; price: number }[];

  // Units
  baseUnit: UnitOfMeasure;
  purchaseUnit?: UnitOfMeasure;
  salesUnit?: UnitOfMeasure;
  unitConversions?: { fromUnit: UnitOfMeasure; toUnit: UnitOfMeasure; factor: number }[];

  // Inventory
  trackingType: ItemTrackingType;
  isStockItem: boolean;
  reorderLevel?: number;
  reorderQuantity?: number;
  safetyStock?: number;
  leadTime?: number;
  defaultWarehouseId?: string;

  // Physical
  weight?: number;
  weightUnit?: 'kg' | 'lb';
  dimensions?: { length: number; width: number; height: number; unit: 'cm' | 'in' };

  // Purchasing
  preferredVendorId?: string;
  preferredVendor?: { id: string; name: string };
  vendorItems?: VendorItem[];

  // Media
  imageUrl?: string;
  images?: string[];

  // Metadata
  notes?: string;
  customFields?: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
  createdBy?: string;
  updatedBy?: string;
}

export interface VendorItem {
  vendorId: string;
  vendorName: string;
  vendorItemCode?: string;
  vendorItemName?: string;
  price: number;
  currency: string;
  leadTime?: number;
  minOrderQty?: number;
  isPreferred?: boolean;
}

export interface ItemCategory {
  id: string;
  code: string;
  name: string;
  description?: string;
  parentId?: string;
  parent?: ItemCategory;
  level: number;
  path: string;
  imageUrl?: string;
  itemCount?: number;
}

export interface ItemFilters {
  status?: EntityStatus;
  type?: ItemType;
  categoryId?: string;
  brandId?: string;
  isStockItem?: boolean;
  trackingType?: ItemTrackingType;
  hasLowStock?: boolean;
}

export interface CreateItemRequest {
  code?: string;
  sku: string;
  barcode?: string;
  type: ItemType;
  name: string;
  description?: string;
  categoryId?: string;
  brandId?: string;
  tags?: string[];
  costPrice: number;
  sellingPrice: number;
  currency?: string;
  taxRate?: number;
  taxCategoryId?: string;
  baseUnit: UnitOfMeasure;
  purchaseUnit?: UnitOfMeasure;
  salesUnit?: UnitOfMeasure;
  trackingType?: ItemTrackingType;
  isStockItem?: boolean;
  reorderLevel?: number;
  reorderQuantity?: number;
  weight?: number;
  weightUnit?: 'kg' | 'lb';
  preferredVendorId?: string;
  imageUrl?: string;
  notes?: string;
  customFields?: Record<string, unknown>;
}

export interface UpdateItemRequest extends Partial<CreateItemRequest> {
  status?: EntityStatus;
}

// ============================================================================
// UNIT OF MEASURE TYPES
// ============================================================================

export interface UnitOfMeasureConfig {
  id: string;
  code: UnitOfMeasure;
  name: string;
  symbol: string;
  category: 'quantity' | 'weight' | 'volume' | 'length' | 'area';
  baseUnit: UnitOfMeasure;
  conversionFactor: number;
  isActive: boolean;
}

// ============================================================================
// PRICE LIST TYPES
// ============================================================================

export interface PriceList {
  id: string;
  code: string;
  name: string;
  description?: string;
  currency: string;
  type: 'sales' | 'purchase';
  isDefault: boolean;
  isActive: boolean;
  validFrom?: Date;
  validTo?: Date;
  priceCount?: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface PriceListItem {
  id: string;
  priceListId: string;
  itemId: string;
  item?: { id: string; code: string; name: string };
  price: number;
  minQuantity?: number;
  maxQuantity?: number;
  validFrom?: Date;
  validTo?: Date;
}
