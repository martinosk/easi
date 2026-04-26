import { Badge } from '@mantine/core';
import { useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { selectDynamicAdditions } from '../../../store/slices/dynamicModeSlice';
import { useCapabilities, useRealizations } from '../../capabilities/hooks/useCapabilities';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { useRelations } from '../../relations/hooks/useRelations';
import {
  getUnexpandedByEdgeType,
  type EdgeType,
  type EntityRef,
  type EntityType,
} from '../utils/dynamicMode';
import { ExpandPopover } from './ExpandPopover';

interface DynamicExpandBadgeProps {
  entityId: string;
  entityType: EntityType;
  entityName: string;
}

export function DynamicExpandBadge({ entityId, entityType, entityName }: DynamicExpandBadgeProps) {
  const enabled = useAppStore((s) => s.dynamicEnabled);
  const dynamicEntities = useAppStore((s) => s.dynamicEntities);
  const filters = useAppStore((s) => s.dynamicFilters);
  const draftAddEntities = useAppStore((s) => s.draftAddEntities);

  const { data: relations = [] } = useRelations();
  const { data: capabilities = [] } = useCapabilities();
  const { data: realizations = [] } = useRealizations();
  const { data: originRelationships = [] } = useOriginRelationshipsQuery();

  const [popoverOpen, setPopoverOpen] = useState(false);

  if (!enabled) return null;

  const includedIds = new Set(dynamicEntities.map((e) => e.id));
  const breakdown = getUnexpandedByEdgeType(
    { relations, capabilities, realizations, originRelationships },
    { id: entityId, type: entityType },
    includedIds,
    filters,
  );
  const total =
    (filters.edges.relation ? breakdown.relation.length : 0) +
    (filters.edges.realization ? breakdown.realization.length : 0) +
    (filters.edges.parentage ? breakdown.parentage.length : 0) +
    (filters.edges.origin ? breakdown.origin.length : 0);

  if (total === 0) return null;

  const inferType = (edge: EdgeType, sourceType: EntityType): EntityType => {
    if (edge === 'relation') return 'component';
    if (edge === 'parentage') return 'capability';
    if (edge === 'realization') return sourceType === 'component' ? 'capability' : 'component';
    return sourceType === 'component' ? 'originEntity' : 'component';
  };

  const idsToRefs = (ids: string[], edge: EdgeType): EntityRef[] =>
    ids.map((id) => ({ id, type: inferType(edge, entityType) }));

  const handleExpandEdgeType = (edge: EdgeType) => {
    draftAddEntities(idsToRefs(breakdown[edge], edge));
    setPopoverOpen(false);
  };

  const handleExpandAll = () => {
    const refs: EntityRef[] = [];
    (['relation', 'realization', 'parentage', 'origin'] as EdgeType[]).forEach((et) => {
      if (filters.edges[et]) refs.push(...idsToRefs(breakdown[et], et));
    });
    draftAddEntities(refs);
    setPopoverOpen(false);
  };

  return (
    <ExpandPopover
      entityName={entityName}
      breakdown={breakdown}
      enabledEdgeTypes={filters.edges}
      opened={popoverOpen}
      onClose={() => setPopoverOpen(false)}
      onExpandEdgeType={handleExpandEdgeType}
      onExpandAll={handleExpandAll}
    >
      <Badge
        size="md"
        variant="filled"
        color="blue"
        style={{
          position: 'absolute',
          top: -8,
          right: -8,
          cursor: 'pointer',
          zIndex: 5,
        }}
        aria-label={`Expand ${entityName} (+${total})`}
        onClick={(e) => {
          e.stopPropagation();
          setPopoverOpen((o) => !o);
        }}
      >
        +{total}
      </Badge>
    </ExpandPopover>
  );
}

// Re-export for clarity at usage site
export { selectDynamicAdditions };
