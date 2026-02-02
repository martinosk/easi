import { ContextMenu, type ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { TreeMultiSelectMenuState, TreeBulkAction } from '../hooks/useTreeMultiSelectMenu';
import type { TreeSelectedItem } from '../hooks/useTreeMultiSelect';

export interface TreeBulkOperationRequest {
  type: TreeBulkAction['type'];
  items: TreeSelectedItem[];
}

interface TreeMultiSelectContextMenuProps {
  menu: TreeMultiSelectMenuState | null;
  onClose: () => void;
  onRequestBulkOperation: (request: TreeBulkOperationRequest) => void;
}

export const TreeMultiSelectContextMenu = ({
  menu,
  onClose,
  onRequestBulkOperation,
}: TreeMultiSelectContextMenuProps) => {
  if (!menu) return null;

  const items: ContextMenuItem[] = menu.actions.map((action) => ({
    label: action.label,
    isDanger: action.isDanger,
    onClick: () => {
      onRequestBulkOperation({
        type: action.type,
        items: menu.items,
      });
      onClose();
    },
  }));

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
