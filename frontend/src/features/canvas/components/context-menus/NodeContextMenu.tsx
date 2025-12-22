import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import { useCurrentView } from '../../../../hooks/useCurrentView';
import { useRemoveComponentFromView, useRemoveCapabilityFromView } from '../../../views/hooks/useViews';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';
import type { CapabilityId, ComponentId } from '../../../../api/types';

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

  const getContextMenuItems = (): ContextMenuItem[] => {
    if (menu.nodeType === 'capability') {
      return [
        {
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
        },
        {
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
        },
      ];
    }

    return [
      {
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
      },
      {
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
      },
    ];
  };

  return (
    <ContextMenu
      x={menu.x}
      y={menu.y}
      items={getContextMenuItems()}
      onClose={onClose}
    />
  );
};
