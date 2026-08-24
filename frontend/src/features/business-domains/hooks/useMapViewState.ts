import { useCallback, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';

export type BoardViewMode = 'board' | 'map' | 'timeline';
type StoredViewMode = Exclude<BoardViewMode, 'timeline'>;
export type MapDepth = 1 | 2 | 3 | 4;

export const PRESENTATION_PARAM = 'presentation';
const TIMELINE = 'timeline';

const VIEW_MODE_KEY = 'business-domains-view';
const DEPTH_KEY = 'business-domains-map-depth';
const DOMAIN_KEY = 'business-domains-map-domain';
const SHOW_APPS_KEY = 'business-domains-map-apps';

const DEFAULT_DEPTH: MapDepth = 2;

function isStoredViewMode(value: string | null): value is StoredViewMode {
  return value === 'board' || value === 'map';
}

function isMapDepth(value: number): value is MapDepth {
  return Number.isInteger(value) && value >= 1 && value <= 4;
}

function usePersistedValue<T>(key: string, parse: (stored: string | null) => T): [T, (next: T) => void] {
  const [value, setValue] = useState<T>(() => parse(localStorage.getItem(key)));

  const set = useCallback(
    (next: T) => {
      setValue(next);
      localStorage.setItem(key, String(next));
    },
    [key],
  );

  return [value, set];
}

export function useViewMode(): [BoardViewMode, (mode: BoardViewMode) => void] {
  const [searchParams, setSearchParams] = useSearchParams();
  const [stored, setStored] = usePersistedValue<StoredViewMode>(VIEW_MODE_KEY, (value) =>
    isStoredViewMode(value) ? value : 'board',
  );

  const mode: BoardViewMode = searchParams.get(PRESENTATION_PARAM) === TIMELINE ? TIMELINE : stored;

  const setMode = useCallback(
    (next: BoardViewMode) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          if (next === TIMELINE) params.set(PRESENTATION_PARAM, TIMELINE);
          else params.delete(PRESENTATION_PARAM);
          return params;
        },
        { replace: true },
      );
      if (next !== TIMELINE) setStored(next);
    },
    [setSearchParams, setStored],
  );

  return [mode, setMode];
}

export function useMapDepth(): [MapDepth, (depth: MapDepth) => void] {
  return usePersistedValue(DEPTH_KEY, (stored) => {
    const parsed = Number(stored);
    return isMapDepth(parsed) ? parsed : DEFAULT_DEPTH;
  });
}

export function useMapShowApps(): [boolean, (show: boolean) => void] {
  return usePersistedValue(SHOW_APPS_KEY, (stored) => stored === 'true');
}

export function resolveMapDomainId(storedId: string | null, domains: BusinessDomain[]): BusinessDomainId | null {
  if (domains.length === 0) return null;
  const match = domains.find((domain) => String(domain.id) === storedId);
  return match ? match.id : domains[0].id;
}

export function useMapDomain(domains: BusinessDomain[]): [BusinessDomainId | null, (id: BusinessDomainId) => void] {
  const [storedId, setStoredId] = usePersistedValue(DOMAIN_KEY, (stored) => stored);
  return [resolveMapDomainId(storedId, domains), setStoredId];
}
