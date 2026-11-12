import { Alert, Badge, Button, Group, Modal, ScrollArea, Skeleton, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, CapabilityDeleteImpact, CapabilityId, ImpactRealization } from '../../../api/types';
import { useCascadeDeleteCapability, useDeleteCapability, useDeleteImpacts } from '../hooks/useCapabilities';

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

const ImpactError = ({ onRetry }: { onRetry: () => void }) => (
  <Alert color="red" data-testid="delete-impact-error">
    <Group justify="space-between">
      <Text size="sm">Could not load what this delete will remove.</Text>
      <Button size="xs" variant="light" color="red" onClick={onRetry} data-testid="delete-impact-retry">
        Retry
      </Button>
    </Group>
  </Alert>
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

interface ImpactProps {
  targets: Capability[];
  impacts: CapabilityDeleteImpact[];
}

const capabilityNames = ({ targets, impacts }: ImpactProps): Map<string, string> =>
  new Map([
    ...targets.map((c) => [c.id, c.name] as const),
    ...impacts.flatMap((impact) => impact.affectedCapabilities.map((c) => [c.id, c.name] as const)),
  ]);

const RealizationsToRemove = ({ targets, impacts }: ImpactProps) => {
  const realizations: ImpactRealization[] = impacts.flatMap((impact) => impact.realizations ?? []);
  if (realizations.length === 0) return null;

  const names = capabilityNames({ targets, impacts });
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

const Summary = ({ targets, impacts }: ImpactProps) => {
  const [capability] = targets;
  const impact = impacts[0];
  if (targets.length > 1) return <Text>Are you sure you want to delete {targets.length} capabilities?</Text>;
  if (impact?.hasDescendants) return <CascadeImpact capability={capability} impact={impact} />;
  return <LeafConfirmation capability={capability} />;
};

const submitLabel = ({ targets, impacts }: ImpactProps) => {
  const impact = impacts[0];
  if (targets.length > 1 || !impact?.hasDescendants) return 'Delete';
  const count = impact.affectedCapabilities.length;
  return `Delete ${targets[0].name} and ${count} ${plural(count, 'child', 'children')}`;
};

const deleteTargets = (capability: Capability | null, capabilitiesToDelete: Capability[]): Capability[] => {
  if (capabilitiesToDelete.length > 1) return capabilitiesToDelete;
  return capability ? [capability] : [];
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

  const targets = deleteTargets(capability, capabilitiesToDelete);
  const { impacts, isLoading, isError, retry } = useDeleteImpacts(
    isOpen ? targets.map((t) => t.id as CapabilityId) : [],
  );
  const impact = { targets, impacts };
  const hasDescendants = impacts[0]?.hasDescendants ?? false;

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const deleteSelected = async () => {
    if (targets.length > 1) {
      for (const cap of targets) {
        await deleteCapabilityMutation.mutateAsync({ capability: cap, parentId: cap.parentId ?? undefined, domainId });
      }
      return;
    }
    const [target] = targets;
    await cascadeDeleteMutation.mutateAsync({
      capability: target,
      cascade: hasDescendants,
      parentId: target.parentId ?? undefined,
      domainId,
    });
  };

  const handleConfirm = async () => {
    setIsDeleting(true);
    setBackendError(null);

    try {
      await deleteSelected();
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
      size={hasDescendants ? 'lg' : 'md'}
      data-testid="delete-capability-dialog"
    >
      <Stack gap="md">
        {isLoading ? (
          <ImpactLoading />
        ) : (
          <>
            <Summary {...impact} />
            <RealizationsToRemove {...impact} />
          </>
        )}

        {isError && <ImpactError onRetry={retry} />}

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
            disabled={impacts.length !== targets.length}
            data-testid="delete-capability-submit"
          >
            {submitLabel(impact)}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
