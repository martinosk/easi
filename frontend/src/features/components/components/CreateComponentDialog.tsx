import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';

interface CreateComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

export const CreateComponentDialog: React.FC<CreateComponentDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const createComponent = useAppStore((state) => state.createComponent);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleClose = () => {
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Application name is required');
      return;
    }

    setIsCreating(true);

    try {
      await createComponent({
        name: name.trim(),
        description: description.trim() || undefined,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create application');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleClose} data-testid="create-component-dialog">
      <div className="dialog-content">
        <h2 className="dialog-title">Create Application</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="component-name" className="form-label">
              Name <span className="required">*</span>
            </label>
            <input
              id="component-name"
              type="text"
              className="form-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter application name"
              autoFocus
              disabled={isCreating}
              data-testid="component-name-input"
            />
          </div>

          <div className="form-group">
            <label htmlFor="component-description" className="form-label">
              Description
            </label>
            <textarea
              id="component-description"
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter application description (optional)"
              rows={3}
              disabled={isCreating}
              data-testid="component-description-input"
            />
          </div>

          {error && <div className="error-message" data-testid="create-component-error">{error}</div>}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-component-cancel"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isCreating || !name.trim()}
              data-testid="create-component-submit"
            >
              {isCreating ? 'Creating...' : 'Create Application'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
