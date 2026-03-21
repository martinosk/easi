import React, { useState, useMemo, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useAppStore } from "../../../store/appStore";
import { useNavigationTree } from "../hooks/useNavigationTree";
import { NavigationTreeContent } from "./NavigationTreeContent";
import { NavigationTreeDialogs } from "./NavigationTreeDialogs";
import { ContextMenu } from "../../../components/shared/ContextMenu";
import type { ContextMenuItem } from "../../../components/shared/ContextMenu";
import type { NavigationTreeProps } from "../types";
import type { TreeSelectedItem } from "../hooks/useTreeMultiSelect";
import type { ComponentId, ValueStream } from "../../../api/types";
import { useValueStreams } from "../../value-streams";
import { useDeleteValueStream } from "../../value-streams/hooks/useValueStreams";
import { getContextMenuPosition } from "../utils/treeUtils";

type OriginEntityDialogType = "acquired" | "vendor" | "team" | null;

function useEscapeToClearSelection(
  selectionCount: number,
  clearFn: () => void,
) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && selectionCount >= 2) {
        clearFn();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [selectionCount, clearFn]);
}

function useClearSingleSelectionOnMulti(
  selectionCount: number,
  selectNode: (id: ComponentId | null) => void,
  setSelectedCapabilityId: (id: string | null) => void,
) {
  useEffect(() => {
    if (selectionCount >= 2) {
      selectNode(null);
      setSelectedCapabilityId(null);
    }
  }, [selectionCount, selectNode, setSelectedCapabilityId]);
}

function useOriginEntityDialog(canCreateOriginEntity: boolean) {
  const [openOriginDialog, setOpenOriginDialog] =
    useState<OriginEntityDialogType>(null);

  const onAddAcquired = useMemo(
    () =>
      canCreateOriginEntity ? () => setOpenOriginDialog("acquired") : undefined,
    [canCreateOriginEntity],
  );
  const onAddVendor = useMemo(
    () =>
      canCreateOriginEntity ? () => setOpenOriginDialog("vendor") : undefined,
    [canCreateOriginEntity],
  );
  const onAddTeam = useMemo(
    () =>
      canCreateOriginEntity ? () => setOpenOriginDialog("team") : undefined,
    [canCreateOriginEntity],
  );
  const closeDialog = useCallback(() => setOpenOriginDialog(null), []);

  return {
    openOriginDialog,
    closeDialog,
    onAddAcquired,
    onAddVendor,
    onAddTeam,
  };
}

function useMultiContextMenu(
  handleMultiSelectContextMenu: (
    event: React.MouseEvent,
    itemId: string,
    selectedItems: TreeSelectedItem[],
  ) => boolean,
  clearMultiSelection: () => void,
) {
  return useCallback(
    (
      event: React.MouseEvent,
      itemId: string,
      selectedItems: TreeSelectedItem[],
    ) => {
      const handled = handleMultiSelectContextMenu(
        event,
        itemId,
        selectedItems,
      );
      if (!handled && selectedItems.length >= 2) {
        const isInSelection = selectedItems.some((item) => item.id === itemId);
        if (!isInSelection) {
          clearMultiSelection();
        }
      }
      return handled;
    },
    [handleMultiSelectContextMenu, clearMultiSelection],
  );
}

function useBulkOperations(
  requestBulkDelete: (items: TreeSelectedItem[]) => void,
  confirmBulkDelete: () => Promise<void>,
  clearMultiSelection: () => void,
) {
  const handleBulkOperation = useCallback(
    (request: { type: "deleteFromModel"; items: TreeSelectedItem[] }) => {
      requestBulkDelete(request.items);
    },
    [requestBulkDelete],
  );

  const handleBulkDeleteConfirm = useCallback(async () => {
    await confirmBulkDelete();
    clearMultiSelection();
  }, [confirmBulkDelete, clearMultiSelection]);

  return { handleBulkOperation, handleBulkDeleteConfirm };
}

