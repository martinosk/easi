import React, { useCallback } from 'react';
import { useDialog } from '../../../../contexts/dialogs';
import { useCreateConnectedEntity } from '../../hooks/useCreateConnectedEntity';
import {
  CreateConnectedEntityDialog,
  type CreateConnectedEntitySubmitData,
} from './CreateConnectedEntityDialog';

export const CreateConnectedEntityDialogContainer: React.FC = () => {
  const dialog = useDialog('create-connected-entity');

  const sourceNodeId = dialog.data?.sourceNodeId ?? '';
  const handlePosition = dialog.data?.handlePosition ?? 'right';
  const sourcePosition = dialog.data?.sourcePosition ?? { x: 0, y: 0 };

  const { createConnectedEntity } = useCreateConnectedEntity(
    sourceNodeId,
    sourcePosition,
    handlePosition,
  );

  const handleSubmit = useCallback(
    (data: CreateConnectedEntitySubmitData) => {
      void createConnectedEntity(data).then(() => {
        dialog.close();
      });
    },
    [createConnectedEntity, dialog],
  );

  if (!dialog.isOpen || !dialog.data) return null;

  return (
    <CreateConnectedEntityDialog
      isOpen={dialog.isOpen}
      sourceNodeId={dialog.data.sourceNodeId}
      sourceNodeType={dialog.data.sourceNodeType}
      handlePosition={dialog.data.handlePosition}
      links={dialog.data.links}
      onSubmit={handleSubmit}
      onClose={dialog.close}
    />
  );
};
