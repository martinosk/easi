import { Group, Select, SegmentedControl, Stack, Switch } from '@mantine/core';
import type { BusinessDomainId } from '../../../api/types';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import { type BoardViewMode, type MapDepth, useMapDepth, useMapDomain, useMapShowApps } from '../hooks/useMapViewState';
import { BoardLegend } from './BoardLegend';
import { CapabilityMap } from './CapabilityMap';
import mapClasses from './CapabilityMap.module.css';
import boardClasses from './DomainBoard.module.css';
import { ToolbarSearchInput } from './ToolbarControls';
import { ViewModeToggle } from './ViewModeToggle';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

export interface CapabilityMapViewProps {
  hookData: BusinessDomainsHookReturn;
  viewMode: BoardViewMode;
  onViewModeChange: (mode: BoardViewMode) => void;
}

const DEPTH_OPTIONS = [
  { value: '1', label: 'L1' },
  { value: '2', label: 'L2' },
  { value: '3', label: 'L3' },
  { value: '4', label: 'L4' },
];

interface MapToolbarProps extends CapabilityMapViewProps {
  depth: MapDepth;
  onDepthChange: (depth: MapDepth) => void;
  selectedDomainId: BusinessDomainId | null;
  onDomainChange: (id: BusinessDomainId) => void;
  showApps: boolean;
  onShowAppsChange: (show: boolean) => void;
}

function MapToolbar({
  hookData,
  viewMode,
  onViewModeChange,
  depth,
  onDepthChange,
  selectedDomainId,
  onDomainChange,
  showApps,
  onShowAppsChange,
}: MapToolbarProps) {
  const domainOptions = hookData.boardDomains.map((vm) => ({ value: String(vm.domain.id), label: vm.domain.name }));

  return (
    <Stack gap="sm">
      <Group justify="space-between" wrap="wrap" gap="md">
        <ViewModeToggle value={viewMode} onChange={onViewModeChange} />
        <Group gap="sm">
          <Select
            value={selectedDomainId ? String(selectedDomainId) : null}
            onChange={(value) => value && onDomainChange(value as BusinessDomainId)}
            data={domainOptions}
            allowDeselect={false}
            aria-label="Mapped domain"
            data-testid="map-domain-select"
          />
          <SegmentedControl
            value={String(depth)}
            onChange={(value) => onDepthChange(Number(value) as MapDepth)}
            data={DEPTH_OPTIONS}
            size="sm"
            data-testid="map-depth-selector"
          />
          <ToolbarSearchInput value={hookData.searchQuery} onChange={hookData.setSearchQuery} />
          <Switch
            checked={showApps}
            onChange={(e) => onShowAppsChange(e.currentTarget.checked)}
            label="Show apps"
            data-testid="map-show-apps-toggle"
          />
        </Group>
      </Group>

      <BoardLegend lens="now" />
    </Stack>
  );
}

export function CapabilityMapView({ hookData, viewMode, onViewModeChange }: CapabilityMapViewProps) {
  const [depth, setDepth] = useMapDepth();
  const [showApps, setShowApps] = useMapShowApps();
  const domains = hookData.boardDomains.map((vm) => vm.domain);
  const [selectedDomainId, setSelectedDomainId] = useMapDomain(domains);
  const viewModel = hookData.boardDomains.find((vm) => vm.domain.id === selectedDomainId);

  return (
    <Stack gap="md" className={boardClasses.page} data-testid="business-domains-map-view">
        <div className={boardClasses.toolbarRow}>
          <MapToolbar
            hookData={hookData}
            viewMode={viewMode}
            onViewModeChange={onViewModeChange}
            depth={depth}
            onDepthChange={setDepth}
            selectedDomainId={selectedDomainId}
            onDomainChange={setSelectedDomainId}
            showApps={showApps}
            onShowAppsChange={setShowApps}
          />
        </div>

        <div className={boardClasses.boardRow}>
          <div className={boardClasses.boardScroll}>
            <div className={mapClasses.scrollInner}>
              {viewModel && (
                <CapabilityMap
                  viewModel={viewModel}
                  depth={depth}
                  searchQuery={hookData.searchQuery}
                  showApps={showApps}
                  getColorForValue={hookData.getColorForValue}
                  onCapabilityClick={(capability, event) =>
                    hookData.handleCapabilityClick(viewModel.domain.id, capability, event)
                  }
                  onChipClick={hookData.handleApplicationClick}
                />
              )}
            </div>
          </div>
        </div>
    </Stack>
  );
}
