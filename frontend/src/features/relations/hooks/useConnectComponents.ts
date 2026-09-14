import { type Component, type CreateRelationRequest, toComponentId } from '../../../api/types';
import type { ConnectComponentsFormData } from '../../../lib/schemas';
import { useAttachComponent } from '../../components/hooks/useComponentContainment';
import { isContainmentKind } from '../utils/connectionKinds';
import { useCreateRelation } from './useRelations';

function relationRequest(data: ConnectComponentsFormData, relationType: 'Triggers' | 'Serves'): CreateRelationRequest {
  return {
    sourceComponentId: toComponentId(data.sourceComponentId),
    targetComponentId: toComponentId(data.targetComponentId),
    relationType,
    name: data.name || undefined,
    description: data.description || undefined,
  };
}

export function useConnectComponents() {
  const createRelationMutation = useCreateRelation();
  const attachMutation = useAttachComponent();

  const connect = async (data: ConnectComponentsFormData, part: Component | undefined): Promise<unknown> => {
    const kind = data.connectionKind;
    if (!isContainmentKind(kind)) {
      return createRelationMutation.mutateAsync(relationRequest(data, kind));
    }
    if (!part) throw new Error('Part not found');
    return attachMutation.mutateAsync({
      component: part,
      request: { parentId: toComponentId(data.targetComponentId), kind },
    });
  };

  return { connect, isPending: createRelationMutation.isPending || attachMutation.isPending };
}
