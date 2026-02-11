import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useNavigationTreeState } from '../hooks/useNavigationTreeState';
import { useTreeContextMenus } from '../hooks/useTreeContextMenus';
import { useTreeMultiSelect } from '../hooks/useTreeMultiSelect';
import type { TreeSelectedItem } from '../hooks/useTreeMultiSelect';
import { useTreeMultiSelectMenu } from '../hooks/useTreeMultiSelectMenu';
import { useTreeBulkDelete } from '../hooks/useTreeBulkDelete';
import { useFilteredTreeData } from '../hooks/useFilteredTreeData';
import { NavigationTreeContent } from './NavigationTreeContent';
import { NavigationTreeDialogs } from './NavigationTreeDialogs';
import type { NavigationTreeProps } from '../types';
import type { ComponentId } from '../../../api/types';

type OriginEntityDialogType = 'acquired' | 'vendor' | 'team' | null;

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

function useEscapeToClearSelection(selectionCount: number, clearFn: () => void) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && selectionCount >= 2) {
        clearFn();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [selectionCount, clearFn]);
}

function useClearSingleSelectionOnMulti(
  selectionCount: number,
  selectNode: (id: ComponentId | null) => void,
  setSelectedCapabilityId: (id: string | null) => void
) {
  useEffect(() => {
    if (selectionCount >= 2) {
      selectNode(null);
      setSelectedCapabilityId(null);
    }
  }, [selectionCount, selectNode, setSelectedCapabilityId]);
}

function useOriginEntityDialog(canCreateOriginEntity: boolean) {
  const [openOriginDialog, setOpenOriginDialog] = useState<OriginEntityDialogType>(null);

  const onAddAcquired = useMemo(
    () => (canCreateOriginEntity ? () => setOpenOriginDialog('acquired') : undefined),
    [canCreateOriginEntity]
  );
  const onAddVendor = useMemo(
    () => (canCreateOriginEntity ? () => setOpenOriginDialog('vendor') : undefined),
    [canCreateOriginEntity]
  );
  const onAddTeam = useMemo(
    () => (canCreateOriginEntity ? () => setOpenOriginDialog('team') : undefined),
    [canCreateOriginEntity]
  );
  const closeDialog = useCallback(() => setOpenOriginDialog(null), []);

  return { openOriginDialog, closeDialog, onAddAcquired, onAddVendor, onAddTeam };
}

function useMultiContextMenu(
  multiSelectMenu: ReturnType<typeof useTreeMultiSelectMenu>,
  clearMultiSelection: () => void
) {
  return useCallback(
    (event: React.MouseEvent, itemId: string, selectedItems: TreeSelectedItem[]) => {
      const handled = multiSelectMenu.handleMultiSelectContextMenu(event, itemId, selectedItems);
      if (!handled && selectedItems.length >= 2) {
        const isInSelection = selectedItems.some((item) => item.id === itemId);
        if (!isInSelection) {
          clearMultiSelection();
        }
      }
      return handled;
    },
    [multiSelectMenu.handleMultiSelectContextMenu, clearMultiSelection]
  );
}

