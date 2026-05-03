import { ContextMenu, type ContextMenuItem, TrashIcon } from '../../../components/shared/ContextMenu';
import type { TreeSelectedItem } from '../hooks/useTreeMultiSelect';
import type { TreeBulkAction, TreeMultiSelectMenuState } from '../hooks/useTreeMultiSelectMenu';

export interface TreeBulkOperationRequest {
  type: TreeBulkAction['type'];
  items: TreeSelectedItem[];
}

interface TreeMultiSelectContextMenuProps {
  menu: TreeMultiSelectMenuState | null;
  onClose: () => void;
  onRequestBulkOperation: (request: TreeBulkOperationRequest) => void;
}

const ACTION_META: Record<TreeBulkAction['type'], { icon: ContextMenuItem['icon']; description: string }> = {
  deleteFromModel: { icon: <TrashIcon />, description: 'Permanently remove selected items' },
};

export const TreeMultiSelectContextMenu = ({
  menu,
  onClose,
  onRequestBulkOperation,
}: TreeMultiSelectContextMenuProps) => {
  if (!menu) return null;

  const items: ContextMenuItem[] = menu.actions.map((action) => {
    const meta = ACTION_META[action.type];
    return {
      label: action.label,
      isDanger: action.isDanger,
      icon: meta?.icon,
      description: meta?.description,
      onClick: () => {
        onRequestBulkOperation({
          type: action.type,
          items: menu.items,
        });
        onClose();
      },
    };
  });

  if (items.length === 0) return null;

  const title = `${menu.items.length} selected`;
  return <ContextMenu x={menu.x} y={menu.y} items={items} title={title} onClose={onClose} />;
};
