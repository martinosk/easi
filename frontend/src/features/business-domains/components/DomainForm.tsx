import { useState, useEffect } from 'react';
import type { BusinessDomain } from '../../../api/types';

interface DomainFormProps {
  domain?: BusinessDomain;
  mode: 'create' | 'edit';
  onSubmit: (name: string, description: string) => Promise<void>;
  onCancel: () => void;
}

interface FormState {
  name: string;
  description: string;
}

interface FormErrors {
  name?: string;
  description?: string;
}

function validateForm(form: FormState): FormErrors {
  const errors: FormErrors = {};

  if (!form.name.trim()) {
    errors.name = 'Name is required';
  } else if (form.name.trim().length > 100) {
    errors.name = 'Name must be 100 characters or less';
  }

  if (form.description.length > 500) {
    errors.description = 'Description must be 500 characters or less';
  }

  return errors;
}

function getInputClassName(hasError: boolean): string {
  return hasError ? 'form-input form-input-error' : 'form-input';
}

function getTextareaClassName(hasError: boolean): string {
  return hasError ? 'form-textarea form-input-error' : 'form-textarea';
}

function getSubmitButtonText(isSubmitting: boolean, mode: 'create' | 'edit'): string {
  if (isSubmitting) return 'Saving...';
  return mode === 'create' ? 'Create' : 'Save';
}

export function DomainForm({ domain, mode, onSubmit, onCancel }: DomainFormProps) {
  const [form, setForm] = useState<FormState>({
    name: domain?.name || '',
    description: domain?.description || '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  useEffect(() => {
    if (domain) {
      setForm({
        name: domain.name,
        description: domain.description || '',
      });
    }
  }, [domain]);

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

    setIsSubmitting(true);

    try {
      await onSubmit(form.name.trim(), form.description.trim());
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="domain-form" data-testid="domain-form">
      <div className="form-group">
        <label htmlFor="domain-name" className="form-label">
          Name <span className="required">*</span>
        </label>
        <input
          id="domain-name"
          type="text"
          className={getInputClassName(!!errors.name)}
          value={form.name}
          onChange={(e) => handleFieldChange('name', e.target.value)}
          placeholder="Enter domain name"
          autoFocus
          disabled={isSubmitting}
          data-testid="domain-name-input"
        />
        {errors.name && (
          <div className="field-error" data-testid="domain-name-error">
            {errors.name}
          </div>
        )}
      </div>

      <div className="form-group">
        <label htmlFor="domain-description" className="form-label">
          Description
        </label>
        <textarea
          id="domain-description"
          className={getTextareaClassName(!!errors.description)}
          value={form.description}
          onChange={(e) => handleFieldChange('description', e.target.value)}
          placeholder="Enter domain description (optional)"
          rows={4}
          disabled={isSubmitting}
          data-testid="domain-description-input"
        />
        {errors.description && (
          <div className="field-error" data-testid="domain-description-error">
            {errors.description}
          </div>
        )}
      </div>

      {backendError && (
        <div className="error-message" data-testid="domain-form-error">
          {backendError}
        </div>
      )}

      <div className="form-actions">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={onCancel}
          disabled={isSubmitting}
          data-testid="domain-form-cancel"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="btn btn-primary"
          disabled={isSubmitting || !form.name.trim()}
          data-testid="domain-form-submit"
        >
          {getSubmitButtonText(isSubmitting, mode)}
        </button>
      </div>
    </form>
  );
}