function useBulkOperations(
  bulkDelete: ReturnType<typeof useTreeBulkDelete>,
  clearMultiSelection: () => void
) {
  const handleBulkOperation = useCallback(
    (request: { type: 'deleteFromModel'; items: TreeSelectedItem[] }) => {
      bulkDelete.requestBulkDelete(request.items);
    },
    [bulkDelete]
  );

  const handleBulkDeleteConfirm = useCallback(async () => {
    await bulkDelete.handleConfirm();
    clearMultiSelection();
  }, [bulkDelete, clearMultiSelection]);

  return { handleBulkOperation, handleBulkDeleteConfirm };
}

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
  onAddComponent,
  onCapabilitySelect,
  onAddCapability,
  onEditCapability,
  onEditComponent,
  onOriginEntitySelect,
  canCreateView = true,
  canCreateOriginEntity = false,
}) => {
  const {
    components, views, filtered, artifactCreators, activeUsers,
    selectedCreatorIds, setSelectedCreatorIds,
    domains, selectedDomainIds, setSelectedDomainIds,
    hasActiveFilters, clearAllFilters,
  } = useFilteredTreeData();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectNode = useAppStore((state) => state.selectNode);

  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);

  const treeState = useNavigationTreeState();
  const contextMenus = useTreeContextMenus({ components, onEditCapability, onEditComponent });

  const multiSelect = useTreeMultiSelect();
  const multiSelectMenu = useTreeMultiSelectMenu();
  const bulkDelete = useTreeBulkDelete();

  useClearSingleSelectionOnMulti(multiSelect.selectionCount, selectNode, setSelectedCapabilityId);
  useEscapeToClearSelection(multiSelect.selectionCount, multiSelect.clearMultiSelection);

  const { handleBulkOperation, handleBulkDeleteConfirm } = useBulkOperations(bulkDelete, multiSelect.clearMultiSelection);

  const selectedEntityIds = useMemo(
    () => extractSelectedEntityIds(selectedNodeId),
    [selectedNodeId]
  );

  const handleMultiContextMenu = useMultiContextMenu(multiSelectMenu, multiSelect.clearMultiSelection);

  const multiSelectProps = useMemo(
    () => ({
      isMultiSelected: multiSelect.isMultiSelected,
      handleItemClick: multiSelect.handleItemClick,
      handleContextMenu: handleMultiContextMenu,
      handleDragStart: multiSelect.handleDragStart,
      selectedItems: multiSelect.selectedItems,
    }),
    [multiSelect.isMultiSelected, multiSelect.handleItemClick, multiSelect.handleDragStart, multiSelect.selectedItems, handleMultiContextMenu]
  );

  const originEntity = useOriginEntityDialog(canCreateOriginEntity);

  return (
    <>
      <div className={`navigation-tree ${treeState.isOpen ? 'open' : 'closed'}`}>
        {treeState.isOpen && (
          <NavigationTreeContent
            components={filtered.components}
            currentView={currentView}
            selectedNodeId={selectedNodeId}
            capabilities={filtered.capabilities}
            views={views}
            acquiredEntities={filtered.acquiredEntities}
            vendors={filtered.vendors}
            internalTeams={filtered.internalTeams}
            selectedCapabilityId={selectedCapabilityId}
            setSelectedCapabilityId={setSelectedCapabilityId}
            selectedEntityIds={selectedEntityIds}
            treeState={treeState}
            contextMenus={contextMenus}
            multiSelect={multiSelectProps}
            selectionCount={multiSelect.selectionCount}
            onComponentSelect={onComponentSelect}
            onViewSelect={onViewSelect}
            onAddComponent={onAddComponent}
            onCapabilitySelect={onCapabilitySelect}
            onAddCapability={onAddCapability}
            onOriginEntitySelect={onOriginEntitySelect}
            canCreateView={canCreateView}
            onAddAcquiredEntity={originEntity.onAddAcquired}
            onAddVendor={originEntity.onAddVendor}
            onAddTeam={originEntity.onAddTeam}
            artifactCreators={artifactCreators}
            users={activeUsers}
            selectedCreatorIds={selectedCreatorIds}
            onCreatorSelectionChange={setSelectedCreatorIds}
            domains={domains}
            selectedDomainIds={selectedDomainIds}
            onDomainSelectionChange={setSelectedDomainIds}
            hasActiveFilters={hasActiveFilters}
            onClearAllFilters={clearAllFilters}
          />
        )}
      </div>

      {!treeState.isOpen && (
        <button
          className="tree-toggle-btn-collapsed"
          onClick={() => treeState.setIsOpen(true)}
          aria-label="Open navigation"
        >
          ›
        </button>
      )}

      <NavigationTreeDialogs
        contextMenus={contextMenus}
        openOriginDialog={originEntity.openOriginDialog}
        onCloseOriginDialog={originEntity.closeDialog}
        multiSelectMenu={multiSelectMenu.menu}
        onCloseMultiSelectMenu={multiSelectMenu.closeMenu}
        onRequestBulkOperation={handleBulkOperation}
        bulkDelete={bulkDelete}
        onBulkDeleteConfirm={handleBulkDeleteConfirm}
      />
    </>
  );
};
