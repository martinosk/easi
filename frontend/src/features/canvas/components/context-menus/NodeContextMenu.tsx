import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import { useCurrentView } from '../../../views/hooks/useCurrentView';
import { useRemoveComponentFromView, useRemoveCapabilityFromView, useRemoveOriginEntityFromView } from '../../../views/hooks/useViews';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';
import type { OriginEntityType } from '../../../../components/canvas';
import type { ViewId } from '../../../../api/types';
import { toCapabilityId, toComponentId } from '../../../../api/types';
import { hasLink } from '../../../../utils/hateoas';

export type NodeDeleteTarget = {
  type: 'component-from-model' | 'capability-from-model' | 'origin-entity-from-model';
  id: string;
  name: string;
  originEntityType?: OriginEntityType;
};

interface NodeContextMenuProps {
  menu: NodeContextMenuType | null;
  onClose: () => void;
  onRequestDelete: (target: NodeDeleteTarget) => void;
}

interface MenuItemBuilderContext {
  menu: NodeContextMenuType;
  canRemoveFromView: boolean;
  canDeleteFromModel: boolean;
  currentViewId: ViewId | null;
  onRequestDelete: (target: NodeDeleteTarget) => void;
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

function buildViewElementItems(ctx: MenuItemBuilderContext, config: ViewElementConfig): ContextMenuItem[] {
  const items: ContextMenuItem[] = [];

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
        });
        ctx.onClose();
      },
      isDanger: true,
      ariaLabel: `Delete ${config.entityLabel} from entire model`,
    });
  }

  return items;
}

function buildOriginEntityItems(ctx: MenuItemBuilderContext, config: ViewElementConfig): ContextMenuItem[] {
  const items: ContextMenuItem[] = [];

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

export const NodeContextMenu = ({ menu, onClose, onRequestDelete }: NodeContextMenuProps) => {
  const { currentViewId } = useCurrentView();
  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();
  const removeOriginEntityFromViewMutation = useRemoveOriginEntityFromView();

  if (!menu) return null;

  const canRemoveFromView = hasLink({ _links: menu.viewElementLinks }, 'x-remove');
  const canDeleteFromModel = hasLink({ _links: menu.modelLinks }, 'delete');

  const removeFromViewHandlers: Record<NodeContextMenuType['nodeType'], (id: string) => void> = {
    capability: (id) => currentViewId && removeCapabilityFromViewMutation.mutate({
      viewId: currentViewId,
      capabilityId: toCapabilityId(id),
    }),
    component: (id) => currentViewId && removeComponentFromViewMutation.mutate({
      viewId: currentViewId,
      componentId: toComponentId(id),
    }),
    originEntity: (originEntityId) => currentViewId && removeOriginEntityFromViewMutation.mutate({
      viewId: currentViewId,
      originEntityId,
    }),
  };

  const ctx: MenuItemBuilderContext = {
    menu,
    canRemoveFromView,
    canDeleteFromModel,
    currentViewId,
    onRequestDelete,
    onClose,
    removeFromView: removeFromViewHandlers[menu.nodeType],
  };

  const viewElementConfig = viewElementConfigs[menu.nodeType];
  const items = menu.nodeType === 'originEntity'
    ? buildOriginEntityItems(ctx, viewElementConfig)
    : buildViewElementItems(ctx, viewElementConfig);
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
