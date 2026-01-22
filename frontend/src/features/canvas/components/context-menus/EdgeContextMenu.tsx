import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import type { EdgeContextMenu as EdgeContextMenuType } from '../../hooks/useContextMenu';
import type { CapabilityId, ComponentId, HATEOASLinks, OriginRelationshipId, OriginRelationshipType } from '../../../../api/types';
import { hasLink } from '../../../../utils/hateoas';

export type DeleteTarget = {
  type: 'relation-from-model' | 'parent-relation' | 'realization' | 'origin-relationship';
  id: string;
  name: string;
  childId?: string;
  capabilityId?: CapabilityId;
  componentId?: ComponentId;
  originRelationshipId?: OriginRelationshipId;
  originRelationshipType?: OriginRelationshipType;
  originEntityId?: string;
  _links?: HATEOASLinks;
};

interface EdgeContextMenuProps {
  menu: EdgeContextMenuType | null;
  onClose: () => void;
  onRequestDelete: (target: DeleteTarget) => void;
}

function extractChildIdFromParentEdge(edgeId: string): string {
  const parentIdStart = edgeId.indexOf('-') + 1;
  const parentIdEnd = edgeId.indexOf('-', parentIdStart + 36);
  return edgeId.substring(parentIdEnd + 1);
}

function buildParentEdgeItems(
  menu: EdgeContextMenuType,
  onRequestDelete: (target: DeleteTarget) => void,
  onClose: () => void
): ContextMenuItem[] {
  const childId = extractChildIdFromParentEdge(menu.edgeId);
  return [{
    label: 'Remove Parent Relationship',
    onClick: () => {
      onRequestDelete({ type: 'parent-relation', id: menu.edgeId, name: 'Parent relationship', childId });
      onClose();
    },
    isDanger: true,
  }];
}

function buildRealizationEdgeItems(
  menu: EdgeContextMenuType,
  canDelete: boolean,
  onRequestDelete: (target: DeleteTarget) => void,
  onClose: () => void
): ContextMenuItem[] {
  const hasRealizationId = menu.realizationId !== undefined;
  const isNotInherited = !menu.isInherited;
  const isDeleteAllowed = hasRealizationId && isNotInherited && canDelete;
  if (!isDeleteAllowed) return [];

  return [{
    label: 'Delete Realization',
    onClick: () => {
      onRequestDelete({
        type: 'realization',
        id: menu.realizationId!,
        name: menu.edgeName,
        capabilityId: menu.capabilityId,
        componentId: menu.componentId,
        _links: menu._links,
      });
      onClose();
    },
    isDanger: true,
    ariaLabel: 'Delete realization link',
  }];
}

function buildOriginRelationshipEdgeItems(
  menu: EdgeContextMenuType,
  canDelete: boolean,
  onRequestDelete: (target: DeleteTarget) => void,
  onClose: () => void
): ContextMenuItem[] {
  if (!menu.originRelationshipId || !canDelete) return [];

  return [{
    label: 'Delete Relationship',
    onClick: () => {
      onRequestDelete({
        type: 'origin-relationship',
        id: menu.originRelationshipId!,
        name: menu.edgeName,
        originRelationshipId: menu.originRelationshipId,
        originRelationshipType: menu.originRelationshipType,
        componentId: menu.componentId,
        originEntityId: menu.originEntityId,
        _links: menu._links,
      });
      onClose();
    },
    isDanger: true,
    ariaLabel: 'Delete origin relationship',
  }];
}

function buildRelationEdgeItems(
  menu: EdgeContextMenuType,
  canDelete: boolean,
  onRequestDelete: (target: DeleteTarget) => void,
  onClose: () => void
): ContextMenuItem[] {
  if (!canDelete) return [];

  return [{
    label: 'Delete from Model',
    onClick: () => {
      onRequestDelete({ type: 'relation-from-model', id: menu.edgeId, name: menu.edgeName });
      onClose();
    },
    isDanger: true,
    ariaLabel: 'Delete relation from entire model',
  }];
}

export const EdgeContextMenu = ({ menu, onClose, onRequestDelete }: EdgeContextMenuProps) => {
  if (!menu) return null;

  const canDelete = hasLink({ _links: menu._links }, 'delete');

  const edgeTypeHandlers: Record<EdgeContextMenuType['edgeType'], () => ContextMenuItem[]> = {
    parent: () => buildParentEdgeItems(menu, onRequestDelete, onClose),
    realization: () => buildRealizationEdgeItems(menu, canDelete, onRequestDelete, onClose),
    'origin-relationship': () => buildOriginRelationshipEdgeItems(menu, canDelete, onRequestDelete, onClose),
    relation: () => buildRelationEdgeItems(menu, canDelete, onRequestDelete, onClose),
  };

  const items = edgeTypeHandlers[menu.edgeType]();
  if (items.length === 0) return null;

  return (
    <ContextMenu
      x={menu.x}
      y={menu.y}
      items={items}
      onClose={onClose}
    />
  );
};
