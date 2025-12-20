import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import type { EdgeContextMenu as EdgeContextMenuType } from '../../hooks/useContextMenu';

interface EdgeContextMenuProps {
  menu: EdgeContextMenuType | null;
  onClose: () => void;
  onRequestDelete: (target: {
    type: 'relation-from-model' | 'parent-relation' | 'realization';
    id: string;
    name: string;
    childId?: string;
  }) => void;
}

export const EdgeContextMenu = ({ menu, onClose, onRequestDelete }: EdgeContextMenuProps) => {
  if (!menu) return null;

  const getContextMenuItems = (): ContextMenuItem[] => {
    if (menu.edgeType === 'parent') {
      const edgeId = menu.edgeId;
      const parentIdStart = edgeId.indexOf('-') + 1;
      const parentIdEnd = edgeId.indexOf('-', parentIdStart + 36);
      const childId = edgeId.substring(parentIdEnd + 1);

      return [
        {
          label: 'Remove Parent Relationship',
          onClick: () => {
            onRequestDelete({
              type: 'parent-relation',
              id: menu.edgeId,
              name: 'Parent relationship',
              childId,
            });
            onClose();
          },
          isDanger: true,
        },
      ];
    }

    if (menu.edgeType === 'realization' && menu.realizationId) {
      if (menu.isInherited) {
        return [];
      }

      return [
        {
          label: 'Delete Realization',
          onClick: () => {
            onRequestDelete({
              type: 'realization',
              id: menu.realizationId!,
              name: menu.edgeName,
            });
            onClose();
          },
          isDanger: true,
          ariaLabel: 'Delete realization link',
        },
      ];
    }

    return [
      {
        label: 'Delete from Model',
        onClick: () => {
          onRequestDelete({
            type: 'relation-from-model',
            id: menu.edgeId,
            name: menu.edgeName,
          });
          onClose();
        },
        isDanger: true,
        ariaLabel: 'Delete relation from entire model',
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
