import { Readable } from 'svelte/store';
import { SelectionState } from '../types/index.js';
export interface UseSelectionOptions<T> {
    initialSelection?: T[];
    getKey?: (item: T) => string | number;
    multiSelect?: boolean;
    maxSelection?: number;
    onChange?: (selection: SelectionState<T>) => void;
}
export interface UseSelectionReturn<T> {
    selection: Readable<SelectionState<T>>;
    selectedItems: Readable<T[]>;
    selectedKeys: Readable<Set<string | number>>;
    isAllSelected: Readable<boolean>;
    isIndeterminate: Readable<boolean>;
    selectedCount: Readable<number>;
    select: (item: T) => void;
    deselect: (item: T) => void;
    toggle: (item: T) => void;
    selectAll: (items: T[]) => void;
    deselectAll: () => void;
    selectRange: (items: T[], startIndex: number, endIndex: number) => void;
    setSelection: (items: T[]) => void;
    isSelected: (item: T) => boolean;
    getSelectedKeys: () => Set<string | number>;
    getSelectedItems: () => T[];
    reset: () => void;
}
export declare function useSelection<T>(options?: UseSelectionOptions<T>): UseSelectionReturn<T>;
export interface UseRowSelectionOptions<T> {
    getRowKey: (row: T) => string | number;
    multiSelect?: boolean;
    onChange?: (selectedRows: T[]) => void;
}
export declare function useRowSelection<T>(options: UseRowSelectionOptions<T>): UseSelectionReturn<T>;
export interface UseCheckboxGroupOptions<T> {
    options: T[];
    initialSelected?: T[];
    getValue?: (option: T) => string | number;
    maxSelected?: number;
    onChange?: (selected: T[]) => void;
}
export declare function useCheckboxGroup<T>(options: UseCheckboxGroupOptions<T>): {
    options: T[];
    toggleAll: () => void;
    selection: Readable<SelectionState<T>>;
    selectedItems: Readable<T[]>;
    selectedKeys: Readable<Set<string | number>>;
    isAllSelected: Readable<boolean>;
    isIndeterminate: Readable<boolean>;
    selectedCount: Readable<number>;
    select: (item: T) => void;
    deselect: (item: T) => void;
    toggle: (item: T) => void;
    selectAll: (items: T[]) => void;
    deselectAll: () => void;
    selectRange: (items: T[], startIndex: number, endIndex: number) => void;
    setSelection: (items: T[]) => void;
    isSelected: (item: T) => boolean;
    getSelectedKeys: () => Set<string | number>;
    getSelectedItems: () => T[];
    reset: () => void;
};
//# sourceMappingURL=useSelection.d.ts.map