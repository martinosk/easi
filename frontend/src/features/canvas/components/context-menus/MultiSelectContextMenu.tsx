import { ContextMenu, type ContextMenuItem } from '../../../../components/shared/ContextMenu';
import type { MultiSelectMenuState, MultiSelectAction } from '../../hooks/useMultiSelectContextMenu';
import type { NodeContextMenu } from '../../hooks/useNodeContextMenu';

export type BulkOperationType = MultiSelectAction['type'];

export interface BulkOperationRequest {
  type: BulkOperationType;
  nodes: NodeContextMenu[];
}

interface MultiSelectContextMenuProps {
  menu: MultiSelectMenuState | null;
  onClose: () => void;
  onRequestBulkOperation: (request: BulkOperationRequest) => void;
}

export const MultiSelectContextMenu = ({
  menu,
  onClose,
  onRequestBulkOperation,
}: MultiSelectContextMenuProps) => {
  if (!menu) return null;

  const items: ContextMenuItem[] = menu.actions.map((action) => ({
    label: action.label,
    isDanger: action.isDanger,
    onClick: () => {
      onRequestBulkOperation({
        type: action.type,
        nodes: menu.selectedNodes,
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
