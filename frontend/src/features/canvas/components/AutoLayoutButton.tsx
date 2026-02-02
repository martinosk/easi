import React from 'react';
import { useAutoLayout } from '../hooks/useAutoLayout';
import { useCanvasNodes } from '../hooks/useCanvasNodes';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { canEdit } from '../../../utils/hateoas';

export const AutoLayoutButton: React.FC = () => {
  const { applyAutoLayout, isLayouting } = useAutoLayout();
  const nodes = useCanvasNodes();
  const { currentView, currentViewId } = useCurrentView();

  const isDisabled = isLayouting || nodes.length === 0 || !currentViewId || !canEdit(currentView);

  return (
    <div className="canvas-auto-layout">
      <button
        type="button"
        className="auto-layout-button"
        onClick={applyAutoLayout}
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
  );
};