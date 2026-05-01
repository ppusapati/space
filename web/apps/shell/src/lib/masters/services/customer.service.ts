/**
 * Customer Service
 * Manages customer CRUD operations
 */

import type {
  Customer,
  CustomerFilters,
  CreateCustomerRequest,
  UpdateCustomerRequest,
  CustomerCategory,
} from '$lib/masters/types';
// import type { ListParams, PaginatedResponse } from '@samavāya/api';

// ============================================================================
// CUSTOMER SERVICE
// ============================================================================

class CustomerService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:9090';
  }

  /**
   * Get paginated list of customers
   */
  // async getCustomers(params?: ListParams<CustomerFilters>): Promise<PaginatedResponse<Customer>> {
  //   const queryParams = new URLSearchParams();

  //   if (params?.page) queryParams.set('page', String(params.page));
  //   if (params?.pageSize) queryParams.set('pageSize', String(params.pageSize));
  //   if (params?.sortBy) queryParams.set('sortBy', params.sortBy);
  //   if (params?.sortOrder) queryParams.set('sortOrder', params.sortOrder);
  //   if (params?.search) queryParams.set('search', params.search);
  //   if (params?.filters?.status) queryParams.set('status', params.filters.status);
  //   if (params?.filters?.type) queryParams.set('type', params.filters.type);
  //   if (params?.filters?.categoryId) queryParams.set('categoryId', params.filters.categoryId);
  //   if (params?.filters?.salesRepId) queryParams.set('salesRepId', params.filters.salesRepId);

  //   const response = await fetch(
  //     `${this.baseUrl}/customers?${queryParams.toString()}`,
  //     {
  //       method: 'GET',
  //       credentials: 'include',
  //     }
  //   );

  //   if (!response.ok) {
  //     throw new Error('Failed to fetch customers');
  //   }

  //   const data = await response.json();

  //   // Parse dates
  //   data.items = data.items.map((customer: Customer) => ({
  //     ...customer,
  //     createdAt: new Date(customer.createdAt),
  //     updatedAt: new Date(customer.updatedAt),
  //   }));

  //   return data;
  // }

  /**
   * Get customer by ID
   */
  async getCustomer(id: string): Promise<Customer> {
    const response = await fetch(`${this.baseUrl}/customers/${id}`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Customer not found');
      }
      throw new Error('Failed to fetch customer');
    }

    const customer = await response.json();

    return {
      ...customer,
      createdAt: new Date(customer.createdAt),
      updatedAt: new Date(customer.updatedAt),
    };
  }

  /**
   * Create new customer
   */
  async createCustomer(request: CreateCustomerRequest): Promise<Customer> {
    const response = await fetch(`${this.baseUrl}/customers`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create customer');
    }

    const customer = await response.json();

    return {
      ...customer,
      createdAt: new Date(customer.createdAt),
      updatedAt: new Date(customer.updatedAt),
    };
  }

  /**
   * Update customer
   */
  async updateCustomer(id: string, request: UpdateCustomerRequest): Promise<Customer> {
    const response = await fetch(`${this.baseUrl}/customers/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to update customer');
    }

    const customer = await response.json();

    return {
      ...customer,
      createdAt: new Date(customer.createdAt),
      updatedAt: new Date(customer.updatedAt),
    };
  }

  /**
   * Delete customer
   */
  async deleteCustomer(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/customers/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to delete customer');
    }
  }

  /**
   * Activate customer
   */
  async activateCustomer(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/customers/${id}/activate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to activate customer');
    }
  }

  /**
   * Deactivate customer
   */
  async deactivateCustomer(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/customers/${id}/deactivate`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to deactivate customer');
    }
  }

  /**
   * Get customer categories
   */
  async getCategories(): Promise<CustomerCategory[]> {
    const response = await fetch(`${this.baseUrl}/customers/categories`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch customer categories');
    }

    return response.json();
  }

  /**
   * Get customer credit summary
   */
  async getCreditSummary(id: string): Promise<{
    creditLimit: number;
    usedCredit: number;
    availableCredit: number;
    overdueAmount: number;
  }> {
    const response = await fetch(`${this.baseUrl}/customers/${id}/credit`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch credit summary');
    }

    return response.json();
  }

  /**
   * Get customer transactions summary
   */
  async getTransactionsSummary(id: string): Promise<{
    totalOrders: number;
    totalRevenue: number;
    averageOrderValue: number;
    lastOrderDate: Date | null;
  }> {
    const response = await fetch(`${this.baseUrl}/customers/${id}/transactions`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transactions summary');
    }

    const data = await response.json();
    return {
      ...data,
      lastOrderDate: data.lastOrderDate ? new Date(data.lastOrderDate) : null,
    };
  }
}

export const customerService = new CustomerService();
