import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InvitationsPage } from './InvitationsPage';
import { invitationApi } from '../api/invitationApi';
import type { Invitation, InvitationsListResponse } from '../types';

vi.mock('../api/invitationApi');
vi.mock('react-hot-toast', () => ({
  default: { error: vi.fn(), success: vi.fn() },
}));
vi.mock('../../../store/userStore', () => ({
  useUserStore: (selector: (state: { hasPermission: (p: string) => boolean }) => boolean) =>
    selector({ hasPermission: () => true }),
}));

const mockInvitations: Invitation[] = [
  {
    id: 'inv-1',
    email: 'pending1@acme.com',
    role: 'viewer',
    status: 'pending',
    invitedBy: 'admin@acme.com',
    createdAt: '2025-01-01T10:00:00Z',
    expiresAt: '2025-01-08T10:00:00Z',
    _links: { self: '/api/v1/invitations/inv-1', revoke: '/api/v1/invitations/inv-1/revoke' },
  },
  {
    id: 'inv-2',
    email: 'accepted@acme.com',
    role: 'editor',
    status: 'accepted',
    invitedBy: 'admin@acme.com',
    createdAt: '2025-01-02T10:00:00Z',
    expiresAt: '2025-01-09T10:00:00Z',
    acceptedAt: '2025-01-03T10:00:00Z',
    _links: { self: '/api/v1/invitations/inv-2' },
  },
  {
    id: 'inv-3',
    email: 'pending2@acme.com',
    role: 'viewer',
    status: 'pending',
    invitedBy: 'admin@acme.com',
    createdAt: '2025-01-03T10:00:00Z',
    expiresAt: '2025-01-10T10:00:00Z',
    _links: { self: '/api/v1/invitations/inv-3', revoke: '/api/v1/invitations/inv-3/revoke' },
  },
  {
    id: 'inv-4',
    email: 'expired@acme.com',
    role: 'viewer',
    status: 'expired',
    invitedBy: 'admin@acme.com',
    createdAt: '2024-12-01T10:00:00Z',
    expiresAt: '2024-12-08T10:00:00Z',
    _links: { self: '/api/v1/invitations/inv-4' },
  },
  {
    id: 'inv-5',
    email: 'revoked@acme.com',
    role: 'editor',
    status: 'revoked',
    invitedBy: 'admin@acme.com',
    createdAt: '2025-01-04T10:00:00Z',
    expiresAt: '2025-01-11T10:00:00Z',
    revokedAt: '2025-01-05T10:00:00Z',
    _links: { self: '/api/v1/invitations/inv-5' },
  },
];

const mockResponse: InvitationsListResponse = {
  data: mockInvitations,
  pagination: { hasMore: false, limit: 50 },
  _links: { self: '/api/v1/invitations' },
};

describe('InvitationsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(invitationApi.listInvitations).mockResolvedValue(mockResponse);
  });

  it('renders all invitations when filter is set to all', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    expect(screen.getAllByTestId(/invitation-row-/)).toHaveLength(5);
  });

  it('filters to show only pending invitations', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');
    fireEvent.change(filterSelect, { target: { value: 'pending' } });

    const rows = screen.getAllByTestId(/invitation-row-/);
    expect(rows).toHaveLength(2);
    expect(screen.getByTestId('invitation-row-inv-1')).toBeInTheDocument();
    expect(screen.getByTestId('invitation-row-inv-3')).toBeInTheDocument();
  });

  it('filters to show only accepted invitations', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');
    fireEvent.change(filterSelect, { target: { value: 'accepted' } });

    const rows = screen.getAllByTestId(/invitation-row-/);
    expect(rows).toHaveLength(1);
    expect(screen.getByTestId('invitation-row-inv-2')).toBeInTheDocument();
  });

  it('filters to show only expired invitations', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');
    fireEvent.change(filterSelect, { target: { value: 'expired' } });

    const rows = screen.getAllByTestId(/invitation-row-/);
    expect(rows).toHaveLength(1);
    expect(screen.getByTestId('invitation-row-inv-4')).toBeInTheDocument();
  });

  it('filters to show only revoked invitations', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');
    fireEvent.change(filterSelect, { target: { value: 'revoked' } });

    const rows = screen.getAllByTestId(/invitation-row-/);
    expect(rows).toHaveLength(1);
    expect(screen.getByTestId('invitation-row-inv-5')).toBeInTheDocument();
  });

  it('shows empty state when filter matches no invitations', async () => {
    vi.mocked(invitationApi.listInvitations).mockResolvedValue({
      ...mockResponse,
      data: mockInvitations.filter((i) => i.status === 'pending'),
    });

    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');
    fireEvent.change(filterSelect, { target: { value: 'revoked' } });

    expect(screen.queryByTestId('invitations-table')).not.toBeInTheDocument();
    expect(screen.getByText('No revoked invitations')).toBeInTheDocument();
  });

  it('returns to showing all invitations when filter is reset to all', async () => {
    render(<InvitationsPage />);

    await waitFor(() => {
      expect(screen.getByTestId('invitations-table')).toBeInTheDocument();
    });

    const filterSelect = screen.getByTestId('status-filter');

    fireEvent.change(filterSelect, { target: { value: 'pending' } });
    expect(screen.getAllByTestId(/invitation-row-/)).toHaveLength(2);

    fireEvent.change(filterSelect, { target: { value: 'all' } });
    expect(screen.getAllByTestId(/invitation-row-/)).toHaveLength(5);
  });
});
