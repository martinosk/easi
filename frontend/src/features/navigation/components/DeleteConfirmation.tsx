import React from 'react';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import type { View, Component, AcquiredEntity, Vendor, InternalTeam } from '../../../api/types';

export type DeleteTarget =
  | { type: 'view'; view: View }
  | { type: 'component'; component: Component }
  | { type: 'acquired'; entity: AcquiredEntity }
  | { type: 'vendor'; entity: Vendor }
  | { type: 'team'; entity: InternalTeam };

interface DeleteConfirmationProps {
  deleteTarget: DeleteTarget | null;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
}

function getDeleteInfo(target: DeleteTarget): { title: string; message: string; itemName: string } {
  switch (target.type) {
    case 'view':
      return {
        title: 'Delete View',
        message: 'Are you sure you want to delete this view?',
        itemName: target.view.name,
      };
    case 'component':
      return {
        title: 'Delete Application',
        message: 'This will delete the application from the entire model, remove it from ALL views, and delete ALL relations involving this application.',
        itemName: target.component.name,
      };
    case 'acquired':
      return {
        title: 'Delete Acquired Entity',
        message: 'This will delete the acquired entity and all relationships to applications.',
        itemName: target.entity.name,
      };
    case 'vendor':
      return {
        title: 'Delete Vendor',
        message: 'This will delete the vendor and all relationships to applications.',
        itemName: target.entity.name,
      };
    case 'team':
      return {
        title: 'Delete Internal Team',
        message: 'This will delete the internal team and all relationships to applications.',
        itemName: target.entity.name,
      };
  }
}

export const DeleteConfirmation: React.FC<DeleteConfirmationProps> = ({
  deleteTarget,
  onConfirm,
  onCancel,
  isLoading,
}) => {
  if (!deleteTarget) return null;

  const { title, message, itemName } = getDeleteInfo(deleteTarget);

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
