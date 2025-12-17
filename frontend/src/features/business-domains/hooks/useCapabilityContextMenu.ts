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
    const parent = capabilities.find(c => c.id === current.parentId);
    if (!parent) break;
    current = parent;
  }

  return current.level === 'L1' ? current : undefined;
}

function getTargetL1Capabilities(
  contextCapability: Capability,
  selectedCapabilities: Set<CapabilityId>,
  capabilities: Capability[]
): Capability[] {
  const isContextSelected = selectedCapabilities.has(contextCapability.id);
  const targetCapabilities = (selectedCapabilities.size > 0 && isContextSelected)
    ? Array.from(selectedCapabilities)
        .map(id => capabilities.find(c => c.id === id))
        .filter((c): c is Capability => c !== undefined)
    : [contextCapability];

  const l1Ancestors = targetCapabilities
    .map(c => findL1Ancestor(c, capabilities))
    .filter((c): c is Capability => c !== undefined);

  return Array.from(new Map(l1Ancestors.map(c => [c.id, c])).values());
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

  const handleCapabilityContextMenu = useCallback((capability: Capability, event: React.MouseEvent) => {
    event.preventDefault();
    setContextMenu({ x: event.clientX, y: event.clientY, capability });
  }, []);

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  const handleRemoveFromDomain = useCallback(async () => {
    if (!contextMenu) return;

    const uniqueL1s = getTargetL1Capabilities(contextMenu.capability, selectedCapabilities, capabilities);
    const domainL1s = uniqueL1s
      .map(l1 => domainCapabilities.find(c => c.id === l1.id))
      .filter((c): c is Capability => c !== undefined);

    await Promise.all(domainL1s.map(l1 => dissociateCapability(l1)));
    await refetch();
    setSelectedCapabilities(new Set());
    closeContextMenu();
  }, [contextMenu, capabilities, domainCapabilities, dissociateCapability, refetch, closeContextMenu, selectedCapabilities, setSelectedCapabilities]);

  const handleDeleteFromModel = useCallback(() => {
    if (!contextMenu) return;

    const uniqueL1s = getTargetL1Capabilities(contextMenu.capability, selectedCapabilities, capabilities);
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
