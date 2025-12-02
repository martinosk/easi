import { useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, CapabilityId, CapabilityLevel, CapabilityRealization } from '../../../api/types';

export interface UseCapabilityRealizationsResult {
  realizations: CapabilityRealization[];
  isLoading: boolean;
  error: Error | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
}

function getLevelNumber(level: CapabilityLevel): number {
  return parseInt(level.substring(1), 10);
}

function isDirectRealization(r: CapabilityRealization): boolean {
  return r.origin === 'Direct';
}

function isInheritedWithHiddenSource(
  r: CapabilityRealization,
  visibleCapabilityIds: Set<CapabilityId>
): boolean {
  return r.origin === 'Inherited' &&
         !!r.sourceCapabilityId &&
         !visibleCapabilityIds.has(r.sourceCapabilityId);
}

function selectDeepestInherited(
  existing: CapabilityRealization,
  candidate: CapabilityRealization,
  levelMap: Map<CapabilityId, number>
): CapabilityRealization {
  const existingLevel = levelMap.get(existing.capabilityId) ?? 0;
  const candidateLevel = levelMap.get(candidate.capabilityId) ?? 0;
  return candidateLevel > existingLevel ? candidate : existing;
}

export function filterVisibleRealizations(
  realizations: CapabilityRealization[],
  capabilityLevels: Map<CapabilityId, number>
): CapabilityRealization[] {
  const visibleCapabilityIds = new Set(capabilityLevels.keys());
  const directRealizations: CapabilityRealization[] = [];
  const inheritedByKey = new Map<string, CapabilityRealization>();

  for (const r of realizations) {
    if (!visibleCapabilityIds.has(r.capabilityId)) {
      continue;
    }

    if (isDirectRealization(r)) {
      directRealizations.push(r);
      continue;
    }

    if (isInheritedWithHiddenSource(r, visibleCapabilityIds)) {
      const key = `${r.componentId}:${r.sourceCapabilityId}`;
      const existing = inheritedByKey.get(key);
      const selected = existing ? selectDeepestInherited(existing, r, capabilityLevels) : r;
      inheritedByKey.set(key, selected);
    }
  }

  return [...directRealizations, ...inheritedByKey.values()];
}

export function useCapabilityRealizations(
  enabled: boolean,
  domainId: BusinessDomainId | null,
  depth: number
): UseCapabilityRealizationsResult {
  const [allRealizations, setAllRealizations] = useState<CapabilityRealization[]>([]);
  const [capabilityLevels, setCapabilityLevels] = useState<Map<CapabilityId, number>>(new Map());
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!enabled || !domainId) {
      setAllRealizations([]);
      setCapabilityLevels(new Map());
      setError(null);
      return;
    }

    const fetchRealizations = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const groups = await apiClient.getCapabilityRealizationsByDomain(domainId, depth);
        const levelMap = new Map<CapabilityId, number>();
        const realizations: CapabilityRealization[] = [];

        for (const group of groups) {
          levelMap.set(group.capabilityId as CapabilityId, getLevelNumber(group.level));
          realizations.push(...group.realizations);
        }

        setCapabilityLevels(levelMap);
        setAllRealizations(realizations);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch capability realizations'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchRealizations();
  }, [enabled, domainId, depth]);

  const filteredRealizations = useMemo(
    () => filterVisibleRealizations(allRealizations, capabilityLevels),
    [allRealizations, capabilityLevels]
  );

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
