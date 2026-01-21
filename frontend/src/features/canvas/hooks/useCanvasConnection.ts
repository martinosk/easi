import { useCallback } from 'react';
import type { Connection } from '@xyflow/react';
import { useChangeCapabilityParent, useLinkSystemToCapability } from '../../capabilities/hooks/useCapabilities';
import {
  useLinkComponentToAcquiredEntity,
  useLinkComponentToVendor,
  useLinkComponentToInternalTeam,
} from '../../origin-entities/hooks';
import {
  toComponentId,
  toCapabilityId,
  toAcquiredEntityId,
  toVendorId,
  toInternalTeamId,
} from '../../../api/types';
import { isOriginEntityNode, getOriginEntityTypeFromNodeId, extractOriginEntityId } from '../utils/nodeFactory';
import toast from 'react-hot-toast';

const CAPABILITY_PREFIX = 'cap-';

const isCapabilityNode = (nodeId: string): boolean => nodeId.startsWith(CAPABILITY_PREFIX);

const isComponentNode = (nodeId: string): boolean =>
  !isCapabilityNode(nodeId) && !isOriginEntityNode(nodeId);

const extractCapabilityId = (nodeId: string): string => nodeId.replace(CAPABILITY_PREFIX, '');

const getErrorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback;

const isHierarchyDepthError = (message: string): boolean =>
  message.includes('depth') || message.includes('L5');

type ConnectionType =
  | 'capability-to-capability'
  | 'component-to-component'
  | 'capability-component-mixed'
  | 'origin-component-mixed'
  | 'invalid';

const getConnectionType = (
  source: string,
  target: string
): ConnectionType => {
  const sourceIsCapability = isCapabilityNode(source);
  const targetIsCapability = isCapabilityNode(target);
  const sourceIsOriginEntity = isOriginEntityNode(source);
  const targetIsOriginEntity = isOriginEntityNode(target);
  const sourceIsComponent = isComponentNode(source);
  const targetIsComponent = isComponentNode(target);

  if (sourceIsCapability && targetIsCapability) return 'capability-to-capability';
  if (sourceIsComponent && targetIsComponent) return 'component-to-component';
  if ((sourceIsCapability && targetIsComponent) || (sourceIsComponent && targetIsCapability)) {
    return 'capability-component-mixed';
  }
  if ((sourceIsOriginEntity && targetIsComponent) || (sourceIsComponent && targetIsOriginEntity)) {
    return 'origin-component-mixed';
  }
  return 'invalid';
};

export const useCanvasConnection = (
  onConnect: (source: string, target: string) => void
) => {
  const changeCapabilityParentMutation = useChangeCapabilityParent();
  const linkSystemToCapabilityMutation = useLinkSystemToCapability();
  const linkComponentToAcquiredEntityMutation = useLinkComponentToAcquiredEntity();
  const linkComponentToVendorMutation = useLinkComponentToVendor();
  const linkComponentToInternalTeamMutation = useLinkComponentToInternalTeam();

  const handleCapabilityParentConnection = useCallback(
    async (source: string, target: string) => {
      const parentId = toCapabilityId(extractCapabilityId(target));
      const childId = toCapabilityId(extractCapabilityId(source));

      try {
        await changeCapabilityParentMutation.mutateAsync({
          id: childId,
          newParentId: parentId,
        });
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create parent relationship');
        if (isHierarchyDepthError(errorMessage)) {
          toast.error('Cannot create this parent relationship: would result in hierarchy deeper than L4');
        }
      }
    },
    [changeCapabilityParentMutation]
  );

  const handleCapabilityComponentConnection = useCallback(
    async (source: string, target: string) => {
      const sourceIsCapability = isCapabilityNode(source);
      const capabilityId = toCapabilityId(
        sourceIsCapability ? extractCapabilityId(source) : extractCapabilityId(target)
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
    [linkSystemToCapabilityMutation]
  );

  const handleOriginComponentConnection = useCallback(
    async (source: string, target: string) => {
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
        switch (originEntityType) {
          case 'acquired':
            await linkComponentToAcquiredEntityMutation.mutateAsync({
              componentId,
              entityId: toAcquiredEntityId(entityId),
            });
            break;
          case 'vendor':
            await linkComponentToVendorMutation.mutateAsync({
              componentId,
              vendorId: toVendorId(entityId),
            });
            break;
          case 'team':
            await linkComponentToInternalTeamMutation.mutateAsync({
              componentId,
              teamId: toInternalTeamId(entityId),
            });
            break;
        }
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create origin relationship');
        toast.error(errorMessage);
      }
    },
    [linkComponentToAcquiredEntityMutation, linkComponentToVendorMutation, linkComponentToInternalTeamMutation]
  );

  const onConnectHandler = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target) return;

      const connectionType = getConnectionType(connection.source, connection.target);

      switch (connectionType) {
        case 'capability-to-capability':
          await handleCapabilityParentConnection(connection.source, connection.target);
          break;
        case 'component-to-component':
          onConnect(connection.target, connection.source);
          break;
        case 'capability-component-mixed':
          await handleCapabilityComponentConnection(connection.source, connection.target);
          break;
        case 'origin-component-mixed':
          await handleOriginComponentConnection(connection.source, connection.target);
          break;
        case 'invalid':
          break;
      }
    },
    [onConnect, handleCapabilityParentConnection, handleCapabilityComponentConnection, handleOriginComponentConnection]
  );

  return { onConnectHandler };
};
