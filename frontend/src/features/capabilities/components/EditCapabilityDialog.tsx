import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../../../store/appStore';
import { apiClient } from '../../../api/client';
import type { Capability, Expert, StatusOption, OwnershipModelOption, StrategyPillarOption } from '../../../api/types';
import { AddExpertDialog } from './AddExpertDialog';
import { AddTagDialog } from './AddTagDialog';

interface EditCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
}

const DEFAULT_MATURITY_LEVELS = ['Genesis', 'Custom Build', 'Product', 'Commodity'];
const DEFAULT_STATUSES: StatusOption[] = [
  { value: 'Active', displayName: 'Active', sortOrder: 1 },
  { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
  { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
];
const DEFAULT_OWNERSHIP_MODELS: OwnershipModelOption[] = [
  { value: 'TribeOwned', displayName: 'Tribe Owned' },
  { value: 'TeamOwned', displayName: 'Team Owned' },
  { value: 'Shared', displayName: 'Shared' },
  { value: 'EnterpriseService', displayName: 'Enterprise Service' },
];
const DEFAULT_STRATEGY_PILLARS: StrategyPillarOption[] = [
  { value: 'AlwaysOn', displayName: 'Always On' },
  { value: 'Grow', displayName: 'Grow' },
  { value: 'Transform', displayName: 'Transform' },
];

interface FormState {
  name: string;
  description: string;
  status: string;
  maturityLevel: string;
  ownershipModel: string;
  primaryOwner: string;
  eaOwner: string;
  strategyPillar: string;
  pillarWeight: number;
}

interface FormErrors {
  name?: string;
  description?: string;
  pillarWeight?: string;
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
  if (form.pillarWeight < 0 || form.pillarWeight > 100) {
    errors.pillarWeight = 'Pillar weight must be between 0 and 100';
  }
  return errors;
};

const EMPTY_FORM: FormState = {
  name: '',
  description: '',
  status: 'Active',
  maturityLevel: '',
  ownershipModel: '',
  primaryOwner: '',
  eaOwner: '',
  strategyPillar: '',
  pillarWeight: 0,
};

const createInitialFormState = (cap?: Capability): FormState => {
  if (!cap) return EMPTY_FORM;
  return {
    name: cap.name ?? '',
    description: cap.description ?? '',
    status: cap.status ?? 'Active',
    maturityLevel: cap.maturityLevel ?? '',
    ownershipModel: cap.ownershipModel ?? '',
    primaryOwner: cap.primaryOwner ?? '',
    eaOwner: cap.eaOwner ?? '',
    strategyPillar: cap.strategyPillar ?? '',
    pillarWeight: cap.pillarWeight ?? 0,
  };
};

interface TextInputProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  disabled?: boolean;
  required?: boolean;
  placeholder?: string;
}

const TextInput: React.FC<TextInputProps> = ({ id, label, value, onChange, error, disabled, required, placeholder }) => (
  <div className="form-group">
    <label htmlFor={id} className="form-label">
      {label} {required && <span className="required">*</span>}
    </label>
    <input
      id={id}
      type="text"
      className={`form-input ${error ? 'form-input-error' : ''}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      disabled={disabled}
      data-testid={`${id}-input`}
    />
    {error && <div className="field-error" data-testid={`${id}-error`}>{error}</div>}
  </div>
);

interface SelectOption {
  value: string;
  displayName: string;
}

interface SelectInputProps {
  id: string;
  label: string;
  value: string;
  options: SelectOption[];
  onChange: (value: string) => void;
  disabled?: boolean;
  placeholder?: string;
  isLoading?: boolean;
}

const SelectInput: React.FC<SelectInputProps> = ({ id, label, value, options, onChange, disabled, placeholder, isLoading }) => (
  <div className="form-group form-group-half">
    <label htmlFor={id} className="form-label">{label}</label>
    <select
      id={id}
      className="form-select"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled || isLoading}
      data-testid={`${id}-select`}
    >
      {isLoading ? (
        <option value="">Loading...</option>
      ) : (
        <>
          {placeholder && <option value="">{placeholder}</option>}
          {options.map((option) => (
            <option key={option.value} value={option.value}>{option.displayName}</option>
          ))}
        </>
      )}
    </select>
  </div>
);

interface ExpertsListProps {
  experts?: Expert[];
  onAddClick: () => void;
  disabled?: boolean;
}

const ExpertsList: React.FC<ExpertsListProps> = ({ experts, onAddClick, disabled }) => (
  <div className="form-group">
    <label className="form-label">Experts</label>
    <div className="expert-list">
      {experts && experts.length > 0 ? (
        experts.map((expert, index) => (
          <div key={index} className="expert-item">
            <span className="expert-name">{expert.name}</span>
            <span className="expert-role">({expert.role})</span>
            <span className="expert-contact">- {expert.contact}</span>
          </div>
        ))
      ) : (
        <div className="empty-list">No experts added</div>
      )}
    </div>
    <button type="button" className="btn btn-link" onClick={onAddClick} disabled={disabled} data-testid="add-expert-button">
      + Add Expert
    </button>
  </div>
);

interface TagsListProps {
  tags?: string[];
  onAddClick: () => void;
  disabled?: boolean;
}

const TagsList: React.FC<TagsListProps> = ({ tags, onAddClick, disabled }) => (
  <div className="form-group">
    <label className="form-label">Tags</label>
    <div className="tag-list">
      {tags && tags.length > 0 ? (
        tags.map((tag, index) => (
          <span key={index} className="tag-badge">{tag}</span>
        ))
      ) : (
        <div className="empty-list">No tags added</div>
      )}
    </div>
    <button type="button" className="btn btn-link" onClick={onAddClick} disabled={disabled} data-testid="add-tag-button">
      + Add Tag
    </button>
  </div>
);

export const EditCapabilityDialog: React.FC<EditCapabilityDialogProps> = ({ isOpen, onClose, capability }) => {
  const [form, setForm] = useState<FormState>(createInitialFormState());
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSaving, setIsSaving] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);
  const [maturityLevels, setMaturityLevels] = useState<string[]>([]);
  const [isLoadingMaturityLevels, setIsLoadingMaturityLevels] = useState(false);
  const [statuses, setStatuses] = useState<StatusOption[]>([]);
  const [isLoadingStatuses, setIsLoadingStatuses] = useState(false);
  const [ownershipModels, setOwnershipModels] = useState<OwnershipModelOption[]>([]);
  const [isLoadingOwnershipModels, setIsLoadingOwnershipModels] = useState(false);
  const [strategyPillars, setStrategyPillars] = useState<StrategyPillarOption[]>([]);
  const [isLoadingStrategyPillars, setIsLoadingStrategyPillars] = useState(false);
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const [isAddTagOpen, setIsAddTagOpen] = useState(false);
  const [currentCapability, setCurrentCapability] = useState<Capability | null>(null);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateCapability = useAppStore((state) => state.updateCapability);
  const updateCapabilityMetadata = useAppStore((state) => state.updateCapabilityMetadata);
  const capabilities = useAppStore((state) => state.capabilities);

  useEffect(() => {
    if (capability) {
      const updated = capabilities.find((c) => c.id === capability.id);
      setCurrentCapability(updated || capability);
    }
  }, [capability, capabilities]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (isOpen && capability) {
      dialog.showModal();
      fetchMetadata();
      setForm(createInitialFormState(capability));
      setErrors({});
      setBackendError(null);
    } else {
      dialog.close();
    }
  }, [isOpen, capability]);

  const fetchMetadata = async () => {
    setIsLoadingMaturityLevels(true);
    setIsLoadingStatuses(true);
    setIsLoadingOwnershipModels(true);
    setIsLoadingStrategyPillars(true);

    try {
      const levels = await apiClient.getMaturityLevels();
      setMaturityLevels(levels);
    } catch {
      setMaturityLevels(DEFAULT_MATURITY_LEVELS);
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

    try {
      const ownershipList = await apiClient.getOwnershipModels();
      setOwnershipModels(ownershipList);
    } catch {
      setOwnershipModels(DEFAULT_OWNERSHIP_MODELS);
    } finally {
      setIsLoadingOwnershipModels(false);
    }

    try {
      const pillarList = await apiClient.getStrategyPillars();
      setStrategyPillars(pillarList);
    } catch {
      setStrategyPillars(DEFAULT_STRATEGY_PILLARS);
    } finally {
      setIsLoadingStrategyPillars(false);
    }
  };

  const handleClose = () => {
    setErrors({});
    setBackendError(null);
    onClose();
  };

  const handleFieldChange = (field: keyof FormState, value: string | number) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field as keyof FormErrors]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
    if (backendError) {
      setBackendError(null);
    }
  };

  const buildMetadataRequest = () => {
    const defaultMaturity = maturityLevels[0] ?? DEFAULT_MATURITY_LEVELS[0];
    return {
      status: form.status,
      maturityLevel: form.maturityLevel || defaultMaturity,
      ownershipModel: form.ownershipModel || undefined,
      primaryOwner: form.primaryOwner.trim() || undefined,
      eaOwner: form.eaOwner.trim() || undefined,
      strategyPillar: form.strategyPillar || undefined,
      pillarWeight: form.pillarWeight || undefined,
    };
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!capability) return;
    setBackendError(null);
    const validationErrors = validateForm(form);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }
    setIsSaving(true);
    try {
      const description = form.description.trim() || undefined;
      await updateCapability(capability.id, { name: form.name.trim(), description });
      await updateCapabilityMetadata(capability.id, buildMetadataRequest());
      handleClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update capability';
      setBackendError(message);
    } finally {
      setIsSaving(false);
    }
  };

  if (!capability) return null;
  const displayCapability = currentCapability || capability;

  return (
    <>
      <dialog ref={dialogRef} className="dialog dialog-large" onClose={handleClose} data-testid="edit-capability-dialog">
        <div className="dialog-content">
          <h2 className="dialog-title">Edit Capability</h2>
          <form onSubmit={handleSubmit}>
            <TextInput id="edit-capability-name" label="Name" value={form.name} onChange={(v) => handleFieldChange('name', v)} error={errors.name} disabled={isSaving} required placeholder="Enter capability name" />

            <div className="form-group">
              <label htmlFor="edit-capability-description" className="form-label">Description</label>
              <textarea id="edit-capability-description" className={`form-textarea ${errors.description ? 'form-input-error' : ''}`} value={form.description} onChange={(e) => handleFieldChange('description', e.target.value)} placeholder="Enter capability description (optional)" rows={3} disabled={isSaving} data-testid="edit-capability-description-input" />
              {errors.description && <div className="field-error" data-testid="edit-capability-description-error">{errors.description}</div>}
            </div>

            <div className="form-row">
              <SelectInput id="edit-capability-status" label="Status" value={form.status} options={statuses || []} onChange={(v) => handleFieldChange('status', v)} disabled={isSaving} isLoading={isLoadingStatuses} />
              <SelectInput id="edit-capability-maturity" label="Maturity Level" value={form.maturityLevel} options={(maturityLevels || []).map((level) => ({ value: level, displayName: level }))} onChange={(v) => handleFieldChange('maturityLevel', v)} disabled={isSaving} isLoading={isLoadingMaturityLevels} />
            </div>

            <div className="form-row">
              <SelectInput id="edit-capability-ownership" label="Ownership Model" value={form.ownershipModel} options={ownershipModels || []} onChange={(v) => handleFieldChange('ownershipModel', v)} disabled={isSaving} placeholder="Select ownership model" isLoading={isLoadingOwnershipModels} />
              <div className="form-group form-group-half">
                <label htmlFor="edit-capability-primary-owner" className="form-label">Primary Owner</label>
                <input id="edit-capability-primary-owner" type="text" className="form-input" value={form.primaryOwner} onChange={(e) => handleFieldChange('primaryOwner', e.target.value)} placeholder="Enter primary owner" disabled={isSaving} data-testid="edit-capability-primary-owner-input" />
              </div>
            </div>

            <div className="form-row">
              <div className="form-group form-group-half">
                <label htmlFor="edit-capability-ea-owner" className="form-label">EA Owner</label>
                <input id="edit-capability-ea-owner" type="text" className="form-input" value={form.eaOwner} onChange={(e) => handleFieldChange('eaOwner', e.target.value)} placeholder="Enter EA owner" disabled={isSaving} data-testid="edit-capability-ea-owner-input" />
              </div>
              <SelectInput id="edit-capability-strategy-pillar" label="Strategy Pillar" value={form.strategyPillar} options={strategyPillars || []} onChange={(v) => handleFieldChange('strategyPillar', v)} disabled={isSaving} placeholder="Select strategy pillar" isLoading={isLoadingStrategyPillars} />
            </div>

            <div className="form-group">
              <label htmlFor="edit-capability-pillar-weight" className="form-label">Pillar Weight (0-100)</label>
              <input id="edit-capability-pillar-weight" type="number" min="0" max="100" className={`form-input form-input-narrow ${errors.pillarWeight ? 'form-input-error' : ''}`} value={form.pillarWeight} onChange={(e) => handleFieldChange('pillarWeight', parseInt(e.target.value, 10) || 0)} disabled={isSaving} data-testid="edit-capability-pillar-weight-input" />
              {errors.pillarWeight && <div className="field-error" data-testid="edit-capability-pillar-weight-error">{errors.pillarWeight}</div>}
            </div>

            <ExpertsList experts={displayCapability.experts} onAddClick={() => setIsAddExpertOpen(true)} disabled={isSaving} />
            <TagsList tags={displayCapability.tags} onAddClick={() => setIsAddTagOpen(true)} disabled={isSaving} />

            {backendError && <div className="error-message" data-testid="edit-capability-error">{backendError}</div>}

            <div className="dialog-actions">
              <button type="button" className="btn btn-secondary" onClick={handleClose} disabled={isSaving} data-testid="edit-capability-cancel">Cancel</button>
              <button type="submit" className="btn btn-primary" disabled={isSaving || isLoadingMaturityLevels || isLoadingStatuses || isLoadingOwnershipModels || isLoadingStrategyPillars || !form.name.trim()} data-testid="edit-capability-submit">{isSaving ? 'Saving...' : 'Save'}</button>
            </div>
          </form>
        </div>
      </dialog>

      <AddExpertDialog isOpen={isAddExpertOpen} onClose={() => setIsAddExpertOpen(false)} capabilityId={capability.id} />
      <AddTagDialog isOpen={isAddTagOpen} onClose={() => setIsAddTagOpen(false)} capabilityId={capability.id} />
    </>
  );
};
