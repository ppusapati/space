export interface SlicerFilter {
  field_id: string;
  label: string;
  data_type: string;
  operator: string; // eq, in, between, gte, lte, contains
  value: unknown;
  values?: unknown[]; // for IN
  second_value?: unknown; // for BETWEEN
}
