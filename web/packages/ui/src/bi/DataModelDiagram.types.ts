export interface DatasetNode {
  id: string;
  name: string;
  module: string;
  fieldCount: number;
  x: number; // position
  y: number;
  fields?: { id: string; label: string; role: string }[];
}

export interface RelationshipEdge {
  id: string;
  sourceDatasetId: string;
  sourceFieldId: string;
  targetDatasetId: string;
  targetFieldId: string;
  joinType: string;
  cardinality: string;
}
