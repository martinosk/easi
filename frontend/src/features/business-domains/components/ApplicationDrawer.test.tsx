import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildComponent, renderWithProviders } from '../../../test/helpers';
import { ApplicationDrawer } from './ApplicationDrawer';

const component = buildComponent({ id: toComponentId('comp-1'), name: 'Phoenix' });

vi.mock('../../components/hooks/useComponents', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../components/hooks/useComponents')>();
  return {
    ...actual,
    useComponents: () => ({ data: [component] }),
  };
});

vi.mock('../../capabilities/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ data: [] }),
  useCapabilitiesByComponent: () => ({ data: [] }),
}));

vi.mock('../hooks/useComponentDetails', () => ({
  useComponentDetails: () => ({ component: null, isLoading: false, error: null }),
}));

describe('ApplicationDrawer', () => {
  it('renders no content when no component is selected', () => {
    renderWithProviders(<ApplicationDrawer componentId={null} onClose={vi.fn()} />);

    expect(screen.queryByText('Phoenix')).not.toBeInTheDocument();
  });

  it('renders the component details for the selected component (resolved from the components store)', async () => {
    renderWithProviders(<ApplicationDrawer componentId={toComponentId('comp-1')} onClose={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('application-drawer')).toHaveTextContent('Phoenix'));
  });
});
