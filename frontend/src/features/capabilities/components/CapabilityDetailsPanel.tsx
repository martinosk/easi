import { Divider, Stack, Title } from '@mantine/core';
import { useState } from 'react';
import type { CapabilityId } from '../../../api/types';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import { useStrategyImportanceByCapability } from '../../business-domains/hooks/useStrategyImportance';
import { useComponents } from '../../components/hooks/useComponents';
import { useCapabilities, useCapabilityRealizations } from '../hooks/useCapabilities';
import { AddExpertDialog } from './AddExpertDialog';
import { CapabilityContent } from './CapabilityDetails';
import { EditCapabilityDialog } from './EditCapabilityDialog';

interface CapabilityDetailsPanelProps {
  capabilityId: string;
}

const noop = () => {};

export function CapabilityDetailsPanel({ capabilityId }: CapabilityDetailsPanelProps) {
  const typedId = capabilityId as CapabilityId;
  const capabilitiesQuery = useCapabilities();
  const { data: components = [] } = useComponents();
  const { data: realizations = [] } = useCapabilityRealizations(typedId);
  const { data: importanceRatings = [] } = useStrategyImportanceByCapability(typedId);
  const [editOpen, setEditOpen] = useState(false);
  const [addExpertOpen, setAddExpertOpen] = useState(false);

  const capability = (capabilitiesQuery.data ?? []).find((c) => c.id === capabilityId);

  if (capabilitiesQuery.isLoading) return <DetailPanelLoading />;
  if (!capability) return <DetailPanelFailure message="Failed to load capability" />;

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Capability Details</Title>
      <Divider />

      <CapabilityContent
        capability={capability}
        capabilityInView={undefined}
        currentView={null}
        realizations={realizations}
        components={components}
        importanceRatings={importanceRatings}
        onColorChange={noop}
        onEdit={() => setEditOpen(true)}
        onRemoveFromView={noop}
        onAddExpert={() => setAddExpertOpen(true)}
      />

      <EditCapabilityDialog isOpen={editOpen} onClose={() => setEditOpen(false)} capability={capability} />
      <AddExpertDialog isOpen={addExpertOpen} onClose={() => setAddExpertOpen(false)} capabilityId={capability.id} />
    </Stack>
  );
}
