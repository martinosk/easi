import { useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import {
  type ContextMenuItem,
  PencilIcon,
  ShareIcon,
  TrashIcon,
  UserPlusIcon,
  UsersIcon,
} from '../../../components/shared/ContextMenu';
import { copyToClipboard, generateDomainShareUrl } from '../../../utils/clipboard';
import { getLink, hasLink } from '../../../utils/hateoas';
import type { ArtifactType } from '../../edit-grants/types';
import type { StewardsTarget } from '../../stewardship';

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
  const [domainForStewards, setDomainForStewards] = useState<StewardsTarget | null>(null);

  const handleContextMenu = (e: React.MouseEvent, domain: BusinessDomain) => {
    setContextMenu({ x: e.clientX, y: e.clientY, domain });
  };

  const getContextMenuItems = (menu: DomainContextMenuState): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [];

    if (hasLink(menu.domain, 'x-edit-grants')) {
      items.push({
        label: 'Invite to Edit...',
        description: 'Grant another user edit access',
        icon: <UserPlusIcon />,
        onClick: () => {
          setDomainToInvite({ id: menu.domain.id, artifactType: 'domain' });
        },
      });
    }

    const stewardshipsHref = getLink(menu.domain, 'x-stewardships');
    if (stewardshipsHref) {
      items.push({
        label: 'Stewards...',
        description: 'See who answers for each concern',
        icon: <UsersIcon />,
        onClick: () => {
          setDomainForStewards({ domainId: menu.domain.id, domainName: menu.domain.name, stewardshipsHref });
        },
      });
    }

    items.push({
      label: 'Share (copy URL)...',
      description: 'Copy a shareable link to clipboard',
      icon: <ShareIcon />,
      onClick: () => {
        const url = generateDomainShareUrl(menu.domain.id);
        copyToClipboard(url);
      },
    });

    if (menu.domain._links.update) {
      items.push({
        label: 'Edit',
        description: 'Open the edit panel',
        icon: <PencilIcon />,
        onClick: () => {
          onEdit(menu.domain);
        },
      });
    }

    const canDelete = menu.domain.capabilityCount === 0 && menu.domain._links.delete;
    if (canDelete) {
      items.push({
        label: 'Delete',
        description: 'Permanently delete this domain',
        icon: <TrashIcon />,
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
    domainForStewards,
    setDomainForStewards,
  };
}
