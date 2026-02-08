import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EditGrantsList } from './EditGrantsList';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import type { EditGrant } from '../types';

vi.mock('../hooks/useEditGrants', () => ({
  useEditGrantsForArtifact: vi.fn(),
  useRevokeEditGrant: vi.fn(),
}));

import { useEditGrantsForArtifact, useRevokeEditGrant } from '../hooks/useEditGrants';

function createGrant(overrides: Partial<EditGrant> = {}): EditGrant {
  return {
    id: 'grant-1',
    grantorId: 'grantor-id',
    grantorEmail: 'grantor@example.com',
    granteeEmail: 'grantee@example.com',
    artifactType: 'capability',
    artifactId: 'cap-123',
    scope: 'write',
    status: 'active',
    createdAt: '2025-01-01T00:00:00Z',
    expiresAt: '2025-01-31T00:00:00Z',
    _links: {},
    ...overrides,
  };
}

describe('EditGrantsList', () => {
  const mockMutate = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useRevokeEditGrant).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      mutateAsync: vi.fn(),
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useRevokeEditGrant>);
  });

  function renderList() {
    return render(
      <MantineTestWrapper>
        <EditGrantsList artifactType="capability" artifactId="cap-123" />
      </MantineTestWrapper>
    );
  }

  describe('Loading state', () => {
    it('should show loading spinner while data is loading', () => {
      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
        isError: false,
        isPending: true,
        isSuccess: false,
        status: 'pending',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.getByTestId('edit-grants-loading')).toBeInTheDocument();
    });
  });

  describe('Empty state', () => {
    it('should show empty state when no grants exist', () => {
      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.getByText('No edit grants found')).toBeInTheDocument();
    });
  });

  describe('Grants table', () => {
    it('should render grants in a table', () => {
      const grants = [
        createGrant({ id: 'g1', granteeEmail: 'alice@example.com', grantorEmail: 'bob@example.com' }),
        createGrant({ id: 'g2', granteeEmail: 'charlie@example.com', grantorEmail: 'bob@example.com', status: 'revoked' }),
      ];

      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: grants,
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.getByTestId('edit-grants-table')).toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-g1')).toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-g2')).toBeInTheDocument();
      expect(screen.getByText('alice@example.com')).toBeInTheDocument();
      expect(screen.getByText('charlie@example.com')).toBeInTheDocument();
    });

    it('should display column headers', () => {
      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [createGrant()],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.getByText('Grantee')).toBeInTheDocument();
      expect(screen.getByText('Granted By')).toBeInTheDocument();
      expect(screen.getByText('Status')).toBeInTheDocument();
      expect(screen.getByText('Expires')).toBeInTheDocument();
      expect(screen.getByText('Actions')).toBeInTheDocument();
    });
  });

  describe('Status filtering', () => {
    const mixedGrants = [
      createGrant({ id: 'active-1', status: 'active', granteeEmail: 'active@example.com' }),
      createGrant({ id: 'revoked-1', status: 'revoked', granteeEmail: 'revoked@example.com' }),
      createGrant({ id: 'expired-1', status: 'expired', granteeEmail: 'expired@example.com' }),
    ];

    beforeEach(() => {
      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: mixedGrants,
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);
    });

    it('should render filter buttons for all statuses', () => {
      renderList();

      expect(screen.getByTestId('filter-all')).toBeInTheDocument();
      expect(screen.getByTestId('filter-active')).toBeInTheDocument();
      expect(screen.getByTestId('filter-revoked')).toBeInTheDocument();
      expect(screen.getByTestId('filter-expired')).toBeInTheDocument();
    });

    it('should show all grants by default', () => {
      renderList();

      expect(screen.getByTestId('edit-grant-row-active-1')).toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-revoked-1')).toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-expired-1')).toBeInTheDocument();
    });

    it('should filter to active grants only', () => {
      renderList();

      fireEvent.click(screen.getByTestId('filter-active'));

      expect(screen.getByTestId('edit-grant-row-active-1')).toBeInTheDocument();
      expect(screen.queryByTestId('edit-grant-row-revoked-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('edit-grant-row-expired-1')).not.toBeInTheDocument();
    });

    it('should filter to revoked grants only', () => {
      renderList();

      fireEvent.click(screen.getByTestId('filter-revoked'));

      expect(screen.queryByTestId('edit-grant-row-active-1')).not.toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-revoked-1')).toBeInTheDocument();
      expect(screen.queryByTestId('edit-grant-row-expired-1')).not.toBeInTheDocument();
    });

    it('should filter to expired grants only', () => {
      renderList();

      fireEvent.click(screen.getByTestId('filter-expired'));

      expect(screen.queryByTestId('edit-grant-row-active-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('edit-grant-row-revoked-1')).not.toBeInTheDocument();
      expect(screen.getByTestId('edit-grant-row-expired-1')).toBeInTheDocument();
    });

    it('should show empty state when filter matches no grants', () => {
      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [createGrant({ id: 'active-1', status: 'active' })],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      fireEvent.click(screen.getByTestId('filter-revoked'));

      expect(screen.getByText('No revoked edit grants')).toBeInTheDocument();
    });
  });

  describe('Revoke action', () => {
    it('should show revoke button when delete link is present', () => {
      const grant = createGrant({
        id: 'g1',
        _links: {
          delete: { href: '/api/v1/edit-grants/g1', method: 'DELETE' },
        },
      });

      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [grant],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.getByTestId('revoke-grant-g1')).toBeInTheDocument();
    });

    it('should not show revoke button when delete link is absent', () => {
      const grant = createGrant({ id: 'g1', status: 'revoked', _links: {} });

      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [grant],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      expect(screen.queryByTestId('revoke-grant-g1')).not.toBeInTheDocument();
    });

    it('should call revokeGrant.mutate with grant id when revoke is clicked', () => {
      const grant = createGrant({
        id: 'g1',
        _links: {
          delete: { href: '/api/v1/edit-grants/g1', method: 'DELETE' },
        },
      });

      vi.mocked(useEditGrantsForArtifact).mockReturnValue({
        data: [grant],
        isLoading: false,
        error: null,
        isError: false,
        isPending: false,
        isSuccess: true,
        status: 'success',
      } as unknown as ReturnType<typeof useEditGrantsForArtifact>);

      renderList();

      fireEvent.click(screen.getByTestId('revoke-grant-g1'));

      expect(mockMutate).toHaveBeenCalledWith('g1');
    });
  });
});
