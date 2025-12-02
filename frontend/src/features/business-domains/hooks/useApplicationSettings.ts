import { useState, useCallback } from 'react';

const SHOW_APPLICATIONS_KEY = 'business-domains-show-applications';

export interface UseApplicationSettingsResult {
  showApplications: boolean;
  setShowApplications: (value: boolean) => void;
}

export function useApplicationSettings(): UseApplicationSettingsResult {
  const [showApplications, setShowApplicationsState] = useState<boolean>(() => {
    const stored = localStorage.getItem(SHOW_APPLICATIONS_KEY);
    return stored ? stored === 'true' : false;
  });

  const setShowApplications = useCallback((value: boolean) => {
    setShowApplicationsState(value);
    localStorage.setItem(SHOW_APPLICATIONS_KEY, String(value));
  }, []);

  return {
    showApplications,
    setShowApplications,
  };
}
