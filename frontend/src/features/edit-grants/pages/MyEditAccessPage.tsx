import { Link } from 'react-router-dom';
import { useMyEditGrants } from '../hooks/useEditGrants';
import type { EditGrant } from '../types';
import './MyEditAccessPage.css';

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function ArtifactCell({ grant }: { grant: EditGrant }) {
  const name = grant.artifactName || 'Deleted artifact';
  const href = grant._links?.artifact?.href;

  if (href) {
    return <Link to={href} className="my-edit-access-artifact-link">{name}</Link>;
  }
  return <span>{name}</span>;
}

export function MyEditAccessPage() {
  const { data: grants, isLoading } = useMyEditGrants();
  const activeGrants = grants?.filter(g => g.status === 'active') ?? [];

  if (isLoading) {
    return (
      <div className="my-edit-access-page">
        <div className="my-edit-access-container">
          <div className="loading-state">
            <div className="loading-spinner" />
            <span>Loading edit access...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="my-edit-access-page">
      <div className="my-edit-access-container">
        <div className="my-edit-access-header">
          <div>
            <h1 className="my-edit-access-title">My Edit Access</h1>
            <p className="my-edit-access-subtitle">
              Artifacts you have been granted write access to.
            </p>
          </div>
        </div>

        {activeGrants.length === 0 ? (
          <div className="empty-state">
            <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M18.5 2.50001C18.8978 2.10219 19.4374 1.87869 20 1.87869C20.5626 1.87869 21.1022 2.10219 21.5 2.50001C21.8978 2.89784 22.1213 3.4374 22.1213 4.00001C22.1213 4.56262 21.8978 5.10219 21.5 5.50001L12 15L8 16L9 12L18.5 2.50001Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <span className="empty-state-text">You have no active edit access grants</span>
          </div>
        ) : (
          <div className="my-edit-access-table-container">
            <table className="my-edit-access-table">
              <thead>
                <tr>
                  <th>Artifact</th>
                  <th>Granted by</th>
                  <th>Reason</th>
                  <th>Expires</th>
                </tr>
              </thead>
              <tbody>
                {activeGrants.map(grant => (
                  <tr key={grant.id}>
                    <td><ArtifactCell grant={grant} /></td>
                    <td>{grant.grantorEmail}</td>
                    <td>{grant.reason || '\u2014'}</td>
                    <td className="my-edit-access-date">{formatDate(grant.expiresAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

export default MyEditAccessPage;
