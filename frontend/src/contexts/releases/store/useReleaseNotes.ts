import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';

const STORAGE_KEY = 'releaseNotesPreferences';

interface ReleaseNotesPreferences {
  dismissedVersion: string;
  dismissMode: 'forever' | 'untilNext';
}

interface UseReleaseNotesResult {
  showOverlay: boolean;
  release: Release | null;
  isLoading: boolean;
  dismiss: (mode: 'forever' | 'untilNext') => void;
}

function getStoredPreferences(): ReleaseNotesPreferences | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return null;
    return JSON.parse(stored) as ReleaseNotesPreferences;
  } catch {
    return null;
  }
}

function setStoredPreferences(prefs: ReleaseNotesPreferences): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
  } catch {
    // Ignore storage errors
  }
}

function shouldShowOverlay(currentVersion: string, preferences: ReleaseNotesPreferences | null): boolean {
  if (!preferences) {
    return true;
  }

  if (preferences.dismissedVersion !== currentVersion) {
    return true;
  }

  return false;
}

export function useReleaseNotes(): UseReleaseNotesResult {
  const [showOverlay, setShowOverlay] = useState(false);
  const [release, setRelease] = useState<Release | null>(null);
  const [currentVersion, setCurrentVersion] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchReleaseInfo() {
      try {
        const [version, latestRelease] = await Promise.all([
          apiClient.getVersion(),
          apiClient.getLatestRelease(),
        ]);

        setCurrentVersion(version);
        setRelease(latestRelease);

        if (latestRelease && version) {
          const preferences = getStoredPreferences();
          const shouldShow = shouldShowOverlay(version, preferences);
          setShowOverlay(shouldShow);
        }
      } catch {
        // Silently fail - don't block app startup for release notes
      } finally {
        setIsLoading(false);
      }
    }

    fetchReleaseInfo();
  }, []);

  const dismiss = useCallback((mode: 'forever' | 'untilNext') => {
    if (!currentVersion) return;

    setStoredPreferences({
      dismissedVersion: currentVersion,
      dismissMode: mode,
    });

    setShowOverlay(false);
  }, [currentVersion]);

  return {
    showOverlay,
    release,
    isLoading,
    dismiss,
  };
}
