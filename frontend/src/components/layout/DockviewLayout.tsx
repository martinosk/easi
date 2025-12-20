import React, { useCallback, useEffect, useRef, useState } from 'react';
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
import { ErrorBoundary, FeatureErrorFallback } from '../shared/ErrorBoundary';
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
      <ErrorBoundary
        fallback={(error, reset) => (
          <FeatureErrorFallback featureName="Navigation" error={error} onReset={reset} />
        )}
      >
        <NavigationTree
          onComponentSelect={props.params.onComponentSelect}
          onCapabilitySelect={props.params.onCapabilitySelect}
          onViewSelect={props.params.onViewSelect}
          onAddComponent={props.params.onAddComponent}
          onAddCapability={props.params.onAddCapability}
          onEditCapability={props.params.onEditCapability}
          onEditComponent={props.params.onEditComponent}
        />
      </ErrorBoundary>
    </div>
  );
};

const CanvasPanel = (props: IDockviewPanelProps<{
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
}>) => {
  return (
    <div style={{ height: '100%', width: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <ViewSelector />
      <div style={{ flex: 1, position: 'relative', overflow: 'hidden' }}>
        <ErrorBoundary
          fallback={(error, reset) => (
            <FeatureErrorFallback featureName="Canvas" error={error} onReset={reset} />
          )}
        >
          <ComponentCanvas
            ref={props.params.canvasRef}
            onConnect={props.params.onConnect}
            onComponentDrop={props.params.onComponentDrop}
          />
        </ErrorBoundary>
      </div>
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
      <ErrorBoundary
        fallback={(error, reset) => (
          <FeatureErrorFallback featureName="Details" error={error} onReset={reset} />
        )}
      >
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
      </ErrorBoundary>
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
  const [panelVisibility, setPanelVisibility] = useState({ navigation: true, details: true });
  const panelSizesRef = useRef<{ navigation: number; details: number }>({ navigation: 280, details: 350 });
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);

  const handleRemoveCapabilityFromView = useCallback(() => {
    if (selectedCapabilityId) {
      removeCapabilityFromCanvas(selectedCapabilityId);
    }
  }, [selectedCapabilityId, removeCapabilityFromCanvas]);

  const togglePanel = (panelId: 'navigation' | 'details') => {
    const api = dockviewApiRef.current;
    if (!api) return;

    const panel = api.getPanel(panelId);
    const navigationPanel = api.getPanel('navigation');
    const detailsPanel = api.getPanel('details');
    const canvasPanel = api.getPanel('canvas');

    if (navigationPanel) {
      panelSizesRef.current.navigation = navigationPanel.api.width;
    }
    if (detailsPanel) {
      panelSizesRef.current.details = detailsPanel.api.width;
    }

    if (panel) {
      api.removePanel(panel);
      setPanelVisibility(prev => ({ ...prev, [panelId]: false }));

      setTimeout(() => {
        const nav = api.getPanel('navigation');
        const details = api.getPanel('details');
        if (nav) nav.api.setSize({ width: panelSizesRef.current.navigation });
        if (details) details.api.setSize({ width: panelSizesRef.current.details });
      }, 0);
    } else {
      if (panelId === 'navigation') {
        api.addPanel({
          id: 'navigation',
          component: 'navigation',
          title: 'Explorer',
          position: { referencePanel: canvasPanel!, direction: 'left' },
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
      } else {
        api.addPanel({
          id: 'details',
          component: 'details',
          title: 'Details',
          position: { referencePanel: canvasPanel!, direction: 'right' },
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
      }

      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));

      setTimeout(() => {
        const nav = api.getPanel('navigation');
        const details = api.getPanel('details');
        if (nav) nav.api.setSize({ width: panelSizesRef.current.navigation });
        if (details) details.api.setSize({ width: panelSizesRef.current.details });
      }, 0);
    }
  };

  const onReady = (event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;

    const savedLayout = localStorage.getItem(LAYOUT_STORAGE_KEY);
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
        console.error('Failed to load saved layout, using default:', error);
        localStorage.removeItem(LAYOUT_STORAGE_KEY);
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
  }, [selectedNodeId, selectedEdgeId, selectedCapabilityId, onEditComponent, onEditRelation, onRemoveFromView, handleRemoveCapabilityFromView]);

  useEffect(() => {
    if (dockviewApiRef.current) {
      const canvasPanel = dockviewApiRef.current.getPanel('canvas');
      if (canvasPanel) {
        canvasPanel.api.updateParameters({
          canvasRef,
          onConnect,
          onComponentDrop,
        });
      }
    }
  }, [canvasRef, onConnect, onComponentDrop]);

  useEffect(() => {
    if (dockviewApiRef.current) {
      const navigationPanel = dockviewApiRef.current.getPanel('navigation');
      if (navigationPanel) {
        navigationPanel.api.updateParameters({
          onComponentSelect,
          onCapabilitySelect,
          onViewSelect,
          onAddComponent,
          onAddCapability,
          onEditCapability,
          onEditComponent,
        });
      }
    }
  }, [onComponentSelect, onCapabilitySelect, onViewSelect, onAddComponent, onAddCapability, onEditCapability, onEditComponent]);

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
        <span style={{ color: 'var(--color-gray-600)', fontWeight: 500 }}>Panels:</span>
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
