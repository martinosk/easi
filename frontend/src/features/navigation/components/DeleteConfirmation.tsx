import React from 'react';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import type { View, Component } from '../../../api/types';

export type DeleteTarget = { type: 'view'; view: View } | { type: 'component'; component: Component };

interface DeleteConfirmationProps {
  deleteTarget: DeleteTarget | null;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
}

export const DeleteConfirmation: React.FC<DeleteConfirmationProps> = ({
  deleteTarget,
  onConfirm,
  onCancel,
  isLoading,
}) => {
  if (!deleteTarget) return null;

  const isView = deleteTarget.type === 'view';
  const title = isView ? 'Delete View' : 'Delete Application';
  const message = isView
    ? 'Are you sure you want to delete this view?'
    : 'This will delete the application from the entire model, remove it from ALL views, and delete ALL relations involving this application.';
  const itemName = isView ? deleteTarget.view!.name : deleteTarget.component!.name;

  return (
    <ConfirmationDialog
      title={title}
      message={message}
      itemName={itemName}
      confirmText="Delete"
      cancelText="Cancel"
      onConfirm={onConfirm}
      onCancel={onCancel}
      isLoading={isLoading}
    />
  );
};
