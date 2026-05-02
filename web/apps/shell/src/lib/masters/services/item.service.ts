/**
 * Item Service
 * Manages item/product CRUD operations
 */

import type {
  Item,
  ItemFilters,
  CreateItemRequest,
  UpdateItemRequest,
  ItemCategory,
  VendorItem,
  PriceList,
  PriceListItem,
} from '$lib/masters/types';
// import type { ListParams, PaginatedResponse } from '@chetana/api';

// ============================================================================
// ITEM SERVICE
// ============================================================================

class ItemService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:9090';
  }

  /**
   * Get paginated list of items
   */
  // async getItems(params?: ListParams<ItemFilters>): Promise<PaginatedResponse<Item>> {
  //   const queryParams = new URLSearchParams();

  //   if (params?.page) queryParams.set('page', String(params.page));
  //   if (params?.pageSize) queryParams.set('pageSize', String(params.pageSize));
  //   if (params?.sortBy) queryParams.set('sortBy', params.sortBy);
  //   if (params?.sortOrder) queryParams.set('sortOrder', params.sortOrder);
  //   if (params?.search) queryParams.set('search', params.search);
  //   if (params?.filters?.status) queryParams.set('status', params.filters.status);
  //   if (params?.filters?.type) queryParams.set('type', params.filters.type);
  //   if (params?.filters?.categoryId) queryParams.set('categoryId', params.filters.categoryId);
  //   if (params?.filters?.brandId) queryParams.set('brandId', params.filters.brandId);
  //   if (params?.filters?.isStockItem !== undefined) {
  //     queryParams.set('isStockItem', String(params.filters.isStockItem));
  //   }
  //   if (params?.filters?.hasLowStock) queryParams.set('hasLowStock', 'true');

  //   const response = await fetch(
  //     `${this.baseUrl}/items?${queryParams.toString()}`,
  //     {
  //       method: 'GET',
  //       credentials: 'include',
  //     }
  //   );

  //   if (!response.ok) {
  //     throw new Error('Failed to fetch items');
  //   }

  //   const data = await response.json();

  //   // Parse dates
  //   data.items = data.items.map((item: Item) => ({
  //     ...item,
  //     createdAt: new Date(item.createdAt),
  //     updatedAt: new Date(item.updatedAt),
  //   }));

  //   return data;
  // }

  /**
   * Get item by ID
   */
  async getItem(id: string): Promise<Item> {
    const response = await fetch(`${this.baseUrl}/items/${id}`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Item not found');
      }
      throw new Error('Failed to fetch item');
    }

    const item = await response.json();

    return {
      ...item,
      createdAt: new Date(item.createdAt),
      updatedAt: new Date(item.updatedAt),
    };
  }

  /**
   * Create new item
   */
  async createItem(request: CreateItemRequest): Promise<Item> {
    const response = await fetch(`${this.baseUrl}/items`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create item');
    }

    const item = await response.json();

    return {
      ...item,
      createdAt: new Date(item.createdAt),
      updatedAt: new Date(item.updatedAt),
    };
  }

  /**
   * Update item
   */
  async updateItem(id: string, request: UpdateItemRequest): Promise<Item> {
    const response = await fetch(`${this.baseUrl}/items/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to update item');
    }

    const item = await response.json();

    return {
      ...item,
      createdAt: new Date(item.createdAt),
      updatedAt: new Date(item.updatedAt),
    };
  }

  /**
   * Delete item
   */
  async deleteItem(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/items/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to delete item');
    }
  }

  /**
   * Activate item
   */
  async activateItem(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/items/${id}/activate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to activate item');
    }
  }

  /**
   * Deactivate item
   */
  async deactivateItem(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/items/${id}/deactivate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to deactivate item');
    }
  }

  /**
   * Get item categories (hierarchical)
   */
  async getCategories(): Promise<ItemCategory[]> {
    const response = await fetch(`${this.baseUrl}/items/categories`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch item categories');
    }

    return response.json();
  }

  /**
   * Get category by ID
   */
  async getCategory(id: string): Promise<ItemCategory> {
    const response = await fetch(`${this.baseUrl}/items/categories/${id}`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch category');
    }

    return response.json();
  }

  /**
   * Get item's vendor list
   */
  async getItemVendors(id: string): Promise<VendorItem[]> {
    const response = await fetch(`${this.baseUrl}/items/${id}/vendors`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch item vendors');
    }

    return response.json();
  }

  /**
   * Add vendor to item
   */
  async addItemVendor(id: string, vendorItem: Omit<VendorItem, 'vendorName'>): Promise<void> {
    const response = await fetch(`${this.baseUrl}/items/${id}/vendors`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(vendorItem),
    });

    if (!response.ok) {
      throw new Error('Failed to add vendor to item');
    }
  }

  /**
   * Get item stock levels across warehouses
   */
  async getStockLevels(id: string): Promise<{
    warehouseId: string;
    warehouseName: string;
    onHand: number;
    available: number;
    reserved: number;
    incoming: number;
  }[]> {
    const response = await fetch(`${this.baseUrl}/items/${id}/stock`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch stock levels');
    }

    return response.json();
  }

  /**
   * Upload item image
   */
  async uploadImage(id: string, file: File): Promise<string> {
    const formData = new FormData();
    formData.append('image', file);

    const response = await fetch(`${this.baseUrl}/items/${id}/image`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });

    if (!response.ok) {
      throw new Error('Failed to upload image');
    }

    const data = await response.json();
    return data.imageUrl;
  }

  /**
   * Get price lists
   */
  async getPriceLists(type?: 'sales' | 'purchase'): Promise<PriceList[]> {
    const queryParams = new URLSearchParams();
    if (type) queryParams.set('type', type);

    const response = await fetch(
      `${this.baseUrl}/price-lists?${queryParams.toString()}`,
      {
        method: 'GET',
        credentials: 'include',
      }
    );

    if (!response.ok) {
      throw new Error('Failed to fetch price lists');
    }

    const data = await response.json();
    return data.map((pl: PriceList) => ({
      ...pl,
      validFrom: pl.validFrom ? new Date(pl.validFrom) : undefined,
      validTo: pl.validTo ? new Date(pl.validTo) : undefined,
      createdAt: new Date(pl.createdAt),
      updatedAt: new Date(pl.updatedAt),
    }));
  }

  /**
   * Get item prices from all price lists
   */
  async getItemPrices(id: string): Promise<PriceListItem[]> {
    const response = await fetch(`${this.baseUrl}/items/${id}/prices`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch item prices');
    }

    return response.json();
  }

  /**
   * Update item price in a price list
   */
  async updateItemPrice(
    id: string,
    priceListId: string,
    price: number,
    minQuantity?: number,
    maxQuantity?: number
  ): Promise<void> {
    const response = await fetch(`${this.baseUrl}/items/${id}/prices/${priceListId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ price, minQuantity, maxQuantity }),
    });

    if (!response.ok) {
      throw new Error('Failed to update item price');
    }
  }

  /**
   * Lookup item by barcode
   */
  async lookupByBarcode(barcode: string): Promise<Item | null> {
    const response = await fetch(
      `${this.baseUrl}/items/lookup?barcode=${encodeURIComponent(barcode)}`,
      {
        method: 'GET',
        credentials: 'include',
      }
    );

    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      throw new Error('Failed to lookup item');
    }

    const item = await response.json();
    return {
      ...item,
      createdAt: new Date(item.createdAt),
      updatedAt: new Date(item.updatedAt),
    };
  }
}

export const itemService = new ItemService();
