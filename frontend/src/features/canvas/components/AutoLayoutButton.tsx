import { Button } from '@mantine/core';
import React, { useState } from 'react';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAutoLayout } from '../hooks/useAutoLayout';
import { useCanvasNodes } from '../hooks/useCanvasNodes';

const AUTO_LAYOUT_WARNING = 'Auto layout is an experimental feature that will completely re-arrange your view.';

const AutoLayoutIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
    <rect x="3" y="3" width="7" height="7" rx="1" />
    <rect x="14" y="3" width="7" height="7" rx="1" />
    <rect x="3" y="14" width="7" height="7" rx="1" />
    <rect x="14" y="14" width="7" height="7" rx="1" />
  </svg>
);

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
      <Button
        variant="default"
        leftSection={isLayouting ? <span className="spinner-small" aria-hidden="true" /> : <AutoLayoutIcon />}
        onClick={handleAutoLayout}
        disabled={isDisabled}
        aria-label="Auto layout canvas"
        aria-busy={isLayouting}
      >
        {isLayouting ? 'Layouting...' : 'Auto Layout'}
      </Button>

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
