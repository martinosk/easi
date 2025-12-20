import { useEffect, useRef, useState } from 'react';
import { DockviewReact } from 'dockview';
import type { DockviewReadyEvent, IDockviewPanelProps } from 'dockview';
import { DomainsSidebar } from './DomainsSidebar';
import { CapabilityExplorerSidebar } from './CapabilityExplorerSidebar';
import { VisualizationArea } from './VisualizationArea';
import { DetailsSidebar } from './DetailsSidebar';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';

const LAYOUT_STORAGE_KEY = 'easi-business-domains-dockview-layout';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

interface DomainsSidebarPanelProps extends IDockviewPanelProps<{
  domains: BusinessDomainsHookReturn['domains'];
  selectedDomainId: BusinessDomainsHookReturn['visualizedDomain'] extends infer T ? T extends null ? undefined : T extends { id: infer U } ? U : undefined : undefined;
  onCreateClick: BusinessDomainsHookReturn['dialogManager']['handleCreateClick'];
  onVisualize: BusinessDomainsHookReturn['handleVisualizeClick'];
  onContextMenu: BusinessDomainsHookReturn['domainContextMenu']['handleContextMenu'];
}> {}

interface VisualizationPanelProps extends IDockviewPanelProps<{
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
}> {}

interface ExplorerPanelProps extends IDockviewPanelProps<{
  visualizedDomain: BusinessDomainsHookReturn['visualizedDomain'];
  capabilities: BusinessDomainsHookReturn['filtering']['allCapabilities'];
  assignedCapabilityIds: BusinessDomainsHookReturn['filtering']['assignedCapabilityIds'];
  onDragStart: BusinessDomainsHookReturn['dragHandlers']['handleDragStart'];
  onDragEnd: BusinessDomainsHookReturn['dragHandlers']['handleDragEnd'];
}> {}

interface DetailsPanelProps extends IDockviewPanelProps<{
  selectedCapability: BusinessDomainsHookReturn['selectedCapability'];
  selectedComponentId: BusinessDomainsHookReturn['selectedComponentId'];
}> {}

const DomainsSidebarPanel = (props: DomainsSidebarPanelProps) => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <DomainsSidebar
        domains={props.params.domains}
        selectedDomainId={props.params.selectedDomainId}
        onCreateClick={props.params.onCreateClick}
        onVisualize={props.params.onVisualize}
        onContextMenu={props.params.onContextMenu}
      />
    </div>
  );
};

const VisualizationPanel = (props: VisualizationPanelProps) => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
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
};

const ExplorerPanel = (props: ExplorerPanelProps) => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
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
};

const DetailsPanel = (props: DetailsPanelProps) => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
      <DetailsSidebar
        selectedCapability={props.params.selectedCapability}
        selectedComponentId={props.params.selectedComponentId}
      />
    </div>
  );
};

interface DockviewBusinessDomainsLayoutProps {
  hookData: BusinessDomainsHookReturn;
}

