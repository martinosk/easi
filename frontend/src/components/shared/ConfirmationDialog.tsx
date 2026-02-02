import React, { useEffect } from 'react';

interface ConfirmationDialogProps {
  title: string;
  message: string;
  itemName?: string;
  itemNames?: string[];
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
  error?: string | null;
}

function useDialogKeyboard(onCancel: () => void, onConfirm: () => void, isLoading: boolean) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
      if (e.key === 'Enter' && !isLoading) onConfirm();
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onCancel, onConfirm, isLoading]);
}

const ItemNamesList: React.FC<{ names: string[] }> = ({ names }) => (
  <ul className="dialog-item-list" style={{ maxHeight: '150px', overflowY: 'auto', margin: '8px 0', paddingLeft: '20px' }}>
    {names.map((name, index) => (
      <li key={index}>{name}</li>
    ))}
  </ul>
);

export const ConfirmationDialog: React.FC<ConfirmationDialogProps> = ({
  title,
  message,
  itemName,
  itemNames,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  isLoading = false,
  error = null,
}) => {
  useDialogKeyboard(onCancel, onConfirm, isLoading);

  return (
    <div className="dialog-overlay" onClick={onCancel}>
      <div className="dialog" onClick={(e) => e.stopPropagation()} role="alertdialog" aria-labelledby="dialog-title" aria-describedby="dialog-description">
        <h3 id="dialog-title">{title}</h3>
        <p id="dialog-description">{message}</p>
        {itemName && <p className="dialog-item-name">"{itemName}"</p>}
        {itemNames && itemNames.length > 0 && <ItemNamesList names={itemNames} />}
        {error && <p className="dialog-error">{error}</p>}
        {!error && <p className="dialog-warning">This action cannot be undone.</p>}
        <div className="dialog-actions">
          <button
            onClick={onCancel}
            className="btn-secondary"
            disabled={isLoading}
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className="btn-danger"
            disabled={isLoading}
          >
            {isLoading ? 'Deleting...' : confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};
