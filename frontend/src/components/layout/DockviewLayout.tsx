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
import { useRemoveCapabilityFromView } from '../../features/views/hooks/useViews';
import { useCurrentView } from '../../hooks/useCurrentView';
import type { Capability, CapabilityId } from '../../api/types';

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

type PanelId = 'navigation' | 'details';

interface PanelSizes {
  navigation: number;
  details: number;
}

function savePanelSizes(api: DockviewReadyEvent['api'], sizesRef: React.MutableRefObject<PanelSizes>) {
  const navigationPanel = api.getPanel('navigation');
  const detailsPanel = api.getPanel('details');
  if (navigationPanel) sizesRef.current.navigation = navigationPanel.api.width;
  if (detailsPanel) sizesRef.current.details = detailsPanel.api.width;
}

function restorePanelSizes(api: DockviewReadyEvent['api'], sizesRef: React.MutableRefObject<PanelSizes>) {
  setTimeout(() => {
    const nav = api.getPanel('navigation');
    const details = api.getPanel('details');
    if (nav) nav.api.setSize({ width: sizesRef.current.navigation });
    if (details) details.api.setSize({ width: sizesRef.current.details });
  }, 0);
}

const panelToggleButtonStyle = {
  padding: '4px 12px',
  border: '1px solid var(--color-gray-300)',
  borderRadius: '4px',
  backgroundColor: 'white',
  cursor: 'pointer',
  fontSize: '13px',
} as const;

const panelBarStyle = {
  height: '32px',
  flexShrink: 0,
  display: 'flex',
  alignItems: 'center',
  gap: '8px',
  padding: '0 12px',
  backgroundColor: 'var(--color-gray-50)',
  borderBottom: '1px solid var(--color-gray-200)',
  fontSize: '13px',
} as const;

interface PanelToggleBarProps {
  panelVisibility: { navigation: boolean; details: boolean };
  onToggle: (panelId: PanelId) => void;
}

function PanelToggleBar({ panelVisibility, onToggle }: PanelToggleBarProps) {
  return (
    <div style={panelBarStyle}>
      <span style={{ color: 'var(--color-gray-600)', fontWeight: 500 }}>Panels:</span>
      <button onClick={() => onToggle('navigation')} style={panelToggleButtonStyle}>
        {panelVisibility.navigation ? '☑' : '☐'} Explorer
      </button>
      <button onClick={() => onToggle('details')} style={panelToggleButtonStyle}>
        {panelVisibility.details ? '☑' : '☐'} Details
      </button>
    </div>
  );
}

function usePanelParams(props: DockviewLayoutProps) {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { currentViewId } = useCurrentView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();

  const handleRemoveCapabilityFromView = useCallback(() => {
    if (selectedCapabilityId && currentViewId) {
      removeCapabilityFromViewMutation.mutate({
        viewId: currentViewId,
        capabilityId: selectedCapabilityId as CapabilityId,
      });
    }
  }, [selectedCapabilityId, currentViewId, removeCapabilityFromViewMutation]);

  const navigation = useCallback(() => ({
    onComponentSelect: props.onComponentSelect,
    onCapabilitySelect: props.onCapabilitySelect,
    onViewSelect: props.onViewSelect,
    onAddComponent: props.onAddComponent,
    onAddCapability: props.onAddCapability,
    onEditCapability: props.onEditCapability,
    onEditComponent: props.onEditComponent,
  }), [props.onComponentSelect, props.onCapabilitySelect, props.onViewSelect, props.onAddComponent, props.onAddCapability, props.onEditCapability, props.onEditComponent]);

  const details = useCallback(() => ({
    selectedNodeId: props.selectedNodeId,
    selectedEdgeId: props.selectedEdgeId,
    selectedCapabilityId,
    onEditComponent: props.onEditComponent,
    onEditRelation: props.onEditRelation,
    onRemoveFromView: props.onRemoveFromView,
    onRemoveCapabilityFromView: handleRemoveCapabilityFromView,
  }), [props.selectedNodeId, props.selectedEdgeId, selectedCapabilityId, props.onEditComponent, props.onEditRelation, props.onRemoveFromView, handleRemoveCapabilityFromView]);

  const canvas = useCallback(() => ({
    canvasRef: props.canvasRef,
    onConnect: props.onConnect,
    onComponentDrop: props.onComponentDrop,
  }), [props.canvasRef, props.onConnect, props.onComponentDrop]);

  return { navigation, details, canvas };
}

