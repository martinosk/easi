import { Box, Button, Center, Group, Loader, Stack, Text, Title } from '@mantine/core';
import { useCallback, useState } from 'react';
import type { BusinessDomain, Capability, ComponentId } from '../../../api/types';
import { EditCapabilityDialog } from '../../capabilities/components/EditCapabilityDialog';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { ComponentDetailsContent } from '../../components/components/ComponentDetails';
import { EditComponentDialog } from '../../components/components/EditComponentDialog';
import { useComponents } from '../../components/hooks/useComponents';
import { useComponentDetails } from '../hooks/useComponentDetails';
import { StrategicImportanceSection } from './StrategicImportanceSection';

interface DetailsSidebarProps {
  selectedCapability: Capability | null;
  selectedComponentId: ComponentId | null;
  visualizedDomain: BusinessDomain | null;
}

function PanelShell({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Stack gap="sm" p="md">
      <Title order={4}>{title}</Title>
      {children}
    </Stack>
  );
}

function EmptyState() {
  return (
    <PanelShell title="Details">
      <Center py="xl">
        <Text c="dimmed">Select a capability or application to view details</Text>
      </Center>
    </PanelShell>
  );
}

interface CapabilityContentProps {
  capability: Capability;
  domain: BusinessDomain | null;
}

function CapabilityField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <Stack gap={2}>
      <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
        {label}
      </Text>
      <Text size="sm">{children}</Text>
    </Stack>
  );
}

function CapabilityContent({ capability, domain }: CapabilityContentProps) {
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const canEdit = capability._links?.edit !== undefined;

  return (
    <PanelShell title="Capability Details">
      {canEdit && (
        <Group justify="flex-end">
          <Button variant="default" size="xs" onClick={() => setEditDialogOpen(true)}>
            Edit
          </Button>
        </Group>
      )}
      <CapabilityField label="Name">{capability.name}</CapabilityField>
      <CapabilityField label="Level">{capability.level}</CapabilityField>
      {capability.description && <CapabilityField label="Description">{capability.description}</CapabilityField>}
      {domain && <StrategicImportanceSection domain={domain} capabilityId={capability.id} />}
      <EditCapabilityDialog
        isOpen={editDialogOpen}
        onClose={() => setEditDialogOpen(false)}
        capability={capability}
      />
    </PanelShell>
  );
}

interface ApplicationContentProps {
  componentId: ComponentId;
}

function ApplicationContent({ componentId }: ApplicationContentProps) {
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  const { data: storeComponents = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: componentRealizations = [] } = useCapabilitiesByComponent(componentId);

  const componentFromStore = storeComponents.find((c) => c.id === componentId);
  const {
    component: componentFromApi,
    isLoading,
    error,
  } = useComponentDetails(componentFromStore ? null : componentId);

  const component = componentFromStore || componentFromApi;

  const handleEdit = useCallback(() => {
    setEditDialogOpen(true);
  }, []);

  const handleCloseEditDialog = useCallback(() => {
    setEditDialogOpen(false);
  }, []);

  if (isLoading) {
    return (
      <PanelShell title="Application Details">
        <Center py="xl">
          <Group gap="xs">
            <Loader size="sm" />
            <Text c="dimmed">Loading...</Text>
          </Group>
        </Center>
      </PanelShell>
    );
  }

  if (error || !component) {
    return (
      <PanelShell title="Application Details">
        <Center py="xl">
          <Text c="red">Failed to load application details</Text>
        </Center>
      </PanelShell>
    );
  }

  return (
    <>
      <ComponentDetailsContent
        component={component}
        realizations={componentRealizations}
        capabilities={capabilities}
        onEdit={handleEdit}
      />
      <EditComponentDialog isOpen={editDialogOpen} onClose={handleCloseEditDialog} component={component} />
    </>
  );
}

export function DetailsSidebar({ selectedCapability, selectedComponentId, visualizedDomain }: DetailsSidebarProps) {
  const hasSelection = selectedCapability || selectedComponentId;

  return (
    <Box component="aside" bg="white" h="100%" w="100%" style={{ overflow: 'auto' }}>
      {!hasSelection && <EmptyState />}
      {selectedCapability && <CapabilityContent capability={selectedCapability} domain={visualizedDomain} />}
      {selectedComponentId && !selectedCapability && <ApplicationContent componentId={selectedComponentId} />}
    </Box>
  );
}
