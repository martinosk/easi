import { Alert, Badge, Button, Group, Modal, ScrollArea, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, CapabilityDeleteImpact } from '../../../api/types';
import {
  type CapabilityDeletionPlan,
  useCapabilityDeletionPlan,
  useExecuteCapabilityDeletion,
} from '../hooks/useCapabilityDeletionPlan';
import { ImpactError, ImpactLoading, plural, RealizationsToRemove } from './CapabilityDeletionImpact';

interface DeleteCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
  onConfirm?: () => void;
  capabilitiesToDelete?: Capability[];
  domainId?: string;
}

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

interface PlanProps {
  targets: Capability[];
  plan: CapabilityDeletionPlan;
}

const Summary = ({ targets, plan }: PlanProps) => {
  const [capability] = targets;
  const impact = plan.impacts[0];
  if (targets.length > 1) return <Text>Are you sure you want to delete {targets.length} capabilities?</Text>;
  if (impact?.hasDescendants) return <CascadeImpact capability={capability} impact={impact} />;
  return <LeafConfirmation capability={capability} />;
};

const ImpactBody = ({ targets, plan }: PlanProps) => {
  if (plan.isLoading) return <ImpactLoading />;
  return (
    <>
      <Summary targets={targets} plan={plan} />
      <RealizationsToRemove realizations={plan.realizations} capabilityNames={plan.capabilityNames} />
      {plan.isError && <ImpactError onRetry={plan.retry} />}
    </>
  );
};

const submitLabel = ({ targets, plan }: PlanProps) => {
  const impact = plan.impacts[0];
  if (targets.length > 1 || !impact?.hasDescendants) return 'Delete';
  const count = impact.affectedCapabilities.length;
  return `Delete ${targets[0].name} and ${count} ${plural(count, 'child', 'children')}`;
};

const deleteTargets = (
  isOpen: boolean,
  capability: Capability | null,
  capabilitiesToDelete: Capability[],
): Capability[] => {
  if (!isOpen) return [];
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

  const targets = deleteTargets(isOpen, capability, capabilitiesToDelete);
  const plan = useCapabilityDeletionPlan(targets);
  const executeDeletion = useExecuteCapabilityDeletion();

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const handleConfirm = async () => {
    setIsDeleting(true);
    setBackendError(null);

    try {
      for (const deletion of plan.deletions) {
        await executeDeletion(deletion, { domainId });
      }
      onConfirm?.();
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to delete capability');
      plan.retry();
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
      size={plan.impacts[0]?.hasDescendants ? 'lg' : 'md'}
      data-testid="delete-capability-dialog"
    >
      <Stack gap="md">
        <ImpactBody targets={targets} plan={plan} />

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
            disabled={!plan.ready}
            data-testid="delete-capability-submit"
          >
            {submitLabel({ targets, plan })}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
