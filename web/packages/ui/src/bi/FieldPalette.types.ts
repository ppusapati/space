export interface DatasetFieldItem {
  id: string;
  label: string;
  column_name: string;
  data_type: string; // string, integer, decimal, date, datetime, boolean, currency, percentage
  role: 'dimension' | 'measure' | 'attribute';
  group_name?: string;
  default_aggregate?: string;
  granularities?: string[];
  icon?: string;
  searchable?: boolean;
  filterable?: boolean;
  sortable?: boolean;
}
