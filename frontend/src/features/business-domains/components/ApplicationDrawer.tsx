import { Drawer } from '@mantine/core';
import type { ComponentId } from '../../../api/types';
import { ComponentDetailsPanel } from '../../components/components/ComponentDetailsPanel';

export interface ApplicationDrawerProps {
  componentId: ComponentId | null;
  onClose: () => void;
}

export function ApplicationDrawer({ componentId, onClose }: ApplicationDrawerProps) {
  return (
    <Drawer
      opened={componentId !== null}
      onClose={onClose}
      position="right"
      size="md"
      title="Application"
      data-testid="application-drawer"
    >
      {componentId && <ComponentDetailsPanel componentId={componentId} />}
    </Drawer>
  );
}
