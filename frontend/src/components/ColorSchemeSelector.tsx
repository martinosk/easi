import React from 'react';
import { useAppStore } from '../store/appStore';

export const ColorSchemeSelector: React.FC = () => {
  const currentView = useAppStore((state) => state.currentView);
  const setColorScheme = useAppStore((state) => state.setColorScheme);

  const colorScheme = currentView?.colorScheme || 'maturity';

  const handleChange = async (event: React.ChangeEvent<HTMLSelectElement>) => {
    await setColorScheme(event.target.value);
  };

  return (
    <div className="color-scheme-selector">
      <label htmlFor="color-scheme-select" className="selector-label">
        Color Scheme
      </label>
      <select
        id="color-scheme-select"
        className="form-select form-select-small"
        value={colorScheme}
        onChange={handleChange}
        aria-label="Select color scheme for canvas elements"
      >
        <option value="maturity">Maturity</option>
        <option value="archimate">Modern</option>
        <option value="archimate-classic">Classic</option>
        <option value="custom">Custom</option>
      </select>
    </div>
  );
};
