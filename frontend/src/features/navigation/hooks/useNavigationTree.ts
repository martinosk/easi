import { useCallback, useMemo, useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useGenerateView } from '../../canvas/hooks/useGenerateView';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { NavigationTreeProps } from '../types';
import { useFilteredTreeData } from './useFilteredTreeData';
import { useNavigationTreeState } from './useNavigationTreeState';
import { useTreeBulkDelete } from './useTreeBulkDelete';
import { type GenerateViewTarget, useTreeContextMenus } from './useTreeContextMenus';
import { useTreeMultiSelect } from './useTreeMultiSelect';
import { useTreeMultiSelectMenu } from './useTreeMultiSelectMenu';

interface SelectedEntityIds {
  acquiredEntityId: string | null;
  vendorId: string | null;
  teamId: string | null;
}

const EMPTY_ENTITY_IDS: SelectedEntityIds = { acquiredEntityId: null, vendorId: null, teamId: null };

const ENTITY_PREFIXES: { prefix: string; key: keyof SelectedEntityIds }[] = [
  { prefix: 'acq-', key: 'acquiredEntityId' },
  { prefix: 'vendor-', key: 'vendorId' },
  { prefix: 'team-', key: 'teamId' },
];

function extractSelectedEntityIds(nodeId: string | null): SelectedEntityIds {
  if (!nodeId) return EMPTY_ENTITY_IDS;
  const result = { ...EMPTY_ENTITY_IDS };
  for (const { prefix, key } of ENTITY_PREFIXES) {
    if (nodeId.startsWith(prefix)) {
      result[key] = nodeId.slice(prefix.length);
      break;
    }
  }
  return result;
}

export function useNavigationTree(props: NavigationTreeProps) {
  const { onEditCapability, onEditComponent, canCreateView = true, canCreateOriginEntity = false } = props;

  const filteredData = useFilteredTreeData();
  const { components, views, filtered } = filteredData;
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);

  const treeState = useNavigationTreeState();
  const { generateView } = useGenerateView();
  const handleGenerateView = useCallback(
    (target: GenerateViewTarget) => generateView(target.entityRef, target.entityName),
    [generateView],
  );
  const contextMenus = useTreeContextMenus({
    components,
    onEditCapability,
    onEditComponent,
    onGenerateView: handleGenerateView,
    canCreateView,
  });

  const multiSelect = useTreeMultiSelect();
  const multiSelectMenu = useTreeMultiSelectMenu();
  const bulkDelete = useTreeBulkDelete();

  const selectedEntityIds = useMemo(() => extractSelectedEntityIds(selectedNodeId), [selectedNodeId]);

  return {
    filteredData,
    currentView,
    selectedNodeId,
    selectedCapabilityId,
    setSelectedCapabilityId,
    treeState,
    contextMenus,
    multiSelect,
    multiSelectMenu,
    bulkDelete,
    selectedEntityIds,
    views,
    filtered,
    canCreateView,
    canCreateOriginEntity,
  };
}
