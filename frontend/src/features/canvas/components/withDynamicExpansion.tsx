import type { ComponentType } from 'react';
import { getEntityId, getEntityType, toNodeId } from '../../../constants/entityIdentifiers';
import { useIsDraftActiveForCurrentView } from '../../views/hooks/useIsDraftActiveForCurrentView';
import type { EntityType } from '../utils/dynamicMode';
import { DynamicExpandBadge } from './DynamicExpandBadge';

interface NodeProps<TData> {
  data: TData;
  id: string;
  selected?: boolean;
}

const NODE_TYPE_TO_ENTITY_TYPE: Record<string, EntityType> = {
  component: 'component',
  capability: 'capability',
  acquired: 'originEntity',
  vendor: 'originEntity',
  team: 'originEntity',
};

export function withDynamicExpansion<TData extends { label: string }>(Inner: ComponentType<NodeProps<TData>>) {
  function Wrapped(props: NodeProps<TData>) {
    const draftActive = useIsDraftActiveForCurrentView();

    if (!draftActive) return <Inner {...props} />;

    const nodeType = getEntityType(toNodeId(props.id));
    const entityId = getEntityId(toNodeId(props.id));
    const entityType = NODE_TYPE_TO_ENTITY_TYPE[nodeType] ?? 'component';

    return (
      <>
        <Inner {...props} />
        <DynamicExpandBadge entityId={entityId} entityType={entityType} entityName={props.data.label} />
      </>
    );
  }
  Wrapped.displayName = `withDynamicExpansion(${Inner.displayName ?? Inner.name ?? 'Component'})`;
  return Wrapped;
}
