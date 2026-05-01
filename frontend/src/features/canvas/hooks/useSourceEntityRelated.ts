import { useMemo } from 'react';
import { getEntityId, getEntityType, toNodeId } from '../../../constants/entityIdentifiers';
import { getPostableRelated, type RelatedLink } from '../../../utils/xRelated';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';

export function useSourceEntityRelated(nodeId: string | null): RelatedLink[] {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: acquired = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: teams = [] } = useInternalTeamsQuery();

  return useMemo(() => {
    if (!nodeId) return [];
    const node = toNodeId(nodeId);
    const type = getEntityType(node);
    const id = getEntityId(node);

    const finder = (list: ReadonlyArray<{ id: string }>) => list.find((e) => e.id === id);

    const entity =
      type === 'component'
        ? finder(components)
        : type === 'capability'
          ? finder(capabilities)
          : type === 'acquired'
            ? finder(acquired)
            : type === 'vendor'
              ? finder(vendors)
              : type === 'team'
                ? finder(teams)
                : undefined;

    return entity ? getPostableRelated(entity) : [];
  }, [nodeId, components, capabilities, acquired, vendors, teams]);
}
