import { useState, useCallback } from 'react';

const SHOW_APPLICATIONS_KEY = 'business-domains-show-applications';
const SHOW_INHERITED_KEY = 'business-domains-show-inherited';

export interface UseApplicationSettingsResult {
  showApplications: boolean;
  showInherited: boolean;
  setShowApplications: (value: boolean) => void;
  setShowInherited: (value: boolean) => void;
}

export function useApplicationSettings(): UseApplicationSettingsResult {
  const [showApplications, setShowApplicationsState] = useState<boolean>(() => {
    const stored = localStorage.getItem(SHOW_APPLICATIONS_KEY);
    return stored ? stored === 'true' : false;
  });

  const [showInherited, setShowInheritedState] = useState<boolean>(() => {
    const stored = localStorage.getItem(SHOW_INHERITED_KEY);
    return stored ? stored === 'true' : false;
  });

  const setShowApplications = useCallback((value: boolean) => {
    setShowApplicationsState(value);
    localStorage.setItem(SHOW_APPLICATIONS_KEY, String(value));
  }, []);

  const setShowInherited = useCallback((value: boolean) => {
    setShowInheritedState(value);
    localStorage.setItem(SHOW_INHERITED_KEY, String(value));
  }, []);

  return {
    showApplications,
    showInherited,
    setShowApplications,
    setShowInherited,
  };
}
