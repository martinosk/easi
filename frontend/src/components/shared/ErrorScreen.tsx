import React from 'react';

interface ErrorScreenProps {
  error: string;
  onRetry: () => void;
  retryLabel?: string;
  title?: string;
}

export const ErrorScreen: React.FC<ErrorScreenProps> = ({
  error,
  onRetry,
  retryLabel = 'Retry',
  title = 'Error Loading Data'
}) => {
  return (
    <div className="error-container">
      <h2>{title}</h2>
      <p>{error}</p>
      <button className="btn btn-primary" onClick={onRetry}>
        {retryLabel}
      </button>
    </div>
  );
};
