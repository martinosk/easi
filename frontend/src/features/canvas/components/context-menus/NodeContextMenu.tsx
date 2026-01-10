import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import { useCurrentView } from '../../../../hooks/useCurrentView';
import { useRemoveComponentFromView, useRemoveCapabilityFromView } from '../../../views/hooks/useViews';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';
import type { CapabilityId, ComponentId } from '../../../../api/types';
import { hasLink } from '../../../../utils/hateoas';

interface NodeContextMenuProps {
  menu: NodeContextMenuType | null;
  onClose: () => void;
  onRequestDelete: (target: {
    type: 'component-from-model' | 'capability-from-model';
    id: string;
    name: string;
  }) => void;
}

export const NodeContextMenu = ({ menu, onClose, onRequestDelete }: NodeContextMenuProps) => {
  const { currentViewId } = useCurrentView();
  const removeComponentFromViewMutation = useRemoveComponentFromView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();

  if (!menu) return null;

  const canRemoveFromView = hasLink({ _links: menu.viewElementLinks }, 'x-remove');
  const canDeleteFromModel = hasLink({ _links: menu.modelLinks }, 'delete');

  const getContextMenuItems = (): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [];

    if (menu.nodeType === 'capability') {
      if (canRemoveFromView) {
        items.push({
          label: 'Remove from View',
          onClick: () => {
            if (currentViewId) {
              removeCapabilityFromViewMutation.mutate({
                viewId: currentViewId,
                capabilityId: menu.nodeId as CapabilityId
              });
            }
            onClose();
          },
        });
      }

      if (canDeleteFromModel) {
        items.push({
          label: 'Delete from Model',
          onClick: () => {
            onRequestDelete({
              type: 'capability-from-model',
              id: menu.nodeId,
              name: menu.nodeName,
            });
            onClose();
          },
          isDanger: true,
          ariaLabel: 'Delete capability from entire model',
        });
      }

      return items;
    }

    if (canRemoveFromView) {
      items.push({
        label: 'Remove from View',
        onClick: () => {
          if (currentViewId) {
            removeComponentFromViewMutation.mutate({
              viewId: currentViewId,
              componentId: menu.nodeId as ComponentId
            });
          }
          onClose();
        },
      });
    }

    if (canDeleteFromModel) {
      items.push({
        label: 'Delete from Model',
        onClick: () => {
          onRequestDelete({
            type: 'component-from-model',
            id: menu.nodeId,
            name: menu.nodeName,
          });
          onClose();
        },
        isDanger: true,
        ariaLabel: 'Delete component from entire model',
      });
    }

    return items;
  };

  const items = getContextMenuItems();
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
