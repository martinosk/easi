import React, { useState } from 'react';
import { Modal, Text, Button, Group, Stack, Alert, Checkbox, Skeleton, ScrollArea, Badge } from '@mantine/core';
import { useDeleteCapability, useDeleteImpact, useCascadeDeleteCapability } from '../hooks/useCapabilities';
import type { Capability, CapabilityId } from '../../../api/types';

interface DeleteCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
  onConfirm?: () => void;
  capabilitiesToDelete?: Capability[];
  domainId?: string;
}

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
  const [deleteApplications, setDeleteApplications] = useState(false);

  const deleteCapabilityMutation = useDeleteCapability();
  const cascadeDeleteMutation = useCascadeDeleteCapability();

  const capabilityId = capability?.id as CapabilityId | undefined;
  const { data: impact, isLoading: impactLoading } = useDeleteImpact(
    isOpen ? capabilityId : undefined
  );

  const handleClose = () => {
    setBackendError(null);
    setDeleteApplications(false);
    onClose();
  };

  const handleConfirm = async () => {
    if (!capability) return;

    setIsDeleting(true);
    setBackendError(null);

    try {
      const isMultiSelect = capabilitiesToDelete.length > 1;

      if (isMultiSelect) {
        for (const cap of capabilitiesToDelete) {
          await deleteCapabilityMutation.mutateAsync({
            capability: cap,
            parentId: cap.parentId ?? undefined,
            domainId,
          });
        }
      } else if (impact?.hasDescendants) {
        await cascadeDeleteMutation.mutateAsync({
          capability,
          cascade: true,
          deleteRealisingApplications: deleteApplications,
          parentId: capability.parentId ?? undefined,
          domainId,
        });
      } else {
        await cascadeDeleteMutation.mutateAsync({
          capability,
          cascade: false,
          deleteRealisingApplications: false,
          parentId: capability.parentId ?? undefined,
          domainId,
        });
      }
      onConfirm?.();
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to delete capability');
    } finally {
      setIsDeleting(false);
    }
  };

  if (!capability) return null;

  const isMultiDelete = capabilitiesToDelete.length > 1;
  const hasCascade = impact?.hasDescendants ?? false;
  const deletableRealizations = impact?.realizationsOnDeletedCapabilities ?? [];
  const retainedRealizations = impact?.realizationsOnRetainedCapabilities ?? [];
  const affectedCount = impact?.affectedCapabilities?.length ?? 0;

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title="Delete Capability?"
      centered
      size={hasCascade ? 'lg' : 'md'}
      data-testid="delete-capability-dialog"
    >
      <Stack gap="md">
        {impactLoading ? (
          <Stack gap="sm" data-testid="delete-impact-loading">
            <Skeleton height={20} width="80%" />
            <Skeleton height={20} width="60%" />
            <Skeleton height={40} />
          </Stack>
        ) : isMultiDelete ? (
          <Text>Are you sure you want to delete {capabilitiesToDelete.length} capabilities?</Text>
        ) : hasCascade ? (
          <>
            <Alert color="orange" data-testid="cascade-warning">
              This will delete <Text span fw={600}>"{capability.name}"</Text> and{' '}
              <Text span fw={600}>{affectedCount} child {affectedCount === 1 ? 'capability' : 'capabilities'}</Text>.
            </Alert>

            {affectedCount > 0 && (
              <Stack gap="xs">
                <Text fw={600} size="sm">Affected Capabilities ({affectedCount})</Text>
                <ScrollArea.Autosize mah={150}>
                  <Stack gap={4}>
                    {impact!.affectedCapabilities.map((cap) => (
                      <Group key={cap.id} gap="xs">
                        <Badge size="xs" variant="light">{cap.level}</Badge>
                        <Text size="sm">{cap.name}</Text>
                      </Group>
                    ))}
                  </Stack>
                </ScrollArea.Autosize>
              </Stack>
            )}

            {deletableRealizations.length > 0 && (
              <Checkbox
                label={`Also delete ${deletableRealizations.length} application ${deletableRealizations.length === 1 ? 'realization' : 'realizations'} that only realise these capabilities`}
                checked={deleteApplications}
                onChange={(e) => setDeleteApplications(e.currentTarget.checked)}
                data-testid="delete-applications-checkbox"
              />
            )}

            {retainedRealizations.length > 0 && (
              <Text size="sm" c="dimmed">
                {retainedRealizations.length} {retainedRealizations.length === 1 ? 'realization' : 'realizations'} will be retained (applications also realise other capabilities).
              </Text>
            )}
          </>
        ) : (
          <>
            <Text>Are you sure you want to delete</Text>
            <Text fw={600} size="lg">"{capability.name}"</Text>
          </>
        )}

        <Text c="orange" size="sm">This action cannot be undone.</Text>

        {backendError && (
          <Alert color="red" data-testid="delete-capability-error">
            {backendError}
          </Alert>
        )}

        <Group justify="flex-end" gap="sm">
          <Button
            variant="default"
            onClick={handleClose}
            disabled={isDeleting}
            data-testid="delete-capability-cancel"
          >
            Cancel
          </Button>
          <Button
            color="red"
            onClick={handleConfirm}
            loading={isDeleting}
            disabled={impactLoading}
            data-testid="delete-capability-submit"
          >
            {hasCascade
              ? `Delete ${capability.name} and ${affectedCount} ${affectedCount === 1 ? 'child' : 'children'}`
              : 'Delete'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
