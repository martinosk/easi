import React, { useRef, useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import type { User } from '../../users/types';

interface ImportUploadStepProps {
  businessDomains: BusinessDomain[];
  eaOwnerCandidates: User[];
  isLoading: boolean;
  error: string | null;
  onUpload: (file: File, businessDomainId?: string, capabilityEAOwner?: string) => void;
  onCancel: () => void;
}

const FileInput: React.FC<{
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  selectedFile: File | null;
  isLoading: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}> = ({ fileInputRef, selectedFile, isLoading, onChange }) => (
  <div className="form-group">
    <label htmlFor="import-file" className="form-label">
      File <span className="required">*</span>
    </label>
    <input
      ref={fileInputRef}
      id="import-file"
      type="file"
      accept=".xml,application/xml,text/xml"
      onChange={onChange}
      className="form-input"
      disabled={isLoading}
      data-testid="file-input"
    />
    {selectedFile && (
      <div className="file-info" data-testid="selected-file">
        {selectedFile.name} ({Math.round(selectedFile.size / 1024)} KB)
      </div>
    )}
  </div>
);

const DomainSelect: React.FC<{
  businessDomains: BusinessDomain[];
  value: string;
  onChange: (value: string) => void;
  disabled: boolean;
}> = ({ businessDomains, value, onChange, disabled }) => (
  <div className="form-group">
    <label htmlFor="business-domain" className="form-label">
      Business Domain (Optional)
    </label>
    <select
      id="business-domain"
      className="form-select"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      data-testid="domain-select"
    >
      <option value="">None - Do not assign to domain</option>
      {businessDomains.map((domain) => (
        <option key={domain.id} value={domain.id}>
          {domain.name}
        </option>
      ))}
    </select>
    <div className="field-help">
      If selected, L1 capabilities will be assigned to this business domain.
    </div>
  </div>
);

const EAOwnerSelect: React.FC<{
  eaOwnerCandidates: User[];
  value: string;
  onChange: (value: string) => void;
  disabled: boolean;
}> = ({ eaOwnerCandidates, value, onChange, disabled }) => (
  <div className="form-group">
    <label htmlFor="ea-owner" className="form-label">
      EA Owner for Capabilities (Optional)
    </label>
    <select
      id="ea-owner"
      className="form-select"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      data-testid="ea-owner-select"
    >
      <option value="">Select EA Owner (optional)</option>
      {eaOwnerCandidates.map((user) => (
        <option key={user.id} value={user.id}>
          {user.name || user.email}
        </option>
      ))}
    </select>
    <div className="field-help">
      If selected, this user will be assigned as EA Owner to all imported capabilities.
    </div>
  </div>
);

export const ImportUploadStep: React.FC<ImportUploadStepProps> = ({
  businessDomains,
  eaOwnerCandidates,
  isLoading,
  error,
  onUpload,
  onCancel,
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<string>('');
  const [selectedEAOwner, setSelectedEAOwner] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) setSelectedFile(file);
  };

  const handleSubmit = () => {
    if (selectedFile) {
      onUpload(selectedFile, selectedDomain || undefined, selectedEAOwner || undefined);
    }
  };

  return (
    <div className="import-step">
      <h3>Upload ArchiMate File</h3>
      <p className="import-step-description">
        Select an ArchiMate Open Exchange XML file to import capabilities and components.
      </p>
      <FileInput
        fileInputRef={fileInputRef}
        selectedFile={selectedFile}
        isLoading={isLoading}
        onChange={handleFileChange}
      />
      <details className="import-options" open>
        <summary className="import-options-summary">Import Options</summary>
        <DomainSelect
          businessDomains={businessDomains}
          value={selectedDomain}
          onChange={setSelectedDomain}
          disabled={isLoading}
        />
        <EAOwnerSelect
          eaOwnerCandidates={eaOwnerCandidates}
          value={selectedEAOwner}
          onChange={setSelectedEAOwner}
          disabled={isLoading}
        />
      </details>
      {error && (
        <div className="error-message" data-testid="upload-error">
          {error}
        </div>
      )}
      <div className="dialog-actions">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={onCancel}
          disabled={isLoading}
          data-testid="cancel-button"
        >
          Cancel
        </button>
        <button
          type="button"
          className="btn btn-primary"
          onClick={handleSubmit}
          disabled={isLoading || !selectedFile}
          data-testid="upload-button"
        >
          {isLoading ? 'Uploading...' : 'Upload'}
        </button>
      </div>
    </div>
  );
};
