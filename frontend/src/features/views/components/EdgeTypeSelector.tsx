import React from 'react';
import { useAppStore } from '../../../store/appStore';

export const EdgeTypeSelector: React.FC = () => {
  const currentView = useAppStore((state) => state.currentView);
  const setEdgeType = useAppStore((state) => state.setEdgeType);

  const edgeType = currentView?.edgeType || 'default';

  const handleChange = async (event: React.ChangeEvent<HTMLSelectElement>) => {
    await setEdgeType(event.target.value);
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
      >
        <option value="default">Bezier</option>
        <option value="step">Step</option>
        <option value="smoothstep">Smooth Step</option>
        <option value="straight">Straight</option>
      </select>
    </div>
  );
};
