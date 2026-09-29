import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { ConcernStewardship, DomainStewardships } from '../types';

vi.mock('react-hot-toast', () => ({ default: { success: vi.fn(), error: vi.fn() } }));
vi.mock('../api/stewardshipApi', () => ({
  stewardshipApi: {
    getDomainStewardships: vi.fn(),
    assign: vi.fn(),
    release: vi.fn(),
  },
}));
const mockUseActiveUsers = vi.fn();
vi.mock('../../users/hooks/useUsers', () => ({
  useActiveUsers: (options: { enabled?: boolean }) => mockUseActiveUsers(options),
}));

import { stewardshipApi } from '../api/stewardshipApi';
import { StewardsDialog } from './StewardsDialog';

const HREF = '/api/v1/stewardships?domainId=ce';
const itemHref = (concern: string) => `/api/v1/stewardships/ce/${concern}`;

const LABELS = {
  ownership: 'Ownership',
  assessment: 'Assessment',
  documentation: 'Documentation',
  planning: 'Planning',
  structure: 'Structure',
} as const;

function concern(
  name: keyof typeof LABELS,
  { steward, writable }: { steward?: string; writable: boolean },
): ConcernStewardship {
  const links: ConcernStewardship['_links'] = { self: { href: itemHref(name), method: 'GET' } };
  if (writable) links['x-assign'] = { href: itemHref(name), method: 'PUT' };
  if (writable && steward) links['x-release'] = { href: itemHref(name), method: 'DELETE' };
  return {
    concern: name,
    label: LABELS[name],
    description: `${LABELS[name]} description`,
    steward: steward ? { id: `${steward.split(' ')[0].toLowerCase()}-id`, name: steward } : null,
    assignedBy: steward ? 'alice@example.com' : null,
    assignedAt: steward ? '2026-09-01T10:00:00Z' : null,
    _links: links,
  };
}

function stewardships({
  writable,
  fallback,
}: {
  writable: boolean;
  fallback?: DomainStewardships['fallback'];
}): DomainStewardships {
  return {
    domainId: 'ce',
    domainName: 'Customer Engagement',
    fallback: fallback === undefined ? { id: 'alice-id', name: 'Alice Smith' } : fallback,
    data: [
      concern('ownership', { steward: 'Mette Gram', writable }),
      concern('assessment', { writable }),
      concern('documentation', { writable }),
      concern('planning', { writable }),
      concern('structure', { writable }),
    ],
    _links: { self: { href: HREF, method: 'GET' } },
  };
}

function renderDialog() {
  renderWithProviders(
    <StewardsDialog
      target={{ domainId: 'ce', domainName: 'Customer Engagement', stewardshipsHref: HREF }}
      onClose={vi.fn()}
    />,
    { withRouter: false },
  );
}

describe('StewardsDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseActiveUsers.mockReturnValue({
      data: [
        { id: 'mette-id', name: 'Mette Gram', email: 'mette@example.com' },
        { id: 'jonas-id', name: 'Jonas Holm', email: 'jonas@example.com' },
      ],
    });
  });

  it('lists every concern with its steward or Unassigned and the fallback', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(stewardships({ writable: false }));

    renderDialog();

    expect(await screen.findByText('Mette Gram')).toBeInTheDocument();
    expect(stewardshipApi.getDomainStewardships).toHaveBeenCalledWith(HREF);
    for (const label of Object.values(LABELS)) {
      expect(screen.getByText(label)).toBeInTheDocument();
      expect(screen.getByText(`${label} description`)).toBeInTheDocument();
    }
    expect(screen.getAllByText('Unassigned')).toHaveLength(4);
    expect(screen.getAllByText('Falls back to Alice Smith')).toHaveLength(4);
  });

  it('names no fallback when the domain has no architect', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(
      stewardships({ writable: false, fallback: null }),
    );

    renderDialog();

    expect(await screen.findAllByText('Unassigned')).toHaveLength(4);
    expect(screen.queryByText(/Falls back to/)).not.toBeInTheDocument();
  });

  it('names an unknown fallback as Unknown user', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(
      stewardships({ writable: false, fallback: { id: 'ghost', name: null } }),
    );

    renderDialog();

    expect(await screen.findAllByText('Falls back to Unknown user')).toHaveLength(4);
  });

  it('shows readers no controls and requests no candidates', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(stewardships({ writable: false }));

    renderDialog();

    await screen.findByText('Mette Gram');
    expect(screen.queryByTestId('steward-select-assessment')).not.toBeInTheDocument();
    expect(screen.queryByTestId('release-steward-ownership')).not.toBeInTheDocument();
    expect(mockUseActiveUsers).toHaveBeenCalled();
    expect(mockUseActiveUsers).not.toHaveBeenCalledWith({ enabled: true });
  });

  it('lets writers assign a steward through the x-assign link', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(stewardships({ writable: true }));
    vi.mocked(stewardshipApi.assign).mockResolvedValue(
      concern('assessment', { steward: 'Jonas Holm', writable: true }),
    );

    renderDialog();

    const select = await screen.findByTestId('steward-select-assessment');
    await userEvent.click(select);
    const listbox = document.getElementById(select.getAttribute('aria-controls') ?? '') as HTMLElement;
    await userEvent.click(within(listbox).getByRole('option', { name: 'Jonas Holm', hidden: true }));

    await waitFor(() =>
      expect(stewardshipApi.assign).toHaveBeenCalledWith({ href: itemHref('assessment'), method: 'PUT' }, 'jonas-id'),
    );
    expect(mockUseActiveUsers).toHaveBeenCalledWith({ enabled: true });
  });

  it('lets writers release an assigned steward through the x-release link', async () => {
    vi.mocked(stewardshipApi.getDomainStewardships).mockResolvedValue(stewardships({ writable: true }));
    vi.mocked(stewardshipApi.release).mockResolvedValue();

    renderDialog();

    await userEvent.click(await screen.findByTestId('release-steward-ownership'));

    await waitFor(() =>
      expect(stewardshipApi.release).toHaveBeenCalledWith({ href: itemHref('ownership'), method: 'DELETE' }),
    );
    expect(screen.queryByTestId('release-steward-assessment')).not.toBeInTheDocument();
  });
});
