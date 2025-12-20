import React, { useEffect, useRef, useState } from 'react';
import { DockviewReact, DockviewDefaultTab } from 'dockview';
import type { DockviewReadyEvent, IDockviewPanelProps, IDockviewPanelHeaderProps } from 'dockview';
import { Toolbar } from './Toolbar';
import { NavigationTree } from '../../features/navigation';
import { ViewSelector } from '../../features/views';
import { ComponentCanvas, type ComponentCanvasRef } from '../../features/canvas';
import { ComponentDetails } from '../../features/components';
import { RelationDetails, RealizationDetails } from '../../features/relations';
import { CapabilityDetails } from '../../features/capabilities';
import { useAppStore } from '../../store/appStore';
import type { Capability } from '../../api/types';

const NonClosableTab = (props: IDockviewPanelHeaderProps) => {
  return <DockviewDefaultTab hideClose={true} {...props} />;
};

interface DockviewLayoutProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  onAddComponent: () => void;
  onAddCapability: () => void;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
  onComponentSelect: (componentId: string) => void;
  onCapabilitySelect: (capabilityId: string) => void;
  onViewSelect: (viewId: string) => Promise<void>;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onEditCapability: (capability: Capability) => void;
  onRemoveFromView: () => void;
}

const isRealizationEdge = (edgeId: string): boolean => edgeId.startsWith('realization-');
const isParentEdge = (edgeId: string): boolean => edgeId.startsWith('parent-');
const isRelationEdge = (edgeId: string): boolean => !isRealizationEdge(edgeId) && !isParentEdge(edgeId);

const LAYOUT_STORAGE_KEY = 'easi-canvas-dockview-layout';

const NavigationTreePanel = (props: IDockviewPanelProps<{
  onComponentSelect: (id: string) => void;
  onCapabilitySelect: (id: string) => void;
  onViewSelect: (id: string) => Promise<void>;
  onAddComponent: () => void;
  onAddCapability: () => void;
  onEditCapability: (capability: Capability) => void;
  onEditComponent: (componentId?: string) => void;
}>) => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
      <NavigationTree
        onComponentSelect={props.params.onComponentSelect}
        onCapabilitySelect={props.params.onCapabilitySelect}
        onViewSelect={props.params.onViewSelect}
        onAddComponent={props.params.onAddComponent}
        onAddCapability={props.params.onAddCapability}
        onEditCapability={props.params.onEditCapability}
        onEditComponent={props.params.onEditComponent}
      />
    </div>
  );
};

const ViewSelectorPanel = () => {
  return (
    <div style={{ height: '100%', width: '100%', overflow: 'auto' }}>
      <ViewSelector />
    </div>
  );
};

const CanvasPanel = (props: IDockviewPanelProps<{
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
}>) => {
  return (
    <div style={{ height: '100%', width: '100%', position: 'relative', overflow: 'hidden' }}>
      <ComponentCanvas
        ref={props.params.canvasRef}
        onConnect={props.params.onConnect}
        onComponentDrop={props.params.onComponentDrop}
      />
    </div>
  );
};

const DetailPanel = (props: IDockviewPanelProps<{
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  selectedCapabilityId: string | null;
  onEditComponent: (componentId?: string) => void;
  onEditRelation: () => void;
  onRemoveFromView: () => void;
  onRemoveCapabilityFromView: () => void;
}>) => {
  const { selectedNodeId, selectedEdgeId, selectedCapabilityId, onEditComponent, onEditRelation, onRemoveFromView, onRemoveCapabilityFromView } = props.params;

  return (
    <div style={{ height: '100%', width: '100%', overflow: 'auto', padding: '1rem' }}>
      {selectedNodeId ? (
        <ComponentDetails onEdit={onEditComponent} onRemoveFromView={onRemoveFromView} />
      ) : selectedEdgeId && isRealizationEdge(selectedEdgeId) ? (
        <RealizationDetails />
      ) : selectedEdgeId && isRelationEdge(selectedEdgeId) ? (
        <RelationDetails onEdit={onEditRelation} />
      ) : selectedCapabilityId ? (
        <CapabilityDetails onRemoveFromView={onRemoveCapabilityFromView} />
      ) : (
        <div style={{ color: 'var(--color-gray-500)' }}>Select a component, relation, or capability to view details</div>
      )}
    </div>
  );
};

