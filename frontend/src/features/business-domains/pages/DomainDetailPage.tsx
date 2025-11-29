import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { DomainForm } from '../components/DomainForm';
import { CapabilityAssociationManager } from '../components/CapabilityAssociationManager';
import { apiClient } from '../../../api/client';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';

interface DomainDetailPageProps {
  domainId: BusinessDomainId;
}

export function DomainDetailPage({ domainId }: DomainDetailPageProps) {
  const [domain, setDomain] = useState<BusinessDomain | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    async function fetchDomain() {
      setIsLoading(true);
      setError(null);
      try {
        const data = await apiClient.getBusinessDomainById(domainId);
        setDomain(data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch domain'));
      } finally {
        setIsLoading(false);
      }
    }

    fetchDomain();
  }, [domainId]);

  const handleUpdateDomain = async (name: string, description: string) => {
    if (!domain) return;

    try {
      const updated = await apiClient.updateBusinessDomain(domain.id, { name, description });
      setDomain(updated);
      setIsEditing(false);
      toast.success('Domain updated successfully');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update domain');
      throw err;
    }
  };

  const handleBack = () => {
    window.location.hash = '#/business-domains';
  };

  if (isLoading) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading domain...</div>
      </div>
    );
  }

  if (error || !domain) {
    return (
      <div className="page-container">
        <div className="error-message" data-testid="domain-detail-error">
          {error?.message || 'Domain not found'}
        </div>
        <button type="button" className="btn btn-secondary" onClick={handleBack}>
          Back to Domains
        </button>
      </div>
    );
  }

  return (
    <div className="page-container" data-testid="domain-detail-page">
      <div className="breadcrumb">
        <button type="button" className="breadcrumb-link" onClick={handleBack} data-testid="back-to-domains">
          Business Domains
        </button>
        <span className="breadcrumb-separator">/</span>
        <span>{domain.name}</span>
      </div>

      <div className="page-header">
        <h1>{domain.name}</h1>
        {domain._links.update && !isEditing && (
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => setIsEditing(true)}
            data-testid="edit-domain-button"
          >
            Edit Details
          </button>
        )}
      </div>

      {isEditing ? (
        <div className="domain-edit-section" data-testid="domain-edit-section">
          <h2>Edit Domain</h2>
          <DomainForm
            mode="edit"
            domain={domain}
            onSubmit={handleUpdateDomain}
            onCancel={() => setIsEditing(false)}
          />
        </div>
      ) : (
        <div className="domain-info-section" data-testid="domain-info-section">
          <div className="domain-detail-field">
            <label className="detail-label">Description:</label>
            <p className="detail-value">{domain.description || 'No description'}</p>
          </div>
          <div className="domain-detail-field">
            <label className="detail-label">Created:</label>
            <p className="detail-value">{new Date(domain.createdAt).toLocaleString()}</p>
          </div>
          {domain.updatedAt && (
            <div className="domain-detail-field">
              <label className="detail-label">Last Updated:</label>
              <p className="detail-value">{new Date(domain.updatedAt).toLocaleString()}</p>
            </div>
          )}
          <div className="domain-detail-field">
            <label className="detail-label">Capability Count:</label>
            <p className="detail-value">{domain.capabilityCount}</p>
          </div>
        </div>
      )}

      <div className="domain-capabilities-section">
        <CapabilityAssociationManager
          capabilitiesLink={domain._links.capabilities}
        />
      </div>
    </div>
  );
}
