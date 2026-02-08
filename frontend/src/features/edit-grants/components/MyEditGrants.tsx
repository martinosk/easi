import { useMyEditGrants } from '../hooks/useEditGrants';

export function MyEditGrants() {
  const { data: grants, isLoading } = useMyEditGrants();

  if (isLoading) {
    return <div className="loading-spinner" data-testid="my-edit-grants-loading" />;
  }

  const activeGrants = grants?.filter(g => g.status === 'active') ?? [];

  if (activeGrants.length === 0) {
    return null;
  }

  return (
    <div className="card-list" data-testid="my-edit-grants">
      <h3>Your Edit Access</h3>
      {activeGrants.map(grant => (
        <div key={grant.id} className="card" data-testid={`my-grant-${grant.id}`}>
          <div className="card-body">
            <span className="card-title">
              {grant.artifactType}: {grant.artifactId}
            </span>
            <span className="card-subtitle">
              Granted by {grant.grantorEmail}
            </span>
            <span className="card-meta">
              Expires {new Date(grant.expiresAt).toLocaleDateString()}
            </span>
            {grant.reason && (
              <span className="card-description">{grant.reason}</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
