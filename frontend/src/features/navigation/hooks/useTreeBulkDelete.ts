import { useState, useCallback, useMemo } from 'react';
import type { Component, Capability } from '../../../api/types';
import { toAcquiredEntityId, toVendorId, toInternalTeamId } from '../../../api/types';
import { useDeleteComponent, useComponents } from '../../components/hooks/useComponents';
import { useDeleteCapability, useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useDeleteAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteVendor } from '../../origin-entities/hooks/useVendors';
import { useDeleteInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import type { TreeSelectedItem, TreeItemType } from './useTreeMultiSelect';

export interface TreeBulkOperationResult {
  succeeded: string[];
  failed: { name: string; error: string }[];
}

async function executeBulkDelete(
  items: TreeSelectedItem[],
  deleteItem: (item: TreeSelectedItem) => Promise<void>
): Promise<TreeBulkOperationResult> {
  const succeeded: string[] = [];
  const failed: TreeBulkOperationResult['failed'] = [];

  for (const item of items) {
    try {
      await deleteItem(item);
      succeeded.push(item.name);
    } catch (error) {
      failed.push({
        name: item.name,
        error: error instanceof Error ? error.message : 'Unknown error',
      });
      break;
    }
  }

  return { succeeded, failed };
}

export function useTreeBulkDelete() {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();

  const deleteComponentMutation = useDeleteComponent();
  const deleteCapabilityMutation = useDeleteCapability();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();

  const [bulkItems, setBulkItems] = useState<TreeSelectedItem[] | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<TreeBulkOperationResult | null>(null);

  const deleteItem = useCallback(
    async (item: TreeSelectedItem): Promise<void> => {
      const handlers: Record<TreeItemType, () => Promise<void>> = {
        component: () => {
          const component = components.find((c: Component) => c.id === item.id);
          if (!component) throw new Error(`Component not found: ${item.name}`);
          return deleteComponentMutation.mutateAsync(component);
        },
        capability: () => {
          const capability = capabilities.find((c: Capability) => c.id === item.id);
          if (!capability) throw new Error(`Capability not found: ${item.name}`);
          return deleteCapabilityMutation.mutateAsync({ capability });
        },
        acquired: () =>
          deleteAcquiredEntityMutation.mutateAsync({ id: toAcquiredEntityId(item.id), name: item.name }),
        vendor: () =>
          deleteVendorMutation.mutateAsync({ id: toVendorId(item.id), name: item.name }),
        team: () =>
          deleteInternalTeamMutation.mutateAsync({ id: toInternalTeamId(item.id), name: item.name }),
      };
      return handlers[item.type]();
    },
    [components, capabilities, deleteComponentMutation, deleteCapabilityMutation, deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation]
  );

  const requestBulkDelete = useCallback((items: TreeSelectedItem[]) => {
    setBulkItems(items);
    setResult(null);
  }, []);

  const handleConfirm = useCallback(async () => {
    if (!bulkItems) return;

    setIsExecuting(true);
    setResult(null);

    const operationResult = await executeBulkDelete(bulkItems, deleteItem);

    setIsExecuting(false);

    if (operationResult.failed.length > 0) {
      setResult(operationResult);
    } else {
      setBulkItems(null);
    }
  }, [bulkItems, deleteItem]);

  const handleCancel = useCallback(() => {
    setBulkItems(null);
    setResult(null);
  }, []);

  const itemNames = useMemo(
    () => bulkItems?.map((item) => item.name) ?? [],
    [bulkItems]
  );

  return {
    bulkItems,
    isExecuting,
    result,
    itemNames,
    requestBulkDelete,
    handleConfirm,
    handleCancel,
  };
}
