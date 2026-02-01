import { useState, useCallback, useMemo } from 'react';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useRemoveComponentFromView, useRemoveCapabilityFromView, useRemoveOriginEntityFromView } from '../../views/hooks/useViews';
import { useDeleteComponent } from '../../components/hooks/useComponents';
import { useDeleteCapability } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useDeleteAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteVendor } from '../../origin-entities/hooks/useVendors';
import { useDeleteInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import { extractOriginEntityId } from '../utils/nodeFactory';
import { toComponentId, toCapabilityId } from '../../../api/types';
import type { ViewId, AcquiredEntityId, VendorId, InternalTeamId, Component, Capability } from '../../../api/types';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import type { NodeContextMenu } from './useNodeContextMenu';
import type { BulkOperationRequest } from '../components/context-menus/MultiSelectContextMenu';

export interface BulkOperationResult {
  succeeded: string[];
  failed: { name: string; error: string }[];
}

function collectSettledResults(
  results: PromiseSettledResult<void>[],
  nodes: NodeContextMenu[]
): BulkOperationResult {
  const succeeded: string[] = [];
  const failed: BulkOperationResult['failed'] = [];

  results.forEach((r, i) => {
    if (r.status === 'fulfilled') {
      succeeded.push(nodes[i].nodeName);
    } else {
      failed.push({ name: nodes[i].nodeName, error: r.reason?.message ?? 'Unknown error' });
    }
  });

  return { succeeded, failed };
}

async function executeSequentially(
  nodes: NodeContextMenu[],
  action: (node: NodeContextMenu) => Promise<void>
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

type RemoveMutations = {
  component: { mutateAsync: (p: { viewId: ViewId; componentId: ReturnType<typeof toComponentId> }) => Promise<void> };
  capability: { mutateAsync: (p: { viewId: ViewId; capabilityId: ReturnType<typeof toCapabilityId> }) => Promise<void> };
  originEntity: { mutateAsync: (p: { viewId: ViewId; originEntityId: string }) => Promise<void> };
};

function removeNodeFromView(node: NodeContextMenu, viewId: ViewId, mutations: RemoveMutations): Promise<void> {
  const handlers: Record<NodeContextMenu['nodeType'], () => Promise<void>> = {
    component: () => mutations.component.mutateAsync({ viewId, componentId: toComponentId(node.nodeId) }),
    capability: () => mutations.capability.mutateAsync({ viewId, capabilityId: toCapabilityId(node.nodeId) }),
    originEntity: () => mutations.originEntity.mutateAsync({ viewId, originEntityId: node.nodeId }),
  };
  return handlers[node.nodeType]();
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
  mutations: OriginEntityDeleteMutations
): Promise<void> {
  const strategies: Record<OriginEntityType, () => Promise<void>> = {
    acquired: () => mutations.acquired.mutateAsync({ id: entityId as AcquiredEntityId, name }),
    vendor: () => mutations.vendor.mutateAsync({ id: entityId as VendorId, name }),
    team: () => mutations.team.mutateAsync({ id: entityId as InternalTeamId, name }),
  };
  return strategies[originEntityType]();
}

function useRemoveMutations(): RemoveMutations {
  const removeComponentMutation = useRemoveComponentFromView();
  const removeCapabilityMutation = useRemoveCapabilityFromView();
  const removeOriginEntityMutation = useRemoveOriginEntityFromView();

  return useMemo<RemoveMutations>(() => ({
    component: removeComponentMutation,
    capability: removeCapabilityMutation,
    originEntity: removeOriginEntityMutation,
  }), [removeComponentMutation, removeCapabilityMutation, removeOriginEntityMutation]);
}

function deleteComponentFromModel(
  node: NodeContextMenu,
  components: Component[],
  mutation: { mutateAsync: (c: Component) => Promise<void> }
): Promise<void> {
  const component = components.find((c) => c.id === node.nodeId);
  if (!component) throw new Error(`Component not found: ${node.nodeName}`);
  return mutation.mutateAsync(component);
}

function deleteCapabilityFromModel(
  node: NodeContextMenu,
  capabilities: Capability[],
  mutation: { mutateAsync: (p: { capability: Capability }) => Promise<void> }
): Promise<void> {
  const capability = capabilities.find((c) => c.id === node.nodeId);
  if (!capability) throw new Error(`Capability not found: ${node.nodeName}`);
  return mutation.mutateAsync({ capability });
}

function deleteOriginEntityFromModel(
  node: NodeContextMenu,
  mutations: OriginEntityDeleteMutations
): Promise<void> {
  const entityId = extractOriginEntityId(node.nodeId);
  if (!entityId || !node.originEntityType) throw new Error(`Origin entity not found: ${node.nodeName}`);
  return deleteOriginEntity(entityId, node.originEntityType, node.nodeName, mutations);
}

function useModelDeleteHandler() {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const deleteComponentMutation = useDeleteComponent();
  const deleteCapabilityMutation = useDeleteCapability();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();

  const originDeleteMutations = useMemo<OriginEntityDeleteMutations>(() => ({
    acquired: deleteAcquiredEntityMutation,
    vendor: deleteVendorMutation,
    team: deleteInternalTeamMutation,
  }), [deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation]);

  return useCallback(
    (node: NodeContextMenu): Promise<void> => {
      const handlers: Record<NodeContextMenu['nodeType'], () => Promise<void>> = {
        component: () => deleteComponentFromModel(node, components, deleteComponentMutation),
        capability: () => deleteCapabilityFromModel(node, capabilities, deleteCapabilityMutation),
        originEntity: () => deleteOriginEntityFromModel(node, originDeleteMutations),
      };
      return handlers[node.nodeType]();
    },
    [components, capabilities, deleteComponentMutation, deleteCapabilityMutation, originDeleteMutations]
  );
}

export const useBulkOperations = () => {
  const { currentViewId } = useCurrentView();

  const removeMutations = useRemoveMutations();
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
      if (!currentViewId) return;
      setIsExecuting(true);
      setResult(null);
      const results = await Promise.allSettled(
        nodes.map((node) => removeNodeFromView(node, currentViewId, removeMutations))
      );
      finishExecution(collectSettledResults(results, nodes));
    },
    [currentViewId, removeMutations, finishExecution]
  );

  const executeBulkDeleteFromModel = useCallback(
    async (nodes: NodeContextMenu[]) => {
      setIsExecuting(true);
      setResult(null);
      const operationResult = await executeSequentially(nodes, deleteFromModel);
      finishExecution(operationResult);
    },
    [deleteFromModel, finishExecution]
  );

  const bulkExecutors = useMemo<Record<BulkOperationRequest['type'], (nodes: NodeContextMenu[]) => Promise<void>>>(() => ({
    removeFromView: executeBulkRemoveFromView,
    deleteFromModel: executeBulkDeleteFromModel,
  }), [executeBulkRemoveFromView, executeBulkDeleteFromModel]);

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
