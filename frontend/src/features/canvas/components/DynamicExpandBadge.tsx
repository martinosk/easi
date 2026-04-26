import { Badge } from '@mantine/core';
import { useState } from 'react';
import { useAppStore } from '../../../store/appStore';
import { selectDynamicAdditions } from '../../../store/slices/dynamicModeSlice';
import { useCapabilities, useRealizations } from '../../capabilities/hooks/useCapabilities';
import { useOriginRelationshipsQuery } from '../../origin-entities/hooks/useOriginRelationships';
import { useRelations } from '../../relations/hooks/useRelations';
import {
  getUnexpandedByEdgeType,
  type DynamicFilters,
  type EdgeType,
  type EntityRef,
  type EntityType,
  type UnexpandedByEdgeType,
} from '../utils/dynamicMode';
import { ExpandPopover } from './ExpandPopover';

interface DynamicExpandBadgeProps {
  entityId: string;
  entityType: EntityType;
  entityName: string;
}

const EDGE_ORDER: EdgeType[] = ['relation', 'realization', 'parentage', 'origin'];

const REALIZATION_TARGET: Record<EntityType, EntityType> = {
  component: 'capability',
  capability: 'component',
  originEntity: 'component',
};

const ORIGIN_TARGET: Record<EntityType, EntityType> = {
  component: 'originEntity',
  capability: 'component',
  originEntity: 'component',
};

function inferNeighborType(edge: EdgeType, sourceType: EntityType): EntityType {
  if (edge === 'relation') return 'component';
  if (edge === 'parentage') return 'capability';
  if (edge === 'realization') return REALIZATION_TARGET[sourceType];
  return ORIGIN_TARGET[sourceType];
}

function totalEnabled(breakdown: UnexpandedByEdgeType, filters: DynamicFilters): number {
  return EDGE_ORDER.reduce((sum, et) => sum + (filters.edges[et] ? breakdown[et].length : 0), 0);
}

function buildAllRefs(
  breakdown: UnexpandedByEdgeType,
  filters: DynamicFilters,
  sourceType: EntityType,
): EntityRef[] {
  return EDGE_ORDER.filter((et) => filters.edges[et]).flatMap((et) =>
    breakdown[et].map((id) => ({ id, type: inferNeighborType(et, sourceType) })),
  );
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
  const total = totalEnabled(breakdown, filters);

  if (total === 0) return null;

  const handleExpandEdgeType = (edge: EdgeType) => {
    draftAddEntities(breakdown[edge].map((id) => ({ id, type: inferNeighborType(edge, entityType) })));
    setPopoverOpen(false);
  };

  const handleExpandAll = () => {
    draftAddEntities(buildAllRefs(breakdown, filters, entityType));
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
        style={{ position: 'absolute', top: -8, right: -8, cursor: 'pointer', zIndex: 5 }}
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

export { selectDynamicAdditions };
