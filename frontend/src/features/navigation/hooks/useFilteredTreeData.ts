import { useState, useMemo } from 'react';
import { useComponents } from '../../components/hooks/useComponents';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useViews } from '../../views/hooks/useViews';
import { useAcquiredEntitiesQuery } from '../../origin-entities/hooks/useAcquiredEntities';
import { useVendorsQuery } from '../../origin-entities/hooks/useVendors';
import { useInternalTeamsQuery } from '../../origin-entities/hooks/useInternalTeams';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { useArtifactCreators } from './useArtifactCreators';
import { useDomainFilterData } from './useDomainFilterData';
import { filterByCreator } from '../utils/filterByCreator';
import { filterByDomain } from '../utils/filterByDomain';
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
  const [selectedDomainIds, setSelectedDomainIds] = useState<string[]>([]);

  const { domains, domainFilterData } = useDomainFilterData(capabilities);

  const creatorMap = useMemo(
    () => new Map(artifactCreators.map((ac) => [ac.aggregateId, ac.creatorId])),
    [artifactCreators]
  );

  const filtered = useMemo(() => {
    const artifacts = { components, capabilities, acquiredEntities, vendors, internalTeams };

    const afterCreator = filterByCreator(artifacts, selectedCreatorIds, creatorMap);
    const afterDomain = filterByDomain(afterCreator, selectedDomainIds, domainFilterData);

    return {
      ...afterDomain,
      capabilities: preserveCapabilityHierarchy(afterDomain.capabilities, capabilities),
    };
  }, [components, capabilities, acquiredEntities, vendors, internalTeams, selectedCreatorIds, creatorMap, selectedDomainIds, domainFilterData]);

  const hasActiveFilters = selectedCreatorIds.length > 0 || selectedDomainIds.length > 0;

  const clearAllFilters = () => {
    setSelectedCreatorIds([]);
    setSelectedDomainIds([]);
  };

  return {
    components,
    views,
    filtered,
    artifactCreators,
    activeUsers,
    selectedCreatorIds,
    setSelectedCreatorIds,
    domains,
    selectedDomainIds,
    setSelectedDomainIds,
    hasActiveFilters,
    clearAllFilters,
  };
}
