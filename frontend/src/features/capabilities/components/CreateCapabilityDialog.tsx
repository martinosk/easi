import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';
import { apiClient } from '../../../api/client';
import type { StatusOption } from '../../../api/types';

interface CreateCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_MATURITY_LEVELS = ['Genesis', 'Custom Build', 'Product', 'Commodity'];
const DEFAULT_STATUSES: StatusOption[] = [
  { value: 'Active', displayName: 'Active', sortOrder: 1 },
  { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
  { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
];

interface FormState {
  name: string;
  description: string;
  status: string;
  maturityLevel: string;
}

interface FormErrors {
  name?: string;
  description?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = 'Name is required';
  } else if (form.name.trim().length > 200) {
    errors.name = 'Name must be 200 characters or less';
  }

  if (form.description.length > 1000) {
    errors.description = 'Description must be 1000 characters or less';
  }

  return errors;
};

export const CreateCapabilityDialog: React.FC<CreateCapabilityDialogProps> = ({
  isOpen,
  onClose,
}) => {
  const [form, setForm] = useState<FormState>({
    name: '',
    description: '',
    status: 'Active',
    maturityLevel: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isCreating, setIsCreating] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);
  const [maturityLevels, setMaturityLevels] = useState<string[]>([]);
  const [isLoadingMaturityLevels, setIsLoadingMaturityLevels] = useState(false);
  const [statuses, setStatuses] = useState<StatusOption[]>([]);
  const [isLoadingStatuses, setIsLoadingStatuses] = useState(false);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const createCapability = useAppStore((state) => state.createCapability);
  const updateCapabilityMetadata = useAppStore((state) => state.updateCapabilityMetadata);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
      fetchMetadata();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const fetchMetadata = async () => {
    setIsLoadingMaturityLevels(true);
    setIsLoadingStatuses(true);

    try {
      const levels = await apiClient.getMaturityLevels();
      setMaturityLevels(levels);
      if (levels.length > 0 && !form.maturityLevel) {
        setForm((prev) => ({ ...prev, maturityLevel: levels[0] }));
      }
    } catch {
      setMaturityLevels(DEFAULT_MATURITY_LEVELS);
      if (!form.maturityLevel) {
        setForm((prev) => ({ ...prev, maturityLevel: DEFAULT_MATURITY_LEVELS[0] }));
      }
    } finally {
      setIsLoadingMaturityLevels(false);
    }

    try {
      const statusList = await apiClient.getStatuses();
      setStatuses(statusList.sort((a, b) => a.sortOrder - b.sortOrder));
    } catch {
      setStatuses(DEFAULT_STATUSES);
    } finally {
      setIsLoadingStatuses(false);
    }
  };

  const resetForm = () => {
    const defaultMaturity = maturityLevels.length > 0 ? maturityLevels[0] : DEFAULT_MATURITY_LEVELS[0];
    setForm({
      name: '',
      description: '',
      status: 'Active',
      maturityLevel: defaultMaturity,
    });
    setErrors({});
    setBackendError(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFieldChange = (
    field: keyof FormState,
    value: string
  ) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field as keyof FormErrors]) {
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

    setIsCreating(true);

    try {
      const capability = await createCapability({
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        level: 'L1',
      });

      await updateCapabilityMetadata(capability.id, {
        status: form.status,
        maturityLevel: form.maturityLevel,
      });

      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to create capability');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <dialog
      ref={dialogRef}
      className="dialog"
      onClose={handleClose}
      data-testid="create-capability-dialog"
    >
      <div className="dialog-content">
        <h2 className="dialog-title">Create Capability</h2>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="capability-name" className="form-label">
              Name <span className="required">*</span>
            </label>
            <input
              id="capability-name"
              type="text"
              className={`form-input ${errors.name ? 'form-input-error' : ''}`}
              value={form.name}
              onChange={(e) => handleFieldChange('name', e.target.value)}
              placeholder="Enter capability name"
              autoFocus
              disabled={isCreating}
              data-testid="capability-name-input"
            />
            {errors.name && (
              <div className="field-error" data-testid="capability-name-error">
                {errors.name}
              </div>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="capability-description" className="form-label">
              Description
            </label>
            <textarea
              id="capability-description"
              className={`form-textarea ${errors.description ? 'form-input-error' : ''}`}
              value={form.description}
              onChange={(e) => handleFieldChange('description', e.target.value)}
              placeholder="Enter capability description (optional)"
              rows={3}
              disabled={isCreating}
              data-testid="capability-description-input"
            />
            {errors.description && (
              <div className="field-error" data-testid="capability-description-error">
                {errors.description}
              </div>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="capability-status" className="form-label">
              Status
            </label>
            <select
              id="capability-status"
              className="form-select"
              value={form.status}
              onChange={(e) => handleFieldChange('status', e.target.value)}
              disabled={isCreating || isLoadingStatuses}
              data-testid="capability-status-select"
            >
              {isLoadingStatuses ? (
                <option value="">Loading...</option>
              ) : (
                statuses.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.displayName}
                  </option>
                ))
              )}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="capability-maturity" className="form-label">
              Maturity Level
            </label>
            <select
              id="capability-maturity"
              className="form-select"
              value={form.maturityLevel}
              onChange={(e) => handleFieldChange('maturityLevel', e.target.value)}
              disabled={isCreating || isLoadingMaturityLevels}
              data-testid="capability-maturity-select"
            >
              {isLoadingMaturityLevels ? (
                <option value="">Loading...</option>
              ) : (
                maturityLevels.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))
              )}
            </select>
          </div>

          {backendError && (
            <div className="error-message" data-testid="create-capability-error">
              {backendError}
            </div>
          )}

          <div className="dialog-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={isCreating}
              data-testid="create-capability-cancel"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isCreating || isLoadingMaturityLevels || isLoadingStatuses || !form.name.trim()}
              data-testid="create-capability-submit"
            >
              {isCreating ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </dialog>
  );
};
