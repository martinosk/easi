import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useReleaseNotes } from './useReleaseNotes';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getVersion: vi.fn(),
    getLatestRelease: vi.fn(),
  },
}));

const STORAGE_KEY = 'releaseNotesPreferences';

const mockRelease: Release = {
  version: '1.2.0',
  releaseDate: '2024-01-15T00:00:00Z',
  notes: '## Features\n- New feature',
  _links: { self: { href: '/api/v1/releases/1.2.0' } },
};

describe('useReleaseNotes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('Initial Loading', () => {
    it('should fetch version and latest release on mount', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(apiClient.getVersion).toHaveBeenCalledTimes(1);
      expect(apiClient.getLatestRelease).toHaveBeenCalledTimes(1);
      expect(result.current.release).toEqual(mockRelease);
    });

    it('should show overlay when no preferences are stored', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);
    });

    it('should not show overlay when version has been dismissed', async () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        dismissedVersion: '1.2.0',
        dismissMode: 'untilNext',
      }));

      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(false);
    });

    it('should show overlay when version is newer than dismissed version', async () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        dismissedVersion: '1.1.0',
        dismissMode: 'untilNext',
      }));

      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors gracefully without showing overlay', async () => {
      vi.mocked(apiClient.getVersion).mockRejectedValue(new Error('Network error'));
      vi.mocked(apiClient.getLatestRelease).mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(false);
      expect(result.current.release).toBeNull();
    });

    it('should handle null release response', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(null);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(false);
      expect(result.current.release).toBeNull();
    });
  });

  describe('Dismiss Functionality', () => {
    it('should dismiss overlay and store preference when using untilNext mode', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);

      act(() => {
        result.current.dismiss('untilNext');
      });

      expect(result.current.showOverlay).toBe(false);

      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
      expect(stored.dismissedVersion).toBe('1.2.0');
      expect(stored.dismissMode).toBe('untilNext');
    });

    it('should dismiss overlay and store preference when using forever mode', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      act(() => {
        result.current.dismiss('forever');
      });

      expect(result.current.showOverlay).toBe(false);

      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
      expect(stored.dismissedVersion).toBe('1.2.0');
      expect(stored.dismissMode).toBe('forever');
    });

    it('should not store preferences if version is not loaded', async () => {
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      act(() => {
        result.current.dismiss('untilNext');
      });

      const stored = localStorage.getItem(STORAGE_KEY);
      expect(stored).toBeNull();

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe('LocalStorage Handling', () => {
    it('should handle corrupted localStorage data gracefully', async () => {
      localStorage.setItem(STORAGE_KEY, 'not valid json');

      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);
    });

    it('should handle localStorage not being available', async () => {
      const originalGetItem = localStorage.getItem;
      const originalSetItem = localStorage.setItem;

      vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
        throw new Error('localStorage not available');
      });
      vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('localStorage not available');
      });

      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);

      act(() => {
        result.current.dismiss('untilNext');
      });

      expect(result.current.showOverlay).toBe(false);

      vi.mocked(localStorage.getItem).mockRestore();
      vi.mocked(localStorage.setItem).mockRestore();
    });
  });

  describe('Version Comparison for Overlay Display', () => {
    it('should show overlay when forever dismissed but version changed', async () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        dismissedVersion: '1.0.0',
        dismissMode: 'forever',
      }));

      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue({
        ...mockRelease,
        version: '2.0.0',
      });

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(true);
    });

    it('should not show overlay when same version is dismissed regardless of mode', async () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        dismissedVersion: '1.2.0',
        dismissMode: 'forever',
      }));

      vi.mocked(apiClient.getVersion).mockResolvedValue('1.2.0');
      vi.mocked(apiClient.getLatestRelease).mockResolvedValue(mockRelease);

      const { result } = renderHook(() => useReleaseNotes());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.showOverlay).toBe(false);
    });
  });
});
