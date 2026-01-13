import React, { useState } from 'react';
import { Modal, Button, Group, Stack, Alert } from '@mantine/core';
import type { Capability } from '../../../api/types';
import { useEditCapabilityForm } from '../hooks/useEditCapabilityForm';
import {
  BasicFields,
  StatusField,
  MaturityField,
  OwnershipFields,
  ExpertsList,
  TagsList,
} from './EditCapabilityFormFields';
import { AddExpertDialog } from './AddExpertDialog';
import { AddTagDialog } from './AddTagDialog';

interface EditCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
}

export const EditCapabilityDialog: React.FC<EditCapabilityDialogProps> = ({
  isOpen,
  onClose,
  capability,
}) => {
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const [isAddTagOpen, setIsAddTagOpen] = useState(false);

  const {
    form,
    currentCapability,
    statusOptions,
    ownershipOptions,
    userOptions,
    isSaving,
    isLoadingMetadata,
    backendError,
    handleSubmit,
    clearError,
  } = useEditCapabilityForm(capability, isOpen, onClose);

  const handleClose = () => {
    clearError();
    onClose();
  };

  const {
    register,
    handleSubmit: formHandleSubmit,
    control,
    formState: { errors, isValid },
  } = form;

  if (!capability) return null;

  const displayCapability = currentCapability || capability;

  return (
    <>
      <Modal
        opened={isOpen}
        onClose={handleClose}
        title="Edit Capability"
        centered
        size="lg"
        data-testid="edit-capability-dialog"
      >
        <form onSubmit={formHandleSubmit(handleSubmit)}>
          <Stack gap="md">
            <BasicFields register={register} errors={errors} disabled={isSaving} />

            <StatusField
              control={control}
              options={statusOptions}
              isLoading={isLoadingMetadata}
              disabled={isSaving}
            />

            <MaturityField control={control} disabled={isSaving} />

            <OwnershipFields
              control={control}
              register={register}
              ownershipOptions={ownershipOptions}
              userOptions={userOptions}
              isLoadingOwnership={isLoadingMetadata}
              isLoadingUsers={isLoadingMetadata}
              disabled={isSaving}
            />

            <ExpertsList
              experts={displayCapability.experts}
              onAddClick={() => setIsAddExpertOpen(true)}
              disabled={isSaving}
            />

            <TagsList
              tags={displayCapability.tags}
              onAddClick={() => setIsAddTagOpen(true)}
              disabled={isSaving}
            />

            {backendError && (
              <Alert color="red" data-testid="edit-capability-error">
                {backendError}
              </Alert>
            )}

            <Group justify="flex-end" gap="sm">
              <Button
                variant="default"
                onClick={handleClose}
                disabled={isSaving}
                data-testid="edit-capability-cancel"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                loading={isSaving}
                disabled={isLoadingMetadata || !isValid}
                data-testid="edit-capability-submit"
              >
                Save
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <AddExpertDialog
        isOpen={isAddExpertOpen}
        onClose={() => setIsAddExpertOpen(false)}
        capabilityId={capability.id}
      />

      <AddTagDialog
        isOpen={isAddTagOpen}
        onClose={() => setIsAddTagOpen(false)}
        capabilityId={capability.id}
      />
    </>
  );
};
