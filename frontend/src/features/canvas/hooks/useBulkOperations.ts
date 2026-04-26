import { useCallback, useMemo, useState } from 'react';
import type { AcquiredEntityId, Capability, Component, InternalTeamId, VendorId } from '../../../api/types';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useDeleteCapability } from '../../capabilities/hooks/useCapabilities';
import { useComponents, useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import { useDeleteVendor } from '../../origin-entities/hooks/useVendors';
import type { BulkOperationRequest } from '../components/context-menus/MultiSelectContextMenu';
import type { NodeContextMenu } from './useNodeContextMenu';

export interface BulkOperationResult {
  succeeded: string[];
  failed: { name: string; error: string }[];
}

async function executeSequentially(
  nodes: NodeContextMenu[],
  action: (node: NodeContextMenu) => Promise<void>,
): Promise<BulkOperationResult> {
  const succeeded: string[] = [];
  const failed: BulkOperationResult['failed'] = [];

  for (const node of nodes) {
    try {
      await action(node);
      succeeded.push(node.nodeName);
    } catch (error) {
      failed.push({ name: node.nodeName, error: error instanceof Error ? error.message : 'Unknown error' });
      break;
    }
  }

  return { succeeded, failed };
}

interface OriginEntityDeleteMutations {
  acquired: { mutateAsync: (p: { id: AcquiredEntityId; name: string }) => Promise<void> };
  vendor: { mutateAsync: (p: { id: VendorId; name: string }) => Promise<void> };
  team: { mutateAsync: (p: { id: InternalTeamId; name: string }) => Promise<void> };
}

function deleteOriginEntity(
  entityId: string,
  originEntityType: OriginEntityType,
  name: string,
  mutations: OriginEntityDeleteMutations,
): Promise<void> {
  const strategies: Record<OriginEntityType, () => Promise<void>> = {
    acquired: () => mutations.acquired.mutateAsync({ id: entityId as AcquiredEntityId, name }),
    vendor: () => mutations.vendor.mutateAsync({ id: entityId as VendorId, name }),
    team: () => mutations.team.mutateAsync({ id: entityId as InternalTeamId, name }),
  };
  return strategies[originEntityType]();
}

function deleteComponentFromModel(
  node: NodeContextMenu,
  components: Component[],
  mutation: { mutateAsync: (c: Component) => Promise<void> },
): Promise<void> {
  const component = components.find((c) => c.id === node.nodeId);
  if (!component) throw new Error(`Component not found: ${node.nodeName}`);
  return mutation.mutateAsync(component);
}

function deleteCapabilityFromModel(
  node: NodeContextMenu,
  capabilities: Capability[],
  mutation: { mutateAsync: (p: { capability: Capability }) => Promise<void> },
): Promise<void> {
  const capability = capabilities.find((c) => c.id === node.nodeId);
  if (!capability) throw new Error(`Capability not found: ${node.nodeName}`);
  return mutation.mutateAsync({ capability });
}

function deleteOriginEntityFromModel(node: NodeContextMenu, mutations: OriginEntityDeleteMutations): Promise<void> {
  if (!node.originEntityType) throw new Error(`Origin entity not found: ${node.nodeName}`);
  return deleteOriginEntity(node.nodeId, node.originEntityType, node.nodeName, mutations);
}

function useModelDeleteHandler() {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const deleteComponentMutation = useDeleteComponent();
  const deleteCapabilityMutation = useDeleteCapability();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();

  const originDeleteMutations = useMemo<OriginEntityDeleteMutations>(
    () => ({
      acquired: deleteAcquiredEntityMutation,
      vendor: deleteVendorMutation,
      team: deleteInternalTeamMutation,
    }),
    [deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation],
  );

  return useCallback(
    (node: NodeContextMenu): Promise<void> => {
      const handlers: Record<NodeContextMenu['nodeType'], () => Promise<void>> = {
        component: () => deleteComponentFromModel(node, components, deleteComponentMutation),
        capability: () => deleteCapabilityFromModel(node, capabilities, deleteCapabilityMutation),
        originEntity: () => deleteOriginEntityFromModel(node, originDeleteMutations),
      };
      return handlers[node.nodeType]();
    },
    [components, capabilities, deleteComponentMutation, deleteCapabilityMutation, originDeleteMutations],
  );
}

export const useBulkOperations = () => {
  const draftRemoveEntities = useAppStore((s) => s.draftRemoveEntities);
  const deleteFromModel = useModelDeleteHandler();

  const [bulkOperation, setBulkOperation] = useState<BulkOperationRequest | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<BulkOperationResult | null>(null);

  const finishExecution = useCallback((operationResult: BulkOperationResult) => {
    setIsExecuting(false);
    if (operationResult.failed.length > 0) {
      setResult(operationResult);
    } else {
      setBulkOperation(null);
    }
  }, []);

  const executeBulkRemoveFromView = useCallback(
    async (nodes: NodeContextMenu[]) => {
      setIsExecuting(true);
      setResult(null);
      draftRemoveEntities(nodes.map((n) => n.nodeId));
      finishExecution({ succeeded: nodes.map((n) => n.nodeName), failed: [] });
    },
    [draftRemoveEntities, finishExecution],
  );

  const executeBulkDeleteFromModel = useCallback(
    async (nodes: NodeContextMenu[]) => {
      setIsExecuting(true);
      setResult(null);
      const operationResult = await executeSequentially(nodes, deleteFromModel);
      finishExecution(operationResult);
    },
    [deleteFromModel, finishExecution],
  );

  const bulkExecutors = useMemo<Record<BulkOperationRequest['type'], (nodes: NodeContextMenu[]) => Promise<void>>>(
    () => ({
      removeFromView: executeBulkRemoveFromView,
      deleteFromModel: executeBulkDeleteFromModel,
    }),
    [executeBulkRemoveFromView, executeBulkDeleteFromModel],
  );

  const handleBulkConfirm = useCallback(async () => {
    if (!bulkOperation) return;
    await bulkExecutors[bulkOperation.type](bulkOperation.nodes);
  }, [bulkOperation, bulkExecutors]);

  const handleBulkCancel = useCallback(() => {
    setBulkOperation(null);
    setResult(null);
  }, []);

  return {
    bulkOperation,
    isExecuting,
    result,
    setBulkOperation,
    handleBulkConfirm,
    handleBulkCancel,
  };
};
