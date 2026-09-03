import { ActionIcon, Tooltip } from '@mantine/core';
import {
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
  IconLayoutSidebarRightCollapse,
  IconLayoutSidebarRightExpand,
} from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';
import type { Capability } from '../../api/types';
import { ComponentCanvas, type ComponentCanvasRef } from '../../features/canvas';
import { CANVAS_COMMANDS_SLOT_ID } from '../../features/canvas/components/CanvasCommandsPortal';
import { NavigationTree } from '../../features/navigation';
import { ViewSelector } from '../../features/views';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { useRemoveCapabilityFromView } from '../../features/views/hooks/useViews';
import { useAppStore } from '../../store/appStore';
import { DetailContentRendererWithPlaceholder } from '../shared/DetailContentRenderer';
import { ErrorBoundary, FeatureErrorFallback } from '../shared/ErrorBoundary';
import { CanvasViewControls } from './CanvasViewControls';
import classes from './CanvasWorkspace.module.css';

const PANEL_VISIBILITY_STORAGE_KEY = 'easi-canvas-panels';

export type PanelId = 'navigation' | 'details';

export interface PanelVisibility {
  navigation: boolean;
  details: boolean;
}

const DEFAULT_PANEL_VISIBILITY: PanelVisibility = { navigation: true, details: true };

function loadPanelVisibility(): PanelVisibility {
  const stored = localStorage.getItem(PANEL_VISIBILITY_STORAGE_KEY);
  if (!stored) return DEFAULT_PANEL_VISIBILITY;
  try {
    const parsed = JSON.parse(stored) as Partial<PanelVisibility>;
    return {
      navigation: parsed.navigation ?? DEFAULT_PANEL_VISIBILITY.navigation,
      details: parsed.details ?? DEFAULT_PANEL_VISIBILITY.details,
    };
  } catch {
    return DEFAULT_PANEL_VISIBILITY;
  }
}

function usePanelVisibility() {
  const [panelVisibility, setPanelVisibility] = useState<PanelVisibility>(loadPanelVisibility);

  useEffect(() => {
    localStorage.setItem(PANEL_VISIBILITY_STORAGE_KEY, JSON.stringify(panelVisibility));
  }, [panelVisibility]);

  const togglePanel = useCallback((panelId: PanelId) => {
    setPanelVisibility((prev) => ({ ...prev, [panelId]: !prev[panelId] }));
  }, []);

  return { panelVisibility, togglePanel };
}

interface PaneHeaderProps {
  title: string;
  collapseLabel: string;
  collapseIcon: React.ReactNode;
  onCollapse: () => void;
}

function PaneHeader({ title, collapseLabel, collapseIcon, onCollapse }: PaneHeaderProps) {
  return (
    <div className={classes.paneHeader}>
      {title}
      <ActionIcon variant="subtle" color="gray" size="sm" aria-label={collapseLabel} onClick={onCollapse}>
        {collapseIcon}
      </ActionIcon>
    </div>
  );
}

interface ReopenPaneButtonProps {
  label: string;
  icon: React.ReactNode;
  className?: string;
  onClick: () => void;
}

function ReopenPaneButton({ label, icon, className, onClick }: ReopenPaneButtonProps) {
  return (
    <Tooltip label={label.replace(' panel', '')}>
      <ActionIcon variant="subtle" color="gray" className={className} aria-label={label} onClick={onClick}>
        {icon}
      </ActionIcon>
    </Tooltip>
  );
}

interface PaneToggleProps {
  panelVisibility: PanelVisibility;
  togglePanel: (panelId: PanelId) => void;
}

function CanvasHeaderBar({ panelVisibility, togglePanel }: PaneToggleProps) {
  return (
    <div className={classes.canvasHeader}>
      {!panelVisibility.navigation && (
        <ReopenPaneButton
          label="Show explorer panel"
          icon={<IconLayoutSidebarLeftExpand size={16} stroke={1.75} />}
          className={classes.headerLeading}
          onClick={() => togglePanel('navigation')}
        />
      )}
      <div className={classes.viewSelectorSlot}>
        <ViewSelector />
      </div>
      <div id={CANVAS_COMMANDS_SLOT_ID} className={classes.canvasCommandsSlot} />
      <CanvasViewControls />
      {!panelVisibility.details && (
        <ReopenPaneButton
          label="Show details panel"
          icon={<IconLayoutSidebarRightExpand size={16} stroke={1.75} />}
          onClick={() => togglePanel('details')}
        />
      )}
    </div>
  );
}

