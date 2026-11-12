import { Alert, Badge, Button, Group, Modal, ScrollArea, Skeleton, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, CapabilityDeleteImpact, CapabilityId, ImpactRealization } from '../../../api/types';
import { useCascadeDeleteCapability, useDeleteCapability, useDeleteImpact } from '../hooks/useCapabilities';

interface DeleteCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
  onConfirm?: () => void;
  capabilitiesToDelete?: Capability[];
  domainId?: string;
}

const plural = (count: number, singular: string, pluralForm: string) => (count === 1 ? singular : pluralForm);

const ImpactLoading = () => (
  <Stack gap="sm" data-testid="delete-impact-loading">
    <Skeleton height={20} width="80%" />
    <Skeleton height={20} width="60%" />
    <Skeleton height={40} />
  </Stack>
);

const CascadeImpact = ({ capability, impact }: { capability: Capability; impact: CapabilityDeleteImpact }) => {
  const affectedCount = impact.affectedCapabilities.length;
  return (
    <>
      <Alert color="orange" data-testid="cascade-warning">
        This will delete{' '}
        <Text span fw={600}>
          "{capability.name}"
        </Text>{' '}
        and{' '}
        <Text span fw={600}>
          {affectedCount} child {plural(affectedCount, 'capability', 'capabilities')}
        </Text>
        .
      </Alert>

      {affectedCount > 0 && (
        <Stack gap="xs">
          <Text fw={600} size="sm">
            Affected Capabilities ({affectedCount})
          </Text>
          <ScrollArea.Autosize mah={150}>
            <Stack gap={4}>
              {impact.affectedCapabilities.map((cap) => (
                <Group key={cap.id} gap="xs">
                  <Badge size="xs" variant="light">
                    {cap.level}
                  </Badge>
                  <Text size="sm">{cap.name}</Text>
                </Group>
              ))}
            </Stack>
          </ScrollArea.Autosize>
        </Stack>
      )}
    </>
  );
};

const LeafConfirmation = ({ capability }: { capability: Capability }) => (
  <>
    <Text>Are you sure you want to delete</Text>
    <Text fw={600} size="lg">
      "{capability.name}"
    </Text>
  </>
);

const capabilityNames = (capability: Capability, impact: CapabilityDeleteImpact): Map<string, string> =>
  new Map([[capability.id, capability.name], ...impact.affectedCapabilities.map((c) => [c.id, c.name] as const)]);

const RealizationsToRemove = ({ capability, impact }: { capability: Capability; impact: CapabilityDeleteImpact }) => {
  const realizations: ImpactRealization[] = impact.realizations ?? [];
  if (realizations.length === 0) return null;

  const names = capabilityNames(capability, impact);
  return (
    <Stack gap="xs" data-testid="realizations-to-remove">
      <Text fw={600} size="sm">
        {realizations.length} application {plural(realizations.length, 'realization', 'realizations')} will be removed
      </Text>
      <ScrollArea.Autosize mah={150}>
        <Stack gap={4}>
          {realizations.map((r) => (
            <Text key={r.id} size="sm" data-testid="realization-to-remove">
              <Text span fw={600}>
                {r.componentName}
              </Text>{' '}
              stops realising{' '}
              <Text span fw={600}>
                {names.get(r.capabilityId) ?? r.capabilityName ?? r.capabilityId}
              </Text>
            </Text>
          ))}
        </Stack>
      </ScrollArea.Autosize>
      <Text size="sm" c="dimmed">
        The applications themselves are kept.
      </Text>
    </Stack>
  );
};

interface ImpactBodyProps {
  capability: Capability;
  impact: CapabilityDeleteImpact | undefined;
  impactLoading: boolean;
  multiDeleteCount: number;
}

const ImpactBody = ({ capability, impact, impactLoading, multiDeleteCount }: ImpactBodyProps) => {
  if (impactLoading) return <ImpactLoading />;
  if (multiDeleteCount > 1) return <Text>Are you sure you want to delete {multiDeleteCount} capabilities?</Text>;
  return (
    <>
      {impact?.hasDescendants ? (
        <CascadeImpact capability={capability} impact={impact} />
      ) : (
        <LeafConfirmation capability={capability} />
      )}
      {impact && <RealizationsToRemove capability={capability} impact={impact} />}
    </>
  );
};

const submitLabel = (capability: Capability, impact: CapabilityDeleteImpact | undefined) => {
  if (!impact?.hasDescendants) return 'Delete';
  const count = impact.affectedCapabilities.length;
  return `Delete ${capability.name} and ${count} ${plural(count, 'child', 'children')}`;
};

export const DeleteCapabilityDialog: React.FC<DeleteCapabilityDialogProps> = ({
  isOpen,
  onClose,
  capability,
  onConfirm,
  capabilitiesToDelete = [],
  domainId,
}) => {
  const [isDeleting, setIsDeleting] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const deleteCapabilityMutation = useDeleteCapability();
  const cascadeDeleteMutation = useCascadeDeleteCapability();

  const capabilityId = capability?.id as CapabilityId | undefined;
  const { data: impact, isLoading: impactLoading } = useDeleteImpact(isOpen ? capabilityId : undefined);

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const deleteSelected = async (target: Capability) => {
    if (capabilitiesToDelete.length > 1) {
      for (const cap of capabilitiesToDelete) {
        await deleteCapabilityMutation.mutateAsync({ capability: cap, parentId: cap.parentId ?? undefined, domainId });
      }
      return;
    }
    await cascadeDeleteMutation.mutateAsync({
      capability: target,
      cascade: impact?.hasDescendants ?? false,
      parentId: target.parentId ?? undefined,
      domainId,
    });
  };

  const handleConfirm = async () => {
    if (!capability) return;

    setIsDeleting(true);
    setBackendError(null);

    try {
      await deleteSelected(capability);
      onConfirm?.();
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to delete capability');
    } finally {
      setIsDeleting(false);
    }
  };

  if (!capability) return null;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Delete Capability?"
      centered
      size={impact?.hasDescendants ? 'lg' : 'md'}
      data-testid="delete-capability-dialog"
    >
      <Stack gap="md">
        <ImpactBody
          capability={capability}
          impact={impact}
          impactLoading={impactLoading}
          multiDeleteCount={capabilitiesToDelete.length}
        />

        <Text c="orange" size="sm">
          This action cannot be undone.
        </Text>

        {backendError && (
          <Alert color="red" data-testid="delete-capability-error">
            {backendError}
          </Alert>
        )}

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={handleClose} disabled={isDeleting} data-testid="delete-capability-cancel">
            Cancel
          </Button>
          <Button
            color="red"
            onClick={handleConfirm}
            loading={isDeleting}
            disabled={impactLoading}
            data-testid="delete-capability-submit"
          >
            {submitLabel(capability, impact)}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
