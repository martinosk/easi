import React from 'react';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useUpdateViewEdgeType } from '../hooks/useViews';
import type { ViewId } from '../../../api/types';

export const EdgeTypeSelector: React.FC = () => {
  const { currentView, currentViewId } = useCurrentView();
  const updateEdgeTypeMutation = useUpdateViewEdgeType();

  const edgeType = currentView?.edgeType || 'default';

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    if (currentViewId) {
      updateEdgeTypeMutation.mutate({
        viewId: currentViewId as ViewId,
        request: { edgeType: event.target.value },
      });
    }
  };

  return (
    <div className="edge-type-selector">
      <label htmlFor="edge-type-select" className="selector-label">
        Edge Type
      </label>
      <select
        id="edge-type-select"
        className="form-select form-select-small"
        value={edgeType}
        onChange={handleChange}
        aria-label="Select edge type for relations"
        disabled={updateEdgeTypeMutation.isPending}
      >
        <option value="default">Bezier</option>
        <option value="step">Step</option>
        <option value="smoothstep">Smooth Step</option>
        <option value="straight">Straight</option>
      </select>
    </div>
  );
};
