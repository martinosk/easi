import React from 'react';

interface CreateViewDialogProps {
  isOpen: boolean;
  viewName: string;
  onViewNameChange: (name: string) => void;
  onClose: () => void;
  onCreate: () => void;
}

export const CreateViewDialog: React.FC<CreateViewDialogProps> = ({
  isOpen,
  viewName,
  onViewNameChange,
  onClose,
  onCreate,
}) => {
  if (!isOpen) return null;

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') onCreate();
    if (e.key === 'Escape') onClose();
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <h3>Create New View</h3>
        <input
          type="text"
          placeholder="View name"
          value={viewName}
          onChange={(e) => onViewNameChange(e.target.value)}
          onKeyDown={handleKeyDown}
          autoFocus
          className="dialog-input"
        />
        <div className="dialog-actions">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={onCreate} className="btn-primary">Create</button>
        </div>
      </div>
    </div>
  );
};