export const DockviewLayout: React.FC<DockviewLayoutProps> = ({
  canvasRef,
  selectedNodeId,
  selectedEdgeId,
  onAddComponent,
  onAddCapability,
  onConnect,
  onComponentDrop,
  onComponentSelect,
  onCapabilitySelect,
  onViewSelect,
  onEditComponent,
  onEditRelation,
  onEditCapability,
  onRemoveFromView,
}) => {
  const dockviewApiRef = useRef<DockviewReadyEvent['api'] | null>(null);
  const [panelVisibility, setPanelVisibility] = useState({ navigation: true, views: true, details: true });
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  const handleRemoveCapabilityFromView = () => {
    if (selectedCapabilityId) {
      removeCapabilityFromCanvas(selectedCapabilityId);
    }
  };

  const togglePanel = (panelId: 'navigation' | 'views' | 'details') => {
    const api = dockviewApiRef.current;
    if (!api) return;

    const panel = api.getPanel(panelId);
    if (panel) {
      api.removePanel(panel);
      setPanelVisibility(prev => ({ ...prev, [panelId]: false }));
    } else {
      const canvasPanel = api.getPanel('canvas');
      if (!canvasPanel) return;

      if (panelId === 'navigation') {
        const newPanel = api.addPanel({
          id: 'navigation',
          component: 'navigation',
          title: 'Explorer',
          position: { referencePanel: canvasPanel, direction: 'left' },
          params: {
            onComponentSelect,
            onCapabilitySelect,
            onViewSelect,
            onAddComponent,
            onAddCapability,
            onEditCapability,
            onEditComponent,
          },
        });
        newPanel.api.setSize({ width: 280 });
      } else if (panelId === 'views') {
        const newPanel = api.addPanel({
          id: 'views',
          component: 'views',
          title: 'Views',
          position: { referencePanel: canvasPanel, direction: 'above' },
        });
        newPanel.api.setSize({ height: 40 });
      } else {
        const newPanel = api.addPanel({
          id: 'details',
          component: 'details',
          title: 'Details',
          position: { referencePanel: canvasPanel, direction: 'right' },
          params: {
            selectedNodeId,
            selectedEdgeId,
            selectedCapabilityId,
            onEditComponent,
            onEditRelation,
            onRemoveFromView,
            onRemoveCapabilityFromView: handleRemoveCapabilityFromView,
          },
        });
        newPanel.api.setSize({ width: 350 });
      }
      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));
    }
  };

  const onReady = (event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;

    // Clear any corrupted saved layout - remove this after layout is fixed
    localStorage.removeItem(LAYOUT_STORAGE_KEY);

    const savedLayout = null; // Temporarily disabled: localStorage.getItem(LAYOUT_STORAGE_KEY);
    if (savedLayout) {
      try {
        event.api.fromJSON(JSON.parse(savedLayout));

        event.api.getPanel('navigation')?.api.updateParameters({
          onComponentSelect,
          onCapabilitySelect,
          onViewSelect,
          onAddComponent,
          onAddCapability,
          onEditCapability,
          onEditComponent,
        });

        event.api.getPanel('canvas')?.api.updateParameters({
          canvasRef,
          onConnect,
          onComponentDrop,
        });

        event.api.getPanel('details')?.api.updateParameters({
          selectedNodeId,
          selectedEdgeId,
          selectedCapabilityId,
          onEditComponent,
          onEditRelation,
          onRemoveFromView,
          onRemoveCapabilityFromView: handleRemoveCapabilityFromView,
        });

        return;
      } catch (error) {
        console.error('Failed to load saved layout:', error);
      }
    }

    const canvasPanel = event.api.addPanel({
      id: 'canvas',
      component: 'canvas',
      title: 'Canvas',
      tabComponent: 'nonClosable',
      params: {
        canvasRef,
        onConnect,
        onComponentDrop,
      },
    });

    const viewsPanel = event.api.addPanel({
      id: 'views',
      component: 'views',
      title: 'Views',
      position: { referencePanel: canvasPanel, direction: 'above' },
    });

    const navigationPanel = event.api.addPanel({
      id: 'navigation',
      component: 'navigation',
      title: 'Explorer',
      position: { referencePanel: canvasPanel, direction: 'left' },
      params: {
        onComponentSelect,
        onCapabilitySelect,
        onViewSelect,
        onAddComponent,
        onAddCapability,
        onEditCapability,
        onEditComponent,
      },
    });

    const detailPanel = event.api.addPanel({
      id: 'details',
      component: 'details',
      title: 'Details',
      position: { referencePanel: canvasPanel, direction: 'right' },
      params: {
        selectedNodeId,
        selectedEdgeId,
        selectedCapabilityId,
        onEditComponent,
        onEditRelation,
        onRemoveFromView,
        onRemoveCapabilityFromView: handleRemoveCapabilityFromView,
      },
    });

    viewsPanel.api.setSize({ height: 40 });
    navigationPanel.api.setSize({ width: 280 });
    detailPanel.api.setSize({ width: 350 });
  };

  useEffect(() => {
    if (dockviewApiRef.current) {
      const detailPanel = dockviewApiRef.current.getPanel('details');
      if (detailPanel) {
        detailPanel.api.updateParameters({
          selectedNodeId,
          selectedEdgeId,
          selectedCapabilityId,
          onEditComponent,
          onEditRelation,
          onRemoveFromView,
          onRemoveCapabilityFromView: handleRemoveCapabilityFromView,
        });
      }
    }
  }, [selectedNodeId, selectedEdgeId, selectedCapabilityId, onEditComponent, onEditRelation, onRemoveFromView]);

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

  const components = {
    navigation: NavigationTreePanel,
    views: ViewSelectorPanel,
    canvas: CanvasPanel,
    details: DetailPanel,
  };

  const tabComponents = {
    nonClosable: NonClosableTab,
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, overflow: 'hidden' }}>
      <Toolbar />
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
          onClick={() => togglePanel('navigation')}
          style={{
            padding: '4px 12px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '13px',
          }}
        >
          {panelVisibility.navigation ? '☑' : '☐'} Explorer
        </button>
        <button
          onClick={() => togglePanel('views')}
          style={{
            padding: '4px 12px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '13px',
          }}
        >
          {panelVisibility.views ? '☑' : '☐'} Views
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
            tabComponents={tabComponents}
            className="dockview-theme-light"
          />
        </div>
      </div>
    </div>
  );
};
