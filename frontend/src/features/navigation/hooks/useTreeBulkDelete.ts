import { useCallback, useMemo, useState } from 'react';
import type { Component } from '../../../api/types';
import { toAcquiredEntityId, toInternalTeamId, toVendorId } from '../../../api/types';
import { useSelectedCapabilityDeletion } from '../../capabilities/hooks/useCapabilityDeletionPlan';
import { useComponents, useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import { useDeleteVendor } from '../../origin-entities/hooks/useVendors';
import type { TreeItemType, TreeSelectedItem } from './useTreeMultiSelect';

export interface TreeBulkOperationResult {
  succeeded: string[];
  failed: { name: string; error: string }[];
}

async function executeBulkDelete(
  items: TreeSelectedItem[],
  deleteItem: (item: TreeSelectedItem) => Promise<void>,
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

function useTreeItemDelete(deleteCapability: (item: TreeSelectedItem) => Promise<void>) {
  const { data: components = [] } = useComponents();
  const deleteComponentMutation = useDeleteComponent();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();

  return useCallback(
    async (item: TreeSelectedItem): Promise<void> => {
      const handlers: Record<TreeItemType, () => Promise<void>> = {
        component: () => {
          const component = components.find((c: Component) => c.id === item.id);
          if (!component) throw new Error(`Component not found: ${item.name}`);
          return deleteComponentMutation.mutateAsync(component);
        },
        capability: () => deleteCapability(item),
        acquired: () => deleteAcquiredEntityMutation.mutateAsync({ id: toAcquiredEntityId(item.id), name: item.name }),
        vendor: () => deleteVendorMutation.mutateAsync({ id: toVendorId(item.id), name: item.name }),
        team: () => deleteInternalTeamMutation.mutateAsync({ id: toInternalTeamId(item.id), name: item.name }),
      };
      return handlers[item.type]();
    },
    [
      components,
      deleteCapability,
      deleteComponentMutation,
      deleteAcquiredEntityMutation,
      deleteVendorMutation,
      deleteInternalTeamMutation,
    ],
  );
}

export function useTreeBulkDelete() {
  const [bulkItems, setBulkItems] = useState<TreeSelectedItem[] | null>(null);
  const capabilityIds = useMemo(
    () => (bulkItems ?? []).filter((i) => i.type === 'capability').map((i) => i.id),
    [bulkItems],
  );
  const { plan: capabilityDeletionPlan, deleteCapability } = useSelectedCapabilityDeletion(capabilityIds);
  const deleteItem = useTreeItemDelete(deleteCapability);
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<TreeBulkOperationResult | null>(null);

  const requestBulkDelete = useCallback((items: TreeSelectedItem[]) => {
    setBulkItems(items);
    setResult(null);
  }, []);

  const handleConfirm = useCallback(async () => {
    if (!bulkItems) return;

    setIsExecuting(true);
    setResult(null);

    const operationResult = await executeBulkDelete(bulkItems, deleteItem);
    if (operationResult.failed.length > 0) capabilityDeletionPlan.retry();

    setIsExecuting(false);

    if (operationResult.failed.length > 0) {
      setResult(operationResult);
    } else {
      setBulkItems(null);
    }
  }, [bulkItems, deleteItem, capabilityDeletionPlan]);

  const handleCancel = useCallback(() => {
    setBulkItems(null);
    setResult(null);
  }, []);

  const itemNames = useMemo(() => bulkItems?.map((item) => item.name) ?? [], [bulkItems]);

  return {
    bulkItems,
    capabilityDeletionPlan,
    isExecuting,
    result,
    itemNames,
    requestBulkDelete,
    handleConfirm,
    handleCancel,
  };
}