export const NavigationTree: React.FC<NavigationTreeProps> = (props) => {
  const {
    onComponentSelect,
    onViewSelect,
    onAddComponent,
    onCapabilitySelect,
    onAddCapability,
    onOriginEntitySelect,
  } = props;

  const tree = useNavigationTree(props);
  const selectNode = useAppStore((state) => state.selectNode);

  useClearSingleSelectionOnMulti(
    tree.multiSelect.selectionCount,
    selectNode,
    tree.setSelectedCapabilityId,
  );
  useEscapeToClearSelection(
    tree.multiSelect.selectionCount,
    tree.multiSelect.clearMultiSelection,
  );

  const { handleBulkOperation, handleBulkDeleteConfirm } = useBulkOperations(
    tree.bulkDelete.requestBulkDelete,
    tree.bulkDelete.handleConfirm,
    tree.multiSelect.clearMultiSelection,
  );

  const handleMultiContextMenu = useMultiContextMenu(
    tree.multiSelectMenu.handleMultiSelectContextMenu,
    tree.multiSelect.clearMultiSelection,
  );

  const multiSelectProps = useMemo(
    () => ({
      isMultiSelected: tree.multiSelect.isMultiSelected,
      handleItemClick: tree.multiSelect.handleItemClick,
      handleContextMenu: handleMultiContextMenu,
      handleDragStart: tree.multiSelect.handleDragStart,
      selectedItems: tree.multiSelect.selectedItems,
    }),
    [
      tree.multiSelect.isMultiSelected,
      tree.multiSelect.handleItemClick,
      tree.multiSelect.handleDragStart,
      tree.multiSelect.selectedItems,
      handleMultiContextMenu,
    ],
  );

  const originEntity = useOriginEntityDialog(tree.canCreateOriginEntity);

  // Value streams
  const navigate = useNavigate();
  const { valueStreams, deleteValueStream } = useValueStreams();
  const [valueStreamContextMenu, setValueStreamContextMenu] = useState<{
    x: number;
    y: number;
    valueStream: ValueStream;
  } | null>(null);

  const handleValueStreamContextMenu = useCallback(
    (e: React.MouseEvent, vs: ValueStream) => {
      e.preventDefault();
      const pos = getContextMenuPosition(e);
      setValueStreamContextMenu({ ...pos, valueStream: vs });
    },
    [],
  );

  const handleAddValueStream = useCallback(
    () => navigate("/value-streams"),
    [navigate],
  );

  const valueStreamContextMenuItems = useMemo((): ContextMenuItem[] => {
    if (!valueStreamContextMenu) return [];
    const vs = valueStreamContextMenu.valueStream;
    const items: ContextMenuItem[] = [];
    if (vs._links?.edit) {
      items.push({
        label: "Edit",
        onClick: () => navigate(`/value-streams/${vs.id}`),
      });
    }
    if (vs._links?.delete) {
      items.push({
        label: "Delete",
        onClick: () => deleteValueStream(vs),
        isDanger: true,
        ariaLabel: "Delete value stream",
      });
    }
    return items;
  }, [valueStreamContextMenu, navigate, deleteValueStream]);

  return (
    <>
      <div
        className={`navigation-tree ${tree.treeState.isOpen ? "open" : "closed"}`}
      >
        {tree.treeState.isOpen && (
          <NavigationTreeContent
            components={tree.filtered.components}
            currentView={tree.currentView}
            selectedNodeId={tree.selectedNodeId}
            capabilities={tree.filtered.capabilities}
            views={tree.views}
            acquiredEntities={tree.filtered.acquiredEntities}
            vendors={tree.filtered.vendors}
            internalTeams={tree.filtered.internalTeams}
            selectedCapabilityId={tree.selectedCapabilityId}
            setSelectedCapabilityId={tree.setSelectedCapabilityId}
            selectedEntityIds={tree.selectedEntityIds}
            treeState={tree.treeState}
            contextMenus={tree.contextMenus}
            multiSelect={multiSelectProps}
            selectionCount={tree.multiSelect.selectionCount}
            onComponentSelect={onComponentSelect}
            onViewSelect={onViewSelect}
            onAddComponent={onAddComponent}
            onCapabilitySelect={onCapabilitySelect}
            onAddCapability={onAddCapability}
            onOriginEntitySelect={onOriginEntitySelect}
            canCreateView={tree.canCreateView}
            onAddAcquiredEntity={originEntity.onAddAcquired}
            onAddVendor={originEntity.onAddVendor}
            onAddTeam={originEntity.onAddTeam}
            onAddValueStream={handleAddValueStream}
            onValueStreamContextMenu={handleValueStreamContextMenu}
            valueStreams={valueStreams}
            artifactCreators={tree.filteredData.artifactCreators}
            users={tree.filteredData.activeUsers}
            selectedCreatorIds={tree.filteredData.selectedCreatorIds}
            onCreatorSelectionChange={tree.filteredData.setSelectedCreatorIds}
            domains={tree.filteredData.domains}
            selectedDomainIds={tree.filteredData.selectedDomainIds}
            onDomainSelectionChange={tree.filteredData.setSelectedDomainIds}
            hasActiveFilters={tree.filteredData.hasActiveFilters}
            onClearAllFilters={tree.filteredData.clearAllFilters}
          />
        )}
      </div>

      {!tree.treeState.isOpen && (
        <button
          className="tree-toggle-btn-collapsed"
          onClick={() => tree.treeState.setIsOpen(true)}
          aria-label="Open navigation"
        >
          ›
        </button>
      )}

      <NavigationTreeDialogs
        contextMenus={tree.contextMenus}
        openOriginDialog={originEntity.openOriginDialog}
        onCloseOriginDialog={originEntity.closeDialog}
        multiSelectMenu={tree.multiSelectMenu.menu}
        onCloseMultiSelectMenu={tree.multiSelectMenu.closeMenu}
        onRequestBulkOperation={handleBulkOperation}
        bulkDelete={tree.bulkDelete}
        onBulkDeleteConfirm={handleBulkDeleteConfirm}
      />

      {valueStreamContextMenu && valueStreamContextMenuItems.length > 0 && (
        <ContextMenu
          x={valueStreamContextMenu.x}
          y={valueStreamContextMenu.y}
          items={valueStreamContextMenuItems}
          onClose={() => setValueStreamContextMenu(null)}
        />
      )}
    </>
  );
};
