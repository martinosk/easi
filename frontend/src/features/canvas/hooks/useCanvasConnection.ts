import type { Connection } from '@xyflow/react';
import { useCallback } from 'react';
import { isOriginEntity, toNodeId } from '../../../constants/entityIdentifiers';
import { isCapabilityNode, useCapabilityConnection } from './useCapabilityConnection';
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
  if (isOriginEntity(toNodeId(nodeId))) return 'origin';
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

export const useCanvasConnection = (onConnect: (source: string, target: string) => void) => {
  const { handleCapabilityParentConnection, handleCapabilityComponentConnection } = useCapabilityConnection();
  const { handleOriginComponentConnection } = useOriginConnection();

  const dispatchByType = useCallback(
    async (type: ConnectionType, source: string, target: string): Promise<void> => {
      const dispatchers: Record<ConnectionType, () => Promise<void> | void> = {
        'capability-to-capability': () => handleCapabilityParentConnection(source, target),
        'component-to-component': () => onConnect(target, source),
        'capability-component-mixed': () => handleCapabilityComponentConnection(source, target),
        'origin-component-mixed': () => handleOriginComponentConnection(source, target),
        invalid: () => undefined,
      };
      await dispatchers[type]();
    },
    [onConnect, handleCapabilityParentConnection, handleCapabilityComponentConnection, handleOriginComponentConnection],
  );

  const onConnectHandler = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target) return;
      if (connection.source === connection.target) return;
      await dispatchByType(getConnectionType(connection.source, connection.target), connection.source, connection.target);
    },
    [dispatchByType],
  );

  return { onConnectHandler };
};
