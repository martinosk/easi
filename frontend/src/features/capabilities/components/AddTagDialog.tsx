import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';

interface AddTagDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

interface FormState {
  tag: string;
}

interface FormErrors {
  tag?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.tag.trim()) {
    errors.tag = 'Tag name is required';
  }

  return errors;
};

export const AddTagDialog: React.FC<AddTagDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [form, setForm] = useState<FormState>({
    tag: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isAdding, setIsAdding] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const addCapabilityTag = useAppStore((state) => state.addCapabilityTag);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const resetForm = () => {
    setForm({
      tag: '',
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (value: string) => {
    setForm({ tag: value });
    if (errors.tag) {
      setErrors({});
    }
    if (backendError) {
      setBackendError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBackendError(null);

    const validationErrors = validateForm(form);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setIsAdding(true);

    try {
      await addCapabilityTag(capabilityId as import('../../../api/types').CapabilityId, form.tag.trim());

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add tag');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onClose={handleClose}
      data-testid="add-tag-dialog"
    >
      <div className="dialog-content">
        <h2 className="dialog-title">Add Tag</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="tag-name" className="form-label">
              Tag Name <span className="required">*</span>
            </label>
            <input
              id="tag-name"
              type="text"
              className={`form-input ${errors.tag ? 'form-input-error' : ''}`}
              value={form.tag}
              onChange={(e) => handleFieldChange(e.target.value)}
              placeholder="Enter tag name"
              autoFocus
              disabled={isAdding}
              data-testid="tag-name-input"
            />
            {errors.tag && (
              <div className="field-error" data-testid="tag-name-error">
                {errors.tag}
              </div>
            )}
          </div>

          {backendError && (
            <div className="error-message" data-testid="add-tag-error">
              {backendError}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isAdding}
              data-testid="add-tag-cancel"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isAdding || !form.tag.trim()}
              data-testid="add-tag-submit"
            >
              {isAdding ? 'Adding...' : 'Add'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
