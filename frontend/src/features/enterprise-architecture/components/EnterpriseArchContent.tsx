import React from 'react';
import { EnterpriseCapabilitiesTable } from './EnterpriseCapabilitiesTable';
import { EnterpriseCapabilitiesEmptyState } from './EnterpriseCapabilitiesEmptyState';
import type { EnterpriseCapability } from '../types';

interface EnterpriseArchContentProps {
  isLoading: boolean;
  error: string | null;
  capabilities: EnterpriseCapability[];
  selectedCapability: EnterpriseCapability | null;
  canWrite: boolean;
  canDelete: boolean;
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  onCreateNew: () => void;
}

export const EnterpriseArchContent = React.memo<EnterpriseArchContentProps>(({
  isLoading,
  error,
  capabilities,
  selectedCapability,
  canWrite,
  canDelete,
  onSelect,
  onDelete,
  onCreateNew,
}) => {
  if (isLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <p>Loading enterprise capabilities...</p>
      </div>
    );
  }

  if (error) {
    return <div className="error-message" data-testid="capabilities-error">{error}</div>;
  }

  if (capabilities.length === 0) {
    return <EnterpriseCapabilitiesEmptyState onCreateNew={onCreateNew} canWrite={canWrite} />;
  }

  return (
    <EnterpriseCapabilitiesTable
      capabilities={capabilities}
      selectedId={selectedCapability?.id}
      onSelect={onSelect}
      onDelete={onDelete}
      canDelete={canDelete}
    />
  );
});

EnterpriseArchContent.displayName = 'EnterpriseArchContent';
