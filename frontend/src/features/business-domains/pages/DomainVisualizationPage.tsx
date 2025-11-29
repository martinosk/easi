import { useState } from 'react';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { DomainFilter } from '../components/DomainFilter';
import type { BusinessDomainId } from '../../../api/types';

export function DomainVisualizationPage() {
  const [selectedDomainId, setSelectedDomainId] = useState<BusinessDomainId | null>(null);
  const { domains, isLoading } = useBusinessDomains();

  const handleDomainSelect = (domainId: BusinessDomainId | null) => {
    setSelectedDomainId(domainId);
  };

  if (isLoading) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading domains...</div>
      </div>
    );
  }

  return (
    <div className="visualization-container" style={{ display: 'flex', height: '100vh' }}>
      <aside style={{ width: '250px', borderRight: '1px solid #e5e7eb', padding: '1rem' }}>
        <h2 style={{ marginBottom: '1rem' }}>Business Domains</h2>
        <DomainFilter
          domains={domains}
          selected={selectedDomainId}
          onSelect={handleDomainSelect}
        />
      </aside>

      <main style={{ flex: 1, padding: '2rem' }}>
        <div style={{ textAlign: 'center', marginTop: '4rem' }}>
          <h1>Grid Visualization</h1>
          <p style={{ color: '#6b7280', marginTop: '1rem' }}>
            {selectedDomainId
              ? `Grid visualization for domain will be implemented here`
              : 'Select a domain from the left sidebar'}
          </p>
        </div>
      </main>
    </div>
  );
}
