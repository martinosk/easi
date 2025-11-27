import type {
  ComponentId,
  RelationId,
  ViewId,
  CapabilityId,
  CapabilityDependencyId,
  RealizationId,
} from '../../api/types';

export type {
  ComponentId,
  RelationId,
  ViewId,
  CapabilityId,
  RealizationId,
};
export type DependencyId = CapabilityDependencyId;

export type RelationType = 'Triggers' | 'Serves';
export type EdgeType = string;
export type LayoutDirection = string;

export interface Position {
  readonly x: number;
  readonly y: number;
}

export interface ViewportState extends Position {
  readonly zoom: number;
}

export interface ComponentData {
  readonly name: string;
  readonly description?: string;
}

export interface RelationData {
  readonly sourceComponentId: ComponentId;
  readonly targetComponentId: ComponentId;
  readonly relationType: RelationType;
  readonly name?: string;
  readonly description?: string;
}

export interface LoadingState {
  isLoading: boolean;
  error: string | null;
}
