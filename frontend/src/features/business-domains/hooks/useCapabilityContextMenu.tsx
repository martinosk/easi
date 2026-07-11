import { useCallback, useMemo, useState } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import { type ContextMenuItem, FolderMinusIcon, TrashIcon, UserPlusIcon } from '../../../components/shared/ContextMenu';
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

function getTargetCapabilities(
  contextCapability: Capability,
  selectedCapabilities: Set<CapabilityId>,
  capabilities: Capability[],
): Capability[] {
  const isContextSelected = selectedCapabilities.has(contextCapability.id);
  if (selectedCapabilities.size === 0 || !isContextSelected) {
    return [contextCapability];
  }
  return Array.from(selectedCapabilities)
    .map((id) => capabilities.find((c) => c.id === id))
    .filter((c): c is Capability => c !== undefined);
}

function useCapabilityPermissions(
  contextMenu: CapabilityContextMenuState | null,
  targets: Capability[],
  domainCapabilities: Capability[],
) {
  const canRemoveFromDomain = useMemo(() => {
    if (!contextMenu || targets.length === 0) return false;
    return targets.every((target) => {
      const domainCap = domainCapabilities.find((c) => c.id === target.id);
      return Boolean(domainCap && hasLink(domainCap, 'x-remove-from-domain'));
    });
  }, [contextMenu, targets, domainCapabilities]);

  const canDeleteFromModel = useMemo(() => {
    if (!contextMenu || targets.length === 0) return false;
    return targets.every((target) => hasLink(target, 'delete'));
  }, [contextMenu, targets]);

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

  const targets = useMemo(() => {
    if (!contextMenu) return [];
    return getTargetCapabilities(contextMenu.capability, selectedCapabilities, capabilities);
  }, [contextMenu, selectedCapabilities, capabilities]);

  const handleRemoveFromDomain = useCallback(async () => {
    if (targets.length === 0) {
      closeContextMenu();
      return;
    }
    const removable = targets
      .map((target) => domainCapabilities.find((c) => c.id === target.id))
      .filter((c): c is Capability => c !== undefined);
    await Promise.all(removable.map((c) => dissociateCapability(c)));
    await refetch();
    clearSelection();
    closeContextMenu();
  }, [targets, domainCapabilities, dissociateCapability, refetch, clearSelection, closeContextMenu]);

  const handleDeleteFromModel = useCallback(() => {
    if (targets.length > 0) {
      setCapabilityToDelete(targets[0]);
      setCapabilitiesToDelete(targets);
    }
    closeContextMenu();
  }, [targets, closeContextMenu]);

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
    targets,
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
