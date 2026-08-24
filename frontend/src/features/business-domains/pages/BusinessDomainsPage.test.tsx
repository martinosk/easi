import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildHookData } from '../testkit/hookData';
import { BusinessDomainsPage } from './BusinessDomainsPage';

vi.mock('../hooks/useBusinessDomainsPage', () => ({
  useBusinessDomainsPage: () => buildHookData(),
}));

beforeEach(() => {
  localStorage.clear();
});

describe('BusinessDomainsPage view switching', () => {
  it('shows the board by default', () => {
    renderWithProviders(<BusinessDomainsPage />);
    expect(screen.getByTestId('domain-board')).toBeInTheDocument();
    expect(screen.queryByTestId('business-domains-map-view')).toBeNull();
  });

  it('switches to the map view and back through the toggle', async () => {
    const user = userEvent.setup();
    renderWithProviders(<BusinessDomainsPage />);

    await user.click(screen.getByRole('radio', { name: 'Map' }));
    expect(screen.getByTestId('business-domains-map-view')).toBeInTheDocument();
    expect(screen.queryByTestId('domain-board')).toBeNull();

    await user.click(screen.getByRole('radio', { name: 'Board' }));
    expect(screen.getByTestId('domain-board')).toBeInTheDocument();
  });

  it('restores the persisted map view on load', () => {
    localStorage.setItem('business-domains-view', 'map');
    renderWithProviders(<BusinessDomainsPage />);
    expect(screen.getByTestId('business-domains-map-view')).toBeInTheDocument();
  });

  it('switches to the timeline view through the toggle', async () => {
    const user = userEvent.setup();
    renderWithProviders(<BusinessDomainsPage />);

    await user.click(screen.getByRole('radio', { name: 'Timeline' }));
    expect(screen.getByTestId('business-domains-timeline-view')).toBeInTheDocument();
    expect(screen.queryByTestId('domain-board')).toBeNull();

    await user.click(screen.getByRole('radio', { name: 'Board' }));
    expect(screen.getByTestId('domain-board')).toBeInTheDocument();
  });

  it('restores the timeline view from a deep link', () => {
    renderWithProviders(<BusinessDomainsPage />, {
      routerProps: { initialEntries: ['/business-domains?presentation=timeline'] },
    });
    expect(screen.getByTestId('business-domains-timeline-view')).toBeInTheDocument();
  });
});
