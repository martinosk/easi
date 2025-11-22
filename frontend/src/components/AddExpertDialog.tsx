import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../store/appStore';

interface AddExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capabilityId: string;
}

interface FormState {
  name: string;
  role: string;
  contact: string;
}

interface FormErrors {
  name?: string;
  role?: string;
  contact?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = 'Name is required';
  }

  if (!form.role.trim()) {
    errors.role = 'Role is required';
  }

  if (!form.contact.trim()) {
    errors.contact = 'Contact is required';
  }

  return errors;
};

export const AddExpertDialog: React.FC<AddExpertDialogProps> = ({
  isOpen,
  onClose,
  capabilityId,
}) => {
  const [form, setForm] = useState<FormState>({
    name: '',
    role: '',
    contact: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isAdding, setIsAdding] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const addCapabilityExpert = useAppStore((state) => state.addCapabilityExpert);

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
      name: '',
      role: '',
      contact: '',
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (field: keyof FormState, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
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
      await addCapabilityExpert(capabilityId, {
        expertName: form.name.trim(),
        expertRole: form.role.trim(),
        contactInfo: form.contact.trim(),
      });

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add expert');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onClose={handleClose}
      data-testid="add-expert-dialog"
    >
      <div className="dialog-content">
        <h2 className="dialog-title">Add Expert</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="expert-name" className="form-label">
              Name <span className="required">*</span>
            </label>
            <input
              id="expert-name"
              type="text"
              className={`form-input ${errors.name ? 'form-input-error' : ''}`}
              value={form.name}
              onChange={(e) => handleFieldChange('name', e.target.value)}
              placeholder="Enter expert name"
              autoFocus
              disabled={isAdding}
              data-testid="expert-name-input"
            />
            {errors.name && (
              <div className="field-error" data-testid="expert-name-error">
                {errors.name}
              </div>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="expert-role" className="form-label">
              Role <span className="required">*</span>
            </label>
            <input
              id="expert-role"
              type="text"
              className={`form-input ${errors.role ? 'form-input-error' : ''}`}
              value={form.role}
              onChange={(e) => handleFieldChange('role', e.target.value)}
              placeholder="Enter expert role"
              disabled={isAdding}
              data-testid="expert-role-input"
            />
            {errors.role && (
              <div className="field-error" data-testid="expert-role-error">
                {errors.role}
              </div>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="expert-contact" className="form-label">
              Contact <span className="required">*</span>
            </label>
            <input
              id="expert-contact"
              type="text"
              className={`form-input ${errors.contact ? 'form-input-error' : ''}`}
              value={form.contact}
              onChange={(e) => handleFieldChange('contact', e.target.value)}
              placeholder="Enter contact information"
              disabled={isAdding}
              data-testid="expert-contact-input"
            />
            {errors.contact && (
              <div className="field-error" data-testid="expert-contact-error">
                {errors.contact}
              </div>
            )}
          </div>

          {backendError && (
            <div className="error-message" data-testid="add-expert-error">
              {backendError}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isAdding}
              data-testid="add-expert-cancel"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isAdding || !form.name.trim() || !form.role.trim() || !form.contact.trim()}
              data-testid="add-expert-submit"
            >
              {isAdding ? 'Adding...' : 'Add'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
