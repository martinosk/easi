import { NativeSelect, Tooltip } from '@mantine/core';
import { IconPalette } from '@tabler/icons-react';
import React from 'react';
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
    <Tooltip label="Colour scheme">
      <NativeSelect
        id="color-scheme-select"
        leftSection={<IconPalette size={14} stroke={1.75} />}
        data={COLOR_SCHEME_OPTIONS}
        value={colorScheme}
        onChange={handleChange}
        aria-label="Select color scheme for canvas elements"
        disabled={updateColorSchemeMutation.isPending}
        size="xs"
      />
    </Tooltip>
  );
};
