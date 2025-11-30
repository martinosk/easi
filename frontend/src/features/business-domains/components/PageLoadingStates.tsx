interface PageLoadingStatesProps {
  isLoading: boolean;
  hasData: boolean;
  error: Error | null;
  children: React.ReactNode;
}

export function PageLoadingStates({ isLoading, hasData, error, children }: PageLoadingStatesProps) {
  if (isLoading && !hasData) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading business domains...</div>
      </div>
    );
  }

  if (error && !hasData) {
    return (
      <div className="page-container">
        <div className="error-message" data-testid="domains-error">
          {error.message}
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
