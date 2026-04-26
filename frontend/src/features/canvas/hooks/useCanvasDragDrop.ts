import type { ReactFlowInstance } from '@xyflow/react';
import { useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { canEdit } from '../../../utils/hateoas';
import type { MultiDragPayload, TreeItemType } from '../../navigation/hooks/useTreeMultiSelect';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { EntityRef, EntityType } from '../utils/dynamicMode';

const TREE_TYPE_TO_ENTITY_TYPE: Record<TreeItemType, EntityType> = {
  component: 'component',
  capability: 'capability',
  acquired: 'originEntity',
  vendor: 'originEntity',
  team: 'originEntity',
};

const MULTI_DROP_OFFSET_Y = 100;

interface ViewPresenceCheck {
  componentIds: Set<string>;
  capabilityIds: Set<string>;
  originEntityIds: Set<string>;
}

function buildViewPresence(
  currentView: {
    components: { componentId: string }[];
    capabilities: { capabilityId: string }[];
    originEntities: { originEntityId: string }[];
  } | null,
): ViewPresenceCheck {
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

function canDropOnView(
  reactFlowInstance: ReactFlowInstance | null,
  currentViewId: string | null,
  currentView: Parameters<typeof canEdit>[0],
): boolean {
  return !!reactFlowInstance && !!currentViewId && canEdit(currentView);
}

function parseDropItems(
  dataTransfer: DataTransfer,
  presence: ViewPresenceCheck,
  draftPresence: Set<string>,
): { type: TreeItemType; id: string }[] {
  const filterUnique = (items: { type: TreeItemType; id: string }[]) =>
    items.filter((item) => !isItemInView(item, presence) && !draftPresence.has(item.id));

  const multiPayload = parseMultiDragPayload(dataTransfer);
  if (multiPayload) return filterUnique(multiPayload.items);
  const singleItem = parseSingleDragItem(dataTransfer);
  if (singleItem) return filterUnique([singleItem]);
  return [];
}

function applyDraftDrop(
  items: { type: TreeItemType; id: string }[],
  origin: { x: number; y: number },
  draftAddEntities: (refs: EntityRef[], positions: Record<string, { x: number; y: number }>) => void,
): void {
  if (items.length === 0) return;
  const refs: EntityRef[] = items.map((it) => ({ id: it.id, type: TREE_TYPE_TO_ENTITY_TYPE[it.type] }));
  const positions: Record<string, { x: number; y: number }> = {};
  items.forEach((it, i) => {
    positions[it.id] = { x: origin.x, y: origin.y + i * MULTI_DROP_OFFSET_Y };
  });
  draftAddEntities(refs, positions);
}

export const useCanvasDragDrop = (
  reactFlowInstance: ReactFlowInstance | null,
  _onComponentDrop?: (componentId: string, x: number, y: number) => void,
) => {
  const { currentViewId, currentView } = useCurrentView();
  const draftAddEntities = useAppStore((s) => s.draftAddEntities);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      if (!canDropOnView(reactFlowInstance, currentViewId, currentView)) return;

      const position = reactFlowInstance!.screenToFlowPosition({ x: event.clientX, y: event.clientY });
      const presence = buildViewPresence(currentView);
      const draftPresence = new Set(dynamicEntities.map((e) => e.id));
      const items = parseDropItems(event.dataTransfer, presence, draftPresence);
      if (items.length === 0) return;

      applyDraftDrop(items, position, draftAddEntities);
    },
    [reactFlowInstance, currentViewId, currentView, draftAddEntities, dynamicEntities],
  );

  return { onDragOver, onDrop };
};