export function DockviewBusinessDomainsLayout({ hookData }: DockviewBusinessDomainsLayoutProps) {
  const dockviewApiRef = useRef<DockviewReadyEvent['api'] | null>(null);
  const [panelVisibility, setPanelVisibility] = useState({ domains: true, explorer: true, details: true });
  const panelSizesRef = useRef<{ domains: number; explorer: number; details: number }>({ domains: 320, explorer: 320, details: 300 });

  const onReady = (event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;

    // Clear any corrupted saved layout - remove this after layout is fixed
    localStorage.removeItem(LAYOUT_STORAGE_KEY);

    const savedLayout = null; // Temporarily disabled: localStorage.getItem(LAYOUT_STORAGE_KEY);
    if (savedLayout) {
      try {
        event.api.fromJSON(JSON.parse(savedLayout));

        event.api.getPanel('visualization')?.api.updateParameters({
          visualizedDomain: hookData.visualizedDomain,
          capabilities: hookData.filtering.capabilitiesWithDescendants,
          capabilitiesLoading: hookData.capabilitiesLoading,
          depth: hookData.depth,
          positions: hookData.positions,
          selectedCapabilities: hookData.selectedCapabilities,
          showApplications: hookData.showApplications,
          isDragOver: hookData.dragHandlers.isDragOver,
          onDepthChange: hookData.setDepth,
          onCapabilityClick: hookData.handleCapabilityClick,
          onContextMenu: hookData.capabilityContextMenu.handleCapabilityContextMenu,
          onShowApplicationsChange: hookData.setShowApplications,
          getRealizationsForCapability: hookData.getRealizationsForCapability,
          onApplicationClick: hookData.handleApplicationClick,
          onDragOver: hookData.dragHandlers.handleDragOver,
          onDragLeave: hookData.dragHandlers.handleDragLeave,
          onDrop: hookData.dragHandlers.handleDrop,
        });

        event.api.getPanel('domains')?.api.updateParameters({
          domains: hookData.domains,
          selectedDomainId: hookData.visualizedDomain?.id,
          onCreateClick: hookData.dialogManager.handleCreateClick,
          onVisualize: hookData.handleVisualizeClick,
          onContextMenu: hookData.domainContextMenu.handleContextMenu,
        });

        event.api.getPanel('explorer')?.api.updateParameters({
          visualizedDomain: hookData.visualizedDomain,
          capabilities: hookData.filtering.allCapabilities,
          assignedCapabilityIds: hookData.filtering.assignedCapabilityIds,
          onDragStart: hookData.dragHandlers.handleDragStart,
          onDragEnd: hookData.dragHandlers.handleDragEnd,
        });

        event.api.getPanel('details')?.api.updateParameters({
          selectedCapability: hookData.selectedCapability,
          selectedComponentId: hookData.selectedComponentId,
        });

        return;
      } catch (error) {
        console.error('Failed to load saved layout:', error);
      }
    }

    const visualizationPanel = event.api.addPanel({
      id: 'visualization',
      component: 'visualization',
      title: 'Visualization',
      params: {
        visualizedDomain: hookData.visualizedDomain,
        capabilities: hookData.filtering.capabilitiesWithDescendants,
        capabilitiesLoading: hookData.capabilitiesLoading,
        depth: hookData.depth,
        positions: hookData.positions,
        selectedCapabilities: hookData.selectedCapabilities,
        showApplications: hookData.showApplications,
        isDragOver: hookData.dragHandlers.isDragOver,
        onDepthChange: hookData.setDepth,
        onCapabilityClick: hookData.handleCapabilityClick,
        onContextMenu: hookData.capabilityContextMenu.handleCapabilityContextMenu,
        onShowApplicationsChange: hookData.setShowApplications,
        getRealizationsForCapability: hookData.getRealizationsForCapability,
        onApplicationClick: hookData.handleApplicationClick,
        onDragOver: hookData.dragHandlers.handleDragOver,
        onDragLeave: hookData.dragHandlers.handleDragLeave,
        onDrop: hookData.dragHandlers.handleDrop,
      },
    });

    const domainsPanel = event.api.addPanel({
      id: 'domains',
      component: 'domains',
      title: 'Business Domains',
      position: { referencePanel: visualizationPanel, direction: 'left' },
      params: {
        domains: hookData.domains,
        selectedDomainId: hookData.visualizedDomain?.id,
        onCreateClick: hookData.dialogManager.handleCreateClick,
        onVisualize: hookData.handleVisualizeClick,
        onContextMenu: hookData.domainContextMenu.handleContextMenu,
      },
    });

    const explorerPanel = event.api.addPanel({
      id: 'explorer',
      component: 'explorer',
      title: 'Capability Explorer',
      position: { referencePanel: visualizationPanel, direction: 'right' },
      params: {
        visualizedDomain: hookData.visualizedDomain,
        capabilities: hookData.filtering.allCapabilities,
        assignedCapabilityIds: hookData.filtering.assignedCapabilityIds,
        onDragStart: hookData.dragHandlers.handleDragStart,
        onDragEnd: hookData.dragHandlers.handleDragEnd,
      },
    });

    const detailsPanel = event.api.addPanel({
      id: 'details',
      component: 'details',
      title: 'Details',
      position: { referencePanel: explorerPanel, direction: 'below' },
      params: {
        selectedCapability: hookData.selectedCapability,
        selectedComponentId: hookData.selectedComponentId,
      },
    });

    domainsPanel.api.setSize({ width: 320 });
    explorerPanel.api.setSize({ width: 320 });
    detailsPanel.api.setSize({ height: 300 });
  };

  useEffect(() => {
    if (!dockviewApiRef.current) return;

    const api = dockviewApiRef.current;

    api.getPanel('domains')?.api.updateParameters({
      domains: hookData.domains,
      selectedDomainId: hookData.visualizedDomain?.id,
      onCreateClick: hookData.dialogManager.handleCreateClick,
      onVisualize: hookData.handleVisualizeClick,
      onContextMenu: hookData.domainContextMenu.handleContextMenu,
    });

    api.getPanel('visualization')?.api.updateParameters({
      visualizedDomain: hookData.visualizedDomain,
      capabilities: hookData.filtering.capabilitiesWithDescendants,
      capabilitiesLoading: hookData.capabilitiesLoading,
      depth: hookData.depth,
      positions: hookData.positions,
      selectedCapabilities: hookData.selectedCapabilities,
      showApplications: hookData.showApplications,
      isDragOver: hookData.dragHandlers.isDragOver,
      onDepthChange: hookData.setDepth,
      onCapabilityClick: hookData.handleCapabilityClick,
      onContextMenu: hookData.capabilityContextMenu.handleCapabilityContextMenu,
      onShowApplicationsChange: hookData.setShowApplications,
      getRealizationsForCapability: hookData.getRealizationsForCapability,
      onApplicationClick: hookData.handleApplicationClick,
      onDragOver: hookData.dragHandlers.handleDragOver,
      onDragLeave: hookData.dragHandlers.handleDragLeave,
      onDrop: hookData.dragHandlers.handleDrop,
    });

    api.getPanel('explorer')?.api.updateParameters({
      visualizedDomain: hookData.visualizedDomain,
      capabilities: hookData.filtering.allCapabilities,
      assignedCapabilityIds: hookData.filtering.assignedCapabilityIds,
      onDragStart: hookData.dragHandlers.handleDragStart,
      onDragEnd: hookData.dragHandlers.handleDragEnd,
    });

    api.getPanel('details')?.api.updateParameters({
      selectedCapability: hookData.selectedCapability,
      selectedComponentId: hookData.selectedComponentId,
    });
  }, [hookData]);

  useEffect(() => {
    const api = dockviewApiRef.current;
    if (!api) return;

    const saveLayout = () => {
      const layout = api.toJSON();
      localStorage.setItem(LAYOUT_STORAGE_KEY, JSON.stringify(layout));
    };

    const disposable = api.onDidLayoutChange(saveLayout);
    return () => disposable.dispose();
  }, []);

  const togglePanel = (panelId: 'domains' | 'explorer' | 'details') => {
    const api = dockviewApiRef.current;
    if (!api) return;

    const panel = api.getPanel(panelId);
    const visualizationPanel = api.getPanel('visualization');
    const domainsPanel = api.getPanel('domains');
    const explorerPanel = api.getPanel('explorer');
    const detailsPanel = api.getPanel('details');
    if (!visualizationPanel) return;

    if (domainsPanel) {
      panelSizesRef.current.domains = domainsPanel.api.width;
    }
    if (explorerPanel) {
      panelSizesRef.current.explorer = explorerPanel.api.width;
    }
    if (detailsPanel) {
      panelSizesRef.current.details = detailsPanel.api.height;
    }

    const restoreAllSizes = () => {
      setTimeout(() => {
        const domains = api.getPanel('domains');
        const explorer = api.getPanel('explorer');
        const details = api.getPanel('details');
        if (domains) domains.api.setSize({ width: panelSizesRef.current.domains });
        if (explorer) explorer.api.setSize({ width: panelSizesRef.current.explorer });
        if (details) details.api.setSize({ height: panelSizesRef.current.details });
      }, 0);
    };

    if (panel) {
      api.removePanel(panel);
      setPanelVisibility(prev => ({ ...prev, [panelId]: false }));
      restoreAllSizes();
    } else {
      if (panelId === 'domains') {
        api.addPanel({
          id: 'domains',
          component: 'domains',
          title: 'Business Domains',
          position: { referencePanel: visualizationPanel, direction: 'left' },
          params: {
            domains: hookData.domains,
            selectedDomainId: hookData.visualizedDomain?.id,
            onCreateClick: hookData.dialogManager.handleCreateClick,
            onVisualize: hookData.handleVisualizeClick,
            onContextMenu: hookData.domainContextMenu.handleContextMenu,
          },
        });
      } else if (panelId === 'explorer') {
        api.addPanel({
          id: 'explorer',
          component: 'explorer',
          title: 'Capability Explorer',
          position: { referencePanel: visualizationPanel, direction: 'right' },
          params: {
            visualizedDomain: hookData.visualizedDomain,
            capabilities: hookData.filtering.allCapabilities,
            assignedCapabilityIds: hookData.filtering.assignedCapabilityIds,
            onDragStart: hookData.dragHandlers.handleDragStart,
            onDragEnd: hookData.dragHandlers.handleDragEnd,
          },
        });
      } else if (explorerPanel) {
        api.addPanel({
          id: 'details',
          component: 'details',
          title: 'Details',
          position: { referencePanel: explorerPanel, direction: 'below' },
          params: {
            selectedCapability: hookData.selectedCapability,
            selectedComponentId: hookData.selectedComponentId,
          },
        });
      }

      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));
      restoreAllSizes();
    }
  };

  const components = {
    domains: DomainsSidebarPanel,
    visualization: VisualizationPanel,
    explorer: ExplorerPanel,
    details: DetailsPanel,
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, overflow: 'hidden' }} data-testid="business-domains-page">
      <div style={{
        height: '32px',
        flexShrink: 0,
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        padding: '0 12px',
        backgroundColor: 'var(--color-gray-50)',
        borderBottom: '1px solid var(--color-gray-200)',
        fontSize: '13px',
      }}>
        <span style={{ color: 'var(--color-gray-600)', fontWeight: 500 }}>View:</span>
        <button
          onClick={() => togglePanel('domains')}
          style={{
            padding: '4px 12px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '13px',
          }}
        >
          {panelVisibility.domains ? '☑' : '☐'} Business Domains
        </button>
        <button
          onClick={() => togglePanel('explorer')}
          style={{
            padding: '4px 12px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '13px',
          }}
        >
          {panelVisibility.explorer ? '☑' : '☐'} Capability Explorer
        </button>
        <button
          onClick={() => togglePanel('details')}
          style={{
            padding: '4px 12px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '13px',
          }}
        >
          {panelVisibility.details ? '☑' : '☐'} Details
        </button>
      </div>
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
