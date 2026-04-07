import { useCallback } from 'react';
import toast from 'react-hot-toast';
import { toCapabilityId, toComponentId } from '../../../api/types';
import {
  useCapabilities,
  useChangeCapabilityParent,
  useLinkSystemToCapability,
} from '../../capabilities/hooks/useCapabilities';

const CAPABILITY_PREFIX = 'cap-';

export const isCapabilityNode = (nodeId: string): boolean => nodeId.startsWith(CAPABILITY_PREFIX);

export const extractCapabilityId = (nodeId: string): string => nodeId.replace(CAPABILITY_PREFIX, '');

const getErrorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback;

const isHierarchyDepthError = (message: string): boolean => message.includes('depth') || message.includes('L5');

export const useCapabilityConnection = () => {
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const linkSystemToCapabilityMutation = useLinkSystemToCapability();
  const { data: capabilities = [] } = useCapabilities();

  const handleCapabilityParentConnection = useCallback(
    async (source: string, target: string) => {
      const parentId = toCapabilityId(extractCapabilityId(target));
      const childId = toCapabilityId(extractCapabilityId(source));
      const oldParentId = capabilities.find((capability) => capability.id === childId)?.parentId ?? undefined;

      try {
        await changeCapabilityParentMutation.mutateAsync({
          id: childId,
          oldParentId,
          newParentId: parentId,
        });
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create parent relationship');
        if (isHierarchyDepthError(errorMessage)) {
          toast.error('Cannot create this parent relationship: would result in hierarchy deeper than L4');
        }
      }
    },
    [capabilities, changeCapabilityParentMutation],
  );

  const handleCapabilityComponentConnection = useCallback(
    async (source: string, target: string) => {
      const sourceIsCapability = isCapabilityNode(source);
      const capabilityId = toCapabilityId(
        sourceIsCapability ? extractCapabilityId(source) : extractCapabilityId(target),
      );
      const componentId = toComponentId(sourceIsCapability ? target : source);

      try {
        await linkSystemToCapabilityMutation.mutateAsync({
          capabilityId,
          request: {
            componentId,
            realizationLevel: 'Full',
          },
        });
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create realization');
        toast.error(errorMessage);
      }
    },
    [linkSystemToCapabilityMutation],
  );

  return { handleCapabilityParentConnection, handleCapabilityComponentConnection };
};
