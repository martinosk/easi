import toast from 'react-hot-toast';
import { useQueryClient } from '@tanstack/react-query';
import { useAppStore } from '../../../store/appStore';
import {
  selectDynamicAdditions,
  selectDynamicDirty,
  selectDynamicPositionDeltas,
  selectDynamicRemovals,
} from '../../../store/slices/dynamicModeSlice';
import { invalidateFor } from '../../../lib/invalidateFor';
import { viewsMutationEffects } from '../../views/mutationEffects';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useSaveDynamicDraft } from '../hooks/useSaveDynamicDraft';
import type { DynamicModeSnapshot } from '../../../store/slices/dynamicModeSlice';
import type { EntityRef } from '../utils/dynamicMode';
import { DynamicModeToolbar } from './DynamicModeToolbar';

function buildSnapshotFromView(view: NonNullable<ReturnType<typeof useCurrentView>['currentView']>): DynamicModeSnapshot {
  const entities: EntityRef[] = [];
  const positions: Record<string, { x: number; y: number }> = {};

  for (const c of view.components) {
    entities.push({ id: c.componentId, type: 'component' });
    positions[c.componentId] = { x: c.x, y: c.y };
  }
  for (const cap of view.capabilities) {
    entities.push({ id: cap.capabilityId, type: 'capability' });
    positions[cap.capabilityId] = { x: cap.x, y: cap.y };
  }
  for (const oe of view.originEntities) {
    entities.push({ id: oe.originEntityId, type: 'originEntity' });
    positions[oe.originEntityId] = { x: oe.x, y: oe.y };
  }

  return { entities, positions };
}

export function DynamicModeContainer() {
  const queryClient = useQueryClient();
  const { currentView, currentViewId } = useCurrentView();
  const enabled = useAppStore((s) => s.dynamicEnabled);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const enterDynamicMode = useAppStore((s) => s.enterDynamicMode);
  const exitDynamicMode = useAppStore((s) => s.exitDynamicMode);

  const dirty = useAppStore(selectDynamicDirty);
  const additions = useAppStore(selectDynamicAdditions);
  const removals = useAppStore(selectDynamicRemovals);
  const positionDeltas = useAppStore(selectDynamicPositionDeltas);
  const positionsState = useAppStore((s) => s.dynamicPositions);

  const { save, isSaving } = useSaveDynamicDraft();

  const handleEnable = () => {
    if (!currentView) return;
    enterDynamicMode(buildSnapshotFromView(currentView));
  };

  const handleSave = async () => {
    if (!currentViewId) return;
    const additionsWithPos = additions.map((e) => ({
      ...e,
      x: positionsState[e.id]?.x ?? 0,
      y: positionsState[e.id]?.y ?? 0,
    }));
    const positionsAsList = Object.entries(positionDeltas).map(([id, pos]) => {
      const ent = dynamicEntities.find((e) => e.id === id);
      return { id, type: ent?.type ?? 'component', ...pos };
    });
    const result = await save({ viewId: currentViewId, additions: additionsWithPos, removals, positionDeltas: positionsAsList });

    invalidateFor(queryClient, viewsMutationEffects.updateDetail(currentViewId));

    if (result.failures.length === 0) {
      toast.success(`Saved ${result.successCount} change${result.successCount === 1 ? '' : 's'}`);
      exitDynamicMode();
    } else {
      const failureMsg = result.failures.map((f) => `${f.operation} ${f.entity.id}`).join(', ');
      toast.error(`Saved ${result.successCount}, failed: ${failureMsg}`);
    }
  };

  const handleDiscard = () => {
    exitDynamicMode();
    if (currentViewId) {
      invalidateFor(queryClient, viewsMutationEffects.updateDetail(currentViewId));
    }
  };

  const totalChanges = additions.length + removals.length + Object.keys(positionDeltas).length;
  const saveLabel = `Save view (${totalChanges})`;

  if (!currentView || !currentViewId) return null;

  return (
    <div style={{ position: 'absolute', top: 12, right: 12, zIndex: 10 }}>
      <DynamicModeToolbar
        enabled={enabled}
        dirty={dirty}
        isSaving={isSaving}
        saveLabel={saveLabel}
        onEnable={handleEnable}
        onSave={() => void handleSave()}
        onDiscard={handleDiscard}
      />
    </div>
  );
}
