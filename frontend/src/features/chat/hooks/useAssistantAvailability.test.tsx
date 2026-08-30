import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import React from 'react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useUserStore } from '../../../store/userStore';
import { createTestQueryClient } from '../../../test/helpers';
import { seedAssistantStatus } from '../../../test/mocks/assistantStatus';
import type { SessionLinks } from '../../auth/types';
import { chatQueryKeys } from '../queryKeys';
import { useAssistantAvailability } from './useAssistantAvailability';

function renderAvailability(client: QueryClient) {
  const wrapper = ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
  return renderHook(() => useAssistantAvailability(), { wrapper });
}

function signInWith(links: Partial<SessionLinks>) {
  useUserStore.setState({
    sessionLinks: { self: '/api/v1/auth/sessions/current', logout: '', user: '', tenant: '', ...links },
  });
}

describe('useAssistantAvailability', () => {
  beforeEach(() => {
    useUserStore.setState({ sessionLinks: null });
  });

  it('offers the assistant when the session carries the link and the assistant is configured', async () => {
    seedAssistantStatus({ configured: true });
    signInWith({ 'x-assistant': '/api/v1/assistant/conversations' });

    const { result } = renderAvailability(createTestQueryClient());

    await waitFor(() => expect(result.current.assistantAvailable).toBe(true));
  });

  it('does not offer the assistant when the tenant has not configured it', async () => {
    seedAssistantStatus({ configured: false });
    signInWith({ 'x-assistant': '/api/v1/assistant/conversations' });
    const client = createTestQueryClient();

    const { result } = renderAvailability(client);

    await waitFor(() => expect(client.getQueryData(chatQueryKeys.status())).toBeDefined());
    expect(result.current.assistantAvailable).toBe(false);
  });

  it('does not request the status when the session lacks the assistant permission link', () => {
    seedAssistantStatus({ configured: true });
    signInWith({});
    const client = createTestQueryClient();

    const { result } = renderAvailability(client);

    expect(result.current.assistantAvailable).toBe(false);
    expect(client.getQueryState(chatQueryKeys.status())?.fetchStatus).not.toBe('fetching');
  });

  it('reports write availability only when the write link is present', async () => {
    seedAssistantStatus({ configured: true });
    signInWith({
      'x-assistant': '/api/v1/assistant/conversations',
      'x-assistant-write': '/api/v1/assistant/conversations',
    });

    const { result } = renderAvailability(createTestQueryClient());

    await waitFor(() => expect(result.current.assistantWriteAvailable).toBe(true));
  });

  it('withholds write availability when the write link is absent', async () => {
    seedAssistantStatus({ configured: true });
    signInWith({ 'x-assistant': '/api/v1/assistant/conversations' });

    const { result } = renderAvailability(createTestQueryClient());

    await waitFor(() => expect(result.current.assistantAvailable).toBe(true));
    expect(result.current.assistantWriteAvailable).toBe(false);
  });
});
