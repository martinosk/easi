import { useCallback } from 'react';
import type { Connection } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import toast from 'react-hot-toast';

const CAPABILITY_PREFIX = 'cap-';

const isCapabilityNode = (nodeId: string): boolean => nodeId.startsWith(CAPABILITY_PREFIX);

const extractCapabilityId = (nodeId: string): string => nodeId.replace(CAPABILITY_PREFIX, '');

const getErrorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback;

const isHierarchyDepthError = (message: string): boolean =>
  message.includes('depth') || message.includes('L5');

type ConnectionType = 'capability-to-capability' | 'component-to-component' | 'mixed';

const getConnectionType = (sourceIsCapability: boolean, targetIsCapability: boolean): ConnectionType => {
  if (sourceIsCapability && targetIsCapability) return 'capability-to-capability';
  if (!sourceIsCapability && !targetIsCapability) return 'component-to-component';
  return 'mixed';
};

export const useCanvasConnection = (
  onConnect: (source: string, target: string) => void
) => {
  const changeCapabilityParent = useAppStore((state) => state.changeCapabilityParent);
  const linkSystemToCapability = useAppStore((state) => state.linkSystemToCapability);

  const handleCapabilityParentConnection = useCallback(
    async (source: string, target: string) => {
      const parentId = extractCapabilityId(target);
      const childId = extractCapabilityId(source);

      try {
        await changeCapabilityParent(childId, parentId);
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create parent relationship');
        if (isHierarchyDepthError(errorMessage)) {
          toast.error('Cannot create this parent relationship: would result in hierarchy deeper than L4');
        }
      }
    },
    [changeCapabilityParent]
  );

  const handleMixedConnection = useCallback(
    async (source: string, target: string, sourceIsCapability: boolean) => {
      const capabilityId = sourceIsCapability
        ? extractCapabilityId(source)
        : extractCapabilityId(target);
      const componentId = sourceIsCapability ? target : source;

      try {
        await linkSystemToCapability(capabilityId, {
          componentId,
          realizationLevel: 'Full',
        });
      } catch (error) {
        const errorMessage = getErrorMessage(error, 'Failed to create realization');
        toast.error(errorMessage);
      }
    },
    [linkSystemToCapability]
  );

  const onConnectHandler = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target) return;

      const sourceIsCapability = isCapabilityNode(connection.source);
      const targetIsCapability = isCapabilityNode(connection.target);
      const connectionType = getConnectionType(sourceIsCapability, targetIsCapability);

      switch (connectionType) {
        case 'capability-to-capability':
          await handleCapabilityParentConnection(connection.source, connection.target);
          break;
        case 'component-to-component':
          onConnect(connection.target, connection.source);
          break;
        case 'mixed':
          await handleMixedConnection(connection.source, connection.target, sourceIsCapability);
          break;
      }
    },
    [onConnect, handleCapabilityParentConnection, handleMixedConnection]
  );

  return { onConnectHandler };
};
