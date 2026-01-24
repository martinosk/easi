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
} from '../../../api/types';
import { isOriginEntityNode, getOriginEntityTypeFromNodeId, extractOriginEntityId } from '../utils/nodeFactory';
import toast from 'react-hot-toast';

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
      notes?: string
    ) => {
      const mutations = {
        acquired: () =>
          linkComponentToAcquiredEntityMutation.mutateAsync({
            componentId,
            entityId: toAcquiredEntityId(entityId),
            notes,
          }),
        vendor: () =>
          linkComponentToVendorMutation.mutateAsync({
            componentId,
            vendorId: toVendorId(entityId),
            notes,
          }),
        team: () =>
          linkComponentToInternalTeamMutation.mutateAsync({
            componentId,
            teamId: toInternalTeamId(entityId),
            notes,
          }),
      };
      await mutations[originEntityType]();
    },
    [linkComponentToAcquiredEntityMutation, linkComponentToVendorMutation, linkComponentToInternalTeamMutation]
  );

  const handleOriginComponentConnection = useCallback(
    async (source: string, target: string, notes?: string): Promise<void> => {
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
        await linkOriginEntity(originEntityType, componentId, entityId, notes);
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create origin relationship');
        toast.error(errorMessage);
      }
    },
    [linkOriginEntity]
  );

  return { handleOriginComponentConnection };
};
