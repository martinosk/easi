import { useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';
import { canEdit } from '../../../utils/hateoas';
import type { DynamicModeSnapshot } from '../../../store/slices/dynamicModeSlice';
import type { EntityRef } from '../utils/dynamicMode';
import type { View, ViewId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { useCurrentView } from '../../views/hooks/useCurrentView';

type Position = { x: number; y: number };

export function buildSnapshotFromView(
  view: View,
  layoutPositions: Record<string, Position>,
): DynamicModeSnapshot {
  const entities: EntityRef[] = [];
  const positions: Record<string, Position> = {};

  for (const c of view.components) {
    entities.push({ id: c.componentId, type: 'component' });
    positions[c.componentId] = layoutPositions[c.componentId] ?? { x: c.x, y: c.y };
  }
  for (const cap of view.capabilities ?? []) {
    entities.push({ id: cap.capabilityId, type: 'capability' });
    positions[cap.capabilityId] = layoutPositions[cap.capabilityId] ?? { x: cap.x, y: cap.y };
  }
  for (const oe of view.originEntities ?? []) {
    entities.push({ id: oe.originEntityId, type: 'originEntity' });
    positions[oe.originEntityId] = { x: oe.x, y: oe.y };
  }
  return { entities, positions };
}

interface DynamicSnapshotInputs {
  currentView: View | null;
  currentViewId: ViewId | null;
  enabled: boolean;
  dynamicViewId: string | null;
  layoutPositions: Record<string, Position>;
}

function shouldExit(inputs: DynamicSnapshotInputs): boolean {
  return !inputs.currentView || !inputs.currentViewId || !canEdit(inputs.currentView);
}

function isReadyForSnapshot(inputs: DynamicSnapshotInputs): inputs is DynamicSnapshotInputs & {
  currentView: View;
  currentViewId: ViewId;
} {
  if (!inputs.currentView || !inputs.currentViewId) return false;
  return inputs.dynamicViewId !== inputs.currentViewId;
}

export function useDynamicSnapshot(): void {
  const { currentView, currentViewId } = useCurrentView();
  const { positions: layoutPositions } = useCanvasLayoutContext();
  const enabled = useAppStore((s) => s.dynamicEnabled);
  const dynamicViewId = useAppStore((s) => s.dynamicViewId);
  const enterDynamicMode = useAppStore((s) => s.enterDynamicMode);
  const exitDynamicMode = useAppStore((s) => s.exitDynamicMode);

  useEffect(() => {
    const inputs: DynamicSnapshotInputs = {
      currentView,
      currentViewId,
      enabled,
      dynamicViewId,
      layoutPositions,
    };
    if (shouldExit(inputs)) {
      if (enabled) exitDynamicMode();
      return;
    }
    if (isReadyForSnapshot(inputs)) {
      enterDynamicMode(
        buildSnapshotFromView(inputs.currentView, layoutPositions),
        inputs.currentViewId,
      );
    }
  }, [
    currentView,
    currentViewId,
    enabled,
    dynamicViewId,
    layoutPositions,
    enterDynamicMode,
    exitDynamicMode,
  ]);
}
