import { useEffect, useRef, useState, useCallback } from 'react';
import type { DockviewReadyEvent, DockviewApi } from 'dockview';
import type { useBusinessDomainsPage } from '../../hooks/useBusinessDomainsPage';
import { buildDomainsParams, buildVisualizationParams, buildExplorerParams, buildDetailsParams } from './panelParams';
import { useUserStore } from '../../../../store/userStore';

const LAYOUT_STORAGE_KEY = 'easi-business-domains-dockview-layout';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;
type PanelId = 'domains' | 'explorer' | 'details';
type PanelSizes = { domains: number; explorer: number; details: number };

const PANEL_DEFINITIONS: Record<PanelId, {
  component: string;
  title: string;
  buildParams: (hookData: BusinessDomainsHookReturn) => Record<string, unknown>;
  position: (api: DockviewApi) => { referencePanel: ReturnType<DockviewApi['getPanel']>; direction: string };
}> = {
  domains: {
    component: 'domains',
    title: 'Business Domains',
    buildParams: buildDomainsParams,
    position: (api) => ({ referencePanel: api.getPanel('visualization')!, direction: 'left' }),
  },
  explorer: {
    component: 'explorer',
    title: 'Capability Explorer',
    buildParams: buildExplorerParams,
    position: (api) => ({ referencePanel: api.getPanel('visualization')!, direction: 'right' }),
  },
  details: {
    component: 'details',
    title: 'Details',
    buildParams: buildDetailsParams,
    position: (api) => ({
      referencePanel: api.getPanel('explorer') ?? api.getPanel('visualization')!,
      direction: 'below',
    }),
  },
};

function addSidePanel(api: DockviewApi, panelId: PanelId, hookData: BusinessDomainsHookReturn) {
  const def = PANEL_DEFINITIONS[panelId];
  return api.addPanel({
    id: panelId,
    component: def.component,
    title: def.title,
    position: def.position(api) as never,
    params: def.buildParams(hookData),
  });
}

function snapshotPanelSizes(api: DockviewApi, sizesRef: React.MutableRefObject<PanelSizes>) {
  const domains = api.getPanel('domains');
  const explorer = api.getPanel('explorer');
  const details = api.getPanel('details');
  if (domains) sizesRef.current.domains = domains.api.width;
  if (explorer) sizesRef.current.explorer = explorer.api.width;
  if (details) sizesRef.current.details = details.api.height;
}

function restoreAllSizes(api: DockviewApi, sizes: PanelSizes) {
  setTimeout(() => {
    const domains = api.getPanel('domains');
    const explorer = api.getPanel('explorer');
    const details = api.getPanel('details');
    if (domains) domains.api.setSize({ width: sizes.domains });
    if (explorer) explorer.api.setSize({ width: sizes.explorer });
    if (details) details.api.setSize({ height: sizes.details });
  }, 0);
}

function useSyncPanelParams(
  dockviewApiRef: React.MutableRefObject<DockviewApi | null>,
  hookData: BusinessDomainsHookReturn,
  showExplorer: boolean,
) {
  useEffect(() => {
    const api = dockviewApiRef.current;
    if (!api) return;

    api.getPanel('domains')?.api.updateParameters(buildDomainsParams(hookData));
    api.getPanel('visualization')?.api.updateParameters(buildVisualizationParams(hookData));
    if (showExplorer) {
      api.getPanel('explorer')?.api.updateParameters(buildExplorerParams(hookData));
    }
    api.getPanel('details')?.api.updateParameters(buildDetailsParams(hookData));
  }, [hookData, showExplorer, dockviewApiRef]);
}

function removeExplorerPanel(api: DockviewApi, panelSizesRef: React.MutableRefObject<PanelSizes>) {
  const explorerPanel = api.getPanel('explorer');
  if (!explorerPanel) return;
  panelSizesRef.current.explorer = explorerPanel.api.width;
  api.removePanel(explorerPanel);
}

function restoreExplorerPanel(api: DockviewApi, hookData: BusinessDomainsHookReturn, panelSizesRef: React.MutableRefObject<PanelSizes>) {
  addSidePanel(api, 'explorer', hookData);
  setTimeout(() => {
    api.getPanel('explorer')?.api.setSize({ width: panelSizesRef.current.explorer });
  }, 0);
}

