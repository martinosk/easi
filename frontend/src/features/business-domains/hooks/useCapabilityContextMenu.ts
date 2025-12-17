import { useState, useCallback, useMemo } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';

interface CapabilityContextMenuState {
  x: number;
  y: number;
  capability: Capability;
}

interface UseCapabilityContextMenuProps {
  capabilities: Capability[];
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
    const parent = capabilities.find(c => c.id === current.parentId);
    if (!parent) break;
    current = parent;
  }

  return current.level === 'L1' ? current : undefined;
}

export function useCapabilityContextMenu({
  capabilities,
  dissociateCapability,
  refetch,
  selectedCapabilities,
  setSelectedCapabilities,
}: UseCapabilityContextMenuProps) {
  const [contextMenu, setContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [capabilityToDelete, setCapabilityToDelete] = useState<Capability | null>(null);
  const [capabilitiesToDelete, setCapabilitiesToDelete] = useState<Capability[]>([]);

  const handleCapabilityContextMenu = useCallback((capability: Capability, event: React.MouseEvent) => {
    event.preventDefault();
    setContextMenu({ x: event.clientX, y: event.clientY, capability });
  }, []);

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  const handleRemoveFromDomain = useCallback(async () => {
    if (!contextMenu) return;

    const isContextCapabilitySelected = selectedCapabilities.has(contextMenu.capability.id);
    const capabilitiesToDissociate = (selectedCapabilities.size > 0 && isContextCapabilitySelected)
      ? Array.from(selectedCapabilities)
          .map(id => capabilities.find(c => c.id === id))
          .filter((c): c is Capability => c !== undefined)
          .map(c => findL1Ancestor(c, capabilities))
          .filter((c): c is Capability => c !== undefined)
      : [contextMenu.capability].map(c => findL1Ancestor(c, capabilities)).filter((c): c is Capability => c !== undefined);

    const uniqueL1s = Array.from(new Map(capabilitiesToDissociate.map(c => [c.id, c])).values());

    await Promise.all(uniqueL1s.map(l1 => dissociateCapability(l1)));
    await refetch();
    setSelectedCapabilities(new Set());
    closeContextMenu();
  }, [contextMenu, capabilities, dissociateCapability, refetch, closeContextMenu, selectedCapabilities, setSelectedCapabilities]);

  const handleDeleteFromModel = useCallback(() => {
    if (!contextMenu) return;

    const isContextCapabilitySelected = selectedCapabilities.has(contextMenu.capability.id);
    const capabilitiesToDeleteList = (selectedCapabilities.size > 0 && isContextCapabilitySelected)
      ? Array.from(selectedCapabilities)
          .map(id => capabilities.find(c => c.id === id))
          .filter((c): c is Capability => c !== undefined)
          .map(c => findL1Ancestor(c, capabilities))
          .filter((c): c is Capability => c !== undefined)
      : [contextMenu.capability].map(c => findL1Ancestor(c, capabilities)).filter((c): c is Capability => c !== undefined);

    const uniqueL1s = Array.from(new Map(capabilitiesToDeleteList.map(c => [c.id, c])).values());

    if (uniqueL1s.length > 0) {
      setCapabilityToDelete(uniqueL1s[0]);
      setCapabilitiesToDelete(uniqueL1s);
    }
    closeContextMenu();
  }, [contextMenu, capabilities, closeContextMenu, selectedCapabilities]);

  const handleDeleteConfirm = useCallback(async () => {
    await refetch();
    setSelectedCapabilities(new Set());
    setCapabilityToDelete(null);
    setCapabilitiesToDelete([]);
  }, [refetch, setSelectedCapabilities]);

  const contextMenuItems: ContextMenuItem[] = useMemo(() => [
    { label: 'Remove from Business Domain', onClick: handleRemoveFromDomain },
    { label: 'Delete from Model', onClick: handleDeleteFromModel, isDanger: true },
  ], [handleRemoveFromDomain, handleDeleteFromModel]);

  return {
    contextMenu,
    capabilityToDelete,
    capabilitiesToDelete,
    handleCapabilityContextMenu,
    closeContextMenu,
    contextMenuItems,
    handleDeleteConfirm,
    setCapabilityToDelete,
  };
}
