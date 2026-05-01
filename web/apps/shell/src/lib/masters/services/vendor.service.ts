/**
 * Vendor Service
 * Manages vendor CRUD operations
 */

import type {
  Vendor,
  VendorFilters,
  CreateVendorRequest,
  UpdateVendorRequest,
  VendorCategory,
} from '$lib/masters/types';
// import type { ListParams, PaginatedResponse } from '@samavāya/api';

// ============================================================================
// VENDOR SERVICE
// ============================================================================

class VendorService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:9090';
  }

  /**
   * Get paginated list of vendors
   */
  // async getVendors(params?: ListParams<VendorFilters>): Promise<PaginatedResponse<Vendor>> {
  //   const queryParams = new URLSearchParams();

  //   if (params?.page) queryParams.set('page', String(params.page));
  //   if (params?.pageSize) queryParams.set('pageSize', String(params.pageSize));
  //   if (params?.sortBy) queryParams.set('sortBy', params.sortBy);
  //   if (params?.sortOrder) queryParams.set('sortOrder', params.sortOrder);
  //   if (params?.search) queryParams.set('search', params.search);
  //   if (params?.filters?.status) queryParams.set('status', params.filters.status);
  //   if (params?.filters?.type) queryParams.set('type', params.filters.type);
  //   if (params?.filters?.categoryId) queryParams.set('categoryId', params.filters.categoryId);
  //   if (params?.filters?.minRating) queryParams.set('minRating', String(params.filters.minRating));

  //   const response = await fetch(
  //     `${this.baseUrl}/vendors?${queryParams.toString()}`,
  //     {
  //       method: 'GET',
  //       credentials: 'include',
  //     }
  //   );

  //   if (!response.ok) {
  //     throw new Error('Failed to fetch vendors');
  //   }

  //   const data = await response.json();

  //   // Parse dates
  //   data.items = data.items.map((vendor: Vendor) => ({
  //     ...vendor,
  //     createdAt: new Date(vendor.createdAt),
  //     updatedAt: new Date(vendor.updatedAt),
  //   }));

  //   return data;
  // }

  /**
   * Get vendor by ID
   */
  async getVendor(id: string): Promise<Vendor> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Vendor not found');
      }
      throw new Error('Failed to fetch vendor');
    }

    const vendor = await response.json();

    return {
      ...vendor,
      createdAt: new Date(vendor.createdAt),
      updatedAt: new Date(vendor.updatedAt),
    };
  }

  /**
   * Create new vendor
   */
  async createVendor(request: CreateVendorRequest): Promise<Vendor> {
    const response = await fetch(`${this.baseUrl}/vendors`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create vendor');
    }

    const vendor = await response.json();

    return {
      ...vendor,
      createdAt: new Date(vendor.createdAt),
      updatedAt: new Date(vendor.updatedAt),
    };
  }

  /**
   * Update vendor
   */
  async updateVendor(id: string, request: UpdateVendorRequest): Promise<Vendor> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to update vendor');
    }

    const vendor = await response.json();

    return {
      ...vendor,
      createdAt: new Date(vendor.createdAt),
      updatedAt: new Date(vendor.updatedAt),
    };
  }

  /**
   * Delete vendor
   */
  async deleteVendor(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to delete vendor');
    }
  }

  /**
   * Activate vendor
   */
  async activateVendor(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}/activate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to activate vendor');
    }
  }

  /**
   * Deactivate vendor
   */
  async deactivateVendor(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}/deactivate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to deactivate vendor');
    }
  }

  /**
   * Get vendor categories
   */
  async getCategories(): Promise<VendorCategory[]> {
    const response = await fetch(`${this.baseUrl}/vendors/categories`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch vendor categories');
    }

    return response.json();
  }

  /**
   * Get vendor performance summary
   */
  async getPerformanceSummary(id: string): Promise<{
    totalOrders: number;
    onTimeDeliveryRate: number;
    qualityScore: number;
    averageLeadTime: number;
  }> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}/performance`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch vendor performance');
    }

    return response.json();
  }

  /**
   * Update vendor rating
   */
  async updateRating(id: string, rating: number): Promise<void> {
    const response = await fetch(`${this.baseUrl}/vendors/${id}/rating`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ rating }),
    });

    if (!response.ok) {
      throw new Error('Failed to update vendor rating');
    }
  }
}

export const vendorService = new VendorService();