interface CanvasWorkspaceProps {
  canvasRef: React.RefObject<ComponentCanvasRef | null>;
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  onAddComponent?: () => void;
  onAddCapability?: () => void;
  canCreateView?: boolean;
  canCreateOriginEntity?: boolean;
  onConnect: (source: string, target: string) => void;
  onComponentDrop: (componentId: string, x: number, y: number) => Promise<void>;
  onComponentSelect: (componentId: string) => void;
  onCapabilitySelect: (capabilityId: string) => void;
  onOriginEntitySelect?: (nodeId: string) => void;
  onViewSelect: (viewId: string) => Promise<void>;
  onEditRelation: () => void;
  onEditCapability: (capability: Capability) => void;
  onRemoveFromView: () => void;
}

function ExplorerPane({ props, onCollapse }: { props: CanvasWorkspaceProps; onCollapse: () => void }) {
  return (
    <div className={classes.explorerPane} data-testid="explorer-pane">
      <PaneHeader
        title="Explorer"
        collapseLabel="Hide explorer panel"
        collapseIcon={<IconLayoutSidebarLeftCollapse size={16} stroke={1.75} />}
        onCollapse={onCollapse}
      />
      <ErrorBoundary
        fallback={(error, reset) => <FeatureErrorFallback featureName="Navigation" error={error} onReset={reset} />}
      >
        <NavigationTree
          onComponentSelect={props.onComponentSelect}
          onCapabilitySelect={props.onCapabilitySelect}
          onOriginEntitySelect={props.onOriginEntitySelect}
          onViewSelect={props.onViewSelect}
          onAddComponent={props.onAddComponent}
          onAddCapability={props.onAddCapability}
          onEditCapability={props.onEditCapability}
          onEditComponent={props.onComponentSelect}
          canCreateView={props.canCreateView}
          canCreateOriginEntity={props.canCreateOriginEntity}
        />
      </ErrorBoundary>
    </div>
  );
}

function DetailsPane({ props, onCollapse }: { props: CanvasWorkspaceProps; onCollapse: () => void }) {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { currentViewId } = useCurrentView();
  const removeCapabilityFromViewMutation = useRemoveCapabilityFromView();

  const handleRemoveCapabilityFromView = useCallback(() => {
    if (selectedCapabilityId && currentViewId) {
      removeCapabilityFromViewMutation.mutate({ viewId: currentViewId, capabilityId: selectedCapabilityId });
    }
  }, [selectedCapabilityId, currentViewId, removeCapabilityFromViewMutation]);

  return (
    <div className={classes.detailsPane} data-testid="details-pane">
      <PaneHeader
        title="Details"
        collapseLabel="Hide details panel"
        collapseIcon={<IconLayoutSidebarRightCollapse size={16} stroke={1.75} />}
        onCollapse={onCollapse}
      />
      <div className={classes.detailsContent}>
        <ErrorBoundary
          fallback={(error, reset) => <FeatureErrorFallback featureName="Details" error={error} onReset={reset} />}
        >
          <DetailContentRendererWithPlaceholder
            selectedNodeId={props.selectedNodeId}
            selectedEdgeId={props.selectedEdgeId}
            selectedCapabilityId={selectedCapabilityId}
            onEditRelation={props.onEditRelation}
            onRemoveFromView={props.onRemoveFromView}
            onRemoveCapabilityFromView={handleRemoveCapabilityFromView}
          />
        </ErrorBoundary>
      </div>
    </div>
  );
}

export const CanvasWorkspace: React.FC<CanvasWorkspaceProps> = (props) => {
  const { panelVisibility, togglePanel } = usePanelVisibility();

  return (
    <div className={classes.workspace}>
      <div className={classes.body}>
        {panelVisibility.navigation && <ExplorerPane props={props} onCollapse={() => togglePanel('navigation')} />}
        <div className={classes.canvasPane} data-testid="canvas-pane">
          <CanvasHeaderBar panelVisibility={panelVisibility} togglePanel={togglePanel} />
          <div className={classes.canvasFlowWrapper}>
            <ErrorBoundary
              fallback={(error, reset) => <FeatureErrorFallback featureName="Canvas" error={error} onReset={reset} />}
            >
              <ComponentCanvas
                ref={props.canvasRef}
                onConnect={props.onConnect}
                onComponentDrop={props.onComponentDrop}
              />
            </ErrorBoundary>
          </div>
        </div>
        {panelVisibility.details && <DetailsPane props={props} onCollapse={() => togglePanel('details')} />}
      </div>
    </div>
  );
};
