import { useEffect, useRef, useState, useCallback } from 'react';
import type { DockviewReadyEvent } from 'dockview';
import type { useBusinessDomainsPage } from '../../hooks/useBusinessDomainsPage';
import { buildDomainsParams, buildVisualizationParams, buildExplorerParams, buildDetailsParams } from './panelParams';
import { useUserStore } from '../../../../store/userStore';

const LAYOUT_STORAGE_KEY = 'easi-business-domains-dockview-layout';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;
type PanelId = 'domains' | 'explorer' | 'details';

export function useDockviewLayout(hookData: BusinessDomainsHookReturn) {
  const userRole = useUserStore((state) => state.user?.role);
  const showExplorer = userRole !== 'stakeholder';
  const dockviewApiRef = useRef<DockviewReadyEvent['api'] | null>(null);
  const [panelVisibility, setPanelVisibility] = useState({ domains: true, explorer: showExplorer, details: true });
  const panelSizesRef = useRef<{ domains: number; explorer: number; details: number }>({ domains: 320, explorer: 320, details: 300 });

  const onReady = useCallback((event: DockviewReadyEvent) => {
    dockviewApiRef.current = event.api;
    localStorage.removeItem(LAYOUT_STORAGE_KEY);

    const visualizationPanel = event.api.addPanel({
      id: 'visualization',
      component: 'visualization',
      title: 'Visualization',
      params: buildVisualizationParams(hookData),
    });

    const domainsPanel = event.api.addPanel({
      id: 'domains',
      component: 'domains',
      title: 'Business Domains',
      position: { referencePanel: visualizationPanel, direction: 'left' },
      params: buildDomainsParams(hookData),
    });

    const explorerPanel = showExplorer
      ? event.api.addPanel({
        id: 'explorer',
        component: 'explorer',
        title: 'Capability Explorer',
        position: { referencePanel: visualizationPanel, direction: 'right' },
        params: buildExplorerParams(hookData),
      })
      : null;

    const detailsPanel = event.api.addPanel({
      id: 'details',
      component: 'details',
      title: 'Details',
      position: { referencePanel: explorerPanel ?? visualizationPanel, direction: 'below' },
      params: buildDetailsParams(hookData),
    });

    domainsPanel.api.setSize({ width: 320 });
    if (explorerPanel) explorerPanel.api.setSize({ width: 320 });
    detailsPanel.api.setSize({ height: 300 });
  }, [showExplorer]);

  useEffect(() => {
    if (!dockviewApiRef.current) return;

    const api = dockviewApiRef.current;
    api.getPanel('domains')?.api.updateParameters(buildDomainsParams(hookData));
    api.getPanel('visualization')?.api.updateParameters(buildVisualizationParams(hookData));
    if (showExplorer) {
      api.getPanel('explorer')?.api.updateParameters(buildExplorerParams(hookData));
    }
    api.getPanel('details')?.api.updateParameters(buildDetailsParams(hookData));
  }, [hookData, showExplorer]);

  useEffect(() => {
    const api = dockviewApiRef.current;
    if (!api) return;

    if (!showExplorer) {
      const explorerPanel = api.getPanel('explorer');
      if (explorerPanel) {
        panelSizesRef.current.explorer = explorerPanel.api.width;
        api.removePanel(explorerPanel);
      }
      setPanelVisibility(prev => (prev.explorer ? { ...prev, explorer: false } : prev));
      return;
    }

    if (panelVisibility.explorer && !api.getPanel('explorer')) {
      const visualizationPanel = api.getPanel('visualization');
      if (!visualizationPanel) return;

      api.addPanel({
        id: 'explorer',
        component: 'explorer',
        title: 'Capability Explorer',
        position: { referencePanel: visualizationPanel, direction: 'right' },
        params: buildExplorerParams(hookData),
      });

      setTimeout(() => {
        api.getPanel('explorer')?.api.setSize({ width: panelSizesRef.current.explorer });
      }, 0);
    }
  }, [hookData, panelVisibility.explorer, showExplorer]);

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

  const togglePanel = useCallback((panelId: PanelId) => {
    const api = dockviewApiRef.current;
    if (!api) return;

    if (panelId === 'explorer' && !showExplorer) return;

    const panel = api.getPanel(panelId);
    const visualizationPanel = api.getPanel('visualization');
    const domainsPanel = api.getPanel('domains');
    const explorerPanel = api.getPanel('explorer');
    const detailsPanel = api.getPanel('details');
    if (!visualizationPanel) return;

    if (domainsPanel) panelSizesRef.current.domains = domainsPanel.api.width;
    if (explorerPanel) panelSizesRef.current.explorer = explorerPanel.api.width;
    if (detailsPanel) panelSizesRef.current.details = detailsPanel.api.height;

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
          params: buildDomainsParams(hookData),
        });
      } else if (panelId === 'explorer') {
        api.addPanel({
          id: 'explorer',
          component: 'explorer',
          title: 'Capability Explorer',
          position: { referencePanel: visualizationPanel, direction: 'right' },
          params: buildExplorerParams(hookData),
        });
      } else {
        api.addPanel({
          id: 'details',
          component: 'details',
          title: 'Details',
          position: { referencePanel: explorerPanel ?? visualizationPanel, direction: 'below' },
          params: buildDetailsParams(hookData),
        });
      }

      setPanelVisibility(prev => ({ ...prev, [panelId]: true }));
      restoreAllSizes();
    }
  }, [hookData, showExplorer]);

  return {
    onReady,
    panelVisibility,
    togglePanel,
    showExplorer,
  };
}
