interface EditGrantsEmptyStateProps {
  statusFilter: string;
}

export function EditGrantsEmptyState({ statusFilter }: EditGrantsEmptyStateProps) {
  return (
    <div className="empty-state">
      <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M15 3H21V9M21 3L13 11M10 5H5C4.46957 5 3.96086 5.21071 3.58579 5.58579C3.21071 5.96086 3 6.46957 3 7V19C3 19.5304 3.21071 20.0391 3.58579 20.4142C3.96086 20.7893 4.46957 21 5 21H17C17.5304 21 18.0391 20.7893 18.4142 20.4142C18.7893 20.0391 19 19.5304 19 19V14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      </svg>
      <p className="empty-state-text">
        {statusFilter === 'all' ? 'No edit grants found' : `No ${statusFilter} edit grants`}
      </p>
    </div>
  );
}
