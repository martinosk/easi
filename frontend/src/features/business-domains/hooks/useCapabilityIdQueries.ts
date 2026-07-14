import { type UseQueryResult, useQueries } from '@tanstack/react-query';
import type { CapabilityRealizationsGroup } from '../../../api/types';

function serializeCapabilityIds(capabilityIds: string[]): string {
  return JSON.stringify([...capabilityIds].sort());
}

export function useCapabilityIdQueries<T>(
  realizationQueries: UseQueryResult<CapabilityRealizationsGroup[]>[],
  queryKey: (capabilityIds: string[]) => readonly unknown[],
  queryFn: (capabilityIds: string[]) => Promise<T>,
): UseQueryResult<T>[] {
  const idsPerDomain = realizationQueries.map((query) => (query.data ?? []).map((group) => group.capabilityId));

  const uniqueByKey = new Map<string, string[]>();
  for (const capabilityIds of idsPerDomain) {
    const key = serializeCapabilityIds(capabilityIds);
    if (!uniqueByKey.has(key)) uniqueByKey.set(key, capabilityIds);
  }
  const uniqueEntries = [...uniqueByKey.entries()];

  const results = useQueries({
    queries: uniqueEntries.map(([, capabilityIds]) => ({
      queryKey: queryKey(capabilityIds),
      queryFn: () => queryFn(capabilityIds),
      enabled: capabilityIds.length > 0,
    })),
  });

  const resultByKey = new Map(uniqueEntries.map(([key], index) => [key, results[index]]));
  return idsPerDomain.map(
    (capabilityIds) => resultByKey.get(serializeCapabilityIds(capabilityIds)) as UseQueryResult<T>,
  );
}
