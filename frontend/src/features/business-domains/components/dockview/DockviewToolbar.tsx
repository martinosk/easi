import React from 'react';

interface PanelVisibility {
  domains: boolean;
  explorer: boolean;
  details: boolean;
}

interface DockviewToolbarProps {
  panelVisibility: PanelVisibility;
  onTogglePanel: (panelId: 'domains' | 'explorer' | 'details') => void;
  showExplorer: boolean;
}

const LeftPanelIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
    <line x1="9" y1="3" x2="9" y2="21" />
  </svg>
);

const RightPanelIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
    <line x1="15" y1="3" x2="15" y2="21" />
  </svg>
);

export const DockviewToolbar: React.FC<DockviewToolbarProps> = ({ panelVisibility, onTogglePanel, showExplorer }) => {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <div className="toolbar-panel-toggles">
          <button
            className={`toolbar-panel-toggle ${panelVisibility.domains ? 'active' : ''}`}
            onClick={() => onTogglePanel('domains')}
            aria-label="Toggle Business Domains panel"
            aria-pressed={panelVisibility.domains}
          >
            <LeftPanelIcon />
            Domains
          </button>
          {showExplorer && (
            <button
              className={`toolbar-panel-toggle ${panelVisibility.explorer ? 'active' : ''}`}
              onClick={() => onTogglePanel('explorer')}
              aria-label="Toggle Capability Explorer panel"
              aria-pressed={panelVisibility.explorer}
            >
              <LeftPanelIcon />
              Explorer
            </button>
          )}
          <button
            className={`toolbar-panel-toggle ${panelVisibility.details ? 'active' : ''}`}
            onClick={() => onTogglePanel('details')}
            aria-label="Toggle Details panel"
            aria-pressed={panelVisibility.details}
          >
            Details
            <RightPanelIcon />
          </button>
        </div>
      </div>
      <div className="toolbar-right" />
    </div>
  );
};
