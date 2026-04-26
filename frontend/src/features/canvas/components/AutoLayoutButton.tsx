import { Button } from '@mantine/core';
import React from 'react';
import { canEdit } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useAutoLayout } from '../hooks/useAutoLayout';
import { useCanvasNodes } from '../hooks/useCanvasNodes';

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

  const isDisabled = isLayouting || nodes.length === 0 || !currentViewId || !canEdit(currentView);

  return (
    <Button
      variant="default"
      leftSection={isLayouting ? <span className="spinner-small" aria-hidden="true" /> : <AutoLayoutIcon />}
      onClick={() => void applyAutoLayout()}
      disabled={isDisabled}
      aria-label="Auto layout canvas"
      aria-busy={isLayouting}
    >
      {isLayouting ? 'Layouting...' : 'Auto Layout'}
    </Button>
  );
};
