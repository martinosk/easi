import React from 'react';
import { useCurrentView } from '../hooks/useCurrentView';
import { useUpdateViewColorScheme } from '../hooks/useViews';
export const ColorSchemeSelector: React.FC = () => {
  const { currentView, currentViewId } = useCurrentView();
  const updateColorSchemeMutation = useUpdateViewColorScheme();

  const colorScheme = currentView?.colorScheme || 'maturity';

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    if (currentViewId) {
      updateColorSchemeMutation.mutate({
        viewId: currentViewId,
        request: { colorScheme: event.target.value },
      });
    }
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
        disabled={updateColorSchemeMutation.isPending}
      >
        <option value="maturity">Maturity</option>
        <option value="classic">Classic</option>
        <option value="custom">Custom</option>
      </select>
    </div>
  );
};
