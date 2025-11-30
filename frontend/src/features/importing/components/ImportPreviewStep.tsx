import React from 'react';
import type { ImportPreview } from '../types';

interface ImportPreviewStepProps {
  preview: ImportPreview;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
}

export const ImportPreviewStep: React.FC<ImportPreviewStepProps> = ({
  preview,
  onConfirm,
  onCancel,
  isLoading,
}) => {
  const { supported, unsupported } = preview;
  const hasUnsupported =
    Object.keys(unsupported.elements).length > 0 ||
    Object.keys(unsupported.relationships).length > 0;

  return (
    <div className="import-step">
      <h3>Import Preview</h3>
      <p className="import-step-description">
        Review what will be imported from the file.
      </p>

      <div className="import-preview">
        <div className="import-section">
          <h4>Will Import</h4>
          <ul className="import-list" data-testid="supported-list">
            <li>
              <strong>{supported.capabilities}</strong> Capabilities
            </li>
            <li>
              <strong>{supported.components}</strong> Components
            </li>
            <li>
              <strong>{supported.parentChildRelationships}</strong> Parent-child relationships
            </li>
            <li>
              <strong>{supported.realizations}</strong> Capability realizations
            </li>
          </ul>
        </div>

        {hasUnsupported && (
          <div className="import-section import-warning">
            <h4>Will NOT Import</h4>
            <p className="import-warning-text">
              The following unsupported elements will be skipped:
            </p>

            {Object.keys(unsupported.elements).length > 0 && (
              <div>
                <h5>Elements:</h5>
                <ul className="import-list" data-testid="unsupported-elements">
                  {Object.entries(unsupported.elements).map(([type, count]) => (
                    <li key={type}>
                      <strong>{count}</strong> {type}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {Object.keys(unsupported.relationships).length > 0 && (
              <div>
                <h5>Relationships:</h5>
                <ul className="import-list" data-testid="unsupported-relationships">
                  {Object.entries(unsupported.relationships).map(([type, count]) => (
                    <li key={type}>
                      <strong>{count}</strong> {type}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}
      </div>

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
          onClick={onConfirm}
          disabled={isLoading}
          data-testid="confirm-button"
        >
          {isLoading ? 'Starting...' : 'Confirm Import'}
        </button>
      </div>
    </div>
  );
};
