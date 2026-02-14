import { useCallback } from 'react';
import type { ReactFlowInstance } from '@xyflow/react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAddCapabilityToView, useAddOriginEntityToView } from '../../views/hooks/useViews';
import { toCapabilityId, toComponentId, toViewId } from '../../../api/types';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { canEdit } from '../../../utils/hateoas';
import type { MultiDragPayload, TreeItemType } from '../../navigation/hooks/useTreeMultiSelect';

const MULTI_DROP_OFFSET_Y = 100;

interface ViewPresenceCheck {
  componentIds: Set<string>;
  capabilityIds: Set<string>;
  originEntityIds: Set<string>;
}

function buildViewPresence(currentView: { components: { componentId: string }[]; capabilities: { capabilityId: string }[]; originEntities: { originEntityId: string }[] } | null): ViewPresenceCheck {
  return {
    componentIds: new Set((currentView?.components ?? []).map((c) => c.componentId)),
    capabilityIds: new Set((currentView?.capabilities ?? []).map((c) => c.capabilityId)),
    originEntityIds: new Set((currentView?.originEntities ?? []).map((oe) => oe.originEntityId)),
  };
}

function isItemInView(item: { type: TreeItemType; id: string }, presence: ViewPresenceCheck): boolean {
  switch (item.type) {
    case 'component':
      return presence.componentIds.has(item.id);
    case 'capability':
      return presence.capabilityIds.has(item.id);
    case 'acquired':
    case 'vendor':
    case 'team':
      return presence.originEntityIds.has(item.id);
  }
}

const SINGLE_DRAG_KEYS: { key: string; type: TreeItemType }[] = [
  { key: 'componentId', type: 'component' },
  { key: 'capabilityId', type: 'capability' },
  { key: 'acquiredEntityId', type: 'acquired' },
  { key: 'vendorId', type: 'vendor' },
  { key: 'internalTeamId', type: 'team' },
];

function parseSingleDragItem(dataTransfer: DataTransfer): { type: TreeItemType; id: string } | null {
  for (const { key, type } of SINGLE_DRAG_KEYS) {
    const id = dataTransfer.getData(key);
    if (id) return { type, id };
  }
  return null;
}

function parseMultiDragPayload(dataTransfer: DataTransfer): MultiDragPayload | null {
  const raw = dataTransfer.getData('multiDragItems');
  if (!raw) return null;
  try {
    return JSON.parse(raw) as MultiDragPayload;
  } catch {
    return null;
  }
}

interface DropHandlers {
  onComponentDrop?: (componentId: string, x: number, y: number) => void;
  updateComponentPosition: (componentId: ReturnType<typeof toComponentId>, x: number, y: number) => Promise<void>;
  addCapability: (viewId: string, capId: ReturnType<typeof toCapabilityId>, x: number, y: number) => Promise<void>;
  updateCapabilityPosition: (capId: ReturnType<typeof toCapabilityId>, x: number, y: number) => Promise<void>;
  addOriginEntity: (viewId: string, originEntityId: string, x: number, y: number) => Promise<void>;
  currentViewId: string;
}

async function addItemToView(item: { type: TreeItemType; id: string }, x: number, y: number, handlers: DropHandlers): Promise<void> {
  switch (item.type) {
    case 'component':
      if (handlers.onComponentDrop) {
        await handlers.onComponentDrop(item.id, x, y);
        await handlers.updateComponentPosition(toComponentId(item.id), x, y);
      }
      break;
    case 'capability': {
      const capId = toCapabilityId(item.id);
      await handlers.addCapability(handlers.currentViewId, capId, x, y);
      await handlers.updateCapabilityPosition(capId, x, y);
      break;
    }
    case 'acquired':
    case 'vendor':
    case 'team':
      await handlers.addOriginEntity(handlers.currentViewId, item.id, x, y);
      break;
  }
}

function canDropOnView(
  reactFlowInstance: ReactFlowInstance | null,
  currentViewId: string | null,
  currentView: Parameters<typeof canEdit>[0]
): boolean {
  return !!reactFlowInstance && !!currentViewId && canEdit(currentView);
}

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

  const getHandlers = useCallback(
    (): DropHandlers => ({
      onComponentDrop,
      updateComponentPosition,
      addCapability: async (viewId, capId, x, y) => {
        await addCapabilityToViewMutation.mutateAsync({ viewId: toViewId(viewId), request: { capabilityId: capId, x, y } });
      },
      updateCapabilityPosition,
      addOriginEntity: async (viewId, originEntityId, x, y) => {
        await addOriginEntityToViewMutation.mutateAsync({ viewId: toViewId(viewId), request: { originEntityId, x, y } });
      },
      currentViewId: currentViewId!,
    }),
    [onComponentDrop, currentViewId, addCapabilityToViewMutation, addOriginEntityToViewMutation, updateComponentPosition, updateCapabilityPosition]
  );

  const onDrop = useCallback(
    async (event: React.DragEvent) => {
      event.preventDefault();

      if (!canDropOnView(reactFlowInstance, currentViewId, currentView)) return;

      const position = reactFlowInstance!.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const handlers = getHandlers();

      const multiPayload = parseMultiDragPayload(event.dataTransfer);
      if (multiPayload) {
        const presence = buildViewPresence(currentView);
        const itemsToAdd = multiPayload.items.filter((item) => !isItemInView(item, presence));
        for (let i = 0; i < itemsToAdd.length; i++) {
          await addItemToView(itemsToAdd[i], position.x, position.y + i * MULTI_DROP_OFFSET_Y, handlers);
        }
        return;
      }

      const singleItem = parseSingleDragItem(event.dataTransfer);
      if (singleItem) {
        await addItemToView(singleItem, position.x, position.y, handlers);
      }
    },
    [reactFlowInstance, currentViewId, currentView, getHandlers]
  );

  return { onDragOver, onDrop };
};
