import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { createTestQueryClient, TestProviders } from '../../../test/helpers/renderWithProviders';
import { useConversations } from './useConversations';
import type { ReactNode } from 'react';
import type { QueryClient } from '@tanstack/react-query';

vi.mock('../api/chatApi', () => ({
  chatApi: {
    createConversation: vi.fn(),
    sendMessageStream: vi.fn(),
    listConversations: vi.fn(),
    getConversation: vi.fn(),
    deleteConversation: vi.fn(),
  },
}));

import { chatApi } from '../api/chatApi';

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <TestProviders withRouter={false} queryClient={queryClient}>
        {children}
      </TestProviders>
    );
  };
}

describe('useConversations', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = createTestQueryClient();
  });

  it('should list conversations from API', async () => {
    const mockConversations = [
      { id: 'conv-1', title: 'First chat', createdAt: '2026-01-01T00:00:00Z', _links: {} },
      { id: 'conv-2', title: 'Second chat', createdAt: '2026-01-02T00:00:00Z', _links: {} },
    ];
    vi.mocked(chatApi.listConversations).mockResolvedValue({
      data: mockConversations,
      _links: {},
    });

    const { result } = renderHook(() => useConversations(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.conversations).toHaveLength(2);
    });

    expect(result.current.conversations[0].title).toBe('First chat');
    expect(result.current.conversations[1].title).toBe('Second chat');
  });

  it('should return empty array when no conversations', async () => {
    vi.mocked(chatApi.listConversations).mockResolvedValue({ data: [], _links: {} });

    const { result } = renderHook(() => useConversations(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.conversations).toEqual([]);
  });

  it('should delete conversation and invalidate list', async () => {
    const mockConversations = [
      { id: 'conv-1', title: 'First chat', createdAt: '2026-01-01T00:00:00Z', _links: {} },
    ];
    vi.mocked(chatApi.listConversations).mockResolvedValue({ data: mockConversations, _links: {} });
    vi.mocked(chatApi.deleteConversation).mockResolvedValue();

    const { result } = renderHook(() => useConversations(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.conversations).toHaveLength(1);
    });

    vi.mocked(chatApi.listConversations).mockResolvedValue({ data: [], _links: {} });

    await act(async () => {
      result.current.deleteConversation('conv-1');
    });

    await waitFor(() => {
      expect(chatApi.deleteConversation).toHaveBeenCalledWith('conv-1');
    });

    await waitFor(() => {
      expect(result.current.conversations).toEqual([]);
    });
  });

  it('should invalidate conversation list on demand', async () => {
    vi.mocked(chatApi.listConversations).mockResolvedValue({ data: [], _links: {} });

    const { result } = renderHook(() => useConversations(), { wrapper: createWrapper(queryClient) });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const callCount = vi.mocked(chatApi.listConversations).mock.calls.length;

    vi.mocked(chatApi.listConversations).mockResolvedValue({
      data: [{ id: 'conv-new', title: 'New chat', createdAt: '2026-01-03T00:00:00Z', _links: {} }],
      _links: {},
    });

    await act(async () => {
      result.current.invalidateList();
    });

    await waitFor(() => {
      expect(vi.mocked(chatApi.listConversations).mock.calls.length).toBeGreaterThan(callCount);
    });
  });
});
