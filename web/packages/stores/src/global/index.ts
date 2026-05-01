/**
 * Global Stores - Export all global application stores
 * @packageDocumentation
 */

// API Provider Bridge (wires stores → api providers, breaking cyclic dep)
// export { initApiProviders } from './apiProviderBridge.js';

// Auth Store
export { authStore } from './auth.store.js';
export type {
  User,
  AuthTokens,
  Permission,
  Role,
  Session,
  AuthState,
  AuthError,
  LoginCredentials,
  RegisterData,
  AuthStoreActions,
} from './auth.store.js';

// Tenant Store
export { tenantStore, defaultSettings, defaultFeatures, defaultLimits } from './tenant.store.js';
export type {
  Tenant,
  TenantSettings,
  TenantFeatures,
  TenantLimits,
  TenantUsage,
  TenantState,
  TenantError,
} from './tenant.store.js';

// Theme Store
export { themeStore, defaultColors } from './theme.store.js';
export type {
  ThemeMode,
  ColorScheme,
  Density,
  FontSize,
  Radius,
  ThemeColors,
  ThemeState,
} from './theme.store.js';

// Notifications Store
export { notificationStore, toastStore } from './notifications.store.js';
export type {
  Notification,
  NotificationAction,
  Toast,
  NotificationPreferences,
  NotificationState,
  ToastState,
  NotificationError,
} from './notifications.store.js';

// Navigation Store
export { navigationStore, sidebarStore, historyStore } from './navigation.store.js';
export type {
  Module,
  ModuleItem,
  NavigationState,
  SidebarState,
  HistoryEntry,
  HistoryState,
} from './navigation.store.js';

// UI Store
export { modalStore, drawerStore, loadingStore, commandPaletteStore } from './ui.store.js';
export type {
  ModalItem,
  DrawerItem,
  LoadingItem,
  CommandItem,
  CommandGroup,
  ModalState,
  DrawerState,
  LoadingState,
  CommandPaletteState,
} from './ui.store.js';

// Session Store
export { sessionStore } from './session.store.js';
export type {
  OrganizationContext,
  SessionTenant,
  SessionCompany,
  SessionBranch,
  BranchType,
  BranchAddress,
  FiscalContext,
  SessionPreferences,
  SessionData,
  DeviceInfo,
  SessionState,
  SessionError,
} from './session.store.js';

// i18n Store
export {
  i18nStore,
  t,
  _,
  formatNumber,
  formatCurrency,
  formatDate,
  formatTime,
  formatRelativeTime,
  formatList,
  setLocale,
  loadTranslations,
  registerTranslations,
  initI18n,
  detectLocale,
  SUPPORTED_LOCALES,
  LOCALE_CONFIGS,
} from './i18nStore.js';
export type {
  SupportedLocale,
  LocaleConfig,
  TranslationValue,
  Translations,
  I18nState,
  PluralOptions,
} from './i18nStore.js';

// Feature Flags Store
export {
  featureFlagStore,
  isEnabled,
  getFlagValue,
  flag,
  setContext,
  setOverride,
  removeOverride,
  clearOverrides,
  fetchFlags,
  registerFlags,
  initFeatureFlags,
} from './featureFlagStore.js';
export type {
  FlagValue,
  TargetingOperator,
  TargetingRule,
  FeatureVariant,
  FeatureFlag,
  FlagContext,
  FeatureFlagState,
} from './featureFlagStore.js';

// Analytics Store
export {
  analyticsStore,
  trackPageView,
  trackEvent,
  identify,
  trackError,
  trackPerformance,
  startSession,
  endSession,
  setConsent,
  registerProvider,
  initAnalytics,
} from './analyticsStore.js';
export type {
  ConsentCategory,
  UserConsent,
  AnalyticsEvent,
  PageView,
  UserSession,
  PerformanceMetric,
  ErrorEvent,
  AnalyticsProvider,
  AnalyticsState,
} from './analyticsStore.js';

// Keyboard Shortcuts Store
export {
  keyboardStore,
  shortcutsByCategory,
  shortcutsForHelp,
  helpVisible,
  activeContext,
  initKeyboard,
  registerShortcut,
  unregisterShortcut,
  setKeyboardContext,
  toggleKeyboardHelp,
  COMMON_SHORTCUTS,
  VIM_SHORTCUTS,
} from './keyboardStore.js';
export type {
  ModifierKey,
  ShortcutContext,
  ShortcutDefinition,
  ShortcutGroup,
  KeyboardState,
} from './keyboardStore.js';

// PWA Store
export {
  pwaStore,
  canInstall,
  isOffline,
  hasUpdate,
  notificationsEnabled,
  initPwa,
  formatBytes,
} from './pwaStore.js';
export type {
  InstallState,
  ServiceWorkerState,
  PwaState,
  BeforeInstallPromptEvent,
} from './pwaStore.js';

// Audit Logging Store
export {
  auditStore,
  recentAuditEntries,
  pendingAuditCount,
  initAudit,
  audit,
  computeChanges,
} from './auditStore.js';
export type {
  AuditAction,
  AuditLevel,
  AuditCategory,
  AuditUser,
  AuditContext,
  DataChange,
  AuditEntry,
  AuditConfig,
  AuditState,
} from './auditStore.js';

// Session Timeout Store
export {
  sessionTimeoutStore,
  isSessionActive,
  isSessionWarning,
  sessionRemainingTime,
  startSessionTimeout,
  stopSessionTimeout,
  extendSession,
} from './sessionTimeoutStore.js';
export type {
  SessionTimeoutConfig,
  SessionTimeoutState,
} from './sessionTimeoutStore.js';

// History Store (Undo/Redo)
export {
  createHistoryStore,
  deriveHistoryState,
  useHistory,
} from './historyStore.js';
export type {
  HistoryEntry as UndoHistoryEntry,
  HistoryConfig as UndoHistoryConfig,
  HistoryState as UndoHistoryState,
} from './historyStore.js';

// Autosave Store
export {
  createAutosaveStore,
  deriveAutosaveState,
  useAutosave,
} from './autosaveStore.js';
export type {
  AutosaveStatus,
  AutosaveConfig,
  AutosaveState,
} from './autosaveStore.js';

// WebSocket Store
export {
  websocketStore,
  wsStatus,
  wsConnected,
  wsQueueSize,
  connectWebSocket,
  disconnectWebSocket,
  sendMessage,
} from './websocketStore.js';
export type {
  WebSocketStatus,
  WebSocketMessage,
  WebSocketConfig,
  WebSocketState,
} from './websocketStore.js';

// Permission Store
export {
  permissionStore,
  currentUser,
  isPermissionLoading,
  isPermissionInitialized,
  usePermission,
  createPermissionGuard,
  requirePermission,
  createResourcePermissions,
} from './permissionStore.js';
export type {
  PermissionAction,
  Permission as PermissionDef,
  Role as RoleDef,
  PermissionUser,
  PermissionConfig,
  PermissionState,
} from './permissionStore.js';
