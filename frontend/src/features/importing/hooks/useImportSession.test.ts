import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ImportSession, CreateImportSessionRequest } from '../types';

const { mockPost, mockGet, mockDelete } = vi.hoisted(() => ({
  mockPost: vi.fn(),
  mockGet: vi.fn(),
  mockDelete: vi.fn(),
}));

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      post: mockPost,
      get: mockGet,
      delete: mockDelete,
      interceptors: {
        response: { use: vi.fn() },
      },
    })),
  },
}));

import { useImportSession } from './useImportSession';

describe('useImportSession', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPost.mockReset();
    mockGet.mockReset();
    mockDelete.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('createSession', () => {
    it('should upload file and create import session', async () => {
      const mockSession: ImportSession = {
        id: 'import-123' as any,
        status: 'pending',
        sourceFormat: 'archimate-openexchange',
        preview: {
          supported: {
            capabilities: 10,
            components: 5,
            parentChildRelationships: 8,
            realizations: 3,
          },
          unsupported: {
            elements: {},
            relationships: {},
          },
        },
        createdAt: '2025-01-15T10:00:00Z',
        _links: {
          self: '/api/v1/imports/import-123',
          confirm: '/api/v1/imports/import-123/confirm',
          delete: '/api/v1/imports/import-123',
        },
      };

      mockPost.mockResolvedValue({ data: mockSession });

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      const request: CreateImportSessionRequest = {
        file,
        sourceFormat: 'archimate-openexchange',
      };

      await act(async () => {
        await result.current.createSession(request);
      });

      await waitFor(() => {
        expect(result.current.session).toEqual(mockSession);
        expect(result.current.isLoading).toBe(false);
        expect(result.current.error).toBeNull();
      });
    });

    it('should handle API errors during session creation', async () => {
      const errorMessage = 'Invalid file format';
      mockPost.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      const request: CreateImportSessionRequest = {
        file,
        sourceFormat: 'archimate-openexchange',
      };

      await act(async () => {
        await result.current.createSession(request);
      });

      await waitFor(() => {
        expect(result.current.error).toBe(errorMessage);
        expect(result.current.isLoading).toBe(false);
        expect(result.current.session).toBeNull();
      });
    });
  });

  describe('confirmSession', () => {
    it('should confirm import session and start importing', async () => {
      const pendingSession: ImportSession = {
        id: 'import-123' as any,
        status: 'pending',
        sourceFormat: 'archimate-openexchange',
        createdAt: '2025-01-15T10:00:00Z',
        _links: { self: '/api/v1/imports/import-123', confirm: '/api/v1/imports/import-123/confirm' },
      };

      const importingSession: ImportSession = {
        ...pendingSession,
        status: 'importing',
        progress: {
          phase: 'creating_components',
          totalItems: 15,
          completedItems: 0,
        },
      };

      mockPost
        .mockResolvedValueOnce({ data: pendingSession })
        .mockResolvedValueOnce({ data: importingSession });

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await waitFor(() => {
        expect(result.current.session?.status).toBe('pending');
      });

      await act(async () => {
        await result.current.confirmSession();
      });

      await waitFor(() => {
        expect(result.current.session?.status).toBe('importing');
        expect(result.current.isLoading).toBe(false);
      });
    });

    it('should handle errors during confirmation', async () => {
      const mockSession: ImportSession = {
        id: 'import-123' as any,
        status: 'pending',
        sourceFormat: 'archimate-openexchange',
        createdAt: '2025-01-15T10:00:00Z',
        _links: { self: '/api/v1/imports/import-123', confirm: '/api/v1/imports/import-123/confirm' },
      };

      const errorMessage = 'Import already started';
      mockPost
        .mockResolvedValueOnce({ data: mockSession })
        .mockRejectedValueOnce(new Error(errorMessage));

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await waitFor(() => {
        expect(result.current.session?.status).toBe('pending');
      });

      await act(async () => {
        await result.current.confirmSession();
      });

      await waitFor(() => {
        expect(result.current.error).toBe(errorMessage);
      });
    });
  });

  describe('cancelSession', () => {
    it('should cancel pending import session', async () => {
      const mockSession: ImportSession = {
        id: 'import-123' as any,
        status: 'pending',
        sourceFormat: 'archimate-openexchange',
        createdAt: '2025-01-15T10:00:00Z',
        _links: {
          self: '/api/v1/imports/import-123',
          delete: '/api/v1/imports/import-123',
        },
      };

      mockPost.mockResolvedValue({ data: mockSession });
      mockDelete.mockResolvedValue({});

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await act(async () => {
        await result.current.cancelSession();
      });

      await waitFor(() => {
        expect(result.current.session).toBeNull();
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe('polling', () => {
    it('should poll for progress when status is importing', async () => {
      const importingSession: ImportSession = {
        id: 'import-123' as any,
        status: 'importing',
        sourceFormat: 'archimate-openexchange',
        progress: {
          phase: 'creating_components',
          totalItems: 15,
          completedItems: 5,
        },
        createdAt: '2025-01-15T10:00:00Z',
        _links: { self: '/api/v1/imports/import-123' },
      };

      const completedSession: ImportSession = {
        ...importingSession,
        status: 'completed',
        progress: undefined,
        result: {
          capabilitiesCreated: 10,
          componentsCreated: 5,
          realizationsCreated: 3,
          domainAssignments: 0,
          errors: [],
        },
        completedAt: '2025-01-15T10:05:00Z',
      };

      mockPost.mockResolvedValue({ data: importingSession });
      mockGet
        .mockResolvedValueOnce({ data: importingSession })
        .mockResolvedValueOnce({ data: completedSession });

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await waitFor(() => {
        expect(result.current.session?.status).toBe('importing');
      });

      await waitFor(
        () => {
          expect(result.current.session?.status).toBe('completed');
        },
        { timeout: 8000 }
      );

      expect(mockGet).toHaveBeenCalled();
    }, 15000);

    it('should stop polling when status is completed', async () => {
      const completedSession: ImportSession = {
        id: 'import-123' as any,
        status: 'completed',
        sourceFormat: 'archimate-openexchange',
        result: {
          capabilitiesCreated: 10,
          componentsCreated: 5,
          realizationsCreated: 3,
          domainAssignments: 0,
          errors: [],
        },
        createdAt: '2025-01-15T10:00:00Z',
        completedAt: '2025-01-15T10:05:00Z',
        _links: { self: '/api/v1/imports/import-123' },
      };

      mockPost.mockResolvedValue({ data: completedSession });

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await waitFor(() => {
        expect(result.current.session?.status).toBe('completed');
      });

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 3000));
      });

      expect(mockGet).not.toHaveBeenCalled();
    }, 10000);
  });

  describe('reset', () => {
    it('should reset session state', async () => {
      const mockSession: ImportSession = {
        id: 'import-123' as any,
        status: 'pending',
        sourceFormat: 'archimate-openexchange',
        createdAt: '2025-01-15T10:00:00Z',
        _links: { self: '/api/v1/imports/import-123' },
      };

      mockPost.mockResolvedValue({ data: mockSession });

      const { result } = renderHook(() => useImportSession());

      const file = new File(['test'], 'test.xml', { type: 'application/xml' });
      await act(async () => {
        await result.current.createSession({ file, sourceFormat: 'archimate-openexchange' });
      });

      await waitFor(() => {
        expect(result.current.session).toEqual(mockSession);
      });

      act(() => {
        result.current.reset();
      });

      await waitFor(() => {
        expect(result.current.session).toBeNull();
        expect(result.current.error).toBeNull();
        expect(result.current.isLoading).toBe(false);
      });
    });
  });
});
