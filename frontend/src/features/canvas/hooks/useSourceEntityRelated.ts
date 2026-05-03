import { useMemo } from 'react';
import { getEntityId, getEntityType, toNodeId } from '../../../constants/entityIdentifiers';
import { getPostableRelated, type RelatedLink, type ResourceWithRelated } from '../../../utils/xRelated';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';

type Identifiable = ResourceWithRelated & { id: string };

const findById = (list: ReadonlyArray<Identifiable>, id: string): ResourceWithRelated | undefined =>
  list.find((e) => e.id === id);

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

    const sources: Record<typeof type, ReadonlyArray<Identifiable>> = {
      component: components,
      capability: capabilities,
      acquired,
      vendor: vendors,
      team: teams,
    };
    const entity = findById(sources[type], id);

    return entity ? getPostableRelated(entity) : [];
  }, [nodeId, components, capabilities, acquired, vendors, teams]);
}
