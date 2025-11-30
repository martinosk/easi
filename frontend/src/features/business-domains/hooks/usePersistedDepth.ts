import { useState, useCallback } from 'react';
import type { DepthLevel } from '../components/DepthSelector';

const STORAGE_KEY = 'business-domains-depth';

export function usePersistedDepth(): [DepthLevel, (depth: DepthLevel) => void] {
  const [depth, setDepthState] = useState<DepthLevel>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? (Number(stored) as DepthLevel) : 1;
  });

  const setDepth = useCallback((newDepth: DepthLevel) => {
    setDepthState(newDepth);
    localStorage.setItem(STORAGE_KEY, String(newDepth));
  }, []);

  return [depth, setDepth];
}
