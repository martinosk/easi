import React, { useRef, useState } from 'react';
import type { BusinessDomain } from '../../../api/types';

interface ImportUploadStepProps {
  businessDomains: BusinessDomain[];
  isLoading: boolean;
  error: string | null;
  onUpload: (file: File, businessDomainId?: string) => void;
  onCancel: () => void;
}

export const ImportUploadStep: React.FC<ImportUploadStepProps> = ({
  businessDomains,
  isLoading,
  error,
  onUpload,
  onCancel,
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedDomain, setSelectedDomain] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  const handleSubmit = () => {
    if (selectedFile) {
      onUpload(selectedFile, selectedDomain || undefined);
    }
  };

  return (
    <div className="import-step">
      <h3>Upload ArchiMate File</h3>
      <p className="import-step-description">
        Select an ArchiMate Open Exchange XML file to import capabilities and components.
      </p>

      <div className="form-group">
        <label htmlFor="import-file" className="form-label">
          File <span className="required">*</span>
        </label>
        <input
          ref={fileInputRef}
          id="import-file"
          type="file"
          accept=".xml,application/xml,text/xml"
          onChange={handleFileChange}
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

      <div className="form-group">
        <label htmlFor="business-domain" className="form-label">
          Business Domain (Optional)
        </label>
        <select
          id="business-domain"
          className="form-select"
          value={selectedDomain}
          onChange={(e) => setSelectedDomain(e.target.value)}
          disabled={isLoading}
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
