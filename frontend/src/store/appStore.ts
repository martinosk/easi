import { create } from 'zustand';
import { createComponentSlice, type ComponentState, type ComponentActions } from './slices/componentSlice';
import { createRelationSlice, type RelationState, type RelationActions } from './slices/relationSlice';
import { createViewSlice, type ViewState, type ViewActions } from './slices/viewSlice';
import { createViewportSlice, type ViewportSliceState, type ViewportActions } from './slices/viewportSlice';
import { createSelectionSlice, type SelectionState, type SelectionActions } from './slices/selectionSlice';
import { createLayoutSlice, type LayoutActions } from './slices/layoutSlice';

export type AppStore =
  & ComponentState
  & ComponentActions
  & RelationState
  & RelationActions
  & ViewState
  & ViewActions
  & ViewportSliceState
  & ViewportActions
  & SelectionState
  & SelectionActions
  & LayoutActions;

export const useAppStore = create<AppStore>()((...args) => ({
  ...createComponentSlice(...args),
  ...createRelationSlice(...args),
  ...createViewSlice(...args),
  ...createViewportSlice(...args),
  ...createSelectionSlice(...args),
  ...createLayoutSlice(...args),
}));

export type {
  ComponentId,
  RelationId,
  ViewId,
  RelationType,
  EdgeType,
  LayoutDirection,
  Position,
  ViewportState,
  ComponentData,
  RelationData,
} from './types/storeTypes';
