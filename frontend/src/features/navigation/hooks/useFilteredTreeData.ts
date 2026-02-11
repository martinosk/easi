import { useState, useMemo } from 'react';
import { useComponents } from '../../components/hooks/useComponents';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useViews } from '../../views/hooks/useViews';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { useArtifactCreators } from './useArtifactCreators';
import { filterByCreator } from '../utils/filterByCreator';
import { preserveCapabilityHierarchy } from '../utils/preserveCapabilityHierarchy';

export function useFilteredTreeData() {
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: views = [] } = useViews();
  const { data: acquiredEntities = [] } = useAcquiredEntitiesQuery();
  const { data: vendors = [] } = useVendorsQuery();
  const { data: internalTeams = [] } = useInternalTeamsQuery();
  const { data: activeUsers = [] } = useActiveUsers();
  const { data: artifactCreators = [] } = useArtifactCreators();

  const [selectedCreatorIds, setSelectedCreatorIds] = useState<string[]>([]);

  const creatorMap = useMemo(
    () => new Map(artifactCreators.map((ac) => [ac.aggregateId, ac.creatorId])),
    [artifactCreators]
  );

  const filtered = useMemo(() => {
    const result = filterByCreator(
      { components, capabilities, acquiredEntities, vendors, internalTeams },
      selectedCreatorIds,
      creatorMap
    );
    return {
      ...result,
      capabilities: preserveCapabilityHierarchy(result.capabilities, capabilities),
    };
  }, [components, capabilities, acquiredEntities, vendors, internalTeams, selectedCreatorIds, creatorMap]);

  return {
    components,
    views,
    filtered,
    artifactCreators,
    activeUsers,
    selectedCreatorIds,
    setSelectedCreatorIds,
  };
}
