import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId, HATEOASLinks } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { ComponentOwnershipSection } from './ComponentOwnershipSection';

vi.mock('../api', () => ({
  componentsApi: {
    nominateOwner: vi.fn(),
    confirmOwnership: vi.fn(),
    assignOwner: vi.fn(),
    clearOwnership: vi.fn(),
    getStatistics: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('../../users/hooks/useUsers', () => ({
  useActiveUsers: () => ({
    data: [{ id: 'u-1', name: 'Alice Smith', email: 'alice@example.com', role: 'architect', status: 'active' }],
  }),
}));

vi.mock('../../origin-entities/hooks/useInternalTeams', () => ({
  useInternalTeamsQuery: () => ({
    data: [{ id: 't-1', name: 'Platform Ops', componentCount: 0, createdAt: '2024-01-01', _links: {} }],
  }),
}));

import { componentsApi } from '../api';

const ownershipLinks = (rels: string[]): HATEOASLinks => {
  const links: HATEOASLinks = { self: { href: '/api/v1/components/comp-1', method: 'GET' } };
  for (const rel of rels) {
    links[rel] = { href: `/api/v1/components/comp-1/ownership`, method: 'POST' };
  }
  return links;
};

describe('ComponentOwnershipSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows the Unknown state with nominate and assign actions', () => {
    const component = buildComponent({
      id: 'comp-1' as ComponentId,
      ownershipState: 'unknown',
      hosting: 'unknown',
      _links: ownershipLinks(['x-nominate-owner', 'x-assign-owner']),
    });

    renderWithProviders(<ComponentOwnershipSection component={component} />, { withRouter: false });

    expect(screen.getByTestId('ownership-state-badge')).toHaveTextContent('Unknown');
    expect(screen.getByRole('button', { name: 'Nominate Owner' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Assign Owner' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Confirm' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Clear' })).not.toBeInTheDocument();
  });

  const directActionCases = [
    {
      name: 'confirms a nominated candidate',
      state: 'nominated' as const,
      badge: 'Nominated',
      links: ['x-confirm-owner', 'x-clear-owner'],
      button: 'Confirm',
      arrange: () => vi.mocked(componentsApi.confirmOwnership).mockResolvedValue(undefined as never),
      calledApi: () => componentsApi.confirmOwnership,
    },
    {
      name: 'clears ownership from the owned state',
      state: 'owned' as const,
      badge: 'Owned',
      links: ['x-clear-owner'],
      button: 'Clear',
      arrange: () => vi.mocked(componentsApi.clearOwnership).mockResolvedValue(undefined),
      calledApi: () => componentsApi.clearOwnership,
    },
  ];

  for (const tc of directActionCases) {
    it(tc.name, async () => {
      const component = buildComponent({
        id: 'comp-1' as ComponentId,
        ownershipState: tc.state,
        owner: { kind: 'user', id: 'u-1', name: 'Alice Smith' },
        _links: ownershipLinks(tc.links),
      });
      tc.arrange();

      renderWithProviders(<ComponentOwnershipSection component={component} />, { withRouter: false });

      expect(screen.getByTestId('ownership-state-badge')).toHaveTextContent(tc.badge);
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();

      await userEvent.click(screen.getByRole('button', { name: tc.button }));

      await waitFor(() => {
        expect(tc.calledApi()).toHaveBeenCalledWith(component);
      });
    });
  }

  it('renders no action buttons without affordance links', () => {
    const component = buildComponent({
      id: 'comp-1' as ComponentId,
      ownershipState: 'managed',
      owner: { kind: 'team', id: 't-1', name: 'Platform Ops' },
      _links: ownershipLinks([]),
    });

    renderWithProviders(<ComponentOwnershipSection component={component} />, { withRouter: false });

    expect(screen.getByTestId('ownership-state-badge')).toHaveTextContent('Managed');
    expect(screen.getByText('Platform Ops')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('opens the owner dialog when nominating', async () => {
    const component = buildComponent({
      id: 'comp-1' as ComponentId,
      ownershipState: 'unknown',
      hosting: 'unknown',
      _links: ownershipLinks(['x-nominate-owner']),
    });

    renderWithProviders(<ComponentOwnershipSection component={component} />, { withRouter: false });

    await userEvent.click(screen.getByRole('button', { name: 'Nominate Owner' }));

    expect(await screen.findByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Nominate Owner', { selector: 'h2' })).toBeInTheDocument();
  });
});
