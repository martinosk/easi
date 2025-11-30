import React from 'react';
import type { ImportResult } from '../types';

interface ImportResultsStepProps {
  result: ImportResult;
  onClose: () => void;
}

export const ImportResultsStep: React.FC<ImportResultsStepProps> = ({ result, onClose }) => {
  const {
    capabilitiesCreated,
    componentsCreated,
    realizationsCreated,
    domainAssignments,
    errors,
  } = result;

  const hasErrors = errors.length > 0;
  const isSuccess = !hasErrors;

  return (
    <div className="import-step">
      <h3>{isSuccess ? 'Import Complete' : 'Import Completed with Errors'}</h3>

      <div className="import-results">
        <div className="import-section">
          <h4>Summary</h4>
          <ul className="import-list" data-testid="results-summary">
            <li>
              <strong>{capabilitiesCreated}</strong> Capabilities created
            </li>
            <li>
              <strong>{componentsCreated}</strong> Components created
            </li>
            <li>
              <strong>{realizationsCreated}</strong> Realizations created
            </li>
            {domainAssignments > 0 && (
              <li>
                <strong>{domainAssignments}</strong> Domain assignments made
              </li>
            )}
          </ul>
        </div>

        {hasErrors && (
          <div className="import-section import-errors">
            <h4>Errors ({errors.length})</h4>
            <div className="error-list" data-testid="error-list">
              {errors.map((error, index) => (
                <div key={index} className="error-item">
                  <div className="error-element">
                    <strong>{error.sourceName}</strong> ({error.sourceElement})
                  </div>
                  <div className="error-message">{error.error}</div>
                  <div className="error-action">Action: {error.action}</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="dialog-actions">
        <button
          type="button"
          className="btn btn-primary"
          onClick={onClose}
          data-testid="close-button"
        >
          Close
        </button>
      </div>
    </div>
  );
};
