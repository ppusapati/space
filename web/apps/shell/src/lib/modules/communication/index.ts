/**
 * Communication domain module.
 *
 * Surfaces every standalone List RPC across the 2 active communication
 * services (chat / currency). The third service in the directory tree
 * (i18n) is intentionally NOT wired — its
 * core/communication/i18n/api/v1/localization/localizationconnect/localizationconnect.go
 * is a stub returning `http.NotFoundHandler()` (not yet implemented).
 *
 * Currently NOT wired (sub-resources requiring parent_id):
 *   - ChatService/ListMessages (needs conversation_id)
 *   - ChatService/ListParticipants (needs conversation_id)
 *   - ChatService/ListReactions (needs message_id)
 *   - ChatService/ListAttachments (needs message_id)
 *   - CurrencyService/ListExchangeRates (needs base_currency or pair)
 *
 * Each of these returns a clean HTTP 400 + `code: invalid_argument`
 * (handlers correctly route through `errors.ToConnectError`) — so the
 * defect class Audit #4 fixed in masters and that DEFER.HANDLER-ERROR-
 * MAPPING-SWEEP tracks for filestorage/queue/sla does NOT apply here.
 *
 * Wire-shape note for currency:
 *   The 3 currency Lists (ListCurrencies / ListConversionLogs /
 *   ListRateAlerts) all carry an embedded `pagination` field of type
 *   `packages.api.v1.pagination.Pagination` — but that type is the
 *   REQUEST shape, not a response with totalCount. The proto declares
 *   only PageOffset/PageSize/Sort/Fields. No totalCount is returned.
 *   ListPage's loader falls back to `rows.length` when responseTotalKey
 *   resolves to undefined, so we simply omit responseTotalKey for
 *   these entities. If currency proto is later refactored to use
 *   `pagination.PaginationResponse` (which has totalCount), update
 *   the registry to set responseTotalKey: 'pagination.totalCount'.
 *
 * Wire-shape note for chat:
 *   chat/ListConversations response uses int64 totalCount (not int32
 *   like most other services). JSON encoding handles the width fine;
 *   captured here in case a future native-decoder change cares.
 */
import type { DomainModule } from '../index.js';

export const communication: DomainModule = {
  id: 'communication',
  label: 'Communication',
  entities: [
    // ---- chat ----
    {
      slug: 'conversations',
      label: 'Conversations',
      formId: 'chat_room',
      listEndpoint: '/communication.chat.api.v1.ChatService/ListConversations',
      responseRowsKey: 'conversations',
      responseTotalKey: 'totalCount',
      columns: ['type', 'title', 'description', 'lastMessageAt', 'isMuted', 'isArchived'],
    },
    // ---- currency ----
    {
      slug: 'currencies',
      label: 'Currencies',
      formId: 'email_configuration',
      listEndpoint: '/communication.currency.api.v1.CurrencyService/ListCurrencies',
      responseRowsKey: 'currencies',
      // No totalCount in proto — loader falls back to rows.length.
      columns: ['code', 'name', 'symbol', 'decimalPlaces', 'status', 'isBaseCurrency'],
    },
    {
      slug: 'conversion-logs',
      label: 'Conversion Logs',
      formId: 'email_configuration',
      listEndpoint: '/communication.currency.api.v1.CurrencyService/ListConversionLogs',
      responseRowsKey: 'logs',
      // No totalCount in proto — loader falls back to rows.length.
      columns: ['fromCurrencyCode', 'toCurrencyCode', 'originalAmount', 'convertedAmount', 'rateUsed', 'convertedBy'],
    },
    {
      slug: 'rate-alerts',
      label: 'Rate Alerts',
      formId: 'email_configuration',
      listEndpoint: '/communication.currency.api.v1.CurrencyService/ListRateAlerts',
      responseRowsKey: 'alerts',
      // No totalCount in proto — loader falls back to rows.length.
      columns: ['baseCurrencyCode', 'targetCurrencyCode', 'thresholdRate', 'alertType', 'thresholdPercent', 'isActive'],
    },
  ],
};
