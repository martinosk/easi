import { ConfirmationDialog } from '../../../../components/shared/ConfirmationDialog';
import type { BulkOperationRequest } from '../context-menus/MultiSelectContextMenu';
import type { BulkOperationResult } from '../../hooks/useBulkOperations';

interface BulkConfirmationDialogProps {
  bulkOperation: BulkOperationRequest | null;
  isExecuting: boolean;
  result: BulkOperationResult | null;
  onConfirm: () => void;
  onCancel: () => void;
}

function buildResultError(result: BulkOperationResult): string {
  const failedNames = result.failed.map((f) => f.name).join(', ');
  return `${result.succeeded.length} item(s) succeeded, ${result.failed.length} failed: ${failedNames}`;
}

export const BulkConfirmationDialog = ({
  bulkOperation,
  isExecuting,
  result,
  onConfirm,
  onCancel,
}: BulkConfirmationDialogProps) => {
  if (!bulkOperation) return null;

  const count = bulkOperation.nodes.length;
  const error = result ? buildResultError(result) : null;

  if (bulkOperation.type === 'removeFromView') {
    return (
      <ConfirmationDialog
        title={`Remove ${count} items from View`}
        message={`Remove ${count} items from the current view? The items will remain in the model.`}
        confirmText="Remove"
        cancelText="Cancel"
        onConfirm={onConfirm}
        onCancel={onCancel}
        isLoading={isExecuting}
        error={error}
      />
    );
  }

  if (bulkOperation.type === 'deleteFromModel') {
    const itemNames = bulkOperation.nodes.map((n) => n.nodeName);
    return (
      <ConfirmationDialog
        title={`Delete ${count} items from Model`}
        message={`This will permanently delete ${count} items from the entire model. They will be removed from ALL views and all associated relations will be deleted. This cannot be undone.`}
        itemNames={itemNames}
        confirmText={`Delete ${count} items`}
        cancelText="Cancel"
        onConfirm={onConfirm}
        onCancel={onCancel}
        isLoading={isExecuting}
        error={error}
      />
    );
  }

  return null;
};
