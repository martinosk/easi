import React, { lazy, Suspense, useState } from 'react';
import { EdgeTypeSelector, ColorSchemeSelector } from '../../features/views';
import { ImportButton } from '../../features/importing';
import { useBusinessDomains } from '../../features/business-domains';

const ImportDialog = lazy(() =>
  import('../../features/importing').then(module => ({ default: module.ImportDialog }))
);

interface PanelToggleProps {
  panelVisibility?: { navigation: boolean; details: boolean };
  onTogglePanel?: (panelId: 'navigation' | 'details') => void;
}

export const Toolbar: React.FC<PanelToggleProps> = ({ panelVisibility, onTogglePanel }) => {
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const { domains } = useBusinessDomains();

  return (
    <>
      <div className="toolbar">
        <div className="toolbar-left">
          {panelVisibility && onTogglePanel && (
            <div className="toolbar-panel-toggles">
              <button
                className={`toolbar-panel-toggle ${panelVisibility.navigation ? 'active' : ''}`}
                onClick={() => onTogglePanel('navigation')}
                aria-label="Toggle Explorer panel"
                aria-pressed={panelVisibility.navigation}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                  <line x1="9" y1="3" x2="9" y2="21" />
                </svg>
                Explorer
              </button>
              <button
                className={`toolbar-panel-toggle ${panelVisibility.details ? 'active' : ''}`}
                onClick={() => onTogglePanel('details')}
                aria-label="Toggle Details panel"
                aria-pressed={panelVisibility.details}
              >
                Details
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                  <line x1="15" y1="3" x2="15" y2="21" />
                </svg>
              </button>
            </div>
          )}
        </div>
        <div className="toolbar-right">
          <EdgeTypeSelector />
          <ColorSchemeSelector />
          <ImportButton onClick={() => setIsImportDialogOpen(true)} />
        </div>
      </div>
      {isImportDialogOpen && (
        <Suspense fallback={null}>
          <ImportDialog
            isOpen={isImportDialogOpen}
            onClose={() => setIsImportDialogOpen(false)}
            businessDomains={domains}
          />
        </Suspense>
      )}
    </>
  );
};
