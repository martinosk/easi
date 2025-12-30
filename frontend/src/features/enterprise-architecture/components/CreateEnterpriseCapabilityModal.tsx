import React, { useState, useEffect, useRef, type FormEvent } from 'react';
import type { CreateEnterpriseCapabilityRequest } from '../types';

interface CreateEnterpriseCapabilityModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateEnterpriseCapabilityRequest) => Promise<void>;
}

export const CreateEnterpriseCapabilityModal = React.memo<CreateEnterpriseCapabilityModalProps>(({ isOpen, onClose, onSubmit }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await onSubmit({
        name,
        description: description || undefined,
        category: category || undefined,
      });
      resetForm();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create enterprise capability');
    } finally {
      setIsSubmitting(false);
    }
  };

  const resetForm = () => {
    setName('');
    setDescription('');
    setCategory('');
    setError(null);
  };

  const handleCancel = () => {
    resetForm();
    onClose();
  };

  return (
    <dialog ref={dialogRef} className="dialog" onClose={handleCancel} data-testid="create-capability-modal">
      <div className="dialog-content">
        <h2 className="dialog-title">Create Enterprise Capability</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="capability-name">
              Name <span className="required">*</span>
            </label>
            <input
              id="capability-name"
              type="text"
              className="form-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={200}
              disabled={isSubmitting}
              placeholder="e.g., Customer Management"
              data-testid="capability-name-input"
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="capability-category">
              Category
            </label>
            <input
              id="capability-category"
              type="text"
              className="form-input"
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              maxLength={100}
              disabled={isSubmitting}
              placeholder="e.g., Core Business"
              data-testid="capability-category-input"
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="capability-description">
              Description
            </label>
            <textarea
              id="capability-description"
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={1000}
              disabled={isSubmitting}
              placeholder="Describe what this enterprise capability represents..."
              rows={3}
              data-testid="capability-description-input"
            />
          </div>

          {error && (
            <div className="error-message" data-testid="create-error-message">
              {error}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleCancel}
              disabled={isSubmitting}
              data-testid="create-cancel-btn"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting || !name.trim()}
              data-testid="create-submit-btn"
            >
              {isSubmitting ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
});

CreateEnterpriseCapabilityModal.displayName = 'CreateEnterpriseCapabilityModal';
