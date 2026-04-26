import { create } from 'zustand';
import {
  type CanvasCapabilityActions,
  type CanvasCapabilityState,
  createCanvasCapabilitySlice,
} from './slices/canvasCapabilitySlice';
import {
  createDynamicModeSlice,
  type DynamicModeActions,
  type DynamicModeState,
} from './slices/dynamicModeSlice';
import { createSelectionSlice, type SelectionActions, type SelectionState } from './slices/selectionSlice';
import { createViewportSlice, type ViewportActions, type ViewportSliceState } from './slices/viewportSlice';
import { createViewSlice, type ViewActions, type ViewState } from './slices/viewSlice';

export type AppStore = ViewState &
  ViewActions &
  ViewportSliceState &
  ViewportActions &
  SelectionState &
  SelectionActions &
  CanvasCapabilityState &
  CanvasCapabilityActions &
  DynamicModeState &
  DynamicModeActions;

export const useAppStore = create<AppStore>()((...args) => ({
  ...createViewSlice(...args),
  ...createViewportSlice(...args),
  ...createSelectionSlice(...args),
  ...createCanvasCapabilitySlice(...args),
  ...createDynamicModeSlice(...args),
}));

export type {
  CapabilityId,
  ComponentData,
  ComponentId,
  DependencyId,
  EdgeType,
  Position,
  RealizationId,
  RelationData,
  RelationId,
  RelationType,
  ViewId,
  ViewportState,
} from './types/storeTypes';
