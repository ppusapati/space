import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import DataGrid from '../DataGrid.svelte';
import type { TableColumn } from '../table.types';

interface TestData {
  id: number;
  name: string;
  email: string;
  age: number;
  status: string;
}

const testColumns: TableColumn<TestData>[] = [
  { key: 'id', header: 'ID', sortable: true },
  { key: 'name', header: 'Name', sortable: true, filterable: true },
  { key: 'email', header: 'Email', sortable: true, filterable: true },
  { key: 'age', header: 'Age', sortable: true, filterType: 'number' },
  { key: 'status', header: 'Status', sortable: true },
];

const testData: TestData[] = [
  { id: 1, name: 'Alice Smith', email: 'alice@example.com', age: 28, status: 'Active' },
  { id: 2, name: 'Bob Johnson', email: 'bob@example.com', age: 35, status: 'Inactive' },
  { id: 3, name: 'Charlie Brown', email: 'charlie@example.com', age: 42, status: 'Active' },
  { id: 4, name: 'Diana Ross', email: 'diana@example.com', age: 31, status: 'Pending' },
  { id: 5, name: 'Edward Chen', email: 'edward@example.com', age: 29, status: 'Active' },
];

describe('DataGrid', () => {
  describe('Rendering', () => {
    it('renders with data and columns', () => {
      render(DataGrid, {
        props: { columns: testColumns, data: testData },
      });

      // Check headers are rendered
      expect(screen.getByText('ID')).toBeInTheDocument();
      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Email')).toBeInTheDocument();

      // Check data is rendered
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
      expect(screen.getByText('bob@example.com')).toBeInTheDocument();
    });

    it('renders empty state when no data', () => {
      render(DataGrid, {
        props: {
          columns: testColumns,
          data: [],
          emptyMessage: 'No records found',
        },
      });

      expect(screen.getByText('No records found')).toBeInTheDocument();
    });

    it('renders with custom test id', () => {
      render(DataGrid, {
        props: { columns: testColumns, data: testData, testId: 'my-grid' },
      });

      expect(screen.getByTestId('my-grid')).toBeInTheDocument();
    });
  });

  describe('Sorting', () => {
    it('sorts by column when header is clicked', async () => {
      const handleSort = vi.fn();
      const { component } = render(DataGrid, {
        props: { columns: testColumns, data: testData },
      });

      component.$on('sort', handleSort);

      // Click on Name header to sort
      const nameHeader = screen.getByText('Name').closest('button');
      if (nameHeader) {
        await fireEvent.click(nameHeader);
      }

      expect(handleSort).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            column: 'name',
            direction: 'asc',
          }),
        })
      );
    });

    it('toggles sort direction on repeated clicks', async () => {
      const handleSort = vi.fn();
      const { component } = render(DataGrid, {
        props: { columns: testColumns, data: testData },
      });

      component.$on('sort', handleSort);

      const nameHeader = screen.getByText('Name').closest('button');

      if (nameHeader) {
        // First click - ascending
        await fireEvent.click(nameHeader);
        expect(handleSort).toHaveBeenLastCalledWith(
          expect.objectContaining({
            detail: expect.objectContaining({ direction: 'asc' }),
          })
        );

        // Second click - descending
        await fireEvent.click(nameHeader);
        expect(handleSort).toHaveBeenLastCalledWith(
          expect.objectContaining({
            detail: expect.objectContaining({ direction: 'desc' }),
          })
        );

        // Third click - no sort
        await fireEvent.click(nameHeader);
        expect(handleSort).toHaveBeenLastCalledWith(
          expect.objectContaining({
            detail: expect.objectContaining({ direction: null }),
          })
        );
      }
    });
  });

  describe('Filtering', () => {
    it('shows filter panel when filter button is clicked', async () => {
      render(DataGrid, {
        props: { columns: testColumns, data: testData, filterable: true },
      });

      const filterButton = screen.getByText('Filter');
      await fireEvent.click(filterButton);

      // Should show add filter button
      expect(screen.getByText('+ Add filter')).toBeInTheDocument();
    });

    it('emits filter event when filter is applied', async () => {
      const handleFilter = vi.fn();
      const { component } = render(DataGrid, {
        props: { columns: testColumns, data: testData, filterable: true },
      });

      component.$on('filter', handleFilter);

      // Open filter panel
      const filterButton = screen.getByText('Filter');
      await fireEvent.click(filterButton);

      // Add a filter
      const addFilterButton = screen.getByText('+ Add filter');
      await fireEvent.click(addFilterButton);

      expect(handleFilter).toHaveBeenCalled();
    });
  });

  describe('Search', () => {
    it('renders search input when searchable is true', () => {
      render(DataGrid, {
        props: { columns: testColumns, data: testData, searchable: true },
      });

      expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
    });

    it('emits search event on input', async () => {
      const handleSearch = vi.fn();
      const { component } = render(DataGrid, {
        props: { columns: testColumns, data: testData, searchable: true },
      });

      component.$on('search', handleSearch);

      const searchInput = screen.getByPlaceholderText('Search...');
      await fireEvent.input(searchInput, { target: { value: 'Alice' } });

      expect(handleSearch).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: { query: 'Alice' },
        })
      );
    });

    it('filters data based on search query', async () => {
      render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          searchable: true,
          paginated: false,
        },
      });

      const searchInput = screen.getByPlaceholderText('Search...');
      await fireEvent.input(searchInput, { target: { value: 'Alice' } });

      // Should show Alice
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();

      // Should not show Bob (different row)
      expect(screen.queryByText('Bob Johnson')).not.toBeInTheDocument();
    });
  });

  describe('Pagination', () => {
    it('renders pagination when paginated is true', () => {
      render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          paginated: true,
          pagination: { page: 1, pageSize: 2, total: 5 },
        },
      });

      expect(screen.getByText(/Showing 1 to 2 of 5 entries/)).toBeInTheDocument();
    });

    it('emits pageChange event when page is changed', async () => {
      const handlePageChange = vi.fn();
      const { component } = render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          paginated: true,
          pagination: { page: 1, pageSize: 2, total: 5 },
        },
      });

      component.$on('pageChange', handlePageChange);

      // Click next page button
      const nextButton = screen.getByLabelText('Next page');
      await fireEvent.click(nextButton);

      expect(handlePageChange).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({ page: 2 }),
        })
      );
    });

    it('emits pageChange event when page size is changed', async () => {
      const handlePageChange = vi.fn();
      const { component } = render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          paginated: true,
          pagination: { page: 1, pageSize: 10, total: 5 },
          pageSizes: [10, 25, 50],
        },
      });

      component.$on('pageChange', handlePageChange);

      const pageSizeSelect = screen.getByRole('combobox');
      await fireEvent.change(pageSizeSelect, { target: { value: '25' } });

      expect(handlePageChange).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({ pageSize: 25, page: 1 }),
        })
      );
    });
  });

  describe('Export', () => {
    it('renders export button when exportable is true', () => {
      render(DataGrid, {
        props: { columns: testColumns, data: testData, exportable: true },
      });

      expect(screen.getByText('Export')).toBeInTheDocument();
    });

    it('shows export dropdown when export button is clicked', async () => {
      render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          exportable: true,
          exportFormats: ['csv', 'xlsx', 'pdf'],
        },
      });

      const exportButton = screen.getByText('Export');
      await fireEvent.click(exportButton);

      expect(screen.getByText('CSV')).toBeInTheDocument();
      expect(screen.getByText('Excel (XLSX)')).toBeInTheDocument();
      expect(screen.getByText('PDF')).toBeInTheDocument();
    });
  });

  describe('Selection', () => {
    it('renders checkboxes when selectable is true', () => {
      render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          selectable: true,
          selectionMode: 'multiple',
        },
      });

      // Should have checkboxes (header + rows)
      const checkboxes = screen.getAllByRole('checkbox');
      expect(checkboxes.length).toBe(testData.length + 1);
    });

    it('emits select event when row is selected', async () => {
      const handleSelect = vi.fn();
      const { component } = render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          selectable: true,
          selectionMode: 'multiple',
        },
      });

      component.$on('select', handleSelect);

      // Click first row checkbox
      const checkboxes = screen.getAllByRole('checkbox');
      await fireEvent.click(checkboxes[1]); // Skip header checkbox

      expect(handleSelect).toHaveBeenCalled();
    });

    it('selects all rows when header checkbox is clicked', async () => {
      const handleSelect = vi.fn();
      const { component } = render(DataGrid, {
        props: {
          columns: testColumns,
          data: testData,
          selectable: true,
          selectionMode: 'multiple',
        },
      });

      component.$on('select', handleSelect);

      // Click header checkbox
      const checkboxes = screen.getAllByRole('checkbox');
      await fireEvent.click(checkboxes[0]);

      expect(handleSelect).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            keys: expect.arrayContaining([1, 2, 3, 4, 5]),
          }),
        })
      );
    });
  });

  describe('Row Click', () => {
    it('emits rowClick event when row is clicked', async () => {
      const handleRowClick = vi.fn();
      const { component } = render(DataGrid, {
        props: { columns: testColumns, data: testData },
      });

      component.$on('rowClick', handleRowClick);

      // Click on a row (click on a cell text)
      const aliceCell = screen.getByText('Alice Smith');
      await fireEvent.click(aliceCell.closest('tr')!);

      expect(handleRowClick).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            row: testData[0],
            index: 0,
          }),
        })
      );
    });
  });
});
