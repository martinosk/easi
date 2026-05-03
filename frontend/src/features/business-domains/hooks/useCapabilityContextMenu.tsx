import { useCallback, useMemo, useState } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import {
  type ContextMenuItem,
  FolderMinusIcon,
  TrashIcon,
  UserPlusIcon,
} from '../../../components/shared/ContextMenu';
import { hasLink } from '../../../utils/hateoas';

interface CapabilityContextMenuState {
  x: number;
  y: number;
  capability: Capability;
}

interface UseCapabilityContextMenuProps {
  capabilities: Capability[];
  domainCapabilities: Capability[];
  dissociateCapability: (capability: Capability) => Promise<void>;
  refetch: () => Promise<void>;
  selectedCapabilities: Set<CapabilityId>;
  setSelectedCapabilities: (selected: Set<CapabilityId>) => void;
}

function findL1Ancestor(capability: Capability, capabilities: Capability[]): Capability | undefined {
  let current = capability;
  const seen = new Set<CapabilityId>();

  while (current.parentId && !seen.has(current.id)) {
    seen.add(current.id);
    const parent = capabilities.find((c) => c.id === current.parentId);
    if (!parent) break;
    current = parent;
  }

  return current.level === 'L1' ? current : undefined;
}

function getTargetL1Capabilities(
  contextCapability: Capability,
  selectedCapabilities: Set<CapabilityId>,
  capabilities: Capability[],
): Capability[] {
  const isContextSelected = selectedCapabilities.has(contextCapability.id);
  const targetCapabilities =
    selectedCapabilities.size > 0 && isContextSelected
      ? Array.from(selectedCapabilities)
          .map((id) => capabilities.find((c) => c.id === id))
          .filter((c): c is Capability => c !== undefined)
      : [contextCapability];

  const l1Ancestors = targetCapabilities
    .map((c) => findL1Ancestor(c, capabilities))
    .filter((c): c is Capability => c !== undefined);

  return Array.from(new Map(l1Ancestors.map((c) => [c.id, c])).values());
}

function useCapabilityPermissions(
  contextMenu: CapabilityContextMenuState | null,
  targetL1s: Capability[],
  domainCapabilities: Capability[],
) {
  const canRemoveFromDomain = useMemo(() => {
    if (!contextMenu || targetL1s.length === 0) return false;
    return targetL1s.every((l1) => {
      const domainCap = domainCapabilities.find((c) => c.id === l1.id);
      return domainCap && hasLink(domainCap, 'x-remove-from-domain');
    });
  }, [contextMenu, targetL1s, domainCapabilities]);

  const canDeleteFromModel = useMemo(() => {
    if (!contextMenu || targetL1s.length === 0) return false;
    return targetL1s.every((l1) => hasLink(l1, 'delete'));
  }, [contextMenu, targetL1s]);

  return { canRemoveFromDomain, canDeleteFromModel };
}

function buildMenuItems(
  contextMenu: CapabilityContextMenuState | null,
  permissions: { canRemoveFromDomain: boolean; canDeleteFromModel: boolean },
  actions: { handleRemoveFromDomain: () => void; handleDeleteFromModel: () => void; handleInviteToEdit: () => void },
): ContextMenuItem[] {
  const items: ContextMenuItem[] = [];
  if (permissions.canRemoveFromDomain) {
    items.push({
      label: 'Remove from Business Domain',
      description: 'Detach from this domain (keep in model)',
      icon: <FolderMinusIcon />,
      onClick: actions.handleRemoveFromDomain,
    });
  }
  if (permissions.canDeleteFromModel) {
    items.push({
      label: 'Delete from Model',
      description: 'Permanently remove this capability',
      icon: <TrashIcon />,
      onClick: actions.handleDeleteFromModel,
      isDanger: true,
    });
  }
  if (contextMenu?.capability && hasLink(contextMenu.capability, 'x-edit-grants')) {
    items.unshift({
      label: 'Invite to Edit...',
      description: 'Grant another user edit access',
      icon: <UserPlusIcon />,
      onClick: actions.handleInviteToEdit,
    });
  }
  return items;
}

export function useCapabilityContextMenu({
  capabilities,
  domainCapabilities,
  dissociateCapability,
  refetch,
  selectedCapabilities,
  setSelectedCapabilities,
}: UseCapabilityContextMenuProps) {
  const [contextMenu, setContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [capabilityToDelete, setCapabilityToDelete] = useState<Capability | null>(null);
  const [capabilitiesToDelete, setCapabilitiesToDelete] = useState<Capability[]>([]);
  const [capabilityToInvite, setCapabilityToInvite] = useState<Capability | null>(null);

  const handleCapabilityContextMenu = useCallback((capability: Capability, event: React.MouseEvent) => {
    event.preventDefault();
    setContextMenu({ x: event.clientX, y: event.clientY, capability });
  }, []);

  const closeContextMenu = useCallback(() => setContextMenu(null), []);
  const clearSelection = useCallback(() => setSelectedCapabilities(new Set()), [setSelectedCapabilities]);

  const targetL1s = useMemo(() => {
    if (!contextMenu) return [];
    return getTargetL1Capabilities(contextMenu.capability, selectedCapabilities, capabilities);
  }, [contextMenu, selectedCapabilities, capabilities]);

  const handleRemoveFromDomain = useCallback(async () => {
    if (targetL1s.length === 0) {
      closeContextMenu();
      return;
    }
    const domainL1s = targetL1s
      .map((l1) => domainCapabilities.find((c) => c.id === l1.id))
      .filter((c): c is Capability => c !== undefined);
    await Promise.all(domainL1s.map((l1) => dissociateCapability(l1)));
    await refetch();
    clearSelection();
    closeContextMenu();
  }, [targetL1s, domainCapabilities, dissociateCapability, refetch, clearSelection, closeContextMenu]);

  const handleDeleteFromModel = useCallback(() => {
    if (targetL1s.length > 0) {
      setCapabilityToDelete(targetL1s[0]);
      setCapabilitiesToDelete(targetL1s);
    }
    closeContextMenu();
  }, [targetL1s, closeContextMenu]);

  const handleDeleteConfirm = useCallback(async () => {
    await refetch();
    clearSelection();
    setCapabilityToDelete(null);
    setCapabilitiesToDelete([]);
  }, [refetch, clearSelection]);

  const handleInviteToEdit = useCallback(() => {
    if (!contextMenu) return;
    setCapabilityToInvite(contextMenu.capability);
    closeContextMenu();
  }, [contextMenu, closeContextMenu]);

  const { canRemoveFromDomain, canDeleteFromModel } = useCapabilityPermissions(
    contextMenu,
    targetL1s,
    domainCapabilities,
  );
  const contextMenuItems = useMemo(
    () =>
      buildMenuItems(
        contextMenu,
        { canRemoveFromDomain, canDeleteFromModel },
        { handleRemoveFromDomain, handleDeleteFromModel, handleInviteToEdit },
      ),
    [
      contextMenu,
      canRemoveFromDomain,
      canDeleteFromModel,
      handleRemoveFromDomain,
      handleDeleteFromModel,
      handleInviteToEdit,
    ],
  );

  return {
    contextMenu,
    capabilityToDelete,
    capabilitiesToDelete,
    capabilityToInvite,
    handleCapabilityContextMenu,
    closeContextMenu,
    contextMenuItems,
    handleDeleteConfirm,
    setCapabilityToDelete,
    setCapabilityToInvite,
  };
}
