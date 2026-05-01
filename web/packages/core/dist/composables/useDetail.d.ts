import { Writable, Readable } from 'svelte/store';
import { DetailState, DetailMode, DetailConfig, DetailError, DetailSection, RelatedDataState, Action } from '../types/index.js';
export interface UseDetailOptions<TEntity, TActions extends string = string> {
    config: DetailConfig<TEntity>;
    sections?: DetailSection<TEntity>[];
    actions?: Record<TActions, Action<TEntity>>;
    fetchEntity?: (id: string | number) => Promise<TEntity>;
    saveEntity?: (entity: TEntity) => Promise<TEntity>;
    deleteEntity?: (entity: TEntity) => Promise<void>;
    onError?: (error: DetailError) => void;
    onSave?: (entity: TEntity) => void;
    onDelete?: (entity: TEntity) => void;
}
export interface UseDetailReturn<TEntity, TActions extends string = string> {
    entity: Writable<TEntity | null>;
    originalEntity: Writable<TEntity | null>;
    state: Writable<DetailState>;
    mode: Writable<DetailMode>;
    sections: Writable<DetailSection<TEntity>[]>;
    activeSectionId: Writable<string | undefined>;
    relatedData: Writable<Record<string, RelatedDataState<unknown>>>;
    validationErrors: Writable<Record<string, string>>;
    isDirty: Readable<boolean>;
    isValid: Readable<boolean>;
    changedFields: Readable<Partial<TEntity>>;
    load: (id: string | number) => Promise<void>;
    reload: () => Promise<void>;
    refresh: () => Promise<void>;
    setMode: (mode: DetailMode) => void;
    edit: () => void;
    view: () => void;
    create: (initialData?: Partial<TEntity>) => void;
    save: () => Promise<void>;
    delete: () => Promise<void>;
    duplicate: () => Promise<TEntity | null>;
    setFieldValue: <K extends keyof TEntity>(field: K, value: TEntity[K]) => void;
    resetField: <K extends keyof TEntity>(field: K) => void;
    resetAll: () => void;
    getChangedFields: () => Partial<TEntity>;
    validate: () => Promise<boolean>;
    validateField: <K extends keyof TEntity>(field: K) => Promise<string | null>;
    setFieldError: (field: string, error: string | null) => void;
    clearErrors: () => void;
    executeAction: (actionId: TActions) => Promise<void>;
    loadRelatedData: (key: string, loader: () => Promise<unknown>) => Promise<void>;
    refreshRelatedData: (key: string, loader: () => Promise<unknown>) => Promise<void>;
    setActiveSection: (sectionId: string) => void;
    toggleSection: (sectionId: string) => void;
}
export declare function useDetail<TEntity, TActions extends string = string>(options: UseDetailOptions<TEntity, TActions>): UseDetailReturn<TEntity, TActions>;
//# sourceMappingURL=useDetail.d.ts.map