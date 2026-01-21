import { useCallback } from 'react';
import {
  useLinkComponentToAcquiredEntity,
  useLinkComponentToVendor,
  useLinkComponentToInternalTeam,
} from '../../origin-entities/hooks';
import {
  toComponentId,
  toAcquiredEntityId,
  toVendorId,
  toInternalTeamId,
  ApiError,
  type RelationshipConflictError,
} from '../../../api/types';
import { isOriginEntityNode, getOriginEntityTypeFromNodeId, extractOriginEntityId } from '../utils/nodeFactory';
import toast from 'react-hot-toast';

const isRelationshipConflict = (error: unknown): error is ApiError & { response: RelationshipConflictError } => {
  return error instanceof ApiError && error.statusCode === 409 && error.response !== undefined;
};

const getErrorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback;

export const useOriginConnection = () => {
  const linkComponentToAcquiredEntityMutation = useLinkComponentToAcquiredEntity();
  const linkComponentToVendorMutation = useLinkComponentToVendor();
  const linkComponentToInternalTeamMutation = useLinkComponentToInternalTeam();

  const linkOriginEntity = useCallback(
    async (
      originEntityType: 'acquired' | 'vendor' | 'team',
      componentId: ReturnType<typeof toComponentId>,
      entityId: string,
      replaceExisting: boolean
    ) => {
      const mutations = {
        acquired: () =>
          linkComponentToAcquiredEntityMutation.mutateAsync({
            componentId,
            entityId: toAcquiredEntityId(entityId),
            replaceExisting,
          }),
        vendor: () =>
          linkComponentToVendorMutation.mutateAsync({
            componentId,
            vendorId: toVendorId(entityId),
            replaceExisting,
          }),
        team: () =>
          linkComponentToInternalTeamMutation.mutateAsync({
            componentId,
            teamId: toInternalTeamId(entityId),
            replaceExisting,
          }),
      };
      await mutations[originEntityType]();
    },
    [linkComponentToAcquiredEntityMutation, linkComponentToVendorMutation, linkComponentToInternalTeamMutation]
  );

  const handleOriginComponentConnection = useCallback(
    async (source: string, target: string, replaceExisting: boolean = false): Promise<void> => {
      const sourceIsOriginEntity = isOriginEntityNode(source);
      const originNodeId = sourceIsOriginEntity ? source : target;
      const componentNodeId = sourceIsOriginEntity ? target : source;

      const originEntityType = getOriginEntityTypeFromNodeId(originNodeId);
      const entityId = extractOriginEntityId(originNodeId);
      const componentId = toComponentId(componentNodeId);

      if (!entityId || !originEntityType) {
        toast.error('Invalid origin entity');
        return;
      }

      try {
        await linkOriginEntity(originEntityType, componentId, entityId, replaceExisting);
      } catch (error) {
        if (isRelationshipConflict(error)) {
          const conflict = error.response;
          const confirmed = window.confirm(
            `This component is already linked to "${conflict.originEntityName}". ` +
              `Do you want to replace it with the new connection?`
          );
          if (confirmed) {
            await handleOriginComponentConnection(source, target, true);
          }
          return;
        }
        const errorMessage = getErrorMessage(error, 'Failed to create origin relationship');
        toast.error(errorMessage);
      }
    },
    [linkOriginEntity]
  );

  return { handleOriginComponentConnection };
};
