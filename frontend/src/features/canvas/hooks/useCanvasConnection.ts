import { useCallback } from 'react';
import type { Connection } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import toast from 'react-hot-toast';

export const useCanvasConnection = (
  onConnect: (source: string, target: string) => void
) => {
  const changeCapabilityParent = useAppStore((state) => state.changeCapabilityParent);
  const linkSystemToCapability = useAppStore((state) => state.linkSystemToCapability);

  const onConnectHandler = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target) return;

      const sourceIsCapability = connection.source.startsWith('cap-');
      const targetIsCapability = connection.target.startsWith('cap-');

      if (sourceIsCapability && targetIsCapability) {
        const parentId = connection.target.replace('cap-', '');
        const childId = connection.source.replace('cap-', '');

        try {
          await changeCapabilityParent(childId, parentId);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to create parent relationship';
          if (errorMessage.includes('depth') || errorMessage.includes('L5')) {
            toast.error('Cannot create this parent relationship: would result in hierarchy deeper than L4');
          }
        }
      } else if (!sourceIsCapability && !targetIsCapability) {
        onConnect(connection.target, connection.source);
      } else {
        const capabilityId = sourceIsCapability
          ? connection.source.replace('cap-', '')
          : connection.target.replace('cap-', '');
        const componentId = sourceIsCapability ? connection.target : connection.source;

        try {
          await linkSystemToCapability(capabilityId, {
            componentId,
            realizationLevel: 'Full',
          });
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to create realization';
          toast.error(errorMessage);
        }
      }
    },
    [onConnect, changeCapabilityParent, linkSystemToCapability]
  );

  return { onConnectHandler };
};
