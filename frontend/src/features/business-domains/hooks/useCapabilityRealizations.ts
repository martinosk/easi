import { useCallback, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../api/client';
import { businessDomainsQueryKeys } from '../queryKeys';
import type { BusinessDomainId, CapabilityId, CapabilityLevel, CapabilityRealization, CapabilityRealizationsGroup } from '../../../api/types';

export interface UseCapabilityRealizationsResult {
  realizations: CapabilityRealization[];
  isLoading: boolean;
  error: Error | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  refetch: () => Promise<void>;
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
  capabilityLevels: Map<CapabilityId, number>,
  visibleCapabilityIds?: Set<CapabilityId>
): CapabilityRealization[] {
  const visibleIds = visibleCapabilityIds ?? new Set(capabilityLevels.keys());
  const directRealizations: CapabilityRealization[] = [];
  const inheritedByKey = new Map<string, CapabilityRealization>();

  for (const r of realizations) {
    if (!visibleIds.has(r.capabilityId)) {
      continue;
    }

    if (isDirectRealization(r)) {
      directRealizations.push(r);
      continue;
    }

    if (isInheritedWithHiddenSource(r, visibleIds)) {
      const key = `${r.componentId}:${r.sourceCapabilityId}`;
      const existing = inheritedByKey.get(key);
      const selected = existing ? selectDeepestInherited(existing, r, capabilityLevels) : r;
      inheritedByKey.set(key, selected);
    }
  }

  return [...directRealizations, ...inheritedByKey.values()];
}

function flattenGroups(groups: CapabilityRealizationsGroup[]): {
  realizations: CapabilityRealization[];
  capabilityLevels: Map<CapabilityId, number>;
} {
  const capabilityLevels = new Map<CapabilityId, number>();
  const realizations: CapabilityRealization[] = [];

  for (const group of groups) {
    capabilityLevels.set(group.capabilityId, getLevelNumber(group.level));
    realizations.push(...group.realizations);
  }

  return { realizations, capabilityLevels };
}

export function useCapabilityRealizations(
  enabled: boolean,
  domainId: BusinessDomainId | null,
  depth: number,
  visibleCapabilityIds?: Set<CapabilityId> | CapabilityId[]
): UseCapabilityRealizationsResult {
  const { data: groups, isLoading, error, refetch } = useQuery({
    queryKey: businessDomainsQueryKeys.realizations(domainId ?? '', depth),
    queryFn: () => apiClient.getCapabilityRealizationsByDomain(domainId!, depth),
    enabled: enabled && !!domainId,
  });

  const { realizations, capabilityLevels } = useMemo(
    () => flattenGroups(groups ?? []),
    [groups]
  );

  const visibleIds = useMemo(() => {
    if (!visibleCapabilityIds) return undefined;
    return visibleCapabilityIds instanceof Set
      ? visibleCapabilityIds
      : new Set(visibleCapabilityIds);
  }, [visibleCapabilityIds]);

  const filteredRealizations = useMemo(
    () => filterVisibleRealizations(realizations, capabilityLevels, visibleIds),
    [realizations, capabilityLevels, visibleIds]
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
    error: error ?? null,
    getRealizationsForCapability,
    refetch: async () => { await refetch(); },
  };
}
