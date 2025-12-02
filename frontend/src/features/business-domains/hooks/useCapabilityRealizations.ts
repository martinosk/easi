import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { apiClient } from '../../../api/client';
import type { CapabilityId, CapabilityLevel, CapabilityRealization } from '../../../api/types';

export interface VisibleCapability {
  id: CapabilityId;
  level: CapabilityLevel;
}

export interface UseCapabilityRealizationsResult {
  realizations: CapabilityRealization[];
  isLoading: boolean;
  error: Error | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
}

function getLevelNumber(level: CapabilityLevel): number {
  return parseInt(level.substring(1), 10);
}

export function useCapabilityRealizations(
  enabled: boolean,
  visibleCapabilities: VisibleCapability[]
): UseCapabilityRealizationsResult {
  const [allRealizations, setAllRealizations] = useState<CapabilityRealization[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const fetchedCapabilityIds = useRef<Set<CapabilityId>>(new Set());

  const visibleCapabilityIds = useMemo(
    () => visibleCapabilities.map((c) => c.id),
    [visibleCapabilities]
  );

  const capabilityLevelMap = useMemo(() => {
    const map = new Map<CapabilityId, number>();
    visibleCapabilities.forEach((c) => map.set(c.id, getLevelNumber(c.level)));
    return map;
  }, [visibleCapabilities]);

  useEffect(() => {
    if (!enabled || visibleCapabilityIds.length === 0) {
      return;
    }

    const newCapabilityIds = visibleCapabilityIds.filter(
      (id) => !fetchedCapabilityIds.current.has(id)
    );

    if (newCapabilityIds.length === 0) {
      return;
    }

    const fetchRealizations = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const results = await Promise.all(
          newCapabilityIds.map((capabilityId) => apiClient.getSystemsByCapability(capabilityId))
        );

        newCapabilityIds.forEach((id) => fetchedCapabilityIds.current.add(id));

        const newRealizations = results.flat();
        setAllRealizations((prev) => {
          const existingIds = new Set(prev.map((r) => r.id));
          const uniqueNew = newRealizations.filter((r) => !existingIds.has(r.id));
          return [...prev, ...uniqueNew];
        });
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch capability realizations'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchRealizations();
  }, [enabled, visibleCapabilityIds]);

  useEffect(() => {
    if (!enabled) {
      setAllRealizations([]);
      fetchedCapabilityIds.current.clear();
      setError(null);
    }
  }, [enabled]);

  const filteredRealizations = useMemo(() => {
    const visibleSet = new Set(visibleCapabilityIds);

    const directRealizations: CapabilityRealization[] = [];
    const inheritedByKey = new Map<string, CapabilityRealization>();

    for (const r of allRealizations) {
      if (!visibleSet.has(r.capabilityId)) {
        continue;
      }

      if (r.origin === 'Direct') {
        directRealizations.push(r);
        continue;
      }

      if (r.origin === 'Inherited' && r.sourceCapabilityId) {
        if (visibleSet.has(r.sourceCapabilityId)) {
          continue;
        }

        const key = `${r.componentId}:${r.sourceCapabilityId}`;
        const existing = inheritedByKey.get(key);

        if (!existing) {
          inheritedByKey.set(key, r);
        } else {
          const existingLevel = capabilityLevelMap.get(existing.capabilityId) ?? 0;
          const currentLevel = capabilityLevelMap.get(r.capabilityId) ?? 0;
          if (currentLevel > existingLevel) {
            inheritedByKey.set(key, r);
          }
        }
      }
    }

    return [...directRealizations, ...inheritedByKey.values()];
  }, [allRealizations, visibleCapabilityIds, capabilityLevelMap]);

  const getRealizationsForCapability = useCallback(
    (capabilityId: CapabilityId): CapabilityRealization[] => {
      return filteredRealizations.filter((r) => r.capabilityId === capabilityId);
    },
    [filteredRealizations]
  );

  return {
    realizations: filteredRealizations,
    isLoading,
    error,
    getRealizationsForCapability,
  };
}
