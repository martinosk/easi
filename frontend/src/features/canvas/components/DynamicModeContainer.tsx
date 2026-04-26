import toast from 'react-hot-toast';
import { useCallback, useMemo } from 'react';
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
import { canEdit } from '../../../utils/hateoas';
import { useSaveDynamicDraft } from '../hooks/useSaveDynamicDraft';
import { useDynamicSnapshot } from '../hooks/useDynamicSnapshot';
import type { EntityRef } from '../utils/dynamicMode';
import { DynamicModeToolbar } from './DynamicModeToolbar';

type Position = { x: number; y: number };

function additionsWithPositions(additions: EntityRef[], positions: Record<string, Position>) {
  return additions.map((e) => ({
    ...e,
    x: positions[e.id]?.x ?? 0,
    y: positions[e.id]?.y ?? 0,
  }));
}

function positionDeltasAsList(deltas: Record<string, Position>, entities: EntityRef[]) {
  return Object.entries(deltas).map(([id, pos]) => {
    const ent = entities.find((e) => e.id === id);
    return { id, type: ent?.type ?? 'component', ...pos };
  });
}

type SaveResult = Awaited<ReturnType<ReturnType<typeof useSaveDynamicDraft>['save']>>;

function notifySaveResult(result: SaveResult) {
  if (result.failures.length === 0) {
    toast.success(`Saved ${result.successCount} change${result.successCount === 1 ? '' : 's'}`);
    return;
  }
  const failureMsg = result.failures.map((f) => `${f.operation} ${f.entity.id}`).join(', ');
  toast.error(`Saved ${result.successCount}, failed: ${failureMsg}`);
}

interface DraftSummary {
  dirty: boolean;
  additions: EntityRef[];
  removals: EntityRef[];
  positionDeltas: Record<string, Position>;
  totalChanges: number;
}

function shouldRenderToolbar(
  editable: boolean,
  currentView: unknown,
  currentViewId: unknown,
  enabled: boolean,
): boolean {
  return editable && Boolean(currentView) && Boolean(currentViewId) && enabled;
}

function useDraftSummary(): DraftSummary {
  const dynamicOriginal = useAppStore((s) => s.dynamicOriginal);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const dynamicPositions = useAppStore((s) => s.dynamicPositions);

  return useMemo(() => {
    const state = { dynamicOriginal, dynamicEntities, dynamicPositions };
    const adds = selectDynamicAdditions(state);
    const rems = selectDynamicRemovals(state);
    const pos = selectDynamicPositionDeltas(state);
    return {
      dirty: selectDynamicDirty(state),
      additions: adds,
      removals: rems,
      positionDeltas: pos,
      totalChanges: adds.length + rems.length + Object.keys(pos).length,
    };
  }, [dynamicOriginal, dynamicEntities, dynamicPositions]);
}

export function DynamicModeContainer() {
  useDynamicSnapshot();

  const queryClient = useQueryClient();
  const { currentView, currentViewId } = useCurrentView();
  const editable = canEdit(currentView);

  const enabled = useAppStore((s) => s.dynamicEnabled);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const dynamicPositions = useAppStore((s) => s.dynamicPositions);
  const enterDynamicMode = useAppStore((s) => s.enterDynamicMode);
  const resetDraft = useAppStore((s) => s.resetDraft);

  const { save, isSaving } = useSaveDynamicDraft();
  const { dirty, additions, removals, positionDeltas, totalChanges } = useDraftSummary();

  const handleSave = useCallback(async () => {
    if (!currentViewId || !currentView) return;
    const result = await save({
      viewId: currentViewId,
      additions: additionsWithPositions(additions, dynamicPositions),
      removals,
      positionDeltas: positionDeltasAsList(positionDeltas, dynamicEntities),
    });

    invalidateFor(queryClient, viewsMutationEffects.updateDetail(currentViewId));
    notifySaveResult(result);

    if (result.failures.length === 0) {
      enterDynamicMode(
        { entities: [...dynamicEntities], positions: { ...dynamicPositions } },
        currentViewId,
      );
    }
  }, [
    currentViewId,
    currentView,
    additions,
    removals,
    positionDeltas,
    dynamicPositions,
    dynamicEntities,
    save,
    queryClient,
    enterDynamicMode,
  ]);

  const handleSaveClick = useCallback(() => {
    void handleSave();
  }, [handleSave]);

  if (!shouldRenderToolbar(editable, currentView, currentViewId, enabled)) return null;

  return (
    <DynamicModeToolbar
      dirty={dirty}
      isSaving={isSaving}
      saveLabel={totalChanges > 0 ? `Save view (${totalChanges})` : 'Save view'}
      onSave={handleSaveClick}
      onDiscard={resetDraft}
    />
  );
}
