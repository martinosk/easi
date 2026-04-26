import { act, renderHook, waitFor } from '@testing-library/react';
import type { RenderHookResult } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';
import { toReleaseVersion } from '../../../api/types';
import { useReleaseNotes } from './useReleaseNotes';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getVersion: vi.fn(),
    getLatestRelease: vi.fn(),
  },
}));

const STORAGE_KEY = 'releaseNotesPreferences';

const mockRelease: Release = {
  version: toReleaseVersion('1.2.0'),
  releaseDate: '2024-01-15T00:00:00Z',
  notes: '## Features\n- New feature',
  _links: { self: { href: '/api/v1/releases/1.2.0', method: 'GET' } },
};

type DismissMode = 'forever' | 'untilNext';

interface StoredPrefs {
  dismissedVersion: string;
  dismissMode: DismissMode;
}

function mockApi(version: string | Error, release: Release | Error | null = mockRelease) {
  const versionMock = vi.mocked(apiClient.getVersion);
  if (version instanceof Error) versionMock.mockRejectedValue(version);
  else versionMock.mockResolvedValue(version);

  const releaseMock = vi.mocked(apiClient.getLatestRelease);
  if (release instanceof Error) releaseMock.mockRejectedValue(release);
  else releaseMock.mockResolvedValue(release);
}

function setStoredPrefs(prefs: StoredPrefs) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
}

function getStoredPrefs(): Partial<StoredPrefs> {
  return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
}

type HookResult = RenderHookResult<ReturnType<typeof useReleaseNotes>, unknown>;

async function renderAndLoad(): Promise<HookResult> {
  const result = renderHook(() => useReleaseNotes());
  await waitFor(() => {
    expect(result.result.current.isLoading).toBe(false);
  });
  return result;
}

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
      mockApi('1.2.0');

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
      mockApi('1.2.0');
      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(true);
    });

    it('should not show overlay when version has been dismissed', async () => {
      setStoredPrefs({ dismissedVersion: '1.2.0', dismissMode: 'untilNext' });
      mockApi('1.2.0');

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(false);
    });

    it('should show overlay when version is newer than dismissed version', async () => {
      setStoredPrefs({ dismissedVersion: '1.1.0', dismissMode: 'untilNext' });
      mockApi('1.2.0');

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(true);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors gracefully without showing overlay', async () => {
      mockApi(new Error('Network error'), new Error('Network error'));

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(false);
      expect(result.current.release).toBeNull();
    });

    it('should handle null release response', async () => {
      mockApi('1.2.0', null);

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(false);
      expect(result.current.release).toBeNull();
    });
  });

  describe('Dismiss Functionality', () => {
    it.each<DismissMode>(['untilNext', 'forever'])(
      'should dismiss overlay and store preference when using %s mode',
      async (mode) => {
        mockApi('1.2.0');

        const { result } = await renderAndLoad();
        act(() => {
          result.current.dismiss(mode);
        });

        expect(result.current.showOverlay).toBe(false);
        const stored = getStoredPrefs();
        expect(stored.dismissedVersion).toBe('1.2.0');
        expect(stored.dismissMode).toBe(mode);
      },
    );

    it('should not store preferences if version is not loaded', async () => {
      mockApi('1.2.0');

      const { result } = renderHook(() => useReleaseNotes());
      act(() => {
        result.current.dismiss('untilNext');
      });

      expect(localStorage.getItem(STORAGE_KEY)).toBeNull();

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe('LocalStorage Handling', () => {
    it('should handle corrupted localStorage data gracefully', async () => {
      localStorage.setItem(STORAGE_KEY, 'not valid json');
      mockApi('1.2.0');

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(true);
    });

    it('should handle localStorage not being available', async () => {
      const getItemSpy = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
        throw new Error('localStorage not available');
      });
      const setItemSpy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('localStorage not available');
      });

      mockApi('1.2.0');

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(true);

      act(() => {
        result.current.dismiss('untilNext');
      });
      expect(result.current.showOverlay).toBe(false);

      getItemSpy.mockRestore();
      setItemSpy.mockRestore();
    });
  });

  describe('Version Comparison for Overlay Display', () => {
    it('should show overlay when forever dismissed but version changed', async () => {
      setStoredPrefs({ dismissedVersion: '1.0.0', dismissMode: 'forever' });
      mockApi('2.0.0', { ...mockRelease, version: toReleaseVersion('2.0.0') });

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(true);
    });

    it('should not show overlay when same version is dismissed regardless of mode', async () => {
      setStoredPrefs({ dismissedVersion: '1.2.0', dismissMode: 'forever' });
      mockApi('1.2.0');

      const { result } = await renderAndLoad();
      expect(result.current.showOverlay).toBe(false);
    });
  });
});
