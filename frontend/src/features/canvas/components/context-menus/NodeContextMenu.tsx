import type { ViewId } from '../../../../api/types';
import { toCapabilityId, toComponentId } from '../../../../api/types';
import type { OriginEntityType } from '../../../../components/canvas';
import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import { hasLink } from '../../../../utils/hateoas';
import type { ArtifactType } from '../../../edit-grants/types';
import { useCurrentView } from '../../../views/hooks/useCurrentView';
import {
  useRemoveCapabilityFromView,
  useRemoveComponentFromView,
  useRemoveOriginEntityFromView,
} from '../../../views/hooks/useViews';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';

export type NodeDeleteTarget = {
  type: 'component-from-model' | 'capability-from-model' | 'origin-entity-from-model';
  id: string;
  name: string;
  originEntityType?: OriginEntityType;
};

export type InviteTarget = {
  id: string;
  artifactType: ArtifactType;
};

export type GenerateViewTarget = {
  entityRef: { id: string; type: 'component' | 'capability' | 'originEntity' };
  entityName: string;
};

interface NodeContextMenuProps {
  menu: NodeContextMenuType | null;
  onClose: () => void;
  onRequestDelete: (target: NodeDeleteTarget) => void;
  onRequestInviteToEdit?: (target: InviteTarget) => void;
  onRequestGenerateView?: (target: GenerateViewTarget) => void;
  canCreateView?: boolean;
}

interface MenuItemBuilderContext {
  menu: NodeContextMenuType;
  canRemoveFromView: boolean;
  canDeleteFromModel: boolean;
  canInviteToEdit: boolean;
  canCreateView: boolean;
  currentViewId: ViewId | null;
  onRequestDelete: (target: NodeDeleteTarget) => void;
  onRequestInviteToEdit?: (target: InviteTarget) => void;
  onRequestGenerateView?: (target: GenerateViewTarget) => void;
  onClose: () => void;
  removeFromView: (id: string) => void;
}

interface ViewElementConfig {
  deleteTargetType: NodeDeleteTarget['type'];
  entityLabel: string;
}

const viewElementConfigs: Record<string, ViewElementConfig> = {
  component: { deleteTargetType: 'component-from-model', entityLabel: 'component' },
  capability: { deleteTargetType: 'capability-from-model', entityLabel: 'capability' },
  originEntity: { deleteTargetType: 'origin-entity-from-model', entityLabel: 'origin entity' },
};

const nodeTypeToArtifactType: Record<NodeContextMenuType['nodeType'], ArtifactType> = {
  component: 'component',
  capability: 'capability',
  originEntity: 'vendor',
};

const originEntityTypeToArtifactType: Record<string, ArtifactType> = {
  acquired: 'acquired_entity',
  vendor: 'vendor',
  team: 'internal_team',
};

function resolveArtifactType(menu: NodeContextMenuType): ArtifactType {
  if (menu.nodeType === 'originEntity' && menu.originEntityType) {
    return originEntityTypeToArtifactType[menu.originEntityType] ?? nodeTypeToArtifactType[menu.nodeType];
  }
  return nodeTypeToArtifactType[menu.nodeType];
}

function buildInviteToEditItem(ctx: MenuItemBuilderContext): ContextMenuItem | null {
  if (!ctx.canInviteToEdit || !ctx.onRequestInviteToEdit) return null;
  const artifactType = resolveArtifactType(ctx.menu);
  return {
    label: 'Invite to Edit...',
    onClick: () => {
      ctx.onRequestInviteToEdit!({ id: ctx.menu.nodeId, artifactType });
      ctx.onClose();
    },
  };
}

const MENU_LABEL_MAX_NAME_LENGTH = 30;

function truncateName(name: string, maxLength: number): string {
  return name.length > maxLength ? name.slice(0, maxLength - 1) + '\u2026' : name;
}

function buildGenerateViewItem(ctx: MenuItemBuilderContext): ContextMenuItem | null {
  if (!ctx.canCreateView || !ctx.onRequestGenerateView) return null;
  const displayName = truncateName(ctx.menu.nodeName, MENU_LABEL_MAX_NAME_LENGTH);
  return {
    label: `Create dynamic view from ${displayName}`,
    onClick: () => {
      ctx.onRequestGenerateView!({
        entityRef: { id: ctx.menu.nodeId, type: ctx.menu.nodeType },
        entityName: ctx.menu.nodeName,
      });
      ctx.onClose();
    },
  };
}

function buildMenuItems(ctx: MenuItemBuilderContext, config: ViewElementConfig): ContextMenuItem[] {
  const items: ContextMenuItem[] = [];

  const generateViewItem = buildGenerateViewItem(ctx);
  if (generateViewItem) items.push(generateViewItem);

  const inviteItem = buildInviteToEditItem(ctx);
  if (inviteItem) items.push(inviteItem);

  if (ctx.canRemoveFromView) {
    items.push({
      label: 'Remove from View',
      onClick: () => {
        ctx.removeFromView(ctx.menu.nodeId);
        ctx.onClose();
      },
    });
  }

  if (ctx.canDeleteFromModel) {
    items.push({
      label: 'Delete from Model',
      onClick: () => {
        ctx.onRequestDelete({
          type: config.deleteTargetType,
          id: ctx.menu.nodeId,
          name: ctx.menu.nodeName,
          originEntityType: ctx.menu.originEntityType,
        });
        ctx.onClose();
      },
      isDanger: true,
      ariaLabel: `Delete ${config.entityLabel} from entire model`,
    });
  }

  return items;
}

export const NodeContextMenu = ({
  menu,
  onClose,
  onRequestDelete,
  onRequestInviteToEdit,
  onRequestGenerateView,
  canCreateView = false,
}: NodeContextMenuProps) => {
  const { currentViewId } = useCurrentView();
  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();
  const removeOriginEntityFromViewMutation = useRemoveOriginEntityFromView();

  if (!menu) return null;

  const canRemoveFromView = hasLink({ _links: menu.viewElementLinks }, 'x-remove');
  const canDeleteFromModel = hasLink({ _links: menu.modelLinks }, 'delete');
  const canInviteToEdit = hasLink({ _links: menu.modelLinks }, 'x-edit-grants');

  const removeFromViewHandlers: Record<NodeContextMenuType['nodeType'], (id: string) => void> = {
    capability: (id) =>
      currentViewId &&
      removeCapabilityFromViewMutation.mutate({
        viewId: currentViewId,
        capabilityId: toCapabilityId(id),
      }),
    component: (id) =>
      currentViewId &&
      removeComponentFromViewMutation.mutate({
        viewId: currentViewId,
        componentId: toComponentId(id),
      }),
    originEntity: (originEntityId) =>
      currentViewId &&
      removeOriginEntityFromViewMutation.mutate({
        viewId: currentViewId,
        originEntityId,
      }),
  };

  const ctx: MenuItemBuilderContext = {
    menu,
    canRemoveFromView,
    canDeleteFromModel,
    canInviteToEdit,
    canCreateView,
    currentViewId,
    onRequestDelete,
    onRequestInviteToEdit,
    onRequestGenerateView,
    onClose,
    removeFromView: removeFromViewHandlers[menu.nodeType],
  };

  const viewElementConfig = viewElementConfigs[menu.nodeType];
  const items = buildMenuItems(ctx, viewElementConfig);
  if (items.length === 0) return null;

  return <ContextMenu x={menu.x} y={menu.y} items={items} onClose={onClose} />;
};
