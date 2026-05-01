/**
 * Entity Store Factory
 * Creates a reusable store for CRUD operations on entities
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// FACTORY
// ============================================================================
export function createEntityStore(config) {
    const { name, idField = 'id', defaultPageSize = 10, defaultSort = { column: null, direction: null }, fetchList: fetchListFn, fetchOne: fetchOneFn, create: createFn, update: updateFn, remove: removeFn, bulkRemove: bulkRemoveFn, } = config;
    // ============================================================================
    // INITIAL STATE
    // ============================================================================
    const initialState = {
        items: [],
        loadingState: 'idle',
        error: null,
        pagination: {
            page: 1,
            pageSize: defaultPageSize,
            total: 0,
            totalPages: 0,
            hasNext: false,
            hasPrevious: false,
        },
        sort: defaultSort,
        filters: {},
        search: '',
        selection: {
            selectedItems: [],
            selectedKeys: new Set(),
            isAllSelected: false,
            isIndeterminate: false,
        },
        current: null,
        currentLoading: false,
        currentError: null,
        lastFetched: null,
    };
    // ============================================================================
    // STORE
    // ============================================================================
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const items = derived(store, ($s) => $s.items);
    const isLoading = derived(store, ($s) => $s.loadingState === 'loading');
    const isEmpty = derived(store, ($s) => $s.items.length === 0 && $s.loadingState !== 'loading');
    const hasError = derived(store, ($s) => $s.error !== null);
    const pagination = derived(store, ($s) => $s.pagination);
    const sort = derived(store, ($s) => $s.sort);
    const filters = derived(store, ($s) => $s.filters);
    const selection = derived(store, ($s) => $s.selection);
    const current = derived(store, ($s) => $s.current);
    // ============================================================================
    // HELPERS
    // ============================================================================
    function getEntityId(entity) {
        return String(entity[idField]);
    }
    function updatePagination(total) {
        update((s) => {
            const totalPages = Math.ceil(total / s.pagination.pageSize);
            return {
                ...s,
                pagination: {
                    ...s.pagination,
                    total,
                    totalPages,
                    hasNext: s.pagination.page < totalPages,
                    hasPrevious: s.pagination.page > 1,
                },
            };
        });
    }
    // ============================================================================
    // LIST ACTIONS
    // ============================================================================
    async function fetchList() {
        if (!fetchListFn) {
            console.warn(`[${name}Store] fetchList not configured`);
            return;
        }
        update((s) => ({ ...s, loadingState: 'loading', error: null }));
        try {
            const state = get(store);
            const result = await fetchListFn({
                page: state.pagination.page,
                pageSize: state.pagination.pageSize,
                sort: state.sort,
                filters: state.filters,
                search: state.search,
            });
            update((s) => ({
                ...s,
                items: result.items,
                loadingState: 'success',
                lastFetched: new Date(),
            }));
            updatePagination(result.total);
        }
        catch (error) {
            update((s) => ({
                ...s,
                loadingState: 'error',
                error: {
                    code: 'FETCH_LIST_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to fetch list',
                },
            }));
        }
    }
    async function refresh() {
        update((s) => ({ ...s, loadingState: 'refreshing' }));
        await fetchList();
    }
    function setPage(page) {
        update((s) => ({
            ...s,
            pagination: { ...s.pagination, page },
        }));
        fetchList();
    }
    function setPageSize(size) {
        update((s) => ({
            ...s,
            pagination: { ...s.pagination, pageSize: size, page: 1 },
        }));
        fetchList();
    }
    function setSort(column, direction) {
        update((s) => ({
            ...s,
            sort: { column, direction },
            pagination: { ...s.pagination, page: 1 },
        }));
        fetchList();
    }
    function setFilters(newFilters) {
        update((s) => ({
            ...s,
            filters: { ...s.filters, ...newFilters },
            pagination: { ...s.pagination, page: 1 },
        }));
        fetchList();
    }
    function resetFilters() {
        update((s) => ({
            ...s,
            filters: {},
            search: '',
            pagination: { ...s.pagination, page: 1 },
        }));
        fetchList();
    }
    function setSearch(query) {
        update((s) => ({
            ...s,
            search: query,
            pagination: { ...s.pagination, page: 1 },
        }));
        fetchList();
    }
    // ============================================================================
    // SELECTION ACTIONS
    // ============================================================================
    function select(entity) {
        const id = getEntityId(entity);
        update((s) => {
            if (s.selection.selectedKeys.has(id))
                return s;
            const selectedKeys = new Set(s.selection.selectedKeys);
            selectedKeys.add(id);
            const selectedItems = [...s.selection.selectedItems, entity];
            return {
                ...s,
                selection: {
                    selectedItems,
                    selectedKeys,
                    isAllSelected: selectedKeys.size === s.items.length,
                    isIndeterminate: selectedKeys.size > 0 && selectedKeys.size < s.items.length,
                },
            };
        });
    }
    function deselect(entity) {
        const id = getEntityId(entity);
        update((s) => {
            if (!s.selection.selectedKeys.has(id))
                return s;
            const selectedKeys = new Set(s.selection.selectedKeys);
            selectedKeys.delete(id);
            const selectedItems = s.selection.selectedItems.filter((e) => getEntityId(e) !== id);
            return {
                ...s,
                selection: {
                    selectedItems,
                    selectedKeys,
                    isAllSelected: false,
                    isIndeterminate: selectedKeys.size > 0,
                },
            };
        });
    }
    function toggleSelect(entity) {
        const id = getEntityId(entity);
        const state = get(store);
        if (state.selection.selectedKeys.has(id)) {
            deselect(entity);
        }
        else {
            select(entity);
        }
    }
    function selectAll() {
        const state = get(store);
        update((s) => ({
            ...s,
            selection: {
                selectedItems: [...s.items],
                selectedKeys: new Set(s.items.map(getEntityId)),
                isAllSelected: true,
                isIndeterminate: false,
            },
        }));
    }
    function deselectAll() {
        update((s) => ({
            ...s,
            selection: {
                selectedItems: [],
                selectedKeys: new Set(),
                isAllSelected: false,
                isIndeterminate: false,
            },
        }));
    }
    function isSelected(entity) {
        const state = get(store);
        return state.selection.selectedKeys.has(getEntityId(entity));
    }
    // ============================================================================
    // CRUD ACTIONS
    // ============================================================================
    async function fetchOne(id) {
        if (!fetchOneFn) {
            console.warn(`[${name}Store] fetchOne not configured`);
            return null;
        }
        update((s) => ({ ...s, currentLoading: true, currentError: null }));
        try {
            const entity = await fetchOneFn(id);
            update((s) => ({ ...s, current: entity, currentLoading: false }));
            return entity;
        }
        catch (error) {
            update((s) => ({
                ...s,
                currentLoading: false,
                currentError: {
                    code: 'FETCH_ONE_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to fetch entity',
                },
            }));
            return null;
        }
    }
    async function create(data) {
        if (!createFn) {
            console.warn(`[${name}Store] create not configured`);
            return null;
        }
        try {
            const entity = await createFn(data);
            update((s) => ({
                ...s,
                items: [entity, ...s.items],
                pagination: { ...s.pagination, total: s.pagination.total + 1 },
            }));
            return entity;
        }
        catch (error) {
            update((s) => ({
                ...s,
                error: {
                    code: 'CREATE_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to create entity',
                },
            }));
            return null;
        }
    }
    async function updateEntity(id, data) {
        if (!updateFn) {
            console.warn(`[${name}Store] update not configured`);
            return null;
        }
        try {
            const entity = await updateFn(id, data);
            update((s) => ({
                ...s,
                items: s.items.map((e) => (getEntityId(e) === id ? entity : e)),
                current: s.current && getEntityId(s.current) === id ? entity : s.current,
            }));
            return entity;
        }
        catch (error) {
            update((s) => ({
                ...s,
                error: {
                    code: 'UPDATE_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to update entity',
                },
            }));
            return null;
        }
    }
    async function remove(id) {
        if (!removeFn) {
            console.warn(`[${name}Store] remove not configured`);
            return false;
        }
        try {
            await removeFn(id);
            update((s) => ({
                ...s,
                items: s.items.filter((e) => getEntityId(e) !== id),
                pagination: { ...s.pagination, total: s.pagination.total - 1 },
                current: s.current && getEntityId(s.current) === id ? null : s.current,
            }));
            return true;
        }
        catch (error) {
            update((s) => ({
                ...s,
                error: {
                    code: 'REMOVE_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to remove entity',
                },
            }));
            return false;
        }
    }
    async function bulkRemove(ids) {
        if (!bulkRemoveFn) {
            console.warn(`[${name}Store] bulkRemove not configured`);
            return false;
        }
        try {
            await bulkRemoveFn(ids);
            const idSet = new Set(ids);
            update((s) => ({
                ...s,
                items: s.items.filter((e) => !idSet.has(getEntityId(e))),
                pagination: { ...s.pagination, total: s.pagination.total - ids.length },
                selection: {
                    selectedItems: [],
                    selectedKeys: new Set(),
                    isAllSelected: false,
                    isIndeterminate: false,
                },
            }));
            return true;
        }
        catch (error) {
            update((s) => ({
                ...s,
                error: {
                    code: 'BULK_REMOVE_ERROR',
                    message: error instanceof Error ? error.message : 'Failed to remove entities',
                },
            }));
            return false;
        }
    }
    // ============================================================================
    // UTILITY
    // ============================================================================
    function getById(id) {
        const state = get(store);
        return state.items.find((e) => getEntityId(e) === id);
    }
    function setCurrent(entity) {
        update((s) => ({ ...s, current: entity }));
    }
    function reset() {
        set(initialState);
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        items,
        isLoading,
        isEmpty,
        hasError,
        pagination,
        sort,
        filters,
        selection,
        current,
        // List actions
        fetchList,
        refresh,
        setPage,
        setPageSize,
        setSort,
        setFilters,
        resetFilters,
        setSearch,
        // Selection actions
        select,
        deselect,
        toggleSelect,
        selectAll,
        deselectAll,
        isSelected,
        // CRUD actions
        fetchOne,
        create,
        update: updateEntity,
        remove,
        bulkRemove,
        // Utility
        getById,
        setCurrent,
        reset,
    };
}
