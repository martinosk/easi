import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { EditGrant } from '../types';

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock('../api/editGrantApi', () => ({
  editGrantApi: {
    create: vi.fn(),
    getMyGrants: vi.fn().mockResolvedValue([]),
    getForArtifact: vi.fn().mockResolvedValue([]),
    revoke: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
import { editGrantApi } from '../api/editGrantApi';
import { useCreateEditGrant } from './useEditGrants';

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

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useCreateEditGrant', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    {
      scenario: 'invitation toast when invitationCreated is true',
      email: 'new@example.com',
      invitationCreated: true as boolean | undefined,
      expectedToast: 'Edit access granted. An invitation to join EASI was also created for new@example.com.',
    },
    {
      scenario: 'standard toast when invitationCreated is false',
      email: 'existing@example.com',
      invitationCreated: false as boolean | undefined,
      expectedToast: 'Edit access granted to existing@example.com',
    },
    {
      scenario: 'standard toast when invitationCreated is undefined',
      email: 'user@example.com',
      invitationCreated: undefined as boolean | undefined,
      expectedToast: 'Edit access granted to user@example.com',
    },
  ])('should show $scenario', async ({ email, invitationCreated, expectedToast }) => {
    const grant = createGrant({ granteeEmail: email, invitationCreated });
    vi.mocked(editGrantApi.create).mockResolvedValue(grant);

    const { result } = renderHook(() => useCreateEditGrant(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      granteeEmail: email,
      artifactType: 'capability',
      artifactId: 'cap-123',
    });

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith(expectedToast);
    });
  });

  it('should show error toast on failure', async () => {
    vi.mocked(editGrantApi.create).mockRejectedValue(new Error('Conflict'));

    const { result } = renderHook(() => useCreateEditGrant(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      granteeEmail: 'user@example.com',
      artifactType: 'capability',
      artifactId: 'cap-123',
    });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Conflict');
    });
  });
});
