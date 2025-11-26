import React from 'react';
import { useAppStore } from '../../../store/appStore';

export const LayoutDirectionSelector: React.FC = () => {
  const currentView = useAppStore((state) => state.currentView);
  const setLayoutDirection = useAppStore((state) => state.setLayoutDirection);

  const layoutDirection = currentView?.layoutDirection || 'TB';

  const handleChange = async (event: React.ChangeEvent<HTMLSelectElement>) => {
    await setLayoutDirection(event.target.value);
  };

  return (
    <div className="layout-direction-selector">
      <label htmlFor="layout-direction-select" className="selector-label">
        Layout Direction
      </label>
      <select
        id="layout-direction-select"
        className="form-select form-select-small"
        value={layoutDirection}
        onChange={handleChange}
        aria-label="Select layout direction for auto layout"
      >
        <option value="TB">Top to Bottom</option>
        <option value="LR">Left to Right</option>
        <option value="BT">Bottom to Top</option>
        <option value="RL">Right to Left</option>
      </select>
    </div>
  );
};
