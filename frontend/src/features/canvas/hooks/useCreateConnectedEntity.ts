import { useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import toast from 'react-hot-toast';
import { httpClient } from '../../../api/core/httpClient';
import type { ComponentId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { componentsMutationEffects } from '../../components/mutationEffects';
import { originRelationshipsQueryKeys } from '../../origin-entities/queryKeys';
import { relationsMutationEffects } from '../../relations/mutationEffects';
import { useViewOperations } from '../../views/hooks/useViewOperations';
import type {
  ConnectedEntityActionType,
  CreateConnectedEntitySubmitData,
} from '../components/dialogs/CreateConnectedEntityDialog';
import { positionFromHandle } from '../utils/handleCalculation';

const PLACEMENT_OFFSET_PX = 250;

export function useCreateConnectedEntity(
  sourceNodeId: string,
  sourcePosition: { x: number; y: number },
  handlePosition: string,
) {
  const queryClient = useQueryClient();
  const { addComponentToView } = useViewOperations();

  const createConnectedEntity = useCallback(
    async (data: CreateConnectedEntitySubmitData) => {
      try {
        const componentResponse = await httpClient.post<{ id: string; name: string }>(
          '/api/v1/components',
          { name: data.name, description: data.description },
        );
        const newComponent = componentResponse.data;
        const newComponentId = newComponent.id as ComponentId;

        if (data.actionType === 'x-add-relation') {
          await httpClient.post(data.actionLink.href, {
            sourceComponentId: sourceNodeId,
            targetComponentId: newComponentId,
            relationType: data.relationType,
          });
        } else if (isOriginAction(data.actionType)) {
          await httpClient.put(data.actionLink.href, {
            componentId: sourceNodeId,
            [`${originEntityFieldName(data.actionType)}`]: newComponentId,
          });
        }

        const newPos = positionFromHandle(sourcePosition, handlePosition, PLACEMENT_OFFSET_PX);
        await addComponentToView(newComponentId, newPos.x, newPos.y);

        invalidateFor(queryClient, [
          ...componentsMutationEffects.create(),
          ...relationsMutationEffects.create(),
          originRelationshipsQueryKeys.lists(),
        ]);

        toast.success(`"${newComponent.name}" created and connected`);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to create connected entity';
        toast.error(message);
        throw error;
      }
    },
    [sourceNodeId, sourcePosition, handlePosition, queryClient, addComponentToView],
  );

  return { createConnectedEntity };
}

function isOriginAction(actionType: ConnectedEntityActionType): boolean {
  return actionType.startsWith('x-set-origin-');
}

function originEntityFieldName(actionType: ConnectedEntityActionType): string {
  switch (actionType) {
    case 'x-set-origin-acquired-via':
      return 'acquiredEntityId';
    case 'x-set-origin-purchased-from':
      return 'vendorId';
    case 'x-set-origin-built-by':
      return 'teamId';
    default:
      return 'entityId';
  }
}
