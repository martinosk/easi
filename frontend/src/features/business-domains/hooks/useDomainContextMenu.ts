import { useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import { copyToClipboard, generateDomainShareUrl } from '../../../utils/clipboard';

interface DomainContextMenuState {
  x: number;
  y: number;
  domain: BusinessDomain;
}

interface UseDomainContextMenuProps {
  onEdit: (domain: BusinessDomain) => void;
  onDelete: (domain: BusinessDomain) => void;
}

export function useDomainContextMenu({ onEdit, onDelete }: UseDomainContextMenuProps) {
  const [contextMenu, setContextMenu] = useState<DomainContextMenuState | null>(null);

  const handleContextMenu = (e: React.MouseEvent, domain: BusinessDomain) => {
    setContextMenu({ x: e.clientX, y: e.clientY, domain });
  };

  const getContextMenuItems = (menu: DomainContextMenuState): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [];

    items.push({
      label: 'Share (copy URL)...',
      onClick: () => {
        const url = generateDomainShareUrl(menu.domain.id);
        copyToClipboard(url);
      },
    });

    if (menu.domain._links.update) {
      items.push({
        label: 'Edit',
        onClick: () => {
          onEdit(menu.domain);
        },
      });
    }

    const canDelete = menu.domain.capabilityCount === 0 && menu.domain._links.delete;
    if (canDelete) {
      items.push({
        label: 'Delete',
        onClick: () => onDelete(menu.domain),
        isDanger: true,
      });
    }

    return items;
  };

  const closeContextMenu = () => setContextMenu(null);

  return {
    contextMenu,
    handleContextMenu,
    getContextMenuItems,
    closeContextMenu,
  };
}
