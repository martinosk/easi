interface LoadingFallbackProps {
  message?: string;
}

export function LoadingFallback({ message = 'Loading...' }: LoadingFallbackProps) {
  return (
    <div className="loading-fallback">
      <div className="loading-fallback-content">
        <div className="loading-fallback-spinner" />
        <span className="loading-fallback-text">{message}</span>
      </div>
    </div>
  );
}
