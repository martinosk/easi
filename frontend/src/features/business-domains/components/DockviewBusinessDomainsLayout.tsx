import { DockviewReact } from 'dockview';
import type { IDockviewPanelProps } from 'dockview';
import { DomainsSidebar } from './DomainsSidebar';
import { CapabilityExplorerSidebar } from './CapabilityExplorerSidebar';
import { VisualizationArea } from './VisualizationArea';
import { DetailsSidebar } from './DetailsSidebar';
import { DockviewToolbar } from './dockview/DockviewToolbar';
import { useDockviewLayout } from './dockview/useDockviewLayout';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

type DomainsSidebarPanelProps = IDockviewPanelProps<{
  domains: BusinessDomainsHookReturn['domains'];
  selectedDomainId: BusinessDomainsHookReturn['visualizedDomain'] extends infer T ? T extends null ? undefined : T extends { id: infer U } ? U : undefined : undefined;
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

const panelContainerStyle = { height: '100%', width: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' } as const;

const DomainsSidebarPanel = (props: DomainsSidebarPanelProps) => (
  <div style={panelContainerStyle}>
    <DomainsSidebar
      domains={props.params.domains}
      selectedDomainId={props.params.selectedDomainId}
      onCreateClick={props.params.onCreateClick}
      onVisualize={props.params.onVisualize}
      onContextMenu={props.params.onContextMenu}
    />
  </div>
);

const VisualizationPanel = (props: VisualizationPanelProps) => (
  <div style={panelContainerStyle}>
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
  </div>
);

const ExplorerPanel = (props: ExplorerPanelProps) => (
  <div style={panelContainerStyle}>
    <CapabilityExplorerSidebar
      visualizedDomain={props.params.visualizedDomain}
      capabilities={props.params.capabilities}
      assignedCapabilityIds={props.params.assignedCapabilityIds}
      isLoading={false}
      onDragStart={props.params.onDragStart}
      onDragEnd={props.params.onDragEnd}
    />
  </div>
);

const DetailsPanel = (props: DetailsPanelProps) => (
  <div style={{ ...panelContainerStyle, overflow: 'auto' }}>
    <DetailsSidebar
      selectedCapability={props.params.selectedCapability}
      selectedComponentId={props.params.selectedComponentId}
      visualizedDomain={props.params.visualizedDomain}
    />
  </div>
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
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, overflow: 'hidden' }} data-testid="business-domains-page">
      <DockviewToolbar panelVisibility={panelVisibility} onTogglePanel={togglePanel} showExplorer={showExplorer} />
      <div style={{ flex: 1, minHeight: 0, position: 'relative' }}>
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}>
          <DockviewReact
            onReady={onReady}
            components={components}
            className="dockview-theme-light"
          />
        </div>
      </div>
    </div>
  );
}
