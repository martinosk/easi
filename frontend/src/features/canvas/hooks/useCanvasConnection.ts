import { useCallback } from 'react';
import type { Connection } from '@xyflow/react';
import { isOriginEntityNode } from '../utils/nodeFactory';
import { useCapabilityConnection, isCapabilityNode } from './useCapabilityConnection';
import { useOriginConnection } from './useOriginConnection';

type ConnectionType =
  | 'capability-to-capability'
  | 'component-to-component'
  | 'capability-component-mixed'
  | 'origin-component-mixed'
  | 'invalid';

type NodeKind = 'capability' | 'component' | 'origin';

const getNodeKind = (nodeId: string): NodeKind => {
  if (isCapabilityNode(nodeId)) return 'capability';
  if (isOriginEntityNode(nodeId)) return 'origin';
  return 'component';
};

const CONNECTION_TYPE_MAP: Record<`${NodeKind}-${NodeKind}`, ConnectionType> = {
  'capability-capability': 'capability-to-capability',
  'component-component': 'component-to-component',
  'capability-component': 'capability-component-mixed',
  'component-capability': 'capability-component-mixed',
  'origin-component': 'origin-component-mixed',
  'component-origin': 'origin-component-mixed',
  'capability-origin': 'invalid',
  'origin-capability': 'invalid',
  'origin-origin': 'invalid',
};

const getConnectionType = (source: string, target: string): ConnectionType => {
  const sourceKind = getNodeKind(source);
  const targetKind = getNodeKind(target);
  return CONNECTION_TYPE_MAP[`${sourceKind}-${targetKind}`];
};

export const useCanvasConnection = (
  onConnect: (source: string, target: string) => void
) => {
  const { handleCapabilityParentConnection, handleCapabilityComponentConnection } = useCapabilityConnection();
  const { handleOriginComponentConnection } = useOriginConnection();

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
