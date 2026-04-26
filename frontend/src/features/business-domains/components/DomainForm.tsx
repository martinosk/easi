import { useEffect, useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import type { User } from '../../users/types';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';

type SubmitFn = (name: string, description: string, domainArchitectId?: string) => Promise<void>;

interface DomainFormProps {
  domain?: BusinessDomain;
  mode: 'create' | 'edit';
  onSubmit: SubmitFn;
  onCancel: () => void;
}

interface FormState {
  name: string;
  description: string;
  domainArchitectId: string;
}

interface FormErrors {
  name?: string;
  description?: string;
}

function initialFormState(domain?: BusinessDomain): FormState {
  return {
    name: domain?.name || '',
    description: domain?.description || '',
    domainArchitectId: domain?.domainArchitectId || '',
  };
}

function validateName(name: string): string | undefined {
  const trimmed = name.trim();
  if (!trimmed) return 'Name is required';
  if (trimmed.length > 100) return 'Name must be 100 characters or less';
  return undefined;
}

function validateDescription(description: string): string | undefined {
  if (description.length > 500) return 'Description must be 500 characters or less';
  return undefined;
}

function validateForm(form: FormState): FormErrors {
  return {
    name: validateName(form.name),
    description: validateDescription(form.description),
  };
}

function hasErrors(errors: FormErrors): boolean {
  return Object.values(errors).some((v) => v !== undefined);
}

function getSubmitButtonText(isSubmitting: boolean, mode: 'create' | 'edit'): string {
  if (isSubmitting) return 'Saving...';
  return mode === 'create' ? 'Create' : 'Save';
}

function useDomainForm(domain: BusinessDomain | undefined, onSubmit: SubmitFn) {
  const [form, setForm] = useState<FormState>(() => initialFormState(domain));
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  useEffect(() => {
    if (domain) {
      setForm(initialFormState(domain));
    }
  }, [domain]);

  const handleFieldChange = (field: keyof FormState, value: string) => {
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
    if (hasErrors(validationErrors)) {
      setErrors(validationErrors);
      return;
    }

    setIsSubmitting(true);

    try {
      await onSubmit(form.name.trim(), form.description.trim(), form.domainArchitectId || undefined);
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  return { form, errors, isSubmitting, backendError, handleFieldChange, handleSubmit };
}

interface TextFieldProps {
  id: string;
  label: React.ReactNode;
  value: string;
  error: string | undefined;
  placeholder: string;
  isSubmitting: boolean;
  onChange: (value: string) => void;
  multiline?: boolean;
  testIdInput: string;
  testIdError: string;
  autoFocus?: boolean;
}

function TextField(props: TextFieldProps) {
  const baseClass = props.multiline ? 'form-textarea' : 'form-input';
  const className = props.error ? `${baseClass} form-input-error` : baseClass;
  const sharedProps = {
    id: props.id,
    className,
    value: props.value,
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => props.onChange(e.target.value),
    placeholder: props.placeholder,
    disabled: props.isSubmitting,
    'data-testid': props.testIdInput,
  };
  return (
    <div className="form-group">
      <label htmlFor={props.id} className="form-label">
        {props.label}
      </label>
      {props.multiline ? (
        <textarea {...sharedProps} rows={4} />
      ) : (
        <input {...sharedProps} type="text" autoFocus={props.autoFocus} />
      )}
      {props.error && (
        <div className="field-error" data-testid={props.testIdError}>
          {props.error}
        </div>
      )}
    </div>
  );
}

interface DomainArchitectFieldProps {
  value: string;
  eligibleUsers: User[];
  isLoadingUsers: boolean;
  isSubmitting: boolean;
  onChange: (value: string) => void;
}

function DomainArchitectField({ value, eligibleUsers, isLoadingUsers, isSubmitting, onChange }: DomainArchitectFieldProps) {
  return (
    <div className="form-group">
      <label htmlFor="domain-architect" className="form-label">
        Domain Architect
      </label>
      <select
        id="domain-architect"
        className="form-input"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={isSubmitting || isLoadingUsers}
        data-testid="domain-architect-select"
      >
        <option value="">-- Select Domain Architect (optional) --</option>
        {eligibleUsers.map((user) => (
          <option key={user.id} value={user.id}>
            {user.name || user.email} ({user.role})
          </option>
        ))}
      </select>
      {isLoadingUsers && <div className="field-hint">Loading eligible users...</div>}
    </div>
  );
}

interface FormActionsProps {
  mode: 'create' | 'edit';
  isSubmitting: boolean;
  canSubmit: boolean;
  onCancel: () => void;
}

function FormActions({ mode, isSubmitting, canSubmit, onCancel }: FormActionsProps) {
  return (
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
        disabled={!canSubmit}
        data-testid="domain-form-submit"
      >
        {getSubmitButtonText(isSubmitting, mode)}
      </button>
    </div>
  );
}

export function DomainForm({ domain, mode, onSubmit, onCancel }: DomainFormProps) {
  const { data: eligibleUsers = [], isLoading: isLoadingUsers } = useEAOwnerCandidates();
  const { form, errors, isSubmitting, backendError, handleFieldChange, handleSubmit } = useDomainForm(domain, onSubmit);
  const canSubmit = !isSubmitting && form.name.trim().length > 0;

  return (
    <form onSubmit={handleSubmit} className="domain-form" data-testid="domain-form">
      <TextField
        id="domain-name"
        label={<>Name <span className="required">*</span></>}
        value={form.name}
        error={errors.name}
        placeholder="Enter domain name"
        isSubmitting={isSubmitting}
        onChange={(value) => handleFieldChange('name', value)}
        autoFocus
        testIdInput="domain-name-input"
        testIdError="domain-name-error"
      />
      <TextField
        id="domain-description"
        label="Description"
        value={form.description}
        error={errors.description}
        placeholder="Enter domain description (optional)"
        isSubmitting={isSubmitting}
        onChange={(value) => handleFieldChange('description', value)}
        multiline
        testIdInput="domain-description-input"
        testIdError="domain-description-error"
      />
      <DomainArchitectField
        value={form.domainArchitectId}
        eligibleUsers={eligibleUsers}
        isLoadingUsers={isLoadingUsers}
        isSubmitting={isSubmitting}
        onChange={(value) => handleFieldChange('domainArchitectId', value)}
      />
      {backendError && (
        <div className="error-message" data-testid="domain-form-error">
          {backendError}
        </div>
      )}
      <FormActions mode={mode} isSubmitting={isSubmitting} canSubmit={canSubmit} onCancel={onCancel} />
    </form>
  );
}