interface ExplorerSyncDeps {
  dockviewApiRef: React.MutableRefObject<DockviewApi | null>;
  panelSizesRef: React.MutableRefObject<PanelSizes>;
  setPanelVisibility: React.Dispatch<React.SetStateAction<{ domains: boolean; explorer: boolean; details: boolean }>>;
}

function useExplorerSync(
  deps: ExplorerSyncDeps,
  hookData: BusinessDomainsHookReturn,
  showExplorer: boolean,
  explorerVisible: boolean,
) {
  useEffect(() => {
    const api = deps.dockviewApiRef.current;
    if (!api) return;

    if (!showExplorer) {
      removeExplorerPanel(api, deps.panelSizesRef);
      deps.setPanelVisibility(prev => (prev.explorer ? { ...prev, explorer: false } : prev));
      return;
    }

    const needsRestore = explorerVisible && !api.getPanel('explorer') && !!api.getPanel('visualization');
    if (needsRestore) {
      restoreExplorerPanel(api, hookData, deps.panelSizesRef);
    }
  }, [hookData, explorerVisible, showExplorer, deps]);
}

function useLayoutPersistence(dockviewApiRef: React.MutableRefObject<DockviewApi | null>) {
  useEffect(() => {
    const api = dockviewApiRef.current;
    if (!api) return;

    const saveLayout = () => {
      const layout = api.toJSON();
      localStorage.setItem(LAYOUT_STORAGE_KEY, JSON.stringify(layout));
    };

    const disposable = api.onDidLayoutChange(saveLayout);
    return () => disposable.dispose();
  }, [dockviewApiRef]);
}

function canTogglePanel(api: DockviewApi | null, panelId: PanelId, showExplorer: boolean): api is DockviewApi {
  if (!api || !api.getPanel('visualization')) return false;
  return panelId !== 'explorer' || showExplorer;
}

function initializePanels(api: DockviewApi, hookData: BusinessDomainsHookReturn, showExplorer: boolean) {
  localStorage.removeItem(LAYOUT_STORAGE_KEY);

  const visualizationPanel = api.addPanel({
    id: 'visualization',
    component: 'visualization',
    title: 'Visualization',
    params: buildVisualizationParams(hookData),
  });

  const domainsPanel = addSidePanel(api, 'domains', hookData);
  const explorerPanel = showExplorer ? addSidePanel(api, 'explorer', hookData) : null;

  api.addPanel({
    id: 'details',
    component: 'details',
    title: 'Details',
    position: { referencePanel: explorerPanel ?? visualizationPanel, direction: 'below' },
    params: buildDetailsParams(hookData),
  });

  domainsPanel.api.setSize({ width: 320 });
  if (explorerPanel) explorerPanel.api.setSize({ width: 320 });
  api.getPanel('details')?.api.setSize({ height: 300 });
}

export function useDockviewLayout(hookData: BusinessDomainsHookReturn) {
  const userRole = useUserStore((state) => state.user?.role);
  const showExplorer = userRole !== 'stakeholder';
  const dockviewApiRef = useRef<DockviewApi | null>(null);
  const [panelVisibility, setPanelVisibility] = useState({ domains: true, explorer: showExplorer, details: true });
  const panelSizesRef = useRef<PanelSizes>({ domains: 320, explorer: 320, details: 300 });
  const explorerSyncDeps: ExplorerSyncDeps = { dockviewApiRef, panelSizesRef, setPanelVisibility };

  const onReady = useCallback((event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;
    initializePanels(event.api, hookData, showExplorer);
  }, [hookData, showExplorer]);

  useSyncPanelParams(dockviewApiRef, hookData, showExplorer);
  useExplorerSync(explorerSyncDeps, hookData, showExplorer, panelVisibility.explorer);
  useLayoutPersistence(dockviewApiRef);

  const togglePanel = useCallback((panelId: PanelId) => {
    const api = dockviewApiRef.current;
    if (!canTogglePanel(api, panelId, showExplorer)) return;

    snapshotPanelSizes(api, panelSizesRef);

    const panel = api.getPanel(panelId);
    if (panel) {
      api.removePanel(panel);
      setPanelVisibility(prev => ({ ...prev, [panelId]: false }));
    } else {
      addSidePanel(api, panelId, hookData);
      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));
    }

    restoreAllSizes(api, panelSizesRef.current);
  }, [hookData, showExplorer]);

  return {
    onReady,
    panelVisibility,
    togglePanel,
    showExplorer,
  };
}
