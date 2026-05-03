import {
  ContextMenu,
  type ContextMenuItem,
  EyeOffIcon,
  TrashIcon,
} from '../../../../components/shared/ContextMenu';
import type { MultiSelectAction, MultiSelectMenuState } from '../../hooks/useMultiSelectContextMenu';
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

const ACTION_META: Record<BulkOperationType, { icon: ContextMenuItem['icon']; description: string }> = {
  removeFromView: { icon: <EyeOffIcon />, description: 'Hide selected; keep in the model' },
  deleteFromModel: { icon: <TrashIcon />, description: 'Permanently remove selected items' },
};

export const MultiSelectContextMenu = ({ menu, onClose, onRequestBulkOperation }: MultiSelectContextMenuProps) => {
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
          nodes: menu.selectedNodes,
        });
        onClose();
      },
    };
  });

  if (items.length === 0) return null;

  const title = `${menu.selectedNodes.length} selected`;
  return <ContextMenu x={menu.x} y={menu.y} items={items} title={title} onClose={onClose} />;
};
