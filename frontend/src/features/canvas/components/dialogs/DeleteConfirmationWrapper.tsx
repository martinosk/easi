import { ConfirmationDialog } from '../../../../components/shared/ConfirmationDialog';
import type { DeleteTarget } from '../../hooks/useDeleteConfirmation';

interface DeleteConfirmationWrapperProps {
  deleteTarget: DeleteTarget | null;
  isDeleting: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export const DeleteConfirmationWrapper = ({
  deleteTarget,
  isDeleting,
  onConfirm,
  onCancel,
}: DeleteConfirmationWrapperProps) => {
  if (!deleteTarget) return null;

  const getTitle = () => {
    switch (deleteTarget.type) {
      case 'component-from-model':
        return 'Delete Component from Model';
      case 'capability-from-model':
        return 'Delete Capability from Model';
      case 'parent-relation':
        return 'Remove Parent Relationship';
      case 'realization':
        return 'Delete Realization';
      case 'origin-entity-from-model':
        return 'Delete Origin Entity from Model';
      case 'origin-relationship':
        return 'Delete Relationship';
      default:
        return 'Delete Relation from Model';
    }
  };

  const getMessage = () => {
    switch (deleteTarget.type) {
      case 'component-from-model':
        return 'This will delete the component from the entire model, remove it from ALL views, and delete ALL relations involving this component.';
      case 'capability-from-model':
        return 'This will delete the capability from the entire model, remove it from ALL views, and affect any child capabilities.';
      case 'parent-relation':
        return 'This will remove the parent-child relationship. The child capability will become a top-level (L1) capability.';
      case 'realization':
        return 'This will remove the link between this capability and application. Any inherited realizations will also be removed.';
      case 'origin-entity-from-model':
        return 'This will delete the origin entity from the entire model and remove all relationships to applications.';
      case 'origin-relationship':
        return 'This will unlink this origin entity from the application.';
      default:
        return 'This will delete the relation from the entire model and remove it from ALL views.';
    }
  };

  return (
    <ConfirmationDialog
      title={getTitle()}
      message={getMessage()}
      itemName={deleteTarget.name}
      confirmText={['parent-relation', 'origin-relationship'].includes(deleteTarget.type) ? 'Remove' : 'Delete'}
      cancelText="Cancel"
      onConfirm={onConfirm}
      onCancel={onCancel}
      isLoading={isDeleting}
    />
  );
};
