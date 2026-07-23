import { useEffect, useRef } from 'react';
import type { View, ViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import type { DraftEntry, DynamicModeSnapshot } from '../../../store/slices/dynamicModeSlice';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { viewPositions } from '../utils/viewPositions';
import { entitiesFromView } from './useCanvasNodes';

export function buildSnapshotFromView(view: View): DynamicModeSnapshot {
  return { entities: entitiesFromView(view), positions: viewPositions(view) };
}

interface DynamicSnapshotInputs {
  currentView: View | null;
  currentViewId: ViewId | null;
  dynamicViewId: string | null;
  draftsByView: Record<string, DraftEntry>;
}

interface SnapshotActions {
  enterDynamicMode: (initial: DynamicModeSnapshot, viewId?: string | null) => void;
  exitDynamicMode: () => void;
  stashCurrentDraft: (viewId: string) => void;
  hydrateDraftForView: (viewId: string) => boolean;
}

function shouldExit(inputs: DynamicSnapshotInputs): boolean {
  return !inputs.currentView || !inputs.currentViewId || !canEdit(inputs.currentView);
}

function shouldRefresh(inputs: DynamicSnapshotInputs): boolean {
  return Boolean(inputs.currentView) && Boolean(inputs.currentViewId) && inputs.dynamicViewId !== inputs.currentViewId;
}

function shouldStashPrev(prevViewId: string | null, nextViewId: ViewId): boolean {
  if (!prevViewId) return false;
  return prevViewId !== nextViewId;
}

function applySwitch(inputs: DynamicSnapshotInputs, prevViewId: string | null, actions: SnapshotActions): void {
  const view = inputs.currentView as View;
  const viewId = inputs.currentViewId as ViewId;
  if (shouldStashPrev(prevViewId, viewId)) {
    actions.stashCurrentDraft(prevViewId as string);
  }
  if (actions.hydrateDraftForView(viewId)) return;
  if (inputs.draftsByView[viewId]) return;
  actions.enterDynamicMode(buildSnapshotFromView(view), viewId);
}

export function useDynamicSnapshot(): void {
  const { currentView, currentViewId } = useCurrentView();
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const draftsByView = useAppStore((s) => s.draftsByView);
  const enterDynamicMode = useAppStore((s) => s.enterDynamicMode);
  const exitDynamicMode = useAppStore((s) => s.exitDynamicMode);
  const stashCurrentDraft = useAppStore((s) => s.stashCurrentDraft);
  const hydrateDraftForView = useAppStore((s) => s.hydrateDraftForView);
  const prevViewIdRef = useRef<string | null>(dynamicViewId);

  useEffect(() => {
    const inputs: DynamicSnapshotInputs = {
      currentView,
      currentViewId,
      dynamicViewId,
      draftsByView,
    };
    if (shouldExit(inputs)) {
      if (dynamicViewId !== null) exitDynamicMode();
      prevViewIdRef.current = null;
      return;
    }
    if (shouldRefresh(inputs)) {
      applySwitch(inputs, prevViewIdRef.current, {
        enterDynamicMode,
        exitDynamicMode,
        stashCurrentDraft,
        hydrateDraftForView,
      });
      prevViewIdRef.current = inputs.currentViewId;
    }
  }, [
    currentView,
    currentViewId,
    dynamicViewId,
    draftsByView,
    enterDynamicMode,
    exitDynamicMode,
    stashCurrentDraft,
    hydrateDraftForView,
  ]);
}
