import { useCallback } from 'react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAddCapabilityToView, useAddOriginEntityToView } from '../../views/hooks/useViews';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { canEdit } from '../../../utils/hateoas';
import { ORIGIN_ENTITY_PREFIXES } from '../utils/nodeFactory';

export const useCanvasDragDrop = (
  reactFlowInstance: ReactFlowInstance | null,
  onComponentDrop?: (componentId: string, x: number, y: number) => void
) => {
  const { currentViewId, currentView } = useCurrentView();
  const addCapabilityToViewMutation = useAddCapabilityToView();
  const addOriginEntityToViewMutation = useAddOriginEntityToView();
  const { updateComponentPosition, updateCapabilityPosition } = useCanvasLayoutContext();

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    async (event: React.DragEvent) => {
      event.preventDefault();

      const componentId = event.dataTransfer.getData('componentId');
      const capabilityId = event.dataTransfer.getData('capabilityId');
      const acquiredEntityId = event.dataTransfer.getData('acquiredEntityId');
      const vendorId = event.dataTransfer.getData('vendorId');
      const internalTeamId = event.dataTransfer.getData('internalTeamId');

      if (!reactFlowInstance || !currentViewId || !canEdit(currentView)) return;

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      if (componentId && onComponentDrop) {
        await onComponentDrop(componentId, position.x, position.y);
        await updateComponentPosition(toComponentId(componentId), position.x, position.y);
      } else if (capabilityId) {
        const capId = toCapabilityId(capabilityId);
        await addCapabilityToViewMutation.mutateAsync({
          viewId: currentViewId,
          request: {
            capabilityId: capId,
            x: position.x,
            y: position.y
          }
        });
        await updateCapabilityPosition(capId, position.x, position.y);
      } else if (acquiredEntityId) {
        const originEntityId = `${ORIGIN_ENTITY_PREFIXES.acquired}${acquiredEntityId}`;
        await addOriginEntityToViewMutation.mutateAsync({
          viewId: currentViewId,
          request: { originEntityId, x: position.x, y: position.y }
        });
      } else if (vendorId) {
        const originEntityId = `${ORIGIN_ENTITY_PREFIXES.vendor}${vendorId}`;
        await addOriginEntityToViewMutation.mutateAsync({
          viewId: currentViewId,
          request: { originEntityId, x: position.x, y: position.y }
        });
      } else if (internalTeamId) {
        const originEntityId = `${ORIGIN_ENTITY_PREFIXES.team}${internalTeamId}`;
        await addOriginEntityToViewMutation.mutateAsync({
          viewId: currentViewId,
          request: { originEntityId, x: position.x, y: position.y }
        });
      }
    },
    [onComponentDrop, reactFlowInstance, currentViewId, currentView, addCapabilityToViewMutation, addOriginEntityToViewMutation, updateComponentPosition, updateCapabilityPosition]
  );

  return { onDragOver, onDrop };
};
