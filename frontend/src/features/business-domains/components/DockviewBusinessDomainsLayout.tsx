import { Box } from '@mantine/core';
import type { IDockviewPanelProps } from 'dockview';
import { DockviewReact, themeLight } from 'dockview';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import { CapabilityExplorerSidebar } from './CapabilityExplorerSidebar';
import { DetailsSidebar } from './DetailsSidebar';
import { DockviewToolbar } from './dockview/DockviewToolbar';
import { useDockviewLayout } from './dockview/useDockviewLayout';
import { DomainsSidebar } from './DomainsSidebar';
import { VisualizationArea } from './VisualizationArea';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

type DomainsSidebarPanelProps = IDockviewPanelProps<{
  domains: BusinessDomainsHookReturn['domains'];
  canCreateDomain: BusinessDomainsHookReturn['canCreateDomain'];
  selectedDomainId: BusinessDomainsHookReturn['visualizedDomain'] extends infer T
    ? T extends null
      ? undefined
      : T extends { id: infer U }
        ? U
        : undefined
    : undefined;
  onCreateClick: BusinessDomainsHookReturn['dialogManager']['handleCreateClick'];
  onVisualize: BusinessDomainsHookReturn['handleVisualizeClick'];
  onContextMenu: BusinessDomainsHookReturn['domainContextMenu']['handleContextMenu'];
}>;

type VisualizationPanelProps = IDockviewPanelProps<{
  visualizedDomain: BusinessDomainsHookReturn['visualizedDomain'];
  capabilities: BusinessDomainsHookReturn['filtering']['capabilitiesWithDescendants'];
  capabilitiesLoading: BusinessDomainsHookReturn['capabilitiesLoading'];
  depth: BusinessDomainsHookReturn['depth'];
  positions: BusinessDomainsHookReturn['positions'];
  selectedCapabilities: BusinessDomainsHookReturn['selectedCapabilities'];
  showApplications: BusinessDomainsHookReturn['showApplications'];
  isDragOver: BusinessDomainsHookReturn['dragHandlers']['isDragOver'];
  onDepthChange: BusinessDomainsHookReturn['setDepth'];
  onCapabilityClick: BusinessDomainsHookReturn['handleCapabilityClick'];
  onContextMenu: BusinessDomainsHookReturn['capabilityContextMenu']['handleCapabilityContextMenu'];
  onShowApplicationsChange: BusinessDomainsHookReturn['setShowApplications'];
  getRealizationsForCapability: BusinessDomainsHookReturn['getRealizationsForCapability'];
  onApplicationClick: BusinessDomainsHookReturn['handleApplicationClick'];
  onDragOver: BusinessDomainsHookReturn['dragHandlers']['handleDragOver'];
  onDragLeave: BusinessDomainsHookReturn['dragHandlers']['handleDragLeave'];
  onDrop: BusinessDomainsHookReturn['dragHandlers']['handleDrop'];
}>;

type ExplorerPanelProps = IDockviewPanelProps<{
  visualizedDomain: BusinessDomainsHookReturn['visualizedDomain'];
  capabilities: BusinessDomainsHookReturn['filtering']['allCapabilities'];
  assignedCapabilityIds: BusinessDomainsHookReturn['filtering']['assignedCapabilityIds'];
  onDragStart: BusinessDomainsHookReturn['dragHandlers']['handleDragStart'];
  onDragEnd: BusinessDomainsHookReturn['dragHandlers']['handleDragEnd'];
}>;

type DetailsPanelProps = IDockviewPanelProps<{
  selectedCapability: BusinessDomainsHookReturn['selectedCapability'];
  selectedComponentId: BusinessDomainsHookReturn['selectedComponentId'];
  visualizedDomain: BusinessDomainsHookReturn['visualizedDomain'];
}>;

function PanelShell({ children, scroll }: { children: React.ReactNode; scroll?: boolean }) {
  return (
    <Box
      h="100%"
      w="100%"
      display="flex"
      style={{ flexDirection: 'column', overflow: scroll ? 'auto' : 'hidden' }}
    >
      {children}
    </Box>
  );
}

const DomainsSidebarPanel = (props: DomainsSidebarPanelProps) => (
  <PanelShell>
    <DomainsSidebar
      domains={props.params.domains}
      canCreateDomain={props.params.canCreateDomain}
      selectedDomainId={props.params.selectedDomainId}
      onCreateClick={props.params.onCreateClick}
      onVisualize={props.params.onVisualize}
      onContextMenu={props.params.onContextMenu}
    />
  </PanelShell>
);

const VisualizationPanel = (props: VisualizationPanelProps) => (
  <PanelShell>
    <VisualizationArea
      visualizedDomain={props.params.visualizedDomain}
      capabilities={props.params.capabilities}
      capabilitiesLoading={props.params.capabilitiesLoading}
      depth={props.params.depth}
      positions={props.params.positions}
      onDepthChange={props.params.onDepthChange}
      onCapabilityClick={props.params.onCapabilityClick}
      onContextMenu={props.params.onContextMenu}
      selectedCapabilities={props.params.selectedCapabilities}
      showApplications={props.params.showApplications}
      onShowApplicationsChange={props.params.onShowApplicationsChange}
      getRealizationsForCapability={props.params.getRealizationsForCapability}
      onApplicationClick={props.params.onApplicationClick}
      isDragOver={props.params.isDragOver}
      onDragOver={props.params.onDragOver}
      onDragLeave={props.params.onDragLeave}
      onDrop={props.params.onDrop}
    />
  </PanelShell>
);

const ExplorerPanel = (props: ExplorerPanelProps) => (
  <PanelShell>
    <CapabilityExplorerSidebar
      visualizedDomain={props.params.visualizedDomain}
      capabilities={props.params.capabilities}
      assignedCapabilityIds={props.params.assignedCapabilityIds}
      isLoading={false}
      onDragStart={props.params.onDragStart}
      onDragEnd={props.params.onDragEnd}
    />
  </PanelShell>
);

const DetailsPanel = (props: DetailsPanelProps) => (
  <PanelShell scroll>
    <DetailsSidebar
      selectedCapability={props.params.selectedCapability}
      selectedComponentId={props.params.selectedComponentId}
      visualizedDomain={props.params.visualizedDomain}
    />
  </PanelShell>
);

const components = {
  domains: DomainsSidebarPanel,
  visualization: VisualizationPanel,
  explorer: ExplorerPanel,
  details: DetailsPanel,
};

interface DockviewBusinessDomainsLayoutProps {
  hookData: BusinessDomainsHookReturn;
}

export function DockviewBusinessDomainsLayout({ hookData }: DockviewBusinessDomainsLayoutProps) {
  const { onReady, panelVisibility, togglePanel, showExplorer } = useDockviewLayout(hookData);

  return (
    <Box
      display="flex"
      flex={1}
      data-testid="business-domains-page"
      style={{ flexDirection: 'column', minHeight: 0, overflow: 'hidden' }}
    >
      <DockviewToolbar panelVisibility={panelVisibility} onTogglePanel={togglePanel} showExplorer={showExplorer} />
      <Box flex={1} style={{ minHeight: 0, position: 'relative' }}>
        <div
          className="dockview-theme-light"
          style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}
        >
          <DockviewReact onReady={onReady} components={components} theme={themeLight} />
        </div>
      </Box>
    </Box>
  );
}
