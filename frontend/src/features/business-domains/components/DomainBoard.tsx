import { Stack, Text } from '@mantine/core';
import { useEffect, useMemo, useRef } from 'react';
import type { BusinessDomainId } from '../../../api/types';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import type { BoardViewMode } from '../hooks/useMapViewState';
import { summaryCounts } from '../lens/journeyIndex';
import { AssignRail } from './AssignRail';
import { BoardLensProvider } from './BoardLensContext';
import { BoardToolbar } from './BoardToolbar';
import classes from './DomainBoard.module.css';
import { DomainBoardCard } from './DomainBoardCard';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

export interface DomainBoardProps {
  hookData: BusinessDomainsHookReturn;
  viewMode: BoardViewMode;
  onViewModeChange: (mode: BoardViewMode) => void;
}

export function DomainBoard({ hookData, viewMode, onViewModeChange }: DomainBoardProps) {
  const cardRefs = useRef(new Map<BusinessDomainId, HTMLDivElement>());

  useEffect(() => {
    if (!hookData.highlightedDomainId) return;
    cardRefs.current.get(hookData.highlightedDomainId)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, [hookData.highlightedDomainId]);

  const summary = useMemo(
    () =>
      summaryCounts(
        hookData.boardDomains.flatMap((vm) => vm.l1Groups.map((group) => group.node)),
        hookData.journeyIndex.getJourney,
      ),
    [hookData.boardDomains, hookData.journeyIndex],
  );

  const showRail = hookData.showAssignRail && hookData.assignRailOpen && hookData.lens === 'now';

  return (
    <BoardLensProvider
      lens={hookData.lens}
      changesOnly={hookData.changesOnly}
      index={hookData.journeyIndex}
      openCapabilityById={hookData.openCapabilityById}
    >
      <Stack gap="md" className={classes.page} data-testid="business-domains-page">
        <div className={classes.toolbarRow}>
          <BoardToolbar
            searchQuery={hookData.searchQuery}
            onSearchChange={hookData.setSearchQuery}
            canCreateDomain={hookData.canCreateDomain}
            onCreateDomain={hookData.dialogManager.handleCreateClick}
            showAssignToggle={hookData.showAssignRail}
            assignRailOpen={hookData.assignRailOpen}
            onToggleAssignRail={hookData.toggleAssignRail}
            lens={hookData.lens}
            onLensChange={hookData.setLens}
            changesOnly={hookData.changesOnly}
            onChangesOnlyChange={hookData.setChangesOnly}
            summary={summary}
            viewMode={viewMode}
            onViewModeChange={onViewModeChange}
          />
        </div>

        <div className={classes.boardRow}>
          <div className={classes.boardScroll}>
            {hookData.boardDomains.length === 0 ? (
              <div className={classes.emptyState}>
                <Text c="dimmed">No business domains yet. Create your first domain to get started.</Text>
              </div>
            ) : (
              <div className={classes.grid} data-testid="domain-board">
                {hookData.boardDomains.map((viewModel) => (
                  <div
                    key={viewModel.domain.id}
                    ref={(el) => {
                      if (el) cardRefs.current.set(viewModel.domain.id, el);
                      else cardRefs.current.delete(viewModel.domain.id);
                    }}
                  >
                    <DomainBoardCard
                      viewModel={viewModel}
                      searchQuery={hookData.searchQuery}
                      selectedCapabilities={hookData.selectedCapabilities}
                      forceOpenL1Ids={hookData.forceOpenL1Ids}
                      isHighlighted={hookData.highlightedDomainId === viewModel.domain.id}
                      isDropTarget={hookData.dragHandlers.dragOverDomainId === viewModel.domain.id}
                      getColorForValue={hookData.getColorForValue}
                      onCapabilityClick={(capability, event) =>
                        hookData.handleCapabilityClick(viewModel.domain.id, capability, event)
                      }
                      onCapabilityContextMenu={(capability, event) =>
                        hookData.capabilityContextMenu.handleCapabilityContextMenu(
                          viewModel.domain.id,
                          capability,
                          event,
                        )
                      }
                      onChipClick={hookData.handleApplicationClick}
                      onDomainMenu={hookData.domainContextMenu.handleContextMenu}
                      onDragOver={(event) => hookData.dragHandlers.handleDragOver(viewModel.domain.id, event)}
                      onDragLeave={hookData.dragHandlers.handleDragLeave}
                      onDrop={(event) => hookData.dragHandlers.handleDrop(viewModel.domain.id, event)}
                    />
                  </div>
                ))}
              </div>
            )}
          </div>

          {showRail && (
            <AssignRail
              allCapabilities={hookData.allCapabilities}
              isLoading={hookData.isLoading}
              globalAssignedCapabilityIds={hookData.globalAssignedCapabilityIds}
              onDragStart={hookData.dragHandlers.handleDragStart}
              onDragEnd={hookData.dragHandlers.handleDragEnd}
            />
          )}
        </div>
      </Stack>
    </BoardLensProvider>
  );
}
