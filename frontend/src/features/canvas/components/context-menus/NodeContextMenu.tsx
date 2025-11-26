import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import { useAppStore } from '../../../../store/appStore';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';

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
  const removeComponentFromView = useAppStore((state) => state.removeComponentFromView);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  if (!menu) return null;

  const getContextMenuItems = (): ContextMenuItem[] => {
    if (menu.nodeType === 'capability') {
      return [
        {
          label: 'Remove from View',
          onClick: () => {
            removeCapabilityFromCanvas(menu.nodeId);
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
          removeComponentFromView(menu.nodeId);
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
