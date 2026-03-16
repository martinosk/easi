import React, { useState } from 'react';
import { useAutoLayout } from '../hooks/useAutoLayout';
import { useCanvasNodes } from '../hooks/useCanvasNodes';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { canEdit } from '../../../utils/hateoas';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';

const AUTO_LAYOUT_WARNING = 'Auto layout is an experimental feature that will completely re-arrange your view.';

export const AutoLayoutButton: React.FC = () => {
  const { applyAutoLayout, isLayouting } = useAutoLayout();
  const nodes = useCanvasNodes();
  const { currentView, currentViewId } = useCurrentView();
  const [isWarningOpen, setIsWarningOpen] = useState(false);

  const isDisabled = isLayouting || nodes.length === 0 || !currentViewId || !canEdit(currentView);

  const handleAutoLayout = () => {
    setIsWarningOpen(true);
  };

  const handleCancelAutoLayout = () => {
    setIsWarningOpen(false);
  };

  const handleConfirmAutoLayout = () => {
    setIsWarningOpen(false);
    void applyAutoLayout();
  };

  return (
    <>
      <div className="canvas-auto-layout">
        <button
          type="button"
          className="auto-layout-button"
          onClick={handleAutoLayout}
          disabled={isDisabled}
          aria-label="Auto layout canvas"
          aria-busy={isLayouting}
        >
          {isLayouting ? (
            <>
              <span className="spinner-small" aria-hidden="true" />
              Layouting...
            </>
          ) : (
            <>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <rect x="3" y="3" width="7" height="7" rx="1" />
                <rect x="14" y="3" width="7" height="7" rx="1" />
                <rect x="3" y="14" width="7" height="7" rx="1" />
                <rect x="14" y="14" width="7" height="7" rx="1" />
              </svg>
              Auto Layout
            </>
          )}
        </button>
      </div>

      {isWarningOpen && (
        <ConfirmationDialog
          title="Auto Layout"
          message={AUTO_LAYOUT_WARNING}
          confirmText="OK"
          onConfirm={handleConfirmAutoLayout}
          onCancel={handleCancelAutoLayout}
        />
      )}
    </>
  );
};