function initializeDefaultLayout(
  api: DockviewReadyEvent['api'],
  panelParams: ReturnType<typeof usePanelParams>
) {
  const canvasPanel = api.addPanel({ id: 'canvas', component: 'canvas', title: 'Canvas', tabComponent: 'nonClosable', params: panelParams.canvas() });
  const navigationPanel = api.addPanel({ id: 'navigation', component: 'navigation', title: 'Explorer', position: { referencePanel: canvasPanel, direction: 'left' }, params: panelParams.navigation() });
  const detailPanel = api.addPanel({ id: 'details', component: 'details', title: 'Details', position: { referencePanel: canvasPanel, direction: 'right' }, params: panelParams.details() });
  navigationPanel.api.setSize({ width: 280 });
  detailPanel.api.setSize({ width: 350 });
}

function restoreSavedLayout(
  api: DockviewReadyEvent['api'],
  panelParams: ReturnType<typeof usePanelParams>
): boolean {
  const savedLayout = localStorage.getItem(LAYOUT_STORAGE_KEY);
  if (!savedLayout) return false;

  try {
    api.fromJSON(JSON.parse(savedLayout));
    api.getPanel('navigation')?.api.updateParameters(panelParams.navigation());
    api.getPanel('canvas')?.api.updateParameters(panelParams.canvas());
    api.getPanel('details')?.api.updateParameters(panelParams.details());
    return true;
  } catch {
    localStorage.removeItem(LAYOUT_STORAGE_KEY);
    return false;
  }
}

function usePanelSync(
  apiRef: React.RefObject<DockviewReadyEvent['api'] | null>,
  panelParams: ReturnType<typeof usePanelParams>
) {
  useEffect(() => { apiRef.current?.getPanel('details')?.api.updateParameters(panelParams.details()); }, [apiRef, panelParams]);
  useEffect(() => { apiRef.current?.getPanel('canvas')?.api.updateParameters(panelParams.canvas()); }, [apiRef, panelParams]);
  useEffect(() => { apiRef.current?.getPanel('navigation')?.api.updateParameters(panelParams.navigation()); }, [apiRef, panelParams]);
}

function useLayoutPersistence(apiRef: React.RefObject<DockviewReadyEvent['api'] | null>) {
  useEffect(() => {
    const api = apiRef.current;
    if (!api) return;
    const disposable = api.onDidLayoutChange(() => localStorage.setItem(LAYOUT_STORAGE_KEY, JSON.stringify(api.toJSON())));
    return () => disposable.dispose();
  }, [apiRef]);
}

function useDockviewLayout(props: DockviewLayoutProps) {
  const dockviewApiRef = useRef<DockviewReadyEvent['api'] | null>(null);
  const [panelVisibility, setPanelVisibility] = useState({ navigation: true, details: true });
  const panelSizesRef = useRef<PanelSizes>({ navigation: 280, details: 350 });
  const panelParams = usePanelParams(props);

  usePanelSync(dockviewApiRef, panelParams);
  useLayoutPersistence(dockviewApiRef);

  const togglePanel = useCallback((panelId: PanelId) => {
    const api = dockviewApiRef.current;
    if (!api) return;
    const panel = api.getPanel(panelId);
    savePanelSizes(api, panelSizesRef);

    if (panel) {
      api.removePanel(panel);
      setPanelVisibility(prev => ({ ...prev, [panelId]: false }));
    } else {
      const canvasPanel = api.getPanel('canvas');
      const isNav = panelId === 'navigation';
      api.addPanel({
        id: panelId,
        component: panelId,
        title: isNav ? 'Explorer' : 'Details',
        position: { referencePanel: canvasPanel!, direction: isNav ? 'left' : 'right' },
        params: isNav ? panelParams.navigation() : panelParams.details(),
      });
      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));
    }
    restorePanelSizes(api, panelSizesRef);
  }, [panelParams]);

  const onReady = useCallback((event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;
    if (!restoreSavedLayout(event.api, panelParams)) {
      initializeDefaultLayout(event.api, panelParams);
    }
  }, [panelParams]);

  return { panelVisibility, togglePanel, onReady };
}

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

const components = {
  navigation: NavigationTreePanel,
  canvas: CanvasPanel,
  details: DetailPanel,
};

const tabComponents = {
  nonClosable: NonClosableTab,
};

export const DockviewLayout: React.FC<DockviewLayoutProps> = (props) => {
  const { panelVisibility, togglePanel, onReady } = useDockviewLayout(props);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, overflow: 'hidden' }}>
      <Toolbar />
      <PanelToggleBar panelVisibility={panelVisibility} onToggle={togglePanel} />
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
