import { useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { ArtifactType } from '../../edit-grants/types';
import { copyToClipboard, generateDomainShareUrl } from '../../../utils/clipboard';
import { hasLink } from '../../../utils/hateoas';

interface DomainContextMenuState {
  x: number;
  y: number;
  domain: BusinessDomain;
}

export interface DomainInviteTarget {
  id: string;
  artifactType: ArtifactType;
}

interface UseDomainContextMenuProps {
  onEdit: (domain: BusinessDomain) => void;
  onDelete: (domain: BusinessDomain) => void;
}

export function useDomainContextMenu({ onEdit, onDelete }: UseDomainContextMenuProps) {
  const [contextMenu, setContextMenu] = useState<DomainContextMenuState | null>(null);
  const [domainToInvite, setDomainToInvite] = useState<DomainInviteTarget | null>(null);

  const handleContextMenu = (e: React.MouseEvent, domain: BusinessDomain) => {
    setContextMenu({ x: e.clientX, y: e.clientY, domain });
  };

  const getContextMenuItems = (menu: DomainContextMenuState): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [];

    if (hasLink(menu.domain, 'x-edit-grants')) {
      items.push({
        label: 'Invite to Edit',
        onClick: () => {
          setDomainToInvite({ id: menu.domain.id, artifactType: 'domain' });
        },
      });
    }

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
    domainToInvite,
    setDomainToInvite,
  };
}
