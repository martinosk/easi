import React from 'react';
import type { ImportProgress } from '../types';

interface ImportProgressStepProps {
  progress: ImportProgress;
}

export const ImportProgressStep: React.FC<ImportProgressStepProps> = ({ progress }) => {
  const { phase, totalItems, completedItems } = progress;
  const percentage = totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0;

  const formatPhase = (phase: string): string => {
    return phase.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
  };

  return (
    <div className="import-step">
      <h3>Importing...</h3>
      <p className="import-step-description">
        Please wait while the import is in progress. Do not close this dialog.
      </p>

      <div className="import-progress">
        <div className="progress-info">
          <div className="progress-phase" data-testid="progress-phase">
            {formatPhase(phase)}
          </div>
          <div className="progress-stats" data-testid="progress-stats">
            {completedItems} / {totalItems} items
          </div>
        </div>

        <div className="progress-bar">
          <div
            className="progress-bar-fill"
            style={{ width: `${percentage}%` }}
            data-testid="progress-bar"
          />
        </div>

        <div className="progress-percentage" data-testid="progress-percentage">
          {percentage}%
        </div>
      </div>
    </div>
  );
};
