import React from 'react';
import { NativeSelect } from '@mantine/core';
import { useCurrentView } from '../hooks/useCurrentView';
import { useUpdateViewEdgeType } from '../hooks/useViews';

const EDGE_TYPE_OPTIONS = [
  { value: 'default', label: 'Bezier' },
  { value: 'step', label: 'Step' },
  { value: 'smoothstep', label: 'Smooth Step' },
  { value: 'straight', label: 'Straight' },
];

export const EdgeTypeSelector: React.FC = () => {
  const { currentView, currentViewId } = useCurrentView();
  const updateEdgeTypeMutation = useUpdateViewEdgeType();

  const edgeType = currentView?.edgeType || 'default';

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    if (currentViewId) {
      updateEdgeTypeMutation.mutate({
        viewId: currentViewId,
        request: { edgeType: event.target.value },
      });
    }
  };

  return (
    <NativeSelect
      id="edge-type-select"
      label="Edge Type"
      data={EDGE_TYPE_OPTIONS}
      value={edgeType}
      onChange={handleChange}
      aria-label="Select edge type for relations"
      disabled={updateEdgeTypeMutation.isPending}
      size="xs"
    />
  );
};
