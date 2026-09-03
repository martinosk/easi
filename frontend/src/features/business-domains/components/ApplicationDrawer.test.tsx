import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildComponent, buildExpert, renderWithProviders, seedDb, server } from '../../../test/helpers';
import { ApplicationDrawer } from './ApplicationDrawer';

const API_BASE = 'http://localhost:8080';

function seedPhoenix() {
  const component = buildComponent({
    id: toComponentId('comp-1'),
    name: 'Phoenix',
    experts: [buildExpert({ name: 'Jane', role: 'Architect' })],
    _links: {
      self: { href: '/api/v1/components/comp-1', method: 'GET' },
      edit: { href: '/api/v1/components/comp-1', method: 'PUT' },
      'x-add-expert': { href: '/api/v1/components/comp-1/experts', method: 'POST' },
      'x-classify-hosting': { href: '/api/v1/components/comp-1/hosting', method: 'PUT' },
    },
  });
  seedDb({ components: [component], capabilities: [] });
  return component;
}

describe('ApplicationDrawer', () => {
  it('renders no content when no component is selected', () => {
    seedPhoenix();
    renderWithProviders(<ApplicationDrawer componentId={null} onClose={vi.fn()} />);

    expect(screen.queryByText('Phoenix')).not.toBeInTheDocument();
  });

  it('renders the application panel with its experts for the selected component', async () => {
    seedPhoenix();
    renderWithProviders(<ApplicationDrawer componentId={toComponentId('comp-1')} onClose={vi.fn()} />);

    expect(await screen.findByRole('heading', { name: 'Phoenix' })).toBeInTheDocument();
    expect(screen.getByText('Experts')).toBeInTheDocument();
    expect(screen.getByText('Jane', { exact: false })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '+ Add Expert' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('re-renders a hosting change from the detail query when the list did not carry the application', async () => {
    seedPhoenix();
    server.use(
      http.get(`${API_BASE}/api/v1/components`, () =>
        HttpResponse.json({ data: [], _links: { self: '/api/v1/components' } }),
      ),
    );
    renderWithProviders(<ApplicationDrawer componentId={toComponentId('comp-1')} onClose={vi.fn()} />);

    const select = await screen.findByTestId('hosting-select');
    expect(select).toHaveValue('Unknown');

    await userEvent.click(select);
    await userEvent.click(await screen.findByRole('option', { name: 'Cloud', hidden: true }));

    await waitFor(() => expect(screen.getByTestId('hosting-select')).toHaveValue('Cloud'));
  });
});
