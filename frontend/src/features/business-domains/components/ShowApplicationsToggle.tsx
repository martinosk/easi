export interface ShowApplicationsToggleProps {
  showApplications: boolean;
  onShowApplicationsChange: (value: boolean) => void;
}

export function ShowApplicationsToggle({
  showApplications,
  onShowApplicationsChange,
}: ShowApplicationsToggleProps) {
  return (
    <div
      style={{
        display: 'flex',
        gap: '0.25rem',
        padding: '0.25rem',
        backgroundColor: '#f3f4f6',
        borderRadius: '0.5rem',
      }}
    >
      <button
        type="button"
        data-selected={showApplications}
        onClick={() => {
          onShowApplicationsChange(!showApplications);
        }}
        style={{
          padding: '0.5rem 0.75rem',
          borderRadius: '0.375rem',
          border: 'none',
          cursor: 'pointer',
          backgroundColor: showApplications ? '#3b82f6' : 'transparent',
          color: showApplications ? 'white' : '#374151',
          fontWeight: 500,
          fontSize: '0.875rem',
          transition: 'all 0.15s ease',
        }}
      >
        Apps
      </button>
    </div>
  );
}
