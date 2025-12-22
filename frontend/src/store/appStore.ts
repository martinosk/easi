import { create } from 'zustand';
import { createViewSlice, type ViewState, type ViewActions } from './slices/viewSlice';
import { createViewportSlice, type ViewportSliceState, type ViewportActions } from './slices/viewportSlice';
import { createSelectionSlice, type SelectionState, type SelectionActions } from './slices/selectionSlice';
import { createCanvasCapabilitySlice, type CanvasCapabilityState, type CanvasCapabilityActions } from './slices/canvasCapabilitySlice';

export type AppStore =
  & ViewState
  & ViewActions
  & ViewportSliceState
  & ViewportActions
  & SelectionState
  & SelectionActions
  & CanvasCapabilityState
  & CanvasCapabilityActions;

export const useAppStore = create<AppStore>()((...args) => ({
  ...createViewSlice(...args),
  ...createViewportSlice(...args),
  ...createSelectionSlice(...args),
  ...createCanvasCapabilitySlice(...args),
}));

export type {
  ComponentId,
  RelationId,
  ViewId,
  CapabilityId,
  DependencyId,
  RealizationId,
  RelationType,
  EdgeType,
  Position,
  ViewportState,
  ComponentData,
  RelationData,
} from './types/storeTypes';
