import { Center, Drawer, Group, Loader, Text } from '@mantine/core';
import { useCallback, useState } from 'react';
import type { Capability, CapabilityRealization, Component, ComponentId } from '../../../api/types';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { ComponentDetailsContent } from '../../components/components/ComponentDetails';
import { EditComponentDialog } from '../../components/components/EditComponentDialog';
import { useComponents } from '../../components/hooks/useComponents';
import { useComponentDetails } from '../hooks/useComponentDetails';

export interface ApplicationDrawerProps {
  componentId: ComponentId | null;
  onClose: () => void;
}

function DrawerLoadingState() {
  return (
    <Center py="xl">
      <Group gap="xs">
        <Loader size="sm" />
        <Text c="dimmed">Loading...</Text>
      </Group>
    </Center>
  );
}

function DrawerErrorState() {
  return (
    <Center py="xl">
      <Text c="red">Failed to load application details</Text>
    </Center>
  );
}

interface DrawerBodyProps {
  isLoading: boolean;
  loadFailed: boolean;
  component: Component | null;
  componentRealizations: CapabilityRealization[];
  capabilities: Capability[];
  editOpen: boolean;
  onEdit: () => void;
  onEditClose: () => void;
}

function DrawerBody({
  isLoading,
  loadFailed,
  component,
  componentRealizations,
  capabilities,
  editOpen,
  onEdit,
  onEditClose,
}: DrawerBodyProps) {
  if (isLoading) return <DrawerLoadingState />;
  if (loadFailed) return <DrawerErrorState />;
  if (!component) return null;

  return (
    <>
      <ComponentDetailsContent
        component={component}
        realizations={componentRealizations}
        capabilities={capabilities}
        onEdit={onEdit}
      />
      <EditComponentDialog isOpen={editOpen} onClose={onEditClose} component={component} />
    </>
  );
}

export function ApplicationDrawer({ componentId, onClose }: ApplicationDrawerProps) {
  const [editOpen, setEditOpen] = useState(false);
  const { data: storeComponents = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: componentRealizations = [] } = useCapabilitiesByComponent(componentId ?? undefined);

  const componentFromStore = storeComponents.find((c) => c.id === componentId);
  const {
    component: componentFromApi,
    isLoading,
    error,
  } = useComponentDetails(componentFromStore ? null : componentId);
  const component = componentFromStore || componentFromApi;
  const loadFailed = Boolean(componentId) && Boolean(error || !component);

  const handleEdit = useCallback(() => setEditOpen(true), []);

  return (
    <Drawer
      opened={componentId !== null}
      onClose={onClose}
      position="right"
      size="md"
      title="Application"
      data-testid="application-drawer"
    >
      <DrawerBody
        isLoading={isLoading}
        loadFailed={loadFailed}
        component={component}
        componentRealizations={componentRealizations}
        capabilities={capabilities}
        editOpen={editOpen}
        onEdit={handleEdit}
        onEditClose={() => setEditOpen(false)}
      />
    </Drawer>
  );
}
