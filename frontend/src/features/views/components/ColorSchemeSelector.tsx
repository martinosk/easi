import React from 'react';
import { NativeSelect } from '@mantine/core';
import { useCurrentView } from '../hooks/useCurrentView';
import { useUpdateViewColorScheme } from '../hooks/useViews';

const COLOR_SCHEME_OPTIONS = [
  { value: 'maturity', label: 'Maturity' },
  { value: 'classic', label: 'Classic' },
  { value: 'custom', label: 'Custom' },
];

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
    <NativeSelect
      id="color-scheme-select"
      label="Color Scheme"
      data={COLOR_SCHEME_OPTIONS}
      value={colorScheme}
      onChange={handleChange}
      aria-label="Select color scheme for canvas elements"
      disabled={updateColorSchemeMutation.isPending}
      size="xs"
    />
  );
};
